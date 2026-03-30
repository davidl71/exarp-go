# GetTask / UpdateTask: query shape and tag–dependency write churn

**Task:** Aether Todo2 T-1774888049604959000 (research)  
**Code:** `internal/database/tasks_crud.go`, benchmarks in `tasks_crud_bench_test.go`  
**Related:** `docs/OPTIMIZATION_RESULTS.md` §4

## Current behavior (2026-03)

### GetTask

- Single `GetContext` round-trip to SQLite.
- Core row plus aggregated `tags_json` / `deps_json` via correlated subqueries (`sqlTaskAggJSON` in `tasks_crud.go`).
- **Implication:** Further “JOIN consolidation” is low ROI unless profiling shows driver/sqlite overhead dominates; the hot cost is more likely JSON parsing, string allocs, and metadata deserialization.

### UpdateTask

- Transaction: optimistic `UPDATE` on `tasks`, then **unconditional** `DELETE` all `task_tags` and `task_dependencies` for the task, then batch `INSERT` for current tags and deps.
- **Implication:** Correct and simple. Churn scales with tag/dep count on *every* full update, even when only `status`/`content` changed and relations are identical.

### Existing fast paths (callers)

- `BatchUpdateTaskStatus` — status/priority-only updates without touching tags/deps.
- `UpdateTaskFields` — partial field updates when applicable.
- Tooling should prefer these when relations do not change.

## Recommendations

1. **Profiling first** — Run `go test -bench=BenchmarkGetTask -benchmem` and `BenchmarkUpdateTask` (and a CPU profile under realistic MCP `task_workflow` update patterns). Promote optimizations only if allocs or wall time show up in user-visible paths.
2. **Tag/dep diff (optional)** — If profiles show `UpdateTask` write amplification: load existing tag/dep sets (or keep sorted slices on the in-memory task), compute set diff, `DELETE` only removed keys and `INSERT` only new keys. Trade: more CPU/branches vs fewer writes. Best when median tag count is small but updates are frequent.
3. **GetTask** — Prefer reducing parse/allocs (e.g. reuse buffers, lazy metadata parse) over splitting/combining SQL unless benchmarks prove otherwise.

## Follow-up implementation task

Open a separate task only if pprof/benchmarks show ≥ meaningful threshold (team-defined) on production-adjacent workloads (e.g. bulk status updates already use batch path; focus on mixed metadata+tag updates).
