//go:build integration

package testwidget

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/dhlab-tech/go-inmemory-platform/pkg/postgres/migrate"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BenchmarkWidgetStore_UpdateDoc exercises generated [WidgetStore.UpdateDoc] (typed switch, no reflection on SQL).
func BenchmarkWidgetStore_UpdateDoc(b *testing.B) {
	dsn := os.Getenv("POSTGRES_TEST_URL")
	if dsn == "" {
		b.Skip("POSTGRES_TEST_URL not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(pool.Close)
	if err := migrate.Up(ctx, pool, MigrationsFS, "migrations"); err != nil {
		b.Fatal(err)
	}
	store := NewWidgetStore(pool, "public", "widgets")
	const base = "bench-widget-updatedoc"
	_, _ = pool.Exec(ctx, `DELETE FROM widgets WHERE id LIKE $1`, base+"%")
	id := base + "-0"
	title := "t0"
	w := &Widget{IDVal: id, Title: &title}
	if _, err := store.Create(ctx, w); err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM widgets WHERE id = $1`, id)
	})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t := fmt.Sprintf("title-%d", i)
		_, err := store.UpdateDoc(ctx, id, map[string]any{"title": t}, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkWidgetStore_Update exercises generated full-row [WidgetStore.Update].
func BenchmarkWidgetStore_Update(b *testing.B) {
	dsn := os.Getenv("POSTGRES_TEST_URL")
	if dsn == "" {
		b.Skip("POSTGRES_TEST_URL not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(pool.Close)
	if err := migrate.Up(ctx, pool, MigrationsFS, "migrations"); err != nil {
		b.Fatal(err)
	}
	store := NewWidgetStore(pool, "public", "widgets")
	const base = "bench-widget-update"
	_, _ = pool.Exec(ctx, `DELETE FROM widgets WHERE id LIKE $1`, base+"%")
	id := base + "-0"
	title := "t0"
	w := &Widget{IDVal: id, Title: &title}
	if _, err := store.Create(ctx, w); err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM widgets WHERE id = $1`, id)
	})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t := fmt.Sprintf("title-%d", i)
		w := &Widget{IDVal: id, Title: &t}
		_, err := store.Update(ctx, w)
		if err != nil {
			b.Fatal(err)
		}
	}
}
