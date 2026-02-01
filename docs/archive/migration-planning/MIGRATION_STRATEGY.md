# Migration Strategy: project-management-automation → exarp-go

**Date:** 2026-01-07  
**Status:** ⚠️ **Superseded** — The canonical migration plan is **`NATIVE_GO_MIGRATION_PLAN.md`** (updated 2026-01-27), which prioritizes Native Go over Python Bridge. This document is retained for historical context.  
**Task:** T-59

---

## Executive Summary

This document defined the migration strategy for completing the migration from `project-management-automation` (Python FastMCP) to `exarp-go` (Go MCP server). **Current status:** See `NATIVE_GO_MIGRATION_PLAN.md`, `MIGRATION_AUDIT_2026-01-12.md`, and `MIGRATION_STATUS_CURRENT.md`. Based on Phase 1 analysis at the time, 6 tools, 18 prompts, and 13 resources remained to migrate; most are now migrated per the Native Go plan.

---

## Migration Patterns

### Pattern 1: Python Bridge (Recommended for Most Tools)

**When to Use:**
- Tools with complex Python dependencies
- Tools with existing Python implementations
- Tools requiring quick migration
- Tools that work well via bridge

**Implementation Steps:**
1. Create Go handler in `internal/tools/handlers.go`
2. Register tool in `internal/tools/registry.go` (Batch 5+)
3. Add tool to `bridge/execute_tool.py`
4. Write integration tests

**Example Pattern:**
```go
// internal/tools/handlers.go
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

```python
# bridge/execute_tool.py
elif tool_name == "tool_name":
    result = _tool_name(
        param1=args.get("param1"),
        param2=args.get("param2"),
        # ... all parameters
    )
```

### Pattern 2: Native Go Implementation

**When to Use:**
- Simple tools with minimal dependencies
- Tools that benefit from Go performance
- Tools with Go-specific functionality (like `lint` for Go linters)

**Implementation Steps:**
1. Implement tool logic directly in Go
2. Create handler in `internal/tools/handlers.go`
3. Register tool in `internal/tools/registry.go`
4. Write unit and integration tests

**Example:** `lint` tool (hybrid - Go for Go linters, Python bridge for others)

### Pattern 3: Hybrid Approach

**When to Use:**
- Tools with multiple components
- Tools where some parts benefit from Go, others from Python

**Example:** `lint` tool
- Go linters: Native Go implementation
- Non-Go linters: Python bridge fallback

---

## Decision Framework

### Tool Migration Decision Tree

```
Is tool simple with minimal dependencies?
├─ YES → Consider Native Go
│   └─ Is performance critical?
│       ├─ YES → Native Go
│       └─ NO → Python Bridge (faster migration)
└─ NO → Python Bridge
    └─ Does tool require FastMCP Context?
        ├─ YES → Document limitations, use stdio mode
        └─ NO → Standard Python Bridge
```

### Remaining Tools Analysis

#### 1. `server_status` - **HIGH PRIORITY**
- **Decision:** Python Bridge
- **Rationale:** Simple tool, quick migration, maintains compatibility
- **Complexity:** Low

#### 2. `context` - **MEDIUM PRIORITY**
- **Decision:** Verify if covered by existing Go tools
- **Rationale:** Go has `context_summarize`, `context_batch`, `context_budget`
- **Action:** Check if Python `context` tool is covered by these
- **Complexity:** Low (may be no-op if covered)

#### 3. `prompt_tracking` - **MEDIUM PRIORITY**
- **Decision:** Python Bridge
- **Rationale:** Useful feature, straightforward migration
- **Complexity:** Low-Medium

#### 4. `demonstrate_elicit` - **LOW PRIORITY**
- **Decision:** Document limitations, optional migration
- **Rationale:** Demo tool, requires FastMCP Context
- **Action:** Document that it requires FastMCP Context, may not work in stdio mode
- **Complexity:** Medium (FastMCP Context dependency)

#### 5. `interactive_task_create` - **LOW PRIORITY**
- **Decision:** Document limitations, optional migration
- **Rationale:** Demo tool, requires FastMCP Context
- **Action:** Document that it requires FastMCP Context, may not work in stdio mode
- **Complexity:** Medium (FastMCP Context dependency)

#### 6. `recommend` - **ALREADY COMPLETE** ✅
- **Decision:** Mark as complete
- **Rationale:** Already split into `recommend_model` and `recommend_workflow` in Go
- **Action:** Document as complete, no migration needed

---

## Prompts Migration Strategy

### Pattern: Python Bridge (All Prompts)

**Implementation Pattern:**
```go
// internal/prompts/registry.go
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

```python
# bridge/get_prompt.py
if prompt_name == "prompt_name":
    return PROMPT_TEMPLATE
```

### Remaining Prompts (18)

**Persona Prompts (8):**
- `arch`, `dev`, `exec`, `pm`, `qa`, `reviewer`, `seceng`, `writer`
- **Strategy:** Migrate as batch (similar structure)
- **Complexity:** Low (template-based)

**Workflow Prompts (10):**
- `auto`, `auto_high`, `automation_setup`, `doc_check`, `doc_quick`, `end_of_day`, `project_health`, `resume_session`, `view_handoffs`, `weekly`
- **Strategy:** Migrate as batch (similar structure)
- **Complexity:** Low-Medium (workflow templates)

