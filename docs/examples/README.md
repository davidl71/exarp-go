# exarp-go configuration examples

Example configuration files for **Cursor**, **OpenCode**, and project-specific tools. Copy the relevant file to the path shown and adjust paths/env for your setup.

**Recommended for client projects:** Use the **portable runner** ([cursor-mcp-portable.json](cursor-mcp-portable.json) for Cursor, [opencode-exarp-go-portable.json](opencode-exarp-go-portable.json) for OpenCode). Copy `scripts/run_exarp_go.sh` from exarp-go into your project’s `scripts/` and point the MCP command at it. See [PORTABLE_MCP_RUNNER.md](../PORTABLE_MCP_RUNNER.md).

**For AI agents:** Use these examples to suggest correct `mcp.json`, `opencode.json`, or other config when users ask how to configure exarp-go. Paths use `{{PROJECT_ROOT}}` (Cursor) or placeholders; replace with the user’s project or exarp-go install path.

---

## Cursor MCP (`.cursor/mcp.json`)

| File | Destination | When to use |
|------|-------------|-------------|
| [cursor-mcp-portable.json](cursor-mcp-portable.json) | **Any** project: `.cursor/mcp.json` | **Recommended.** Portable runner: copy `scripts/run_exarp_go.sh` into your project; resolves exarp-go (sibling, PATH, EXARP_GO_ROOT). |
| [cursor-mcp-portable-global.json](cursor-mcp-portable-global.json) | **Any** project: `.cursor/mcp.json` | Use when exarp-go is installed globally (`make install`). Point `command` at `$(go env GOPATH)/bin/run_exarp_go.sh`; set `PROJECT_ROOT` in env. |
| [cursor-mcp-per-project.json](cursor-mcp-per-project.json) | Any project: `.cursor/mcp.json` | Direct binary path (e.g. sibling `../exarp-go/bin/exarp-go`). No script. |
| [cursor-mcp-wrapper.json](cursor-mcp-wrapper.json) | **exarp-go repo only**: `.cursor/mcp.json` | When your workspace is the exarp-go repo; uses root `run-exarp-go.sh` (build-if-needed). |
| [cursor-mcp-binary.json](cursor-mcp-binary.json) | exarp-go repo: `.cursor/mcp.json` | Direct `bin/exarp-go` when workspace is exarp-go. |
| [cursor-mcp-absolute-path.json](cursor-mcp-absolute-path.json) | Any project: `.cursor/mcp.json` | Fallback if `{{PROJECT_ROOT}}` in command fails; use absolute paths. |

- **Global config:** Most users have `~/.cursor/mcp.json` for shared servers (devwisdom, context7, etc.). Use **project** `.cursor/mcp.json` only for the `exarp-go` entry (and any project-specific overrides) so `PROJECT_ROOT` is correct per workspace.
- **Placeholders:** Replace `../exarp-go` or `/path/to/exarp-go` with the actual path to your exarp-go clone or binary. Use `{{PROJECT_ROOT}}` in Cursor; it is substituted by the IDE with the workspace root.

---

## OpenCode / OAC (`opencode.json`)

| File | Destination | When to use |
|------|-------------|-------------|
| [opencode-exarp-go-portable.json](opencode-exarp-go-portable.json) | Global or project: `opencode.json` | **Recommended.** Portable runner: copy `scripts/run_exarp_go.sh` into your project; set `command` and `environment.PROJECT_ROOT` to your project paths. |
| [opencode-exarp-go.json](opencode-exarp-go.json) | Global: `~/.config/opencode/opencode.json` or project: `opencode.json` | Direct exarp-go binary path; set `environment.PROJECT_ROOT` to the project root. |

- Set `environment.PROJECT_ROOT` to the project root (or use the portable runner script which derives it).
- Optional: `timeout` (e.g. `10000` ms) for slow tools like report.

---

## Project config (optional)

| File | Destination | When to use |
|------|-------------|-------------|
| [task_tool_rules.yaml](task_tool_rules.yaml) | Project: `.cursor/task_tool_rules.yaml` | Override or extend tag → tool hints for task_workflow `enrich_tool_hints`. |
| [hooks.json](hooks.json) | Project: `.cursor/hooks.json` | Run a script on session start (e.g. session-prime). |

---

## Quick reference

- **Cursor (client project):** Copy [cursor-mcp-portable.json](cursor-mcp-portable.json) and copy `scripts/run_exarp_go.sh` from exarp-go into your project’s `scripts/`. See [PORTABLE_MCP_RUNNER.md](../PORTABLE_MCP_RUNNER.md).
- **Cursor (exarp-go repo):** Use [cursor-mcp-wrapper.json](cursor-mcp-wrapper.json) or [cursor-mcp-binary.json](cursor-mcp-binary.json).
- **Fix existing config:** After global install, run `make fix-mcp-config` (eval only) or `make fix-mcp-config MCP_CONFIG_FLAGS="--cursor-global"` to update Cursor MCP config to use the installed runner. See [PORTABLE_MCP_RUNNER.md](../PORTABLE_MCP_RUNNER.md).
- **OpenCode:** Prefer [opencode-exarp-go-portable.json](opencode-exarp-go-portable.json) with the portable runner; or [opencode-exarp-go.json](opencode-exarp-go.json) with absolute binary path and `environment.PROJECT_ROOT`. For global install, use path `$(go env GOPATH)/bin/run_exarp_go.sh` and set `PROJECT_ROOT`.
- **AI agents:** When suggesting config, point users to `docs/examples/` and [PORTABLE_MCP_RUNNER.md](../PORTABLE_MCP_RUNNER.md) for the portable runner.

See also: [PORTABLE_MCP_RUNNER.md](../PORTABLE_MCP_RUNNER.md), [CURSOR_MCP_SETUP.md](../CURSOR_MCP_SETUP.md), [OPENCODE_INTEGRATION.md](../OPENCODE_INTEGRATION.md), [.cursor/rules/mcp-configuration.mdc](../../.cursor/rules/mcp-configuration.mdc).
