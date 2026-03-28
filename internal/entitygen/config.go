package entitygen

// Config drives one entitygen run (single struct → store + goose migration).
type Config struct {
	Src     string // path to .go source file
	Types   []string
	OutDir  string
	Table   string
	PK      string // primary key column name, default "id"
	Seq     string // migration numeric prefix e.g. "00001"
	PkgName string // if empty, taken from source file package
}
