# Lessons Learned from devwisdom-go Project

**Date:** 2026-01-07  
**Source:** `/Users/davidl/Projects/devwisdom-go`  
**Purpose:** Extract actionable lessons for model-assisted workflow and Go migration

---

## Executive Summary

The `devwisdom-go` project is a successful Go MCP server implementation that provides valuable patterns and lessons for:
1. **Go MCP Server Architecture** - Clean, simple, performant design
2. **Performance Optimization** - Massive improvements over Python (517x faster startup)
3. **Development Workflow** - Makefile, watchdog, testing patterns
4. **MCP Protocol Implementation** - Custom JSON-RPC 2.0 handler
5. **Dual-Mode Binary** - CLI + MCP server in same executable

**Key Takeaway:** A well-structured Go MCP server can achieve 500x+ performance improvements while maintaining simplicity and maintainability.

---

## Architecture Lessons

### 1. Simple Main Entry Point ✅

**Pattern:**
```go
// cmd/server/main.go
func main() {
    server := mcp.NewWisdomServer()
    if err := server.Run(context.Background(), os.Stdin, os.Stdout); err != nil {
        log.Fatalf("Server error: %v", err)
    }
}
```

**Lessons:**
- ✅ **Minimal main function** - Just create server and run
- ✅ **Explicit stdio** - Pass `os.Stdin` and `os.Stdout` explicitly
- ✅ **Context usage** - Always use `context.Background()` for root context
- ✅ **Error handling** - Fail fast with `log.Fatalf` on startup errors

**Apply to exarp-go:**
- Keep `main.go` simple - delegate all logic to internal packages
- Use explicit stdio transport (not hidden in framework)
- Initialize components before starting server loop

### 2. Clean Internal Structure ✅

**Pattern:**
```
internal/
├── mcp/          # MCP protocol handler
├── wisdom/       # Business logic (wisdom engine)
├── config/       # Configuration management
├── logging/      # Structured logging
└── cli/          # CLI commands (separate from server)
```

**Lessons:**
- ✅ **Package separation** - Each package has single responsibility
- ✅ **No circular dependencies** - Clean dependency graph
- ✅ **Internal visibility** - Use `internal/` to prevent external imports
- ✅ **Domain-driven** - Packages organized by domain, not layer

**Apply to exarp-go:**
```
internal/
├── framework/    # Framework abstraction (already planned)
├── tools/        # Tool implementations
├── prompts/      # Prompt handlers
├── resources/    # Resource handlers
├── bridge/       # Python bridge (for MLX/tools)
├── models/       # Model integration (NEW for model-assisted workflow)
│   ├── router.go      # Model selection
│   ├── mlx.go         # MLX handler
│   ├── ollama.go      # Ollama handler
│   ├── breakdown.go   # Task breakdown
│   └── execution.go   # Auto-execution
└── config/      # Configuration
```

### 3. Custom JSON-RPC Implementation ✅

**Pattern:**
- Implemented JSON-RPC 2.0 manually (not using framework)
- Full control over protocol handling
- Custom error handling and request routing

**Lessons:**
- ✅ **Full control** - Can customize protocol behavior
- ✅ **No framework lock-in** - Not dependent on external MCP framework
- ✅ **Explicit error handling** - Clear error codes and messages
- ✅ **Compact JSON** - No indentation for stdio compatibility

**Trade-offs:**
- ⚠️ **More code** - Must implement protocol yourself
- ⚠️ **Maintenance** - Must keep up with MCP spec changes
- ⚠️ **Testing** - Must test protocol implementation

**Decision for exarp-go:**
- **Recommendation:** Use official Go SDK initially (faster development)
- **Future:** Consider custom implementation if framework becomes limiting
- **Hybrid:** Use SDK but wrap with custom error handling/logging

### 4. Dual-Mode Binary (CLI + MCP Server) ✅

