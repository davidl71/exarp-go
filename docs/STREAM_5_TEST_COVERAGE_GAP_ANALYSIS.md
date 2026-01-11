# Stream 5: Test Coverage Gap Analysis

**Date:** 2026-01-12  
**Status:** ✅ **ANALYSIS COMPLETE**  
**Objective:** Identify test coverage gaps for Stream 5 implementation

---

## Executive Summary

Analysis of test coverage gaps in native Go implementations. Identified **29+ tool implementations** missing unit tests. Existing test infrastructure is solid with mock servers and test helpers available.

---

## Current Test Coverage Status

### Test Files Found

**Existing Test Files (8 files):**
- `registry_test.go` - Tool registration tests ✅
- `handlers_test.go` - Handler tests ✅
- `memory_maint_test.go` - Memory maintenance tests ✅
- `statistics_test.go` - Statistics tests ✅
- `graph_helpers_test.go` - Graph operation tests ✅
- `critical_path_test.go` - Critical path tests ✅
- `apple_foundation_helpers_test.go` - Apple Foundation Models helpers ✅
- `apple_foundation_test.go` - Apple Foundation Models tests ✅

**Benchmark Files (4 files):**
- `statistics_bench_test.go` - Statistics benchmarks ✅
- `task_analysis_bench_test.go` - Task analysis benchmarks ✅
- `graph_helpers_bench_test.go` - Graph operation benchmarks ✅

### Source Files vs Test Files

**Total Source Files:** 49 Go files (excluding tests)
**Test Files:** 8 test files + 4 benchmark files = 12 test-related files
**Coverage Ratio:** ~24% of source files have dedicated tests

---

## Tools Missing Unit Tests

### High Priority (Core Tools)

These tools are frequently used and need comprehensive test coverage:

1. **`alignment_analysis.go`** - Task alignment analysis
   - Priority: High
   - Complexity: Medium
   - Dependencies: Task workflow, database
   - Actions to test: `todo2`, `prd`

2. **`attribution_check.go`** - Attribution compliance checking
   - Priority: High
   - Complexity: Medium
   - Dependencies: None
   - Actions to test: All actions

3. **`config_generator.go`** - IDE config generation
   - Priority: High
   - Complexity: Medium
   - Dependencies: File I/O
   - Actions to test: `rules`, `ignore`, `simplify`

4. **`external_tool_hints.go`** - External tool hints
   - Priority: High
   - Complexity: Low
   - Dependencies: File I/O
   - Actions to test: All actions

5. **`git_tools.go`** - Git operations
   - Priority: High
   - Complexity: Medium
   - Dependencies: Git commands
   - Actions to test: `commits`, `branches`, `tasks`, `diff`, `graph`, `merge`, `set_branch`

6. **`health_check.go`** - Health checks
   - Priority: High
   - Complexity: Medium
   - Dependencies: File system, git
   - Actions to test: `server`, `git`, `docs`, `dod`, `cicd`

7. **`hooks_setup.go`** - Git hooks setup
   - Priority: High
   - Complexity: Low
   - Dependencies: File I/O, git
   - Actions to test: `git`, `patterns`

8. **`memory.go`** - Memory operations
   - Priority: High
   - Complexity: Medium
   - Dependencies: Memory system
   - Actions to test: `save`, `recall`, `search`

9. **`report.go`** - Report generation
   - Priority: High
   - Complexity: High
   - Dependencies: Scorecard, database, tools
   - Actions to test: `overview`, `scorecard`, `briefing`, `prd`

10. **`security.go`** - Security scanning
    - Priority: High
    - Complexity: Medium
    - Dependencies: External security tools
    - Actions to test: `scan`, `alerts`, `report`

11. **`server_status.go`** - Server status
    - Priority: High
    - Complexity: Low
    - Dependencies: None
    - Actions to test: Status check

12. **`session.go`** - Session management
    - Priority: High
    - Complexity: Medium
    - Dependencies: Task workflow, prompts
    - Actions to test: `prime`, `handoff`, `prompts`, `assignee`
    - **Note:** Build error - missing `handleSessionPrompts` and `handleSessionAssignee`

13. **`session_mode_inference.go`** - Session mode inference
    - Priority: High
    - Complexity: Medium
    - Dependencies: Task workflow
    - Actions to test: Mode inference

14. **`task_workflow.go`** (native) - Task workflow management
    - Priority: High
    - Complexity: High
    - Dependencies: Database, task workflow
    - Actions to test: `sync`, `approve`, `clarify`, `clarity`, `cleanup`, `create`

### Medium Priority (Specialized Tools)

15. **`automation_native.go`** - Automation workflows
    - Priority: Medium
    - Complexity: High
    - Dependencies: Multiple tools
    - Actions to test: `daily`, `nightly`, `sprint`, `discover`

16. **`context.go`** / **`context_native.go`** - Context management
    - Priority: Medium
    - Complexity: Medium
    - Dependencies: None
    - Actions to test: `summarize`, `budget`, `batch`

