# Stream 5: Test Patterns & Templates

**Date:** 2026-01-12  
**Status:** ✅ **DOCUMENTATION COMPLETE**  
**Objective:** Comprehensive test patterns and templates for Stream 5 implementation

---

## Executive Summary

This document provides test patterns, templates, and best practices for testing native Go implementations in exarp-go. Based on existing test patterns and Go testing best practices.

---

## Test Infrastructure

### Mock Server

**Location:** `tests/fixtures/mock_server.go`

**Usage:**
```go
import "github.com/davidl71/exarp-go/tests/fixtures"

func TestMyTool(t *testing.T) {
    server := fixtures.NewMockServer("test-server")
    
    // Register tools
    err := RegisterAllTools(server)
    if err != nil {
        t.Fatalf("RegisterAllTools() error = %v", err)
    }
    
    // Test tool registration
    tool, exists := server.GetTool("tool_name")
    if !exists {
        t.Error("tool not registered")
    }
    
    // Test tool execution
    ctx := context.Background()
    args := fixtures.MustToJSONRawMessage(map[string]interface{}{
        "action": "test",
    })
    
    result, err := server.CallTool(ctx, "tool_name", args)
    if err != nil {
        t.Errorf("CallTool() error = %v", err)
    }
    
    // Verify result
    if len(result) == 0 {
        t.Error("CallTool() returned empty result")
    }
}
```

### Test Helpers

**Location:** `tests/fixtures/test_helpers.go`

**Key Functions:**
- `TestContext(timeout)` - Create test context with timeout
- `MustToJSONRawMessage(v)` - Convert value to json.RawMessage (panics on error)
- `ToJSONRawMessage(v)` - Convert value to json.RawMessage (returns error)
- `JSONRPCRequest(method, params)` - Build JSON-RPC request
- `JSONRPCResponse(id, result)` - Build JSON-RPC response
- `JSONRPCError(id, code, message)` - Build JSON-RPC error

**Usage:**
```go
import "github.com/davidl71/exarp-go/tests/fixtures"

func TestMyTool(t *testing.T) {
    ctx, cancel := fixtures.TestContext(5 * time.Second)
    defer cancel()
    
    args := fixtures.MustToJSONRawMessage(map[string]interface{}{
        "action": "test",
    })
    
    result, err := handleMyTool(ctx, args)
    // ... assertions
}
```

---

## Test Patterns

### Pattern 1: Simple Tool Test (No Dependencies)

**Use Case:** Tools with no external dependencies, simple I/O

**Template:**
```go
func TestHandleToolName(t *testing.T) {
    ctx := context.Background()
    args := fixtures.MustToJSONRawMessage(map[string]interface{}{})
    
    result, err := handleToolName(ctx, args)
    if err != nil {
        t.Fatalf("handleToolName() error = %v", err)
    }
    
    if len(result) == 0 {
        t.Error("handleToolName() returned empty result")
    }
    
    // Verify result structure (if applicable)
    // Verify result content (if applicable)
}
```

**Examples:**
- `server_status.go`
- `tool_catalog.go`
- `workflow_mode.go`

---

### Pattern 2: Tool with Actions

**Use Case:** Tools with multiple actions (action parameter)

