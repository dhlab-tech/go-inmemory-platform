package domain

// Entity is the minimal contract for types stored in the projection layer and
// synchronized from durable storage (MongoDB, PostgreSQL, etc.).
type Entity interface {
	any
	ID() string
	Version() *int64
	SetDeleted(bool)
}
