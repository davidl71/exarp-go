# Shared Patterns in Task-Related Tools

Task-related tools (`task_workflow`, `task_analysis`, `task_discovery`, `estimation`) share several patterns. This doc summarizes them for consistency and future refactors.

## 1. Handler entry (handlers.go)

- **Parse:** `Parse*Request(args)` → `req, params, err`
- **Protobuf:** If `req != nil`, set `params = *RequestToParams(req)` and `request.ApplyDefaults(params, defaults)`
- **Dispatch:** Call native handler with `(ctx, params)` or `(ctx, projectRoot, params)` (estimation)

## 2. Project root resolution

- **Task tools (workflow, analysis, discovery):** Resolve inside native handler with `FindProjectRoot()` from `todo2_utils.go` (looks for `.todo2`, respects `PROJECT_ROOT` env).
- **Estimation:** Resolved in `handlers.go` with `FindProjectRoot()` and passed into native handler (aligned with other task tools as of 2026-01-28).

## 3. Native handler structure

- Get `projectRoot` (via `FindProjectRoot()` in task tools; passed in for estimation).
- Read `action, _ := params["action"].(string)`; default if empty.
- `switch action { case "X": return handle...(ctx, projectRoot, params) }`.

## 4. Todo2 / database access

- **Load/save (compat):** `LoadTodo2Tasks(projectRoot)`, `SaveTodo2Tasks(projectRoot, tasks)` from `todo2_utils.go`.
- **Direct DB (when available):** `database.GetDB()`, then `database.ListTasks`, `GetTask`, `UpdateTask`, `CreateTask`, `AddComments`, `DeleteTask` with `TaskFilters`.
- **Pattern:** Prefer DB for filtered/single-task ops; use LoadTodo2Tasks/SaveTodo2Tasks when loading/saving full set or when DB may be unavailable.

## 5. FM (Apple Foundation Models) usage

- **Check:** `FMAvailable()` before using Apple FM.
- **Generate:** `DefaultFMProvider().Generate(ctx, prompt, maxTokens, temperature)`.
- **Used in:** task_discovery (semantic extraction), task_workflow (clarity), task_analysis (hierarchy/classify), estimation (estimate action).
- **Fallback:** When `!FMAvailable()`, tools either return a clear error (task_analysis hierarchy) or use statistical/non-FM path (estimation, task_discovery).

## 6. CGO vs nocgo split

- **With Apple FM:** `*_native.go` (build tag `darwin && arm64 && cgo`).
- **Without:** `*_native_nocgo.go` (build tag `!(darwin && arm64 && cgo)`).
- **Shared logic:** `*_shared.go` or `*_common.go` (no FM-specific code so both builds compile).

## 7. Cross-tool reuse

- **task_workflow** calls estimation logic: `handleEstimationNative(ctx, projectRoot, estimationParams)` from `task_workflow_common.go` when adding estimates to tasks (no MCP round-trip).
- **Shared types:** e.g. `EstimationResult`, `Todo2Task`; helpers in `todo2_utils.go`, `task_discovery_common.go`, `task_analysis_shared.go`, `estimation_shared.go`.

## 8. File layout (task-related)

| Concern              | Files |
|----------------------|--------|
| Todo2 + project root | `todo2_utils.go` (FindProjectRoot, LoadTodo2Tasks, SaveTodo2Tasks, SyncTodo2Tasks) |
| Task workflow        | `task_workflow_common.go`, `task_workflow_native.go`, `task_workflow_native_nocgo.go` |
| Task analysis        | `task_analysis_shared.go` |
| Task discovery       | `task_discovery_common.go`, `task_discovery_native.go`, `task_discovery_native_nocgo.go` |
| Estimation           | `estimation_shared.go`, `estimation_native.go`, `estimation_native_nocgo.go` |
| Handlers + parse     | `handlers.go`, `protobuf_helpers.go` |

## Summary

- **Entry:** Parse request → params + defaults → native handler.
- **Project root:** `FindProjectRoot()` in task tools; estimation uses `security.GetProjectRoot(".")` in handler (could align to FindProjectRoot).
- **Data:** LoadTodo2Tasks/SaveTodo2Tasks + database package when DB available.
- **FM:** FMAvailable() then DefaultFMProvider().Generate; graceful fallback per tool.
- **Structure:** action dispatch; shared types and helpers in `*_shared.go` / `*_common.go`; CGO/nocgo split where Apple FM is used.
