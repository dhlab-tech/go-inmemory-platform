# Positioning: go-inmemory-platform

**Status:** public documentation  
**Purpose:** canonical positioning and differentiation

---

## Core positioning

**go-inmemory-platform** is a **Go module** that implements an **in-process, strongly consistent projection** over **MongoDB** and/or **PostgreSQL**, behind a single storage-agnostic API: [`projection.Projection[T]`](../pkg/projection/projection.go).

It is **not** a generic cache, **not** a Redis replacement, and **not** a distributed cache.

---

## What this is

### Pattern

- The **database** (Mongo or Postgres, per integration) is the **source of truth**.
- All authoritative **writes** go through the database APIs you wire (`mongo.Processor` / `Updater`, or generated **`RowStore`**).
- The **in-memory layer** is a **projection**: indexes, listeners, fast reads.
- **Each process** holds its own projection.
- **`Await*`** methods give **read-after-write** guarantees **inside that process** after a write.

### MongoDB path

- Sync from **Change Streams** (replica set).
- Rich BSON-oriented **`Processor`** / **`Searcher`** stack.

### PostgreSQL path

- Sync via **application writes** through typed **`RowStore`** (codegen from `cmd/entitygen`).
- Optional **logical replication** (`pgoutput`) to push remote changes into the same listener model.
- Hot SQL paths avoid reflection when using generated stores.

### Shared projection features

- Tag-driven **indexes** (inverse, inverse unique, sorted, suffix).
- Same **`Projection[T]`** surface: `AwaitCreate` / `AwaitUpdate` / `AwaitDelete`, `AwaitUpdateDoc`, non-blocking `Create` / `Update` / `Delete` / `UpdateDoc`, `Spawn`.

---

## What this is not

- **Not** TTL/LRU eviction.
- **Not** shared RAM across services.
- **Not** pub/sub or messaging.
- **Not** durable in-memory state (restart → rebuild / catch up).
- **Not** a CDC platform (no Kafka replacement); Mongo uses streams, Postgres may use replication for **app** consistency, not warehouse ETL.
- **Not** polyglot—**Go only**.

---

## Target scenarios

Backend services that:

- Use **Go**.
- Primary data lives in **MongoDB** and/or **PostgreSQL**.
- Need **fast in-process reads** with **indexes**.
- Want **`Await*`** for strict read-after-write **after local writes** in one process.
- Prefer **one module** instead of separate `go-mongo-platform` + `go-postgres-platform` imports.

Typical pain this addresses:

> “We already trust the database, but Redis / hand-rolled cache invalidation gives stale reads and races. We want one write path and predictable reads in-process.”

---

## Why other options miss the mark (for this niche)

| Approach | Gap |
|----------|-----|
| Redis / distributed cache | Dual write or invalidation complexity; eventual consistency; network on every read. |
| TTL caches | Poor fit for domain invariants; still not DB-native ordering. |
| Heavy CDC (Kafka + connectors) | Operational weight; aimed at pipelines, not per-request app consistency. |

---

## Differentiation (summary)

| Aspect | go-inmemory-platform | Redis / dist. cache |
|--------|----------------------|---------------------|
| Source of truth | Mongo or Postgres | Often dual-write or cache-aside |
| Read-after-write (`Await*`) | Per process, strong | Not guaranteed |
| Shared state across replicas | No | Yes (by design) |
| Read latency | In-process | Network |
| Backends | Mongo + Postgres in one module | N/A |

---

## When to choose / not choose

**Choose** when you accept process-local projection, use one of the supported databases, and benefit from unified **`Projection[T]`** + codegen on Postgres.

**Do not choose** for cross-service shared cache, TTL-as-a-service, polyglot clients, or distributed coordination—see [NON_GOALS.md](../NON_GOALS.md) and [decision-tree.md](decision-tree.md).

---

## Related documentation

- [decision-tree.md](decision-tree.md)
- [production-guide.md](production-guide.md)
- [troubleshooting.md](troubleshooting.md)
- [integration.md](integration.md)
