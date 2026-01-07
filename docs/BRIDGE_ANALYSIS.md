# Python Bridge vs Native Go Implementation Analysis

**Date:** 2026-01-07  
**Status:** ✅ Complete Analysis

## Executive Summary

This document provides a comprehensive comparison of which tools, resources, and prompts use the Python bridge versus native Go implementation, and identifies FastMCP dependencies.

**Key Findings:**
- **23 tools** use Python bridge (1 has hybrid: Go for Go linters, Python for others)
- **6 resources** all use Python bridge
- **15 prompts** all use Python bridge
- **FastMCP is NOT used** in the bridge scripts - they import directly from `project_management_automation`
- **Dict issues** are handled in bridge scripts via JSON conversion

---

## Tools Analysis

### Total Tools: 24

#### Native Go Implementation (1 tool - partial)
1. **`lint`** - Hybrid implementation
   - **Go linters** (golangci-lint, go-vet, gofmt, goimports): Native Go implementation
   - **Non-Go linters** (ruff, etc.): Python bridge fallback
   - **Location:** `internal/tools/linting.go`
   - **Bridge Usage:** Conditional (only for non-Go linters)

#### Python Bridge Implementation (23 tools)

**Batch 1 Tools (6):**
1. `analyze_alignment` - Python bridge
2. `generate_config` - Python bridge
3. `health` - Python bridge
4. `setup_hooks` - Python bridge
5. `check_attribution` - Python bridge
6. `add_external_tool_hints` - Python bridge

**Batch 2 Tools (8):**
7. `memory` - Python bridge
8. `memory_maint` - Python bridge
9. `report` - Python bridge
10. `security` - Python bridge
11. `task_analysis` - Python bridge
12. `task_discovery` - Python bridge
13. `task_workflow` - Python bridge
14. `testing` - Python bridge

**Batch 3 Tools (9):**
15. `automation` - Python bridge
16. `tool_catalog` - Python bridge
17. `workflow_mode` - Python bridge
18. `estimation` - Python bridge
19. `git_tools` - Python bridge
20. `session` - Python bridge
21. `infer_session_mode` - Python bridge
22. `ollama` - Python bridge
23. `mlx` - Python bridge

### Tool Implementation Pattern

**Go Handler Pattern:**
```go
func handleToolName(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
    var params map[string]interface{}
    if err := json.Unmarshal(args, &params); err != nil {
        return nil, fmt.Errorf("failed to parse arguments: %w", err)
    }
    
    result, err := bridge.ExecutePythonTool(ctx, "tool_name", params)
    if err != nil {
        return nil, fmt.Errorf("tool_name failed: %w", err)
    }
    
    return []framework.TextContent{
        {Type: "text", Text: result},
    }, nil
}
```

**Bridge Script Pattern:**
```python
# bridge/execute_tool.py imports from:
from project_management_automation.tools.consolidated import (
    analyze_alignment as _analyze_alignment,
    generate_config as _generate_config,
    # ... all tools
)
```

---

## Resources Analysis

### Total Resources: 6

**All resources use Python bridge:**

1. `stdio://scorecard` - Python bridge
2. `stdio://memories` - Python bridge
3. `stdio://memories/category/{category}` - Python bridge
4. `stdio://memories/task/{task_id}` - Python bridge
5. `stdio://memories/recent` - Python bridge
6. `stdio://memories/session/{date}` - Python bridge

### Resource Implementation Pattern

**Go Handler Pattern:**
```go
func handleResourceName(ctx context.Context, uri string) ([]byte, string, error) {
    return bridge.ExecutePythonResource(ctx, uri)
}
```

**Bridge Script Pattern:**
```python
# bridge/execute_resource.py imports from:
from project_management_automation.resources.memories import (
    get_memories_resource,
    # ... all resource handlers
)
from project_management_automation.tools.project_scorecard import (
    generate_project_scorecard,
)
```

---

## Prompts Analysis

### Total Prompts: 15

**All prompts use Python bridge:**

**Original Prompts (8):**
1. `align` - Python bridge
2. `discover` - Python bridge
3. `config` - Python bridge
4. `scan` - Python bridge
5. `scorecard` - Python bridge
6. `overview` - Python bridge
7. `dashboard` - Python bridge
8. `remember` - Python bridge

**High-Value Workflow Prompts (7):**
9. `daily_checkin` - Python bridge
10. `sprint_start` - Python bridge
11. `sprint_end` - Python bridge
12. `pre_sprint` - Python bridge
13. `post_impl` - Python bridge
14. `sync` - Python bridge
15. `dups` - Python bridge

### Prompt Implementation Pattern

**Go Handler Pattern:**
```go
func createPromptHandler(promptName string) framework.PromptHandler {
    return func(ctx context.Context, args map[string]interface{}) (string, error) {
        promptText, err := bridge.GetPythonPrompt(ctx, promptName)
        if err != nil {
            return "", fmt.Errorf("failed to get prompt %s: %w", promptName, err)
        }
        return promptText, nil
    }
}
```

**Bridge Script Pattern:**
```python
# bridge/get_prompt.py imports from:
from project_management_automation.prompts import (
    TASK_ALIGNMENT_ANALYSIS,
    TASK_DISCOVERY,
    # ... all prompt templates
)
```

---

## FastMCP Usage Analysis

