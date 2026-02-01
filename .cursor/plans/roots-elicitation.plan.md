---
name: MCP Roots and Elicitation
overview: Add MCP Roots (client workspace boundaries) and Elicitation (server-initiated user input) support to exarp-go.
todos:
  - id: T-1769981328913
    content: Add MCP Roots support (adapter + handlers)
    status: done
  - id: T-1769981329793
    content: Add MCP Elicitation support (adapter + tools)
    status: done
  - id: T-1769981330518
    content: Document MCP Roots and Elicitation
    status: done
isProject: false
---

# Plan: MCP Roots and Elicitation

**Tag hints:** `#mcp` `#feature` `#docs`

**Generated:** 2026-02-01

---

## 1. Purpose & Success Criteria

**Purpose:** Enable exarp-go to use two MCP client features that Cursor supports ([MCP docs](https://cursor.com/docs/context/mcp)): **Roots** (workspace boundaries the client exposes to the server) and **Elicitation** (server-initiated requests for user input via forms or URLs).

**Success criteria:**

- **Roots:** When Cursor sends workspace roots, exarp-go receives them (e.g. via `RootsListChangedHandler`), and resource handlers that serve path-sensitive or `file://` content can call `ListRoots` and respect client boundaries.
- **Elicitation:** Tool handlers can request user input (form or URL) via the session; at least one tool optionally uses elicitation when the client supports it; unsupported clients are handled without errors.
- Documentation describes both features and how exarp-go uses them.

---

## 2. Technical Foundation

- **MCP:** [Model Context Protocol](https://modelcontextprotocol.io) (vendored go-sdk already implements Roots and Elicitation).
- **exarp-go:** Framework-agnostic `MCPServer` interface; gosdk adapter in `mcp-go-core` wraps `mcp.Server`; tools/resources registered in `internal/tools` and resources.
- **Roots (client → server):** Client sends `notifications/roots/list_changed`; server can call `Session.ListRoots(ctx, nil)`. SDK uses roots in built-in `fileResourceHandler` for path checks.
- **Elicitation (server → client):** Server calls `Session.Elicit(ctx, params)`; client shows form or URL and returns result. Client must advertise `elicitation` capability.
- **Invariants:** Use Makefile for build/test; no direct edits to `.todo2/state.todo2.json`; prefer framework adapter changes in mcp-go-core (or local adapter) so exarp-go stays thin.

---

## 3. Iterative Milestones

Each milestone maps to a Todo2 task. Check off as done.

- [x] **Roots: adapter + handlers** (T-1769981328913)  
  - Create server with `ServerOptions{RootsListChangedHandler: ...}` in gosdk adapter.  
  - Cache or expose roots when client sends `roots/list_changed`.  
  - Extend framework so resource handlers can access roots/session (context or new param).  
  - Use `ListRoots` in any path-sensitive resource logic.
- [x] **Elicitation: adapter + tools** (T-1769981329793)  
  - Pass session (or Elicit wrapper) into tool handlers (e.g. via context or extended handler signature).  
  - In gosdk adapter, get `req.Session` from `CallToolRequest` and inject into handler.  
  - Add optional elicitation in at least one tool (e.g. session prime, task_workflow, report).  
  - Graceful fallback when client does not support elicitation.
- [x] **Documentation** (T-1769981330518)  
  - Document Roots and Elicitation in `docs/MCP_ROOTS_ELICITATION.md` or `docs/CURSOR_MCP_SETUP.md`.  
  - Describe when exarp-go uses roots and when tools use elicitation.  
  - Link to Cursor MCP docs.

---

## 4. Open Questions

- Prefer extending **mcp-go-core** (shared library) vs. exarp-go-only adapter changes for Roots/Elicitation?
- Which tool(s) should use elicitation first (session prime, report, task_workflow, or a new “clarify” helper)?
- Do we need any `file://` resources that must respect roots, or is Roots support primarily for future path-safe resources?

---

## 5. References

- [Cursor — Model Context Protocol (MCP)](https://cursor.com/docs/context/mcp)
- [MCP — Elicitation](https://modelcontextprotocol.io/docs/concepts/elicitation)
- [MCP — Roots](https://modelcontextprotocol.io/docs/concepts/roots)
- Vendored SDK: `vendor/github.com/modelcontextprotocol/go-sdk/mcp/server.go` (`RootsListChangedHandler`, `ServerSession.ListRoots`, `ServerSession.Elicit`)

