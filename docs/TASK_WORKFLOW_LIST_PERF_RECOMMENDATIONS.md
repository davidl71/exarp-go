## Task Workflow List — Perf Recommendations (SQL + Hashing)

This doc consolidates the current performance recommendations for `task_workflow` `action=list`,
with emphasis on **reducing allocations / GC** for MCP stdio consumers.

### Where the time/allocs go (today)

Primary costs for `list` are typically:
- **Wide task row loads** (including large text + metadata)
- **Tag/dependency fanout** (extra queries + slice growth)
- **JSON response construction** (`map[string]interface{}` + per-task slices)

Benchmarks are the source of truth. See:
- `docs/PERFORMANCE_GUIDE.md` (how to measure)
- `docs/TASK_WORKFLOW_LIST_SQL_AUDIT.md` (what SQL we run today)

---

## SQL recommendations

All SQL currently used by the list path is documented in:
- `docs/TASK_WORKFLOW_LIST_SQL_AUDIT.md`

Recommended changes (ordered by ROI):

1) **Add a DB “summary list” query**
   - Default list response (especially for MCP) should not require:
     - `long_description`
     - `metadata` / `metadata_protobuf`
   - Summary columns should be:
     - `id`, `name`, `status`, `priority`, `created_at/last_modified` (plus optional `parent_id`)
     - include `content` only if `name` is empty (or behind `include_content=true`)

2) **Make related data opt-in**
   - Tags, dependencies, locks all add memory churn and DB work.
   - Prefer explicit flags:
     - `include_tags`
     - `include_dependencies`
     - `include_locks` (already exists; default false)

3) **Lock lookup should be filtered in SQL**
   - Avoid “load all active locks then filter”.
   - Query `WHERE id IN (?) AND lock_until >= ?` when `include_locks=true`.

4) **Push filters into SQL**
   - `handleTaskWorkflowList` currently filters many dimensions in memory.
   - Passing `TaskFilters` down to `store.ListTasks(ctx, filters)` avoids loading unrelated rows.
   - **Implemented (partial)**:
     - When `task_id` is not set and `order` is not `execution|dependency`, we now push down:
       - default open-only (`statuses IN (Todo, In Progress, Blocked)`)
       - `status`, `priority`, `tag`
     - When `order=execution|dependency`, we still load the full task set (project-scoped)
       so dependency ordering remains correct, then apply filters in memory.

### Why this helps GC (and why it compounds)

`task_workflow list` is called frequently (MCP stdio clients tend to poll). The expensive part is
often not “filtering” itself; it’s **the work done before you can filter**:

- decoding wide task rows
- allocating per-task slices (tags, deps)
- building the output maps/slices

By pushing `status/priority/tag` down into SQLite, the handler **avoids touching tasks that won’t
be returned**, which reduces allocations roughly in proportion to “how much of the DB you can skip”.

This compounds most when:
- your DB has many closed tasks (Done/Cancelled)
- clients repeatedly request the default open-only list
- clients filter by `status`/`tag`/`priority` (common operator workflows)

### Enum-first filtering (status_enum/priority_enum)

For SQLite-backed tasks, prefer filtering on integer enum columns (`status_enum`, `priority_enum`)
when typed filters are available, and treat string fields as boundary/compat-only.

See `docs/ENUM_FIRST_SQLITE_SCHEMA.md`.

---

## Hashing / checksum recommendations

Hashing helps when you want **fast change detection** or **conditional fetch**, not faster DB lookup.
For this project, hashes are most useful as **change tokens for large fields**.

### Prefer “version tokens” first

SQLite tasks already have a `version` column (optimistic locking) and timestamps (`last_modified`).
These are usually better than hashes for “did the task change?” because they’re cheap and human-inspectable.

### Where hashes could help (if needed)

1) **Conditional fetch of large fields**
   - In summary list responses, return:
     - `version` (or `last_modified`) as the task-wide change token
     - optionally `long_description_hash` and/or `metadata_hash` as *field-level* tokens
   - Clients can fetch full details only when tokens change.

2) **Payload dedup / caching**
   - For derived resources (execution packs, briefing payloads), a `payload_hash` can serve as a stable cache key.

### Where hashes are not worth it

- Hashing `task_id` does not beat indexed lookup and adds CPU.
- Replacing `version`/timestamps with hashes is usually worse operationally.

### Suggested minimal change-token shape

- Summary list:
  - `id`, `name`, `status`, `priority`, `last_modified`, `version`
- Detail fetch (by id):
  - full task with `long_description`, `metadata`, tags, deps, recent runs if requested
- Optional:
  - `long_description_hash` / `metadata_hash` only if we need field-level conditional fetch

---

## Suggested next steps (concrete)

1) **Implement `ListTaskSummaries` in `internal/database`**
   - lightweight struct + SQL selecting minimal columns
   - no metadata decode by default

2) **Add `task_workflow list` flags + defaults**
   - default MCP list: summary-only (name-first)
   - add `include_content/include_tags/include_dependencies` toggles

3) **Tighten lock query**
   - implement “locks for IDs” SQL and use it when `include_locks=true`

4) **Re-run benchmark + pprof**
   - `BenchmarkHandleTaskWorkflowList` before/after
   - keep a short note in `docs/OPTIMIZATION_RESULTS.md` with deltas

5) **Backfill empty `name`**
   - Tool maintenance action: `task_workflow` `action=fix_empty_names` (supports `dry_run`)
   - Goal: enable name-first summary queries and keep list UIs stable

