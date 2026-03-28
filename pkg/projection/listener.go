package projection

import (
	"context"

	"github.com/dhlab-tech/go-inmemory-platform/pkg/domain"
)

// StreamEventListener handles change events for typed entities.
// Implementations receive notifications when entities are added, updated, or deleted.
type StreamEventListener[T domain.Entity] interface {
	Add(ctx context.Context, v T)
	Update(ctx context.Context, id string, updatedFields T, removedFields []string)
	Delete(ctx context.Context, id string)
}

// Listener coordinates multiple StreamEventListeners and manages the execution order.
// BeforeListeners are called before cache operations, regular listeners are called after.
type Listener[T domain.Entity] struct {
	cache           Cache[T]
	listeners       []StreamEventListener[T]
	beforeListeners []StreamEventListener[T]
}

// Add processes an Add event by calling before listeners, updating the cache, then calling after listeners.
func (c *Listener[T]) Add(ctx context.Context, v T) {
	for _, listener := range c.beforeListeners {
		listener.Add(ctx, v)
	}
	c.cache.Add(ctx, v)
	for _, listener := range c.listeners {
		listener.Add(ctx, v)
	}
}

// Update processes an Update event by calling before listeners, updating the cache, then calling after listeners.
func (c *Listener[T]) Update(ctx context.Context, id string, updatedFields T, removedFields []string) {
	for _, listener := range c.beforeListeners {
		listener.Update(ctx, id, updatedFields, removedFields)
	}
	c.cache.Update(ctx, id, updatedFields, removedFields)
	for _, listener := range c.listeners {
		listener.Update(ctx, id, updatedFields, removedFields)
	}
}

// Delete processes a Delete event by calling before listeners, deleting from the cache, then calling after listeners.
func (c *Listener[T]) Delete(ctx context.Context, id string) {
	for _, listener := range c.beforeListeners {
		listener.Delete(ctx, id)
	}
	c.cache.Delete(ctx, id)
	for _, listener := range c.listeners {
		listener.Delete(ctx, id)
	}
}

// AddListener registers a new StreamEventListener.
// If before is true, the listener is called before cache operations; otherwise, it's called after.
func (c *Listener[T]) AddListener(listener StreamEventListener[T], before bool) (idx int) {
	if before {
		c.beforeListeners = append(c.beforeListeners, listener)
		return
	}
	c.listeners = append(c.listeners, listener)
	return
}

// NewListener creates a new Listener that coordinates cache operations and event listeners.
func NewListener[T domain.Entity](cache Cache[T]) *Listener[T] {
	return &Listener[T]{
		cache:           cache,
		listeners:       []StreamEventListener[T]{},
		beforeListeners: []StreamEventListener[T]{},
	}
}

// AddCallbackListener is a StreamEventListener that calls a callback function only for Add events.
type AddCallbackListener[T domain.Entity] struct {
	callback func(ctx context.Context, v T)
}

// NewAddCallbackListener creates a new AddCallbackListener with the specified callback.
func NewAddCallbackListener[T domain.Entity](callback func(ctx context.Context, v T)) *AddCallbackListener[T] {
	return &AddCallbackListener[T]{
		callback: callback,
	}
}

func (s *AddCallbackListener[T]) Add(ctx context.Context, v T) {
	s.callback(ctx, v)
}

func (s *AddCallbackListener[T]) Update(ctx context.Context, id string, updatedFields T, removedFields []string) {
}

func (s *AddCallbackListener[T]) Delete(ctx context.Context, id string) {

}

// UpdateCallbackListener is a StreamEventListener that calls a callback function only for Update events.
type UpdateCallbackListener[T domain.Entity] struct {
	callback func(ctx context.Context, id string, v T, removedFields []string)
}

// NewUpdateCallbackListener creates a new UpdateCallbackListener with the specified callback.
func NewUpdateCallbackListener[T domain.Entity](callback func(ctx context.Context, id string, v T, removedFields []string)) *UpdateCallbackListener[T] {
	return &UpdateCallbackListener[T]{
		callback: callback,
	}
}

func (s *UpdateCallbackListener[T]) Add(ctx context.Context, v T) {
}

func (s *UpdateCallbackListener[T]) Update(ctx context.Context, id string, updatedFields T, removedFields []string) {
	s.callback(ctx, id, updatedFields, removedFields)
}

func (s *UpdateCallbackListener[T]) Delete(ctx context.Context, id string) {
}

// DeleteCallbackListener is a StreamEventListener that calls a callback function only for Delete events.
type DeleteCallbackListener[T domain.Entity] struct {
	callback func(ctx context.Context, id string)
}

// NewDeleteCallbackListener creates a new DeleteCallbackListener with the specified callback.
func NewDeleteCallbackListener[T domain.Entity](callback func(ctx context.Context, id string)) *DeleteCallbackListener[T] {
	return &DeleteCallbackListener[T]{
		callback: callback,
	}
}

func (s *DeleteCallbackListener[T]) Add(ctx context.Context, v T) {
}

func (s *DeleteCallbackListener[T]) Update(ctx context.Context, id string, updatedFields T, removedFields []string) {
}

func (s *DeleteCallbackListener[T]) Delete(ctx context.Context, id string) {
	s.callback(ctx, id)
}
