# Task Backend Usage by Tool

This document catalogs which tools and modules use which task backend access pattern.

## Access Patterns

| Pattern | Description |
|---------|-------------|
| **TaskStore** | `store.ListTasks/GetTask/UpdateTask/CreateTask/DeleteTask` - DB-first with JSON fallback via `NewDefaultTaskStore()` |
| **Direct DB** | `database.ListTasks/GetTask/UpdateTask/CreateTask/DeleteTask` - Direct database access (no JSON fallback) |
| **LoadTodo2Tasks** | `LoadTodo2Tasks/SaveTodo2Tasks` - DB-first with JSON fallback (standalone functions) |

---

## Tools Using TaskStore (Recommended ✓)

These tools use the `TaskStore` interface which provides automatic DB-first, JSON-fallback:

| Tool/Module | File | Operations |
|-------------|------|------------|
| session (handoff) | `session_handoff.go` | ListTasks |
| session (restore) | `session_handoff.go` | ListTasks, UpdateTask, CreateTask |
| task_workflow (CRUD) | `task_workflow_crud.go` | ListTasks, GetTask, UpdateTask |
| task_workflow (actions) | `task_workflow_actions.go` | ListTasks, GetTask, UpdateTask |
| task_workflow (maintenance) | `task_workflow_maintenance.go` | ListTasks, DeleteTask |
| task_workflow (maintenance helpers) | `task_workflow_maintenance_helpers.go` | ListTasks, UpdateTask |
| task_workflow (AI run) | `task_workflow_ai_run.go` | GetTask |
| task_workflow (create AI) | `task_workflow_create_ai.go` | ListTasks, CreateTask, UpdateTask |
| plan_sync | `plan_sync.go` | ListTasks, GetTask, CreateTask, UpdateTask |
| task_execute | `task_execute.go` | GetTask |
| infer_task_progress | `infer_task_progress.go` | ListTasks, GetTask, UpdateTask |
| task_analysis (shared) | `task_analysis_shared.go` | ListTasks, UpdateTask, DeleteTask |
| task_analysis (deps) | `task_analysis_deps.go` | ListTasks |
| task_analysis (deps analysis) | `task_analysis_deps_analysis.go` | ListTasks, UpdateTask |
| task_analysis (tags) | `task_analysis_tags.go` | ListTasks, UpdateTask |
| task_analysis (suggest deps) | `task_analysis_suggest_deps.go` | ListTasks |
| task_analysis (graph) | `task_analysis_graph.go` | ListTasks |
| session_mode_inference | `session_mode_inference.go` | ListTasks |
| report_data | `report_data.go` | ListTasks |
| report_plan | `report_plan.go` | ListTasks |
| report_plan_generate | `report_plan_generate.go` | ListTasks |
| git_tools_actions | `git_tools_actions.go` | ListTasks, UpdateTask |
| estimation_analytics | `estimation_analytics.go` | ListTasks |
| alignment_analysis | `alignment_analysis.go` | ListTasks |
| automation_scheduled | `automation_scheduled.go` | ListTasks |
| task_discovery (common) | `task_discovery_common.go` | ListTasks |
| task_discovery (native) | `task_discovery_native.go` | ListTasks |
| task_discovery (nocgo) | `task_discovery_native_nocgo.go` | ListTasks |

---

## Tools Using Direct Database (Not Recommended)

These tools directly call `database.*` functions, bypassing the TaskStore interface:

| Tool/Module | File | Reason |
|-------------|------|--------|
| task_workflow_actions | `task_workflow_actions.go` | Uses `GetTasksByStatus` for Review tasks |
| task_workflow_maintenance | `task_workflow_maintenance.go` | Legacy - mixes with TaskStore |
| task_workflow_maintenance_helpers | `task_workflow_maintenance_helpers.go` | Legacy - mixes with TaskStore |
| todo2_db_adapter | `todo2_db_adapter.go` | DB-only adapter (intentional - DB sync) |
| todo2_utils | `todo2_utils.go` | Sync function uses direct DB |
| session_assignee | `session_assignee.go` | Uses `GetTasksByStatus` helper |
| recommend | `recommend.go` | Single task lookup |
| report_plan | `report_plan.go` | Updates task metadata |
| attribution_check | `attribution_check.go` | Creates tasks |
| alignment_analysis | `alignment_analysis.go` | Creates tasks |

### Issues with Direct Database Access

1. **No JSON fallback** - Fails if DB is unavailable
2. **Inconsistent behavior** - Some tools use TaskStore, some don't
3. **Maintenance burden** - Duplicated fallback logic possible

---

## Tools Using LoadTodo2Tasks/SaveTodo2Tasks

These use the standalone functions (DB-first with JSON fallback):

| Tool/Module | File | Usage |
|-------------|------|-------|
| session_handoff | `session_handoff.go` | Legacy fallback for tasks without ProjectID |
| task_store | `task_store.go` | TaskStore implementation |
| task_workflow_create_ai | `task_workflow_create_ai.go` | Persist corrected tasks |

---

## Tools/Resources NOT Accessing Tasks

These tools don't interact with tasks at all:

| Tool/Module | Type |
|-------------|------|
| session (prime, prompts) | No task access |
| setup_hooks | No task access |
| testing | No task access |
| ollama | No task access |
| mlx | No task access |
| apple_foundation | No task access |
| context | No task access |
| memory | No task access |
| memory_maint | No task access |
| generate_config | No task access |
| security | No task access |
| lint | No task access |
| health | No task access (except git) |
| report (scorecard, overview) | Uses TaskStore |
| task_discovery | Uses TaskStore |
| task_analysis (most) | Uses TaskStore |

---

## Recommendations

1. **Prefer TaskStore** - Use `NewDefaultTaskStore(projectRoot)` and call `store.ListTasks()`, etc.
2. **Migrate direct DB calls** - Tools using `database.*` directly should use TaskStore
3. **Keep LoadTodo2Tasks** - Still useful for one-off loads/saves outside TaskStore context
4. **Avoid new direct DB access** - Any new code should use TaskStore

### Migration Example

```go
// Before (direct DB - not recommended)
tasks, err := database.ListTasks(ctx, filters)

// After (TaskStore - recommended)
store := NewDefaultTaskStore(projectRoot)
tasks, err := store.ListTasks(ctx, filters)
```
