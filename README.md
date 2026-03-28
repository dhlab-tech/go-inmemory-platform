# go-inmemory-platform

A single Go module that keeps an **in-process typed cache** (projection) in sync with **MongoDB** (Change Streams) or **PostgreSQL** (typed `RowStore`, optional logical replication). The storage-facing surface is [`projection.Projection[T]`](pkg/projection/projection.go): **one API** for both backends.

**What this is:** A strongly consistent **projection** inside your process—indexes, listeners, and `Await*` read-after-write semantics—fed from the database you already use. Postgres paths pair **codegen** (`cmd/entitygen`) with explicit SQL and hot paths without reflection on queries.

**What this is not:** Not a distributed cache, not a Redis replacement, not shared RAM across replicas. No built-in TTL eviction or cross-service invalidation. See [NON_GOALS.md](NON_GOALS.md).

Use it when you want **fast reads from memory** while the **database stays the source of truth**, and you can accept **per-process** cache semantics.

## When to use / when not to

**Good fit**

- Go service with **MongoDB** (replica set / Change Streams) **and/or** **PostgreSQL** with a typed store.
- Read-heavy paths with **in-memory indexes** (inverse, sorted, suffix, …).
- You want **`Await*`** for deterministic read-after-write **within one process**.
- You are migrating from separate **`go-mongo-platform`** / **`go-postgres-platform`** imports into **one module**.

**Poor fit**

- **Shared mutable cache** across all instances → use Redis or a dedicated data service.
- **TTL-driven** expiry as a product feature without app logic → use a cache with TTL.
- **Polyglot** clients to the same in-memory layer → not applicable (Go-only, in-process).
- **Beginner-only** teams unfamiliar with Change Streams or Postgres replication—operational complexity remains yours.

For normative boundaries, read [ARCHITECTURE_CONTRACT.md](ARCHITECTURE_CONTRACT.md).

## Why not “just add Redis”?

Classic **DB + external cache** stacks often add:

- dual-write or invalidation bugs,
- races between DB and cache,
- stale reads and hard-to-reason visibility.

**go-inmemory-platform** instead keeps **one write path** to Mongo or Postgres and applies **ordered** updates to the projection (streams / replication). **`Await*`** ties your goroutine to that path when you need it. Reads from the projection avoid extra network hops—but **each replica still has its own RAM** ([ARCHITECTURE_CONTRACT.md](ARCHITECTURE_CONTRACT.md)).

## What’s in the box

