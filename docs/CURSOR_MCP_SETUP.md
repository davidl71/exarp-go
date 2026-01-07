# Cursor MCP Server Setup Guide

## Current Configuration

The `exarp-go` MCP server is configured in `.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "exarp-go": {
      "command": "/Users/davidl/Projects/exarp-go/bin/exarp-go",
      "args": [],
      "env": {
        "PROJECT_ROOT": "{{PROJECT_ROOT}}"
      },
      "description": "Crew Role: Executor - Go-based MCP server - 24 tools, 15 prompts, 6 resources"
    }
  }
}
```

## Server Status

✅ **Binary Location:** `/Users/davidl/Projects/exarp-go/bin/exarp-go`
✅ **Binary Status:** Built and executable
✅ **Server Type:** STDIO-based MCP server (JSON-RPC 2.0)
✅ **Mode Detection:** Auto-detects TTY (CLI mode) vs non-TTY (MCP server mode)

## Loading the Server in Cursor

### Step 1: Verify Binary Exists

```bash
cd /Users/davidl/Projects/exarp-go
ls -la bin/exarp-go
# Should show: -rwxr-xr-x ... bin/exarp-go
```

### Step 2: Test Server Startup

```bash
# Test that server responds to initialize request
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | ./bin/exarp-go
```

Expected: Server should respond with JSON-RPC response (may show EOF error when stdin closes, which is normal).

### Step 3: Restart Cursor

1. **Quit Cursor completely** (Cmd+Q on macOS)
2. **Reopen Cursor** to load the updated MCP configuration
3. **Check MCP Status:**
   - Open Cursor Settings
   - Navigate to "Features" → "Model Context Protocol"
   - Verify `exarp-go` server is listed and shows "Connected" status

### Step 4: Verify Server is Loaded

In Cursor chat, you should be able to:
- See `exarp-go` tools available
- Use tools like `lint`, `health`, `report`, etc.
- Access prompts and resources

## Troubleshooting

### Server Not Loading

1. **Check binary path:**
   ```bash
   test -f /Users/davidl/Projects/exarp-go/bin/exarp-go && echo "✅ Binary exists" || echo "❌ Binary missing"
   ```

2. **Check binary permissions:**
   ```bash
   chmod +x /Users/davidl/Projects/exarp-go/bin/exarp-go
   ```

3. **Rebuild binary:**
   ```bash
   cd /Users/davidl/Projects/exarp-go
   go build -o bin/exarp-go ./cmd/server
   ```

4. **Check Cursor logs:**
   - Open Cursor Settings
   - Navigate to "Features" → "Model Context Protocol"
   - Click on `exarp-go` server
   - View error logs if server failed to start

### Server Crashes on Startup

1. **Check Go version:**
   ```bash
   go version
   # Should be 1.24.0 or later
   ```

2. **Check dependencies:**
   ```bash
   cd /Users/davidl/Projects/exarp-go
   go mod download
   go mod verify
   ```

3. **Test server manually:**
   ```bash
   ./bin/exarp-go
   # Should wait for input (STDIO mode)
   # Press Ctrl+C to exit
   ```

### Configuration Issues

1. **Validate JSON:**
   ```bash
   cat .cursor/mcp.json | python3 -m json.tool
   # Should show valid JSON without errors
   ```

2. **Check environment variables:**
   - `PROJECT_ROOT` should be set to `{{PROJECT_ROOT}}` (Cursor will substitute)
   - No other environment variables required

## Alternative: Using Wrapper Script

If you prefer to use the wrapper script (with auto-rebuild), update the config:

```json
{
  "mcpServers": {
    "exarp-go": {
      "command": "/Users/davidl/Projects/exarp-go/run-exarp-go.sh",
      "args": [],
      "env": {
        "PROJECT_ROOT": "{{PROJECT_ROOT}}",
        "EXARP_WATCH": "0"
      }
    }
  }
}
```

**Note:** Set `EXARP_WATCH=0` to disable watch mode for Cursor (watch mode is for development only).

## Server Capabilities

Once loaded, the `exarp-go` server provides:

- **24 Tools:** lint, health, report, memory, security, testing, etc.
- **15 Prompts:** align, discover, config, scan, scorecard, etc.
- **6 Resources:** scorecard, memories (by category/task/session/recent)

## Next Steps

After server is loaded:
1. Test a tool: `@exarp-go lint`
2. Check health: `@exarp-go health`
3. View scorecard: Access `stdio://scorecard` resource

---

**Last Updated:** 2026-01-07
**Status:** ✅ Configuration ready for Cursor