17. **`estimation_native.go`** - Task estimation
    - Priority: Medium
    - Complexity: Medium
    - Dependencies: Database, statistics
    - Actions to test: `estimate`, `analyze`, `stats`

18. **`linting.go`** - Linting operations
    - Priority: Medium
    - Complexity: Medium
    - Dependencies: External linters
    - Actions to test: `run`, `analyze`

19. **`ollama_native.go`** - Ollama integration
    - Priority: Medium
    - Complexity: Medium
    - Dependencies: Ollama API
    - Actions to test: `status`, `models`, `generate`, `pull`, `hardware`, `docs`, `quality`, `summary`

20. **`prompt_tracking.go`** - Prompt tracking
    - Priority: Medium
    - Complexity: Low
    - Dependencies: Database
    - Actions to test: `log`, `analyze`

21. **`recommend.go`** - Recommendations
    - Priority: Medium
    - Complexity: Low
    - Dependencies: None
    - Actions to test: `model`, `workflow`, `advisor`

22. **`testing.go`** - Testing tool
    - Priority: Medium
    - Complexity: Medium
    - Dependencies: Go test, Python bridge
    - Actions to test: `run`, `coverage`, `suggest`, `validate`

23. **`tool_catalog.go`** - Tool catalog
    - Priority: Medium
    - Complexity: Low
    - Dependencies: None
    - Actions to test: `help`

24. **`workflow_mode.go`** - Workflow mode management
    - Priority: Medium
    - Complexity: Low
    - Dependencies: File I/O
    - Actions to test: `focus`, `suggest`, `stats`

### Lower Priority (Supporting/Platform-Specific)

25. **`scorecard_go.go`** - Go scorecard generation
    - Priority: Medium
    - Complexity: High
    - Dependencies: Go toolchain, file system
    - Note: Complex, may benefit from integration tests

26. **`scorecard_mlx.go`** - MLX scorecard generation
    - Priority: Low
    - Complexity: High
    - Dependencies: MLX, platform-specific
    - Note: Platform-specific, conditional compilation

27. **`report_mlx.go`** - MLX report generation
    - Priority: Low
    - Complexity: Medium
    - Dependencies: MLX, platform-specific
    - Note: Platform-specific

28. **`apple_foundation_registry.go`** / **`apple_foundation_registry_nocgo.go`** - Apple Foundation Models
    - Priority: Low
    - Complexity: Medium
    - Dependencies: Platform-specific
    - Note: Already has test files (`apple_foundation_test.go`)

29. **`todo2_db_adapter.go`** / **`todo2_utils.go`** - Todo2 database utilities
    - Priority: Medium
    - Complexity: Medium
    - Dependencies: Database
    - Note: May be tested indirectly via task_workflow tests

---

## Test Infrastructure Available

### Mock Server (`tests/fixtures/mock_server.go`)

**Features:**
- ✅ Mock MCPServer implementation
- ✅ Tool/Prompt/Resource registration
- ✅ Handler invocation for testing
- ✅ Thread-safe operations
- ✅ Count tracking

**Usage Pattern:**
```go
server := fixtures.NewMockServer("test-server")
err := RegisterAllTools(server)
// Test tools are registered
tool, exists := server.GetTool("tool_name")
// Test tool execution
result, err := server.CallTool(ctx, "tool_name", args)
```

### Test Helpers (`tests/fixtures/test_helpers.go`)

**Features:**
- ✅ Test context creation with timeout
- ✅ JSON-RPC request/response builders
- ✅ JSON-RPC error builders
- ✅ JSON conversion utilities

**Usage Pattern:**
```go
ctx, cancel := fixtures.TestContext(5 * time.Second)
defer cancel()
args := fixtures.MustToJSONRawMessage(map[string]interface{}{"action": "test"})
```

### Python Test Helpers (`tests/fixtures/test_helpers.py`)

**Features:**
- ✅ JSONRPCClient for stdio testing
- ✅ Server spawning utilities
- ✅ JSON-RPC helpers

---

## Test Coverage Goals

### Target Coverage: 80%+ for Native Implementations

**Current Estimate:**
- Tools with tests: ~8-10 tools (registry, handlers, memory_maint, statistics, graph_helpers, critical_path, apple_foundation)
- Tools without tests: ~29+ tools
- Coverage: ~25-30% estimated

**Gap Analysis:**
- Need tests for: ~29 tools
- Priority 1 (High): ~14 tools
- Priority 2 (Medium): ~11 tools
- Priority 3 (Low): ~4 tools

---

## Recommended Test Implementation Order

### Phase 1: High Priority Core Tools (Week 1-2)

1. `server_status.go` - Simple, no dependencies
2. `external_tool_hints.go` - Low complexity
3. `hooks_setup.go` - Low complexity
4. `attribution_check.go` - Medium complexity
5. `config_generator.go` - Medium complexity
6. `tool_catalog.go` - Low complexity
7. `workflow_mode.go` - Low complexity
8. `recommend.go` - Low complexity

### Phase 2: Medium Priority Tools (Week 2-3)

