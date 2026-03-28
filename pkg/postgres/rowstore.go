package postgres

import (
	"context"

	"github.com/dhlab-tech/go-inmemory-platform/pkg/domain"
	"github.com/jackc/pglogrepl"
)

// RowStore is implemented by cmd/entitygen output: typed columns, explicit Scan and arguments (no reflection on hot paths).
type RowStore[T domain.Entity] interface {
	All(ctx context.Context) ([]T, error)
	// FindWithFilter applies a trusted SQL fragment after WHERE (no "WHERE" keyword).
	FindWithFilter(ctx context.Context, whereSQL string) ([]T, error)
	Create(ctx context.Context, v T) (id string, err error)
	Update(ctx context.Context, v T) (T, error)
	Delete(ctx context.Context, id string) error
	// UpdateDoc applies partial updates; keys are BSON/json field names matching struct tags used at codegen time.
	UpdateDoc(ctx context.Context, id string, set, unset map[string]any) (found bool, err error)
}

// TupleDecoder maps a logical-replication tuple to T (typically generated; may use [DecodeReplicationCell]).
type TupleDecoder[T any] func(rel *pglogrepl.RelationMessageV2, tuple *pglogrepl.TupleData) (T, error)
