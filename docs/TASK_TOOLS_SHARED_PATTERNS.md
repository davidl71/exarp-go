# Shared Patterns in Task-Related Tools

Task-related tools (`task_workflow`, `task_analysis`, `task_discovery`, `estimation`) share several patterns. This doc summarizes them for consistency and future refactors.

## 1. Handler entry (handlers.go) via WrapHandler

- **Parse:** `Parse*Request(args)` → `req, params, err`
- **CRITICAL:** Defaults must be applied for BOTH protobuf AND JSON paths (not just protobuf):
  ```go
  if req != nil {
      params = RequestToParams(req)
  }
  // Apply defaults for both paths
  if defaults != nil {
      framework.ApplyDefaults(params, defaults)
  }
  ```
  See `handlers_wrap.go` for the correct implementation.
- **Dispatch:** Call native handler with `(ctx, params)` or `(ctx, projectRoot, params)` (estimation)

## 2. Project root resolution

- **Task tools (workflow, analysis, discovery):** Resolve inside native handler with `FindProjectRoot()` from `todo2_utils.go` (looks for `.todo2`, respects `PROJECT_ROOT` env).
- **Estimation:** Resolved in `handlers.go` with `FindProjectRoot()` and passed into native handler (aligned with other task tools as of 2026-01-28).

## 3. Native handler structure

- Get `projectRoot` (via `FindProjectRoot()` in task tools; passed in for estimation).
- Read `action, _ := params["action"].(string)`; default if empty.
- `switch action { case "X": return handle...(ctx, projectRoot, params) }`.

## 4. Todo2 / database access

- **Canonical task command path:** user-facing task CRUD goes through `task_workflow`.
- **Tool-layer CRUD path:** `internal/tools` should prefer `getTaskStore(ctx)` / `TaskStore` for ordinary task reads and writes.
- **Load/save (compat):** `LoadTodo2Tasks(projectRoot)`, `SaveTodo2Tasks(projectRoot, tasks)` from `todo2_utils.go`.
- **Direct DB is now the exception:** use `database.*` directly only for DB-native capabilities such as locks, execution runs, comments, verifications, or migration/repair helpers.
- **Pattern:** if the operation is “load/update/create/delete a task”, default to `TaskStore`; if it is inherently SQLite-specific, use `database.*`.

## 5. FM / local LLM usage

- **Primary task-tool abstraction:** `DefaultFMProvider().Generate(ctx, prompt, maxTokens, temperature)`.
- **Routing:** FM chain now prefers the configured/default local provider path (for example Ollama-backed flows) instead of a dedicated Apple Foundation Models tool.
- **Used in:** task_discovery (semantic extraction), task_workflow (clarity), task_analysis (hierarchy/classify), estimation (estimate action).
- **Fallback:** When `!FMAvailable()`, tools either return a clear error or use statistical/non-FM paths depending on the feature.

## 6. Build-tag split

- **Current rule:** prefer portable Go implementations with no platform-specific build split unless the feature truly requires it.
- **If a split is needed:** keep platform- or CGO-specific wiring in `*_native.go` / `*_nocgo.go` or similarly named files, and keep the business logic in `*_shared.go` / `*_common.go`.
- **Current direction:** Apple FM-specific build-tag branches were removed, so task tools should not introduce new darwin-only CRUD or FM paths.

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

- **Entry via WrapHandler:** Parse request → convert proto (if any) → apply defaults for BOTH paths → native handler.
- **Entry manual:** Parse request → params + defaults → native handler.
- **Project root:** `FindProjectRoot()` in task tools; estimation uses `security.GetProjectRoot(".")` in handler (could align to FindProjectRoot).
- **Data:** `task_workflow` is the single task command backend; `TaskStore` is the preferred tool-layer CRUD abstraction.
- **FM:** `FMAvailable()` then `DefaultFMProvider().Generate`; graceful fallback per tool.
- **Structure:** action dispatch; shared types and helpers in `*_shared.go` / `*_common.go`; use build-tag splits sparingly and only for genuinely platform-specific code.
