package inmemory

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// benchMongoEntity satisfies mongo entity constraints for Spawn benchmarks.
type benchMongoEntity struct {
	Id      primitive.ObjectID `bson:"_id"`
	V       *int64             `bson:"version"`
	Deleted *bool              `bson:"deleted"`
}

func (e *benchMongoEntity) ID() string {
	if e.Id.IsZero() {
		e.Id = primitive.NewObjectID()
	}
	return e.Id.Hex()
}

func (e *benchMongoEntity) Version() *int64 { return e.V }

func (e *benchMongoEntity) SetDeleted(d bool) {
	b := d
	e.Deleted = &b
}

// BenchmarkSpawn_pointerEntity measures reflect-based allocation in Spawn (pointer entity).
func BenchmarkSpawn_pointerEntity(b *testing.B) {
	ctx := context.Background()
	im := &inMemory[*benchMongoEntity]{}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = im.Spawn(ctx)
	}
}

// benchMongoValueEntity is a non-pointer entity; Spawn uses reflect.New for value types.
type benchMongoValueEntity struct {
	Id      primitive.ObjectID `bson:"_id"`
	V       *int64             `bson:"version"`
	Deleted *bool              `bson:"deleted"`
}

func (e benchMongoValueEntity) ID() string {
	// value receiver: ID() on zero value allocates new ObjectID each call — not used for bench entity lifecycle
	if e.Id.IsZero() {
		return primitive.NewObjectID().Hex()
	}
	return e.Id.Hex()
}

func (e benchMongoValueEntity) Version() *int64 { return e.V }

func (e benchMongoValueEntity) SetDeleted(d bool) {}

// BenchmarkSpawn_valueEntity measures Spawn when T is a struct (not a pointer type).
func BenchmarkSpawn_valueEntity(b *testing.B) {
	ctx := context.Background()
	im := &inMemory[benchMongoValueEntity]{}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = im.Spawn(ctx)
	}
}
