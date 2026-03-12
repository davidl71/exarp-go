# JSON Handling Guidelines

**Date:** 2026-01-13  
**Status:** Active Guidelines

---

## Principle: Use Standard Library Directly

**We use Go's standard `encoding/json` package directly - no wrappers or utilities.**

### Rationale

1. **Standard library is sufficient** - No performance requirements justify wrappers
2. **No unnecessary indirection** - Direct calls are clearer and more maintainable
3. **Go best practices** - Handle errors inline, don't hide complexity
4. **Avoid reinventing the wheel** - Standard library is well-tested and optimized

---

## Usage Patterns

### Compact JSON (Machine-to-Machine Transfer)

Use `json.Marshal()` for compact JSON when:
- Exporting data for sharing between agents/machines
- Network transfer (smaller payloads, ~20-30% size reduction)
- Performance-critical paths
- Internal data serialization

```go
import "encoding/json"

data, err := json.Marshal(result)
if err != nil {
    return nil, fmt.Errorf("failed to marshal: %w", err)
}
```

### Formatted JSON (Human-Readable Output)

Use `json.MarshalIndent()` for formatted JSON when:
- Logging or debugging output
- Configuration files
- Human-readable responses
- Development/testing

```go
import "encoding/json"

data, err := json.MarshalIndent(result, "", "  ")
if err != nil {
    return nil, fmt.Errorf("failed to marshal: %w", err)
}
```

### Error Handling

**Always handle errors explicitly:**

```go
data, err := json.Marshal(result)
if err != nil {
    // Handle error appropriately:
    // - Return error to caller
    // - Log error
    // - Use fallback value (only if safe)
    return nil, fmt.Errorf("marshal failed: %w", err)
}
```

**For error-safe marshaling (only when absolutely necessary):**

```go
data, err := json.Marshal(result)
if err != nil {
    // Return empty object only if protocol requires it
    // (e.g., embedding JSON in resource responses)
    data = []byte("{}")
}
```

---

## When NOT to Create Wrappers

❌ **Don't create wrappers for:**
- `json.Marshal()` → `MarshalCompact()` (no value added)
- `json.MarshalIndent()` → `MarshalFormatted()` (no value added)
- Simple error handling → `MustMarshal*()` (handle errors inline)

✅ **Only create utilities if:**
- You need to swap in jsoniter/sonic for performance (add abstraction then)
- You have complex custom marshaling logic (domain-specific)
- You need shared error handling patterns (document the pattern, not wrap)

---

## Performance Considerations

### Standard Library Performance

- **`json.Marshal()`**: Fast enough for most use cases
- **`json.MarshalIndent()`**: Slightly slower due to formatting (~10-20% overhead)
- **Memory**: Efficient, minimal allocations

### When to Consider Alternatives

Only consider faster libraries (jsoniter, sonic) if:
- JSON operations are in hot paths (profiled and confirmed bottleneck)
- Processing very large JSON documents (>10MB)
- High-throughput scenarios (>100k ops/sec)

**Current projects (exarp-go, devwisdom-go):**
- Standard library is perfectly adequate
- No performance bottlenecks identified
- No need for jsoniter/sonic

---

## Examples

### Session Handoff Export (Machine-to-Machine)

```go
// Compact JSON for export files
data, err := json.Marshal(handoffData)
if err != nil {
    return nil, fmt.Errorf("failed to marshal handoff data: %w", err)
}
```

### Tool Response (Human-Readable)

```go
// Formatted JSON for tool responses
output, err := json.MarshalIndent(result, "", "  ")
if err != nil {
    return nil, fmt.Errorf("failed to marshal response: %w", err)
}
return []framework.TextContent{
    {Type: "text", Text: string(output)},
}, nil
```

### Error-Safe Resource Response

```go
// Only use error-safe marshaling when protocol requires it
data, err := json.Marshal(consultations)
if err != nil {
    // Return empty object to avoid breaking protocol
    data = []byte("{}")
}
```

---

## Migration from Wrappers

If you find wrapper functions in the codebase:

1. **Identify usage** - Find all call sites
2. **Replace with standard library** - Use `json.Marshal()` or `json.MarshalIndent()` directly
3. **Handle errors inline** - Don't hide error handling
4. **Remove wrapper** - Delete the unnecessary abstraction

---

## Related Documentation

- [Go JSON Package Documentation](https://pkg.go.dev/encoding/json)
- [Go Best Practices: Error Handling](https://go.dev/doc/effective_go#errors)
- [mcp-go-core README](../README.md)

---

**Remember: Simple, direct, and clear. Use the standard library.**
