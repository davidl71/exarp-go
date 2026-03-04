# Dead code report (deadcode -test ./...)

Generated from `deadcode -test ./...`. "Unreachable" = not reachable from `main` or `*_test.go` entry points.

## Summary

- **Total unreachable items:** 52 (functions/methods).
- **Keep:** Public API, test helpers, or reflection/interface entry points.
- **Candidates for removal:** Internal helpers that have been superseded by proto/other paths.

## By package

### internal/api
- `GetProjectRoot` — API context helper; consider keeping if HTTP API is used.

### internal/cli
- `tui3270State.sessionNames`, `model.taskParentMap` — TUI internals; keep if TUI is in use.

### internal/config
- `LoadConfigYAML`, `detectConfigFormat` — legacy YAML-only path; remove if only Load/LoadConfig is used.

### internal/database
- `GetAgentIDFromSession`, `CleanupExpiredLocksWithReport`, `RunMigrations`, `GetCurrentDriver`, `GetDialect`, `GetTasksByPriority`, `BatchClaimTasks`, `GetTagsForTask` — some are public API; verify before removing.

### internal/logging
- `Error` — logger method; may be used via interface.

### internal/security
- `AllowRequest` — rate limit; may be used by middleware.

### internal/tools
- ~~`loadInProgressTasks`~~ — **removed** (legacy wrapper; callers use `loadTasksByStatus`).
- ~~`aggregateProjectData`~~ — **removed** (map-based overview; report uses `aggregateProjectDataProto`).
- ~~`formatOverviewText`, `formatOverviewMarkdown`, `formatOverviewHTML`~~ — **removed** (report uses `formatOverviewTextProto` / proto formatters).
- ~~`getFloatParam`~~ — **removed** (unused).
- ~~`getProjectRoot` (scorecard_go_format)~~ — **removed** (unused).
- `CursorIgnoreGenerator.hasFile`, `discoverTagsFromMarkdown`, `UpdatePlanningDocWithTaskRefs`, `UpdateTaskStatus` — internal helpers; verify no dynamic calls.
- `HealthDataToProto`, `ProjectOverviewDataToProto`, `GitToolsRequestToParams` — proto helpers; may be called from generated or other packages.
- `handleInferSessionModeNative` — session mode; ensure dispatched from handlers.
- **llamacpp_model_manager.go** — whole file unreachable (model manager not wired); remove or wire.
- **mlx_native_nocgo.go** — `MLXNativeAvailable`; stub for non-CGO builds.

### internal/utils
- `TaskLock`, `StateFileLock` — filelock; may be used by migration or legacy code.

### tests/fixtures
- `TestContext`, `JSONRPCRequest`, `JSONRPCResponse`, `JSONRPCError` — test helpers; ensure tests use them (or remove if obsolete).

### internal/web
- `FS` — embedded FS; may be used by HTTP server.

## Recommended next steps

1. ~~**Done:** Remove `loadInProgressTasks` (infer_task_progress.go).~~
2. ~~**Done:** Remove `aggregateProjectData` and `formatOverviewText` / `formatOverviewMarkdown` / `formatOverviewHTML` (report uses proto path). Remove `getFloatParam` and `getProjectRoot` (scorecard_go_format).~~
3. **Optional:** Remove or wire `llamacpp_model_manager.go` and related unreachable functions.
4. **Audit:** For each remaining item in deadcode output, confirm "keep as API/test" or "safe to remove."
