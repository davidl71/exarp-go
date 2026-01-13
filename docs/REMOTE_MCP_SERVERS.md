# Remote MCP Server Configuration Guide

## Overview

Yes, Cursor can use remote MCP servers! There are several approaches depending on your needs.

## Method 1: SSH-based Remote MCP Servers (Recommended)

You can configure Cursor to run MCP server commands via SSH, allowing you to use remote servers.

### Configuration Example

**Local Cursor connecting to Remote MCP Server:**

```json
{
  "mcpServers": {
    "remote-exarp-go": {
      "command": "ssh",
      "args": [
        "davids-mac-mini.tailf62197.ts.net",
        "/Users/davidl/Projects/exarp-go/bin/exarp-go"
      ],
      "env": {
        "PROJECT_ROOT": "{{PROJECT_ROOT}}"
      },
      "description": "Remote exarp-go MCP server via SSH"
    },
    "remote-advisor": {
      "command": "ssh",
      "args": [
        "davids-mac-mini.tailf62197.ts.net",
        "/Users/davidl/Projects/devwisdom-go/run-advisor.sh"
      ],
      "env": {
        "PROJECT_ROOT": "{{PROJECT_ROOT}}"
      },
      "description": "Remote advisor MCP server via SSH"
    }
  }
}
```

### Requirements

1. **SSH Key Authentication:**
   - Set up SSH keys for passwordless access
   - Test: `ssh davids-mac-mini.tailf62197.ts.net "echo 'Connected'"`

2. **SSH Config:**
   - Add to `~/.ssh/config` for easier management:
   ```ssh-config
   Host davids-mac-mini
       HostName davids-mac-mini.tailf62197.ts.net
       User davidl
       StrictHostKeyChecking accept-new
   ```

3. **Remote Binary Paths:**
   - Ensure paths exist on remote machine
   - Binaries must be executable
   - Test: `ssh davids-mac-mini.tailf62197.ts.net "/path/to/binary --help"`

### Advantages

- ✅ Use remote resources (GPU, MLX, Ollama)
- ✅ Centralized server management
- ✅ Leverage remote hardware capabilities
- ✅ Works with stdio-based MCP servers

### Disadvantages

- ⚠️ Network latency (small, but present)
- ⚠️ Requires stable SSH connection
- ⚠️ Remote machine must be accessible

## Method 2: Cursor Background Agents

Cursor has built-in Background Agents feature that runs agents in isolated remote environments.

### Features

- **Background Agents**: Spawn asynchronous agents that run in remote environments
- **Parallel Execution**: Run up to 8 agents in parallel
- **Isolated Environments**: Each agent works in its own environment
- **Remote Machines**: Can use git worktrees or remote machines

### Access

- **Background Agent Sidebar**: View in Cursor interface
- **Trigger**: `Ctrl+E` (or `Cmd+E` on macOS)
- **Status**: View agent status, send follow-ups, or take over tasks

### Requirements

- ✅ Cursor Pro plan (for unlimited agent requests)
- ✅ Internet connection
- ✅ Remote machine access (if using custom remote setup)

## Method 3: Network-Based MCP Servers (SSE/HTTP)

Some MCP servers support network transport (SSE/HTTP) instead of stdio.

### Configuration Example

```json
{
  "mcpServers": {
    "remote-http-server": {
      "url": "https://remote-server.example.com/mcp",
      "transport": "sse"
    }
  }
}
```

### Limitations

- ⚠️ Most MCP servers use stdio (local only)
- ⚠️ Requires server to support HTTP/SSE transport
- ⚠️ More complex setup (authentication, security)

## Recommended Setup for Your Use Case

Based on your setup (local Linux machine, remote macOS M4):

### Option A: SSH-based Remote MCP Servers (Best for Remote Resources)

**Use when:**
- You want to leverage remote MLX/Ollama on macOS M4
- Remote has better hardware/resources
- You want centralized server management

**Configuration:**
```json
{
  "mcpServers": {
    "remote-exarp-go": {
      "command": "ssh",
      "args": [
        "davids-mac-mini.tailf62197.ts.net",
        "/Users/davidl/Projects/exarp-go/bin/exarp-go"
      ],
      "env": {
        "PROJECT_ROOT": "{{PROJECT_ROOT}}"
      }
    },
    "remote-ollama": {
      "command": "ssh",
      "args": [
        "davids-mac-mini.tailf62197.ts.net",
        "/Users/davidl/Projects/exarp-go/bin/exarp-go"
      ],
      "env": {
        "PROJECT_ROOT": "{{PROJECT_ROOT}}",
        "OLLAMA_HOST": "http://localhost:11434"
      }
    }
  }
}
```

### Option B: Local MCP Servers (Best for Performance)

**Use when:**
- You want lowest latency
- Local resources are sufficient
- No need for remote hardware

**Configuration:**
```json
{
  "mcpServers": {
    "local-exarp-go": {
      "command": "/home/dlowes/projects/exarp-go/bin/exarp-go",
      "args": [],
      "env": {
        "PROJECT_ROOT": "{{PROJECT_ROOT}}"
      }
    }
  }
}
```

### Option C: Hybrid Approach (Best of Both Worlds)

**Use both local and remote:**
- Local: Fast tools (lint, format, etc.)
- Remote: Resource-intensive tools (MLX, Ollama, CodeLlama)

