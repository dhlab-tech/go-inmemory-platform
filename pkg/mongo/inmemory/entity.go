package inmemory

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	mng "go.mongodb.org/mongo-driver/mongo"
)

// MongoDeps contains MongoDB connection dependencies required for creating an InMemory instance.
type MongoDeps struct {
	Client            *mng.Client
	Db                string
	ConnectionTimeout time.Duration
}

// Entity configures an entity type with its collection name, listeners, and options.
// BeforeListeners are called before cache operations, AfterListeners are called after.
// Notify is used for event notifications, and Option allows customizing the InMemory instance.
//
// WarmupFilter, if non-nil, restricts the initial full sync (Searcher.FindWithFilter).
// Use e.g. bson.M{"deleted": bson.M{"$ne": true}} to skip soft-deleted documents and shorten startup.
// Nil means the entire collection is loaded (same as before).
type Entity[T d] struct {
	Collection      string
	WarmupFilter    *bson.M
	BeforeListeners []StreamEventListener[T]
	AfterListeners  []StreamEventListener[T]
	Notify          Notify[T]
	Option          func(InMemory[T])
}
