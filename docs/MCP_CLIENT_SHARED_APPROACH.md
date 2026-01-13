# MCP Client - Shared Library Approach

**Date:** 2026-01-13  
**Context:** Both `exarp-go` and `devwisdom-go` use `mcp-go-core` shared library  
**Opportunity:** Add MCP client functionality to `mcp-go-core` for shared use

---

## Current Situation

### Projects Using mcp-go-core

1. **exarp-go** ✅
   - Uses `mcp-go-core` for framework abstraction, security, types
   - Needs MCP client to call devwisdom-go server
   - 3 actions need client: `recommend`/advisor, `report`/briefing, `memory_maint`/dream

2. **devwisdom-go** ✅
   - Uses `mcp-go-core` for framework abstraction, security, types
   - Could potentially use MCP client to call other servers (future)

3. **mcp-go-core** 📦
   - Shared library with framework abstraction, security utilities, types
   - **Currently server-only** - no client functionality
   - Opportunity to add client support for both projects

---

## Proposed Solution: Add MCP Client to mcp-go-core

### Benefits

1. **Shared Implementation** ✅
   - Both projects benefit from same client code
   - Single implementation to maintain
   - Consistent behavior across projects

2. **Reusability** ✅
   - Any project using `mcp-go-core` can use client functionality
   - Future projects can leverage it
   - Standard pattern for MCP client calls

3. **Consistency** ✅
   - Same patterns as server implementation
   - Uses same types and utilities
   - Framework-agnostic design

### Implementation Approach

#### Option 1: Add Client Package to mcp-go-core

**Structure:**
```
mcp-go-core/pkg/mcp/
  client/
    client.go          # MCP client implementation
    stdio_client.go    # STDIO transport client
    types.go           # Client-specific types
```

**API Design:**
```go
package client

type Client interface {
    Connect(ctx context.Context) error
    CallTool(ctx context.Context, toolName string, args map[string]interface{}) ([]TextContent, error)
    ListTools(ctx context.Context) ([]ToolInfo, error)
    Close() error
}

// NewStdioClient creates a new STDIO-based MCP client
func NewStdioClient(command string, args []string) Client {
    // Implementation
}
```

**Usage in exarp-go:**
```go
import "github.com/davidl71/mcp-go-core/pkg/mcp/client"

// Create client for devwisdom-go
devwisdomClient := client.NewStdioClient(
    "/path/to/devwisdom",
    []string{},
)

// Connect and call tool
err := devwisdomClient.Connect(ctx)
if err != nil {
    return err
}

result, err := devwisdomClient.CallTool(ctx, "consult_advisor", map[string]interface{}{
    "metric": "security",
    "score": 40.0,
    "context": "Working on improving project security",
})
```

#### Option 2: Reuse Existing Server Infrastructure

**Leverage existing types:**
- Use `mcp-go-core/pkg/mcp/types` for shared types
- Use `mcp-go-core/pkg/mcp/framework` patterns
- Reuse JSON-RPC 2.0 handling from framework

**Key Insight:**
- Client is "reverse" of server (initiate vs receive requests)
- Same protocol (JSON-RPC 2.0)
- Same transport (STDIO, HTTP, SSE)
- Similar patterns, different direction

---

## Implementation Plan

### Phase 1: Research and Design

1. **Research MCP Client Protocol**
   - Review MCP specification for client requirements
   - Understand JSON-RPC 2.0 client patterns
   - Study existing client implementations (Python, other languages)

2. **Design Client API**
   - Define client interface
   - Design connection lifecycle
   - Plan error handling
   - Consider transport options (STDIO first)

3. **Review mcp-go-core Structure**
   - Understand current architecture
   - Identify where client package fits
   - Ensure consistency with existing patterns

### Phase 2: Implementation

1. **Implement Core Client** (`mcp-go-core/pkg/mcp/client`)
   - STDIO transport client
   - JSON-RPC 2.0 message handling
   - Connection management
   - Tool calling interface

2. **Add Client Types**
   - Reuse existing types from `mcp-go-core/pkg/mcp/types`
   - Add client-specific types if needed
   - Ensure compatibility

3. **Testing**
   - Unit tests for client functionality
   - Integration tests with devwisdom-go server
   - Test error handling and edge cases

### Phase 3: Integration

1. **Update exarp-go**
   - Use client from `mcp-go-core`
   - Replace Python bridge calls with native Go client
   - Update handlers: `recommend`, `report`, `memory_maint`

2. **Documentation**
   - Document client API
   - Add usage examples
   - Update migration guide

3. **Future Use**
   - devwisdom-go can use client for future features
   - Other projects can leverage it
   - Community benefit

---

## Comparison with Current Approach

### Current: Python Bridge

**Pros:**
- ✅ Works now
- ✅ No implementation needed
- ✅ Python has mature MCP client libraries

**Cons:**
- ❌ Requires Python bridge
- ❌ Adds Python dependency
- ❌ Not fully native Go
- ❌ Inconsistent with migration goals

### Proposed: mcp-go-core Client

**Pros:**
- ✅ Fully native Go
- ✅ Shared implementation (both projects benefit)
- ✅ Consistent with existing patterns
- ✅ Framework-agnostic design
- ✅ Reusable for future projects
- ✅ Single binary deployment
- ✅ Better performance

**Cons:**
- ⚠️ Requires implementation in mcp-go-core
- ⚠️ More initial work
- ⚠️ Need to maintain client code

---

## Next Steps

### Short-term (Research)

1. ✅ Understand current mcp-go-core structure
2. ✅ Review MCP client protocol requirements
3. ✅ Design client API
4. ⏳ Create implementation plan

### Medium-term (Implementation)

1. ⏳ Implement client in mcp-go-core
2. ⏳ Add tests and documentation
3. ⏳ Release new mcp-go-core version

### Long-term (Integration)

1. ⏳ Update exarp-go to use client
2. ⏳ Replace Python bridge calls
3. ⏳ Update devwisdom-go to use client (if needed)

---

## Considerations

### Protocol Support

- **STDIO**: Primary transport (simplest, works well)
- **HTTP**: Future enhancement (if needed)
- **SSE**: Future enhancement (if needed)

### Error Handling

- Connection errors
- Protocol errors
- Tool call errors
- Timeout handling

### Lifecycle Management

- Connection initialization
- Session management
- Graceful shutdown
- Resource cleanup

---

## Conclusion

**Adding MCP client to mcp-go-core is the right approach** because:

1. ✅ **Shared benefit** - Both exarp-go and devwisdom-go can use it
2. ✅ **Consistent patterns** - Same library, same patterns
3. ✅ **Future-proof** - Other projects can benefit
4. ✅ **Native Go** - Fully native implementation
5. ✅ **Better architecture** - Client and server in same ecosystem

**Recommendation:** Implement MCP client in `mcp-go-core` as a shared package, then use it in `exarp-go` to replace Python bridge calls.

---

## References

- **mcp-go-core:** `/home/dlowes/projects/mcp-go-core`
- **exarp-go:** `/home/dlowes/projects/exarp-go`
- **devwisdom-go:** `/home/dlowes/projects/devwisdom-go`
- **MCP Specification:** https://modelcontextprotocol.io
- **Current Solution:** `docs/MCP_CLIENT_SOLUTION.md`
