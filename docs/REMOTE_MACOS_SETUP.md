# Remote macOS (Apple Silicon M4) Setup Guide

## Overview

This guide helps configure Cursor and MCP servers on a remote macOS machine (Apple Silicon M4).

## Remote Machine Details

- **Hostname:** davids-mac-mini.tailf62197.ts.net
- **OS:** macOS 26.3 (Sequoia)
- **Architecture:** Apple Silicon M4 (ARM64)
- **User:** davidl
- **Cursor:** Installed at `/Applications/Cursor.app`

## Prerequisites

### 1. Verify Development Tools

```bash
# Check Go installation
go version

# Check Python 3
python3 --version

# Check uvx (Python package manager)
uvx --version

# Check npx (Node.js)
npx --version
```

### 2. Verify Project Directories

Projects should be in `~/Projects/`:
- `exarp-go/` - Go-based MCP server
- `devwisdom-go/` - Advisor MCP server
- `project-management-automation/` - Coordinator tools

## Step 1: Build Required Binaries

### Build exarp-go (ARM64)

```bash
cd ~/Projects/exarp-go
go build -o bin/exarp-go ./cmd/server
chmod +x bin/exarp-go

# Verify
./bin/exarp-go --version
```

### Build devwisdom-go (ARM64)

```bash
cd ~/Projects/devwisdom-go
go build -o devwisdom ./cmd/server
chmod +x devwisdom

# Verify
./devwisdom --help
```

### Create devwisdom wrapper script

```bash
cd ~/Projects/devwisdom-go
cat > run-advisor.sh << 'EOF'
#!/bin/bash
cd "$(dirname "$0")" || exit 1
exec ./devwisdom "$@"
EOF
chmod +x run-advisor.sh
```

## Step 2: Configure Cursor MCP Servers

### Create/Update `.cursor/mcp.json`

For projects on the remote machine, create `.cursor/mcp.json` in each project workspace:

**Example for `exarp-go` project:**

```json
{
  "mcpServers": {
    "advisor": {
      "command": "/Users/davidl/Projects/devwisdom-go/run-advisor.sh",
      "args": [],
      "env": {
        "PROJECT_ROOT": "{{PROJECT_ROOT}}"
      },
      "description": "Crew Role: Advisor - DevWisdom Go MCP Server"
    },
    "exarp-go": {
      "command": "/Users/davidl/Projects/exarp-go/bin/exarp-go",
      "args": [],
      "env": {
        "PROJECT_ROOT": "{{PROJECT_ROOT}}"
      },
      "description": "Crew Role: Executor - Go-based MCP server"
    }
  }
}
```

**Note:** Update paths to match your actual project locations.

## Step 3: Verify MCP Server Paths

### Check Binary Paths

```bash
# Verify exarp-go
test -f ~/Projects/exarp-go/bin/exarp-go && echo "✅ exarp-go ready" || echo "❌ Build exarp-go"

# Verify devwisdom
test -f ~/Projects/devwisdom-go/devwisdom && echo "✅ devwisdom ready" || echo "❌ Build devwisdom"

# Verify wrapper script
test -f ~/Projects/devwisdom-go/run-advisor.sh && echo "✅ wrapper ready" || echo "❌ Create wrapper"
```

### Test Servers

```bash
# Test exarp-go
~/Projects/exarp-go/bin/exarp-go --help

# Test devwisdom
~/Projects/devwisdom-go/devwisdom --help
```

## Step 4: Architecture Considerations (Apple Silicon M4)

### Building for ARM64

All Go binaries compile natively for ARM64:

```bash
# Check architecture
uname -m
# Should show: arm64

# Go will build ARM64 binaries by default
go env GOARCH
# Should show: arm64
```

### Universal Binaries (if needed)

If you need universal binaries for compatibility:

```bash
# Build universal binary (ARM64 + x86_64)
go build -o bin/exarp-go ./cmd/server

# Or explicitly for ARM64
GOARCH=arm64 go build -o bin/exarp-go-arm64 ./cmd/server
```

## Step 5: Python Dependencies (if needed)

