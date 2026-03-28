package inmemory

import (
	"context"
	"errors"
	"reflect"

	"github.com/dhlab-tech/go-inmemory-platform/pkg/projection"
	"github.com/dhlab-tech/go-inmemory-platform/pkg/postgres"
	"github.com/dhlab-tech/go-inmemory-platform/pkg/postgres/migrate"
	"github.com/rs/zerolog"
)

type noOpHandler[T d] struct{}

func (n *noOpHandler[T]) Add(ctx context.Context, v T) {}

func (n *noOpHandler[T]) Update(ctx context.Context, id string, updatedFields T, removedFields []string) {
}

func (n *noOpHandler[T]) Delete(ctx context.Context, id string) {}

func streamOn(repl *postgres.LogicalReplicationConfig) bool {
	return repl != nil && repl.ConnString != "" && repl.Publication != "" && repl.SlotName != ""
}

// InMemory combines PostgreSQL access with the shared projection layer.
type InMemory[T d] interface {
	projection.Projection[T]
	GetPostgres() *postgres.Postgres[T]
}

type inMemory[T d] struct {
	CacheWithEventListener *CacheWithEventListener[T]
	PG                     *postgres.Postgres[T]
}

func (im *inMemory[T]) GetCacheWithEventListener() *CacheWithEventListener[T] {
	return im.CacheWithEventListener
}

func (im *inMemory[T]) GetPostgres() *postgres.Postgres[T] {
	return im.PG
}

func (im *inMemory[T]) Spawn(ctx context.Context) (instance T) {
	_t := reflect.TypeOf(instance)
	if _t.Kind() == reflect.Ptr {
		instance = reflect.New(_t.Elem()).Interface().(T)
		instance.ID()
		return
	}
	instance = reflect.New(_t).Elem().Interface().(T)
	instance.ID()
	return
}

func (p *inMemory[T]) AwaitCreate(ctx context.Context, ps T) (id string, err error) {
	if p.CacheWithEventListener == nil {
		return "", errors.New("cache is not initialized, AwaitCreate requires cache")
	}
	ch := make(chan struct{})
	defer close(ch)
	ui := p.CacheWithEventListener.AwaitNotify.AddListenerCreate(ps.ID(), func() {
		ch <- struct{}{}
	})
	_, err = p.PG.Store.Create(ctx, ps)
	if err != nil {
		p.CacheWithEventListener.AwaitNotify.DeleteListenerCreate(ps.ID(), ui)
		return
	}
	<-ch
	id = ps.ID()
	return
}

func (p *inMemory[T]) AwaitUpdate(ctx context.Context, ps T) (res T, err error) {
	if p.CacheWithEventListener == nil {
		return res, errors.New("cache is not initialized, AwaitUpdate requires cache")
	}
	ch := make(chan struct{})
	defer close(ch)
	ui := p.CacheWithEventListener.AwaitNotify.AddListenerUpdate(ps.ID(), func() {
		ch <- struct{}{}
	})
	res, err = p.PG.Store.Update(ctx, ps)
	if err != nil {
		p.CacheWithEventListener.AwaitNotify.DeleteListenerUpdate(ps.ID(), ui)
		if errors.Is(err, postgres.ErrNothingToUpdate) {
			err = nil
		}
		return
	}
	<-ch
	return
}

func (p *inMemory[T]) AwaitUpdateDoc(ctx context.Context, id string, set, unset map[string]any) (found bool, err error) {
	if p.CacheWithEventListener == nil {
		return false, errors.New("cache is not initialized, AwaitUpdateDoc requires cache")
	}
	ch := make(chan struct{})
	defer close(ch)
	ui := p.CacheWithEventListener.AwaitNotify.AddListenerUpdate(id, func() {
		ch <- struct{}{}
	})
	found, err = p.PG.Store.UpdateDoc(ctx, id, set, unset)
	if err != nil {
		p.CacheWithEventListener.AwaitNotify.DeleteListenerUpdate(id, ui)
		if errors.Is(err, postgres.ErrNothingToUpdate) {
			err = nil
		}
		return
	}
	<-ch
	return
}

// UpdateDoc applies partial updates via [postgres.RowStore.UpdateDoc] without waiting for the projection.
func (p *inMemory[T]) UpdateDoc(ctx context.Context, id string, set, unset map[string]any) (found bool, err error) {
	found, err = p.PG.Store.UpdateDoc(ctx, id, set, unset)
	if err != nil && errors.Is(err, postgres.ErrNothingToUpdate) {
		return found, nil
	}
	return found, err
}

