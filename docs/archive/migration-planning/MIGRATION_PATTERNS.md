# Migration Patterns Reference Guide

**Date:** 2026-01-07  
**Purpose:** Standard patterns for migrating tools, prompts, and resources

---

## Tool Migration Pattern

### Step 1: Create Go Handler

**File:** `internal/tools/handlers.go`

```go
// handleToolName handles the tool_name tool
func handleToolName(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
    // Parse arguments
    var params map[string]interface{}
    if err := json.Unmarshal(args, &params); err != nil {
        return nil, fmt.Errorf("failed to parse arguments: %w", err)
    }

    // Execute via Python bridge
    result, err := bridge.ExecutePythonTool(ctx, "tool_name", params)
    if err != nil {
        return nil, fmt.Errorf("tool_name failed: %w", err)
    }

    return []framework.TextContent{
        {Type: "text", Text: result},
    }, nil
}
```

### Step 2: Register Tool

**File:** `internal/tools/registry.go`

```go
// In registerBatch5Tools() or appropriate batch function
if err := server.RegisterTool(
    "tool_name",
    "[HINT: Tool description. action=action1|action2. Unified tool description.]",
    framework.ToolSchema{
        Type: "object",
        Properties: map[string]interface{}{
            "action": map[string]interface{}{
                "type":    "string",
                "enum":    []string{"action1", "action2"},
                "default": "action1",
            },
            "param1": map[string]interface{}{
                "type": "string",
            },
            "param2": map[string]interface{}{
                "type":    "boolean",
                "default": false,
            },
        },
    },
    handleToolName,
); err != nil {
    return fmt.Errorf("failed to register tool_name: %w", err)
}
```

### Step 3: Add to Python Bridge

**File:** `bridge/execute_tool.py`

```python
# Add import at top
from project_management_automation.tools.consolidated import (
    tool_name as _tool_name,
)

# Add handler in main execution block
elif tool_name == "tool_name":
    result = _tool_name(
        action=args.get("action", "action1"),
        param1=args.get("param1"),
        param2=args.get("param2", False),
        # ... all parameters
    )
```

### Step 4: Write Tests

**File:** `tests/integration/mcp/test_server_tools.go`

```go
func TestToolName(t *testing.T) {
    // Test tool registration
    // Test tool execution
    // Test error handling
}
```

---

## Prompt Migration Pattern

### Step 1: Add to Go Registry

**File:** `internal/prompts/registry.go`

```go
// In RegisterAllPrompts()
{"prompt_name", "Prompt description."},
```

### Step 2: Add to Python Bridge

**File:** `bridge/get_prompt.py`

```python
# Add import
from project_management_automation.prompts import (
    PROMPT_TEMPLATE,
)

# Add handler
if prompt_name == "prompt_name":
    return PROMPT_TEMPLATE
```

---

## Resource Migration Pattern

### Step 1: Create Go Handler

**File:** `internal/resources/handlers.go`

```go
// handleResourceName handles the stdio://resource_name resource
func handleResourceName(ctx context.Context, uri string) ([]byte, string, error) {
    return bridge.ExecutePythonResource(ctx, uri)
}
```

### Step 2: Register Resource

**File:** `internal/resources/handlers.go`

```go
// In RegisterAllResources()
if err := server.RegisterResource(
    "stdio://resource_name",
    "Resource Name",
    "Resource description.",
    "application/json",
    handleResourceName,
); err != nil {
    return fmt.Errorf("failed to register resource_name resource: %w", err)
}
```

### Step 3: Add to Python Bridge

**File:** `bridge/execute_resource.py`

```python
# Add handler
if uri.startswith("automation://resource_name") or uri.startswith("stdio://resource_name"):
    return get_resource_data(uri), "application/json"
```

---

## URI Scheme Migration

### Pattern: automation:// → stdio://

**Old URI:** `automation://resource_name`  
**New URI:** `stdio://resource_name`

**Implementation:**
- Support both schemes during transition
- Update documentation
- Eventually deprecate `automation://` scheme

---

## FastMCP Context Tools

### Pattern: Document Limitations

**Tools:** `demonstrate_elicit`, `interactive_task_create`

**Approach:**
1. Document that tool requires FastMCP Context
2. Mark as optional/low priority
3. Note that stdio mode doesn't support Context
4. Users can use FastMCP server for these features

---

## Testing Patterns

### Unit Test Pattern

```go
func TestHandleToolName(t *testing.T) {
    tests := []struct {
        name    string
        args    json.RawMessage
        wantErr bool
    }{
        {
            name:    "valid args",
            args:    json.RawMessage(`{"action": "action1"}`),
            wantErr: false,
        },
        {
            name:    "invalid args",
            args:    json.RawMessage(`{}`),
            wantErr: false, // Tool should handle defaults
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := handleToolName(context.Background(), tt.args)
            if (err != nil) != tt.wantErr {
                t.Errorf("handleToolName() error = %v, wantErr %v", err, tt.wantErr)
            }
            if !tt.wantErr && len(result) == 0 {
                t.Error("handleToolName() returned empty result")
            }
        })
    }
}
```

### Integration Test Pattern

```go
func TestIntegrationToolName(t *testing.T) {
    // Test via MCP server
    // Verify tool registration
    // Test tool execution
    // Validate response format
}
```

---

## Common Pitfalls

### 1. Parameter Type Mismatches
- **Issue:** Python expects different types than Go provides
- **Solution:** Convert types in bridge script (string → int, etc.)

### 2. JSON String vs Dict
- **Issue:** Tools may return dict or JSON string
- **Solution:** Bridge script handles both (see `execute_tool.py`)

### 3. Error Propagation
- **Issue:** Python errors not properly propagated
- **Solution:** Wrap errors in bridge script, return proper error format

### 4. Missing Parameters
- **Issue:** Optional parameters not handled
- **Solution:** Use `.get()` with defaults in bridge script

---

## Checklist for Each Tool

- [ ] Go handler created in `handlers.go`
- [ ] Tool registered in `registry.go` (appropriate batch)
- [ ] Tool added to `bridge/execute_tool.py`
- [ ] All parameters mapped correctly
- [ ] Default values handled
- [ ] Error handling implemented
- [ ] Integration test written
- [ ] Documentation updated
- [ ] Tool tested via MCP interface

---

**Last Updated:** 2026-01-07  
**Status:** Reference Guide Complete

