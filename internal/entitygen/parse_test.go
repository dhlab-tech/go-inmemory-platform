package entitygen

import (
	"path/filepath"
	"testing"
)

func TestParseFile_Sample(t *testing.T) {
	src := filepath.Join("testdata", "sample.go")
	meta, err := ParseFile(src, "Sample")
	if err != nil {
		t.Fatal(err)
	}
	if meta.Package != "fixtures" {
		t.Fatalf("package: got %q", meta.Package)
	}
	if len(meta.Fields) != 3 {
		t.Fatalf("fields: got %d", len(meta.Fields))
	}
	if meta.Fields[0].Column != "id" || meta.Fields[0].Kind != KindString || meta.Fields[0].Nullable {
		t.Fatalf("id field: %+v", meta.Fields[0])
	}
	if meta.Fields[1].Column != "title" || meta.Fields[1].UpdateKey != "ttl" {
		t.Fatalf("title field: %+v", meta.Fields[1])
	}
	if meta.Fields[2].Column != "version" || meta.Fields[2].Kind != KindInt64 || meta.Fields[2].Nullable {
		t.Fatalf("version field: %+v", meta.Fields[2])
	}
}
