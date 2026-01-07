# MCP Go Framework Comparison: Official SDK vs mcp-go

**Date:** 2026-01-07  
**Reference:** [Building a MCP Server in Go](https://en.bioerrorlog.work/entry/hello-mcp-golang)

---

## Current Implementation

**We're using:** `github.com/modelcontextprotocol/go-sdk v1.2.0` (Official SDK)

**Blog post discusses:** `github.com/mark3labs/mcp-go v0.18.0` (Third-party SDK)

---

## Key Differences

### 1. Official Status

| Framework | Status | Release Timeline |
|-----------|--------|------------------|
| **go-sdk** (Official) | ✅ Official | Released after April 2025 |
| **mcp-go** (mark3labs) | ⚠️ Third-party | Available since before April 2025 |

**Our Choice:** Official SDK - Better long-term support and spec compliance

---

### 2. API Style Comparison

#### mcp-go (mark3labs) - Options Pattern

```go
// From blog post
addTool := mcp.NewTool(
    "add",
    mcp.WithDescription("Add two numbers"),
    mcp.WithNumber("x", mcp.Required()),
    mcp.WithNumber("y", mcp.Required()),
)
s.AddTool(addTool, addToolHandler)
```

**Characteristics:**
- Uses Options Pattern (functional options)
- More verbose but explicit
- Type-safe with helper functions (`WithNumber`, `WithString`, etc.)

#### go-sdk (Official) - JSON Schema

```go
// Our implementation
inputSchemaMap := map[string]interface{}{
    "type":       "object",
    "properties": map[string]interface{}{
        "x": map[string]interface{}{
            "type": "number",
        },
        "y": map[string]interface{}{
            "type": "number",
        },
    },
    "required": []string{"x", "y"},
}
```

**Characteristics:**
- Uses JSON Schema directly
- More flexible (can define complex schemas)
- Matches MCP protocol specification exactly

---

### 3. Code Verbosity

**mcp-go (from blog):**
- More verbose per tool/resource/prompt
- Explicit type helpers
- Clear intent but more lines of code

**go-sdk (our implementation):**
- Less verbose for simple cases
- More flexible for complex schemas
- Framework-agnostic abstraction layer reduces boilerplate

---

## Why We Chose Official SDK

### Advantages

1. ✅ **Official Support** - Maintained by MCP team
2. ✅ **Spec Compliance** - Guaranteed to match MCP specification
3. ✅ **Future-Proof** - Will receive updates with spec changes
4. ✅ **Better Documentation** - Official documentation and examples
5. ✅ **Framework-Agnostic Design** - Our abstraction allows switching frameworks

### Trade-offs

1. ⚠️ **Newer** - Less community examples (but growing)
2. ⚠️ **Different API** - JSON Schema approach vs Options Pattern

---

## Framework-Agnostic Design Benefits

Our implementation uses a **framework-agnostic abstraction layer**, which means:

1. ✅ **Easy Framework Switching** - Can switch between go-sdk, mcp-go, go-mcp via config
2. ✅ **Consistent API** - Our tools/prompts/resources use same interface regardless of framework
3. ✅ **Future Flexibility** - Can add mcp-go adapter if needed

**Current Adapters:**
- ✅ `gosdk` - Official SDK (currently used)
- 📋 `mcpgo` - Could add mcp-go adapter if needed
- 📋 `gomcp` - Could add go-mcp adapter if needed

---

## Should We Consider mcp-go?

### Reasons to Consider

1. **Larger Community** - 730 forks, more examples
2. **Options Pattern** - Some developers prefer explicit type helpers
3. **Mature Ecosystem** - More real-world usage examples

### Reasons to Stay with Official SDK

1. ✅ **Official Support** - Better long-term maintenance
2. ✅ **Spec Compliance** - Guaranteed to match MCP spec
3. ✅ **Our Abstraction** - Framework choice is hidden from tool implementations
4. ✅ **Already Working** - 24 tools, 15 prompts, 6 resources all working

---

## Recommendation

**Stay with Official SDK** ✅

**Reasons:**
1. We're already successfully using it
2. Official support is more valuable than community size
3. Our framework-agnostic design means we can add mcp-go later if needed
4. Official SDK will receive updates with spec changes

**If Needed:**
- Can add mcp-go adapter as alternative option
- Framework selection via config file
- No code changes needed for tools/prompts/resources

---

## References

- [Blog Post: Building a MCP Server in Go](https://en.bioerrorlog.work/entry/hello-mcp-golang)
- [Official Go SDK](https://github.com/modelcontextprotocol/go-sdk)
- [mcp-go by mark3labs](https://github.com/mark3labs/mcp-go)
- [Our Framework Comparison](docs/MCP_FRAMEWORKS_COMPARISON.md)

