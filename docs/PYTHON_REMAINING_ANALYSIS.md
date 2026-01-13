# Python Code Remaining Analysis

**Date:** 2026-01-12  
**Status:** Current state after native Go migration progress

## Executive Summary

**Total Python Files:** ~20 files (excluding `project_management_automation/` and test files)

**Python Code Categories:**
1. **Bridge Scripts** (3 core files + supporting modules) - ✅ Active
2. **Test Files** (10+ files) - ✅ Active
3. **Supporting Modules** (bridge/context, bridge/recommend) - ✅ Active
4. **`project_management_automation/`** (100+ files) - ✅ Active (Python tool implementations)

---

## Active Bridge Scripts

### Core Bridge Scripts (3 files)

1. **`bridge/execute_tool.py`** (361 lines)
   - Routes 19 tools to Python implementations
   - Imports from `project_management_automation.tools.consolidated`
   - Handles context and recommend tools from bridge directory

2. **`bridge/execute_resource.py`** (109 lines)
   - Routes 6 resources to Python implementations
   - Handles memories, prompts, scorecard resources

3. **Supporting Modules:**
   - `bridge/context/` - Context tool implementation
   - `bridge/recommend/` - Recommend tool implementation
   - `bridge/agent_evaluation.py` - Agent evaluation utilities
   - `bridge/statistics_bridge.py` - Statistics bridge

---

## Tools Still Using Python Bridge

**Total: 19 tools still routed through Python bridge**

### Fully Python Bridge (No Native Go)
1. `check_attribution` - Attribution compliance check
2. `memory` - Memory search/save (fully Python)
3. `security` - Security scanning (partial: scan is native, alerts/report use Python)
4. `testing` - Testing tools (fully Python)
5. `lint` - Linting (hybrid: Go linters native, non-Go linters Python)
6. `estimation` - Task duration estimation (fully Python)
7. `session` - Session management (fully Python)
8. `ollama` - Ollama integration (fully Python)
9. `mlx` - MLX integration (fully Python)
10. `context` - Context management (fully Python)
11. `recommend` - Model/workflow recommendation (partial: advisor action native via devwisdom-go)

### Hybrid (Native + Python Fallback)
12. `analyze_alignment` - Native for "todo2", Python for "prd"
13. `setup_hooks` - Native for "git"/"patterns", Python fallback
14. `memory_maint` - Native for "health"/"gc"/"prune"/"consolidate"/"dream", Python fallback
15. `report` - Native for "scorecard" (Go projects), "overview" (partial), "briefing" (native via devwisdom-go), Python for "overview" (automation)
16. `task_analysis` - Native for most actions, Python fallback
17. `task_discovery` - Native for basic scanning, Python fallback
18. `task_workflow` - Native for most actions, Python fallback
19. `add_external_tool_hints` - Fully Python (used by generate_config)

---

## Python Bridge Usage in Go Code

### Files Using Python Bridge

1. **`internal/tools/handlers.go`**
   - 17 tools with `ExecutePythonTool` calls
   - Mix of hybrid and fully Python tools

2. **`internal/tools/automation_native.go`**
   - `runDailyTaskPython` for:
     - `memory_maint/consolidate` (in nightly automation)
     - `report/overview` (in sprint automation)

3. **`internal/tools/testing.go`**
   - All testing actions use Python bridge

4. **`internal/tools/security.go`**
   - `alerts` and `report` actions use Python bridge
   - `scan` action is native Go

5. **`internal/tools/report_mlx.go`**
   - MLX enhancement uses Python bridge

---

## Resources Still Using Python Bridge

**Total: 6 resources routed through Python bridge**

1. `stdio://scorecard` - Project scorecard
2. `stdio://memories` - All memories
3. `stdio://memories/category/{category}` - Memories by category
4. `stdio://memories/task/{task_id}` - Memories by task
5. `stdio://memories/recent` - Recent memories
6. `stdio://memories/session/{date}` - Session memories
7. `stdio://prompts` - All prompts (compact)
8. `stdio://prompts/mode/{mode}` - Prompts by mode
9. `stdio://prompts/persona/{persona}` - Prompts by persona
10. `stdio://prompts/category/{category}` - Prompts by category
11. `stdio://session/mode` - Session mode

**Note:** All resources are fully Python bridge (no native Go implementations yet)

---

## Test Files

**Python Test Files (10+ files):**
- `tests/unit/python/test_execute_tool.py` - Bridge tool routing tests
- `tests/unit/python/test_execute_resource.py` - Bridge resource tests
- `tests/unit/python/test_get_prompt.py` - Prompt retrieval tests
- `tests/integration/bridge/test_bridge_integration.py` - Integration tests
- `tests/integration/mcp/test_server_startup.py` - Server startup tests
- `tests/integration/mcp/test_mcp_server.py` - MCP protocol tests
- `tests/integration/mcp/test_apple_foundation_models.py` - Apple FM tests
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
- `resources/prompt_discovery.py` - Prompt resources
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

### Resources (All Python)
- ❌ All 11 resources are fully Python bridge

### Prompts (All Python)
- ❌ All 15 prompts use Python bridge

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