**Template:**
```go
func TestHandleToolName(t *testing.T) {
    tests := []struct {
        name    string
        action  string
        params  map[string]interface{}
        wantErr bool
    }{
        {
            name:    "valid action 1",
            action:  "action1",
            params:  map[string]interface{}{"param1": "value1"},
            wantErr: false,
        },
        {
            name:    "valid action 2",
            action:  "action2",
            params:  map[string]interface{}{"param2": 42},
            wantErr: false,
        },
        {
            name:    "invalid action",
            action:  "invalid",
            params:  map[string]interface{}{},
            wantErr: true,
        },
        {
            name:    "default action",
            action:  "",
            params:  map[string]interface{}{},
            wantErr: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := context.Background()
            
            params := map[string]interface{}{
                "action": tt.action,
            }
            for k, v := range tt.params {
                params[k] = v
            }
            
            args := fixtures.MustToJSONRawMessage(params)
            
            result, err := handleToolName(ctx, args)
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

**Examples:**
- `config_generator.go` (actions: rules, ignore, simplify)
- `health_check.go` (actions: server, git, docs, dod, cicd)
- `git_tools.go` (actions: commits, branches, tasks, diff, graph, merge, set_branch)
- `hooks_setup.go` (actions: git, patterns)

---

### Pattern 3: Tool with File I/O

**Use Case:** Tools that read/write files

**Template:**
```go
func TestHandleToolName(t *testing.T) {
    // Setup: Create temporary directory
    tempDir := t.TempDir()
    projectRoot := filepath.Join(tempDir, "test-project")
    os.MkdirAll(projectRoot, 0755)
    
    // Create test files
    testFile := filepath.Join(projectRoot, "test.txt")
    os.WriteFile(testFile, []byte("test content"), 0644)
    
    // Change to test directory (if needed)
    oldDir, _ := os.Getwd()
    os.Chdir(projectRoot)
    defer os.Chdir(oldDir)
    
    ctx := context.Background()
    args := fixtures.MustToJSONRawMessage(map[string]interface{}{
        "action": "test",
    })
    
    result, err := handleToolName(ctx, args)
    if err != nil {
        t.Fatalf("handleToolName() error = %v", err)
    }
    
    // Verify files created/modified
    // Verify result
}
```

**Examples:**
- `config_generator.go`
- `external_tool_hints.go`
- `hooks_setup.go`
- `workflow_mode.go`

---

### Pattern 4: Tool with Database Dependencies

**Use Case:** Tools that use Todo2 database

**Template:**
```go
func TestHandleToolName(t *testing.T) {
    // Setup: Create test database
    tempDir := t.TempDir()
    projectRoot := filepath.Join(tempDir, "test-project")
    os.MkdirAll(projectRoot, 0755)
    
    // Initialize test database
    todo2Dir := filepath.Join(projectRoot, ".todo2")
    os.MkdirAll(todo2Dir, 0755)
    
    // Initialize database (may need helper function)
    dbPath := filepath.Join(todo2Dir, "todo2.db")
    // ... database initialization ...
    
    // Create test data
    // ... create test tasks ...
    
    ctx := context.Background()
    args := fixtures.MustToJSONRawMessage(map[string]interface{}{
        "action": "test",
    })
    
    result, err := handleToolName(ctx, args)
    if err != nil {
        t.Fatalf("handleToolName() error = %v", err)
    }
    
    // Verify database changes
    // Verify result
}
```

**Examples:**
- `alignment_analysis.go`
- `memory.go`
- `task_workflow.go`
- `session.go`
- `estimation_native.go`
- `prompt_tracking.go`

**Note:** May need test database helper functions to simplify setup.

---

### Pattern 5: Tool with External Commands

**Use Case:** Tools that execute external commands (git, go, etc.)

**Template:**
```go
func TestHandleToolName(t *testing.T) {
    // Option 1: Use real command (if available and fast)
    // Option 2: Mock command (if complex or slow)
    // Option 3: Skip test if command not available
    
    ctx := context.Background()
    args := fixtures.MustToJSONRawMessage(map[string]interface{}{
        "action": "test",
    })
    
    result, err := handleToolName(ctx, args)
    
    // Handle case where command not available
    if err != nil && strings.Contains(err.Error(), "command not found") {
        t.Skip("Required command not available")
    }
    
    if err != nil {
        t.Fatalf("handleToolName() error = %v", err)
    }
    
    // Verify result
}
```

**Examples:**
- `git_tools.go` (git commands)
- `health_check.go` (go commands, git commands)
- `linting.go` (linter commands)
- `testing.go` (go test commands)

---

### Pattern 6: Tool with External API

**Use Case:** Tools that call external APIs

**Template:**
```go
func TestHandleToolName(t *testing.T) {
    // Option 1: Mock API calls (recommended for unit tests)
    // Option 2: Skip test if API not available
    // Option 3: Use test API endpoint
    
    ctx := context.Background()
    args := fixtures.MustToJSONRawMessage(map[string]interface{}{
        "action": "test",
    })
    
    result, err := handleToolName(ctx, args)
    
    // Handle case where API not available
    if err != nil && strings.Contains(err.Error(), "connection refused") {
        t.Skip("API not available")
    }
    
    if err != nil {
        t.Fatalf("handleToolName() error = %v", err)
    }
    
    // Verify result
}
```

**Examples:**
- `ollama_native.go` (Ollama API)
- `security.go` (Security API endpoints)
- `report.go` (External APIs for some features)

---

### Pattern 7: Tool with Hybrid Pattern (Native + Python Bridge)

**Use Case:** Tools that try native first, fall back to Python bridge

**Template:**
```go
func TestHandleToolName_Native(t *testing.T) {
    // Test native implementation
    ctx := context.Background()
    args := fixtures.MustToJSONRawMessage(map[string]interface{}{
        "action": "native_action",
    })
    
    result, err := handleToolNameNative(ctx, args)
    if err != nil {
        t.Fatalf("handleToolNameNative() error = %v", err)
    }
    
    // Verify native result
}

