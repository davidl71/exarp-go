# Task-Related Code Migration to mcp-go-core

**Date:** 2026-01-13  
**Status:** Analysis Complete  
**Purpose:** Identify task-related code that can be migrated to `mcp-go-core` shared library

---

## Executive Summary

After analyzing task-related code, **3 high-priority patterns** are excellent candidates for migration to `mcp-go-core`:

1. ✅ **Protobuf/JSON Request Parsing** - Generic pattern (20+ instances)
2. ✅ **Result Formatting Helper** - Generic JSON response formatting
3. ✅ **Default Value Application** - Generic parameter defaults

**NOT candidates:** Task-specific business logic (filtering, storage, domain models)

---

## Current mcp-go-core Packages

**Already Available:**
- `pkg/mcp/framework` - Framework abstraction (MCPServer interface)
- `pkg/mcp/security` - Path validation, project root detection
- `pkg/mcp/types` - Common types (TextContent, ToolSchema, ToolInfo)
- `pkg/mcp/logging` - Structured logging
- `pkg/mcp/cli` - CLI utilities (TTY detection)
- `pkg/mcp/factory` - Server factory pattern
- `pkg/mcp/protocol` - JSON-RPC types

**Missing (Opportunities):**
- ❌ Generic request parsing (protobuf/JSON fallback)
- ❌ Result formatting utilities
- ❌ Default value application helpers

---

## Migration Candidates

### 1. ⭐ **Protobuf/JSON Request Parsing Pattern** - HIGH PRIORITY

**Current Problem:**
The exact same parsing pattern is repeated **20+ times** across `protobuf_helpers.go`:

```go
// Repeated 20+ times with identical structure
func ParseXRequest(args json.RawMessage) (*proto.XRequest, map[string]interface{}, error) {
    var req proto.XRequest
    
    // Try protobuf binary first
    if err := protobuf.Unmarshal(args, &req); err == nil {
        return &req, nil, nil
    }
    
    // Fall back to JSON
    var params map[string]interface{}
    if err := json.Unmarshal(args, &params); err != nil {
        return nil, nil, fmt.Errorf("failed to parse arguments: %w", err)
    }
    
    return nil, params, nil
}
```

**Files Affected:**
- `internal/tools/protobuf_helpers.go` - 20+ functions (ParseMemoryRequest, ParseTaskWorkflowRequest, ParseContextRequest, etc.)

**Proposed mcp-go-core Package:**
```go
// pkg/mcp/request/parser.go
package request

import (
    "encoding/json"
    "google.golang.org/protobuf/proto"
)

// ParseRequest is a generic function for parsing protobuf or JSON requests
// T must be a protobuf message type
func ParseRequest[T proto.Message](
    args json.RawMessage,
    newMessage func() T,
) (T, map[string]interface{}, error) {
    var zero T
    
    // Try protobuf first
    req := newMessage()
    if err := proto.Unmarshal(args, req); err == nil {
        return req, nil, nil
    }
    
    // Fall back to JSON
    var params map[string]interface{}
    if err := json.Unmarshal(args, &params); err != nil {
        return zero, nil, fmt.Errorf("failed to parse arguments: %w", err)
    }
    
    return zero, params, nil
}
```

**Usage After Migration:**
```go
// Before (20+ lines per function):
func ParseTaskWorkflowRequest(args json.RawMessage) (*proto.TaskWorkflowRequest, map[string]interface{}, error) {
    // ... 20 lines of boilerplate ...
}

// After (1 line):
func ParseTaskWorkflowRequest(args json.RawMessage) (*proto.TaskWorkflowRequest, map[string]interface{}, error) {
    return request.ParseRequest(args, func() *proto.TaskWorkflowRequest { return &proto.TaskWorkflowRequest{} })
}
```

**Benefits:**
- ✅ Eliminates 400+ lines of duplicate code
- ✅ Single implementation to maintain
- ✅ Type-safe with generics
- ✅ Reusable across all MCP projects

**Migration Impact:**
- **Lines Removed:** ~400 lines from `protobuf_helpers.go`
- **Complexity:** Low (generic function)
- **Risk:** Low (well-tested pattern)

---

### 2. ⭐ **Result Formatting Helper** - HIGH PRIORITY

**Current Problem:**
Every handler repeats this pattern:

```go
// Repeated in every handler
result := map[string]interface{}{...}
output, _ := json.MarshalIndent(result, "", "  ")

outputPath, _ := params["output_path"].(string)
if outputPath != "" {
    if err := os.WriteFile(outputPath, output, 0644); err == nil {
        result["output_path"] = outputPath
    }
}

return []framework.TextContent{
    {Type: "text", Text: string(output)},
}, nil
```

**Files Affected:**
- `internal/tools/task_workflow_common.go` - 6 handlers
- `internal/tools/handlers.go` - 20+ handlers
- All tool handlers

**Proposed mcp-go-core Package:**
```go
// pkg/mcp/response/formatter.go
package response

import (
    "encoding/json"
    "os"
    "github.com/davidl71/mcp-go-core/pkg/mcp/types"
)

// FormatResult formats a result map as JSON and optionally writes to file
func FormatResult(
    result map[string]interface{},
    outputPath string,
) ([]types.TextContent, error) {
    output, err := json.MarshalIndent(result, "", "  ")
    if err != nil {
        return nil, fmt.Errorf("failed to marshal result: %w", err)
    }
    
    if outputPath != "" {
        if err := os.WriteFile(outputPath, output, 0644); err == nil {
            result["output_path"] = outputPath
            // Re-marshal with output_path included
            output, _ = json.MarshalIndent(result, "", "  ")
        }
    }
    
    return []types.TextContent{
        {Type: "text", Text: string(output)},
    }, nil
}
```

**Usage After Migration:**
```go
// Before (10+ lines):
result := map[string]interface{}{...}
output, _ := json.MarshalIndent(result, "", "  ")
outputPath, _ := params["output_path"].(string)
if outputPath != "" {
    // ... file writing logic ...
}
return []framework.TextContent{{Type: "text", Text: string(output)}}, nil

// After (1 line):
return response.FormatResult(result, outputPath)
```

**Benefits:**
- ✅ Eliminates 200+ lines of duplicate code
- ✅ Consistent formatting across all tools
- ✅ Centralized file writing logic
- ✅ Reusable across all MCP projects

**Migration Impact:**
- **Lines Removed:** ~200 lines across handlers
- **Complexity:** Low (simple helper)
- **Risk:** Low (straightforward extraction)

---

### 3. ⭐ **Default Value Application** - MEDIUM PRIORITY

**Current Problem:**
Default value setting is verbose and repeated:

```go
// Repeated in multiple handlers
if req.Action == "" {
    params["action"] = "sync"
}
if req.SubAction == "" {
    params["sub_action"] = "list"
}
if req.OutputFormat == "" {
    params["output_format"] = "text"
}
if req.Status == "" {
    params["status"] = "Review"
}
```

**Files Affected:**
- `internal/tools/handlers.go` - handleTaskWorkflow, handleTesting, etc.
- Multiple tool handlers

**Proposed mcp-go-core Package:**
```go
// pkg/mcp/request/defaults.go
package request

// ApplyDefaults applies default values to a params map
// Only sets values if key doesn't exist or is empty string
func ApplyDefaults(params map[string]interface{}, defaults map[string]interface{}) {
    for key, defaultValue := range defaults {
        if val, exists := params[key]; !exists || val == "" {
            params[key] = defaultValue
        }
    }
}
```

**Usage After Migration:**
```go
// Before (4+ lines per default):
if req.Action == "" {
    params["action"] = "sync"
}
// ... repeat for each default ...

// After (1 line):
request.ApplyDefaults(params, map[string]interface{}{
    "action":        "sync",
    "sub_action":    "list",
    "output_format": "text",
    "status":        "Review",
})
```

**Benefits:**
- ✅ Eliminates 50+ lines of repetitive default setting
- ✅ More maintainable (defaults in one place)
- ✅ Reusable across all tools
- ✅ Less error-prone

**Migration Impact:**
- **Lines Removed:** ~50 lines
- **Complexity:** Low (simple helper)
- **Risk:** Low (straightforward)

---

## NOT Migration Candidates (Task-Specific)

### ❌ Task Filtering Logic
**Why:** Business logic specific to task management
**Location:** `task_workflow_common.go` - filterTasks, filtering by status/tag/priority
**Reason:** Domain-specific, not generic MCP infrastructure

