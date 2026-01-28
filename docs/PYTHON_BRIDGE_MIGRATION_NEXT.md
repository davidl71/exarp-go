# Python Bridge → Native Go: Next Migration Steps

**Date:** 2026-01-28  
**Status:** Active  
**Based on:** `MIGRATION_STATUS_CURRENT.md`, `SAFE_PYTHON_REMOVAL_PLAN.md`, codebase grep of `ExecutePythonTool`

---

## Executive Summary

- **Current:** 96% tool coverage (27/28 have native); 1 tool bridge-only (`mlx`). Most tools are **hybrid** (native first, bridge fallback).
- **Next:** Remove optional bridge fallbacks where the native path is complete; then slim the bridge to essential tools only.
- **This pass:** Memory tool made **fully native** (bridge fallback removed). Audit and doc added for future removals.

---

## Audit: Where Bridge Is Still Called

| Tool / area | Call site | Role | Safe to remove? |
|-------------|-----------|------|------------------|
| **memory** | ~~handlers.go~~ | ~~Fallback when native fails~~ | ✅ **Done** – Native-only (2026-01-28). Bridge branch removed from execute_tool.py. |
| **task_discovery** | ~~handlers.go~~ | ~~Fallback when native fails~~ | ✅ **Done** – Native-only (2026-01-28). Bridge branch removed from execute_tool.py. |
| **report** | ~~handlers.go~~ | ~~Fallback when native failed~~ | ✅ **Done** – Native-only (overview, scorecard, briefing, prd). Bridge branch removed from execute_tool.py. |
| **security** | ~~security.go~~ | ~~Fallback when native failed~~ | ✅ **Done** – Native-only (Go projects only; errors otherwise). Bridge branch removed. |
| **task_workflow** | `handlers.go`, `task_workflow_common.go` | Fallback when native fails; **external sync** (agentic-tools) | ❌ Keep – External sync is bridge-only for now. |
| **testing** | ~~testing.go~~ | ~~Fallback when native failed~~ | ✅ **Done** – Native-only (Go projects only; errors otherwise). Bridge branch removed. |
| **lint** | ~~handlers.go~~ | ~~Fallback for non-native linters (e.g. ruff)~~ | ✅ **Done** – Native-only (golangci-lint, go-vet, gofmt, goimports, markdownlint, shellcheck, auto); unsupported linter returns error. Bridge branch removed. |
| **mlx** | `handlers.go`, `insight_provider.go` | Primary (no Go bindings) | ❌ Keep – Intentional bridge-only. |
| **context** | `handlers.go` | Unhandled actions / when native fails | Optional – Could narrow to specific actions. |
| **recommend** | ~~handlers.go~~ | ~~Fallback when native failed~~ | ✅ **Done** – Native-only (model, workflow, advisor via devwisdom-go). Bridge branch removed. |
| **ollama** | `ollama_provider.go` | Fallback when native HTTP fails | ❌ Keep – Resilient fallback. |

---

## Handlers That Do Not Call Bridge (Fully Native)

These handlers have **no** `bridge.ExecutePythonTool` call; they are already fully native in Go:

- `analyze_alignment`, `generate_config`, `health`, `setup_hooks`, `check_attribution`, `add_external_tool_hints`
- `memory_maint` (all actions native, including dream per current implementation)
- `session`, `estimation`, `task_analysis`, `git_tools`, `infer_session_mode`, `tool_catalog`, `workflow_mode`, `prompt_tracking`
- `report` (overview, scorecard, briefing, prd – native only; unsupported action returns error)
- `recommend` (model, workflow, advisor – native only via devwisdom-go; unsupported action returns error)
- `security` (scan, alerts, report – Go projects only; errors otherwise)
- `testing` (run, coverage, validate – Go projects only; unsupported action returns error)
- `lint` (golangci-lint, go-vet, gofmt, goimports, markdownlint, shellcheck, auto – unsupported linter returns error)

---

## Bridge Routes Still in `bridge/execute_tool.py`

The bridge still routes these tool names (for fallback or primary use):

- `task_workflow`, `ollama`, `mlx`, `context`

**Removed 2026-01-28:** `memory`, `task_discovery`, `report`, `recommend`, `security`, `testing`, `lint` – branches and imports removed from execute_tool.py (Go handlers are fully native).

---

## Next Steps (Recommended Order)

1. ~~**Done:** Remove memory’s bridge fallback.~~
2. ~~**Done:** Remove task_discovery bridge fallback and slim bridge (memory + task_discovery branches removed from execute_tool.py).~~
3. **Optional later:** Remove or narrow context bridge usage after verifying behavior.
4. **Optional later:** Remove more dead branches from `execute_tool.py` as tools become fully native; see `SAFE_PYTHON_REMOVAL_PLAN.md`.
5. **Regression:** Keep `toolsWithNoBridge` and regression tests in sync with handler behavior (see `internal/tools/regression_test.go`).

---

## References

- **Status:** `docs/MIGRATION_STATUS_CURRENT.md`
- **Safe removal:** `docs/SAFE_PYTHON_REMOVAL_PLAN.md`, `docs/PYTHON_FALLBACKS_SAFE_TO_REMOVE.md`
- **Regression:** `internal/tools/regression_test.go` (`toolsWithNoBridge`, `TestRegressionFeatureParity`)
