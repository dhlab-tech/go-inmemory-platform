package postgres

import (
	"time"

	"github.com/dhlab-tech/go-inmemory-platform/pkg/domain"
	"github.com/dhlab-tech/go-inmemory-platform/pkg/projection"
	"github.com/jackc/pgx/v5/pgxpool"
)

type d = domain.Entity

// Postgres aggregates PostgreSQL operations for one projection table using a [RowStore] (codegen).
type Postgres[T d] struct {
	Store    RowStore[T]
	Listener StreamListener // logical replication consumer; nil if replication disabled
}

// NewPostgres wires a typed [RowStore], optional logical replication, and a tuple decoder.
// When repl is non-nil, decode must be non-nil (typically generated).
func NewPostgres[T d](
	pool *pgxpool.Pool,
	schema string,
	table string,
	store RowStore[T],
	connectionTimeout time.Duration,
	handler projection.StreamEventListener[T],
	repl *LogicalReplicationConfig,
	decode TupleDecoder[T],
) *Postgres[T] {
	_ = connectionTimeout // reserved for future statement timeouts
	_ = pool
	sch := schema
	if sch == "" {
		sch = "public"
	}

	var listener StreamListener
	if repl != nil && repl.ConnString != "" && repl.Publication != "" && repl.SlotName != "" && decode != nil {
		listener = NewLogicalReplicationListener[T](*repl, sch, table, handler, decode)
	}

	return &Postgres[T]{
		Store:    store,
		Listener: listener,
	}
}