**Pattern:**
- Same binary can run as CLI (when TTY) or MCP server (when stdio)
- Detects mode based on `isatty(stdin)`

**Lessons:**
- ✅ **Single binary** - One executable for both use cases
- ✅ **Mode detection** - Automatic based on stdin type
- ✅ **Shared code** - Business logic reused between CLI and server
- ✅ **User experience** - CLI provides convenient access to functionality

**Apply to exarp-go:**
- Consider adding CLI mode for:
  - Direct tool invocation (testing)
  - Task breakdown (interactive)
  - Prompt optimization (iterative refinement)
  - Model testing (benchmarking)

**Example:**
```go
// cmd/exarp-go/main.go
func main() {
    if isatty(os.Stdin.Fd()) {
        // CLI mode
        cli.Run()
    } else {
        // MCP server mode
        server.Run(context.Background(), os.Stdin, os.Stdout)
    }
}
```

---

## Performance Lessons

### 1. Massive Performance Gains ✅

**Results from devwisdom-go:**
- **Startup time:** 0.387ms (vs Python ~200ms) = **517x faster**
- **Response time:** 0.01748µs (vs Python ~5-10ms) = **286,000-572,000x faster**
- **Memory:** < 1KB per operation (vs Python ~1-2KB)
- **Binary size:** < 10MB (self-contained)

**Lessons:**
- ✅ **Compiled language advantage** - Go's compiled nature provides massive speedups
- ✅ **Zero allocations** - Hot paths can achieve zero allocations
- ✅ **Map lookups** - O(1) lookups are extremely fast
- ✅ **Minimal overhead** - No interpreter or runtime startup cost

**Apply to exarp-go:**
- **Expect similar gains** - Python bridge will add overhead, but still much faster
- **Optimize hot paths** - Tool registration, request routing
- **Benchmark early** - Establish baseline, measure improvements
- **Profile regularly** - Use `go tool pprof` to find bottlenecks

### 2. Benchmark-Driven Development ✅

**Pattern:**
- Comprehensive benchmark suite
- Performance goals defined upfront
- Regular benchmarking during development

**Lessons:**
- ✅ **Define goals** - Set performance targets before implementation
- ✅ **Benchmark early** - Establish baseline immediately
- ✅ **Profile hot paths** - Focus optimization where it matters
- ✅ **Document results** - Keep performance docs up to date

**Apply to exarp-go:**
- Set performance goals:
  - Startup: < 100ms (with Python bridge)
  - Tool execution: < 50ms (including bridge overhead)
  - Model calls: < 500ms (local models)
- Create benchmark suite:
  - Tool registration
  - Request routing
  - Python bridge calls
  - Model inference

### 3. Zero-Allocation Hot Paths ✅

**Pattern:**
- Critical paths (GetWisdom, GetAeonLevel) achieve zero allocations
- Returns pointers instead of copying structs
- Reuses structures where possible

**Lessons:**
- ✅ **Return pointers** - Avoid copying large structs
- ✅ **Reuse structures** - Pool allocations for high-throughput
- ✅ **Map lookups** - O(1) lookups with minimal overhead
- ✅ **Minimal allocations** - Hot paths should allocate as little as possible

**Apply to exarp-go:**
- **Tool handlers** - Return pointers to results
- **Request routing** - Use map-based routing (O(1))
- **Model responses** - Reuse response structures
- **Python bridge** - Minimize JSON marshaling/unmarshaling

---

## Development Workflow Lessons

### 1. Comprehensive Makefile ✅

**Pattern:**
- Build targets (server, CLI, all platforms)
- Test targets (unit, coverage, HTML)
- Benchmark targets (CPU, memory profiling)
- Cross-compilation targets
- Release packaging

**Lessons:**
- ✅ **Standard targets** - `build`, `test`, `clean`, `install`
- ✅ **Development targets** - `watchdog`, `fmt`, `lint`
- ✅ **Profiling targets** - `bench-cpu`, `bench-mem`, `pprof-web`
- ✅ **Release targets** - `build-release`, `build-all-platforms`

