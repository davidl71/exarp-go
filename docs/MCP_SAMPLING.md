# MCP Sampling Implementation

## Overview

MCP Sampling allows the server to request the client to generate text using its LLM. This is useful when the server needs AI analysis but doesn't have its own LLM access.

## Current Status

**Implemented:**
- `framework.Sampler` interface in mcp-go-core
- `SamplerFromContext(ctx)` helper in exarp-go
- Re-exports in `internal/framework/server.go`

**Not yet implemented:**
- gosdk adapter integration (to inject Sampler into context)
- Tool that uses sampling

## Research Findings ✅

**Key discovery:** `CallToolRequest = ServerRequest[*CallToolParamsRaw]` which has:
- `Session *ServerSession` - the MCP session

**Implementation approach:**
1. In gosdk adapter tool handler, capture `req.Session` 
2. Store session in adapter struct
3. Create Sampler that wraps `session.CreateMessage()`
4. Create RootsHandler that wraps `session.ListRoots()`

**Code path:**
```
ToolHandler (gosdk adapter)
  → req.Session (ServerSession)
  → session.CreateMessage(ctx, params)
  → return CreateMessageResult
```

## gosdk Adapter Integration Required

The gosdk adapter needs to:

1. **Detect client capabilities**: Check if client advertises sampling capability during initialization
2. **Create Sampler implementation**: Wrap `ServerSession.CreateMessage` 
3. **Inject into context**: Use middleware or handler wrapper to inject Sampler before tool/resource handlers

**Challenge**: The adapter's tool handlers receive a Go `context.Context`, but the MCP `ServerSession` is needed to call `CreateMessage`. The session must be captured during initialization.

### Implementation Options

**Option A: Middleware approach**
- Add `WithSamplingSupport` adapter option
- Middleware wraps tool handlers and injects Sampler into context

**Option B: Handler wrapper approach** (similar to how TaskStore is injected)
- In native handlers, inject Sampler like TaskStore is injected in task_workflow_native.go

**Option C: Full integration**
- Modify gosdk adapter to capture ServerSession
- Inject Sampler automatically in all handlers

## Usage Pattern

Once fully implemented, tools can use sampling like this:

```go
import "github.com/davidl71/exarp-go/internal/framework"

func myToolHandler(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
    // Check if client supports sampling
    sampler := framework.SamplerFromContext(ctx)
    if sampler == nil {
        // Client doesn't support sampling - degrade gracefully
        return nil, fmt.Errorf("sampling not supported by client")
    }

    // Request sampling from client
    result, err := sampler.CreateMessage(ctx, framework.CreateMessageParams{
        Messages: []framework.SamplingMessage{
            {Role: "user", Content: "Analyze this code and suggest improvements..."},
        },
        Temperature: 0.7,
        MaxTokens:   512,
    })
    if err != nil {
        return nil, fmt.Errorf("sampling failed: %w", err)
    }

    // Use result.Content
    return framework.FormatResult(result.Content, "")
}
```

## Use Cases

Potential tools that could use sampling:
- `ask_agent` - Ask the client to analyze something
- `code_review` - Request client LLM to review code
- `explain_error` - Ask client to explain an error

## References

- [MCP Sampling Protocol](https://modelcontextprotocol.io/docs/concepts/sampling)
- `vendor/github.com/modelcontextprotocol/go-sdk/mcp/server.go` - ServerSession.CreateMessage
- `internal/tools/task_workflow_native.go` - Example of context injection pattern
