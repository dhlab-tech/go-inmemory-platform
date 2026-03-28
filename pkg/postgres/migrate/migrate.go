// Package migrate wraps [github.com/pressly/goose] for PostgreSQL with [pgxpool.Pool].
//
// Use an embedded filesystem in the application binary (//go:embed migrations/*.sql) and call [Up]
// when PostgresDeps.AutoMigrate is true (often disabled in production or run under a migration role).
package migrate

import (
	"context"
	"database/sql"
	"io/fs"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// Up runs goose migrations from fsys rooted at dir (e.g. "migrations"). dialect is usually "postgres".
func Up(ctx context.Context, pool *pgxpool.Pool, fsys fs.FS, dir string) error {
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()
	goose.SetBaseFS(fsys)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.UpContext(ctx, db, dir, goose.WithAllowMissing())
}

// UpSQL runs goose on an existing *sql.DB (caller owns lifecycle).
func UpSQL(ctx context.Context, db *sql.DB, fsys fs.FS, dir string) error {
	goose.SetBaseFS(fsys)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.UpContext(ctx, db, dir, goose.WithAllowMissing())
}

// CurrentVersion returns the goose revision recorded in the database (see goose_db_version).
func CurrentVersion(pool *pgxpool.Pool) (int64, error) {
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()
	return goose.GetDBVersion(db)
}
