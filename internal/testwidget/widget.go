package testwidget

import "github.com/dhlab-tech/go-inmemory-platform/pkg/domain"

//go:generate go run ../../cmd/entitygen -src=widget.go -types=Widget -table=widgets -out=.

// Widget is a minimal domain entity for tests and docs (run entitygen against this file).
var _ domain.Entity = (*Widget)(nil)
type Widget struct {
	IDVal string  `db:"id" bson:"_id"`
	Title *string `db:"title" bson:"title"`
	Ver   *int64  `db:"version" bson:"version"`
	Del   *bool   `db:"deleted" bson:"deleted"`
}

func (w *Widget) ID() string { return w.IDVal }

func (w *Widget) Version() *int64 { return w.Ver }

func (w *Widget) SetDeleted(d bool) {
	b := d
	w.Del = &b
}
