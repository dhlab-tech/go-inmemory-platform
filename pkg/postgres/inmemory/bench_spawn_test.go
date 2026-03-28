package inmemory

import (
	"context"
	"testing"
)

// benchPGEntity is a minimal [domain.Entity] for Spawn benchmarks.
type benchPGEntity struct {
	IDVal string
}

func (e *benchPGEntity) ID() string {
	if e.IDVal == "" {
		e.IDVal = "bench-id"
	}
	return e.IDVal
}

func (e *benchPGEntity) Version() *int64 { return nil }

func (e *benchPGEntity) SetDeleted(bool) {}

// BenchmarkSpawn_pointerEntity measures reflect-based allocation in Spawn (pointer entity).
func BenchmarkSpawn_pointerEntity(b *testing.B) {
	ctx := context.Background()
	im := &inMemory[*benchPGEntity]{}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = im.Spawn(ctx)
	}
}

type benchPGValueEntity struct {
	IDVal string
}

func (e benchPGValueEntity) ID() string {
	if e.IDVal == "" {
		return "bench"
	}
	return e.IDVal
}

func (e benchPGValueEntity) Version() *int64 { return nil }

func (e benchPGValueEntity) SetDeleted(bool) {}

// BenchmarkSpawn_valueEntity measures Spawn when T is a struct.
func BenchmarkSpawn_valueEntity(b *testing.B) {
	ctx := context.Background()
	im := &inMemory[benchPGValueEntity]{}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = im.Spawn(ctx)
	}
}