### ❌ Database/File Fallback Pattern
**Why:** Specific to exarp-go's task storage architecture
**Location:** `task_workflow_common.go` - database-first, file-fallback pattern
**Reason:** Storage strategy is project-specific, not generic

### ❌ Task Metadata Serialization
**Why:** Specific to Todo2Task domain model
**Location:** `database/tasks.go` - protobuf/JSON metadata handling
**Reason:** Domain-specific data structures

### ❌ Task ID Generation
**Why:** Specific to epoch-based task IDs
**Location:** `task_workflow_common.go` - generateEpochTaskID
**Reason:** Domain-specific ID format

### ❌ Task Status Normalization
**Why:** Specific to Todo2 workflow states
**Location:** `task_workflow_common.go` - normalizeStatus
**Reason:** Domain-specific business rules

---

## Migration Plan

### Phase 1: Add Generic Utilities to mcp-go-core

**1.1 Add Request Parsing Package**
- Create `pkg/mcp/request/parser.go`
- Implement generic `ParseRequest` function
- Add tests

**1.2 Add Response Formatting Package**
- Create `pkg/mcp/response/formatter.go`
- Implement `FormatResult` function
- Add tests

**1.3 Add Defaults Package**
- Create `pkg/mcp/request/defaults.go`
- Implement `ApplyDefaults` function
- Add tests

### Phase 2: Migrate exarp-go to Use New Packages

**2.1 Update protobuf_helpers.go**
- Replace 20+ ParseXRequest functions with generic calls
- Remove duplicate code

**2.2 Update All Handlers**
- Replace result formatting with `response.FormatResult`
- Replace default setting with `request.ApplyDefaults`

**2.3 Update Dependencies**
- Update `go.mod` to use new mcp-go-core version
- Test all tool handlers

### Phase 3: Share with Other Projects

**3.1 devwisdom-go Integration**
- Use same request parsing utilities
- Use same result formatting
- Use same default helpers

**3.2 Future Projects**
- Any new MCP server can use these utilities
- Standardized patterns across ecosystem

---

## Estimated Impact

### Code Reduction
- **Protobuf Parsing:** ~400 lines removed
- **Result Formatting:** ~200 lines removed
- **Default Application:** ~50 lines removed
- **Total:** ~650 lines of duplicate code eliminated

### Benefits
- ✅ **Maintainability:** Single implementation to fix/improve
- ✅ **Consistency:** Same patterns across all projects
- ✅ **Reusability:** Any MCP project can use these utilities
- ✅ **Testing:** Centralized test coverage

### Risks
- ⚠️ **Breaking Changes:** Need to coordinate mcp-go-core updates
- ⚠️ **Version Management:** Need to keep mcp-go-core version in sync
- ⚠️ **Testing:** Need comprehensive tests before migration

---

## Implementation Priority

### High Priority (Quick Wins)
1. ✅ **Result Formatting** - Simple extraction, high impact
2. ✅ **Default Application** - Simple helper, immediate benefit

### Medium Priority (Larger Refactor)
3. ✅ **Protobuf Parsing** - Requires generics, but eliminates most duplication

### Low Priority (Future)
4. Consider additional generic patterns as they emerge
5. Evaluate if other projects need similar utilities

---

## Next Steps

1. **Create mcp-go-core Issues:**
   - Add request parsing package
   - Add response formatting package
   - Add defaults package

2. **Implement in mcp-go-core:**
   - Add packages with tests
   - Document usage patterns
   - Version release

3. **Migrate exarp-go:**
   - Update to use new packages
   - Remove duplicate code
   - Test all handlers

4. **Share with devwisdom-go:**
   - Update devwisdom-go to use same utilities
   - Standardize patterns

---

## Conclusion

**3 high-priority patterns** are excellent candidates for migration to `mcp-go-core`:

1. ✅ **Protobuf/JSON Request Parsing** - Eliminates 400+ lines
2. ✅ **Result Formatting** - Eliminates 200+ lines
3. ✅ **Default Value Application** - Eliminates 50+ lines

**Total Impact:** ~650 lines of duplicate code eliminated, improved maintainability, and standardized patterns across MCP projects.

**Task-specific code** (filtering, storage, domain models) should remain in exarp-go as it's business logic, not generic infrastructure.
