# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Project governance and policy docs aligned with the consolidated module: [API_STABILITY.md](API_STABILITY.md), [ARCHITECTURE_CONTRACT.md](ARCHITECTURE_CONTRACT.md), [NON_GOALS.md](NON_GOALS.md), [CONTRIBUTING.md](CONTRIBUTING.md), [SECURITY.md](SECURITY.md), [SUPPORT.md](SUPPORT.md), and [LICENSE](LICENSE) (MIT).
- Extended guides under `docs/`: [positioning.md](docs/positioning.md), [decision-tree.md](docs/decision-tree.md), [anti-patterns.md](docs/anti-patterns.md), [production-guide.md](docs/production-guide.md), [troubleshooting.md](docs/troubleshooting.md) (adapted from the legacy Mongo-focused docs for the unified Mongo + Postgres module).

## [0.0.1] - 2026-03-28

### Added

- Initial tagged release of **`github.com/dhlab-tech/go-inmemory-platform`**: one module for in-memory projection with **MongoDB** and **PostgreSQL**.
- **`pkg/domain`**: shared `Entity` constraint (`ID`, `Version`, `SetDeleted`).
- **`pkg/projection`**: `Cache`, tag-driven indexes (inverse, inverse unique, sorted, suffix), listener chain, and storage-agnostic **`Projection[T]`** with:
  - **`AwaitCreate` / `AwaitUpdate` / `AwaitDelete`** — persist and block until the in-memory view reflects the change.
  - **`Create` / `Update` / `Delete`** — persist without waiting on the projection.
  - **`AwaitUpdateDoc` / `UpdateDoc`** — partial updates via `set` / `unset` maps (BSON field names); Postgres uses generated `RowStore.UpdateDoc`, Mongo uses `Updater.UpdateOne` with optional **`AwaitUpdateDocBSON` / `UpdateDocBSON`** on the Mongo `inmemory` facade.
- **`pkg/mongo`**: `Searcher`, `Processor`, change-stream listener, `Updater`, and related wiring.
- **`pkg/mongo/inmemory`**: **`NewInMemory`** combining Mongo client, optional change stream, and projection.
- **`pkg/postgres`**: typed **`RowStore[T]`**, **`TupleDecoder`** for logical replication (pgoutput), errors and helpers.
- **`pkg/postgres/migrate`**: goose migrations on `pgxpool`.
- **`pkg/postgres/inmemory`**: **`NewInMemory`** combining generated store, projection, and optional logical replication.
- **`cmd/entitygen`** + **`internal/entitygen`**: codegen for **`RowStore`**, goose SQL, and tuple decoding from struct tags (`pkg/entitymeta`).
- **Examples**: Mongo CRUD, indexes, listeners; Postgres widgets sample (`examples/`).
- **`docs/integration.md`**: wiring a consumer service (`go.mod`, replication, string IDs on listeners).
- **Tests and benchmarks**: unit tests across packages; integration test fixture under `internal/testwidget` (`-tags=integration`, `POSTGRES_TEST_URL`); benchmarks for reflection-heavy paths, `Processor` prepare, `Spawn`, and generated `WidgetStore` (integration).

### Notes

- **`0.0.x`** indicates early API surface; prefer pinning a version or `replace` in `go.mod` while downstreams migrate from **`go-mongo-platform`** / **`go-postgres-platform`**.
- **`StreamEventListener`** uses **string** document IDs (`Delete(ctx, string)`, `Update(ctx, id string, …)`), consistent across both backends.

[Unreleased]: https://github.com/dhlab-tech/go-inmemory-platform/compare/v0.0.1...HEAD
[0.0.1]: https://github.com/dhlab-tech/go-inmemory-platform/releases/tag/v0.0.1
