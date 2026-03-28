# Integrating `go-inmemory-platform` into a service

Single module `github.com/dhlab-tech/go-inmemory-platform`: import only the packages you need. The module lists Mongo and Postgres drivers; your binary only links what you import.

## Package map

| Goal | Import path |
|------|-------------|
| Domain entity constraint | `github.com/dhlab-tech/go-inmemory-platform/pkg/domain` |
| Projection cache, indexes, `Projection[T]` | `github.com/dhlab-tech/go-inmemory-platform/pkg/projection` |
| Mongo `Searcher` / `Processor` / change stream | `github.com/dhlab-tech/go-inmemory-platform/pkg/mongo` |
| Mongo-backed `NewInMemory` | `github.com/dhlab-tech/go-inmemory-platform/pkg/mongo/inmemory` |
| Postgres pool types, replication, `RowStore` | `github.com/dhlab-tech/go-inmemory-platform/pkg/postgres` |
| Goose migrations helper | `github.com/dhlab-tech/go-inmemory-platform/pkg/postgres/migrate` |
| Postgres-backed `NewInMemory` | `github.com/dhlab-tech/go-inmemory-platform/pkg/postgres/inmemory` |
| Entitygen contract (tags) | `github.com/dhlab-tech/go-inmemory-platform/pkg/entitymeta` |

## MongoDB

1. Build `mongo.Stream` from a change stream on your database.
2. Call `mongo/inmemory.NewInMemory` with `MongoDeps` (client, DB name, timeouts) and `Entity` (collection, hooks, warmup filter).

Imports migrated from the old module:

- `github.com/dhlab-tech/go-mongo-platform/pkg/inmemory` → `github.com/dhlab-tech/go-inmemory-platform/pkg/mongo/inmemory`
- `github.com/dhlab-tech/go-mongo-platform/pkg/mongo` → `github.com/dhlab-tech/go-inmemory-platform/pkg/mongo`

## PostgreSQL (codegen)

1. Define structs with `db:"column"` (or `bson:"..."` fallback for column name).
2. Run:

   ```bash
   go run github.com/dhlab-tech/go-inmemory-platform/cmd/entitygen \
     -src=./internal/models/widget.go -types=Widget -table=widgets -out=./gen
   ```

3. Embed generated `migrations/*.sql` and call `postgres/migrate.Up` on startup if you use `AutoMigrate` (or run goose separately in production).
4. Wire `postgres/inmemory.NewInMemory` with `Entity.Store` = `NewXxxStore(pool, schema, table)` and, if logical replication is enabled, `Entity.TupleDecode` = `XxxTupleDecoder()`.

Imports migrated from the old module:

- `github.com/dhlab-tech/go-postgres-platform/pkg/postgres/...` → `github.com/dhlab-tech/go-inmemory-platform/pkg/postgres/...`
- `.../pkg/inmemory` (postgres) → `.../pkg/postgres/inmemory`

## `go.mod` (monorepo)

Use a version tag or a local replace while developing:

```go
require github.com/dhlab-tech/go-inmemory-platform v0.0.0
replace github.com/dhlab-tech/go-inmemory-platform => ../go-inmemory-platform
```

Adjust the relative path to your layout.

## Migrating an existing app (e.g. `server/go`)

[`projection.StreamEventListener`](https://pkg.go.dev/github.com/dhlab-tech/go-inmemory-platform/pkg/projection#StreamEventListener) uses **`Delete(ctx, string)`** and **`Update(ctx, id string, ...)`** (string document IDs). If your indexes or processors still implement `Delete(ctx, primitive.ObjectID)`, update them to take `string` (typically hex IDs) before switching imports from `go-mongo-platform` to this module.

Until that refactor is done, keep depending on **`github.com/dhlab-tech/go-mongo-platform`** as a thin wrapper over `go-inmemory-platform`, or pin a `replace` to a compatible tag.

## Further reading

- [positioning.md](positioning.md) — product positioning vs caches  
- [decision-tree.md](decision-tree.md) — when this module fits  
- [anti-patterns.md](anti-patterns.md) — mistakes to avoid  
- [production-guide.md](production-guide.md) — operations (Mongo + Postgres)  
- [troubleshooting.md](troubleshooting.md) — common failures  
- [ARCHITECTURE_CONTRACT.md](../ARCHITECTURE_CONTRACT.md) — normative guarantees
