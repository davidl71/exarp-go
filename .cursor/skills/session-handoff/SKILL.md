---
name: session-handoff
description: Session handoff via exarp-go. Use when ending a session (create handoff note), listing all handoffs, resuming from handoff, or exporting handoff data. Prefer session MCP tool with action=handoff, sub_action=end|list|resume|latest|export.
---

# exarp-go Session Handoff

Apply this skill when the user wants to end a session, list handoffs, resume from a handoff, or export handoff data.

## Prefer MCP: session tool

Use the exarp-go `session` tool with `action=handoff` and `sub_action`:

| Operation | MCP args |
|-----------|----------|
| **End session** (create handoff) | `action=handoff`, `sub_action=end`, `summary` (required) |
| **List all handoffs** | `action=handoff`, `sub_action=list`, optional: `limit` (default 5) |
| Resume (latest handoff) | `action=handoff`, `sub_action=resume` |
| Latest only | `action=handoff`, `sub_action=latest` |
| Export | `action=handoff`, `sub_action=export`, optional: `output_path`, `export_latest` |

## End session (create handoff)

**Required:** `summary` — what was done or left off.

**Optional:**
- `blockers` — array of strings or JSON
- `next_steps` — array of strings or JSON
- `include_tasks` — include in-progress tasks (default true)
- `include_git_status` — include git status (default true)
- `unassign_my_tasks` — release task locks for this agent (default true)
- `dry_run` — preview without saving

**Examples (MCP):**
```json
{"action":"handoff","sub_action":"end","summary":"Finished ModelRouter work, pushed v0.3.3"}
{"action":"handoff","sub_action":"end","summary":"Blocked on API key","blockers":["Waiting for OAuth approval"]}
```

## List all handoffs

**Examples (MCP):**
```json
{"action":"handoff","sub_action":"list"}
{"action":"handoff","sub_action":"list","limit":10}
```

## Fallback: CLI

```bash
exarp-go -tool session -args '{"action":"handoff","sub_action":"end","summary":"Your summary"}'
exarp-go -tool session -args '{"action":"handoff","sub_action":"list"}'
```

## TUI (exarp-go interactive)

In the Handoffs view:
- **i** — Start an **interactive** agent with the handoff as the initial prompt (agent opens for conversation; handoff is not closed).
- **e** — Run handoff in child agent (non-interactive) and close the handoff.

**Backend:** The TUI uses the **Cursor CLI** (`agent` / `agent -p`). For other backends, use the correct command locally:
- **Cursor (Claude / Cursor models):** `agent "prompt"` or `agent -p "prompt"`
- **Ollama (local):** `ollama run <model> "prompt"` or exarp-go `ollama` / `text_generate` tools
- See [CURSOR_API_AND_CLI_INTEGRATION.md](../../docs/CURSOR_API_AND_CLI_INTEGRATION.md) §2 for the full backend table.

## Storage

Handoffs are stored in `.todo2/handoffs.json` (last 20 kept). Session prime includes handoff alerts when notes exist from another host.