**Apply to exarp-go:**
```makefile
# Add model-specific targets
bench-models:          # Benchmark model inference
test-models:           # Test model integration
watchdog:              # Development hot reload
build-release:         # Cross-platform release
```

### 2. Watchdog for Development ✅

**Pattern:**
- Shell script monitors server for crashes
- File watching for hot reload
- Auto-rebuild on source changes

**Lessons:**
- ✅ **Crash detection** - Auto-restart on crashes
- ✅ **File watching** - Reload on config/data changes
- ✅ **Auto-rebuild** - Rebuild binary on Go source changes
- ✅ **Graceful shutdown** - Handle SIGTERM properly

**Apply to exarp-go:**
- Create `watchdog.sh` for development
- Watch:
  - Go source files (rebuild)
  - Config files (reload)
  - Python bridge scripts (restart)
  - Model configs (reload)

### 3. Comprehensive Testing ✅

**Pattern:**
- Unit tests for all packages
- Integration tests for MCP protocol
- Benchmark tests for performance
- Coverage reporting (text + HTML)

**Lessons:**
- ✅ **Table-driven tests** - Go idiom for test cases
- ✅ **Integration tests** - Test full MCP protocol flow
- ✅ **Benchmark tests** - Performance regression detection
- ✅ **Coverage goals** - Aim for >80% coverage

**Apply to exarp-go:**
- Test structure:
  - Unit tests for each package
  - Integration tests for MCP protocol
  - Bridge tests (Python execution)
  - Model tests (MLX/Ollama integration)
  - End-to-end tests (full workflow)

---

## MCP Protocol Lessons

### 1. Compact JSON for STDIO ✅

**Pattern:**
```go
encoder.SetIndent("", "") // No indentation
```

**Lessons:**
- ✅ **Compact JSON** - Some clients have issues with indented JSON
- ✅ **Explicit setting** - Don't rely on defaults
- ✅ **Compatibility** - Better compatibility with MCP clients

**Apply to exarp-go:**
- Always use compact JSON for stdio transport
- Test with multiple MCP clients (Cursor, others)

### 2. Proper Error Handling ✅

**Pattern:**
- JSON-RPC 2.0 error codes
- Parse errors (id = null)
- Method not found
- Invalid parameters
- Internal errors

**Lessons:**
- ✅ **Standard error codes** - Follow JSON-RPC 2.0 spec
- ✅ **Parse errors** - Special handling (id must be null)
- ✅ **Error messages** - Clear, actionable error messages
- ✅ **Error data** - Include additional context in error.data

**Apply to exarp-go:**
- Implement proper JSON-RPC error handling
- Include helpful error messages
- Add error context for debugging

### 3. Request Logging & Tracking ✅

**Pattern:**
- Log all requests with unique IDs
- Measure request duration
- Track request/response pairs

**Lessons:**
- ✅ **Request IDs** - Format and track request IDs
- ✅ **Duration tracking** - Measure request processing time
- ✅ **Structured logging** - Use structured logger
- ✅ **Debug logging** - Optional debug-level logging

**Apply to exarp-go:**
- Add request logging middleware
- Track:
  - Request ID
  - Method name
  - Duration
  - Error status
  - Model calls (if applicable)

---

## Model Integration Lessons (NEW)

### 1. Model Router Pattern ✅

**From devwisdom-go structure:**
- Clean separation of concerns
- Interface-based design
- Easy to add new models

**Apply to model-assisted workflow:**
```go
// internal/models/router.go
type ModelType string

const (
    ModelMLXCodeLlama ModelType = "mlx-codellama"
    ModelMLXPhi35     ModelType = "mlx-phi35"
    ModelOllamaLlama  ModelType = "ollama-llama"
    ModelOllamaCode   ModelType = "ollama-codellama"
)

type ModelRouter interface {
    SelectModel(taskType TaskType, requirements ModelRequirements) ModelType
    CallModel(ctx context.Context, model ModelType, prompt string) (string, error)
}
```

