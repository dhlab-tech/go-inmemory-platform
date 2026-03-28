package main

import "github.com/dhlab-tech/go-inmemory-platform/pkg/domain"

//go:generate go run ../../cmd/entitygen -src=model.go -types=Widget -table=widgets -out=.

// Widget is a minimal entity for the example; regenerate store with:
//
//	go run github.com/dhlab-tech/go-inmemory-platform/cmd/entitygen -src=model.go -types=Widget -table=widgets -out=.
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

var _ domain.Entity = (*Widget)(nil)
