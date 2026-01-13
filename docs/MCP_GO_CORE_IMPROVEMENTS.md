# 🚀 mcp-go-core Improvements for exarp-go

**Date:** 2026-01-13  
**Status:** Analysis and Recommendations

---

## Summary

This document identifies improvements from `mcp-go-core` v0.3.0 that could benefit `exarp-go`. Currently, `exarp-go` uses `mcp-go-core` v0.3.0 but maintains its own adapter implementation instead of using the library's adapter directly.

---

## Current State

### ✅ What's Already Integrated

- ✅ **Framework Interface** - Uses `mcp-go-core/pkg/mcp/framework` interfaces
- ✅ **Security Utilities** - Uses `mcp-go-core/pkg/mcp/security` (with local extensions)
- ✅ **Types** - Uses `mcp-go-core/pkg/mcp/types`
- ✅ **Factory Pattern** - Uses `mcp-go-core/pkg/mcp/factory`

### ⚠️ What's Still Local

- ⚠️ **Adapter Implementation** - `internal/framework/adapters/gosdk/adapter.go` is a local copy, not using `mcp-go-core/pkg/mcp/framework/adapters/gosdk` directly

---

## Key Improvements Available in mcp-go-core

### 1. **Adapter Options Pattern** ⭐ **HIGH PRIORITY**

**What:** The mcp-go-core adapter supports an options pattern for configuration.

**Current (exarp-go):**
```go
// internal/factory/server.go
return gosdk.NewGoSDKAdapter(name, version), nil
```

**Improved (mcp-go-core):**
```go
// With options support
import (
    "github.com/davidl71/mcp-go-core/pkg/mcp/framework/adapters/gosdk"
    "github.com/davidl71/mcp-go-core/pkg/mcp/logging"
)

logger := logging.NewLogger()
adapter := gosdk.NewGoSDKAdapter(name, version,
    gosdk.WithLogger(logger),
    gosdk.WithMiddleware(myMiddleware), // Optional
)
```

**Benefits:**
- ✅ Built-in logging support
- ✅ Optional middleware configuration
- ✅ Extensible for future options
- ✅ Better testability

**Files to Update:**
- `internal/factory/server.go` - Use mcp-go-core adapter directly

---

### 2. **Validation Helpers** ⭐ **HIGH PRIORITY**

**What:** Reusable validation functions in `validation.go`.

**Current (exarp-go):**
```go
// Manual validation in adapter
if name == "" {
    return fmt.Errorf("tool name cannot be empty")
}
if description == "" {
    return fmt.Errorf("tool description cannot be empty")
}
// ... repeated in RegisterTool, RegisterPrompt, RegisterResource
```

**Improved (mcp-go-core):**
```go
import "github.com/davidl71/mcp-go-core/pkg/mcp/framework/adapters/gosdk"

// Reusable validation
if err := gosdk.ValidateRegistration(name, description, handler); err != nil {
    return fmt.Errorf("tool registration: %w", err)
}

// Request validation
if err := gosdk.ValidateCallToolRequest(req); err != nil {
    return nil, err
}
```

**Available Functions:**
- `ValidateRegistration()` - Common registration validation
- `ValidateResourceRegistration()` - Resource-specific validation
- `ValidateCallToolRequest()` - Tool request validation
- `ValidateGetPromptRequest()` - Prompt request validation
- `ValidateReadResourceRequest()` - Resource request validation

**Benefits:**
- ✅ Consistent validation across all handlers
- ✅ Reduced code duplication
- ✅ Easier to maintain
- ✅ Better error messages

---

### 3. **Context Validation Helper** ⭐ **MEDIUM PRIORITY**

**What:** Optimized context validation function.

**Current (exarp-go):**
```go
if ctx.Err() != nil {
    return fmt.Errorf("context cancelled: %w", ctx.Err())
}
```

**Improved (mcp-go-core):**
```go
import "github.com/davidl71/mcp-go-core/pkg/mcp/framework/adapters/gosdk"

// Optimized context validation
if err := gosdk.ValidateContext(ctx); err != nil {
    return nil, err
}
```

**Benefits:**
- ✅ More efficient context cancellation detection
- ✅ Consistent error handling
- ✅ Better performance (non-blocking select pattern)