9. `git_tools.go` - Medium complexity
10. `health_check.go` - Medium complexity
11. `memory.go` - Medium complexity
12. `linting.go` - Medium complexity
13. `testing.go` - Medium complexity
14. `prompt_tracking.go` - Low complexity
15. `context.go` - Medium complexity
16. `session_mode_inference.go` - Medium complexity

### Phase 3: High Complexity Tools (Week 3-4)

17. `alignment_analysis.go` - Requires database
18. `report.go` - Complex, many dependencies
19. `security.go` - External dependencies
20. `automation_native.go` - Complex workflows
21. `session.go` - **Blocked by build error**
22. `task_workflow.go` - Complex, requires database
23. `estimation_native.go` - Requires database
24. `ollama_native.go` - External API dependencies

### Phase 4: Platform-Specific / Lower Priority (Week 4-5)

25. `scorecard_go.go` - Complex, may need integration tests
26. `scorecard_mlx.go` - Platform-specific
27. `report_mlx.go` - Platform-specific
28. `todo2_db_adapter.go` / `todo2_utils.go` - May be tested indirectly

---

## Test Pattern Recommendations

### Pattern 1: Simple Tool Test (No Dependencies)

```go
func TestHandleServerStatus(t *testing.T) {
    ctx := context.Background()
    args := fixtures.MustToJSONRawMessage(map[string]interface{}{})
    
    result, err := handleServerStatus(ctx, args)
    if err != nil {
        t.Fatalf("handleServerStatus() error = %v", err)
    }
    
    if len(result) == 0 {
        t.Error("handleServerStatus() returned empty result")
    }
}
```

### Pattern 2: Tool with Actions

```go
func TestHandleToolName(t *testing.T) {
    tests := []struct {
        name    string
        action  string
        wantErr bool
    }{
        {"valid action", "action1", false},
        {"invalid action", "invalid", true},
        {"default action", "", false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            args := fixtures.MustToJSONRawMessage(map[string]interface{}{
                "action": tt.action,
            })
            
            result, err := handleToolName(context.Background(), args)
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

### Pattern 3: Tool with Database Dependencies

```go
func TestHandleToolNameWithDatabase(t *testing.T) {
    // Setup: Create test database
    projectRoot := setupTestProject(t)
    defer cleanupTestProject(t, projectRoot)
    
    ctx := context.Background()
    args := fixtures.MustToJSONRawMessage(map[string]interface{}{
        "action": "test",
    })
    
    result, err := handleToolName(ctx, args)
    // ... assertions
}
```

---

## Action Items

### Immediate Actions

- [x] Complete coverage gap analysis (this document)
- [ ] Prioritize tool test implementation order
- [ ] Create test templates for each pattern
- [ ] Document test setup/teardown patterns
- [ ] Fix build error in `session.go` (Stream 1 dependency)

### Short-Term Actions (Week 1-2)

- [ ] Implement tests for Phase 1 tools (8 tools)
- [ ] Achieve 30-40% coverage
- [ ] Document test patterns
- [ ] Create test utilities if needed

### Medium-Term Actions (Week 3-4)

- [ ] Implement tests for Phase 2 tools (8 tools)
- [ ] Implement tests for Phase 3 tools (8 tools - after build error fixed)
- [ ] Achieve 60-70% coverage
- [ ] Add integration tests where needed

### Long-Term Actions (Week 5+)

- [ ] Implement tests for Phase 4 tools (4 tools)
- [ ] Achieve 80%+ coverage
- [ ] Performance benchmarks for all tools
- [ ] Regression test framework
- [ ] CI/CD integration

---

## Dependencies & Blockers

### Build Error (Blocks Unit Tests)

**File:** `internal/tools/session.go`
**Error:** Undefined functions `handleSessionPrompts` and `handleSessionAssignee`
**Impact:** Blocks all unit tests (compilation fails)
**Resolution:** Stream 1 dependency - complete session tool actions

### Test Dependencies

**Database Tests:**
- Need test database setup/teardown
- May need test fixtures for Todo2 data
- Consider using in-memory SQLite for tests

**External Dependencies:**
- Git commands (may need mock or test repo)
- External tools (may need mock or skip tests)
- API calls (may need mock or skip tests)

---

## Success Metrics

### Coverage Targets

- **Phase 1:** 30-40% coverage (8 tools)
- **Phase 2:** 50-60% coverage (+8 tools)
- **Phase 3:** 70-80% coverage (+8 tools)
- **Phase 4:** 80%+ coverage (+4 tools)

### Quality Targets

- All high-priority tools have tests
- All actions covered for each tool
- Error handling tested
- Edge cases covered
- Integration with dependencies tested

---

## References

- **Test Patterns:** `docs/MIGRATION_PATTERNS.md#testing-patterns`
- **Mock Server:** `tests/fixtures/mock_server.go`
- **Test Helpers:** `tests/fixtures/test_helpers.go`
- **Existing Tests:** `internal/tools/*_test.go`
- **Stream 5 Plan:** `docs/STREAM_5_TESTING_VALIDATION.md`

---

**Last Updated:** 2026-01-12  
**Status:** ✅ Analysis Complete - Ready for Implementation Planning
