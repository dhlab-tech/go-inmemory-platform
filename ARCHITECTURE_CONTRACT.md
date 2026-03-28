# Architecture contract (normative)

**Status:** Public normative document  
**Version:** 1.0 (go-inmemory-platform)  
**Purpose:** Architectural constraints and behavioral guarantees shared by Mongo and PostgreSQL backends

> If other docs conflict with this contract, **this contract wins**.

---

## 1. Core constraints

### 1.1 Single-process semantics

**MUST:** The in-memory projection is **process-local** and **instance-scoped**.

- Each process holds its own projection; there is no shared in-memory state across hosts or pods.
- **MUST NOT** assume cross-instance coherence of the cache.

---

### 1.2 Durable store as source of truth

**MUST:** All authoritative writes go to the **configured database** (MongoDB or PostgreSQL). The projection is a **derived view**, not a second system of record.

- **MongoDB path:** applications persist via `mongo.Processor` / `Updater` (and related APIs); the cache is updated from **Change Streams** (and optional warmup).
- **PostgreSQL path:** applications persist via generated **`RowStore`** (or compatible implementations); the cache is fed by the same listener pipeline, optionally including **logical replication** (`pgoutput`) when enabled.

**MUST NOT:** Treat in-memory structures as durable or bypass the database for authoritative state.

---

### 1.3 No distributed coordination

**MUST NOT:** Provide distributed locks, leader election, or cross-service cache invalidation.

**SHOULD NOT:** Use this library as a coordination or shared-state layer between services.

---

### 1.4 Ephemeral projection

**MUST:** In-memory state is **lost on restart**.

**SHOULD:** Rebuild or warm the projection from the database (Mongo: query / stream; Postgres: load + replication or equivalent) on startup.

---

### 1.5 No default eviction

**MUST NOT:** Ship TTL, LRU, or other automatic eviction policies for the projection.

**SHOULD:** Size and lifecycle of cached data remain an **application** concern.

---

### 1.6 No exactly-once across restarts

**MUST NOT:** Guarantee exactly-once delivery of replication/stream events across crashes.

**SHOULD:** Rely on versioning / idempotency in the app when hard deduplication is required.

---

## 2. Consistency guarantees

### 2.1 Read-after-write (`Await*`)

**MUST:** After `AwaitCreate`, `AwaitUpdate`, `AwaitDelete`, or `AwaitUpdateDoc` returns successfully in a process, the in-memory projection observable through that process’s `Projection[T]` reflects the corresponding write (subject to normal error handling).

This is a **single-process** guarantee, not cross-instance.

---

### 2.2 Synchronization with the database

**MongoDB:** **MUST** drive cache updates from **Change Streams** when streaming is enabled; ordering follows the stream. Resume tokens apply where used.

**PostgreSQL:** With logical replication configured, **MUST** apply decoded tuples through the same listener model as other updates so indexes stay consistent. Without replication, the process **SHOULD** still load or refresh data via `RowStore` / startup hooks as implemented by the app.

**SHOULD:** Handle disconnects (network, DB restart) with reconnection or controlled degradation as appropriate for your deployment.

---

### 2.3 Non-`Await` writes

`Create`, `Update`, `Delete`, and `UpdateDoc` **persist** without waiting for the projection; readers may observe stale in-memory data until the sync path catches up.

---

## 3. Explicitly not guaranteed

- Consistency **across** service instances.
- Exactly-once event delivery across process restarts.
- Zero replication lag under partitions or overload.
- Cross-service cache coherence or TTL semantics.

---

## 4. Responsibility boundaries

| Concern | Owner |
|--------|--------|
| Durability, relational/document correctness | MongoDB / PostgreSQL |
| Event ordering | Change Streams or replication/WAL as configured |
| Building and indexing the projection | go-inmemory-platform (`pkg/projection` + adapters) |
| Memory limits, sharding, multi-region | Application / infrastructure |
| Generated SQL and `RowStore` correctness | Maintainers + consumers (codegen inputs) |

---

## 5. Contract stability

Normative rules in this document are meant to stay **stable across `v0.x`**; if they change, the change **MUST** be called out in [CHANGELOG.md](CHANGELOG.md) and reflected in [API_STABILITY.md](API_STABILITY.md) where relevant.

---

## Related documentation

- [NON_GOALS.md](NON_GOALS.md)
- [README.md](README.md)
- [docs/integration.md](docs/integration.md)
- [docs/positioning.md](docs/positioning.md)
- [docs/anti-patterns.md](docs/anti-patterns.md)
- [docs/production-guide.md](docs/production-guide.md)
