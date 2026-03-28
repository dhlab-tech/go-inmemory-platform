package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dhlab-tech/go-inmemory-platform/pkg/domain"
	"github.com/dhlab-tech/go-inmemory-platform/pkg/projection"
	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/rs/zerolog"
)

// StreamListener applies WAL events to the projection (same role as Mongo change stream listener).
type StreamListener interface {
	Listen(ctx context.Context) error
}

// LogicalReplicationConfig selects the pgoutput stream for one publication/slot.
type LogicalReplicationConfig struct {
	// ConnString must use replication= (e.g. postgres://.../db?replication=database).
	ConnString  string
	Publication string
	SlotName    string
	// TemporarySlot creates a temporary slot (tests); otherwise persistent.
	TemporarySlot bool
}

// logicalReplListener applies Insert/Update/Delete from pgoutput to a [projection.StreamEventListener].
// Full row updates use Delete+Add so indexes stay consistent without partial merge from WAL.
type logicalReplListener[T d] struct {
	cfg          LogicalReplicationConfig
	schema       string
	table        string
	handler      projection.StreamEventListener[T]
	decode       TupleDecoder[T]
	relations    map[uint32]*pglogrepl.RelationMessageV2
	inStream     bool
	clientXLog   pglogrepl.LSN
	standbyEvery time.Duration
}

// NewLogicalReplicationListener builds a [StreamListener] for pgoutput logical replication.
// decode must map pgoutput tuples to T (typically generated); WAL events are applied via handler.
func NewLogicalReplicationListener[T d](
	cfg LogicalReplicationConfig,
	schema, table string,
	handler projection.StreamEventListener[T],
	decode TupleDecoder[T],
) StreamListener {
	sch := schema
	if sch == "" {
		sch = "public"
	}
	return &logicalReplListener[T]{
		cfg:          cfg,
		schema:       sch,
		table:        table,
		handler:      handler,
		decode:       decode,
		relations:    map[uint32]*pglogrepl.RelationMessageV2{},
		standbyEvery: 10 * time.Second,
	}
}

// Listen blocks, processing WAL until ctx is cancelled. Sends standby status updates so the master can recycle WAL.
func (l *logicalReplListener[T]) Listen(ctx context.Context) error {
	if l.cfg.ConnString == "" || l.cfg.Publication == "" || l.cfg.SlotName == "" {
		return errors.New("postgres: logical replication: empty conn string, publication, or slot name")
	}
	conn, err := pgconn.Connect(ctx, l.cfg.ConnString)
	if err != nil {
		return fmt.Errorf("postgres: replication connect: %w", err)
	}
	defer conn.Close(ctx)

	sysident, err := pglogrepl.IdentifySystem(ctx, conn)
	if err != nil {
		return fmt.Errorf("postgres: IdentifySystem: %w", err)
	}
	zerolog.Ctx(ctx).Info().
		Str("slot", l.cfg.SlotName).
		Str("system", sysident.SystemID).
		Str("xlogpos", sysident.XLogPos.String()).
		Msg("postgres logical replication: identified system")

	_, err = pglogrepl.CreateReplicationSlot(ctx, conn, l.cfg.SlotName, "pgoutput", pglogrepl.CreateReplicationSlotOptions{
		Temporary: l.cfg.TemporarySlot,
	})
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("postgres: CreateReplicationSlot: %w", err)
	}

	pluginArgs := []string{
		"proto_version '2'",
		fmt.Sprintf("publication_names '%s'", l.cfg.Publication),
		"messages 'true'",
	}
	err = pglogrepl.StartReplication(ctx, conn, l.cfg.SlotName, sysident.XLogPos, pglogrepl.StartReplicationOptions{
		PluginArgs: pluginArgs,
	})
	if err != nil {
		return fmt.Errorf("postgres: StartReplication: %w", err)
	}

	l.clientXLog = sysident.XLogPos
	nextStandby := time.Now().Add(l.standbyEvery)

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if time.Now().After(nextStandby) {
			if err := pglogrepl.SendStandbyStatusUpdate(ctx, conn, pglogrepl.StandbyStatusUpdate{
				WALWritePosition: l.clientXLog,
			}); err != nil {
				return fmt.Errorf("postgres: SendStandbyStatusUpdate: %w", err)
			}
			nextStandby = time.Now().Add(l.standbyEvery)
		}

		subCtx, cancel := context.WithDeadline(ctx, time.Now().Add(l.standbyEvery))
		rawMsg, err := conn.ReceiveMessage(subCtx)
		cancel()
		if err != nil {
			if pgconn.Timeout(err) {
				continue
			}
			return fmt.Errorf("postgres: ReceiveMessage: %w", err)
		}

		if errMsg, ok := rawMsg.(*pgproto3.ErrorResponse); ok {
			return fmt.Errorf("postgres replication error: %s", errMsg.Message)
		}

		msg, ok := rawMsg.(*pgproto3.CopyData)
		if !ok {
			continue
		}

		switch msg.Data[0] {
		case pglogrepl.PrimaryKeepaliveMessageByteID:
			pkm, err := pglogrepl.ParsePrimaryKeepaliveMessage(msg.Data[1:])
			if err != nil {
				return err
			}
			if pkm.ServerWALEnd > l.clientXLog {
				l.clientXLog = pkm.ServerWALEnd
			}
		case pglogrepl.XLogDataByteID:
			xld, err := pglogrepl.ParseXLogData(msg.Data[1:])
			if err != nil {
				return err
			}
			if err := l.processV2(ctx, xld.WALData); err != nil {
				return err
			}
			if xld.WALStart > l.clientXLog {
				l.clientXLog = xld.WALStart
			}
		}
	}
}

