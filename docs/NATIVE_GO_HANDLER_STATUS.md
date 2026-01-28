# Native Go Handler Status

**Date:** 2026-01-09  
**Last Updated:** 2026-01-28 (native migration: 4 removed fallbacks, memory/task_discovery full native)  
**Status:** ✅ Handlers use native implementations; bridge only where required

---

## Executive Summary

**Key Finding:** Handlers in `internal/tools/handlers.go` use native Go first. Many tools are **fully native** (no Python bridge). Others are **hybrid** (native first, bridge fallback for specific actions or when native fails).

**Python fallbacks removed (2026-01-27 / 2026-01-28):** `setup_hooks`, `check_attribution`, `session`, `memory_maint` (4 tools); `memory`, `task_discovery` (2 tools); `report` (1 tool). Bridge no longer routes these; handlers never call `ExecutePythonTool` for them.

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

### Hybrid Tools (Native First, Bridge Fallback)

| Tool | Native | Bridge used when | Handler |
|------|--------|------------------|---------|
| `security` | Go scan, `gh` | Non-Go projects; Go scan/gh fails | `handleSecurity` / `security.go` |
| `task_workflow` | sync, approve, clarity, cleanup, create; clarify (Apple FM) | Apple FM–related errors; **external sync** (agentic-tools) | `handleTaskWorkflow`, `task_workflow_common` |
| `testing` | Go test/coverage/validate | Non-Go; native fails | `handleTesting` / `testing.go` |
| `lint` | Go linters | Non-Go linters (e.g. ruff) | `handleLint` |
| `context` | summarize, budget (with Apple FM) | Unhandled actions; native fails | `handleContext` |
| `recommend` | model, workflow | advisor; native fails | `handleRecommend` |
| `ollama` | HTTP API (docs, quality, summary, etc.) | Fallback when native HTTP fails | `ollama_provider` |

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
