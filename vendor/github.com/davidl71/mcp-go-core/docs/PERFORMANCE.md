# 🚀 Performance Optimizations

This document describes performance optimizations applied to `mcp-go-core` and best practices for optimal performance.

**Last Updated:** 2026-01-12

---

## Applied Optimizations

### 1. Fast Path for Empty Middleware Chains

**Location:** `pkg/mcp/framework/adapters/gosdk/middleware.go`

**Optimization:** Skip middleware wrapping if no middleware is registered.

**Before:**
```go
func (mc *MiddlewareChain) WrapToolHandler(handler ToolHandlerFunc) ToolHandlerFunc {
    wrapped := handler
    for i := len(mc.toolMiddlewares) - 1; i >= 0; i-- {
        wrapped = mc.toolMiddlewares[i](wrapped)
    }
    return wrapped
}
```

**After:**
```go
func (mc *MiddlewareChain) WrapToolHandler(handler ToolHandlerFunc) ToolHandlerFunc {
    if len(mc.toolMiddlewares) == 0 {
        return handler // Fast path: no middleware
    }
    // ... apply middleware
}
```

**Impact:** Eliminates unnecessary function wrapping when no middleware is used (common case).

---

### 2. Pre-allocated Slice Capacity

**Location:** `pkg/mcp/framework/adapters/gosdk/converters.go`, `adapter.go`

**Optimization:** Pre-allocate slices with exact capacity to avoid reallocations.

**Before:**
```go
result := make([]mcp.Content, 0)
for _, tc := range contents {
    result = append(result, &mcp.TextContent{Text: tc.Text})
}
```

**After:**
```go
result := make([]mcp.Content, len(contents))
for i := range contents {
    result[i] = &mcp.TextContent{Text: contents[i].Text}
}
```

**Impact:** Reduces allocations and improves memory efficiency.

---

### 3. Nil Slice Returns

**Location:** `pkg/mcp/framework/adapters/gosdk/adapter.go`

**Optimization:** Return nil slices instead of empty slices for better memory usage.

**Before:**
```go
if len(a.toolInfo) == 0 {
    return []types.ToolInfo{}
}
```

**After:**
```go
if len(a.toolInfo) == 0 {
    return nil // Better than empty slice
}
```

**Impact:** Nil slices use no memory, empty slices allocate header.

---

### 4. Optimized Context Validation

**Location:** `pkg/mcp/framework/adapters/gosdk/context.go`

**Optimization:** Use select for non-blocking context cancellation check.

**Before:**
```go
if ctx.Err() != nil {
    return ctx.Err()
}
```

**After:**
```go
select {
case <-ctx.Done():
    return fmt.Errorf("context cancelled: %w", ctx.Err())
default:
    return nil
}
```

**Impact:** More efficient context cancellation detection.

---

## Performance Characteristics

### Time Complexity

| Operation | Complexity | Notes |
|-----------|------------|-------|
| Tool Registration | O(1) | Map insertion |
| Tool Lookup | O(1) | Map lookup |
| Tool Execution | O(n) | n = middleware chain length (typically 0-3) |
| List Tools | O(n) | n = number of registered tools |
| Middleware Wrapping | O(1) | Fast path when no middleware |

### Space Complexity

| Operation | Complexity | Notes |
|-----------|------------|-------|
| Tool Storage | O(n) | n = number of tools |
| Middleware Chain | O(m) | m = number of middleware |
| Request Processing | O(1) | Constant space per request |

---

## Best Practices

### 1. Tool Registration

**Good:**
```go
// Register tools at startup (not per-request)
func setupServer() error {
    return server.RegisterTool("my_tool", "Description", schema, handler)
}
```

**Avoid:**
```go
// Don't register tools per-request
func handleRequest(req *Request) {
    server.RegisterTool("temp_tool", ...) // BAD
}
```

### 2. Middleware Usage

**Good:**
```go
// Use middleware for cross-cutting concerns
adapter := gosdk.NewGoSDKAdapter("server", "1.0.0",
    gosdk.WithMiddleware(&LoggingMiddleware{}),
)
```

**Avoid:**
```go
// Don't add unnecessary middleware
adapter := gosdk.NewGoSDKAdapter("server", "1.0.0",
    gosdk.WithMiddleware(&EmptyMiddleware{}), // BAD
)
```

### 3. Context Handling

**Good:**
```go
// Check context early
func handler(ctx context.Context, args json.RawMessage) ([]types.TextContent, error) {
    if err := ctx.Err(); err != nil {
        return nil, err
    }
    // ... process request
}
```

**Avoid:**
```go
// Don't ignore context cancellation
func handler(ctx context.Context, args json.RawMessage) ([]types.TextContent, error) {
    // Process without checking context - BAD
    time.Sleep(10 * time.Second) // Blocking operation
}
```

### 4. Memory Management

**Good:**
```go
// Reuse slices when possible
var buffer []byte
for _, item := range items {
    buffer = buffer[:0] // Reuse slice
    // ... process item
}
```

**Avoid:**
```go
// Don't allocate unnecessarily
for _, item := range items {
    buffer := make([]byte, 0) // New allocation each iteration - BAD
    // ... process item
}
```

---

## Benchmarking

### Running Benchmarks

```bash
cd pkg/mcp/framework/adapters/gosdk
go test -bench=. -benchmem
```

### Expected Results

- **Tool Registration:** < 1μs per tool
- **Tool Lookup:** < 100ns per lookup
- **Middleware Wrapping:** < 50ns per middleware (with fast path)
- **Context Validation:** < 10ns per check

---

## Performance Profiling

### CPU Profiling

```bash
go test -bench=. -cpuprofile=cpu.prof
go tool pprof cpu.prof
```

### Memory Profiling

```bash
go test -bench=. -memprofile=mem.prof
go tool pprof mem.prof
```

### Heap Profiling

```bash
go test -bench=. -memprofile=heap.prof
go tool pprof -http=:8080 heap.prof
```

---

## Future Optimizations

### Potential Improvements

1. **Pool for TextContent slices** - Reuse slice allocations
2. **Cache schema conversions** - Memoize ToolSchemaToMCP results
3. **Batch tool registration** - Reduce map rehashing
4. **Connection pooling** - For SSE/HTTP transports
5. **Request batching** - Process multiple requests together

### When to Optimize

- **Profile first** - Identify actual bottlenecks
- **Measure impact** - Ensure optimizations help
- **Maintain clarity** - Don't sacrifice readability
- **Test thoroughly** - Optimizations can introduce bugs

---

## Performance Monitoring

### Metrics to Track

1. **Tool call latency** - P50, P95, P99
2. **Memory usage** - Peak and average
3. **Goroutine count** - Detect leaks
4. **GC pauses** - Frequency and duration
5. **Error rate** - Failed tool calls

### Tools

- **pprof** - CPU and memory profiling
- **expvar** - Runtime metrics
- **prometheus** - Metrics collection (if needed)
- **tracing** - Request tracing (if needed)

---

*These optimizations provide baseline performance improvements. Profile your specific use case to identify additional optimization opportunities.*
