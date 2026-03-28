package main

import (
	"context"
	"embed"
	"log"
	"os"

	"github.com/dhlab-tech/go-inmemory-platform/pkg/postgres/inmemory"
	"github.com/dhlab-tech/go-inmemory-platform/pkg/postgres/migrate"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	ctx := logger.WithContext(context.Background())

	dsn := "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	if v := os.Getenv("POSTGRES_DSN"); v != "" {
		dsn = v
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("pgx pool: %v", err)
	}
	defer pool.Close()

	if err := migrate.Up(ctx, pool, migrationsFS, "migrations"); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	logger.Info().Msg("migrations applied")

	store := NewWidgetStore(pool, "public", "widgets")
	title := "demo"
	w := &Widget{IDVal: "ex1", Title: &title}
	if _, err := store.Create(ctx, w); err != nil {
		log.Fatalf("create: %v", err)
	}

	im, err := inmemory.NewInMemory[*Widget](ctx, inmemory.PostgresDeps{
		Pool:        pool,
		Schema:      "public",
		AutoMigrate: false,
	}, inmemory.Entity[*Widget]{
		Table:   "widgets",
		Store:   store,
		// TupleDecode only required when PostgresDeps.Replication is set (logical replication).
	})
	if err != nil {
		log.Fatalf("inmemory: %v", err)
	}
	if im == nil {
		log.Fatal("unexpected nil InMemory")
	}

	all, err := im.GetPostgres().Store.All(ctx)
	if err != nil {
		log.Fatalf("all: %v", err)
	}
	logger.Info().Int("rows", len(all)).Msg("projection store read ok")
}
