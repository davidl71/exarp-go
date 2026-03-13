# A2A Protocol Support for exarp-go

**Task:** T-1773431009286563000
**Date:** 2026-03-13
**Status:** Research / Pre-implementation

---

## What is A2A?

Google's Agent-to-Agent (A2A) protocol is an open standard for autonomous agents to discover, communicate with, and delegate tasks to each other across vendor boundaries. It is transport-agnostic (HTTP/JSON-RPC) and complementary to MCP: MCP connects agents to tools and data; A2A connects agents to other agents.

Spec: https://google.github.io/A2A
Current version: 0.2 (early 2025, still evolving)

---

## Protocol Overview

### Agent Card

Every A2A agent exposes a machine-readable capability manifest:

```
GET /.well-known/agent.json
```

```json
{
  "name": "exarp-go",
  "description": "MCP server and task orchestrator",
  "url": "http://localhost:8080",
  "version": "0.48.0",
  "capabilities": {
    "streaming": true,
    "pushNotifications": false
  },
  "skills": [
    {
      "id": "task_workflow",
      "name": "Task Workflow",
      "description": "Create, list, update, and execute Todo2 tasks",
      "inputModes": ["text"],
      "outputModes": ["text", "data"]
    }
  ]
}
```

Skills map naturally to exarp-go's registered MCP tools.

### Task Lifecycle

```
Client                         exarp-go (A2A server)
  |                                    |
  |-- POST /tasks ─────────────────────>|  submit task
  |<─ 202 {taskId, status: submitted} --|
  |                                    |
  |-- GET /tasks/{id} ─────────────────>|  poll status
  |<─ {status: working, ...} -----------|
  |                                    |
  |-- GET /tasks/{id}/events ──────────>|  SSE stream (optional)
  |<═══════════════════════════════════ |  progress events
  |<─ {status: completed, result} ------|
```

A2A task states: `submitted` → `working` → `completed` | `failed` | `canceled`

These map directly to Todo2 statuses: `Todo` → `In Progress` → `Done` | (error) | (deleted)

---

## Fit Assessment for exarp-go

### Strong fits

| A2A requirement | exarp-go today | Gap |
|----------------|---------------|-----|
| HTTP server | `api.NewServer` running on configurable port | None |
| Task lifecycle | Todo2 (`Todo → In Progress → Done`) | Mapping layer needed |
| Capability manifest | 37 tools in tool registry | Auto-generation needed |
| SSE streaming | MCP SSE transport exists as precedent | New endpoint needed |
| Task execution | `task run-with-ai`, `task_execute` tool | Bridge to A2A input needed |

### Gaps

1. **`/.well-known/agent.json` endpoint** — not implemented; straightforward to add
2. **`POST /tasks` handler** — needs JSON-RPC parsing and Todo2 task creation
3. **`GET /tasks/{id}/events` SSE** — needs per-task event bus (goroutine + channel)
4. **Auth middleware** — bearer token validation; no key management today
5. **Task execution bridge** — mapping free-text A2A task descriptions to specific exarp-go tools

---

## Proposed Architecture

```
internal/a2a/
  agent_card.go      # AgentCard struct; Generate(registry) → JSON
  handler.go         # http.Handler: routes /tasks, /.well-known/agent.json
  task_handler.go    # POST /tasks → create Todo2 task → return taskId
  events_handler.go  # GET /tasks/{id}/events → SSE goroutine
  executor.go        # bridge: A2A task description → tool dispatch
  auth.go            # BearerAuth middleware; key from env/config
  types.go           # A2A wire types (Task, Artifact, Message, etc.)
```

Mount in `cmd/server/main.go` alongside existing `api.NewServer`.

### Agent card generation

Auto-generate from the tool registry at startup:

```go
func Generate(tools []ToolRegistration) AgentCard {
    skills := make([]Skill, 0, len(tools))
    for _, t := range tools {
        skills = append(skills, Skill{
            ID:          t.Name,
            Name:        t.Name,
            Description: stripHint(t.Description),
        })
    }
    return AgentCard{Name: "exarp-go", Skills: skills, ...}
}
```

### Task execution bridge

The hardest part. Two approaches:

**Option A — Free-text routing (v0)**
Accept any text task description, create a Todo2 task, run `task run-with-ai` against it. Simple but low-fidelity — the LLM decides what to do.

**Option B — Skill-addressed routing (v1)**
A2A client specifies `skillId` (matching a tool name). Route directly to that tool handler with the task's `message` as params. Higher fidelity, requires structured input from caller.

Recommend: ship Option A first, layer Option B when a real client exists.

---

## Implementation Phases

### Phase 0 — Spike (1 day)
- Add `/.well-known/agent.json` endpoint with static agent card
- Verify it is parseable by an A2A client (e.g. `a2a-cli` or custom script)
- No task execution

### Phase 1 — Minimal task round-trip (2 days)
- `POST /tasks` → create Todo2 task, return task ID + `submitted` status
- `GET /tasks/{id}` → return current Todo2 status mapped to A2A state
- Bearer token auth middleware (env var `A2A_API_KEY`)

### Phase 2 — Execution + streaming (2 days)
- Connect `POST /tasks` to `task run-with-ai` (Option A executor)
- `GET /tasks/{id}/events` SSE — stream log lines from execution as A2A `working` events
- Final `completed`/`failed` event when task finishes

### Phase 3 — Skill-addressed routing (future)
- Map `skillId` → specific tool handler
- Structured param extraction from A2A `Message`
- Requires a real A2A client (Cursor/OpenCode/Codex) to validate

---

## Client Support Status (as of 2026-03)

| Client | A2A support |
|--------|-------------|
| Claude (Anthropic) | Not yet; MCP is primary protocol |
| Cursor | Not yet |
| OpenCode | Not yet |
| Codex (OpenAI) | Not yet |
| Google ADK agents | Yes (reference implementation) |
| LangChain agents | Experimental |

**Implication:** Phase 0–2 can be built and tested with `curl` or a small test script today. Real client integration is 2026 H2 at earliest.

---

## Related Protocols (context)

| Protocol | Layer | Relevance |
|----------|-------|-----------|
| **MCP** | Tool/data access | Primary protocol; A2A is additive |
| **ANP** | Internet-scale agent discovery | Not relevant; clients are known |
| **Gibberlink** | Binary encoding between LLMs | Not relevant; human-in-the-loop |
| **Semantic comm** | Meaning-layer compression | Research stage; no tooling |

---

## Risks

- **Spec churn** — A2A 0.2 is not stable; Phase 3 work may need rework
- **No real clients** — hard to validate end-to-end until Cursor/OpenCode adopt
- **Auth surface** — adding an HTTP endpoint for task execution is a new attack surface; bearer token auth is minimum viable

---

## Decision

Build Phase 0 (agent card endpoint) when there is a concrete client to test against. Defer Phase 1+ until at least one of Claude/Cursor/OpenCode ships A2A client support or an internal multi-agent use case justifies it.

**Next action when ready:** Create `internal/a2a/agent_card.go` and mount `/.well-known/agent.json` in `cmd/server/main.go`.
