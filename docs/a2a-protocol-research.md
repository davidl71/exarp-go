# A2A Protocol Research: Feasibility for exarp-go

**Task:** T-1773431009286563000
**Date:** 2026-03-15
**Status:** Research complete — see also `docs/research/a2a-protocol.md` (earlier spike notes)

---

## What is A2A?

Google's **Agent-to-Agent (A2A)** protocol is an open standard for autonomous agents to discover, communicate with, and delegate work to each other across vendor boundaries. It uses JSON-RPC 2.0 over HTTP and is designed to be complementary to MCP: MCP connects agents to tools and data sources; A2A connects agents to other agents.

- Spec: https://google.github.io/A2A
- Current version: 0.2 (early 2025; still evolving)
- License: Apache 2.0

### How A2A differs from MCP

| Dimension | MCP | A2A |
|-----------|-----|-----|
| Purpose | Expose tools and resources to an LLM client | Enable agents to delegate tasks to other agents |
| Transport | stdio, SSE, Streamable HTTP | HTTP (JSON-RPC) |
| Discovery | Not standardized (client knows the server) | `/.well-known/agent.json` agent card |
| Primary primitive | Tool call + resource read | Task (submitted → working → completed) |
| Streaming | SSE or Streamable HTTP | SSE per-task event stream |
| Auth | None specified (bearer in practice) | Optional OAuth2 / bearer token |
| Initiator | Human or LLM agent | Another autonomous agent |
| Client support (2026-03) | Wide (Cursor, Claude, OpenCode, Zed) | Narrow (Google ADK; experimental elsewhere) |

MCP and A2A are additive: an MCP server can also expose an A2A endpoint on the same HTTP port. They serve different callers.

---

## A2A Protocol Mechanics

### Agent Card

Every A2A server exposes a self-describing capability manifest at:

```
GET /.well-known/agent.json
```

The card lists the agent's name, URL, version, declared capabilities (streaming, push notifications), and skills. Skills are analogous to MCP tools.

### Task Lifecycle

```
Client                         A2A Server
  |                                 |
  |-- POST /tasks ─────────────────>|  submit task
  |<─ 202 {taskId, status:"submitted"} -|
  |                                 |
  |-- GET /tasks/{id} ─────────────>|  poll status
  |<─ {status: "working", ...} -----|
  |                                 |
  |-- GET /tasks/{id}/events ──────>|  SSE stream (optional)
  |<══════════════════════════════ |  progress events
  |<─ {status: "completed", result} |
```

A2A task states: `submitted` → `working` → `completed` | `failed` | `canceled`

### Streaming

Progress is delivered via Server-Sent Events (SSE) on `GET /tasks/{id}/events`. Each event carries a partial result or status update. Streaming is optional; clients may poll instead.

### Auth

Authentication is optional in the spec. The recommended pattern is bearer token (`Authorization: Bearer <key>`) validated by middleware. OAuth2 flows are specified for production use.

---

## What exarp-go Already Has

| Capability needed for A2A | exarp-go today |
|--------------------------|----------------|
| HTTP server | `internal/api.NewServer` (used with `-serve`) + MCP HTTP (`-mcp-http`) |
| Task storage and lifecycle | Todo2 (SQLite-backed), states: `Todo → In Progress → Review → Done` |
| Tool registry | 37 registered MCP tools; all have names and descriptions |
| Async task execution | `task run-with-ai`, `task_execute` tool |
| SSE precedent | MCP Streamable HTTP transport (`runMCPHTTPMode`) |
| Auth infrastructure | None today |
| `/.well-known/agent.json` | Not implemented |

The existing ACP (`-acp`) implementation in `internal/acp/server.go` is a useful structural analogy: it wraps the MCP server in a different protocol transport using the same tool infrastructure. An A2A layer would follow the same pattern but over HTTP rather than stdio.

---

## Agent Card for exarp-go

A representative agent card reflecting the current tool set:

```json
{
  "name": "exarp-go",
  "description": "MCP server and task orchestrator with local AI (Ollama, Apple FM, MLX)",
  "url": "http://localhost:8080",
  "version": "0.48.0",
  "capabilities": {
    "streaming": true,
    "pushNotifications": false
  },
  "defaultInputModes": ["text"],
  "defaultOutputModes": ["text", "data"],
  "skills": [
    {
      "id": "task_workflow",
      "name": "Task Workflow",
      "description": "Create, list, update, and execute Todo2 tasks",
      "inputModes": ["text"],
      "outputModes": ["text", "data"]
    },
    {
      "id": "text_generate",
      "name": "Text Generation",
      "description": "Generate text via local AI (Ollama, Apple FM, MLX)",
      "inputModes": ["text"],
      "outputModes": ["text"]
    },
    {
      "id": "session",
      "name": "Session Management",
      "description": "Prime session context, emit handoffs",
      "inputModes": ["text"],
      "outputModes": ["text", "data"]
    },
    {
      "id": "report",
      "name": "Reports",
      "description": "Overview, scorecard, and briefing reports",
      "inputModes": ["text"],
      "outputModes": ["text", "data"]
    }
  ]
}
```

The full 37-tool list can be auto-generated from the tool registry at startup rather than hardcoded. Skills map 1-to-1 to tool registrations.

---

## Implementation Complexity

### Component-by-component estimate

