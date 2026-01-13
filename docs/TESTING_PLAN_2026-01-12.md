# Testing Plan for Native Go Migration

**Date:** 2026-01-12  
**Status:** In Progress  
**Scope:** Comprehensive testing for all native Go implementations

---

## Executive Summary

This document outlines the comprehensive testing strategy for all native Go implementations migrated in Streams 1, 2, and 3. The goal is to achieve 80%+ test coverage and ensure feature parity with Python bridge implementations.

---

## Testing Strategy

### 1. Unit Tests

**Goal:** Test individual tool/resource functions in isolation

**Coverage Target:** 80%+ for all native implementations

**Test Files:**
- `internal/tools/*_test.go` - Tool unit tests
- `internal/resources/*_test.go` - Resource unit tests

**Test Categories:**
- ✅ Argument parsing and validation
- ✅ Error handling
- ✅ Edge cases
- ✅ Return format validation
- ✅ Native vs Python bridge fallback behavior

### 2. Integration Tests

**Goal:** Test tools and resources via MCP interface

**Test Files:**
- `tests/integration/mcp/test_tools.go` - Tool integration tests
- `tests/integration/mcp/test_resources.go` - Resource integration tests

**Test Categories:**
- ✅ Tool registration and discovery
- ✅ Parameter passing via MCP
- ✅ Response format validation
- ✅ Error propagation
- ✅ Resource URI handling

### 3. Performance Benchmarks

**Goal:** Measure and document performance improvements

**Test Files:**
- `internal/tools/*_bench_test.go` - Tool benchmarks
- `internal/resources/*_bench_test.go` - Resource benchmarks

**Metrics:**
- Execution time (native vs Python bridge)
- Memory usage
- Startup time
- Throughput

### 4. Regression Testing

**Goal:** Verify feature parity and no regressions

**Test Files:**
- `tests/regression/test_feature_parity.go` - Feature parity tests
- `tests/regression/test_output_format.go` - Output format tests

**Test Categories:**
- ✅ Output format comparison
- ✅ Feature parity verification
- ✅ Backward compatibility
- ✅ No breaking changes

---

## Test Coverage Status

### Tools (27 native implementations)

#### Fully Native Tools (5 tools)
- [x] `tool_catalog` - Has basic tests
- [ ] `workflow_mode` - Needs unit tests
- [ ] `git_tools` - Needs unit tests
- [ ] `infer_session_mode` - Needs unit tests
- [ ] `prompt_tracking` - Needs unit tests

#### Hybrid Tools - Recently Migrated (8 tools)
- [ ] `session` - Needs tests for prompts/assignee actions
- [ ] `ollama` - Needs tests for docs/quality/summary actions
- [ ] `recommend` - Needs tests for workflow action
- [ ] `setup_hooks` - Needs tests for patterns action
- [ ] `analyze_alignment` - Needs tests for prd action
- [ ] `health` - Needs tests for docs/dod/cicd actions
- [ ] `automation` - Has basic tests, needs comprehensive tests
- [ ] `estimation` - Needs tests for analyze action

#### Hybrid Tools - Previously Migrated (14 tools)
- [x] `analyze_alignment` - Has basic tests (todo2 action)
- [x] `automation` - Has basic tests
- [ ] Other tools - Need comprehensive tests

### Resources (21 native implementations)

- [x] `stdio://scorecard` - Has basic tests
- [x] `stdio://memories/*` - Has basic tests
- [ ] `stdio://prompts/*` - Needs tests
- [ ] `stdio://session/mode` - Needs tests
- [ ] Other resources - Need tests

---

## Priority Order

### Phase 1: Critical Tools (High Priority)
1. `session` - prompts/assignee actions
2. `health` - docs/dod/cicd actions
3. `automation` - nightly/sprint actions
4. `ollama` - docs/quality/summary actions

### Phase 2: Important Tools (Medium Priority)
5. `recommend` - workflow action
6. `setup_hooks` - patterns action
7. `analyze_alignment` - prd action
8. `estimation` - analyze action

### Phase 3: Remaining Tools (Lower Priority)
9. All other hybrid tools
10. Fully native tools

### Phase 4: Resources
11. All resource implementations

---

## Test Implementation Patterns

### Pattern 1: Unit Test Structure

```go
func TestHandleToolAction(t *testing.T) {
    tests := []struct {
        name      string
        params    map[string]interface{}
        wantError bool
        validate  func(*testing.T, []framework.TextContent)
    }{
        {
            name: "valid request",
            params: map[string]interface{}{
                "action": "action_name",
                "param1": "value1",
            },
            wantError: false,
            validate: func(t *testing.T, result []framework.TextContent) {
                if len(result) == 0 {
                    t.Error("expected non-empty result")
                }
                // Validate JSON format
                var data map[string]interface{}
                if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
                    t.Errorf("invalid JSON: %v", err)
                }
            },
        },
        // ... more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := context.Background()
            result, err := handleToolNative(ctx, tt.params)
            if (err != nil) != tt.wantError {
                t.Errorf("error = %v, wantError %v", err, tt.wantError)
                return
            }
            if !tt.wantError && tt.validate != nil {
                tt.validate(t, result)
            }
        })
    }
}
```

### Pattern 2: Integration Test Structure

```go
func TestToolViaMCP(t *testing.T) {
    server := fixtures.NewMockServer("test-server")
    err := RegisterAllTools(server)
    if err != nil {
        t.Fatalf("RegisterAllTools() error = %v", err)
    }
    
    ctx := context.Background()
    args := map[string]interface{}{
        "action": "action_name",
    }
    argsJSON, _ := json.Marshal(args)
    
    result, err := server.CallTool(ctx, "tool_name", argsJSON)
    if err != nil {
        t.Fatalf("CallTool() error = %v", err)
    }
    
    // Validate result
    // ...
}
```

### Pattern 3: Benchmark Structure

```go
func BenchmarkToolNative(b *testing.B) {
    ctx := context.Background()
    params := map[string]interface{}{
        "action": "action_name",
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = handleToolNative(ctx, params)
    }
}
```

---

## Test Execution

### Run All Tests
```bash
make test
```

### Run Unit Tests Only
```bash
make test-go
```

### Run Integration Tests
```bash
make test-integration
```

### Run Benchmarks
```bash
make go-bench
```

### Generate Coverage Report
```bash
make test-coverage-go
```

---

## Success Criteria

### Unit Tests
- ✅ 80%+ code coverage for all native implementations
- ✅ All actions tested
- ✅ Error cases tested
- ✅ Edge cases tested

### Integration Tests
- ✅ All tools work via MCP interface
- ✅ All resources work via MCP interface
- ✅ Parameter parsing works correctly
- ✅ Response format is valid

### Performance Benchmarks
- ✅ Native implementations are faster than Python bridge
- ✅ Performance improvements documented
- ✅ Memory usage is acceptable

### Regression Tests
- ✅ Feature parity verified
- ✅ Output format matches (where applicable)
- ✅ No breaking changes
- ✅ Backward compatibility maintained

---

## Timeline

- **Week 1:** Critical tools unit tests (session, health, automation, ollama)
- **Week 2:** Important tools unit tests (recommend, setup_hooks, analyze_alignment, estimation)
- **Week 3:** Remaining tools and resources unit tests
- **Week 4:** Integration tests
- **Week 5:** Performance benchmarks and regression testing

---

**Last Updated:** 2026-01-12  
**Status:** In Progress  
**Next Review:** After Phase 1 completion