### 2. Performance Considerations ✅

**From devwisdom-go benchmarks:**
- Local operations are extremely fast
- Model inference will be the bottleneck

**Apply to model-assisted workflow:**
- **Model calls:** Expect 100-500ms per call
- **Task breakdown:** Can run in parallel for multiple tasks
- **Auto-execution:** Batch simple tasks for efficiency
- **Prompt optimization:** Iterative refinement (3-5 iterations)

### 3. Error Handling for Models ✅

**From devwisdom-go error patterns:**
- Graceful degradation
- Clear error messages
- Fallback strategies

**Apply to model-assisted workflow:**
- **Model unavailable:** Fall back to standard AI processing
- **Model timeout:** Use shorter timeout, fall back
- **Model errors:** Log error, continue without model assistance
- **Invalid responses:** Validate model output, retry if needed

---

## Configuration Lessons

### 1. Minimal Dependencies ✅

**Pattern:**
```go
// go.mod
module github.com/davidl71/devwisdom-go
go 1.21
// No external dependencies!
```

**Lessons:**
- ✅ **Standard library first** - Use Go stdlib when possible
- ✅ **Minimal deps** - Only add dependencies when necessary
- ✅ **Easy builds** - Faster builds, fewer vulnerabilities
- ✅ **Self-contained** - Binary includes everything needed

**Apply to exarp-go:**
- **Minimize dependencies:**
  - Go SDK (required for MCP)
  - Python bridge (required for tools)
  - HTTP client (for Ollama, if not using bridge)
  - Avoid: Heavy frameworks, unnecessary abstractions

### 2. Environment-Based Configuration ✅

**Pattern:**
- Environment variables for configuration
- File-based config (sources.json)
- Sensible defaults

**Lessons:**
- ✅ **Environment variables** - Easy to override
- ✅ **Config files** - For complex configuration
- ✅ **Defaults** - Sensible defaults for all settings
- ✅ **Validation** - Validate config on startup

**Apply to exarp-go:**
- **Configuration:**
  - Framework selection (env var or config file)
  - Model selection (config file)
  - Python bridge path (env var)
  - Model timeouts (config file)

---

## Testing Lessons

### 1. Table-Driven Tests ✅