func TestHandleToolName_Hybrid(t *testing.T) {
    // Test hybrid handler (native + fallback)
    // This tests the handler in handlers.go
    ctx := context.Background()
    args := fixtures.MustToJSONRawMessage(map[string]interface{}{
        "action": "native_action",
    })
    
    result, err := handleToolName(ctx, args)
    if err != nil {
        t.Fatalf("handleToolName() error = %v", err)
    }
    
    // Verify result (should use native implementation)
}

func TestHandleToolName_Fallback(t *testing.T) {
    // Test fallback to Python bridge
    // May need to mock Python bridge or skip test
    // This is complex - may be better suited for integration tests
}
```

**Examples:**
- `session.go` (native: prime, handoff; bridge: prompts, assignee)
- `health_check.go` (native: server; bridge: git, docs, dod, cicd)
- `automation_native.go` (native: daily, discover; bridge: nightly, sprint)

**Note:** Testing Python bridge fallback may require integration tests or mocking.

---

### Pattern 8: Benchmark Test

**Use Case:** Performance testing and optimization

**Template:**
```go
func BenchmarkToolName(b *testing.B) {
    sizes := []struct {
        name string
        size int
    }{
        {"Small_100", 100},
        {"Medium_1000", 1000},
        {"Large_10000", 10000},
    }
    
    for _, size := range sizes {
        b.Run(size.name, func(b *testing.B) {
            // Setup test data
            data := generateTestData(size.size)
            
            b.ResetTimer()
            b.ReportAllocs()
            
            for i := 0; i < b.N; i++ {
                _ = ToolFunction(data)
            }
        })
    }
}
```

**Examples:**
- `statistics_bench_test.go`
- `task_analysis_bench_test.go`
- `graph_helpers_bench_test.go`

---

## Best Practices

### 1. Table-Driven Tests

**Always use table-driven tests for multiple test cases:**

```go
func TestHandleToolName(t *testing.T) {
    tests := []struct {
        name    string
        args    map[string]interface{}
        wantErr bool
    }{
        {"valid case 1", map[string]interface{}{"action": "test1"}, false},
        {"valid case 2", map[string]interface{}{"action": "test2"}, false},
        {"invalid case", map[string]interface{}{"action": "invalid"}, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ... test implementation
        })
    }
}
```

### 2. Test Helpers

**Use fixtures package for common operations:**

```go
import "github.com/davidl71/exarp-go/tests/fixtures"

// Use test context
ctx, cancel := fixtures.TestContext(5 * time.Second)
defer cancel()

// Use JSON conversion
args := fixtures.MustToJSONRawMessage(params)