**Configuration:**
```json
{
  "mcpServers": {
    "local-exarp-go": {
      "command": "/home/dlowes/projects/exarp-go/bin/exarp-go",
      "args": [],
      "description": "Local tools (fast, low latency)"
    },
    "remote-ollama": {
      "command": "ssh",
      "args": [
        "davids-mac-mini.tailf62197.ts.net",
        "/Users/davidl/Projects/exarp-go/bin/exarp-go"
      ],
      "env": {
        "PROJECT_ROOT": "{{PROJECT_ROOT}}",
        "OLLAMA_HOST": "http://localhost:11434"
      },
      "description": "Remote MLX/Ollama tools (powerful hardware)"
    }
  }
}
```

## Setup Instructions

### Step 1: Verify SSH Access

```bash
# Test SSH connection
ssh davids-mac-mini.tailf62197.ts.net "echo '✅ SSH working'"

# Test remote binary
ssh davids-mac-mini.tailf62197.ts.net "/Users/davidl/Projects/exarp-go/bin/exarp-go --help"
```

### Step 2: Configure SSH Keys (if needed)

```bash
# Generate SSH key (if you don't have one)
ssh-keygen -t ed25519 -C "cursor-mcp"

# Copy to remote machine
ssh-copy-id davids-mac-mini.tailf62197.ts.net

# Test passwordless access
ssh davids-mac-mini.tailf62197.ts.net "echo '✅ Passwordless SSH working'"
```

### Step 3: Add to Cursor MCP Config

Edit `.cursor/mcp.json` in your workspace:

```json
{
  "mcpServers": {
    "remote-exarp-go": {
      "command": "ssh",
      "args": [
        "davids-mac-mini.tailf62197.ts.net",
        "/Users/davidl/Projects/exarp-go/bin/exarp-go"
      ],
      "env": {
        "PROJECT_ROOT": "{{PROJECT_ROOT}}"
      },
      "description": "Remote exarp-go MCP server (macOS M4)"
    }
  }
}
```

### Step 4: Restart Cursor

1. Quit Cursor completely
2. Reopen Cursor
3. Check MCP status in Settings → Features → Model Context Protocol

## Troubleshooting

### SSH Connection Issues

**Problem:** Connection timeout or refused

**Solution:**
1. Verify SSH works: `ssh davids-mac-mini.tailf62197.ts.net`
2. Check Tailscale connectivity: `ping 100.117.185.32`
3. Verify SSH keys are set up correctly

### Remote Binary Not Found

**Problem:** "command not found" or "No such file"

**Solution:**
1. Verify path exists: `ssh davids-mac-mini.tailf62197.ts.net "ls -la /path/to/binary"`
2. Check executable permissions: `ssh davids-mac-mini.tailf62197.ts.net "chmod +x /path/to/binary"`
3. Use absolute paths in MCP config

### Environment Variables Not Working

**Problem:** Remote server can't find dependencies

**Solution:**
1. Add PATH to env:
   ```json
   "env": {
     "PATH": "/opt/homebrew/bin:/usr/local/bin:...",
     "PROJECT_ROOT": "{{PROJECT_ROOT}}"
   }
   ```
2. Create wrapper script on remote that sets environment

### Performance Issues

**Problem:** Slow response times

**Solution:**
1. Use local servers for frequently-used tools
2. Only use remote for resource-intensive operations
3. Check network latency: `ping davids-mac-mini.tailf62197.ts.net`
4. Ensure Tailscale connection is stable

## Best Practices

1. **Use SSH Keys**: Passwordless authentication for reliability
2. **Test First**: Verify remote commands work before adding to MCP config
3. **Monitor Performance**: Watch for latency issues
4. **Hybrid Approach**: Use local for speed, remote for resources
5. **Fallback**: Keep local servers as backup if remote is unavailable

## Examples

### Remote Ollama + MLX Setup

```json
{
  "mcpServers": {
    "remote-exarp-go": {
      "command": "ssh",
      "args": [
        "davids-mac-mini.tailf62197.ts.net",
        "/Users/davidl/Projects/exarp-go/bin/exarp-go"
      ],
      "env": {
        "PROJECT_ROOT": "{{PROJECT_ROOT}}",
        "PATH": "/opt/homebrew/bin:/usr/local/bin:...",
        "OLLAMA_HOST": "http://localhost:11434"
      },
      "description": "Remote exarp-go with Ollama and MLX (M4)"
    }
  }
}
```

### Multiple Remote Servers

```json
{
  "mcpServers": {
    "remote-exarp-go": {
      "command": "ssh",
      "args": ["davids-mac-mini.tailf62197.ts.net", "/Users/davidl/Projects/exarp-go/bin/exarp-go"],
      "description": "Remote exarp-go"
    },
    "remote-advisor": {
      "command": "ssh",
      "args": ["davids-mac-mini.tailf62197.ts.net", "/Users/davidl/Projects/devwisdom-go/run-advisor.sh"],
      "description": "Remote advisor"
    }
  }
}
```

## Additional Resources

- [Cursor Background Agents Documentation](https://docs.cursor.com/en/background-agents)
- [MCP Protocol Specification](https://modelcontextprotocol.io/)
- [SSH Configuration Guide](./TAILSCALE_SSH_SETUP.md)

