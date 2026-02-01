# MCP Roots and Elicitation in exarp-go

exarp-go supports two MCP client features: **Roots** (workspace boundaries) and **Elicitation** (server-initiated user input). This document describes how they work and how to use them.

**References:**

- [Cursor — Model Context Protocol (MCP)](https://cursor.com/docs/context/mcp)
- [MCP — Roots](https://modelcontextprotocol.io/docs/concepts/roots)
- [MCP — Elicitation](https://modelcontextprotocol.io/docs/concepts/elicitation)
- [FastMCP — User Elicitation (server)](https://gofastmcp.com/servers/elicitation) — Python server API (`ctx.elicit()`)
- [FastMCP — fastmcp.server.elicitation](https://gofastmcp.com/python-sdk/fastmcp-server-elicitation) — Schema and response types

---

## Roots

**What it is:** The MCP *client* (e.g. Cursor) exposes workspace roots to the *server*. Roots define the boundaries of where the server can operate (e.g. workspace folders). The client sends `notifications/roots/list_changed` when roots change; the server can call `roots/list` to get the current list.

**How exarp-go uses it:**

- The gosdk adapter creates the MCP server with `ServerOptions{RootsListChangedHandler}` so the server receives roots/list_changed notifications.
- When a **resource** is read, the adapter calls `Session.ListRoots(ctx, nil)` and attaches the result to the request context. Resource handlers can read client roots via **`framework.RootsFromContext(ctx)`** (from `github.com/davidl71/mcp-go-core/pkg/mcp/framework`).
- Handlers that serve path-sensitive or `file://` content can use roots to validate paths and stay within client-allowed boundaries.

**Example (resource handler):**

```go
import mcpframework "github.com/davidl71/mcp-go-core/pkg/mcp/framework"

func myResourceHandler(ctx context.Context, uri string) ([]byte, string, error) {
    roots := mcpframework.RootsFromContext(ctx)
    if len(roots) > 0 {
        // Client provided roots; restrict path resolution to these URIs
        for _, r := range roots {
            // r.URI, r.Name
        }
    }
    // ... serve resource
}
```

**When roots are not set:** If the client does not support roots or has not sent roots, `RootsFromContext(ctx)` returns `nil`. Handlers should treat that as “no roots restriction” and use existing behavior.

---

## Elicitation

**What it is:** The MCP *server* can ask the *client* for user input (form or URL). The client shows a form or URL to the user and returns the result. The client must advertise the `elicitation` capability.

**How exarp-go uses it:**

- The gosdk adapter injects an **Eliciter** into the request context for **tool** handlers. Tools can call **`framework.EliciterFromContext(ctx)`** (from `github.com/davidl71/mcp-go-core/pkg/mcp/framework`). If non-nil, the client supports elicitation and the tool can call `ElicitForm(ctx, message, schema)` to ask the user for structured input.
- **Session tool:** When `action=prime` and `ask_preferences=true`, the session tool optionally uses elicitation to ask the user for “include task summary” and “include hints” preferences. If the client does not support elicitation or the user declines, defaults are used (graceful fallback).

**Example (tool handler):**

```go
import mcpframework "github.com/davidl71/mcp-go-core/pkg/mcp/framework"

func myToolHandler(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
    eliciter := mcpframework.EliciterFromContext(ctx)
    if eliciter != nil {
        schema := map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "confirm": map[string]interface{}{"type": "boolean", "description": "Proceed?"},
            },
        }
        action, content, err := eliciter.ElicitForm(ctx, "Confirm action?", schema)
        if err == nil && action == "accept" && content != nil {
            // Use content["confirm"], etc.
        }
        // On decline/cancel/error: fall back to default behavior
    }
    // ...
}
```

**When elicitation is not available:** In CLI mode or when the client does not support elicitation, `EliciterFromContext(ctx)` returns `nil`. Tools should not call it in that case; they should use defaults or skip the prompt.

**Timeouts:** All elicitation calls use a bounded context (`context.WithTimeout`) so they never block indefinitely. Session prime uses 5 seconds; task_workflow confirm (approve/delete) uses 15 seconds. On timeout, the tool falls back (e.g. defaults for prime, cancelled result for approve/delete) and may report "elicitation timed out" in the message.

### FastMCP reference (why it used to show in Cursor)

**What is “Context”?** In FastMCP, **Context** is the object the framework passes into each tool (e.g. `async def my_tool(ctx: Context)`). It represents the current request/execution environment and exposes methods like `ctx.elicit(message, response_type=...)` for user input. Context is only available when the client connects in a way that provides it (e.g. Cursor using FastMCP integration); in plain stdio MCP there is no FastMCP Context — the server still gets a Go `context.Context` and, when the client advertises elicitation, an **Eliciter** via `EliciterFromContext(ctx)`.

Elicitation **used to work in Cursor** when the project used a **FastMCP (Python) server**. FastMCP uses the same MCP elicitation protocol under the hood: `ctx.elicit(message, response_type=...)` sends an MCP `elicitation/create` request with a JSON schema. The difference is **how the client surfaces it**:

- **FastMCP mode (Python server):** When Cursor connected to a FastMCP server, it had “FastMCP Context” and could show elicitation **inline in chat** — the question appeared as part of the conversation and the user replied in the same thread. See [DEMONSTRATE_ELICIT_EXAMPLE.md](DEMONSTRATE_ELICIT_EXAMPLE.md) (“This question appeared inline in chat, not as a pop-up”).
- **Stdio MCP (exarp-go):** exarp-go sends the same MCP `elicitation/create` (form mode) via the official Go SDK. Cursor may advertise the elicitation capability (so we get an Eliciter and the request is sent), but **Cursor’s stdio MCP path may not yet render form elicitation in the UI**. So the server asks, but the user never sees a prompt; after our timeout we fall back to defaults and set `elicitation: "timeout"` in the result.

So: same protocol, different client behavior. When Cursor adds (or exposes) UI for MCP form elicitation in stdio mode, `ask_preferences=true` will show the prompt without code changes in exarp-go.

---

## Elicitation support plan

**Current state:** Session prime is the primary elicitation tool: when `ask_preferences=true` and the client supports elicitation, it asks for include_tasks / include_hints. Scope is split: **mcp-go-core** defines the `Eliciter` interface and injects it in the gosdk adapter; **exarp-go** uses `EliciterFromContext(ctx)` in tool handlers and decides when to elicit.

**Pattern:** Always check `EliciterFromContext(ctx) != nil` before calling `ElicitForm`. On decline, cancel, or error, use defaults or skip the prompt (graceful fallback). Do not block CLI or non-elicitation clients.

**Optional second use:** The task_workflow tool supports optional confirmation for batch actions (e.g. approve, delete) via `confirm_via_elicitation`: when true and the client supports elicitation, the user is prompted to confirm (and optionally set dry_run) before the action runs.

---

## Summary

| Feature     | Direction   | Where it’s used in exarp-go | How to access in code                          |
|------------|------------|-----------------------------|-----------------------------------------------|
| **Roots**  | Client → server | Resource handlers          | `framework.RootsFromContext(ctx)`             |
| **Elicitation** | Server → client | Tool handlers        | `framework.EliciterFromContext(ctx)` → `ElicitForm(...)` |

Both features are optional: handlers must check for `nil` and fall back when the client does not support them.

**Elicitation pattern (tool handler):** Check `EliciterFromContext(ctx) != nil`, then call `ElicitForm(ctx, message, schema)`; on decline, cancel, or error use defaults or skip the prompt.
