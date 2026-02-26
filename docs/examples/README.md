# exarp-go configuration examples

Example configuration files for **Cursor**, **OpenCode**, and project-specific tools. Copy the relevant file to the path shown and adjust paths/env for your setup.

**For AI agents:** Use these examples to suggest correct `mcp.json`, `opencode.json`, or other config when users ask how to configure exarp-go. Paths use `{{PROJECT_ROOT}}` (Cursor) or placeholders; replace with the user’s project or exarp-go install path.

---

## Cursor MCP (`.cursor/mcp.json`)

| File | Destination | When to use |
|------|-------------|-------------|
| [cursor-mcp-binary.json](cursor-mcp-binary.json) | Project: `.cursor/mcp.json` | Run exarp-go binary directly (e.g. sibling clone or installed path). |
| [cursor-mcp-wrapper.json](cursor-mcp-wrapper.json) | Project: `.cursor/mcp.json` | Use `run-exarp-go.sh` so the server builds if needed; set `EXARP_WATCH=0` for Cursor. |
| [cursor-mcp-per-project.json](cursor-mcp-per-project.json) | **Any** project (not exarp-go repo): `.cursor/mcp.json` | Per-project exarp-go: each workspace gets its own `PROJECT_ROOT` and `.todo2`. |

- **Global config:** Most users have `~/.cursor/mcp.json` for shared servers (devwisdom, context7, etc.). Use **project** `.cursor/mcp.json` only for the `exarp-go` entry (and any project-specific overrides) so `PROJECT_ROOT` is correct per workspace.
- **Placeholders:** Replace `../exarp-go` or `/path/to/exarp-go` with the actual path to your exarp-go clone or binary. Use `{{PROJECT_ROOT}}` in Cursor; it is substituted by the IDE with the workspace root.

---

## OpenCode / OAC (`opencode.json`)

| File | Destination | When to use |
|------|-------------|-------------|
| [opencode-exarp-go.json](opencode-exarp-go.json) | Global: `~/.config/opencode/opencode.json` or project: `opencode.json` | Add exarp-go as a local MCP server so OpenCode agents can call task_workflow, report, session, health. |

- Set `environment.PROJECT_ROOT` to the project root (or use a wrapper script that sets it from the current directory).
- Optional: `timeout` (e.g. `10000` ms) for slow tools like report.

---

## Project config (optional)

| File | Destination | When to use |
|------|-------------|-------------|
| [task_tool_rules.yaml](task_tool_rules.yaml) | Project: `.cursor/task_tool_rules.yaml` | Override or extend tag → tool hints for task_workflow `enrich_tool_hints`. |
| [hooks.json](hooks.json) | Project: `.cursor/hooks.json` | Run a script on session start (e.g. session-prime). |

---

## Quick reference

- **Cursor:** Copy one of the `cursor-mcp-*.json` examples to your project’s `.cursor/mcp.json`. Ensure the `command` path points to your exarp-go binary or `run-exarp-go.sh`.
- **OpenCode:** Copy `opencode-exarp-go.json` into `~/.config/opencode/opencode.json` or your project’s `opencode.json`; set `command` and `environment.PROJECT_ROOT`.
- **AI agents:** When suggesting config, point users to `docs/examples/` and the file that matches their client (Cursor vs OpenCode) and style (binary vs wrapper, per-project vs global).

See also: [CURSOR_MCP_SETUP.md](../CURSOR_MCP_SETUP.md), [OPENCODE_INTEGRATION.md](../OPENCODE_INTEGRATION.md), [.cursor/rules/mcp-configuration.mdc](../../.cursor/rules/mcp-configuration.mdc).
