# Cursor Remote - SSH Setup with Tailscale

## Overview

Using Cursor's Remote - SSH extension with Tailscale is an **excellent solution**! This allows you to:
- Connect Cursor directly to your remote macOS M4 machine
- Run Cursor on the remote machine (via SSH)
- Access MCP servers that are "local" to the remote machine
- Use remote MLX, Ollama, and CodeLlama without SSH-based MCP overhead

## Advantages Over SSH-Based MCP Servers

✅ **Better Performance:**
- MCP servers run locally on remote machine (no SSH overhead per call)
- Direct stdio communication
- Lower latency

✅ **Full Remote Environment:**
- Access remote filesystem directly
- Run commands in remote environment
- Use remote resources seamlessly

✅ **Simpler Configuration:**
- MCP servers use normal local paths
- No SSH command wrapping needed
- Standard MCP configuration works

✅ **Better Integration:**
- Extensions work in remote environment
- Terminal runs on remote machine
- Full remote development experience

## Setup Instructions

### Step 1: Configure SSH for Remote - SSH Extension

**Your SSH config already includes:**
```ssh-config
Host davids-mac-mini.tailf62197.ts.net
    User davidl
    # Tailscale SSH - Accept host keys automatically
```

**Optional: Add Short Alias**

For easier connection, you can add a short alias:

```ssh-config
Host davids-mac-mini
    HostName davids-mac-mini.tailf62197.ts.net
    User davidl
    StrictHostKeyChecking accept-new
    UserKnownHostsFile ~/.ssh/known_hosts
    ServerAliveInterval 60
    ServerAliveCountMax 3
```

Then you can connect to just `davids-mac-mini` instead of the full hostname.

### Step 2: Connect via Remote - SSH

1. **Open Command Palette:**
   - Press `F1` or `Cmd+Shift+P` (macOS) / `Ctrl+Shift+P` (Windows/Linux)

2. **Connect to Remote:**
   - Type: "Remote-SSH: Connect to Host"
   - Select: `davids-mac-mini` (or `davids-mac-mini.tailf62197.ts.net`)

3. **Wait for Connection:**
   - Cursor will connect via SSH
   - It will install Cursor Server on remote machine (first time only)
   - New window opens connected to remote

4. **Open Remote Workspace:**
   - Open folder: `/Users/davidl/Projects/exarp-go`
   - Or any other project folder

### Step 3: Verify Remote Connection

**Check you're on remote:**
- Bottom-left corner shows: `SSH: davids-mac-mini`
- Terminal shows remote machine hostname
- Files are from remote machine

### Step 4: MCP Servers Work Automatically!

Once connected via Remote - SSH:
- MCP servers configured on remote machine work normally
- `.cursor/mcp.json` on remote machine is used
- All MCP servers run locally on remote machine
- No SSH command wrapping needed!

**Example Remote MCP Config** (already configured on remote):
```json
{
  "mcpServers": {
    "advisor": {
      "command": "/Users/davidl/Projects/devwisdom-go/run-advisor.sh",
      "args": [],
      "env": {
        "PROJECT_ROOT": "{{PROJECT_ROOT}}"
      }
    },
    "exarp-go": {
      "command": "/Users/davidl/Projects/exarp-go/bin/exarp-go",
      "args": [],
      "env": {
        "PROJECT_ROOT": "{{PROJECT_ROOT}}",
        "OLLAMA_HOST": "http://localhost:11434",
        "PATH": "/opt/homebrew/bin:..."
      }
    }
  }
}
```

## Benefits for Your Setup

### 1. Native Remote MCP Servers
- ✅ All MCP servers run on remote macOS M4
- ✅ Full access to MLX, Ollama, CodeLlama
- ✅ No SSH overhead for each MCP call
- ✅ Standard stdio communication

### 2. Remote Hardware Access
- ✅ Apple Silicon M4 (ARM64)
- ✅ MLX optimized for M4
- ✅ Ollama with CodeLlama models
- ✅ All resources local to remote machine

### 3. Workspace Management
- ✅ Open remote project folders directly
- ✅ Edit files on remote machine
- ✅ Run commands in remote environment
- ✅ Full development environment on remote

## Remote - SSH vs SSH-Based MCP

### Remote - SSH (Recommended) ✅
- **Performance:** Excellent (local MCP servers)
- **Setup:** One-time SSH connection
- **Overhead:** Minimal (only SSH for connection)
- **Experience:** Full remote development
- **MCP Calls:** Direct stdio (fast)

