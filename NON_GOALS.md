# Non-goals (normative)

**Status:** Public normative document  
**Version:** 1.0 (go-inmemory-platform)  
**Purpose:** Prevent misaligned expectations and misuse

> If marketing or secondary docs conflict with this file, **this file wins**.

---

## Out of scope

The project **does not** aim to provide:

- **Distributed cache semantics** — no shared in-memory state across services or nodes.
- **Cross-service synchronization** — no built-in coordination between instances.
- **TTL / automatic eviction** — no default expiration or LRU for the projection.
- **Generic key-value or Redis-style cache** — the API is typed entities + indexes + `Projection[T]`, not a KV store.
- **Polyglot SDKs** — Go only.
- **Persistence of the projection** — no snapshot/WAL of the in-memory layer; rebuild from the DB.
- **Exactly-once delivery across restarts** — duplicates or gaps are possible under failure; apps may deduplicate.
- **Distributed primitives** — no locks, leader election, or shared mutable cache.
- **A second source of truth** — MongoDB or PostgreSQL remains authoritative.
- **“Works without understanding the datastore”** — Mongo paths assume Change Streams (replica set); Postgres paths assume migrations, `RowStore`, and optionally logical replication and generated decoders.

---

## Anti-audience

**Not** aimed at:

- Engineers who need a **drop-in Redis replacement** for shared mutable state.
- Teams that require **one global in-RAM view** across all replicas.
- **Polyglot** stacks needing a language-neutral cache protocol.
- **TTL-first** product rules without app-level enforcement.
- **Heavy analytics / ETL** where an application projection layer is the wrong tool.

---

## Expected user profile

You likely fit if you:

- Run **Go** services and own **MongoDB and/or PostgreSQL** operations.
- Want **fast in-process reads** with the database as source of truth.
- Can use **`Await*`** when you need read-after-write within one process.
- Accept **process-local** cache semantics and operational tradeoffs of Change Streams or replication.

---

## When to use something else

| Need | Consider |
|------|----------|
| Shared mutable cache across services | Redis, memcached, or a dedicated data service |
| TTL / automatic expiry | A cache or store with TTL support |
| Cross-language clients | HTTP/gRPC + a neutral store |
| Distributed locking / consensus | etcd, Consul, DB-native patterns, etc. |
| Strict exactly-once across crashes | App-level idempotency + your broker/DB guarantees |

---

## Why this matters

Stating non-goals **reduces** wrong tickets, **clarifies** positioning next to `go-mongo-platform` / raw drivers, and **aligns** senior engineers on what the module will never become.

---

## Related documentation

- [ARCHITECTURE_CONTRACT.md](ARCHITECTURE_CONTRACT.md)
- [API_STABILITY.md](API_STABILITY.md)
- [README.md](README.md)
- [docs/positioning.md](docs/positioning.md)
- [docs/decision-tree.md](docs/decision-tree.md)
- [docs/anti-patterns.md](docs/anti-patterns.md)
