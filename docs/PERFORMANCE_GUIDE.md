# Performance guide (exarp-go)

**Last updated:** 2026-03-30

Use this page as the **index** for performance work: general Go patterns (external reference), how they apply to this repo, and links to detailed notes.

**Rule:** Confirm bottlenecks with **benchmarks** and **`pprof`** (CPU + heap) before micro-optimizing. The MCP server is often **I/O- and SQLite-bound**, not CPU-bound.

---

## Quick reference: common Go performance patterns

Aligned with the [Go Optimization Guide — common patterns](https://goperf.dev/01-common-patterns/) (memory, concurrency, I/O, compiler).

| Area | Representative techniques | Primary goal | Relevance to exarp-go |
|------|---------------------------|--------------|------------------------|
| **Memory** | Preallocate slices/maps; `sync.Pool` where reuse is clear; struct field alignment; avoid unnecessary interface boxing; trim allocations on hot decode paths | Less GC churn, better cache locality | JSON-RPC args maps, task graph helpers, large `ListTasks` payloads; see `OPTIMIZATION_RESULTS.md` |
| **Concurrency** | Worker pools; atomics / small critical sections; `sync.Once` for one-time init; immutable snapshots; **`context` for cancel/timeouts** | Controlled parallelism, fewer contentions | Tool semaphore (`internal/factory`, `internal/security`); queue worker; claim/lease paths — see `research/STABILITY_AND_PERFORMANCE_REMAINING.md` |
| **I/O** | Buffered readers/writers; **batching** DB ops and fewer full-tree walks | Fewer syscalls / round trips | SQLite batch patterns; consolidated `filepath.Walk` (see `PERFORMANCE_OPTIMIZATION_RESEARCH.md`); file cache — `PERFORMANCE_FILE_CACHE.md` |
| **Compiler / analysis** | Build flags when justified; **escape analysis** to keep hot structs stack-friendly | Last-mile wins after profiling | Use after identifying hot functions; avoid `-gcflags` churn without CI discipline |

**External catalog (detail per topic):** [goperf.dev — Common Go Patterns for Performance](https://goperf.dev/01-common-patterns/)

---

## exarp-go documentation map

| Document | Focus |
|----------|--------|
| [research/STABILITY_AND_PERFORMANCE_REMAINING.md](research/STABILITY_AND_PERFORMANCE_REMAINING.md) | Queue DB init, tool semaphore vs config, large-repo walks, locking tests |
| [PERFORMANCE_OPTIMIZATION_RESEARCH.md](PERFORMANCE_OPTIMIZATION_RESEARCH.md) | Scorecard multilang walks, testing validation walks, maintenance batch deletes |
| [PERFORMANCE_FILE_CACHE.md](PERFORMANCE_FILE_CACHE.md) | `internal/cache` mtime + TTL file cache |
| [OPTIMIZATION_RESULTS.md](OPTIMIZATION_RESULTS.md) | Benchmarks: task levels, duplicate detection, median |
| [SCORECARD_PERFORMANCE_ANALYSIS.md](SCORECARD_PERFORMANCE_ANALYSIS.md) | Historical scorecard latency (walks + external commands) |
| [CONTEXT_REDUCTION_FOLLOWUP_TASKS.md](CONTEXT_REDUCTION_FOLLOWUP_TASKS.md) | Token/context size (operator UX, not CPU) |

---

## Suggested workflow

1. Reproduce slowness with **`time`**, **`go test -bench`**, or **`pprof`** on a representative workload.
2. Check **stability/performance inventory** above for known hotspots.
3. Apply patterns from the **quick reference table** that match the profile (alloc-heavy vs I/O vs contention).
4. Re-run **`make test`** / **`make lint`**; add or extend benchmarks if the path is regression-prone.
