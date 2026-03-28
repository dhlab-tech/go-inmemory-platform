package projection

import (
	"context"
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BenchmarkPrepareIdxs_DocSetTitle walks struct tags and nested fields (reflection).
func BenchmarkPrepareIdxs_DocSetTitle(b *testing.B) {
	var doc DocSetTitle
	v := reflect.ValueOf(doc)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = prepareIdxs(v)
	}
}

// BenchmarkPrepareIdxs_Image is a larger shape (nested indexes on embedded D + flat fields).
func BenchmarkPrepareIdxs_Image(b *testing.B) {
	var image Image
	v := reflect.ValueOf(image)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = prepareIdxs(v)
	}
}

// BenchmarkNewCacheWithEventListener_DocSetTitle builds listeners + all index types from entity tags (reflection-heavy startup).
func BenchmarkNewCacheWithEventListener_DocSetTitle(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewCacheWithEventListener[*DocSetTitle](nil, nil, nil)
	}
}

// BenchmarkCache_prepareCreate_Image mirrors the in-memory cache path that clones entities from BSON/stream payloads (reflection).
func BenchmarkCache_prepareCreate_Image(b *testing.B) {
	ctx := context.Background()
	c := NewCache[*Image](map[string]*Image{}).(*cache[*Image])
	_id := primitive.NewObjectID()
	v := int64(1)
	del := false
	w, h := 640, 480
	name := "bench.png"
	img := &Image{
		D:      D{Id: _id, V: &v, Deleted: &del},
		Name:   &name,
		Width:  &w,
		Height: &h,
	}
	ps := reflect.ValueOf(img)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.prepareCreate(ctx, ps)
	}
}

// BenchmarkUpdateStringFieldValuesByName_inverseKey is a hot path for composite index keys (reflection + string concat).
func BenchmarkUpdateStringFieldValuesByName_inverseKey(b *testing.B) {
	_id := primitive.NewObjectID()
	v := int64(1)
	del := false
	doc := &DocSetTitle{
		D:         D{Id: _id, V: &v, Deleted: &del},
		CatalogID: ptrString("c1"),
		ItemID:    ptrString("i1"),
		Title:     ptrString("t"),
	}
	fields := []string{"CatalogID", "ItemID"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = updateStringFieldValuesByName(doc, fields)
	}
}

func ptrString(s string) *string { return &s }
