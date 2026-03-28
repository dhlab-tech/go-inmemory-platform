package inmemory

import (
	"io/fs"
	"time"

	"github.com/dhlab-tech/go-inmemory-platform/pkg/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresDeps holds connection pool and defaults for [NewInMemory].
type PostgresDeps struct {
	Pool              *pgxpool.Pool
	Schema            string
	ConnectionTimeout time.Duration
	// Replication enables logical replication (pgoutput). Nil disables streaming (no-op handler).
	Replication *postgres.LogicalReplicationConfig
	// AutoMigrate runs goose Up from Migrate on startup when true (often disabled in production).
	AutoMigrate bool
	// Migrate is used when AutoMigrate is true; FS is typically //go:embed migrations/*.sql.
	Migrate *MigrateSpec
}

// MigrateSpec points at goose SQL files (embedded or os.DirFS).
type MigrateSpec struct {
	FS  fs.FS
	Dir string // e.g. "migrations"; empty defaults to "migrations"
}

// Entity configures hooks, codegen [postgres.RowStore], optional warmup, and replication decoding.
type Entity[T d] struct {
	Table string
	// Store is required when Table is non-empty (generated [postgres.RowStore]).
	Store postgres.RowStore[T]
	// TupleDecode is required when PostgresDeps.Replication is enabled (e.g. generated XTupleDecoder()).
	TupleDecode postgres.TupleDecoder[T]
	// WarmupWhere is a SQL WHERE fragment (without WHERE), e.g. `deleted IS DISTINCT FROM true`.
	WarmupWhere *string
	BeforeListeners []StreamEventListener[T]
	AfterListeners  []StreamEventListener[T]
	Notify          Notify[T]
	Option          func(InMemory[T])
}
