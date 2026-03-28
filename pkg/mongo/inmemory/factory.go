package inmemory

import (
	"context"
	"errors"
	"reflect"

	"github.com/dhlab-tech/go-inmemory-platform/pkg/projection"
	"github.com/dhlab-tech/go-inmemory-platform/pkg/mongo"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
)

// noOpHandler is a no-op implementation of mongo handler interface
type noOpHandler[T d] struct{}

func (n *noOpHandler[T]) Add(ctx context.Context, v T) {}

func (n *noOpHandler[T]) Update(ctx context.Context, id string, updatedFields T, removedFields []string) {
}

func (n *noOpHandler[T]) Delete(ctx context.Context, id string) {}

// isStreamValid checks if stream interface is not nil and contains a valid value
func isStreamValid(s stream) bool {
	if s == nil {
		return false
	}
	v := reflect.ValueOf(s)
	if !v.IsValid() {
		return false
	}
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return false
	}
	return true
}

type stream interface {
	AddListener(ctx context.Context, db, col string, listener streamListener)
}

type streamListener = interface {
	Listen(ctx context.Context, change []byte) (err error)
}

// InMemory combines MongoDB access with the shared projection layer (see [projection.Projection]).
type InMemory[T d] interface {
	projection.Projection[T]
	GetMongo() *mongo.Mongo[T]
	// AwaitUpdateDocBSON applies a partial update using MongoDB bson.D fragments, then waits for the projection.
	AwaitUpdateDocBSON(ctx context.Context, id string, set, unset bson.D) (found bool, err error)
	// UpdateDocBSON applies a partial update using bson.D without waiting for the projection.
	UpdateDocBSON(ctx context.Context, id string, set, unset bson.D) (found bool, err error)
}

type inMemory[T d] struct {
	CacheWithEventListener *CacheWithEventListener[T]
	Mongo                  *mongo.Mongo[T]
}

func mapToBsonD(m map[string]any) bson.D {
	if m == nil {
		return nil
	}
	d := make(bson.D, 0, len(m))
	for k, v := range m {
		d = append(d, bson.E{Key: k, Value: v})
	}
	return d
}

// Spawn creates a new instance of the entity type T.
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

// GetCacheWithEventListener returns the cache with event listeners and indexes.
func (im *inMemory[T]) GetCacheWithEventListener() *CacheWithEventListener[T] {
	return im.CacheWithEventListener
}

// GetMongo returns the MongoDB operations instance.
func (im *inMemory[T]) GetMongo() *mongo.Mongo[T] {
	return im.Mongo
}

// AwaitCreate creates an entity in MongoDB and waits until the change is reflected in the in-memory cache.
func (p *inMemory[T]) AwaitCreate(ctx context.Context, ps T) (id string, err error) {
	if p.CacheWithEventListener == nil {
		return "", errors.New("cache is not initialized, AwaitCreate requires cache")
	}
	ch := make(chan struct{})
	defer close(ch)
	ui := p.CacheWithEventListener.AwaitNotify.AddListenerCreate(ps.ID(), func() {
		ch <- struct{}{}
	})
	_, err = p.Mongo.Processor.Create(ctx, ps)
	if err != nil {
		p.CacheWithEventListener.AwaitNotify.DeleteListenerCreate(ps.ID(), ui)
		return
	}
	<-ch
	id = ps.ID()
	return
}

// AwaitUpdate updates an entity in MongoDB and waits until the change is reflected in the in-memory cache.
func (p *inMemory[T]) AwaitUpdate(ctx context.Context, ps T) (res T, err error) {
	if p.CacheWithEventListener == nil {
		return res, errors.New("cache is not initialized, AwaitUpdate requires cache")
	}
	ch := make(chan struct{})
	defer close(ch)
	ui := p.CacheWithEventListener.AwaitNotify.AddListenerUpdate(ps.ID(), func() {
		ch <- struct{}{}
	})
	res, err = p.Mongo.Processor.Update(ctx, ps)
	if err != nil {
		p.CacheWithEventListener.AwaitNotify.DeleteListenerUpdate(ps.ID(), ui)
		if errors.Is(err, mongo.ErrNothingToUpdate) {
			err = nil
		}
		return
	}
	<-ch
	return
}

// AwaitUpdateDoc applies field-level updates using maps (BSON field names), then waits for the projection.
func (p *inMemory[T]) AwaitUpdateDoc(ctx context.Context, id string, set, unset map[string]any) (found bool, err error) {
	return p.AwaitUpdateDocBSON(ctx, id, mapToBsonD(set), mapToBsonD(unset))
}

// UpdateDoc applies field-level updates using maps without waiting for the projection.
func (p *inMemory[T]) UpdateDoc(ctx context.Context, id string, set, unset map[string]any) (found bool, err error) {
	return p.UpdateDocBSON(ctx, id, mapToBsonD(set), mapToBsonD(unset))
}

