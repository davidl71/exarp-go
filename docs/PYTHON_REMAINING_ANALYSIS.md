# Python Code Remaining Analysis

**Date:** 2026-01-12  
**Last updated:** 2026-01-29 (4 removed Python fallbacks reflected)  
**Status:** Current state after native Go migration progress

**Migration update (2026-01-27):** Python fallbacks were removed for **`setup_hooks`**, **`check_attribution`**, **`session`**, and **`memory_maint`**. These tools are now full native only (no bridge). See `NATIVE_GO_HANDLER_STATUS.md` and `NEXT_MIGRATION_STEPS.md` for current handler status.

## Executive Summary

**Total Python Files:** ~20 files (excluding `project_management_automation/` and test files)

**Python Code Categories:**
1. **Bridge Scripts** (3 core files + supporting modules) - ✅ Active
2. **Test Files** (10+ files) - ✅ Active
3. **Supporting Modules** (bridge/context, bridge/recommend) - ✅ Active
4. **`project_management_automation/`** (100+ files) - ✅ Active (Python tool implementations)

---

## Active Bridge Scripts

### Core Bridge Scripts (2 files; resources are Go-only)

1. **`bridge/execute_tool.py`**
   - Routes selected tools to Python implementations when Go delegates
   - Imports from `project_management_automation.tools.consolidated`
   - Handles context and recommend tools from bridge directory
   - **Resources are not routed here** — exarp-go serves all 22 resources natively from `internal/resources/`

2. **Supporting Modules:**
   - `bridge/context/` - Context tool implementation
   - `bridge/recommend/` - Recommend tool implementation
   - `bridge/agent_evaluation.py` - Agent evaluation utilities
   - `bridge/statistics_bridge.py` - Statistics bridge

---

## Tools Still Using Python Bridge

**Current (2026-01-29):** For authoritative status see **`NATIVE_GO_HANDLER_STATUS.md`**. Summary: **4 Python fallbacks removed (2026-01-27):** `setup_hooks`, `check_attribution`, `session`, `memory_maint` are full native only. Most other tools with native implementations are also full native (no bridge). Only **`mlx`** remains hybrid (native status/hardware; bridge for generate when needed).

### Tools no longer using Python bridge (fallbacks removed or full native)
- `setup_hooks`, `check_attribution`, `session`, `memory_maint` — fallbacks removed 2026-01-27
- `memory`, `task_discovery`, `report`, `recommend`, `security`, `testing`, `lint`, `ollama`, `context`, `task_workflow` — full native (no bridge). See NATIVE_GO_HANDLER_STATUS.md.

### Hybrid (native first, bridge only when needed)
- **`mlx`** — Native status/hardware (darwin+CGO); bridge for models/generate when native unavailable.

### Bridge script routing (historical)
`bridge/execute_tool.py` may still list tools for backward compatibility; Go handlers do not call the bridge for the tools above. Only `mlx` (generate) uses the bridge from Go when needed.

---

## Python Bridge Usage in Go Code

**Current (2026-01-29):** Bridge is used only for **`mlx`** (generate action when native unavailable). Automation uses native handlers only (`runDailyTask` for report, memory_maint, etc.). See `NATIVE_GO_HANDLER_STATUS.md`.

### Files that may call Python bridge

1. **`internal/tools/handlers.go`**
   - **`mlx`** — bridge for generate when native (luxfi/mlx) unavailable. All other tools with native implementations use native only.

2. **`internal/tools/automation_native.go`**
   - Uses `runDailyTask` (native) for report, memory_maint, health, etc. No `runDailyTaskPython` for those tools.

3. **`internal/tools/mlx_invoke.go`** / **`handlers.go`**
   - MLX tool may call bridge for generate.

---

## Resources (All Native Go)

**Total: 22 resources — all served by exarp-go (no Python bridge).**

There is no `execute_resource.py`; the Go MCP server registers and serves all resources in `internal/resources/`:

- **Scorecard:** `stdio://scorecard` — native Go (`scorecard.go`, uses `tools.GenerateGoScorecard`)
- **Memories:** `stdio://memories`, `stdio://memories/category/{category}`, `stdio://memories/task/{task_id}`, `stdio://memories/recent`, `stdio://memories/session/{date}` — native Go (`memories.go`, uses `tools.LoadAllMemories` etc.)
- **Prompts:** `stdio://prompts`, `stdio://prompts/mode/{mode}`, `stdio://prompts/persona/{persona}`, `stdio://prompts/category/{category}` — native Go (`prompts.go`, uses `internal/prompts`)
- **Session:** `stdio://session/mode` — native Go (`session.go`, uses infer_session_mode)
- **Server/models/tools/tasks:** `stdio://server/status`, `stdio://models`, `stdio://cursor/skills`, `stdio://tools`, `stdio://tools/{category}`, `stdio://tasks`, `stdio://tasks/{task_id}`, `stdio://tasks/status/{status}`, `stdio://tasks/priority/{priority}`, `stdio://tasks/tag/{tag}`, `stdio://tasks/summary` — all native Go

