# Task Completion Check: Native Go (Default)

## Summary

**`make check-tasks`** and **`make update-completed-tasks`** use the **native Go** implementation by default. They run **`exarp-go -tool infer_task_progress`** and do **not** depend on Python or the agentic-tools MCP server.

## Current Flow (Native Go)

| Makefile target              | Implementation | Dependency |
|-----------------------------|----------------|------------|
| `make check-tasks`          | `exarp-go -tool infer_task_progress` (dry run) | None (Go binary only) |
| `make update-completed-tasks` | Same tool, `auto_update_tasks=true` | None |

- The **infer_task_progress** tool is implemented in **internal/tools/infer_task_progress.go** and **infer_task_progress_evidence.go**.
- It loads **In Progress** tasks only from the database (or `.todo2/state.todo2.json` fallback), gathers codebase evidence (file walk, snippets), scores tasks with heuristics and optionally Apple FM when available, and optionally updates task status to Done and writes a markdown report.
- Reports are written to **docs/TASK_COMPLETION_CHECK.md** (dry run) and **docs/TASK_COMPLETION_UPDATE.md** (update run) when the Makefile targets are used.

## What Native Go Implements

| Area | What it does |
|------|-------------------------------|
| **Evidence** | File walk under project root with configurable `scan_depth` and `file_extensions`; path and snippet collection (see `infer_task_progress_evidence.go`) |
| **Scoring** | Heuristic keyword/path matching; optional FM-based scoring when `FMAvailable()` (see `infer_task_progress.go`) |
| **Output** | JSON result with `inferred_results`, `tasks_updated`; optional markdown report when `output_path` is set |
| **Apply** | When `auto_update_tasks=true` and `dry_run=false`, calls `database.UpdateTask` (or `SaveTodo2Tasks` when DB unavailable) to set status Done and completed |

## Python / Legacy Path (Deprecated)

The previous flow used **project_management_automation/tools/auto_update_task_status.py**, which called the **agentic-tools MCP** `infer_task_progress` tool. That path is **deprecated**. The Makefile no longer invokes Python for task completion check; the native Go tool is the only path used by `make check-tasks` and `make update-completed-tasks`.

If you need the same contract (inputs/outputs) from Python or another MCP client, call the **exarp-go** MCP tool **infer_task_progress** instead of agentic-tools.

## Reference

- **Migration plan (completed):** [TASK_COMPLETION_INFERENCE_NATIVE_MIGRATION.md](TASK_COMPLETION_INFERENCE_NATIVE_MIGRATION.md)
- **Tool registration:** internal/tools/registry.go, handlers.go
- **Evidence + handler:** internal/tools/infer_task_progress.go, infer_task_progress_evidence.go
