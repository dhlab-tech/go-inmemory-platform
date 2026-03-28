package projection

import (
	"context"

	"github.com/dhlab-tech/go-inmemory-platform/pkg/domain"
)

// Projection is the storage-agnostic API for typed entities with an in-memory
// projection and read-after-write Await* semantics.
type Projection[T domain.Entity] interface {
	GetCacheWithEventListener() *CacheWithEventListener[T]
	Spawn(ctx context.Context) T
	AwaitCreate(ctx context.Context, ps T) (id string, err error)
	AwaitUpdate(ctx context.Context, ps T) (res T, err error)
	// AwaitUpdateDoc applies a partial document update then waits until the projection reflects it.
	// set/unset use BSON field names (e.g. "title") to match struct bson tags on T.
	AwaitUpdateDoc(ctx context.Context, id string, set, unset map[string]any) (found bool, err error)
	AwaitDelete(ctx context.Context, ps T) (err error)
	// Create, Update, and Delete persist through the backing store (e.g. mongo.Processor or postgres.RowStore)
	// without blocking until the in-memory projection reflects the change. Use Await* when you need read-after-write consistency.
	Create(ctx context.Context, ps T) (id string, err error)
	Update(ctx context.Context, ps T) (T, error)
	Delete(ctx context.Context, id string) (err error)
	// UpdateDoc applies the same partial update as AwaitUpdateDoc without waiting for the in-memory projection.
	UpdateDoc(ctx context.Context, id string, set, unset map[string]any) (found bool, err error)
}