// UpdateDocBSON applies bson.D fragments without waiting for the projection.
func (p *inMemory[T]) UpdateDocBSON(ctx context.Context, id string, set, unset bson.D) (found bool, err error) {
	found, err = p.Mongo.Updater.UpdateOne(ctx, id, nil, set, unset)
	if err != nil && errors.Is(err, mongo.ErrNothingToUpdate) {
		return found, nil
	}
	return found, err
}

// AwaitUpdateDocBSON updates a document using bson.D fragments and waits until the change is reflected.
func (p *inMemory[T]) AwaitUpdateDocBSON(ctx context.Context, id string, set, unset bson.D) (found bool, err error) {
	if p.CacheWithEventListener == nil {
		return false, errors.New("cache is not initialized, AwaitUpdateDoc requires cache")
	}
	ch := make(chan struct{})
	defer close(ch)
	ui := p.CacheWithEventListener.AwaitNotify.AddListenerUpdate(id, func() {
		ch <- struct{}{}
	})
	found, err = p.Mongo.Updater.UpdateOne(ctx, id, nil, set, unset)
	if err != nil {
		p.CacheWithEventListener.AwaitNotify.DeleteListenerUpdate(id, ui)
		if errors.Is(err, mongo.ErrNothingToUpdate) {
			err = nil
		}
		return
	}
	<-ch
	return
}

// AwaitDelete deletes an entity from MongoDB and waits until the change is reflected in the in-memory cache.
func (p *inMemory[T]) AwaitDelete(ctx context.Context, ps T) (err error) {
	if p.CacheWithEventListener == nil {
		return errors.New("cache is not initialized, AwaitDelete requires cache")
	}
	ch := make(chan struct{})
	defer close(ch)
	ui := p.CacheWithEventListener.AwaitNotify.AddListenerDelete(ps.ID(), func() {
		ch <- struct{}{}
	})
	err = p.Mongo.Processor.Delete(ctx, ps.ID())
	if err != nil {
		p.CacheWithEventListener.AwaitNotify.DeleteListenerDelete(ps.ID(), ui)
		return
	}
	<-ch
	return
}

// Create inserts into MongoDB via [mongo.Processor.Create] without waiting for the in-memory projection.
func (p *inMemory[T]) Create(ctx context.Context, ps T) (id string, err error) {
	return p.Mongo.Processor.Create(ctx, ps)
}

// Update writes to MongoDB via [mongo.Processor.Update] without waiting for the projection.
func (p *inMemory[T]) Update(ctx context.Context, ps T) (T, error) {
	return p.Mongo.Processor.Update(ctx, ps)
}

// Delete removes a document via [mongo.Processor.Delete] without waiting for the projection.
func (p *inMemory[T]) Delete(ctx context.Context, id string) error {
	return p.Mongo.Processor.Delete(ctx, id)
}

// NewInMemory creates a new InMemory instance for a typed entity.
// Returns nil if the collection name is empty (no-op mode).
func NewInMemory[T d](ctx context.Context, stream stream, deps MongoDeps, entityDeps Entity[T]) (InMemory[T], error) {
	if entityDeps.Collection == "" {
		return nil, nil
	}
	var im *CacheWithEventListener[T]
	var m *mongo.Mongo[T]
	if isStreamValid(stream) {
		im = NewCacheWithEventListener[T](
			entityDeps.BeforeListeners,
			entityDeps.AfterListeners,
			entityDeps.Notify,
		)
		m = mongo.NewMongo[T](
			deps.Client,
			deps.Db,
			entityDeps.Collection,
			deps.ConnectionTimeout,
			im.Cache,
			im.EventListener,
		)
	} else {
		m = mongo.NewMongo[T](
			deps.Client,
			deps.Db,
			entityDeps.Collection,
			deps.ConnectionTimeout,
			nil,
			&noOpHandler[T]{},
		)
	}
	if isStreamValid(stream) {
		stream.AddListener(ctx, deps.Db, entityDeps.Collection, m.Listener)
	}
	i := inMemory[T]{
		CacheWithEventListener: im,
		Mongo:                  m,
	}
	if entityDeps.Option != nil {
		entityDeps.Option(&i)
	}
	zerolog.Ctx(ctx).Debug().Str("collection", entityDeps.Collection).Any("im", im).Msg("in-memory initialized")
	if im != nil {
		var (
			its []T
			err error
		)
		if entityDeps.WarmupFilter != nil {
			its, err = m.Searcher.FindWithFilter(ctx, *entityDeps.WarmupFilter)
		} else {
			its, err = m.Searcher.All(ctx)
		}
		if err != nil {
			return nil, err
		}
		for _, it := range its {
			im.EventListener.Add(ctx, it)
		}
	}
	return &i, nil
}
