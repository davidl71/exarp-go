# TaskStore Migration Scope

Documents remaining TaskStore migration candidates and code that is intentionally **not** a migration target.

---

## Project Root Unification (T-1770831825325)

**Canonical:** `internal/projectroot` with configurable markers.

| Function | Markers | Use |
|----------|---------|-----|
| `projectroot.Find()` | .exarp, .todo2 | Exarp project (PROJECT_ROOT env + cwd walk) |
| `projectroot.FindFrom(path)` | .exarp, .todo2 | Exarp from specific path |
| `projectroot.FindGoMod(path)` | go.mod | Generic Go project root |
| `projectroot.FindFromWithMarkers(path, markers)` | custom | Pluggable markers |

- **tools.FindProjectRoot()** → `projectroot.Find()`
- **config.FindProjectRoot()** → `projectroot.Find()` (returns cwd on error)
- **security.GetProjectRoot(path)** → `projectroot.FindGoMod(path)` (unified; was mcp-go-core go.mod)

## Remaining Migration Tasks

| Task ID | Description | Priority |
|---------|-------------|----------|
| T-1770831732740 | Migrate `internal/resources/tasks.go` to TaskStore | medium |
| T-1770831733303 | Migrate `todo2_utils` GetSuggestedNextTasks and WriteTodo2Overview to TaskStore | medium |
| T-1770831733871 | Migrate `todo2_utils` GetTaskByID and UpdateTaskStatus to TaskStore | low |

See `.todo2/` for full task details.

---

## Not Candidates (By Design)

Code that uses `LoadTodo2Tasks` or `SaveTodo2Tasks` but is **not** a TaskStore migration target:

### 1. `internal/tools/task_store.go`

**Why:** This file **implements** the DefaultTaskStore interface. It uses `LoadTodo2Tasks` and `SaveTodo2Tasks` internally as the storage primitives. Migration would be circular—TaskStore is the abstraction that wraps these functions.

### 2. `internal/tools/todo2_utils.go` (LoadTodo2Tasks, SaveTodo2Tasks definitions)

**Why:** These are the **primitives**—the functions that TaskStore and database fallback logic call. They are not migration targets; they are the low-level implementation.

### 3. `internal/tools/estimation_shared.go` (line ~561)

**Why:** Contains only a **comment** referencing the LoadTodo2Tasks pattern (e.g., "Try database first (same pattern as LoadTodo2Tasks)"). No direct call to migrate.

### 4. Test Files

**Why:** Tests use `LoadTodo2Tasks` for setup, assertions, or temp-directory state. Migrating tests to TaskStore would add complexity without clear benefit. Low priority; can remain as-is unless consistency is desired.

| File | Usage |
|------|-------|
| `task_workflow_test.go` | Setup/assertions with tmpDir |
| `critical_path_test.go` | Load tasks for test |
| `protobuf_integration_test.go` | Skip if no .todo2 |
| `plan_sync_test.go` | Comment about .todo2 dir |
| `infer_task_progress_test.go` | Comment about empty state |

---

## Migration Pattern

For code that **is** a candidate:

- **Tool handlers (with context):** Use `getTaskStore(ctx)` and `store.ListTasks`, `store.GetTask`, `store.UpdateTask`.
- **Resources / standalone utils (projectRoot only):** Use `NewDefaultTaskStore(projectRoot)` and the same store methods.

The goal is one code path (TaskStore) instead of duplicate DB-then-JSON logic.
