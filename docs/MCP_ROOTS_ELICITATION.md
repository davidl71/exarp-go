# MCP Roots and Elicitation in exarp-go

exarp-go supports two MCP client features: **Roots** (workspace boundaries) and **Elicitation** (server-initiated user input). This document describes how they work and how to use them.

**References:**

- [Cursor — Model Context Protocol (MCP)](https://cursor.com/docs/context/mcp)
- [MCP — Roots](https://modelcontextprotocol.io/docs/concepts/roots)
- [MCP — Elicitation](https://modelcontextprotocol.io/docs/concepts/elicitation)

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
