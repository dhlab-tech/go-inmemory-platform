package testwidget

import "embed"

// MigrationsFS is example embedded goose SQL for tests and CI.
//
//go:embed migrations/*.sql
var MigrationsFS embed.FS
