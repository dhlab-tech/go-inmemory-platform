package fixtures

// Sample is used by entitygen parser tests.
type Sample struct {
	ID    string  `db:"id"`
	Title *string `db:"title" bson:"ttl"`
	Ver   int64   `db:"version"`
}
