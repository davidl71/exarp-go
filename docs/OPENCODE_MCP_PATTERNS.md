# OpenCode MCP Server Patterns - exarp-go Implementation Status

This document analyzes which OpenCode MCP server configuration patterns are implemented in exarp-go.

## OpenCode MCP Configuration Options

| Feature | OpenCode Config | exarp-go Status |
|---------|-----------------|-----------------|
| Local MCP | `type: "local"` | ✅ Implemented |
| Command array | `command: ["path", "args"]` | ✅ Implemented |
| Environment variables | `environment: {...}` | ✅ Implemented |
| Enabled/disabled | `enabled: true/false` | ✅ Implemented |
| Timeout | `timeout: 5000` | ⚠️ Not exposed in config |
| Remote MCP | `type: "remote"` | ⏳ Not implemented |
| Headers | `headers: {...}` | N/A for local |
| OAuth | `oauth: {...}` | N/A for local |
| Glob tool filtering | `tools: {"pattern*": false}` | ⚠️ Basic support |

## Current exarp-go opencode.json

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "exarp-go": {
      "type": "local",
      "command": ["/path/to/exarp-go"],
      "enabled": true,
      "environment": {
        "PROJECT_ROOT": "/path/to/project",
        "EXARP_MIGRATIONS_DIR": "/path/to/migrations",
        "EXARP_WATCH": "0"
      }
    }
  }
}
```

## Implemented Patterns

### 1. Local MCP Server ✅
- Uses `type: "local"`
- Command points to binary or wrapper script
- Environment variables set PROJECT_ROOT and other options

### 2. Environment Variables ✅
- `PROJECT_ROOT` - Required for task/todo2 storage
- `EXARP_MIGRATIONS_DIR` - Optional, for database migrations
- `EXARP_WATCH` - Disable file watching (set to "0")

### 3. Enabled/Disabled ✅
- `enabled: true` to activate
- Set to `false` to disable without removing config

## Patterns That Could Be Implemented

### 1. Timeout Configuration
OpenCode supports `timeout` for fetching tools. Currently exarp-go doesn't expose this.

**Potential enhancement**: Add to environment or config:
```json
{
  "environment": {
    "EXARP_MCP_TIMEOUT": "10000"
  }
}
```

### 2. Remote MCP (HTTP Transport)
OpenCode supports `type: "remote"` with `url`. exarp-go could support:
- SSE transport (Server-Sent Events)
- StreamableHTTP transport

**Current status**: 
- mcp-go-core already has `SSETransport` and supports StreamableHTTP via go-sdk
- exarp-go server doesn't expose transport selection
- ACP mode (`-acp`) provides an alternative protocol

**Related task**: T-1773328157358788000 - "Add Streamable HTTP transport for remote MCP"

### 3. Tool Filtering via Glob Patterns
OpenCode supports disabling tools via glob patterns:
```json
{
  "tools": {
    "exarp-go_task_workflow": true,
    "exarp-go_*": false
  }
}
```

**Current status**: exarp-go has basic tool filtering via `tools.ToolFilterForMode()` but not exposed via OpenCode config.

### 4. Per-Agent Tool Configuration
OpenCode supports enabling tools per-agent:
```json
{
  "agent": {
    "my-agent": {
      "tools": {
        "exarp-go_*": true
      }
    }
  }
}
```

**Status**: Not implemented - would require OpenCode-side configuration.

## Recommendations

1. **Document current capabilities** - This document serves as reference
2. **Consider adding timeout env var** - Low effort, useful for slow environments
3. **Remote MCP is a medium effort feature** - Requires transport selection in server
4. **Tool filtering could be enhanced** - Could read from environment or config

## Environment Variables Reference

| Variable | Description | Required |
|----------|-------------|----------|
| PROJECT_ROOT | Project directory for .todo2 and config | Yes |
| EXARP_MIGRATIONS_DIR | Path to migrations folder | No |
| EXARP_WATCH | Set to "0" to disable file watching | No |
| EXARP_MCP_TIMEOUT | Timeout for MCP operations (ms) | No |
