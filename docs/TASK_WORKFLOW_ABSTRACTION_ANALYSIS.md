# Task Workflow and Todo2 Abstraction Layer Analysis

## Abstraction Layer (todo2_utils.go)

| Function | Behavior |
|----------|----------|
| **LoadTodo2Tasks(projectRoot)** | Load from DB first; on failure, load from `.todo2/state.todo2.json`. Returns one source only. |
| **SaveTodo2Tasks(projectRoot, tasks)** | Save to DB first; if DB save succeeds, return (no JSON write). On DB failure, save to JSON only. |
| **SyncTodo2Tasks(projectRoot)** | Load from **both** DB and JSON, merge by task ID (DB wins), then save to **both** (DB gets list minus AUTO-*, JSON gets full list). |

## Why Duplicates Reappear After Merge Then Sync

1. **Merge (task_analysis duplicates auto_fix)** calls `SaveTodo2Tasks(projectRoot, tasks)` with the merged list (e.g. 504 tasks).
2. **SaveTodo2Tasks** calls `saveTodo2TasksToDB(tasks)`; it succeeds, so the function returns **without** writing to JSON.
3. JSON still contains the pre-merge 560 tasks (including the 56 duplicate IDs that were removed from the in-memory list and written only to DB).
4. **Sync** runs: loads DB (504) + JSON (560), merges by task ID. The 56 IDs that exist only in JSON (the merged-away duplicates) are added back. Result: 560 tasks. Sync then saves 560 to both stores.

So any mutator that uses **SaveTodo2Tasks** (merge, approve, fix_missing_deps, etc.) updates only the DB when DB is available; JSON stays stale. A later sync merges stale JSON back in and undoes the mutation.

## Fix: Use the Abstraction So Both Stores Stay in Sync

**Option A (implemented):** When DB save succeeds in **SaveTodo2Tasks**, also write the same task list to JSON. That way every mutator keeps both stores in sync; sync no longer reintroduces removed tasks.

- **LoadTodo2Tasks** – unchanged (single source: DB then JSON fallback).
- **SaveTodo2Tasks** – after successful `saveTodo2TasksToDB(tasks)`, call `saveTodo2TasksToJSON(projectRoot, tasks)` so the canonical list is written to both.
- **SyncTodo2Tasks** – unchanged; it remains the path that merges two sources and writes to both (with AUTO-* handling).

Callers (task_analysis merge, task_workflow approve, infer_task_progress apply, etc.) already use **SaveTodo2Tasks**; no caller changes required. After this change, merge → sync keeps 504 tasks in both stores.

## Call Sites Using SaveTodo2Tasks

- **task_analysis_shared.go** – duplicates (merge), tags (consolidate), fix_missing_deps
- **task_workflow_common.go** – approve (batch status update)
- **task_workflow_native.go** – create, approve
- **infer_task_progress.go** – apply inferred completions (JSON fallback path)
- **git_tools.go** – branch/merge task updates

All benefit from SaveTodo2Tasks writing to both when DB is available.