- **`pkg/domain`** — minimal entity contract: `ID()`, `Version()`, `SetDeleted(bool)`.
- **`pkg/projection`** — in-memory `Cache`, tag-driven **indexes** (inverse, inverse-unique, sorted, suffix), listener chain, and **`Projection[T]`**.
- **`pkg/mongo`** — `Searcher`, `Processor` (reflective BSON prepare for create/update), `Updater` (partial `$set`/`$unset`), change-stream **listener** wiring.
- **`pkg/mongo/inmemory`** — `NewInMemory`: wires Mongo + projection; optional warmup from DB; `AwaitUpdateDoc` / `UpdateDoc` (maps keyed by BSON field names) plus `AwaitUpdateDocBSON` / `UpdateDocBSON` for raw `bson.D`.
- **`pkg/postgres`** — typed **`RowStore[T]`** (no reflection on hot SQL paths when using generated stores), **`TupleDecoder`** for pgoutput, optional **logical replication** listener.
- **`pkg/postgres/migrate`** — [pressly/goose](https://github.com/pressly/goose) on `pgxpool`.
- **`pkg/postgres/inmemory`** — `NewInMemory`: wires generated `RowStore` + projection + optional replication.
- **`cmd/entitygen`** — generates `RowStore`, goose SQL, and `TupleDecoder` from structs (`db:` / `bson:` tags); contract in **`pkg/entitymeta`**.

## `Projection[T]` at a glance

| Direction | Blocking until cache matches DB |
|-----------|----------------------------------|
| Create / full entity update / delete | `AwaitCreate`, `AwaitUpdate`, `AwaitDelete` |
| Partial update by field maps (`set` / `unset`) | `AwaitUpdateDoc` (Postgres + Mongo) |
| Same writes **without** waiting on the projection | `Create`, `Update`, `Delete`, `UpdateDoc` |

`Spawn` allocates a zero entity of type `T` (with `ID()` initialized where applicable). Indexes are built from struct field tags (`indexes:"..."`); see tests under `pkg/projection` for shapes with nested structs.

## Layout

```
pkg/domain/
pkg/projection/          # Cache, indexes, listeners, Projection[T]
pkg/mongo/               # Searcher, Processor, Stream, Updater, …
pkg/mongo/inmemory/      # NewInMemory (Mongo + projection)
pkg/postgres/            # RowStore, replication, errors, …
pkg/postgres/migrate/    # goose + pgx
pkg/postgres/inmemory/   # NewInMemory (Postgres + projection)
pkg/entitymeta/          # Tags / docs for entitygen
cmd/entitygen/
internal/entitygen/      # Generator + tests
internal/testwidget/     # Sample entity, embed migrations, integration test
examples/                # Runnable Mongo and Postgres programs
docs/integration.md      # Wiring into a real service (replace, replication, IDs)
```

## Quick start

**MongoDB** (replica set required for change streams):

```bash
export MONGODB_URI='mongodb://.../?replicaSet=rs0'
go run ./examples/mongo_crud
```

See also [`examples/mongo_indexes`](examples/mongo_indexes/) and [`examples/mongo_listeners`](examples/mongo_listeners/).

**PostgreSQL** (typed store + projection):

```bash
export POSTGRES_DSN='postgres://...'
go run ./examples/postgres_widgets
```

**Codegen only** (generate `RowStore` + SQL next to your models):

```bash
go run github.com/dhlab-tech/go-inmemory-platform/cmd/entitygen -h
```

Full wiring (migrations, `migrate.Up`, `TupleDecoder`, `replace` in `go.mod`) is described in **[docs/integration.md](docs/integration.md)**.

## Examples (from repo root)

| Directory | Command | Needs |
|-----------|---------|--------|
| [examples/mongo_crud](examples/mongo_crud/) | `go run ./examples/mongo_crud` | `MONGODB_URI`, replica set |
| [examples/mongo_indexes](examples/mongo_indexes/) | `go run ./examples/mongo_indexes` | same |
| [examples/mongo_listeners](examples/mongo_listeners/) | `go run ./examples/mongo_listeners` | same |
| [examples/postgres_widgets](examples/postgres_widgets/) | `go run ./examples/postgres_widgets` | `POSTGRES_DSN` |

## Migrating from older modules

| Before | After |
|--------|--------|
| `github.com/dhlab-tech/go-mongo-platform/pkg/inmemory` | `github.com/dhlab-tech/go-inmemory-platform/pkg/mongo/inmemory` |
| `github.com/dhlab-tech/go-mongo-platform/pkg/mongo` | `github.com/dhlab-tech/go-inmemory-platform/pkg/mongo` |
| `github.com/dhlab-tech/go-postgres-platform/pkg/...` | `github.com/dhlab-tech/go-inmemory-platform/pkg/postgres/...` |

`StreamEventListener` uses **string** document IDs (`Delete(ctx, id string)`, `Update(ctx, id string, …)`). Code that still uses `primitive.ObjectID` in those signatures must be updated when switching from legacy `go-mongo-platform`.

## Development

```bash
go test ./...
go test -race ./...
go test -tags=integration -count=1 ./internal/testwidget/...   # POSTGRES_TEST_URL

# Optional benchmarks (reflection, processor prepare paths, Spawn, …)
go test ./pkg/projection ./pkg/mongo ./pkg/mongo/inmemory ./pkg/postgres/inmemory -run='^$' -bench=. -benchmem -count=1
```

## Module dependencies

The `go.mod` lists both **mongo-driver** and **pgx** / goose. Import only the packages your binary uses; the linker drops unused packages, but `go mod` still records the module’s direct requirements.

## Project policies & changelog

| Document | Role |
|----------|------|
| [CHANGELOG.md](CHANGELOG.md) | Release history |
| [API_STABILITY.md](API_STABILITY.md) | Public API scope and semver intent |
| [ARCHITECTURE_CONTRACT.md](ARCHITECTURE_CONTRACT.md) | Normative architecture and guarantees |
| [NON_GOALS.md](NON_GOALS.md) | Explicit out-of-scope items |
| [CONTRIBUTING.md](CONTRIBUTING.md) | How to contribute |
| [SECURITY.md](SECURITY.md) | Vulnerability reporting |
| [SUPPORT.md](SUPPORT.md) | Community vs commercial support |
| [LICENSE](LICENSE) | MIT License |

**Guides (in `docs/`):** [positioning.md](docs/positioning.md) · [decision-tree.md](docs/decision-tree.md) · [anti-patterns.md](docs/anti-patterns.md) · [production-guide.md](docs/production-guide.md) · [troubleshooting.md](docs/troubleshooting.md) · [integration.md](docs/integration.md)