func (l *logicalReplListener[T]) processV2(ctx context.Context, walData []byte) error {
	logicalMsg, err := pglogrepl.ParseV2(walData, l.inStream)
	if err != nil {
		return err
	}
	switch m := logicalMsg.(type) {
	case *pglogrepl.RelationMessageV2:
		l.relations[m.RelationID] = m
	case *pglogrepl.InsertMessageV2:
		if err := l.onInsert(ctx, m); err != nil {
			return err
		}
	case *pglogrepl.UpdateMessageV2:
		if err := l.onUpdate(ctx, m); err != nil {
			return err
		}
	case *pglogrepl.DeleteMessageV2:
		if err := l.onDelete(ctx, m); err != nil {
			return err
		}
	case *pglogrepl.StreamStartMessageV2:
		l.inStream = true
	case *pglogrepl.StreamStopMessageV2:
		l.inStream = false
	}
	return nil
}

func (l *logicalReplListener[T]) relMatches(rel *pglogrepl.RelationMessageV2) bool {
	if rel == nil {
		return false
	}
	ns := rel.Namespace
	if ns == "" {
		ns = "public"
	}
	return ns == l.schema && rel.RelationName == l.table
}

func (l *logicalReplListener[T]) onInsert(ctx context.Context, m *pglogrepl.InsertMessageV2) error {
	rel := l.relations[m.RelationID]
	if !l.relMatches(rel) {
		return nil
	}
	ent, err := l.tupleToEntity(rel, m.Tuple)
	if err != nil {
		return err
	}
	l.handler.Add(ctx, ent)
	return nil
}

func (l *logicalReplListener[T]) onUpdate(ctx context.Context, m *pglogrepl.UpdateMessageV2) error {
	rel := l.relations[m.RelationID]
	if !l.relMatches(rel) {
		return nil
	}
	if m.NewTuple == nil {
		return nil
	}
	ent, err := l.tupleToEntity(rel, m.NewTuple)
	if err != nil {
		return err
	}
	id := entityID(ent)
	l.handler.Delete(ctx, id)
	l.handler.Add(ctx, ent)
	return nil
}

func (l *logicalReplListener[T]) onDelete(ctx context.Context, m *pglogrepl.DeleteMessageV2) error {
	rel := l.relations[m.RelationID]
	if !l.relMatches(rel) {
		return nil
	}
	if m.OldTuple == nil {
		return nil
	}
	ent, err := l.tupleToEntity(rel, m.OldTuple)
	if err != nil {
		return err
	}
	l.handler.Delete(ctx, entityID(ent))
	return nil
}

func entityID[T domain.Entity](v T) string {
	return v.ID()
}

func (l *logicalReplListener[T]) tupleToEntity(rel *pglogrepl.RelationMessageV2, tuple *pglogrepl.TupleData) (T, error) {
	var zero T
	if l.decode == nil {
		return zero, errors.New("postgres: logical replication: nil TupleDecoder")
	}
	return l.decode(rel, tuple)
}
