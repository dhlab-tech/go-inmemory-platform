# Anti-patterns

**Status:** public documentation  
**Purpose:** typical mistakes when using **go-inmemory-platform** and how to avoid them

Applies to **MongoDB** (Change Streams) and **PostgreSQL** (typed `RowStore`, optional logical replication) paths unless noted.

---

## 1. Redis (or another cache) for the same rows

**Smell:** Duplicating the same entity data in Redis “for speed” while also using this projection.

**Why it’s bad:** Dual writes, invalidation races, stale reads, second source of truth.

**Instead:** Use the in-memory projection as the only RAM view of that database data. Use Redis only for **cross-cutting** concerns (sessions, rate limits, genuinely shared keyspace) that are **not** this projection.

---

## 2. Assuming shared RAM across replicas

**Smell:** Expecting all pods/instances to see the same in-memory state.

**Why it’s bad:** The projection is **process-local**; there is no cross-instance sync.

**Instead:** Design APIs and UX for per-process views; use external stores for shared mutable state.

---

## 3. Persisting the projection (snapshots / WAL)

**Smell:** Writing the cache to disk to “survive” restarts.

**Why it’s bad:** Extra failure modes; the design is **rebuild from the database** (Mongo query + stream, or Postgres load + replication / polling as you wired it).

**Instead:** Treat RAM as ephemeral; tune **startup load** and **DB** availability.

---

## 4. Expecting built-in TTL / LRU

**Smell:** Waiting for automatic eviction from the projection.

**Why it’s bad:** Not provided; memory is an **application** concern.

**Instead:** Cap working set in Mongo/Postgres, archive old data, or use a **TTL-capable** system for TTL-specific use cases—not this layer.

---

## 5. Treating `Await*` as distributed consistency

**Smell:** Believing `AwaitCreate` / `AwaitUpdate` / … synchronize **all** replicas.

**Why it’s bad:** `Await*` is **single-process** read-after-write only.

**Instead:** Use `Await*` inside one process after its own writes; accept **eventual** differences across instances.

---

## 6. Global mutex for “cluster-wide” coordination

**Smell:** In-process locks to coordinate multiple services.

**Why it’s bad:** Mutexes don’t span processes; wrong layer.

**Instead:** etcd, Consul, Redis primitives, or DB constraints—outside this library.

---

## 7. Reordering replication / stream events (Mongo)

**Smell:** Retry or batching logic that reorders **Change Stream** events.

**Why it’s bad:** Ordering is part of the consistency story; reordering can corrupt the view.

**Instead:** Preserve order; retry failed **handling** without permuting the stream.

---

## 8. Ignoring ordering or gaps (Postgres replication)

**Smell:** Applying logical replication messages out of order, or skipping error handling on the replication connection.

**Why it’s bad:** The projection and indexes assume a coherent apply order; silent drops desync RAM from the database.

**Instead:** Follow the listener pipeline you configured; monitor replication lag and errors; restart with a controlled resync path if your app defines one.

---

## 9. Second source of truth

**Smell:** Treating the projection (or a side table) as authoritative over Mongo/Postgres.

**Why it’s bad:** Violates the [architecture contract](../ARCHITECTURE_CONTRACT.md).

**Instead:** **Database first**; projection is a derived index-friendly view.

---

## 10. Exactly-once across restarts

**Smell:** Assuming no duplicate or missed events after crash.

**Why it’s bad:** Not guaranteed by the stack; duplicates and edge-case gaps are possible.

**Instead:** Idempotent handlers, versioning, or explicit dedup where you need hard guarantees.

---

## 11. Using the library for distributed locking

**Smell:** Distributed locks “because we already have the cache.”

**Why it’s bad:** No coordination primitives here.

**Instead:** Dedicated lock services or DB-native patterns.

---

## 12. Postgres: replication configured but wrong / missing decoder

**Smell:** Enabling logical replication without a correct **`TupleDecoder`** (generated `XxxTupleDecoder()`), or mismatched publication columns.

**Why it’s bad:** WAL tuples won’t map cleanly to your entity; subtle desync or failures.

**Instead:** Follow [integration.md](integration.md): codegen, `REPLICA IDENTITY FULL` where required, aligned schema and decoder.

---

## Related documentation

- [ARCHITECTURE_CONTRACT.md](../ARCHITECTURE_CONTRACT.md)
- [NON_GOALS.md](../NON_GOALS.md)
- [decision-tree.md](decision-tree.md)
- [production-guide.md](production-guide.md)
