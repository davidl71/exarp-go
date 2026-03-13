# Agent Client Compatibility: Claude Code, Cursor, OpenCode

**Date:** 2026-03-13
**Status:** Research / Backlog

---

## Overview

exarp-go is an MCP server consumed by AI coding agents. Each client (Claude Code, Cursor, OpenCode) has different MCP behaviours, session lifecycle patterns, and feature surfaces. This document maps gaps and improvement opportunities per client.

---

## Claude Code

Claude Code is the primary client and is well-served today. Gaps are minor.

### Current state
- `CLAUDE.md` fully documents exarp tools, session workflow, and CLI shortcuts
- `session action=prime` provides full context in one call
- Compact JSON defaults (just added) reduce per-call token cost
- Shorter HINT strings reduce tool-selection overhead for subagents

### Gaps

#### `cursor_cli_suggestion` in session prime (T-1773431268285525000)
**Problem:** `session prime` always includes `cursor_cli_suggestion` — a Cursor-specific field that wastes tokens for Claude Code callers.
**Fix:** Add optional `client=claude|cursor|opencode` param to `handleSessionPrime`; suppress Cursor-only fields when `client=claude`.
**Effort:** Small (1–2h)

---

## Cursor

Cursor uses MCP but agents start cold — no automatic `session prime` call. Discovery and background task execution are the main opportunities.

### Current state
- `.cursor/mcp.json` example in `docs/examples/`
- `cursor_cli_suggestion` in session prime points agents at the right CLI command
- `generate_config` tool can create `.cursor/rules/*.mdc` files
- A2A Phase 0 researched (see `docs/research/a2a-protocol.md`)

### Gaps

#### Cursor Background Agent task claim (T-1773431269560179000)
**Problem:** Cursor Background Agents run autonomously but have no convention for claiming a Todo2 task before starting work, risking duplication when multiple agents are active.
**Fix:** Document (and potentially expose as a tool action) the `ClaimTaskForAgent` database lock so Cursor agents can safely claim tasks before execution.
**Effort:** Small–Medium (doc + possible `task_workflow action=claim`)

#### `X-Client-Name` header detection in SSE mode (T-1773431269377758000)
**Problem:** exarp serves all MCP clients identically over SSE. Cursor-specific behaviour (e.g. including `cursor_cli_suggestion`, richer error context) requires knowing the caller.
**Fix:** Read `X-Client-Name: cursor` (or equivalent) from SSE request headers; store in context; use in handlers to tailor output.
**Effort:** Medium (cross-cutting; touches SSE transport + multiple handlers)

#### `agent://card` resource (T-1773431269013462000)
**Problem:** Cursor agents reading tool schemas cold consume many tokens. A single resource listing all capabilities would let agents bootstrap faster.
**Fix:** Expose `agent://card` MCP resource (via `TrackResource`) that returns name, version, and all tool names + one-line descriptions. Auto-generated from the tool registry at startup.
**Effort:** Small (1 day)

---

## OpenCode

OpenCode is the newest client with the least mature MCP support. Focus on transport reliability and schema completeness.

### Current state
- `opencode.json` example in `docs/examples/`
- SSE transport exists but not validated against OpenCode
- No session bootstrap convention (OpenCode has no equivalent of `session prime` auto-call)

### Gaps

#### SSE transport compatibility (T-1773431268466254000)
**Problem:** OpenCode's MCP SSE implementation may differ from Claude Code / Cursor in subtle ways (keep-alive intervals, reconnect behaviour, header requirements). Not yet tested.
**Fix:** Stand up OpenCode pointed at exarp SSE endpoint; run through tool calls, streaming, and error cases; document any quirks and fix transport-layer issues found.
**Effort:** Small (1 day testing + fixes)

#### Tool schema `default` fields audit (T-1773431268648656000)
**Problem:** OpenCode uses JSON Schema `default` values to pre-fill tool params in its UI. Many exarp tools omit `default` on optional fields, degrading the OpenCode experience.
**Fix:** Audit all 37 tool schemas; add `default` to optional fields where a sensible default exists (e.g. `output_format: "text"`, `compact: true`, `priority: "medium"`).
**Effort:** Small (half day)

#### `prime://context` resource for session bootstrap (T-1773431268833199000)
**Problem:** OpenCode has no session start hook, so agents never call `session prime`. Without it they lack task context, hints, and handoff state.
**Fix:** Expose `prime://context` as an MCP resource that returns the same payload as `session action=prime`. OpenCode (and other clients) can subscribe to it on connect without an explicit tool call.
**Effort:** Medium (1–2 days; resource infrastructure + session prime refactor to share logic)

---

## Cross-Client (all three)

These benefit Claude Code, Cursor, and OpenCode equally.

### `agent://card` resource (T-1773431269013462000)
Auto-generated from the tool registry at server startup. Exposes:
```json
{
  "name": "exarp-go",
  "version": "0.48.0",
  "tools": [
    {"name": "task_workflow", "description": "Todo2 task CRUD and AI execution"},
    ...
  ],
  "prompts": [...],
  "resources": [...]
}
```
Single read gives any agent a full capability map without scanning all tool schemas.

### Tool schema `examples` field (T-1773431269195579000)
MCP spec supports an `examples` array on tool schemas. Adding one canonical example per tool reduces agent trial-and-error on first call.

```go
"examples": []map[string]interface{}{
    {"action": "list", "status": "Todo", "priority": "high"},
},
```
**Effort:** Small–Medium (37 tools; could be done incrementally starting with top-10 most-called)

### Compact JSON as MCP default (done)
Already implemented: `task_workflow action=list output_format=json` and `session action=prime` both default to compact JSON. No client needs to pass `compact=true` explicitly.

---

## Priority Matrix

| Task | Client | Priority | Effort | ID |
|------|--------|----------|--------|----|
| `agent://card` resource | All | Medium | Small | T-1773431269013462000 |
| `prime://context` resource | OpenCode | Medium | Medium | T-1773431268833199000 |
| SSE compatibility audit | OpenCode | Medium | Small | T-1773431268466254000 |
| Cursor Background Agent claim | Cursor | Medium | Small | T-1773431269560179000 |
| Tool schema `examples` | All | Low | Medium | T-1773431269195579000 |
| Tool schema `default` audit | OpenCode | Low | Small | T-1773431268648656000 |
| `client=` param in session prime | Claude Code | Low | Small | T-1773431268285525000 |
| `X-Client-Name` SSE detection | Cursor/OpenCode | Low | Medium | T-1773431269377758000 |

---

## Recommended sequencing

1. **`agent://card` resource** — highest leverage, benefits all three clients, small effort
2. **SSE compatibility audit** — blocking for OpenCode adoption, quick to validate
3. **`prime://context` resource** — enables OpenCode session context, medium effort
4. **Cursor Background Agent claim** — enables safe multi-agent workflows in Cursor
5. **`default` field audit** — polish; do alongside any schema work
6. **`examples` field** — good but low urgency; do incrementally

A2A Phase 0 (agent card HTTP endpoint) is tracked separately in `docs/research/a2a-protocol.md` (T-1773431009286563000).
