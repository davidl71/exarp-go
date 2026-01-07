# Shared Tool Migration Research - Complete Reference

**Date:** 2026-01-07  
**Status:** ✅ **Complete Research Reference**  
**Purpose:** Comprehensive research document for all 24 tool migration tasks (T-22 through T-45)

---

## Overview

This document consolidates all research findings from batch research execution for individual tool migration tasks. Each tool task should reference this document and include tool-specific implementation details.

**Research Sources:**
- Batch Research: `docs/BATCH2_PARALLEL_RESEARCH_REMAINING_TASKS.md`
- Python Implementation: `mcp_stdio_tools/server.py`
- Framework Design: `docs/TEST_PARALLEL_RESEARCH_T2.md`
- Foundation Research: `docs/BATCH1_PARALLEL_RESEARCH_TNAN_T8.md`

---

## Section 1: Local Codebase Analysis

### Python Tool Implementation Location

**File:** `mcp_stdio_tools/server.py`

**Tool Registration Pattern:**
```python
@server.list_tools()
async def list_tools() -> list[Tool]:
    tool_definitions = [
        {
            "name": "tool_name",
            "description": "[HINT: ...]",
            "inputSchema": {
                "type": "object",
                "properties": {
                    # Tool-specific parameters
                }
            }
        }
    ]
```

**Tool Execution Pattern:**
```python
@server.call_tool()
async def call_tool(name: str, arguments: dict[str, Any]) -> list[TextContent]:
    if name == "tool_name":
        result = _tool_function(
            param1=arguments.get("param1", default),
            # ... other parameters
        )
    return [TextContent(type="text", text=result)]
```

**Current Tool List:**
- **Batch 1 (6 tools):** `analyze_alignment`, `generate_config`, `health`, `setup_hooks`, `check_attribution`, `add_external_tool_hints`
- **Batch 2 (8 tools):** `memory`, `memory_maint`, `report`, `security`, `task_analysis`, `task_discovery`, `task_workflow`, `testing`
- **Batch 3 (8 tools):** `automation`, `tool_catalog`, `workflow_mode`, `lint`, `estimation`, `git_tools`, `session`, `infer_session_mode`

**Python Bridge Location:**
- Bridge script: `bridge/execute_tool.py`
- Go bridge implementation: `internal/bridge/python.go`

---

## Section 2: Internet Research (2026)

### Go MCP Tool Registration

🔗 **[Go SDK Tool Registration](https://github.com/modelcontextprotocol/go-sdk)**

**Key Patterns:**
- Use `RegisterTool` method with handler function
- Handler signature: `func(ctx context.Context, args json.RawMessage) ([]mcp.TextContent, error)`
- Tool schema defined via JSON Schema
- Return `TextContent` array for results

**Example:**
```go
server.RegisterTool("tool_name", func(ctx context.Context, args json.RawMessage) ([]mcp.TextContent, error) {
    // Parse arguments
    var params ToolParams
    if err := json.Unmarshal(args, &params); err != nil {
        return nil, fmt.Errorf("invalid arguments: %w", err)
    }
    
    // Execute tool via Python bridge
    result, err := bridge.ExecutePythonTool(ctx, "tool_name", params)
    if err != nil {
        return nil, fmt.Errorf("tool execution failed: %w", err)
    }
    
    return []mcp.TextContent{
        {Type: "text", Text: result},
    }, nil
})
```

### Go Python Bridge Patterns

🔗 **[Go os/exec Package](https://golang.org/pkg/os/exec/)**

**Key Patterns:**
- Use `exec.CommandContext` for subprocess execution
- JSON marshaling/unmarshaling for data exchange
- Context for timeout and cancellation
- Error handling for Python execution failures

**Best Practices:**
- Set timeouts for Python calls (30s default, configurable)
- Handle stderr for error messages
- Validate JSON responses
- Use structured logging
- Support cancellation via context

**Implementation:**
```go
func ExecutePythonTool(ctx context.Context, toolName string, args map[string]interface{}) (string, error) {
    argsJSON, _ := json.Marshal(args)
    cmd := exec.CommandContext(ctx, "python3", "bridge/execute_tool.py", toolName, string(argsJSON))
    
    output, err := cmd.Output()
    if err != nil {
        return "", fmt.Errorf("python execution failed: %w", err)
    }
    
    return string(output), nil
}
```

### MCP Protocol Specification

🔗 **[MCP Specification](https://modelcontextprotocol.io/specification)**

**Tool Definition Requirements:**
- Name: Unique identifier
- Description: Human-readable description
- InputSchema: JSON Schema defining parameters
- Annotations: Optional metadata

**Tool Calling Protocol:**
- Client sends `tools/call` request with name and arguments
- Server executes tool and returns content array
- Content can be text, image, or resource
- `isError` flag indicates success/failure

**Response Format:**
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Tool output"
      }
    ],
    "isError": false
  }
}
```

---

## Section 3: Synthesis from Batch Research

### Tractatus Analysis

**Session ID:** 167a4f8a-4e12-4b74-9dd6-38105ab498f7  
**Concept:** "What is tool migration strategy for migrating Python MCP tools to Go?"

**Key Propositions:**

1. **Migration Approaches**
   - **Python Bridge:** Execute Python tools via subprocess (recommended initially)
   - **Direct Port:** Rewrite tools in Go (long-term goal)
   - **Hybrid:** Bridge for complex tools, port simple ones (balanced approach)

2. **Python Bridge Strategy**
   - Use `exec.Command` for Python execution
   - JSON-RPC 2.0 for communication
   - Error propagation from Python to Go
   - Timeout handling for long operations

3. **Tool Registration**
   - Register tools with framework-agnostic interface
   - Tools execute via Python bridge initially
   - Bridge handles argument marshaling/unmarshaling
   - Results converted to MCP TextContent format

### Context7 Documentation

**Library:** Model Context Protocol Specification  
**Library ID:** `/websites/modelcontextprotocol_io_specification`

**Key Findings:**

**Tool Definition:**
- Name: Unique identifier
- Description: Human-readable description
- InputSchema: JSON Schema defining parameters
- Annotations: Optional metadata

**Tool Calling:**
- Client sends `tools/call` request with name and arguments
- Server executes tool and returns content array
- Content can be text, image, or resource
- `isError` flag indicates success/failure

**Recommendations:**
- Use JSON Schema for input validation
- Return TextContent array for tool results
- Handle errors gracefully with isError flag
- Support multiple content types (text, image, resource)

### Web Search Findings

**Query:** "Go Python bridge subprocess execution JSON-RPC 2026"

**Key Findings:**

**Go Python Bridge Patterns:**
- Use `os/exec` package for subprocess execution
- JSON for data exchange (marshaling/unmarshaling)
- Context for timeout and cancellation
- Error handling for Python execution failures

**Best Practices:**
- Set timeouts for Python calls
- Handle stderr for error messages
- Validate JSON responses
- Use structured logging

**Recommendations:**
- Implement bridge with timeout support
- Add retry logic for transient failures
- Cache Python environment detection
- Monitor bridge performance

---

## Section 4: Implementation Strategy

### Common Migration Pattern

**For All Tools:**

1. **Tool Registration**
   - Define tool schema in Go (match Python schema)
   - Register with framework-agnostic interface
   - Use `RegisterTool` method

2. **Handler Implementation**
   - Create handler function
   - Parse arguments from JSON
   - Call Python bridge with tool name and arguments
   - Convert Python result to MCP TextContent format

3. **Error Handling**
   - Validate input arguments
   - Handle Python execution errors
   - Return meaningful error messages
   - Support timeout and cancellation

4. **Testing**
   - Unit tests for handler function
   - Integration tests with Python bridge
   - Cursor integration verification

### Framework-Agnostic Design

**Key Principles:**
- All tools register via `MCPServer` interface
- Tools work with any MCP framework (go-sdk, mcp-go, go-mcp)
- Configuration-based framework selection
- Easy framework switching

**Implementation:**
```go
type MCPServer interface {
    RegisterTool(name, description string, schema ToolSchema, handler ToolHandler) error
    // ... other methods
}

