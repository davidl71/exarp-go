**Tag hints:** `#performance` `#database` `#concurrency` `#mcp`

# exarp-go — remaining stability and performance work (inventory)

**Status:** Research note (2026-03-30). No implementation commitment; use for backlog triage.  
**Related tasks:** See exarp-go Todo2 for current IDs (titles: queue worker DB init vs pool; tool semaphore permits vs config).

**See also:** [PERFORMANCE_GUIDE.md](../PERFORMANCE_GUIDE.md) (concise pattern table + links to all performance docs) and [Go Optimization Guide — common patterns](https://goperf.dev/01-common-patterns/).

---

## 1. Queue worker: per-job database init

**Area:** `internal/queue/worker.go` — `handleTaskExecuteJob` calls `database.InitWithCentralizedConfig` or `database.Init` on every Asynq task.

**Why it matters:** Under high job volume, repeated init may add latency or connection churn depending on how `Init*` is implemented. A long-lived pool or init-once keyed by `ProjectRoot` may be preferable if profiling confirms cost.

**Next step:** Measure worker throughput with/without caching init; align with `database` package semantics (safe no-op vs reopen).

---

## 2. Tool semaphore: hard-coded permits and `sync.Once`

**Area:** `internal/factory/server.go` — `toolSemaphoreMiddleware` uses `security.GetToolSemaphore(10)`.  
**Area:** `internal/security/semaphore.go` — `GetToolSemaphore` uses `sync.Once`, so the **first** `permits` argument wins for process lifetime; later values are ignored.

**Why it matters:** Operators cannot tune concurrent tool execution via config. `ToolsConfig.ToolLimit` in `internal/config/schema.go` is a separate concept (tool allowlists), not wired to this semaphore.

**Next step:** Add a dedicated config field (e.g. `max_concurrent_tools`), initialize the semaphore from loaded config once at server startup, and document the relationship to rate limiting and Redis queue concurrency.

---

## 3. Full-tree walks (scale)

**Area:** `health` docs scanning, task analysis / ownership walks, and similar `filepath.Walk` / `WalkDir` paths.

**Why it matters:** Large client repos pay linear disk traversal on each invocation.

**Next step:** Optional caching, narrower roots, or incremental scans; prioritize if scorecard/health latency is reported as a problem.

---

## 4. Concurrency correctness

**Area:** `internal/database/tasks_lock*.go` — task claims and leases; tests such as `tasks_lock_concurrency_test.go` (concurrent `ClaimTaskForAgent`).

**Why it matters:** Multi-agent MCP clients and queue workers depend on exclusive claims under contention.

**Next step:** Keep concurrent claim tests in CI; extend coverage for batch claim and lease renewal if gaps appear.

---

## 5. Existing positive patterns (keep)

- **Report overview:** `internal/tools/report.go` uses `singleflight` to dedupe concurrent overview requests.
- **Database:** Centralized `MaxConnections` in config (`internal/config/schema.go`) for SQLite pool sizing when init uses it.

---

## References (code)

| Topic | Location |
|-------|----------|
| Queue worker handler | `internal/queue/worker.go` |
| Tool semaphore middleware | `internal/factory/server.go` |
| Semaphore + `sync.Once` | `internal/security/semaphore.go` |
| DB max connections | `internal/config/schema.go` (database section) |
