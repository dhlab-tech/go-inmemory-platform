package entitygen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerate_SampleSQL(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		Src:    filepath.Join("testdata", "sample.go"),
		Types:  []string{"Sample"},
		OutDir: dir,
		Table:  "samples",
		PK:     "id",
		Seq:    "00001",
	}
	if err := Generate(cfg); err != nil {
		t.Fatal(err)
	}
	sqlPath := filepath.Join(dir, "migrations", "00001_samples_init.up.sql")
	raw, err := os.ReadFile(sqlPath)
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	if !strings.Contains(s, `CREATE TABLE IF NOT EXISTS "samples"`) {
		t.Fatalf("missing CREATE: %s", s)
	}
	if !strings.Contains(s, `REPLICA IDENTITY FULL`) {
		t.Fatalf("missing replica identity: %s", s)
	}
	goSrc, err := os.ReadFile(filepath.Join(dir, "sample_store.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(goSrc), "func NewSampleStore(") {
		t.Fatal("missing constructor")
	}
	if !strings.Contains(string(goSrc), "SampleTupleDecoder()") {
		t.Fatal("missing tuple decoder")
	}
}
