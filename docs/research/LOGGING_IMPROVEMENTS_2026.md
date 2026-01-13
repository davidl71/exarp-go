# Future Logging Improvements Research

**Date:** 2026-01-13  
**Status:** Research Complete  
**Project:** exarp-go

---

## Current State Analysis

### Current Logging Architecture

**MCP Framework Logger** (`mcp-go-core/pkg/mcp/logging`):
- ✅ Structured logging with levels (DEBUG, INFO, WARN, ERROR)
- ✅ Context-aware logging (request IDs, operation names)
- ✅ Performance logging (slow operation detection)
- ✅ Stderr output (MCP protocol compatible)
- ✅ Git hook suppression (GIT_HOOK=1 sets level to WARN via factory)
- ⚠️ Custom implementation (not standard library)
- ⚠️ Uses MCP_DEBUG env var (not GIT_HOOK directly)
- ⚠️ No JSON output support
- ✅ Used in `internal/factory/server.go` for MCP framework operations

**exarp-go Internal Logger** (`internal/logging/logger.go` - NEW):
- ✅ Uses slog (Go 1.21+ standard library)
- ✅ JSON output support (LOG_FORMAT=json)
- ✅ GIT_HOOK suppression
- ✅ Context-aware logging
- ✅ Performance tracking
- ⚠️ Duplicate of mcp-go-core functionality
- ⚠️ Used only for CLI operations (not MCP framework)

**exarp-go Internal Logger** (`internal/logging/logger.go` - NEW, Phases 1-3):
- ✅ Uses slog (Go 1.21+ standard library)
- ✅ JSON output support (LOG_FORMAT=json)
- ✅ GIT_HOOK suppression
- ✅ Context-aware logging
- ✅ Performance tracking
- ⚠️ Duplicate of mcp-go-core functionality
- ⚠️ Used only for CLI operations (not MCP framework)

**Standard Go Log Package Usage** (Legacy - being migrated):
- ⚠️ Previously 26 instances of `log.Printf()` in `internal/cli/` (now migrated to slog)
- ⚠️ No log level control
- ⚠️ Not suppressed in git hooks
- ⚠️ No structured logging
- ⚠️ No context information

**Issues Identified**:
1. **Triple Logging Systems**: 
   - mcp-go-core logger (used for MCP framework)
   - exarp-go internal logger (used for CLI, just created)
   - standard log package (legacy, being migrated)
2. **Code Duplication**: Two structured loggers with similar functionality
3. **Inconsistent Features**: mcp-go-core lacks JSON output and direct GIT_HOOK support
4. **Git Hook Inconsistency**: Different env vars (MCP_DEBUG vs GIT_HOOK)
5. **No Unified Approach**: Should consolidate to single logger

---

## Research Findings (2026)

### 1. Go Standard Library: `log/slog` (Go 1.21+)