// Use mock server
server := fixtures.NewMockServer("test-server")
```

### 3. Error Handling

**Test both success and error cases:**

```go
tests := []struct {
    name    string
    args    map[string]interface{}
    wantErr bool
    errMsg  string // Expected error message substring
}{
    {"success", map[string]interface{}{"action": "test"}, false, ""},
    {"error", map[string]interface{}{"action": "invalid"}, true, "invalid action"},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        result, err := handleTool(ctx, args)
        if (err != nil) != tt.wantErr {
            t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
        }
        if tt.wantErr && tt.errMsg != "" {
            if !strings.Contains(err.Error(), tt.errMsg) {
                t.Errorf("error message = %q, want to contain %q", err.Error(), tt.errMsg)
            }
        }
    })
}
```

### 4. Edge Cases

**Test edge cases and boundary conditions:**

```go
tests := []struct {
    name    string
    args    map[string]interface{}
    wantErr bool
}{
    {"empty args", map[string]interface{}{}, false},
    {"nil args", nil, false}, // If supported
    {"missing required param", map[string]interface{}{"other": "value"}, true},
    {"invalid type", map[string]interface{}{"action": 123}, true},
    {"empty string", map[string]interface{}{"action": ""}, false}, // Default behavior
    {"very long string", map[string]interface{}{"action": strings.Repeat("a", 10000)}, false},
}
```

### 5. Cleanup

**Always clean up test resources:**

```go
func TestHandleToolName(t *testing.T) {
    tempDir := t.TempDir() // Automatically cleaned up
    defer func() {
        // Additional cleanup if needed
    }()
    
    // ... test implementation
}
```

### 6. Parallel Testing

**Use `t.Parallel()` for independent tests:**

```go
func TestHandleToolName(t *testing.T) {
    t.Parallel() // Allows test to run in parallel
    
    // ... test implementation
}
```

**Note:** Don't use `t.Parallel()` if tests share resources or modify global state.

### 7. Subtests

**Use subtests for better test organization:**

```go
func TestHandleToolName(t *testing.T) {
    tests := []struct {
        name string
        // ... test cases
    }{
        {"test case 1", ...},
        {"test case 2", ...},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ... test implementation
        })
    }
}
```

---

## Test Organization

### File Naming

- Unit tests: `*_test.go` (e.g., `tool_name_test.go`)
- Benchmark tests: `*_bench_test.go` (e.g., `tool_name_bench_test.go`)
- Integration tests: `tests/integration/` directory

### Test Structure

```go
package tools

import (
    "context"
    "testing"
    
    "github.com/davidl71/exarp-go/tests/fixtures"
)

// TestHandleToolName tests the tool handler
func TestHandleToolName(t *testing.T) {
    // ... implementation
}

// TestHandleToolName_EdgeCases tests edge cases
func TestHandleToolName_EdgeCases(t *testing.T) {
    // ... implementation
}

// BenchmarkToolName benchmarks the tool
func BenchmarkToolName(b *testing.B) {
    // ... implementation
}
```

---

## Common Patterns

### Pattern: Testing Handler via Mock Server

```go
func TestToolViaMockServer(t *testing.T) {
    server := fixtures.NewMockServer("test-server")
    err := RegisterAllTools(server)
    if err != nil {
        t.Fatalf("RegisterAllTools() error = %v", err)
    }
    
    ctx := context.Background()
    args := fixtures.MustToJSONRawMessage(map[string]interface{}{
        "action": "test",
    })
    
    result, err := server.CallTool(ctx, "tool_name", args)
    if err != nil {
        t.Errorf("CallTool() error = %v", err)
    }
    
    // Verify result
}
```

### Pattern: Testing Direct Handler

```go
func TestToolHandler(t *testing.T) {
    ctx := context.Background()
    args := fixtures.MustToJSONRawMessage(map[string]interface{}{
        "action": "test",
    })
    
    result, err := handleToolName(ctx, args)
    if err != nil {
        t.Fatalf("handleToolName() error = %v", err)
    }
    
    // Verify result
}
```

### Pattern: Testing with Project Root

```go
func TestToolWithProjectRoot(t *testing.T) {
    tempDir := t.TempDir()
    projectRoot := filepath.Join(tempDir, "test-project")
    os.MkdirAll(projectRoot, 0755)
    
    // Set PROJECT_ROOT environment variable or use FindProjectRoot
    oldRoot := os.Getenv("PROJECT_ROOT")
    os.Setenv("PROJECT_ROOT", projectRoot)
    defer os.Setenv("PROJECT_ROOT", oldRoot)
    
    ctx := context.Background()
    args := fixtures.MustToJSONRawMessage(map[string]interface{}{})
    
    result, err := handleToolName(ctx, args)
    // ... assertions
}
```

---

## References

- **Mock Server:** `tests/fixtures/mock_server.go`
- **Test Helpers:** `tests/fixtures/test_helpers.go`
- **Existing Tests:** `internal/tools/*_test.go`
- **Go Testing:** https://go.dev/doc/tutorial/add-a-test
- **Testing Best Practices:** `docs/MIGRATION_PATTERNS.md#testing-patterns`
- **Stream 5 Plan:** `docs/STREAM_5_TESTING_VALIDATION.md`
- **Coverage Analysis:** `docs/STREAM_5_TEST_COVERAGE_GAP_ANALYSIS.md`

---

**Last Updated:** 2026-01-12  
**Status:** ✅ Documentation Complete - Ready for Test Implementation
