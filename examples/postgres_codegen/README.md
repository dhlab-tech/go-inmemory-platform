# Postgres codegen only

Generate typed `RowStore` and goose SQL from structs with `db` / `bson` tags:

```bash
# from repository root
go run github.com/dhlab-tech/go-inmemory-platform/cmd/entitygen \
  -src=./path/to/models.go -types=MyEntity -table=my_table -out=./gen
```

See [docs/integration.md](../docs/integration.md) and [examples/postgres_widgets](../postgres_widgets/) for a full flow (embed migrations, `migrate.Up`, `postgres/inmemory`).