**Pattern:**
```go
func TestGetWisdom(t *testing.T) {
    tests := []struct {
        name     string
        source   string
        score    int
        wantErr  bool
    }{
        {"valid source", "stoic", 75, false},
        {"invalid source", "invalid", 75, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

**Lessons:**
- ✅ **Go idiom** - Standard Go testing pattern
- ✅ **Easy to extend** - Add test cases easily
- ✅ **Clear structure** - Test cases are self-documenting
- ✅ **Parallel execution** - Can run subtests in parallel

**Apply to exarp-go:**
- Use table-driven tests for:
  - Tool handlers
  - Model router
  - Task breakdown
  - Prompt optimization

### 2. Integration Tests ✅

**Pattern:**
- Test full MCP protocol flow
- Test with real JSON-RPC messages
- Test error scenarios

**Lessons:**
- ✅ **End-to-end** - Test complete request/response cycle
- ✅ **Real protocol** - Use actual JSON-RPC messages
- ✅ **Error scenarios** - Test error handling paths
- ✅ **Edge cases** - Test boundary conditions

**Apply to exarp-go:**
- Integration tests for:
  - MCP protocol (initialize, tools, prompts, resources)
  - Python bridge (tool execution)
  - Model integration (MLX, Ollama)
  - Full workflow (task breakdown → execution)

---

## Documentation Lessons

### 1. Comprehensive README ✅

**Pattern:**
- Quick start
- Usage examples
- CLI documentation
- Architecture overview
- Performance benchmarks

**Lessons:**
- ✅ **Quick start** - Get users running in < 5 minutes
- ✅ **Examples** - Real-world usage examples
- ✅ **API docs** - Document all public APIs
- ✅ **Performance** - Include benchmark results

**Apply to exarp-go:**
- README should include:
  - Quick start (build, run, configure)
  - Tool usage examples
  - Model integration guide
  - Performance benchmarks
  - Migration guide (Python → Go)

### 2. Performance Documentation ✅

**Pattern:**
- Dedicated `docs/PERFORMANCE.md`
- Benchmark results
- Comparison with Python
- Optimization opportunities

**Lessons:**
- ✅ **Track performance** - Document performance characteristics
- ✅ **Compare versions** - Show improvements over Python
- ✅ **Optimization guide** - Document optimization opportunities
- ✅ **Real-world metrics** - Include actual usage metrics

**Apply to exarp-go:**
- Create `docs/PERFORMANCE.md`:
  - Benchmark results
  - Python bridge overhead
  - Model inference times
  - Optimization opportunities

---

## Actionable Recommendations

### Immediate Actions (Phase 1)

1. **Adopt Simple Main Pattern**
   - Keep `main.go` minimal
   - Delegate to internal packages
   - Explicit stdio transport

2. **Create Model Package Structure**
   - `internal/models/` package
   - Router, handlers, breakdown, execution
   - Follow devwisdom-go's clean structure

3. **Add Benchmark Suite**
   - Define performance goals
   - Create benchmark tests
   - Establish baseline

4. **Create Comprehensive Makefile**
   - Build, test, benchmark targets
   - Watchdog for development
   - Cross-compilation support

### Short-Term Actions (Phase 2-3)

5. **Implement Request Logging**
   - Structured logging
   - Request ID tracking
   - Duration measurement

6. **Add Integration Tests**
   - MCP protocol tests
   - Python bridge tests
   - Model integration tests

7. **Create Performance Documentation**
   - Benchmark results
   - Comparison with Python
   - Optimization guide

### Long-Term Actions (Phase 4+)

8. **Consider Dual-Mode Binary**
   - CLI mode for testing
   - Interactive model tools
   - Direct tool invocation

9. **Optimize Hot Paths**
   - Zero allocations where possible
   - Map-based routing
   - Pointer returns

10. **Add Watchdog Script**
    - Crash detection
    - File watching
    - Auto-rebuild

---

## Key Takeaways

### Architecture
- ✅ **Simple main** - Keep entry point minimal
- ✅ **Clean structure** - Package by domain, not layer
- ✅ **Explicit stdio** - Pass stdin/stdout explicitly

### Performance
- ✅ **Massive gains** - Expect 500x+ improvements
- ✅ **Benchmark early** - Establish baseline immediately
- ✅ **Zero allocations** - Optimize hot paths

### Development
- ✅ **Comprehensive Makefile** - Standard targets + profiling
- ✅ **Watchdog** - Development hot reload
- ✅ **Table-driven tests** - Go testing idiom

### MCP Protocol
- ✅ **Compact JSON** - No indentation for stdio
- ✅ **Proper errors** - Follow JSON-RPC 2.0 spec
- ✅ **Request logging** - Track all requests

### Models
- ✅ **Router pattern** - Clean model selection
- ✅ **Performance aware** - Model calls are bottleneck
- ✅ **Graceful degradation** - Fall back on errors

---

## References

- **devwisdom-go:** `/Users/davidl/Projects/devwisdom-go`
- **Performance Docs:** `docs/PERFORMANCE.md`
- **MCP Implementation:** `internal/mcp/server.go`
- **Makefile:** `Makefile`
- **Benchmarks:** `internal/wisdom/benchmark_test.go`

---

**Status:** Ready to apply lessons to exarp-go migration and model-assisted workflow implementation.