### Install uv (if not present)

```bash
# Install uv
curl -LsSf https://astral.sh/uv/install.sh | sh

# Add to PATH (add to ~/.zshrc)
export PATH="$HOME/.cargo/bin:$PATH"
```

### Verify Python Bridge Dependencies

If using Python bridge features:

```bash
cd ~/Projects/exarp-go
# Check if Python bridge scripts exist
ls -la bridge/*.py

# Test Python execution
python3 bridge/execute_tool.py --help
```

## Step 6: Configure Cursor

### 1. Open Cursor on Remote Machine

```bash
# Open Cursor
open -a Cursor

# Or from terminal
open -a Cursor ~/Projects/exarp-go
```

### 2. Verify MCP Configuration

1. Open Cursor Settings (Cmd+,)
2. Navigate to "Features" → "Model Context Protocol"
3. Verify servers are listed and show "Connected" status

### 3. Test MCP Servers

In Cursor chat, try:
- Using exarp-go tools: `@exarp-go report scorecard`
- Getting wisdom: `@advisor get wisdom`

## Troubleshooting

### Binary Not Found

**Issue:** Cursor can't find MCP server binary

**Solution:**
1. Verify absolute paths in `.cursor/mcp.json`
2. Check binary exists: `ls -la /Users/davidl/Projects/exarp-go/bin/exarp-go`
3. Check binary is executable: `chmod +x /path/to/binary`
4. Restart Cursor after changes

### Architecture Mismatch

**Issue:** Binary built for wrong architecture

**Solution:**
```bash
# Rebuild for ARM64
cd ~/Projects/exarp-go
go build -o bin/exarp-go ./cmd/server

# Verify architecture
file bin/exarp-go
# Should show: Mach-O 64-bit executable arm64
```

### Permission Denied

**Issue:** Can't execute binary or scripts

**Solution:**
```bash
# Make binaries executable
chmod +x ~/Projects/exarp-go/bin/exarp-go
chmod +x ~/Projects/devwisdom-go/devwisdom
chmod +x ~/Projects/devwisdom-go/run-advisor.sh

# Check Cursor has permissions
# System Settings → Privacy & Security → Full Disk Access
```

### MCP Server Not Loading

**Check:**
1. Restart Cursor completely (Cmd+Q)
2. Check Cursor MCP logs (Settings → Features → Model Context Protocol)
3. Verify JSON syntax in `.cursor/mcp.json`
4. Test binary manually from terminal

## Best Practices for Remote macOS

1. **Use Absolute Paths:** Always use full paths in MCP config (`/Users/davidl/Projects/...`)
2. **Verify Architecture:** Ensure binaries are ARM64 (native for M4)
3. **Test Locally:** Test binaries from terminal before configuring Cursor
4. **Wrapper Scripts:** Use wrapper scripts for servers that need specific working directories
5. **Environment Variables:** Use `{{PROJECT_ROOT}}` for workspace-aware paths

## Quick Setup Script

```bash
#!/bin/bash
# Quick setup script for remote macOS M4

set -e

echo "🚀 Setting up MCP servers on macOS M4..."

# Build exarp-go
cd ~/Projects/exarp-go
echo "📦 Building exarp-go..."
go build -o bin/exarp-go ./cmd/server
chmod +x bin/exarp-go

# Build devwisdom-go
cd ~/Projects/devwisdom-go
echo "📦 Building devwisdom..."
go build -o devwisdom ./cmd/server
chmod +x devwisdom

# Create wrapper
cat > run-advisor.sh << 'EOF'
#!/bin/bash
cd "$(dirname "$0")" || exit 1
exec ./devwisdom "$@"
EOF
chmod +x run-advisor.sh

echo "✅ Setup complete!"
echo "📝 Next: Configure .cursor/mcp.json in your workspace"
```

## Additional Resources

- [Cursor MCP Setup Guide](./CURSOR_MCP_SETUP.md)
- [Tailscale SSH Setup](./TAILSCALE_SSH_SETUP.md)
- [Apple Silicon Development](https://developer.apple.com/documentation/apple-silicon)

