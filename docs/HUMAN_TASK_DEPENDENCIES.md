# Human-Only Task Dependencies and Breakdown

**Tag hints:** `#docs` `#planning` `#opencode` `#cursor`

This document records dependencies for human-only tasks and proposes breakdown into implementable subtasks.

---

## 1. T-1771355092457125000 — Validate exarp-go with OpenCode MCP

**Status:** Human-only (Todo)  
**Dependencies:** None (can run once OpenCode is installed and config exists)

### What blocks completion

| Blocker | Type | Description |
|--------|------|-------------|
| OpenCode runtime | External | Must have OpenCode installed (`opencode` on PATH) |
| Interactive session | Human | Must run OpenCode TUI, ask agent to list/call tools, visually verify |
| Config correctness | Human | User must add exarp-go to `~/.config/opencode/opencode.json` or project `opencode.json` |

### Implementable subtasks (can create)

| Subtask | Description | Deliverable |
|---------|-------------|-------------|
| **OpenCode config validation script** | Script or `exarp-go` subcommand that checks: opencode config file exists, exarp-go entry present, `command` path exists, `PROJECT_ROOT` set | `scripts/validate-opencode-config.sh` or `exarp-go validate opencode` |
| **MCP stdio smoke test** | Run exarp-go in stdio mode, send minimal JSON-RPC (e.g. `initialize` + `tools/list`), assert response | CI or local script; validates exarp-go MCP without OpenCode |
| **Validation checklist doc** | Structured checklist in OPENCODE_INTEGRATION.md with explicit pass/fail criteria | Update §1 "Verification" with numbered checklist |

### Remaining human-only

- Run OpenCode, connect to agent, ask "list my tasks" and confirm exarp-go tools appear and work.
- Record result (e.g. add note to task).

---

## 2. T-1771379764381696000 — Validate and submit exarp-go Cursor plugin

**Status:** Human-only (Todo)  
**Dependencies:** T-1771379759827804000 (manifest), T-1771379763585865000 (MCP config) — both Done.

### What blocks completion

| Blocker | Type | Description |
|--------|------|-------------|
| Cursor plugin validation | Semi-automated | `node scripts/validate-template.mjs` from cursor/plugin-template — requires Node, cloning template |
| Cursor Marketplace submission | Human | Submit repo link to Cursor team (Slack or kniparko@anysphere.com) |
| Plugin install test | Human | Install plugin in Cursor, verify MCP appears |

### Implementable subtasks (can create)

| Subtask | Description | Deliverable |
|---------|-------------|-------------|
| **Makefile target for plugin validation** | `make validate-plugin` — clone or use cursor/plugin-template validate script, run against `.cursor-plugin/` and `mcp.json` | Makefile target |
| **Plugin structure validation** | Script that checks: `.cursor-plugin/plugin.json` exists, required fields present, `mcp.json` valid JSON, exarp-go server entry | `scripts/validate-cursor-plugin.sh` or similar |
| **Submission runbook** | Step-by-step doc: how to run validation, what to send Cursor, where to submit | Add §3.1 to CURSOR_PLUGIN_PLAN.md or `docs/CURSOR_PLUGIN_SUBMISSION.md` |

### Remaining human-only

- Run validation (can be scripted; human runs it).
- Submit plugin to Cursor (email/Slack — inherently human).
- Test installed plugin in Cursor (human).

---

## 3. Implementable tasks (created) — all Done

| Task ID | Name | Delivered |
|---------|------|-----------|
| T-1771431470154388000 | OpenCode: Add config validation script | `scripts/validate-opencode-config.sh`, `make validate-opencode-config` |
| T-1771431471680197000 | OpenCode: Add MCP stdio smoke test | `scripts/test-mcp-stdio.sh`, `make test-mcp-stdio` |
| T-1771431474062739000 | Cursor plugin: Add Makefile validate-plugin target | `scripts/validate-cursor-plugin.sh`, `make validate-plugin` |
| T-1771431476232391000 | Cursor plugin: Add submission runbook to docs | `docs/CURSOR_PLUGIN_SUBMISSION.md` |

---

## 4. References

- [OPENCODE_INTEGRATION.md](OPENCODE_INTEGRATION.md) — OpenCode MCP setup and verification checklist
- [CURSOR_PLUGIN_PLAN.md](CURSOR_PLUGIN_PLAN.md) — Plugin phases and validation
- [docs/opencode-exarp-go.example.json](opencode-exarp-go.example.json) — Example OpenCode config