type ToolHandler func(ctx context.Context, args json.RawMessage) ([]TextContent, error)
```

### Python Bridge Integration

**Bridge Script:** `bridge/execute_tool.py`
- Accepts tool name and JSON arguments
- Imports and executes Python tool
- Returns JSON result

**Go Bridge:** `internal/bridge/python.go`
- Executes Python script via subprocess
- Handles JSON marshaling/unmarshaling
- Manages timeouts and cancellation
- Propagates errors

---

## Tool-Specific Implementation Notes

### Batch 1 Tools (Simple)

**Common Characteristics:**
- Simple parameter structures
- Straightforward Python function calls
- No complex state management
- Quick execution times

**Migration Approach:**
- Direct Python bridge execution
- Minimal Go wrapper code
- Standard error handling

### Batch 2 Tools (Medium)

**Common Characteristics:**
- More complex parameters
- May involve file I/O
- Some state management
- Variable execution times

**Migration Approach:**
- Python bridge with enhanced error handling
- File path handling
- State management via Python
- Timeout configuration

### Batch 3 Tools (Advanced)

**Common Characteristics:**
- Complex parameter structures
- May involve external APIs
- Resource management
- Longer execution times

**Migration Approach:**
- Python bridge with extended timeouts
- Resource cleanup handling
- API error propagation
- Performance monitoring

---

## Testing Requirements

### Unit Tests
- Test handler function with mock bridge
- Test argument parsing
- Test error handling
- Test timeout behavior

### Integration Tests
- Test with actual Python bridge
- Test tool execution end-to-end
- Test error propagation
- Test cancellation

### Cursor Integration
- Verify tool appears in tool list
- Test tool execution in Cursor
- Verify result format
- Test error display

---

## Performance Considerations

### Python Bridge Overhead
- Subprocess creation: ~50-100ms
- JSON marshaling: ~1-5ms
- Python execution: varies by tool
- Total overhead: ~100-200ms per call

### Optimization Strategies
- Cache Python environment detection
- Reuse subprocess connections (future)
- Batch tool calls when possible
- Monitor and log performance

---

## Error Handling Patterns

### Input Validation
```go
if param == "" {
    return nil, fmt.Errorf("param is required")
}
```

### Python Execution Errors
```go
result, err := bridge.ExecutePythonTool(ctx, toolName, args)
if err != nil {
    return nil, fmt.Errorf("tool execution failed: %w", err)
}
```

### Timeout Handling
```go
ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
defer cancel()
```

---

## References

- **Batch Research:** `docs/BATCH2_PARALLEL_RESEARCH_REMAINING_TASKS.md`
- **Framework Design:** `docs/TEST_PARALLEL_RESEARCH_T2.md`
- **Foundation Research:** `docs/BATCH1_PARALLEL_RESEARCH_TNAN_T8.md`
- **Python Implementation:** `mcp_stdio_tools/server.py`
- **Go SDK:** https://github.com/modelcontextprotocol/go-sdk
- **MCP Specification:** https://modelcontextprotocol.io/specification

---

**Status:** ✅ Complete research reference for all 24 tool migration tasks.

