//go:build integration

package testwidget

import (
	"context"
	"os"
	"testing"

	"github.com/dhlab-tech/go-inmemory-platform/pkg/postgres/migrate"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestMigrateAndWidgetStore(t *testing.T) {
	dsn := os.Getenv("POSTGRES_TEST_URL")
	if dsn == "" {
		t.Skip("POSTGRES_TEST_URL not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := migrate.Up(ctx, pool, MigrationsFS, "migrations"); err != nil {
		t.Fatal(err)
	}
	store := NewWidgetStore(pool, "public", "widgets")
	title := "hello"
	w := &Widget{IDVal: "w1", Title: &title}
	if _, err := store.Create(ctx, w); err != nil {
		t.Fatal(err)
	}
	all, err := store.All(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 || all[0].IDVal != "w1" {
		t.Fatalf("unexpected rows: %+v", all)
	}
}