---

### 4. **Converter Helpers** ⭐ **MEDIUM PRIORITY**

**What:** Helper functions for type conversion.

**Current (exarp-go):**
```go
// Manual conversion in adapter
inputSchemaMap := map[string]interface{}{
    "type":       schema.Type,
    "properties": schema.Properties,
}
if len(schema.Required) > 0 {
    inputSchemaMap["required"] = schema.Required
}

// Manual TextContent conversion
contents := []mcp.Content{}
for _, content := range result {
    contents = append(contents, &mcp.TextContent{Text: content.Text})
}
```

**Improved (mcp-go-core):**
```go
import "github.com/davidl71/mcp-go-core/pkg/mcp/framework/adapters/gosdk"

// Schema conversion
inputSchemaMap := gosdk.ToolSchemaToMCP(schema)

// TextContent conversion
contents := gosdk.TextContentToMCP(result)
```

**Benefits:**
- ✅ Consistent conversion logic
- ✅ Reduced code duplication
- ✅ Pre-allocated slices for better performance
- ✅ Easier to maintain

---

### 5. **Middleware Support** ⭐ **LOW PRIORITY** (Future Enhancement)

**What:** Built-in middleware chain support for tools, prompts, and resources.

**Current:** Not available in exarp-go's adapter

**Improved (mcp-go-core):**
```go
import (
    "github.com/davidl71/mcp-go-core/pkg/mcp/framework/adapters/gosdk"
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Define middleware
loggingMiddleware := func(next gosdk.ToolHandlerFunc) gosdk.ToolHandlerFunc {
    return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        log.Printf("Tool call: %s", req.Params.Name)
        return next(ctx, req)
    }
}

// Use with adapter
adapter := gosdk.NewGoSDKAdapter(name, version,
    gosdk.WithMiddleware(loggingMiddleware),
)
```

**Benefits:**
- ✅ Cross-cutting concerns (logging, metrics, auth)
- ✅ Clean separation of concerns
- ✅ Fast path when no middleware (performance optimized)
- ✅ Composable middleware chains

**Use Cases:**
- Request logging
- Performance metrics
- Authentication/authorization
- Error handling
- Request/response transformation

---

### 6. **Logging Integration** ⭐ **MEDIUM PRIORITY**

**What:** Built-in structured logging support.

**Current:** No logging in exarp-go's adapter

**Improved (mcp-go-core):**
```go
import (
    "github.com/davidl71/mcp-go-core/pkg/mcp/framework/adapters/gosdk"
    "github.com/davidl71/mcp-go-core/pkg/mcp/logging"
)

logger := logging.NewLogger()
adapter := gosdk.NewGoSDKAdapter(name, version,
    gosdk.WithLogger(logger),
)

// Adapter automatically logs:
// - Tool registration (Debug level)
// - Tool registration success (Info level)
// - Prompt registration (Debug level)
// - Resource registration (Debug level)
```

**Benefits:**
- ✅ Consistent logging across all operations
- ✅ Debug-level logging for registration
- ✅ Info-level logging for successful operations
- ✅ Easy to enable/disable via logger configuration

---

### 7. **Performance Optimizations** ⭐ **MEDIUM PRIORITY**

**What:** Performance improvements in mcp-go-core v0.3.0.

**Optimizations:**
1. **Fast Path for Empty Middleware Chains** - Skip wrapping when no middleware
2. **Pre-allocated Slice Capacity** - Avoid reallocations
3. **Nil Slice Returns** - Better memory usage
4. **Optimized Context Validation** - Non-blocking select pattern

**Impact:**
- ✅ Faster tool execution (no middleware overhead when unused)
- ✅ Reduced memory allocations
- ✅ Better context cancellation detection

**Reference:** See `mcp-go-core/docs/PERFORMANCE.md` for details.

---

### 8. **Better Error Handling** ⭐ **MEDIUM PRIORITY**

**What:** More consistent and detailed error messages.

**Current (exarp-go):**
```go
if err != nil {
    return fmt.Errorf("tool execution error: %v", err)
}
```

**Improved (mcp-go-core):**
```go
// Validation errors include context
if err := gosdk.ValidateRegistration(name, description, handler); err != nil {
    return fmt.Errorf("tool registration: %w", err) // Wrapped with context
}
```