### SSH-Based MCP (Alternative)
- **Performance:** Good (SSH per MCP call)
- **Setup:** SSH command in each MCP config
- **Overhead:** SSH for each tool call
- **Experience:** Local Cursor, remote tools
- **MCP Calls:** SSH wrapper (some latency)

## Configuration Examples

### SSH Config for Tailscale

**Recommended SSH Config:**
```ssh-config
# Tailscale hosts
Host davids-mac-mini
    HostName davids-mac-mini.tailf62197.ts.net
    User davidl
    StrictHostKeyChecking accept-new
    UserKnownHostsFile ~/.ssh/known_hosts
    # Optional: Keep connections alive
    ServerAliveInterval 60
    ServerAliveCountMax 3

# All Tailscale hosts
Host *.ts.net
    StrictHostKeyChecking accept-new
    UserKnownHostsFile ~/.ssh/known_hosts
```

### Remote Workspace Configuration

**When connected via Remote - SSH:**
- Open folder: `/Users/davidl/Projects/exarp-go`
- MCP config at: `/Users/davidl/Projects/exarp-go/.cursor/mcp.json`
- All paths are remote paths (no SSH commands needed)

## Troubleshooting

### Remote - SSH Connection Issues

**Problem:** Cannot connect to remote host

**Solution:**
1. Verify SSH works: `ssh davids-mac-mini.tailf62197.ts.net`
2. Check Tailscale connectivity: `ping 100.117.185.32`
3. Verify SSH keys are set up
4. Check Remote - SSH extension is installed

### Cursor Server Installation Failed

**Problem:** Cursor Server won't install on remote

**Solution:**
1. Check remote machine has sufficient space
2. Verify write permissions: `ls -la ~/.cursor-server`
3. Check remote machine connectivity
4. Try manual installation if needed

### MCP Servers Not Loading

**Problem:** MCP servers don't appear when connected remotely

**Solution:**
1. Verify `.cursor/mcp.json` exists on remote machine
2. Check paths in MCP config are absolute (remote paths)
3. Verify binaries exist on remote: `ls -la /Users/davidl/Projects/exarp-go/bin/exarp-go`
4. Restart Cursor after connecting remotely

### Remote Terminal Not Working

**Problem:** Terminal doesn't open or commands fail

**Solution:**
1. Check shell is available: `ssh davids-mac-mini.tailf62197.ts.net "echo $SHELL"`
2. Verify PATH in remote environment
3. Check remote machine permissions

## Best Practices

### 1. Use SSH Config for Easy Access
- Add all Tailscale hosts to `~/.ssh/config`
- Use short hostnames (e.g., `davids-mac-mini`)
- Enable connection keep-alive

### 2. Organize Remote Workspaces
- Create workspace folders on remote
- Use consistent project structure
- Document remote paths

### 3. Sync Settings (Optional)
- Use Settings Sync in Cursor
- Share keybindings and extensions
- Sync across local and remote

### 4. Remote Extensions
- Some extensions work in remote mode
- Others need remote installation
- Check extension compatibility

## Quick Start Checklist

- [ ] Install Remote - SSH extension in Cursor
- [ ] Configure SSH config (`~/.ssh/config`)
- [ ] Test SSH connection: `ssh davids-mac-mini.tailf62197.ts.net`
- [ ] Connect via Remote - SSH (F1 → "Remote-SSH: Connect to Host")
- [ ] Open remote workspace: `/Users/davidl/Projects/exarp-go`
- [ ] Verify MCP servers work (Settings → Features → Model Context Protocol)
- [ ] Test remote MLX/Ollama: `@exarp-go mlx status`

## Advantages Summary

**Remote - SSH + Tailscale = Perfect Combination!**

✅ **Secure:** Tailscale provides encrypted mesh network  
✅ **Fast:** Direct connection, no VPN overhead  
✅ **Native:** MCP servers run locally on remote  
✅ **Powerful:** Full access to remote hardware (M4, MLX, Ollama)  
✅ **Seamless:** Cursor runs as if on remote machine  

This is the **best approach** for using remote MLX/Ollama with Cursor!

## Additional Resources

- [Remote - SSH Extension Documentation](https://code.visualstudio.com/docs/remote/ssh)
- [Tailscale SSH Guide](./TAILSCALE_SSH_SETUP.md)
- [Remote MCP Servers Guide](./REMOTE_MCP_SERVERS.md)