func (p *inMemory[T]) AwaitDelete(ctx context.Context, ps T) (err error) {
	if p.CacheWithEventListener == nil {
		return errors.New("cache is not initialized, AwaitDelete requires cache")
	}
	ch := make(chan struct{})
	defer close(ch)
	ui := p.CacheWithEventListener.AwaitNotify.AddListenerDelete(ps.ID(), func() {
		ch <- struct{}{}
	})
	err = p.PG.Store.Delete(ctx, ps.ID())
	if err != nil {
		p.CacheWithEventListener.AwaitNotify.DeleteListenerDelete(ps.ID(), ui)
		return
	}
	<-ch
	return
}

// Create inserts via [postgres.RowStore.Create] without waiting for the in-memory projection.
func (p *inMemory[T]) Create(ctx context.Context, ps T) (id string, err error) {
	return p.PG.Store.Create(ctx, ps)
}

// Update writes via [postgres.RowStore.Update] without waiting for the projection.
func (p *inMemory[T]) Update(ctx context.Context, ps T) (T, error) {
	return p.PG.Store.Update(ctx, ps)
}

// Delete removes a row via [postgres.RowStore.Delete] without waiting for the projection.
func (p *inMemory[T]) Delete(ctx context.Context, id string) error {
	return p.PG.Store.Delete(ctx, id)
}

// NewInMemory builds a projection backed by PostgreSQL. Returns nil if Table is empty.
// When Replication is set, TupleDecode must be non-nil and starts [postgres.StreamListener.Listen] in a goroutine.
func NewInMemory[T d](ctx context.Context, deps PostgresDeps, entity Entity[T]) (InMemory[T], error) {
	if entity.Table == "" {
		return nil, nil
	}
	if entity.Store == nil {
		return nil, errors.New("postgres inmemory: Entity.Store is required when Table is set")
	}
	if streamOn(deps.Replication) && entity.TupleDecode == nil {
		return nil, errors.New("postgres inmemory: Entity.TupleDecode is required when Replication is enabled")
	}

	if deps.AutoMigrate && deps.Migrate != nil && deps.Migrate.FS != nil {
		dir := deps.Migrate.Dir
		if dir == "" {
			dir = "migrations"
		}
		if err := migrate.Up(ctx, deps.Pool, deps.Migrate.FS, dir); err != nil {
			return nil, err
		}
		if ver, err := migrate.CurrentVersion(deps.Pool); err == nil {
			zerolog.Ctx(ctx).Info().Int64("goose_db_version", ver).Msg("postgres migrations applied")
		}
	}

	var im *CacheWithEventListener[T]
	var handler projection.StreamEventListener[T]
	if streamOn(deps.Replication) {
		im = NewCacheWithEventListener[T](
			entity.BeforeListeners,
			entity.AfterListeners,
			entity.Notify,
		)
		handler = im.EventListener
	} else {
		handler = &noOpHandler[T]{}
	}

	store := entity.Store
	decode := entity.TupleDecode
	if !streamOn(deps.Replication) {
		decode = nil
	}

	pg := postgres.NewPostgres[T](
		deps.Pool,
		deps.Schema,
		entity.Table,
		store,
		deps.ConnectionTimeout,
		handler,
		deps.Replication,
		decode,
	)

	out := &inMemory[T]{CacheWithEventListener: im, PG: pg}
	if entity.Option != nil {
		entity.Option(out)
	}

	zerolog.Ctx(ctx).Debug().Str("table", entity.Table).Msg("postgres in-memory initialized")

	if im != nil {
		var (
			its []T
			err error
		)
		if entity.WarmupWhere != nil && *entity.WarmupWhere != "" {
			its, err = pg.Store.FindWithFilter(ctx, *entity.WarmupWhere)
		} else {
			its, err = pg.Store.All(ctx)
		}
		if err != nil {
			return nil, err
		}
		for _, it := range its {
			im.EventListener.Add(ctx, it)
		}
	}

	if im != nil && pg.Listener != nil {
		go func() {
			if err := pg.Listener.Listen(ctx); err != nil && !errors.Is(err, context.Canceled) {
				zerolog.Ctx(ctx).Err(err).Msg("postgres logical replication stopped")
			}
		}()
	}

	return out, nil
}
