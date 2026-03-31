# Performance guide (exarp-go)

**Last updated:** 2026-03-30

Use this page as the **index** for performance work: general Go patterns (external reference), how they apply to this repo, and links to detailed notes.

**Rule:** Confirm bottlenecks with **benchmarks** and **`pprof`** (CPU + heap) before micro-optimizing. The MCP server is often **I/O- and SQLite-bound**, not CPU-bound.

---

## AI / agent: commands to measure before optimizing

**Context:** Run from **exarp-go repo root**. Use **`make b`** (or **`make silent`**) to build `./bin/exarp-go`. For Go tests/benches, prefer repo **`Makefile` targets** where they exist; otherwise use the `go test` lines below (they match `internal/database/tasks_crud_bench_test.go`).

### Hints (read first)

- **One-shot CLI** (`task list` + `-cpuprof`) captures **startup + one query**; samples are sparse. Prefer **`go test -bench`** on the hot package for stable before/after numbers.
- **Allocations:** `-benchmem` on benchmarks, or **`-memprofile`** + `go tool pprof -sample_index=alloc_space`.
- **SQLite / disk:** not CPU pprof — use **`health` `action=database`** (status, WAL, size) or filesystem `du` on `.todo2/`.
- **`PROJECT_ROOT`:** set when the workload should use another project’s `.todo2` (CLI finds DB via project root).

### Benchmarks (regression-style)

```bash
# Tool handlers and related (default bench suite; includes -benchmem)
make bench
# Equivalent: see Makefile target go-bench (CGO_ENABLED=0, skips unit tests via -run=^$, ./internal/tools/...)
```

### Database CRUD (CPU + heap profiles)

```bash
mkdir -p logs
CGO_ENABLED=0 go test -c -o logs/database.test ./internal/database/

# CPU (while iterating benchmarks)
# Bench names match tasks_crud_bench_test.go (batch is BenchmarkBatchUpdateTaskStatus_64).
CGO_ENABLED=0 go test -run='^$' -bench='Benchmark(CreateTask|GetTask|UpdateTask|DeleteTask|BatchUpdateTaskStatus_64)' \
  -cpuprofile=logs/crud_cpu.pprof ./internal/database/
go tool pprof -text -nodecount=40 logs/database.test logs/crud_cpu.pprof

# Heap / allocations
CGO_ENABLED=0 go test -run='^$' -bench='Benchmark(CreateTask|GetTask|UpdateTask|DeleteTask|BatchUpdateTaskStatus_64)' \
  -benchmem -memprofile=logs/crud_mem.pprof ./internal/database/
go tool pprof -text -sample_index=alloc_space -nodecount=40 logs/database.test logs/crud_mem.pprof
```

**Hint:** If symbols look wrong after code changes, re-run **`go test -c -o logs/database.test ./internal/database/`** before `go tool pprof`. For a quick peek without a local binary, **`go tool pprof -text logs/crud_cpu.pprof`** often still ranks hot functions.

### Reprofile and compare (CRUD regression)

Script: **`scripts/crud_reprofile_compare.sh`** (same bench regex as above; **`CGO_ENABLED=0`**).

**Benchmarks + `benchstat`**

1. On a known-good commit (or before a change): `make crud-bench-reprofile` then `make crud-bench-save-baseline` → writes `logs/crud_bench_baseline.txt`.
2. After your change: `make crud-bench-reprofile` → writes `logs/crud_bench_latest.txt` and, if the baseline file exists, prints **`benchstat` baseline vs latest**.
3. Install **`benchstat`** once: `go install golang.org/x/perf/cmd/benchstat@latest`

Optional env: **`CRUD_BENCH_COUNT`** (default `5`), **`CRUD_BENCH_TIME`** (e.g. `2s` → `-benchtime=2s`). **`make crud-bench-compare`** re-runs **`benchstat`** only (no `go test`).

**CPU profiles + `pprof -base`**

- `make crud-pprof` — builds **`logs/database.test`**, runs the CRUD bench suite once with timestamped **`logs/crud_cpu_<UTC>.pprof`** and **`logs/crud_mem_<UTC>.pprof`**, then prints short **`-text`** summaries.
- Compare a new capture to an older CPU profile (same **test binary** build is best for stable symbols):

```bash
bash scripts/crud_reprofile_compare.sh pprof-diff logs/crud_cpu_OLD.pprof logs/crud_cpu_NEW.pprof
# optional 4th arg: path to test binary (default logs/database.test)
```

### CLI: backlog read / `task_workflow` path (CPU)

```bash
make b
./bin/exarp-go -cpuprof=/tmp/cli_tasklist_cpu.pprof task list
# Optional: limit noise — ./bin/exarp-go -cpuprof=... task list --limit 50
# Other project DB:
#   PROJECT_ROOT=/path/to/project ./bin/exarp-go -cpuprof=... task list

go tool pprof -text -nodecount=40 ./bin/exarp-go /tmp/cli_tasklist_cpu.pprof
```

### MCP tool: one-shot `task_workflow list` (CPU)

This captures **startup + one list call** (still sparse, but useful as a quick sanity check).
For stable before/after numbers, prefer `go test -run=^$ -bench=...` in `./internal/tools`.

```bash
make b
./bin/exarp-go -cpuprof=/tmp/mcp_taskworkflow_list_cpu.pprof \
  -tool task_workflow \
  -args '{"action":"list","output_format":"json","compact":true,"limit":200}' >/dev/null
go tool pprof -text -nodecount=40 ./bin/exarp-go /tmp/mcp_taskworkflow_list_cpu.pprof
```

### MCP stdio session (long-running CPU)

```bash
./bin/exarp-go -cpuprof=/tmp/mcp_stdio_cpu.pprof
# interact, then exit client; analyze as above with ./bin/exarp-go
```

### Notes: reducing allocs/GC in `task_workflow list`

- `task_workflow list` supports `include_locks=true` (default false). Avoid locks unless needed.
- For SQL field/row tightening opportunities, see `docs/TASK_WORKFLOW_LIST_SQL_AUDIT.md`.

### SQLite / disk health (not CPU)

```bash
./bin/exarp-go -tool health -args '{"action":"database","operation":"status"}'
# Optional: checkpoint / vacuum / analyze — see health tool schema (operation field)
```

### Ad hoc UpdateTask micro-bench (binary entry)

```bash
# 100x UpdateTaskFields for a real task ID; writes /tmp/cpu.prof, /tmp/mem.prof, /tmp/goroutine.prof
./bin/exarp-go -benchprof=T-YOUR-ID
```

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
