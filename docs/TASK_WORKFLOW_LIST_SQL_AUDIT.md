## Task Workflow List — SQL Audit (Fields + Tightening Options)

This note is a **quick reference for future profiling runs** when tuning `task_workflow` list performance.
It documents the SQL queries currently used (directly or indirectly) by the list path, and highlights
places where we can **pull fewer fields** or **avoid extra queries**.

### Scope

- **Tool**: `task_workflow`
- **Action**: `list`
- **Primary handler**: `internal/tools/task_workflow_crud.go` → `handleTaskWorkflowList`
- **Primary DB entrypoint**: `internal/database/tasks_list.go` → `ListTasks`

### Perf checklist (quick rerun)

Use this when iterating on SQL or payload shaping so you can quickly confirm impact on **GC pressure**
and wall time.

#### Bench (stable before/after)

```bash
cd /Users/davidl/Projects/mcp/exarp-go

# Focused benchmark for list handler (skips unit tests).
CGO_ENABLED=0 go test -run='^$' -bench=BenchmarkHandleTaskWorkflowList -benchmem -benchtime=3s ./internal/tools
```

Watch for:
- `ns/op` (latency)
- `B/op` and `allocs/op` (GC pressure)

#### One-shot CPU profile (CLI path)

```bash
cd /Users/davidl/Projects/mcp/exarp-go
make b
./bin/exarp-go -cpuprof=/tmp/cli_tasklist_cpu.pprof task list --json --limit 200 >/dev/null
go tool pprof -text -nodecount=40 ./bin/exarp-go /tmp/cli_tasklist_cpu.pprof
```

#### One-shot CPU profile (tool path)

```bash
cd /Users/davidl/Projects/mcp/exarp-go
make b
./bin/exarp-go -cpuprof=/tmp/mcp_taskworkflow_list_cpu.pprof \
  -tool task_workflow \
  -args '{"action":"list","output_format":"json","compact":true,"limit":200}' >/dev/null
go tool pprof -text -nodecount=40 ./bin/exarp-go /tmp/mcp_taskworkflow_list_cpu.pprof
```

#### Flags that materially affect DB work / allocations

- `include_locks=true` (default false): adds lock lookups and output mapping.
- `include_metadata=true`: includes task metadata in output (can be large / decode-heavy).
- `include_full_long_description=true`: prevents truncation in list output.
- `limit`: caps tasks returned (reduces tag/dep fanout and JSON payload size).

### Queries checked (current)

#### 1) Load tasks (base rows)

- **File**: `internal/database/tasks_list.go`
- **Function**: `ListTasks(ctx, filters)`
- **Query**:

```sql
SELECT DISTINCT
  t.id, t.name, t.content, t.long_description, t.status, t.priority, t.completed,
  t.created, t.last_modified, t.completed_at,
  t.metadata, t.metadata_protobuf, t.metadata_format,
  t.parent_id, t.project_id, t.assigned_to, t.host, t.agent, t.version
FROM tasks t
-- optional: INNER JOIN task_tags tt ON t.id = tt.task_id
-- optional: WHERE predicates (status/priority/tag/project/assigned_to/host/agent)
ORDER BY t.created_at DESC
```

**Notes**
- This is intentionally “wide” (loads metadata + long_description for every row) even though most list
  renderings do not need all fields.
- The query uses `DISTINCT` due to optional join on `task_tags`.
- `task_workflow list` now passes `TaskFilters` down to this query for the common case (no `order=execution|dependency`),
  so the DB can filter by `status` (including the default open-only set), `priority`, and `tag` before decoding rows.

#### 2) Batch load tags

- **File**: `internal/database/tasks_list.go`
- **Function**: `ListTasks(ctx, filters)`
- **Query**:

```sql
SELECT task_id, tag
FROM task_tags
WHERE task_id IN (?)
ORDER BY task_id, tag
```

#### 3) Batch load dependencies

- **File**: `internal/database/tasks_list.go`
- **Function**: `ListTasks(ctx, filters)`
- **Query**:

```sql
SELECT task_id, depends_on_id
FROM task_dependencies
WHERE task_id IN (?)
ORDER BY task_id, depends_on_id
```

#### 4) Active locks (optional in `task_workflow list`)

- **File**: `internal/database/lock_monitoring.go`
- **Function**: `GetActiveLocks(ctx)`
- **Query**:

```sql
SELECT id, assignee, assigned_at, lock_until
FROM tasks
WHERE assignee IS NOT NULL
  AND lock_until IS NOT NULL
  AND lock_until >= ?
ORDER BY lock_until ASC
```

**Notes**
- `GetActiveLockMapForTasks` currently calls `GetActiveLocks` (loads **all** active locks) then filters
  in-memory to the requested task IDs.
- In `task_workflow list`, locks are now **opt-in** via `include_locks=true`.

#### 5) Execution runs / verifications / progress (only when listing a specific `task_id`)

- **File**: `internal/database/task_execution.go`
- **Functions**:
  - `ListTaskExecutionRuns(ctx, taskID, status, limit)`
  - `ListTaskVerifications(...)`
  - `ListTaskProgressEntries(...)`
- **Query example** (`ListTaskExecutionRuns`):

```sql
SELECT run_id, task_id, agent_id, host, status, summary, files_touched, commands_run, notes, started_at, ended_at
FROM task_execution_runs
WHERE (? = '' OR task_id = ?)
  AND (? = '' OR status = ?)
ORDER BY started_at DESC
LIMIT ?
```

### Tightening opportunities (recommended)

#### A) Add a “summary rows” query for list

For list views (especially MCP stdio), we can avoid pulling big columns:

- **Avoid**: `t.long_description`, `t.metadata*` for bulk list unless explicitly requested
- **Keep**: `id, name, status, priority, created_at/last_modified` (and optionally `parent_id`)
- **Consider**: include `content` only when `name` is empty, or behind an `include_content=true` flag

Suggested approach:
- Implement a new DB function, e.g. `ListTaskSummaries(ctx, filters, opts)` that returns a lightweight
  struct slice and only queries needed columns.
- Keep existing `ListTasks` for full fidelity (show/detail, exports, migrations).

#### B) Lock lookup: query only requested task IDs

Instead of “load all active locks then filter”, add a query like:

```sql
SELECT id, assignee, assigned_at, lock_until
FROM tasks
WHERE id IN (?)
  AND assignee IS NOT NULL
  AND lock_until IS NOT NULL
  AND lock_until >= ?
```

This reduces rows scanned when there are many locks but you’re listing a small page of tasks.

#### C) Tag/deps load: skip when not requested

For ultra-compact list output, consider an option like `include_tags=false` / `include_dependencies=false`
to avoid the extra `IN (?)` queries entirely.

#### D) Push filters into SQL from tool layer

`handleTaskWorkflowList` currently filters in memory (status/priority/tag/owned_file/openOnly) after loading.
If we pass `TaskFilters` down to `store.ListTasks(ctx, filters)` we can avoid loading unrelated rows.

### Internet references (2026)

These aren’t project-specific, but they back the “select fewer columns into smaller structs” approach:

- `https://jmoiron.github.io/sqlx/` (SQLx docs: mapping query columns into structs)
- `https://dev.to/jones_charles_ad50858dbc0/sqlx-your-go-to-database-toolkit-for-go-developers-53n8`

### Related docs

- `docs/TASK_WORKFLOW_LIST_PERF_RECOMMENDATIONS.md` (SQL + hashing recommendations + next steps)

