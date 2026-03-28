// Entitygen reads explicitly listed struct types from a Go source file and emits a typed
// [github.com/dhlab-tech/go-inmemory-platform/pkg/postgres.RowStore], goose SQL, and a [postgres.TupleDecoder].
//
// Example:
//
//	go run github.com/dhlab-tech/go-inmemory-platform/cmd/entitygen \
//	  -src=./internal/models/widget.go -types=Widget -table=widgets -out=./gen
package main

import (
	"flag"
	"log"
	"strings"

	"github.com/dhlab-tech/go-inmemory-platform/internal/entitygen"
)

func main() {
	src := flag.String("src", "", "path to .go file containing entity struct(s)")
	types := flag.String("types", "", "comma-separated struct type names to generate")
	out := flag.String("out", ".", "output directory (store .go + migrations/)")
	table := flag.String("table", "", "PostgreSQL table name")
	pk := flag.String("pk", "id", "primary key column (must match a db/bson-tagged field)")
	seq := flag.String("seq", "00001", "numeric prefix for goose migration filename")
	pkg := flag.String("pkg", "", "package name for generated Go (default: source file package)")
	flag.Parse()

	var typeList []string
	for _, p := range strings.Split(*types, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			typeList = append(typeList, p)
		}
	}
	cfg := entitygen.Config{
		Src:     *src,
		Types:   typeList,
		OutDir:  *out,
		Table:   *table,
		PK:      *pk,
		Seq:     *seq,
		PkgName: *pkg,
	}
	if err := entitygen.Generate(cfg); err != nil {
		log.Fatal(err)
	}
}