**Note:** All resources are fully native Go (no Python bridge for resources).

---

## Test Files

**Python Test Files (10+ files):**
- `tests/unit/python/test_execute_tool.py` - Bridge tool routing tests
- `tests/unit/python/test_execute_resource.py` - Not present (resources are Go-only)
- `tests/unit/python/test_get_prompt.py` - Prompt retrieval tests
- `tests/integration/bridge/test_bridge_integration.py` - Integration tests
- `tests/integration/mcp/test_server_startup.py` - Server startup tests
- `tests/integration/mcp/test_mcp_server.py` - MCP protocol tests
- `tests/fixtures/mock_python.py` - Test fixtures
- `tests/fixtures/test_helpers.py` - Test helpers

---

## `project_management_automation/` Directory

**Status:** ✅ Active dependency (100+ Python files)

This directory contains the actual Python implementations of all tools:
- `tools/consolidated.py` - Consolidated tool implementations
- `tools/external_tool_hints.py` - External tool hints
- `tools/attribution_check.py` - Attribution checking
- `resources/memories.py` - Memory resources
- `resources/prompt_discovery.py` - **Removed**; manifest and discovery inlined into `context_primer.py`
- `utils/` - Utility modules
- And many more...

**Purpose:** This is the Python library that implements the actual tool logic. The bridge scripts call into these implementations.

---

## Migration Status Summary

### Fully Native Go Tools (6 tools)
- ✅ `health` - All actions native
- ✅ `automation` - All actions native
- ✅ `tool_catalog` - Fully native
- ✅ `workflow_mode` - Fully native
- ✅ `git_tools` - Fully native
- ✅ `infer_session_mode` - Fully native

### Hybrid Tools (13 tools)
- ⚠️ `analyze_alignment` - Native "todo2", Python "prd"
- ⚠️ `generate_config` - All native, uses Python for external hints
- ⚠️ `setup_hooks` - Native "git"/"patterns", Python fallback
- ⚠️ `memory_maint` - Native "health"/"gc"/"prune"/"consolidate"/"dream", Python fallback
- ⚠️ `report` - Native "scorecard"/"briefing", Python "overview" (automation)
- ⚠️ `security` - Native "scan", Python "alerts"/"report"
- ⚠️ `task_analysis` - Native most actions, Python fallback
- ⚠️ `task_discovery` - Native basic scanning, Python fallback
- ⚠️ `task_workflow` - Native most actions, Python fallback
- ⚠️ `lint` - Native Go linters, Python non-Go linters
- ⚠️ `recommend` - Native "advisor", Python "model"/"workflow"

### Fully Python Tools (10 tools)
- ❌ `check_attribution` - Fully Python
- ❌ `memory` - Fully Python
- ❌ `testing` - Fully Python
- ❌ `estimation` - Fully Python
- ❌ `session` - Fully Python
- ❌ `ollama` - Fully Python
- ❌ `mlx` - Fully Python
- ❌ `context` - Fully Python
- ❌ `add_external_tool_hints` - Fully Python (utility)

### Resources (All Native Go)
- ✅ All 22 resources are native Go (`internal/resources/`; no Python bridge for resources)

### Prompts (All Native Go)
- ✅ All 34 prompts are native Go (`internal/prompts/`; server registers and serves them)

---

## Next Migration Priorities

### High Priority (Quick Wins)
1. **`memory_maint/consolidate`** - Already partially native, complete migration
2. **`report/overview`** - Already partially native, complete migration
3. **`security/alerts` and `security/report`** - Complete security tool migration
4. **`recommend/model` and `recommend/workflow`** - Already has advisor native

### Medium Priority
5. **`memory` tool** - Core memory operations
6. **`session` tool** - Session management
7. **`context` tool** - Context management
8. **`testing` tool** - Testing utilities
9. **`estimation` tool** - Task estimation
10. **`check_attribution`** - Attribution checking

### Low Priority (ML/AI Integrations)
11. **`ollama` tool** - Ollama integration
12. **`mlx` tool** - MLX integration

### Resources & Prompts
13. **Memory resources** - Migrate to native Go
14. **Prompt resources** - Migrate to native Go
15. **Prompts** - Migrate to native Go

---

## Conclusion

**Current State:**
- ✅ 6 tools fully native (25%)
- ⚠️ 13 tools hybrid (54%)
- ❌ 10 tools fully Python (42%)
- ❌ All resources Python
- ❌ All prompts Python

**Progress:** ~25% of tools fully native, ~79% have some native implementation

**Remaining Work:**
- Complete migration of hybrid tools
- Migrate fully Python tools
- Migrate resources
- Migrate prompts

**Python Bridge Remains Critical:**
- Bridge scripts are essential for gradual migration
- `project_management_automation/` provides implementations
- Test files validate bridge functionality
- Migration strategy: gradual, tool-by-tool approach
