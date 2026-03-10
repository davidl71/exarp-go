# Portable MCP runner for exarp-go

This document describes the **portable runner** script that Cursor, OpenCode, and other MCP clients use to run exarp-go with the correct project context (`PROJECT_ROOT`, `.todo2`, task store) across different hosts and installs.

## What it is

- **Script:** `scripts/run_exarp_go.sh` (in this repo).
- **Purpose:** Resolve the exarp-go binary (or `go run`) and set `PROJECT_ROOT` to the **client project** so exarp-go uses that project’s `.todo2` and config.
- **Use case:** You open a different repo (e.g. `my-app`) in Cursor; that repo’s `.cursor/mcp.json` points to a copy of this script. The script ensures `PROJECT_ROOT` is `my-app` and runs exarp-go from a sibling exarp-go clone, PATH, or `EXARP_GO_ROOT`.

## Setup in a client project

You can use the runner in two ways: **copy into project** (recommended when you don’t install exarp-go globally) or **use globally installed runner** (after `make install`).

### Option A: Copy the script into your project

1. **Copy the script** from exarp-go into your project:
   ```bash
   mkdir -p scripts
   cp /path/to/exarp-go/scripts/run_exarp_go.sh scripts/
   chmod +x scripts/run_exarp_go.sh
   ```

