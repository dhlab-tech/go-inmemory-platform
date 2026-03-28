package postgres

import "errors"

var (
	// ErrNothingToUpdate is returned when an update would not change the row.
	ErrNothingToUpdate = errors.New("nothing to update")
	// ErrNotFound is returned when a row does not exist.
	ErrNotFound = errors.New("not found")
)
