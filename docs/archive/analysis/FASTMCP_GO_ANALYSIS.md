# FastMCP Go SDK Analysis

**Date:** 2026-01-07  
**Reference:** [FastMCP Go SDK](https://pkg.go.dev/github.com/SetiabudiResearch/mcp-go-sdk/pkg/mcp/fastmcp)

---

## Overview

**Package:** `github.com/SetiabudiResearch/mcp-go-sdk/pkg/mcp/fastmcp`  
**Purpose:** FastMCP-style simplified interface for Go MCP servers  
**Status:** Third-party, MIT licensed

---

## Key Features

### 1. Simplified Fluent API

```go
app := fastmcp.New("My App").
    Tool("greet", func(name string) string {
        return "Hello, " + name + "!"
    }, "Greet a person").
    Resource("files/{path}", func(path string) ([]byte, error) {
        return ioutil.ReadFile(path)
    }, "Access files").
    RunStdio()
```

**Benefits:**
- ✅ Chainable methods
- ✅ Less boilerplate
- ✅ Similar to Python FastMCP
- ✅ Handler function type inference

### 2. Multiple Transport Support

```go
app.RunStdio()        // STDIO (for CLI)
app.RunWebSocket(":8080")  // WebSocket
app.RunSSE(":8080")   // Server-Sent Events
```

**Our Current:** Only STDIO (via official SDK)

### 3. Automatic Capabilities

FastMCP automatically configures:
- Tool list change notifications
- Resource subscriptions
- Prompt list change notifications
- Logging support

**Our Current:** Manual capability configuration

### 4. Handler Type Inference

Supports various function signatures:
```go
// Simple
func(string) string
func(int) (int, error)

// Complex
func(User) (*User, error)
func(SearchParams) ([]Result, error)
```

**Our Current:** Explicit JSON schema definition

---

## Comparison: FastMCP Go vs Our Implementation

### Code Verbosity

#### FastMCP Go Style:
```go
app.Tool("lint", func(path string) string {
    // handler logic
    return result
}, "Lint code")
```

#### Our Current Style:
```go
// In registry.go
if err := server.RegisterTool(
    "lint",
    "[HINT: Linting tool...]",
    framework.ToolSchema{
        Type: "object",
        Properties: map[string]interface{}{
            "path": map[string]interface{}{
                "type": "string",
            },
        },
    },
    handleLint,
); err != nil {
    return fmt.Errorf("failed to register lint: %w", err)
}

// In handlers.go
func handleLint(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
    var params struct {
        Path string `json:"path"`
    }
    if err := json.Unmarshal(args, &params); err != nil {
        return nil, fmt.Errorf("failed to parse arguments: %w", err)
    }
    // handler logic
    return []framework.TextContent{{Type: "text", Text: result}}, nil
}
```

**Verdict:** FastMCP Go is significantly less verbose

---

## Should We Consider FastMCP Go?

### Advantages

1. ✅ **Less Boilerplate** - Much simpler tool registration
2. ✅ **Fluent API** - Chainable, readable code
3. ✅ **Type Inference** - Automatic schema generation
4. ✅ **Multiple Transports** - WebSocket, SSE support
5. ✅ **Familiar** - Similar to Python FastMCP

### Disadvantages

1. ⚠️ **Third-Party** - Not official SDK
2. ⚠️ **Less Control** - Automatic schema generation may limit flexibility
3. ⚠️ **Migration Effort** - Would need to rewrite all 24 tools
4. ⚠️ **Framework Lock-in** - Less framework-agnostic
5. ⚠️ **Unknown Maintenance** - Repository activity unclear

### Our Current Advantages

1. ✅ **Official SDK** - Guaranteed spec compliance
2. ✅ **Framework-Agnostic** - Can switch frameworks easily
3. ✅ **Full Control** - Explicit schema definition
4. ✅ **Already Working** - 24 tools, 15 prompts, 6 resources
5. ✅ **Proven** - Production-ready implementation

---

## Migration Complexity

### If We Switched to FastMCP Go

**Required Changes:**
1. Replace framework adapter
2. Rewrite all 24 tool registrations
3. Rewrite all 15 prompt registrations
4. Rewrite all 6 resource registrations
5. Update handler signatures
6. Test all functionality

**Estimated Effort:** High (days/weeks of work)

**Risk:** Medium-High (breaking changes, unknown framework maturity)

---

## Recommendation

### ❌ **Do NOT migrate to FastMCP Go**

**Reasons:**

1. **We're Already Successful** - 24 tools, 15 prompts, 6 resources all working
2. **Official SDK is Better** - Long-term support and spec compliance
3. **Framework-Agnostic Design** - Our abstraction is more flexible
4. **Migration Cost** - High effort for minimal benefit
5. **Unknown Quality** - Third-party, unclear maintenance status

### ✅ **Alternative: Create FastMCP-Style Wrapper**

**Better Approach:**
- Keep official SDK as backend
- Create FastMCP-style convenience wrapper
- Best of both worlds: simplicity + official support

**Example:**
```go
// internal/framework/fastmcp.go
type FastMCP struct {
    server framework.MCPServer
}

func (f *FastMCP) Tool(name string, handler interface{}, desc string) *FastMCP {
    // Auto-generate schema from handler signature
    // Register with official SDK backend
    return f
}
```

---

## Current Status Summary

**Our Implementation:**
- ✅ Official SDK (best long-term support)
- ✅ Framework-agnostic design (flexibility)
- ✅ 24 tools working
- ✅ 15 prompts working
- ✅ 6 resources working
- ✅ Production-ready

**FastMCP Go:**
- ⚠️ Third-party (unknown maintenance)
- ⚠️ Less flexible (automatic schema)
- ✅ Simpler API (less boilerplate)
- ✅ Multiple transports (we only need STDIO)

**Verdict:** Stay with current implementation ✅

---

## References

- [FastMCP Go SDK Documentation](https://pkg.go.dev/github.com/SetiabudiResearch/mcp-go-sdk/pkg/mcp/fastmcp)
- [Our Framework Comparison](docs/MCP_FRAMEWORKS_COMPARISON.md)
- [Our Framework-Agnostic Design](docs/FRAMEWORK_AGNOSTIC_DESIGN.md)

