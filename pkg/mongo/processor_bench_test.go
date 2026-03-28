package mongo_test

import (
	"context"
	"testing"

	"github.com/dhlab-tech/go-inmemory-platform/pkg/mongo"
	"github.com/dhlab-tech/go-inmemory-platform/pkg/projection"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BenchmarkProcessor_PrepareCreate_D measures BSON document preparation for a minimal pointer entity (reflection).
func BenchmarkProcessor_PrepareCreate_D(b *testing.B) {
	ctx := context.Background()
	cache := projection.NewCache(map[string]*D{})
	p := mongo.NewProcessor[*D](cache, nil, nil, nil)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := primitive.NewObjectID()
		v := int64(1)
		del := false
		ps := &D{Id: id, V: &v, Deleted: &del}
		_, _, _ = p.PrepareCreate(ctx, ps)
	}
}

// BenchmarkProcessor_PrepareCreate_Config measures a wider struct (maps, slices, nested values).
func BenchmarkProcessor_PrepareCreate_Config(b *testing.B) {
	ctx := context.Background()
	cache := projection.NewCache(map[string]*Config{})
	p := mongo.NewProcessor[*Config](cache, nil, nil, nil)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg := newBenchConfig()
		_, _, _ = p.PrepareCreate(ctx, cfg)
	}
}

// BenchmarkProcessor_PrepareUpdate_D compares new vs cached entity and builds $set fields (reflection).
func BenchmarkProcessor_PrepareUpdate_D(b *testing.B) {
	ctx := context.Background()
	id := primitive.NewObjectID()
	v := int64(1)
	v2 := int64(2)
	del := false
	old := &D{Id: id, V: &v, Deleted: &del}
	cache := projection.NewCache(map[string]*D{old.ID(): old})
	p := mongo.NewProcessor[*D](cache, nil, nil, nil)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		next := &D{Id: id, V: &v2, Deleted: &del}
		_, _, _, _ = p.PrepareUpdate(ctx, next)
	}
}

// BenchmarkProcessor_PrepareUpdate_Config is a heavier diff path on the Config shape.
func BenchmarkProcessor_PrepareUpdate_Config(b *testing.B) {
	ctx := context.Background()
	cfgOld := newBenchConfig()
	cache := projection.NewCache(map[string]*Config{cfgOld.ID(): cfgOld})
	p := mongo.NewProcessor[*Config](cache, nil, nil, nil)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfgNew := cloneBenchConfig(cfgOld)
		cfgNew.Labels[0] = "mutated-label"
		_, _, _, _ = p.PrepareUpdate(ctx, cfgNew)
	}
}

func newBenchConfig() *Config {
	id := primitive.NewObjectID()
	v := int64(1)
	del := false
	schema := "s"
	host := "h"
	path := map[string]string{"a": "1", "b": "2"}
	return &Config{
		D:        D{Id: id, V: &v, Deleted: &del},
		Schema:   &schema,
		Host:     &host,
		Path:     path,
		PageSize: ptrInt(10),
		Labels:   []string{"l1", "l2"},
		Fits:     []Fit{{Width: width1, Height: height1}},
	}
}

func cloneBenchConfig(c *Config) *Config {
	cp := *c
	cp.Path = map[string]string{}
	for k, v := range c.Path {
		cp.Path[k] = v
	}
	cp.JSVars = nil
	cp.Menu = nil
	cp.Fits = append([]Fit(nil), c.Fits...)
	cp.Labels = append([]string(nil), c.Labels...)
	return &cp
}

func ptrInt(n int) *int { return &n }