**Benefits:**
- ✅ Error wrapping with context
- ✅ Consistent error format
- ✅ Better debugging information

---

## Migration Strategy

### Option 1: Direct Migration (Recommended) ⭐

**Replace local adapter with mcp-go-core adapter directly.**

**Steps:**
1. Remove `internal/framework/adapters/gosdk/adapter.go` (local copy)
2. Update `internal/factory/server.go` to use `mcp-go-core/pkg/mcp/framework/adapters/gosdk` directly
3. Add optional logging support (if desired)
4. Update tests if needed

**Benefits:**
- ✅ Single source of truth (no duplication)
- ✅ Automatic updates when mcp-go-core improves
- ✅ Reduced maintenance burden
- ✅ Access to all features (options, middleware, validation helpers)

**Drawbacks:**
- ⚠️ Requires testing to ensure compatibility
- ⚠️ May need to migrate any local customizations

### Option 2: Gradual Adoption

**Adopt features incrementally while keeping local adapter.**

**Steps:**
1. Import validation helpers from mcp-go-core
2. Import converter helpers from mcp-go-core
3. Add logging support using mcp-go-core logger
4. Consider middleware support later
5. Eventually migrate to direct adapter usage (Option 1)

**Benefits:**
- ✅ Lower risk (gradual migration)
- ✅ Can test each feature independently
- ✅ Keep local customizations temporarily

**Drawbacks:**
- ⚠️ Still maintaining duplicate code
- ⚠️ Missing some optimizations
- ⚠️ More work long-term

---

## Recommended Actions

### High Priority (Do First)

1. **✅ Use mcp-go-core adapter directly** (Option 1)
   - Remove local adapter copy
   - Update factory to use mcp-go-core adapter
   - Add optional logging support
   - Test thoroughly

2. **✅ Use validation helpers**
   - Replace manual validation with `ValidateRegistration()`
   - Use request validation helpers
   - Reduce code duplication

3. **✅ Use converter helpers**
   - Replace manual conversions with `ToolSchemaToMCP()` and `TextContentToMCP()`
   - Improve consistency

### Medium Priority (Do Next)

4. **✅ Add logging support**
   - Use `WithLogger()` option
   - Enable debug-level logging for development
   - Improve observability

5. **✅ Use context validation helper**
   - Replace manual context checks with `ValidateContext()`
   - Better performance

### Low Priority (Future)

6. **✅ Consider middleware support**
   - Add middleware for cross-cutting concerns
   - Enable request logging, metrics, etc.
   - Only if needed for specific use cases

---

## Implementation Checklist

- [ ] Review current adapter implementation
- [ ] Identify any local customizations that need preservation
- [ ] Test mcp-go-core adapter compatibility
- [ ] Update `internal/factory/server.go` to use mcp-go-core adapter
- [ ] Add optional logging support
- [ ] Remove local adapter copy (`internal/framework/adapters/gosdk/adapter.go`)
- [ ] Update tests to work with mcp-go-core adapter
- [ ] Verify all tools/prompts/resources still work
- [ ] Update documentation
- [ ] Consider adding middleware for logging/metrics (optional)

---

## References

- **mcp-go-core Repository:** `github.com/davidl71/mcp-go-core`
- **mcp-go-core Version:** v0.3.0
- **Migration Guide:** `mcp-go-core/docs/MIGRATION_GUIDE.md`
- **Performance Guide:** `mcp-go-core/docs/PERFORMANCE.md`
- **Adapter Options:** `mcp-go-core/pkg/mcp/framework/adapters/gosdk/options.go`
- **Middleware:** `mcp-go-core/pkg/mcp/framework/adapters/gosdk/middleware.go`
- **Validation Helpers:** `mcp-go-core/pkg/mcp/framework/adapters/gosdk/validation.go`
- **Converter Helpers:** `mcp-go-core/pkg/mcp/framework/adapters/gosdk/converters.go`

---

## Questions?

- Are there local customizations in the adapter that need to be preserved?
- Should we migrate directly (Option 1) or gradually (Option 2)?
- Do we want logging support enabled?
- Are there use cases that would benefit from middleware support?
