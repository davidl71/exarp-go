# Native Go Handler Status

**Date:** 2026-01-09  
**Last Updated:** 2026-01-28 (task_workflow full native: external sync returns error)  
**Status:** ✅ Handlers use native implementations; bridge only where required

---

## Executive Summary

**Key Finding:** Handlers in `internal/tools/handlers.go` use native Go first. Many tools are **fully native** (no Python bridge). Others are **hybrid** (native first, bridge fallback for specific actions or when native fails).

**Python fallbacks removed (2026-01-27 / 2026-01-28):** `setup_hooks`, `check_attribution`, `session`, `memory_maint` (4 tools); `memory`, `task_discovery`, `report`, `recommend`, `security`, `testing`, `lint`, `ollama` (8 tools). Bridge no longer routes these; handlers never call `ExecutePythonTool` for them.

---

## Handler Implementation Status

### Full Native Go Tools (No Python Bridge)

These tools have complete native implementations and **never** use the Python bridge. See `toolsWithNoBridge` in `internal/tools/regression_test.go`.

| Tool | Handler | Native Implementation | Status |
|------|---------|------------------------|--------|
| `server_status` | `handleServerStatus` | `handleServerStatusNative` | ✅ Full Native |
| `tool_catalog` | `handleToolCatalog` | `handleToolCatalogNative` | ✅ Full Native |
| `workflow_mode` | `handleWorkflowMode` | `handleWorkflowModeNative` | ✅ Full Native |
| `infer_session_mode` | `handleInferSessionMode` | `handleInferSessionModeNative` | ✅ Full Native |
| `git_tools` | `handleGitTools` | `handleGitToolsNative` | ✅ Full Native |
| `generate_config` | `handleGenerateConfig` | `handleGenerateConfigNative` | ✅ Full Native |
| `add_external_tool_hints` | `handleAddExternalToolHints` | `handleAddExternalToolHintsNative` | ✅ Full Native |
| `setup_hooks` | `handleSetupHooks` | `handleSetupHooksNative` | ✅ Full Native (fallback removed 2026-01-27) |
| `check_attribution` | `handleCheckAttribution` | `handleCheckAttributionNative` | ✅ Full Native (fallback removed 2026-01-27) |
| `session` | `handleSession` | `handleSessionNative` | ✅ Full Native (fallback removed 2026-01-27) |
| `memory_maint` | `handleMemoryMaint` | `handleMemoryMaintNative` | ✅ Full Native (fallback removed 2026-01-27) |
| `memory` | `handleMemory` | `handleMemoryNative` | ✅ Full Native (fallback removed 2026-01-28) |
| `analyze_alignment` | `handleAnalyzeAlignment` | `handleAnalyzeAlignmentNative` | ✅ Full Native |
| `estimation` | `handleEstimation` | `handleEstimationNative` | ✅ Full Native |
| `task_analysis` | `handleTaskAnalysis` | `handleTaskAnalysisNative` | ✅ Full Native |
| `task_discovery` | `handleTaskDiscovery` | `handleTaskDiscoveryNative` | ✅ Full Native (fallback removed 2026-01-28) |
| `prompt_tracking` | `handlePromptTracking` | `handlePromptTrackingNative` | ✅ Full Native |
| `health` | `handleHealth` | `handleHealthNative` | ✅ Full Native |
| `report` | `handleReport` | overview, scorecard (Go-only), briefing, prd (native); unsupported action returns error | ✅ Full Native (fallback removed) |
| `recommend` | `handleRecommend` | model, workflow, advisor (devwisdom-go in-process) | ✅ Full Native (fallback removed) |
| `security` | `handleSecurity` | scan, alerts, report (Go projects only; errors otherwise) | ✅ Full Native (fallback removed) |
| `testing` | `handleTesting` | run, coverage, validate (Go projects only; unsupported action returns error) | ✅ Full Native (fallback removed) |
| `lint` | `handleLint` | golangci-lint, go-vet, gofmt, goimports, markdownlint, shellcheck, auto; unsupported linter returns error | ✅ Full Native (fallback removed) |
| `ollama` | `handleOllama` | status, models, generate, docs, quality, summary, etc. (native HTTP only; errors on failure) | ✅ Full Native (fallback removed) |
| `context` | `handleContext` | summarize (Apple FM), budget, batch; unknown action returns error | ✅ Full Native (fallback removed 2026-01-28) |
| `task_workflow` | `handleTaskWorkflow` | sync (SQLite↔JSON), approve, clarify (Apple FM), clarity, cleanup, create; external=true returns error | ✅ Full Native (fallback removed 2026-01-28) |

### Hybrid Tools (Native First, Bridge Fallback)

None. All tools with native implementations are now full native; bridge only where required (e.g. `mlx`).

### Python Bridge Only

| Tool | Reason | Handler |
|------|--------|---------|
| `mlx` | No Go bindings; intentional bridge-only | `handleMlx` / `insight_provider` |

---

## Four Removed Python Fallbacks (2026-01-27)

These were previously hybrid (native + fallback). **Fallbacks removed**; handlers now return errors instead of calling the bridge:

1. **`setup_hooks`** — `handleSetupHooksNative` (git + patterns). No bridge call.
2. **`check_attribution`** — `handleCheckAttributionNative`. No bridge call.
3. **`session`** — `handleSessionNative` (prime, handoff, prompts, assignee). No bridge call.
4. **`memory_maint`** — `handleMemoryMaintNative` (health, gc, prune, consolidate, dream). No bridge call.

See `docs/PYTHON_FALLBACKS_SAFE_TO_REMOVE.md` for implementation details.

---

## References

- **Handlers:** `internal/tools/handlers.go`
- **Bridge routing:** `bridge/execute_tool.py`
- **Regression / `toolsWithNoBridge`:** `internal/tools/regression_test.go`
- **Migration plan:** `docs/NATIVE_GO_MIGRATION_PLAN.md`
- **Safe removal:** `docs/PYTHON_FALLBACKS_SAFE_TO_REMOVE.md`, `docs/PYTHON_BRIDGE_MIGRATION_NEXT.md`
