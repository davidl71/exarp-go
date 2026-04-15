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

## Non-Go quickstarts

These are the shortest practical setup flows for client repos that are not Go projects.

### JavaScript / TypeScript repo with Cursor

Example client repo:

```text
my-web-app/
  .cursor/
  scripts/
  package.json
  src/
```

Steps:

1. Copy `scripts/run_exarp_go.sh` from the `exarp-go` repo into `my-web-app/scripts/`.
2. Create `my-web-app/.cursor/mcp.json` from [cursor-mcp-portable.json](cursor-mcp-portable.json).
3. Restart Cursor or reload MCP.

Resulting config:

```json
{
  "mcpServers": {
    "exarp-go": {
      "command": "{{PROJECT_ROOT}}/scripts/run_exarp_go.sh",
      "args": [],
      "env": {
        "PROJECT_ROOT": "{{PROJECT_ROOT}}"
      }
    }
  }
}
```

Verify from the JavaScript/TypeScript repo:

```bash
cd /path/to/my-web-app
PROJECT_ROOT="$PWD" EXARP_GO_VERBOSE=1 ./scripts/run_exarp_go.sh -list
```

You should see the runner resolve an `exarp-go` binary or repo and print the available tools.

### Python repo with Cursor

Example client repo:

```text
my-python-service/
  .cursor/
  scripts/
  pyproject.toml
  app/
```

Steps:

1. Copy `scripts/run_exarp_go.sh` from the `exarp-go` repo into `my-python-service/scripts/`.
2. Create `my-python-service/.cursor/mcp.json` from [cursor-mcp-portable.json](cursor-mcp-portable.json).
3. Restart Cursor or reload MCP.

Use the same config as above:

```json
{
  "mcpServers": {
    "exarp-go": {
      "command": "{{PROJECT_ROOT}}/scripts/run_exarp_go.sh",
      "args": [],
      "env": {
        "PROJECT_ROOT": "{{PROJECT_ROOT}}"
      }
    }
  }
}
```

Verify from the Python repo:

```bash
cd /path/to/my-python-service
PROJECT_ROOT="$PWD" EXARP_GO_VERBOSE=1 ./scripts/run_exarp_go.sh -list
```

The Python repo does not need `go.mod` or any Go sources. The only requirement is that the runner can locate `exarp-go`.

### OpenCode in any non-Go repo

For OpenCode, use [opencode-exarp-go-portable.json](opencode-exarp-go-portable.json), but note the main difference from Cursor:

- `command` must be an absolute path to the runner script
- `environment.PROJECT_ROOT` must be the absolute path to the client repo

Minimal example:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "exarp-go": {
      "type": "local",
      "command": ["/absolute/path/to/your/project/scripts/run_exarp_go.sh"],
      "enabled": true,
      "environment": {
        "PROJECT_ROOT": "/absolute/path/to/your/project"
      },
      "timeout": 10000
    }
  }
}
```

---

## Verification checklist

After wiring a client repo:

1. Confirm the runner script exists at `scripts/run_exarp_go.sh`.
2. Confirm the MCP config file lives in the client repo, not in the `exarp-go` repo.
3. Confirm `PROJECT_ROOT` resolves to the client repo.
4. Run `PROJECT_ROOT="$PWD" EXARP_GO_VERBOSE=1 ./scripts/run_exarp_go.sh -list` from the client repo.
5. In Cursor or OpenCode, confirm `exarp-go` appears as a connected MCP server.

If the runner cannot find `exarp-go`, set `EXARP_GO_ROOT=/path/to/exarp-go` explicitly and rerun the debug command.

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
