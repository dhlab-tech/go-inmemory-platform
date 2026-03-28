# Decision tree: cache / Redis vs go-inmemory-platform

**Status:** public documentation  
**Purpose:** choose between external caches and this module

---

## Quick tree

```
Need shared mutable state across many service instances?
├─ YES → Redis, DB, or another shared store — not this projection alone
└─ NO → Continue

Need TTL / LRU as a product feature without app logic?
├─ YES → Use a cache with TTL
└─ NO → Continue

Need cross-language clients to the same in-RAM dataset?
├─ YES → Expose APIs + a neutral store — not in-process Go-only RAM
└─ NO → Continue

Building a Go service backed by MongoDB and/or PostgreSQL?
├─ NO → This module is unlikely to fit
└─ YES → Continue

Mongo path: can you run a replica set (Change Streams)?
├─ NO (Mongo only) → Fix topology or don’t use the Mongo in-memory path
└─ YES or using Postgres → Continue

Need strong read-after-write after your own writes in ONE process?
├─ NO → You might still use non-Await writes + projection reads; understand staleness
└─ YES → Strong fit: use Await*

Want to avoid dual-write to DB + external cache for the same entities?
├─ NO → You may still combine tools deliberately
└─ YES → Strong fit: single write path to DB, projection follows
```

---

## Use go-inmemory-platform when

**Architecture**

- Go service(s); primary data in **MongoDB** and/or **Postgres**.
- **Process-local** cache per instance is OK.
- No requirement that **all** replicas share one RAM image.

**Consistency**

- **`Await*`** for write-then-read **in the same process** after DB write.
- Tolerate **eventual** differences **across** instances.

**Operations**

- Mongo: **replica set** for Change Streams.
- Postgres: migrations (**goose**), generated **`RowStore`**, optional **logical replication** if you need WAL-driven updates—see [integration.md](integration.md).

**Performance**

- Read-heavy, in-memory indexes/filters, minimal read latency.

---

## Use Redis (or similar) when

- **Shared** keyspace across services or regions.
- **TTL** / eviction as a core feature.
- **Polyglot** access patterns.
- Pub/sub, rate limiting, or session store—not entity projection.

---

## Use a distributed cache when

- Large shared working set, cross-region, explicit cache cluster operations—and you accept eventual consistency and ops cost.

---

## Migration notes

### From Redis cache-aside (same entities)

1. Stop writing entity duplicates to Redis for those types.
2. Route reads through **`Projection[T]`** / indexes after warmup.
3. Replace “write then read from Redis” with **`Await*`** where you need freshness in-process.
4. Keep Redis only for unrelated concerns.

### From legacy `go-mongo-platform` / `go-postgres-platform`

- Map imports to **`github.com/dhlab-tech/go-inmemory-platform/...`** (see [README.md](../README.md)).
- Listener **`Delete` / `Update`** use **string** IDs.

---

## Comparison tables

### Consistency

| Solution | Model | Read-after-write in one process |
|----------|--------|----------------------------------|
| go-inmemory-platform | `Await*` strong locally | Yes (after `Await*`) |
| Redis | Eventual | Not guaranteed |

### Operations

| Solution | Extra infra |
|----------|-------------|
| go-inmemory-platform | Your Mongo and/or Postgres (replication optional) |
| Redis | Redis cluster + client semantics |

### Reads

| Solution | Typical read path |
|----------|-------------------|
| go-inmemory-platform | In-process |
| Redis | Network |

---

## Misconceptions

- **“Redis replacement”** — False. Different role (shared vs process-local).
- **“Shared state across pods”** — False. One projection per process.
- **“TTL built in”** — False. No automatic eviction.
- **“Faster than Redis”** — Misleading; in-process reads are low-latency, but the problems solved differ.

---

## Checklist

Before adopting:

- [ ] Go service
- [ ] Mongo (replica set) and/or Postgres with a clear persistence story
- [ ] OK with per-instance RAM; cross-instance eventual
- [ ] No mandatory TTL from this library
- [ ] Understand `Await*` scope (single process)

---

## Related documentation

- [positioning.md](positioning.md)
- [anti-patterns.md](anti-patterns.md)
- [production-guide.md](production-guide.md)
- [troubleshooting.md](troubleshooting.md)