| Component | Complexity | Rationale |
|-----------|-----------|-----------|
| `/.well-known/agent.json` endpoint | **Easy** | Static JSON; add one route to existing HTTP mux in `runServeMode` |
| Agent card auto-generation | **Easy** | Iterate tool registry, map Name+Description to Skill structs |
| `POST /tasks` handler | **Medium** | Parse A2A JSON-RPC envelope, create Todo2 task, return task ID |
| `GET /tasks/{id}` status poll | **Easy** | Read Todo2 task, map status enum to A2A states |
| `GET /tasks/{id}/events` SSE | **Medium** | Need per-task channel/goroutine for progress events; no fan-out bus today |
| Task execution bridge | **Medium** | Connect A2A task description to `task run-with-ai`; free-text LLM routing is straightforward |
| Bearer token auth middleware | **Easy** | Single middleware function; key from env/config |
| OAuth2 auth | **Hard** | Full OAuth2 flow; not needed for initial implementation |
| Skill-addressed routing (v1) | **Hard** | Structured param extraction from A2A Message → specific tool handler; requires real client to validate |
| New `-a2a` server mode in `main.go` | **Easy** | Same pattern as existing `-acp` and `-mcp-http` mode dispatch |
| Tests | **Medium** | HTTP handler table tests; SSE harder to test deterministically |

**Overall: Medium.** Phase 0 (agent card) is easy. Phase 1 (task round-trip) is achievable in 2–3 days. Phase 2 (streaming execution) adds complexity but is bounded.

---

## Coexistence with MCP and ACP

exarp-go currently supports three transport modes; A2A would be a fourth:

| Mode flag | Protocol | Transport | Caller |
|-----------|----------|-----------|--------|
| (default) | MCP | stdio | Claude, Cursor, OpenCode |
| `-mcp-http` | MCP Streamable HTTP | HTTP | Remote MCP clients |
| `-acp` | Agent Client Protocol | stdio | Zed, JetBrains |
| `-serve` | REST API + PWA | HTTP | Browser, curl |
| `-a2a` (proposed) | A2A | HTTP | Other agents |

A2A does **not** conflict with MCP or ACP because:

1. It runs on a separate port (or a sub-path of `-serve`).
2. It uses the same underlying tool infrastructure (registries, `framework.MCPServer.CallTool`).
3. The `-a2a` flag would follow the same dispatch pattern in `main.go` as `-acp` and `-mcp-http`.

The cleanest initial approach is to mount A2A routes alongside the existing API in `runServeMode` (i.e. add `/a2a/` path prefix and `/.well-known/agent.json` to the existing `api.Server` mux). This avoids a new port and makes the agent card reachable from the same base URL as the REST API.

**Proposed package layout:**

```
internal/a2a/
  agent_card.go      # AgentCard struct; Generate(tools) → JSON
  handler.go         # http.Handler: routes for A2A endpoints
  task_handler.go    # POST /tasks, GET /tasks/{id}
  events_handler.go  # GET /tasks/{id}/events SSE
  executor.go        # bridge: A2A task → tool dispatch
  auth.go            # BearerAuth middleware
  types.go           # A2A wire types (Task, Artifact, Message)
```

---

## Recommended Next Steps

### If worth pursuing (recommendation: deferred)

A2A client support in the tools exarp-go users actually use (Claude Code, Cursor, OpenCode) is absent as of 2026-03. Google ADK agents are the only production-ready A2A clients. Building A2A infrastructure now carries spec-churn risk and has no testable client.

**Recommended decision:** Build Phase 0 only when a concrete A2A client exists to test against. Defer Phase 1+ until at least one of Claude/Cursor/OpenCode ships A2A support or an internal multi-agent use case (e.g. exarp-go delegating to another exarp-go instance) justifies it.

### Phase roadmap (when to build)

**Phase 0 — Agent card spike (1 day)**
- Add `/.well-known/agent.json` to the `-serve` mux
- Auto-generate from tool registry
- Validate with `curl` or `a2a-cli`
- No task execution; zero risk

**Phase 1 — Minimal task round-trip (2 days)**
- `POST /a2a/tasks` → create Todo2 task → return task ID + `submitted`
- `GET /a2a/tasks/{id}` → map Todo2 status to A2A state
- Bearer token auth middleware (env var `A2A_API_KEY`)

**Phase 2 — Execution + streaming (2 days)**
- Connect `POST /a2a/tasks` to `task run-with-ai` (free-text LLM routing)
- `GET /a2a/tasks/{id}/events` SSE — stream execution progress as A2A events
- Validate end-to-end with Google ADK or test script

**Phase 3 — Skill-addressed routing (future, requires real client)**
- Map A2A `skillId` → specific tool handler
- Structured param extraction from A2A `Message`

### Immediate action

No code changes needed now. When a real A2A client becomes available:

1. Create `internal/a2a/agent_card.go`
2. Mount `/.well-known/agent.json` in `runServeMode` in `cmd/server/main.go`
3. Validate the card is parseable by the client

---

## Risk Summary

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| A2A spec changes (currently v0.2) | High | Thin wire-type layer; isolate in `internal/a2a/types.go` |
| No real clients to test against | High | Build Phase 0 only; do not invest in Phase 1+ until clients exist |
| Auth surface (new HTTP task execution endpoint) | Medium | Bearer token minimum; rate limit; scope to localhost by default |
| Task execution ambiguity (free-text → tool routing) | Medium | LLM routing via `text_generate`; acceptable for v0 |

---

## References

- A2A spec: https://google.github.io/A2A
- A2A GitHub: https://github.com/google/A2A
- Existing exarp-go research note: `docs/research/a2a-protocol.md`
- ACP integration doc: `docs/ACP_INTEGRATION.md`
- ACP implementation: `internal/acp/server.go`
- HTTP server entry points: `cmd/server/main.go` (`runServeMode`, `runMCPHTTPMode`)
