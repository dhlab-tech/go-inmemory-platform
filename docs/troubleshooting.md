# Troubleshooting

**Status:** public documentation  
**Purpose:** common failures and operational responses for **go-inmemory-platform**

---

## MongoDB: Change Stream disconnect

**Symptoms:** projection stops updating; writes succeed but RAM reads look stale; stream errors in logs.

**Causes:** network blips, RS step-down, timeouts, server restart.

**What to do:**

1. Verify connectivity and **`rs.status()`**.
2. Inspect MongoDB logs.
3. Library typically **reconnects**; if stuck, rolling **restart** forces reload + new stream.
4. Alert on reconnect storm / sustained disconnect.

**Prevention:** healthy RS, correct URI, pool limits, monitoring.

---

## MongoDB: “Change Streams require a replica set”

**Fix:** deploy as **replica set** (even single-node RS for dev); add **`replicaSet=`** to the connection string.

---

## Process restart

**Symptoms:** empty or rebuilding cache; slow readiness after deploy.

**Expected:** RAM is **ephemeral**; reload from DB + stream catch-up.

**What to do:** readiness probe until **warmup** finished; track **duration**; optimize queries/indexes if too slow.

---

## MongoDB unavailable

**Symptoms:** writes fail; projection **freezes** at last state.

**Expected:** no speculative authoritative writes to RAM; resume when DB returns.

---

## PostgreSQL: replication lag or disconnect

**Symptoms:** Postgres writes visible in SQL but RAM outdated when using logical replication; replication errors in logs.

**What to do:**

1. Check **slot**, **publication**, **network**, **disk** (slot retention).
2. Confirm **decoder** matches table layout (`TupleDecoder` / column names).
3. After extended outage, you may need a defined **resync** (reload entities via `RowStore`) per your app policy.

---

## PostgreSQL: migration / schema drift

**Symptoms:** `Scan` errors, `UpdateDoc` type switches missing columns, replication decode failures.

**What to do:** run **goose** to head in the right order; regenerate **entitygen** outputs when structs change; avoid manual schema drift vs generated SQL.

---

## Stale projection reads

**Causes:** stream/replication paused, restart mid-rebuild, used **`Create`/`Update`** without **`Await*`** and read immediately.

**What to do:** confirm listener health; for strict local read-after-write use **`Await*`**; otherwise accept short lag.

---

## `cache is not initialized, AwaitCreate requires cache`

**Cause:** **`mongo/inmemory.NewInMemory`** or **`postgres/inmemory.NewInMemory`** built **without** a live projection bundle (e.g. Mongo path with **no** valid stream so cache isn’t wired—see your `Entity` / deps).

**What to do:** ensure **Change Stream** path is enabled when you need **`Await*`**; for Postgres, wire **store + listener** per [integration.md](integration.md).

---

## High memory use

**Causes:** full collection/table loaded; leaks in app code holding references.

**What to do:** measure **RSS** vs entity count; narrow **warmup**; archive old data; profile leaks.

---

## Slow startup

**Causes:** huge full table scan / `All()` for warmup; missing DB indexes.

**What to do:** reduce loaded set, add **DB indexes**, async readiness.

---

## Health checks (checklist)

1. Database **ping**.
2. **Listener** running (stream or replication).
3. **Readiness** after initial load.
4. **Memory** within budget.

---

## FAQ

**Q: Projection not updating?**  
**A:** Mongo: Change Stream health. Postgres: replication worker + decoder + publication.

**Q: Single-node Mongo?**  
**A:** Use a **replica set** configuration (can be one node) for Change Streams.

**Q: DB down?**  
**A:** Writes fail; RAM may be stale until recovery and sync resume.

**Q: TTL eviction?**  
**A:** Not from this library—handle in app or another system.

**Q: Shared state across instances?**  
**A:** Not supported; use an external store.

**Q: Consistency model?**  
**A:** See [ARCHITECTURE_CONTRACT.md](../ARCHITECTURE_CONTRACT.md): **`Await*`** is **process-local** read-after-write.

---

## Getting help

- **GitHub Issues / Discussions** for the `dhlab-tech/go-inmemory-platform` repo.
- Commercial: **[SUPPORT.md](../SUPPORT.md)** — **support@digital-heroes.tech**

---

## Related documentation

- [positioning.md](positioning.md)
- [decision-tree.md](decision-tree.md)
- [production-guide.md](production-guide.md)
- [anti-patterns.md](anti-patterns.md)