2. **Configure MCP** to use it.

   **Cursor** (`.cursor/mcp.json`):
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
   Cursor substitutes `{{PROJECT_ROOT}}` with the workspace root. The script also derives `PROJECT_ROOT` from the script path if unset.

   **OpenCode** (`opencode.json` or `~/.config/opencode/opencode.json`):
   ```json
   {
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

3. **Restart** Cursor or reload MCP so the new command is used.

### Option B: Use the globally installed runner

When you install exarp-go with `make install`, the portable runner is installed to `$(go env GOPATH)/bin/run_exarp_go.sh` (alongside the `exarp-go` binary). You can point client projects at this script so you don't need to copy it into each repo. If the runner is started **without** `PROJECT_ROOT` set (e.g. your agent runs it with the project directory as the current working directory), it uses the **current directory** as the project root.

1. **Install exarp-go** (from the exarp-go repo):
   ```bash
   make build
   make install
   ```
   This installs `exarp-go` and `run_exarp_go.sh` to `GOPATH/bin`. Ensure `GOPATH/bin` is on your PATH.

2. **In each client project**, set up MCP to use the global runner. You **must** set `PROJECT_ROOT` to the client project root (the runner cannot derive it when it lives in GOPATH/bin).

   **Cursor** (`.cursor/mcp.json`). Replace the path with your `GOPATH/bin`: run `go env GOPATH` and use that value + `/bin/run_exarp_go.sh`, e.g. `/Users/me/go/bin/run_exarp_go.sh`:
   ```json
   {
     "mcpServers": {
       "exarp-go": {
         "command": "/Users/me/go/bin/run_exarp_go.sh",
         "args": [],
         "env": {
           "PROJECT_ROOT": "{{PROJECT_ROOT}}"
         }
       }
     }
   }
   ```

   **OpenCode** — set `command` to the absolute path of the runner and `environment.PROJECT_ROOT` to the project root:
   ```json
   {
     "mcp": {
       "exarp-go": {
         "type": "local",
         "command": ["/Users/me/go/bin/run_exarp_go.sh"],
         "enabled": true,
         "environment": {
           "PROJECT_ROOT": "/absolute/path/to/your/project"
         },
         "timeout": 10000
       }
     }
   }
   ```

3. **Restart** Cursor or reload MCP.

To install only the binary (no runner), run `make install-binary`. To install only the runner (e.g. after copying a new binary into GOPATH/bin), run `make install-runner`.

**Evaluate or fix existing Cursor MCP config:** After installing, you can have the project evaluate and optionally fix the exarp-go entry in your Cursor MCP config so it uses the installed runner:

```bash
make fix-mcp-config                                    # report current config and recommendation
make fix-mcp-config MCP_CONFIG_FLAGS="--cursor-global" # update ~/.cursor/mcp.json
make fix-mcp-config MCP_CONFIG_FLAGS="--cursor-project=/path/to/repo"
make fix-mcp-config MCP_CONFIG_FLAGS="--cursor-global --dry-run"  # show what would change
```

See `scripts/fix-exarp-mcp-config.sh --help` for all options.

## Resolution order

The script chooses which exarp-go to run in this order:

| Priority | Source | Condition |
|----------|--------|-----------|
| 1 | `EXARP_GO_ROOT/bin/exarp-go` | `EXARP_GO_ROOT` set and executable present |
| 2 | CWD exarp-go repo | Walk up from CWD; use `bin/exarp-go` or `go run ./cmd/server` |
| 3 | PATH | `exarp-go` on PATH |
| 4 | Sibling | `PROJECT_ROOT/../exarp-go/bin/exarp-go` |
| 5 | Fallback paths | `~/go/bin/exarp-go`, `~/Projects/mcp/exarp-go/bin/exarp-go`, `/usr/local/bin/exarp-go` |

Exarp-go repo detection: directory has `go.mod` containing `exarp-go` and either `cmd/server/main.go` or executable `bin/exarp-go`.

## Environment variables

| Variable | Purpose |
|----------|---------|
| `PROJECT_ROOT` | Project exarp-go serves (`.todo2`, task store). Set by Cursor/OpenCode or caller. If unset: when the script is inside a project (e.g. `project/scripts/`), that project is used; when the script is installed globally (e.g. in GOPATH/bin), the **current working directory** is used. |
| `EXARP_GO_ROOT` | exarp-go repo root; used for working-dir build and `EXARP_MIGRATIONS_DIR`. |
| `EXARP_GO_VERBOSE=1` | Log to stderr which binary or `go run` is used. |
| `EXARP_MIGRATIONS_DIR` | Set automatically from `EXARP_GO_ROOT/migrations` when using a repo build. |

## Override exarp-go location

```bash
EXARP_GO_ROOT=/path/to/exarp-go ./scripts/run_exarp_go.sh
```

## Debug

```bash
EXARP_GO_VERBOSE=1 ./scripts/run_exarp_go.sh -list
```

## Example configs in this repo

- **Cursor (copy script):** [docs/examples/cursor-mcp-portable.json](examples/cursor-mcp-portable.json) — portable runner in project.
- **Cursor (global install):** [docs/examples/cursor-mcp-portable-global.json](examples/cursor-mcp-portable-global.json) — use when exarp-go is installed via `make install`.
- **Cursor (binary):** [docs/examples/cursor-mcp-per-project.json](examples/cursor-mcp-per-project.json) — direct binary path (no script).
- **OpenCode:** [docs/examples/opencode-exarp-go-portable.json](examples/opencode-exarp-go-portable.json) — portable runner for OpenCode.

See [docs/examples/README.md](examples/README.md) for the full index.

## Evaluate or fix existing config

After `make install`, run `make fix-mcp-config` to report the current exarp-go entry in your Cursor MCP config and the recommended command. To update the config to use the installed runner:

- `make fix-mcp-config MCP_CONFIG_FLAGS="--cursor-global"` — update `~/.cursor/mcp.json`
- `make fix-mcp-config MCP_CONFIG_FLAGS="--cursor-project=/path/to/repo"` — update a project's `.cursor/mcp.json`
- Add `--dry-run` to see what would change without writing.

## Related

- [CURSOR_MCP_SETUP.md](CURSOR_MCP_SETUP.md) — Cursor MCP setup.
- [OPENCODE_INTEGRATION.md](OPENCODE_INTEGRATION.md) — OpenCode integration.
- [.cursor/rules/mcp-configuration.mdc](../.cursor/rules/mcp-configuration.mdc) — MCP configuration rules.
