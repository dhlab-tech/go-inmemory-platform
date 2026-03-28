# Production guide

**Status:** public documentation  
**Purpose:** deployment and operations for **go-inmemory-platform** (Mongo and Postgres paths)

---

## MongoDB: Change Streams

### Replica set

Change Streams require a **replica set** (not standalone-only topology).

- Set **`replicaSet=`** in the URI and include members as needed:
  `mongodb://h1:27017,h2:27017/db?replicaSet=rs0`

### Behaviour (typical)

- **Resume tokens** when the driver/stream layer provides them.
- On disconnect: reconnect; if resume is impossible, implementations often **reload** from the database—confirm in your version and logs.
- **Restart:** in-memory state is **empty** until warmup + stream catch-up completes.

### Resilience

On repeated failures: check RS health, networking, auth, and Oplog; add app-level backoff/alerts around listener lifecycle.

---

## PostgreSQL: migrations and replication

### Schema and migrations

- Use **goose** SQL (often **generated** alongside `RowStore` via `entitygen`).
- Run **`postgres/migrate.Up`** in controlled environments or a job; many teams disable auto-migrate in prod or run under a restricted role.
- Keep migrations **additive** when evolving live DBs (`IF NOT EXISTS`, etc.).

### Logical replication (optional)

If you use **`NewLogicalReplicationListener`**:

- Configure **publication**, **slot**, connection string, and **`TupleDecoder`** (generated `XxxTupleDecoder()`).
- **`REPLICA IDENTITY FULL`** is commonly required for full row images on updates—match your generator/docs.
- Monitor **replication lag**, slot disk usage, and connection drops.
- On prolonged outage: plan whether to **resync** (reload from `RowStore`) in addition to catching up WAL—app-specific.

### Connection pool

- Size **`pgxpool`** for concurrency; watch **max connections** vs DB limits.
- Timeouts and context cancellation should match your SLOs.

---

## Observability

### Metrics (suggested)

- **Mongo:** RS lag indicators, pool usage, write latency, stream errors/reconnects.
- **Postgres:** pool wait time, query latency, replication lag, slot health.
- **App:** projection document count (if cheap), **heap**, **Await** latency, rebuild duration, listener error rate.

### Logging

Libraries often use **zerolog**; correlate **stream/replication** events with deploys.

### Health checks

- DB **connectivity**.
- **Listener** alive (Mongo stream or replication worker).
- **Readiness** after **warmup** / initial load (don’t serve traffic that assumes a full cache until your init completes).

---

## Performance

### Memory

- Projection size tracks **working set** you load; **no eviction**—plan data volume per instance.
- Partition archives in the DB or shrink what you attach to `NewInMemory` (filters / separate collections or tables).

### Startup

- Large collections/tables → long **first load**; gate readiness probes until done.
- **Mongo:** indexes on queried fields speed full scans.
- **Postgres:** PK and query patterns for `All` / warmup filters.

### Writes

- **`Await*`** adds latency until the projection observes the change—expected.
- Non-**`Await`** writes return after DB ack; readers may be briefly stale in RAM.

---

## Horizontal scaling

- **N replicas ⇒ N independent projections** and **N × memory** for the same loaded data.
- No coordination between instances; scale DB and replication capacity accordingly.

## Rolling deploys

- New pods **rebuild** RAM; stagger if you want to avoid synchronized DB read spikes.
- **Blue/green:** new instances warm independently; no shared cache to migrate.

---

## Shutdown

1. Stop new traffic (LB / readiness).
2. Drain in-flight work.
3. Close **stream / replication** consumers and **DB** pools cleanly.

---

## Limitations (reminder)

- Process-local state; no cross-pod coherence.
- No TTL eviction.
- Mongo path needs **replica set**.
- Postgres replication path needs correct **decoder** and **schema** alignment.

---

## Troubleshooting

See [troubleshooting.md](troubleshooting.md).

---

## Related documentation

- [positioning.md](positioning.md)
- [decision-tree.md](decision-tree.md)
- [integration.md](integration.md)
- [ARCHITECTURE_CONTRACT.md](../ARCHITECTURE_CONTRACT.md)