**Source:** [Go 1.21 slog documentation](https://pkg.go.dev/log/slog)

**Key Features:**
- ✅ Built into Go standard library (no external dependencies)
- ✅ Structured logging with key-value pairs
- ✅ Multiple handlers (JSON, text)
- ✅ Log levels (DEBUG, INFO, WARN, ERROR)
- ✅ Context support
- ✅ Type-safe attributes

**Example:**
```go
import "log/slog"

handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
    Level: slog.LevelInfo,
})
logger := slog.New(handler)
slog.SetDefault(logger)

slog.Info("Database initialized", 
    "path", dbPath,
    "mode", "centralized",
)
```

**Benefits:**
- No external dependencies
- Standard library support
- JSON output for log aggregation
- Contextual logging with structured fields

---

### 2. High-Performance Logging: Zap

**Source:** [Uber Zap Documentation](https://github.com/uber-go/zap)

**Key Features:**
- ✅ Zero-allocation JSON logging
- ✅ Type-safe field construction
- ✅ High performance (minimal overhead)
- ✅ Structured logging
- ✅ Multiple output formats

**Use Case:** Production environments requiring maximum performance

**Trade-offs:**
- External dependency
- More complex API than slog
- Better for high-throughput scenarios

---

### 3. Zero-Allocation Logging: Zerolog

**Source:** [Zerolog Documentation](https://github.com/rs/zerolog)

**Key Features:**
- ✅ Zero-allocation JSON logging
- ✅ Simple, composable API
- ✅ Contextual logging
- ✅ Conditional logging
- ✅ Chainable API

**Use Case:** Applications primarily outputting JSON logs

**Trade-offs:**
- External dependency
- JSON-focused (less flexible than slog)

---

### 4. Structured Logging Best Practices

**Source:** [Honeycomb Logging Best Practices](https://www.honeycomb.io/blog/engineers-checklist-logging-best-practices)

**Key Recommendations:**
1. **Structured Logging**: Use JSON format for machine-readable logs
2. **Standardized Fields**: Use consistent field names (OpenTelemetry conventions)
3. **Appropriate Log Levels**: Use DEBUG, INFO, WARN, ERROR appropriately
4. **Context Information**: Include request IDs, operation names, user IDs
5. **Avoid Sensitive Data**: Mask passwords, API keys, PII
6. **Centralized Logging**: Aggregate logs for analysis
7. **Log Rotation**: Implement rotation and retention policies

---

### 5. MCP Server Logging Requirements

**Current Implementation:**
- ✅ Logs to stderr (MCP protocol requirement)
- ✅ JSON-RPC on stdout (protocol requirement)
- ✅ Context-aware logging (request IDs)
- ✅ Performance tracking (slow operations)

**Best Practices for MCP Servers:**
1. **Stderr for Logs**: All logs must go to stderr (stdout reserved for JSON-RPC)
2. **Structured Format**: JSON logs for easy parsing by MCP clients
3. **Request Tracing**: Include request IDs in all logs
4. **Performance Metrics**: Track tool call durations
5. **Error Context**: Include operation context in error logs

---

## Recommended Improvements

### Phase 1: Unify Logging Systems (High Priority)

**Goal:** Replace standard `log` package with structured logging

**Approach:**
1. **Migrate to `log/slog`** (Go 1.21+ standard library)
   - No external dependencies
   - Structured logging built-in
   - JSON output support
   - Context-aware logging

2. **Create Unified Logger Interface**
   ```go
   // internal/logging/logger.go
   package logging
   
   import "log/slog"
   
   // UnifiedLogger wraps slog with MCP-specific features
   type UnifiedLogger struct {
       *slog.Logger
       // MCP-specific features
   }
   ```

3. **Replace Standard Log Calls**
   - Replace `log.Printf()` with `slog.Info()`, `slog.Warn()`, etc.
   - Add structured fields (path, operation, context)
   - Maintain stderr output for MCP compatibility

**Benefits:**
- ✅ Single logging system
- ✅ Structured logs (JSON)
- ✅ Git hook suppression works for all logs
- ✅ Context-aware logging
- ✅ No external dependencies

---

### Phase 2: Enhanced Context Logging (Medium Priority)

**Goal:** Add request tracing and operation context

**Approach:**
1. **Request Context Propagation**
   ```go
   // Add request ID to all logs
   ctx := context.WithValue(ctx, "request_id", requestID)
   slog.InfoContext(ctx, "Tool call started", "tool", toolName)
   ```

2. **Operation Context**
   ```go
   // Include operation context
   slog.Info("Database initialized",
       "path", dbPath,
       "mode", "centralized",
       "operation", "database_init",
   )
   ```

3. **Performance Context**
   ```go
   // Track slow operations
   start := time.Now()
   // ... operation ...
   duration := time.Since(start)
   if duration > slowThreshold {
       slog.Warn("Slow operation",
           "operation", "database_query",
           "duration", duration,
           "threshold", slowThreshold,
       )
   }
   ```

**Benefits:**
- ✅ Better debugging (request tracing)
- ✅ Performance monitoring
- ✅ Operation context in all logs

---

### Phase 3: Git Hook Log Suppression (Low Priority)

**Goal:** Suppress all logs (including standard log) in git hooks

**Approach:**
1. **Check GIT_HOOK Environment Variable**
   ```go
   // In logging initialization
   if os.Getenv("GIT_HOOK") == "1" {
       // Set log level to WARN or ERROR
       handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
           Level: slog.LevelWarn,
       })
       slog.SetDefault(slog.New(handler))
   }
   ```

2. **Suppress Standard Log Package**
   ```go
   // Redirect standard log to discard in git hooks
   if os.Getenv("GIT_HOOK") == "1" {
       log.SetOutput(io.Discard)
   }
   ```

**Benefits:**
- ✅ Clean git hook output
- ✅ Reduced token usage in AI contexts
- ✅ Consistent behavior

---

### Phase 4: JSON Output Format (Medium Priority)

**Goal:** Output logs in JSON format for log aggregation

**Approach:**
1. **Use JSON Handler**
   ```go
   handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
       Level: slog.LevelInfo,
   })
   logger := slog.New(handler)
   ```

2. **Structured Fields**
   ```go
   slog.Info("Database initialized",
       "timestamp", time.Now().Format(time.RFC3339),
       "level", "INFO",
       "service", "exarp-go",
       "operation", "database_init",
       "path", dbPath,
       "mode", "centralized",
   )
   ```

**Benefits:**
- ✅ Machine-readable logs
- ✅ Easy log aggregation (ELK, Datadog, etc.)
- ✅ Better search and filtering

---

## Implementation Plan

### Task 1: Migrate Standard Log to slog
- Replace `log.Printf()` calls with `slog.Info()`, `slog.Warn()`, etc.
- Add structured fields to all log calls
- Maintain stderr output
- **Estimated:** 2-3 hours

### Task 2: Create Unified Logger Interface
- Create `internal/logging/logger.go` wrapper
- Provide MCP-specific logging helpers
- Support git hook suppression
- **Estimated:** 1-2 hours

### Task 3: Add Request Context Propagation
- Add request ID to context
- Propagate context through call chain
- Include context in all logs
- **Estimated:** 2-3 hours

### Task 4: Implement JSON Output
- Configure JSON handler
- Add structured fields to all logs
- Test with log aggregation tools
- **Estimated:** 1-2 hours

### Task 5: Git Hook Log Suppression
- Suppress standard log package in git hooks
- Ensure all logs respect GIT_HOOK=1
- Test git hook output
- **Estimated:** 1 hour

---

## References

### Go Logging Libraries
- **slog (Go 1.21+)**: https://pkg.go.dev/log/slog
- **Zap**: https://github.com/uber-go/zap
- **Zerolog**: https://github.com/rs/zerolog
- **Logrus**: https://github.com/sirupsen/logrus

### Best Practices
- **Honeycomb Logging Best Practices**: https://www.honeycomb.io/blog/engineers-checklist-logging-best-practices
- **Go Logging Best Practices**: https://www.dash0.com/guides/logging-best-practices
- **Structured Logging in Go**: https://peerdh.com/blogs/programming-insights/structured-logging-best-practices-in-go

### MCP Protocol
- **MCP Specification**: https://modelcontextprotocol.io
- **JSON-RPC 2.0**: https://www.jsonrpc.org/specification

---

## Decision Matrix

| Approach | Dependencies | Performance | Structured | Git Hook Support | Recommendation |
|----------|--------------|-------------|------------|-----------------|----------------|
| **slog (Go 1.21+)** | None (stdlib) | Good | ✅ JSON/Text | ✅ Yes | ⭐ **Recommended** |
| **Zap** | External | Excellent | ✅ JSON | ✅ Yes | Consider for high-throughput |
| **Zerolog** | External | Excellent | ✅ JSON only | ✅ Yes | Consider for JSON-only needs |
| **Current (dual)** | None | Good | ⚠️ Partial | ⚠️ Partial | ❌ Not recommended |

---

## Conclusion

**Recommended Approach:** Migrate to Go's standard library `log/slog` (Go 1.21+)

**Rationale:**
1. ✅ No external dependencies (standard library)
2. ✅ Structured logging built-in
3. ✅ JSON output support
4. ✅ Context-aware logging
5. ✅ Git hook suppression support
6. ✅ Future-proof (standard library)

**Migration Path:**
1. Phase 1: Replace standard log with slog (unify logging)
2. Phase 2: Add request context propagation
3. Phase 3: Implement JSON output
4. Phase 4: Enhance git hook suppression

**Estimated Total Effort:** 7-11 hours

---

**Next Steps:**
1. Review and approve this research
2. Create implementation tasks
3. Begin Phase 1 migration
4. Test with git hooks
5. Deploy and monitor