### ❌ FastMCP is NOT Used in Bridge Scripts

**Key Finding:** The bridge scripts (`bridge/execute_tool.py`, `bridge/execute_resource.py`, `bridge/get_prompt.py`) do **NOT** use FastMCP.

**Evidence:**
1. **No FastMCP imports** in bridge scripts
2. **Direct imports** from `project_management_automation.tools.consolidated`
3. **No decorators** (`@mcp.tool()`, `@mcp.resource()`, `@mcp.prompt()`)
4. **No FastMCP Context** usage

**Bridge Script Architecture:**
```python
# bridge/execute_tool.py
# NO FastMCP imports - direct function calls
from project_management_automation.tools.consolidated import (
    analyze_alignment as _analyze_alignment,
    # ... direct imports
)

# Direct function calls (not FastMCP tools)
result = _analyze_alignment(
    action=args.get("action", "todo2"),
    # ... parameters
)
```

### FastMCP Usage in Source Project

**Note:** The actual Python tool implementations in `project_management_automation` may or may not use FastMCP, but:
- The bridge scripts don't care - they call functions directly
- The tools were migrated FROM FastMCP TO stdio to avoid FastMCP static analysis issues
- The bridge bypasses FastMCP entirely by calling functions directly

---

## Dict Issues Analysis

### Current Dict Handling

**Bridge Scripts Handle Dict/JSON Conversion:**

```python
# bridge/execute_tool.py lines 371-377
# Handle result - tools may return dict or JSON string
if isinstance(result, dict):
    result_json = json.dumps(result, indent=2)
elif isinstance(result, str):
    result_json = result
else:
    result_json = json.dumps({"result": str(result)}, indent=2)

print(result_json)
```

**This pattern:**
- ✅ Handles dict returns (converts to JSON)
- ✅ Handles string returns (passes through)
- ✅ Handles other types (wraps in dict)
- ✅ Avoids FastMCP dict issues by bypassing FastMCP entirely

### FastMCP Dict Issues (Historical Context)

**From Documentation:**
- Tools were migrated FROM FastMCP TO stdio to avoid FastMCP static analysis issues
- FastMCP has problems with dict types in static analysis
- Stdio mode bypasses FastMCP static analysis entirely

**Current Solution:**
- Bridge scripts don't use FastMCP
- Direct function calls avoid FastMCP dict issues
- JSON conversion handles dict returns safely

---

## Simplification Opportunities

### Can We Bypass FastMCP Further?

**Current State:**
- ✅ Bridge scripts already bypass FastMCP
- ✅ Direct function calls (no FastMCP decorators)
- ✅ JSON conversion handles dict issues

**Potential Simplifications:**

1. **Remove FastMCP from Source Project (if present)**
   - If `project_management_automation` tools use FastMCP decorators
   - Could refactor to plain Python functions
   - Would eliminate FastMCP dependency entirely

2. **Simplify Dict Handling**
   - Current JSON conversion is already simple
   - Could standardize on always returning JSON strings
   - Would eliminate dict/string branching

3. **Native Go Migration**
   - Migrate more tools to native Go
   - Only lint tool has partial Go implementation
   - 23 tools still use Python bridge

### Recommendation: Bypass FastMCP as First Step

**✅ YES - FastMCP is Already Bypassed**

The bridge scripts already bypass FastMCP by:
1. Direct function imports (not FastMCP tools)
2. Direct function calls (not FastMCP decorators)
3. JSON conversion (handles dict issues)

**Next Steps for Further Simplification:**

1. **Verify source project** - Check if `project_management_automation` tools use FastMCP
2. **If FastMCP present** - Refactor to plain Python functions
3. **Standardize returns** - Always return JSON strings (eliminate dict/string branching)
4. **Consider native Go** - Migrate more tools to native Go implementation

---

## Summary Tables

### Tools by Implementation Type

| Type | Count | Tools |
|------|-------|-------|
| Native Go (full) | 0 | - |
| Native Go (partial) | 1 | lint (Go linters only) |
| Python Bridge | 23 | All others |

### Resources by Implementation Type

| Type | Count | Resources |
|------|-------|-----------|
| Native Go | 0 | - |
| Python Bridge | 6 | All resources |

### Prompts by Implementation Type

| Type | Count | Prompts |
|------|-------|---------|
| Native Go | 0 | - |
| Python Bridge | 15 | All prompts |

### FastMCP Usage

| Component | Uses FastMCP | Notes |
|-----------|--------------|-------|
| Bridge Scripts | ❌ NO | Direct function calls |
| Go Handlers | ❌ NO | Call bridge scripts |
| Source Tools | ❓ UNKNOWN | May use FastMCP, but bridge bypasses it |

---

## Conclusion

**Current Architecture:**
- Bridge scripts bypass FastMCP entirely
- Direct function calls avoid FastMCP static analysis
- JSON conversion handles dict issues
- Only 1 tool has partial native Go implementation

**Simplification Status:**
- ✅ FastMCP already bypassed in bridge layer
- ⚠️ Source project may still use FastMCP (needs verification)
- ✅ Dict issues handled via JSON conversion
- 📋 Opportunity: Migrate more tools to native Go

**Recommendation:**
1. Verify if source project uses FastMCP
2. If yes, refactor to plain Python functions
3. Standardize on JSON string returns
4. Consider native Go migration for frequently-used tools

