package inmemory

import (
	"context"

	"github.com/dhlab-tech/go-inmemory-platform/pkg/domain"
	"github.com/dhlab-tech/go-inmemory-platform/pkg/projection"
)

type d = domain.Entity

type (
	Cache[T domain.Entity]                         = projection.Cache[T]
	CacheWithEventListener[T domain.Entity]       = projection.CacheWithEventListener[T]
	StreamEventListener[T domain.Entity]          = projection.StreamEventListener[T]
	EventListener[T domain.Entity]                = projection.EventListener[T]
	Notify[T domain.Entity]                       = projection.Notify[T]
	Notifier[T domain.Entity]                     = projection.Notifier[T]
	InverseIndex[T domain.Entity]                 = projection.InverseIndex[T]
	InverseUniqueIndex[T domain.Entity]           = projection.InverseUniqueIndex[T]
	SortedIndex[T domain.Entity]                  = projection.SortedIndex[T]
	SuffixIndex[T domain.Entity]                  = projection.SuffixIndex[T]
	Listener[T domain.Entity]                     = projection.Listener[T]
	AddCallbackListener[T domain.Entity]          = projection.AddCallbackListener[T]
	UpdateCallbackListener[T domain.Entity]       = projection.UpdateCallbackListener[T]
	DeleteCallbackListener[T domain.Entity]       = projection.DeleteCallbackListener[T]
)

const (
	InverseIndexType       = projection.InverseIndexType
	InverseUniqueIndexType = projection.InverseUniqueIndexType
	SortdIndexType         = projection.SortdIndexType
	SuffixIndexType        = projection.SuffixIndexType
)

type (
	M         = projection.M
	Sorted    = projection.Sorted
	Intersect = projection.Intersect
)

func NewCache[T d](data map[string]T) Cache[T] { return projection.NewCache[T](data) }

func NewCacheWithEventListener[T d](
	beforeListeners []StreamEventListener[T],
	afterListeners []StreamEventListener[T],
	notify Notify[T],
) *CacheWithEventListener[T] {
	return projection.NewCacheWithEventListener[T](beforeListeners, afterListeners, notify)
}

func NewListener[T d](cache Cache[T]) *Listener[T] { return projection.NewListener[T](cache) }

func NewNotifier[T d](
	listenersCreate map[string]map[string]func(),
	listenersUpdate map[string]map[string]func(),
	listenersDelete map[string]map[string]func(),
) *Notifier[T] {
	return projection.NewNotifier[T](listenersCreate, listenersUpdate, listenersDelete)
}

func NewAddCallbackListener[T d](callback func(ctx context.Context, v T)) *AddCallbackListener[T] {
	return projection.NewAddCallbackListener[T](callback)
}

func NewUpdateCallbackListener[T d](callback func(ctx context.Context, id string, v T, removedFields []string)) *UpdateCallbackListener[T] {
	return projection.NewUpdateCallbackListener[T](callback)
}

func NewDeleteCallbackListener[T d](callback func(ctx context.Context, id string)) *DeleteCallbackListener[T] {
	return projection.NewDeleteCallbackListener[T](callback)
}

func NewInverseIndex[T d](
	data map[string][]string,
	nilData []string,
	cache Cache[T],
	from []string,
	to *string,
) InverseIndex[T] {
	return projection.NewInverseIndex[T](data, nilData, cache, from, to)
}

func NewInverseUniqIndex[T d](
	data map[string]string,
	cache Cache[T],
	field []string,
	to *string,
) InverseUniqueIndex[T] {
	return projection.NewInverseUniqIndex[T](data, cache, field, to)
}

func NewSortedIndex[T d](sorted Sorted, cache Cache[T], from []string, to *string) SortedIndex[T] {
	return projection.NewSortedIndex[T](sorted, cache, from, to)
}

func NewSuffixIndex[T d](cache Cache[T], btreeDegree int, from []string, to *string) (SuffixIndex[T], SuffixIndex[T]) {
	return projection.NewSuffixIndex[T](cache, btreeDegree, from, to)
}

func NewSuffix[T d](index M, cache Cache[T], from []string, to *string) SuffixIndex[T] {
	return projection.NewSuffix[T](index, cache, from, to)
}

func NewUpdateSuffix[T d](index M, cache Cache[T], from []string, to *string) SuffixIndex[T] {
	return projection.NewUpdateSuffix[T](index, cache, from, to)
}

func NewM[T d](cache Cache[T], tree projection.SuffixTree) M {
	return projection.NewM[T](cache, tree)
}

func NewSorted(degree int, ids []string) Sorted { return projection.NewSorted(degree, ids) }

func BuildSorted() Sorted { return projection.BuildSorted() }

func BuildM[T d](cache Cache[T]) M { return projection.BuildM[T](cache) }

func NewIntersect() *Intersect { return projection.NewIntersect() }
