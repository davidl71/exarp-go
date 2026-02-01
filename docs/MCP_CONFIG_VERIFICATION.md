# MCP Config Verification

**Date:** 2026-02-01  
**Task:** T-404 — Verify MCP config uses Go binary (not Python server)

---

## Overview

exarp-go runs as an MCP server via a **Go binary**. The MCP configuration must point to the compiled Go binary (`bin/exarp-go`), **not** to a Python script or interpreter.

---

## Correct Configuration

**Location:** `~/.cursor/mcp.json` (or project `.cursor/mcp.json` if used)

```json
{
  "mcpServers": {
    "exarp-go": {
      "command": "{{PROJECT_ROOT}}/../exarp-go/bin/exarp-go",
      "args": [],
      "env": {
        "PROJECT_ROOT": "{{PROJECT_ROOT}}"
      },
      "description": "exarp-go MCP tools"
    }
  }
}
```

**Correct command patterns:**
- `path/to/exarp-go/bin/exarp-go` — Go binary
- `{{PROJECT_ROOT}}/../exarp-go/bin/exarp-go` — Relative to workspace
- `/absolute/path/to/bin/exarp-go` — Absolute path

---

## Incorrect Configuration (Python)

**Do NOT use:**
- `python` / `python3` / `uv run python`
- `uv run python -m mcp_stdio_tools.server`
- Any `.py` script path
- `python bridge/execute_tool.py`

If your config uses Python, exarp-go tools will fail or behave unexpectedly. Migrate to the Go binary.

---

## Verification Steps

1. **Locate config**
   ```bash
   cat ~/.cursor/mcp.json | jq '.mcpServers["exarp-go"]'
   ```

2. **Check command**
   - The `command` value must be a path to the exarp-go binary.
   - It must NOT contain `python`, `python3`, `uv run`, or `.py`.

3. **Build binary**
   ```bash
   cd /path/to/exarp-go
   make build
   ls -la bin/exarp-go
   ```

4. **Test MCP server**
   ```bash
   ./bin/exarp-go -list
   ```
   If this works, the binary is ready. Point MCP config to `bin/exarp-go`.

---

## Checklist

- [ ] `command` points to `bin/exarp-go` (or equivalent path)
- [ ] No `python`, `python3`, `uv run`, or `.py` in command
- [ ] `PROJECT_ROOT` set in `env` when using `{{PROJECT_ROOT}}`
- [ ] Binary exists and is executable (`chmod +x bin/exarp-go`)

---

## Troubleshooting: "python tool execution failed" / "execute_tool.py"

**Error:** `report failed: python tool execution failed: ... can't open file '/Users/davidl/go/bridge/execute_tool.py'`

**Cause:** You're running a different `exarp-go` binary (e.g. from `$GOPATH/bin` or `/Users/davidl/go/bin`) that expects a Python bridge. The exarp-go project in this repo is **fully native** and does not use the bridge.

**Fix:**
1. **Use the local binary** from this project:
   ```bash
   cd /path/to/exarp-go
   make build
   ./bin/exarp-go -tool report -args '{"action":"plan"}'
   ```
2. **Or use Makefile targets** (they use the local binary + PROJECT_ROOT):
   ```bash
   make report-plan      # Generate .cursor/plans/<project>.plan.md
   make scorecard-plans  # Generate scorecard improvement plans
   ```
3. **If using `exarp-go` from PATH:** Ensure it's the binary from this project. Run `which exarp-go` — if it points to another directory (e.g. `/Users/davidl/go/bin`), use `./bin/exarp-go` or `make report-plan` instead.

---

## Related

- [.cursor/rules/mcp-configuration.mdc](../.cursor/rules/mcp-configuration.mdc) — MCP configuration guidelines
- [README.md](../README.md) — MCP Configuration section
