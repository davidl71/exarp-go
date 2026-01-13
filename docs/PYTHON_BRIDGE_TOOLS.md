# Python Bridge Tools - Remaining Tools

**Date:** 2026-01-13  
**Status:** 8 tools/actions still use Python bridge (24% of total tools)

---

## Summary

Out of 34 total tools, **26 tools (76%)** are now fully or mostly native Go. The remaining **8 tools/actions** use Python bridge for legitimate reasons:

1. **MCP Client Dependencies** (3 actions) - Require calling other MCP servers
2. **Specialized Analysis Features** (3 actions) - Benefit from Python analysis libraries
3. **External Framework Dependencies** (4 actions) - MLX requires Python MLX library

---

## Tools Using Python Bridge

### 1. MCP Client Dependencies (3 actions)

These tools call other MCP servers (devwisdom-go) and appropriately use Python bridge:

#### `recommend` tool - **advisor** action ⚠️
- **Status:** 2/3 actions native (model ✅, workflow ✅), advisor ⚠️ Python bridge
- **Reason:** Requires calling devwisdom-go MCP server for advisor consultation
- **Location:** `internal/tools/handlers.go:handleRecommend`
- **Python Bridge:** `bridge/execute_tool.py` → `recommend.tool.recommend`
- **Native Alternative:** Would require implementing MCP client in Go (future enhancement)

#### `report` tool - **briefing** action ⚠️
- **Status:** 3/4 actions native (overview ✅, scorecard ✅, prd ✅), briefing ⚠️ Python bridge
- **Reason:** Requires calling devwisdom-go MCP server for wisdom briefings
- **Location:** `internal/tools/handlers.go:handleReport`
- **Python Bridge:** `bridge/execute_tool.py` → `consolidated_reporting.report`
- **Native Alternative:** Would require implementing MCP client in Go (future enhancement)

#### `memory_maint` tool - **dream** action ⚠️
- **Status:** 4/5 actions native (health ✅, gc ✅, prune ✅, consolidate ✅), dream ⚠️ Python bridge
- **Reason:** Requires calling devwisdom-go MCP server for advisor-based dreaming
- **Location:** `internal/tools/handlers.go:handleMemoryMaint`
- **Python Bridge:** `bridge/execute_tool.py` → `consolidated_memory.memory_maint`
- **Native Alternative:** Would require implementing MCP client in Go (future enhancement)

---

### 2. Specialized Analysis Features (3 actions)

These tools provide specialized documentation/analysis features that benefit from Python libraries:

#### `ollama` tool - **docs, quality, summary** actions ⚠️
- **Status:** 5/8 actions native (status ✅, models ✅, generate ✅, pull ✅, hardware ✅), docs ⚠️, quality ⚠️, summary ⚠️ Python bridge
- **Reason:** Specialized analysis features (documentation generation, code quality analysis, text summarization) that benefit from Python analysis libraries
- **Location:** `internal/tools/handlers.go:handleOllama`
- **Python Bridge:** `bridge/execute_tool.py` → `ollama_integration.ollama`
- **Native Alternative:** Could be implemented in Go if needed, but Python provides richer analysis libraries
- **Note:** Core Ollama functionality (status, models, generate, pull, hardware) is fully native Go

---

### 3. External Framework Dependencies (4 actions)

These tools require external Python libraries that are not available in Go:

#### `mlx` tool - **all actions** ⚠️
- **Status:** 0/4 actions native (status ⚠️, hardware ⚠️, models ⚠️, generate ⚠️) - All Python bridge
- **Reason:** Requires Python MLX library (Apple Silicon ML framework) which is Python-native
- **Location:** `internal/tools/handlers.go:handleMlx`
- **Python Bridge:** `bridge/execute_tool.py` → `mlx_integration.mlx`
- **Native Alternative:** MLX is a Python framework with no Go equivalent. Python bridge is appropriate.
- **Note:** MLX is used for Apple Silicon-optimized ML inference, which requires the Python MLX library

---

## Python Bridge Tool Mapping

### Bridge File: `bridge/execute_tool.py`

The following tools are routed to Python bridge:

```python
# MCP Client Dependencies
"recommend" → recommend.tool.recommend (advisor action only)
"report" → consolidated_reporting.report (briefing action only)
"memory_maint" → consolidated_memory.memory_maint (dream action only)

# Specialized Analysis
"ollama" → ollama_integration.ollama (docs, quality, summary actions only)

# External Framework
"mlx" → mlx_integration.mlx (all actions)
```

### Python Tool Files

1. **`project_management_automation/tools/recommend/tool.py`**
   - Used for: `recommend` advisor action

2. **`project_management_automation/tools/consolidated_reporting.py`**
   - Used for: `report` briefing action

3. **`project_management_automation/tools/consolidated_memory.py`**
   - Used for: `memory_maint` dream action

4. **`project_management_automation/tools/ollama_integration.py`**
   - Used for: `ollama` docs, quality, summary actions

5. **`project_management_automation/tools/mlx_integration.py`**
   - Used for: `mlx` all actions

---

## Migration Recommendations

### ✅ Keep as Python Bridge (Appropriate)

These should remain Python bridge due to legitimate technical reasons:

1. **MCP Client Dependencies** (3 actions)
   - Would require implementing MCP client in Go
   - Current Python implementation works well
   - Low priority for migration

2. **MLX Integration** (4 actions)
   - MLX is Python-native framework
   - No Go equivalent available
   - Appropriate to keep as Python bridge

### 🤔 Consider Migration (Optional)

These could be migrated to Go if needed, but provide less value:

1. **Ollama Specialized Actions** (3 actions)
   - docs, quality, summary actions
   - Could be implemented in Go using HTTP API calls
   - Python provides richer analysis libraries
   - Low priority - core Ollama functionality is native

---

## Statistics

**By Category:**
- **MCP Client Dependencies:** 3 actions (38%)
- **Specialized Analysis:** 3 actions (38%)
- **External Framework:** 4 actions (50%)

**By Tool:**
- **Fully Python Bridge:** 1 tool (mlx - 4 actions)
- **Partial Python Bridge:** 4 tools (recommend, report, memory_maint, ollama - 1-3 actions each)

**Overall:**
- **Total Tools:** 34
- **Fully Native:** 23 (68%)
- **Mostly Native:** 3 (9%)
- **Python Bridge:** 8 actions (23%)
- **Native Coverage:** 26/34 (76%)

---

## Conclusion

The remaining Python bridge usage is **intentional and appropriate**:

1. **MCP Client Dependencies** - Require calling other MCP servers (would need Go MCP client)
2. **Specialized Analysis** - Benefit from Python analysis libraries (optional to migrate)
3. **External Framework** - MLX requires Python MLX library (must stay Python)

All core functionality tools are now native Go. The Python bridge usage is for specialized features and external dependencies that are appropriately handled in Python.

**Status: Production Ready** ✅

The remaining Python bridge usage does not impact the core functionality of the project management system, which is fully native Go.
