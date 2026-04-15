# Cursor MCP Setup Guide

This guide explains how to attach the `exarp-go` MCP server to a project opened in Cursor. The default case is a client repo of any language: Go, Python, JavaScript/TypeScript, or otherwise.

## Choose the right setup

Use one of these two paths:

1. Client repo in any language: use the portable runner.
2. The `exarp-go` repo itself: use the repo-local wrapper or binary config.

For most users, option 1 is the right one.

## Option 1: Client repo in any language

This is the recommended setup when the repo you opened in Cursor is not the `exarp-go` repo.

### What you need

- A client repo open in Cursor
- Access to an `exarp-go` install, sibling clone, or a checkout referenced by `EXARP_GO_ROOT`
- A copy of `scripts/run_exarp_go.sh` inside the client repo

### Step 1: Copy the runner into the client repo

From the `exarp-go` repo:

```bash
mkdir -p /path/to/your/project/scripts
cp /path/to/exarp-go/scripts/run_exarp_go.sh /path/to/your/project/scripts/
chmod +x /path/to/your/project/scripts/run_exarp_go.sh
```

### Step 2: Add `.cursor/mcp.json` in the client repo

Create or update `.cursor/mcp.json` in the client repo:

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

Why this works:
- Cursor replaces `{{PROJECT_ROOT}}` with the workspace root
- `exarp-go` uses that value as the project context for `.todo2`, config, and related tools
- the client repo does not need to be a Go project

### Step 3: Restart Cursor or reload MCP

After editing `.cursor/mcp.json`, restart Cursor or reload MCP so the new server entry is picked up.

### Step 4: Verify from the client repo

Use checks that match the client-repo workflow, not the `exarp-go` repo internals:

1. In Cursor settings, verify `exarp-go` appears under Model Context Protocol and shows as connected.
2. In Cursor chat, confirm the `exarp-go` tools are available.
3. Run a simple tool from the client repo context, such as `health`, `report`, or `session prime`.

If you want a shell-level debug check from the client repo:

```bash
cd /path/to/your/project
PROJECT_ROOT="$PWD" EXARP_GO_VERBOSE=1 ./scripts/run_exarp_go.sh -list
```

That confirms the runner can resolve `exarp-go` and start the server entrypoint for the client repo.

### Common client-repo mistakes

- Putting the config in the `exarp-go` repo instead of the client repo
- Forgetting to copy `scripts/run_exarp_go.sh`
- Setting `PROJECT_ROOT` to the `exarp-go` repo instead of the client repo
- Assuming the client repo must contain Go files or a `go.mod`

## Option 2: The `exarp-go` repo itself

Use this only when your Cursor workspace is the `exarp-go` repository.

### Repo-local wrapper example

```json
{
  "mcpServers": {
    "exarp-go": {
      "command": "{{PROJECT_ROOT}}/run-exarp-go.sh",
      "args": [],
      "env": {
        "PROJECT_ROOT": "{{PROJECT_ROOT}}",
        "EXARP_WATCH": "0"
      }
    }
  }
}
```

### Repo-local binary example

```json
{
  "mcpServers": {
    "exarp-go": {
      "command": "{{PROJECT_ROOT}}/bin/exarp-go",
      "args": [],
      "env": {
        "PROJECT_ROOT": "{{PROJECT_ROOT}}"
      }
    }
  }
}
```

Use the wrapper when you want build-if-needed behavior. Use the binary when `bin/exarp-go` already exists and you want a direct command.

## Troubleshooting

### Server not showing as connected

Check:

- `.cursor/mcp.json` is valid JSON
- the `command` path exists
- `PROJECT_ROOT` points at the workspace you actually opened in Cursor
- Cursor was restarted or MCP was reloaded after config changes

### Portable runner cannot find `exarp-go`

The runner resolves `exarp-go` in this order:

1. `EXARP_GO_ROOT/bin/exarp-go`
2. current working tree if it is an `exarp-go` repo
3. `exarp-go` on `PATH`
4. sibling `../exarp-go/bin/exarp-go`
5. common fallback install paths

If needed, force the location explicitly:

```bash
EXARP_GO_ROOT=/path/to/exarp-go PROJECT_ROOT=/path/to/your/project ./scripts/run_exarp_go.sh -list
```

### Config validates but tools still use the wrong project

This is almost always a `PROJECT_ROOT` issue. For Cursor, use `{{PROJECT_ROOT}}` in `.cursor/mcp.json` and keep that config file inside the client repo you want `exarp-go` to serve.

## Related docs

- [PORTABLE_MCP_RUNNER.md](PORTABLE_MCP_RUNNER.md)
- [docs/examples/README.md](examples/README.md)
- [OPENCODE_INTEGRATION.md](OPENCODE_INTEGRATION.md)
- [MCP_ROOTS_ELICITATION.md](MCP_ROOTS_ELICITATION.md)

**Last Updated:** 2026-03-08
**Status:** Client-repo-first setup guide