---

## Resources Migration Strategy

### Pattern: Python Bridge (All Resources)

**Implementation Pattern:**
```go
// internal/resources/handlers.go
func handleResourceName(ctx context.Context, uri string) ([]byte, string, error) {
    return bridge.ExecutePythonResource(ctx, uri)
}
```

```python
# bridge/execute_resource.py
if uri.startswith("automation://resource_name"):
    return get_resource_data(uri)
```

### URI Scheme Migration

**Current:** `automation://` (Python)  
**Target:** `stdio://` (Go)

**Migration Approach:**
1. Map `automation://` URIs to `stdio://` equivalents
2. Update resource handlers to support both schemes (backward compatibility)
3. Document URI scheme changes

### Remaining Resources (13)

**Automation Resources:**
- `automation://agents` → `stdio://agents`
- `automation://cache` → `stdio://cache`
- `automation://history` → `stdio://history`
- `automation://linters` → `stdio://linters`
- `automation://models` → `stdio://models`
- `automation://problem-categories` → `stdio://problem-categories`
- `automation://tts-backends` → `stdio://tts-backends`
- `automation://status` → `stdio://status`
- `automation://tools` → `stdio://tools`
- `automation://tasks` → `stdio://tasks`
- `automation://tasks/agent/{agent_name}` → `stdio://tasks/agent/{agent_name}`
- `automation://tasks/status/{status}` → `stdio://tasks/status/{status}`
- `automation://memories/health` → `stdio://memories/health`

**Strategy:** Migrate as batch, maintain backward compatibility during transition

---

## FastMCP Context Dependencies

### Tools Requiring FastMCP Context

1. **`demonstrate_elicit`**
   - Uses `elicit()` API for inline chat questions
   - **Limitation:** Requires FastMCP Context (not available in stdio mode)
   - **Solution:** Document limitation, mark as optional

2. **`interactive_task_create`**
   - Uses `elicit()` API for interactive task creation
   - **Limitation:** Requires FastMCP Context (not available in stdio mode)
   - **Solution:** Document limitation, mark as optional

### Handling Strategy

**Option 1: Document and Skip**
- Document that these tools require FastMCP Context
- Mark as optional/low priority
- Users can use FastMCP server for these features

**Option 2: Alternative Implementation**
- Create stdio-compatible alternatives
- Use different interaction patterns
- More complex, but maintains functionality

**Recommendation:** Option 1 (Document and Skip) - These are demo tools, not core functionality

---

## Testing Strategy

### Unit Tests
- Test Go handlers with mock bridge calls
- Test parameter parsing and error handling
- Test edge cases

### Integration Tests
- Test tool execution via MCP interface
- Test Python bridge communication
- Test error propagation

### Test Files Structure
```
tests/
├── unit/
│   └── go/
│       └── tools/
│           └── handlers_test.go
└── integration/
    ├── mcp/
    │   └── test_server_tools.go
    └── bridge/
        └── test_python_bridge.go
```

---

## Migration Phases

### Phase 3: High-Priority Tools (Batch 1)
**Tools:** `server_status`, `context` (verify), `prompt_tracking`  
**Estimated Duration:** 1 week  
**Approach:** Python bridge for all

### Phase 4: Remaining Tools (Batch 2)
**Tools:** `demonstrate_elicit`, `interactive_task_create` (optional)  
**Estimated Duration:** 1 week  
**Approach:** Document limitations, optional migration

### Phase 5: Prompts and Resources
**Prompts:** 18 prompts (persona + workflow)  
**Resources:** 13 resources (automation:// → stdio://)  
**Estimated Duration:** 2 weeks  
**Approach:** Batch migration, template-based

### Phase 6: Testing and Validation
**Scope:** Full integration testing, documentation, cleanup  
**Estimated Duration:** 1 week

---

## Success Criteria

### Tools
- ✅ All 6 tools migrated (or documented as complete/optional)
- ✅ All tools tested and working
- ✅ Integration tests passing

### Prompts
- ✅ All 18 prompts migrated
- ✅ All prompts tested
- ✅ Prompt templates working

### Resources
- ✅ All 13 resources migrated
- ✅ URI scheme migration complete
- ✅ Backward compatibility maintained (if needed)

---

## Risk Mitigation

### High Risk
- **FastMCP Context Dependencies:** Document limitations, mark as optional
- **URI Scheme Changes:** Maintain backward compatibility during transition

### Medium Risk
- **Resource URI Migration:** Test thoroughly, document changes
- **Prompt Template Compatibility:** Verify template rendering

### Low Risk
- **Simple Tool Migration:** Follow established patterns
- **Template-Based Prompts:** Straightforward migration

---

## Next Steps

1. ✅ **Phase 2 Complete** - Strategy documented
2. 📋 **Begin Phase 3** - Start with high-priority tools
3. 📋 **Verify `context` tool** - Check if covered by existing Go tools
4. 📋 **Create tool migration templates** - Standardize approach

---

**Last Updated:** 2026-01-07  
**Status:** Strategy Design Complete  
**Next Phase:** Phase 3 - High-Priority Tools Migration

