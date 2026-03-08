# OpenCode Setup Guide

**Last Updated**: 2026-03-08  
**Version**: v0.3.5  
**Audience**: OpenCode users integrating exarp-go

This guide provides step-by-step instructions for setting up exarp-go with OpenCode.

---

## Quick Start

### 1. Install exarp-go

```bash
# Clone the repository
git clone https://github.com/davidl71/exarp-go.git
cd exarp-go

# Build the binary
make go-build

# Verify installation
./bin/exarp-go -list
```

### 2. Create OpenCode Configuration

**For exarp-go development** (working in exarp-go repo):

```bash
# Create opencode.json in project root
cat > opencode.json <<'EOF'
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "exarp-go": {
      "type": "local",
      "command": ["/Users/davidl/Projects/mcp/exarp-go/bin/exarp-go"],
      "enabled": true,
      "environment": {
        "PROJECT_ROOT": "/Users/davidl/Projects/mcp/exarp-go",
        "EXARP_MIGRATIONS_DIR": "/Users/davidl/Projects/mcp/exarp-go/migrations",
        "EXARP_WATCH": "0"
      }
    }
  }
}
EOF
```

**For other projects** (using exarp-go in your project):

```bash
# Create opencode.json in your project root
cat > opencode.json <<'EOF'
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "exarp-go": {
      "type": "local",
      "command": ["/Users/davidl/go/bin/run_exarp_go.sh"],
      "enabled": true,
      "environment": {
        "PROJECT_ROOT": "/path/to/your/project",
        "EXARP_MIGRATIONS_DIR": "/Users/davidl/Projects/mcp/exarp-go/migrations",
        "EXARP_WATCH": "0"
      }
    }
  }
}
EOF

# Replace /path/to/your/project with actual path
sed -i '' 's|/path/to/your/project|'"$(pwd)"'|g' opencode.json
```

### 3. Test Configuration

```bash
# Verify MCP server starts
PROJECT_ROOT=$(pwd) ./bin/exarp-go -list

# Should see:
# Available tools (39 total):
#   task_workflow
#   task_analysis
#   ...
```

### 4. Start OpenCode

```bash
# Launch OpenCode with configuration
opencode --config opencode.json

# Or if opencode.json is in current directory:
opencode
```

---

## Configuration Options

### Direct Binary vs Wrapper Script

| Approach | When to Use | Command | Pros | Cons |
|----------|-------------|---------|------|------|
| **Direct Binary** | exarp-go development | `bin/exarp-go` | Fastest, simple | Requires local build |
| **Wrapper Script** | Other projects | `run_exarp_go.sh` | Auto-resolves binary | Slightly slower startup |

### Environment Variables

| Variable | Required? | Purpose | Default | Example |
|----------|-----------|---------|---------|---------|
| `PROJECT_ROOT` | ✅ Yes | Project workspace root | CWD | `/Users/me/myproject` |
| `EXARP_MIGRATIONS_DIR` | ⚠️ Recommended | Todo2 migrations | `$EXARP_GO_ROOT/migrations` | `/path/to/exarp-go/migrations` |
| `EXARP_WATCH` | ❌ No | Auto-reload on changes | `0` | `0` (off) or `1` (on) |
| `EXARP_GO_VERBOSE` | ❌ No | Verbose logging | `0` | `0` or `1` |

**Best Practices**:
- Always set `PROJECT_ROOT` explicitly
- Use absolute paths (no `~` or relative paths)
- Set `EXARP_WATCH=0` for production stability
- Enable `EXARP_GO_VERBOSE=1` for troubleshooting

---

## Installation Methods

### Method 1: From Source (Recommended)

```bash
# Clone repository
git clone https://github.com/davidl71/exarp-go.git
cd exarp-go

# Build binary
make go-build

# Install wrapper script (optional)
cp scripts/run_exarp_go.sh ~/go/bin/
chmod +x ~/go/bin/run_exarp_go.sh

# Verify
./bin/exarp-go -list
```

### Method 2: Pre-built Binary (Future)

```bash
# Download latest release
# (Not yet available - build from source for now)

# Install to PATH
mv exarp-go ~/go/bin/
chmod +x ~/go/bin/exarp-go
```

### Method 3: Go Install (Future)

```bash
# Install via go install
# (Not yet published to Go modules - build from source for now)
```

---

## Project-Specific Setup

### For exarp-go Repository

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "exarp-go": {
      "type": "local",
      "command": ["/Users/davidl/Projects/mcp/exarp-go/bin/exarp-go"],
      "enabled": true,
      "environment": {
        "PROJECT_ROOT": "/Users/davidl/Projects/mcp/exarp-go",
        "EXARP_MIGRATIONS_DIR": "/Users/davidl/Projects/mcp/exarp-go/migrations",
        "EXARP_WATCH": "0"
      }
    }
  }
}
```

**Why?**
- Uses local binary (fastest)
- PROJECT_ROOT points to exarp-go repo
- Migrations directory in local repo

### For Go Projects

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "exarp-go": {
      "type": "local",
      "command": ["/Users/davidl/go/bin/run_exarp_go.sh"],
      "enabled": true,
      "environment": {
        "PROJECT_ROOT": "/Users/davidl/Projects/my-go-app",
        "EXARP_MIGRATIONS_DIR": "/Users/davidl/Projects/mcp/exarp-go/migrations",
        "EXARP_WATCH": "0"
      }
    }
  }
}
```

**Why?**
- Uses wrapper script (auto-resolves exarp-go)
- PROJECT_ROOT points to your Go project
- Migrations from exarp-go installation

### For Multi-Language Projects

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "exarp-go": {
      "type": "local",
      "command": ["/Users/davidl/go/bin/run_exarp_go.sh"],
      "enabled": true,
      "environment": {
        "PROJECT_ROOT": "/Users/davidl/Projects/polyglot-app",
        "EXARP_MIGRATIONS_DIR": "/Users/davidl/Projects/mcp/exarp-go/migrations",
        "EXARP_WATCH": "0"
      }
    }
  }
}
```

**Why?**
- exarp-go supports Go, Python, Rust, Node, Shell, Ansible, Markdown
- Same configuration works for all languages

---

## Verification Checklist

After setup, verify everything works:

### 1. Binary Check

```bash
# Check binary exists and is executable
ls -la /Users/davidl/Projects/mcp/exarp-go/bin/exarp-go

# Should show: -rwxr-xr-x ... exarp-go
```

### 2. Tool Discovery

```bash
# List all tools
./bin/exarp-go -list | head -10

# Should show 39 tools
```

### 3. Environment Variables

```bash
# Test with environment
PROJECT_ROOT=$(pwd) ./bin/exarp-go -list | grep "39 total"

# Should show: Available tools (39 total):
```

### 4. Health Check

```bash
# Run health check
./bin/exarp-go -tool health -args '{"action":"server"}'

# Should return: {"status":"operational",...}
```

### 5. Task List

```bash
# List tasks
./bin/exarp-go -tool task_workflow -args '{"action":"sync","sub_action":"list","limit":5}'

# Should return JSON task list
```

### 6. OpenCode Integration

```bash
# Start OpenCode
opencode --config opencode.json

# In OpenCode, verify exarp-go tools are available
```

---

## Troubleshooting

### Issue: "exarp-go not found"

**Symptoms**: OpenCode can't start MCP server

**Solution**:
```bash
# Check binary exists
ls -la /Users/davidl/Projects/mcp/exarp-go/bin/exarp-go

# If missing, rebuild
cd /Users/davidl/Projects/mcp/exarp-go
make go-build

# Update opencode.json with correct path
```

### Issue: "PROJECT_ROOT not set"

**Symptoms**: Tools can't find project files

**Solution**:
```json
{
  "environment": {
    "PROJECT_ROOT": "/absolute/path/to/your/project"
  }
}
```

**Important**: Use absolute paths, not relative or `~`

### Issue: "Migrations directory not found"

**Symptoms**: Warning about missing migrations

**Solution**:
```bash
# Check migrations exist
ls -la /Users/davidl/Projects/mcp/exarp-go/migrations/

# Set correct path in opencode.json
"EXARP_MIGRATIONS_DIR": "/Users/davidl/Projects/mcp/exarp-go/migrations"
```

### Issue: Slow startup

**Symptoms**: OpenCode takes long to start

**Solution**:
1. Use direct binary instead of wrapper (faster)
2. Disable watch mode: `"EXARP_WATCH": "0"`
3. Check binary is built with optimizations: `make go-build`

### Issue: Tools not appearing

**Symptoms**: OpenCode doesn't see exarp-go tools

**Solution**:
```bash
# Verify MCP config is valid JSON
cat opencode.json | jq .

# Check server starts
./bin/exarp-go -list

# Restart OpenCode
```

### Issue: Permission denied

**Symptoms**: Can't execute exarp-go

**Solution**:
```bash
# Make binary executable
chmod +x /Users/davidl/Projects/mcp/exarp-go/bin/exarp-go

# If using wrapper
chmod +x /Users/davidl/go/bin/run_exarp_go.sh
```

---

## Advanced Configuration

### Multiple Projects

You can configure exarp-go for multiple projects:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "exarp-go-project1": {
      "type": "local",
      "command": ["/Users/davidl/go/bin/run_exarp_go.sh"],
      "enabled": true,
      "environment": {
        "PROJECT_ROOT": "/Users/davidl/Projects/project1"
      }
    },
    "exarp-go-project2": {
      "type": "local",
      "command": ["/Users/davidl/go/bin/run_exarp_go.sh"],
      "enabled": true,
      "environment": {
        "PROJECT_ROOT": "/Users/davidl/Projects/project2"
      }
    }
  }
}
```

### Development Mode

For exarp-go development with auto-reload:

```json
{
  "environment": {
    "PROJECT_ROOT": "/Users/davidl/Projects/mcp/exarp-go",
    "EXARP_WATCH": "1",
    "EXARP_GO_VERBOSE": "1"
  }
}
```

**Note**: This rebuilds on file changes - slower but convenient for development

### Custom Migrations

For projects with custom Todo2 schemas:

```json
{
  "environment": {
    "PROJECT_ROOT": "/Users/davidl/Projects/myproject",
    "EXARP_MIGRATIONS_DIR": "/Users/davidl/Projects/myproject/.todo2/migrations"
  }
}
```

---

## Integration with Other Tools

### Cursor + OpenCode

You can use both Cursor and OpenCode with exarp-go:

**Cursor** (`~/.cursor/mcp.json`):
```json
{
  "exarp-go": {
    "command": ["/Users/davidl/go/bin/run_exarp_go.sh"],
    "env": {
      "PROJECT_ROOT": "{{PROJECT_ROOT}}"
    }
  }
}
```

**OpenCode** (`opencode.json`):
```json
{
  "mcp": {
    "exarp-go": {
      "command": ["/Users/davidl/go/bin/run_exarp_go.sh"],
      "environment": {
        "PROJECT_ROOT": "/path/to/project"
      }
    }
  }
}
```

**Note**: Both can coexist; OpenCode's config takes precedence when running `opencode`

### GitHub Copilot

exarp-go and Copilot are complementary:
- **Copilot**: Code completion, inline suggestions
- **exarp-go**: Task management, project automation, health checks

No configuration needed - they work independently.

---

## Best Practices

### 1. Configuration Management

✅ **Do**:
- Keep `opencode.json` in project root
- Use version control for project-specific config
- Create `opencode.json.template` for team sharing
- Document custom environment variables

❌ **Don't**:
- Hardcode username in shared configs
- Use relative paths in `command`
- Commit sensitive data in config

### 2. Performance

✅ **Do**:
- Use direct binary for exarp-go development
- Set `EXARP_WATCH=0` for production
- Build with optimizations: `make go-build`
- Use stdio:// resources instead of tools when possible

❌ **Don't**:
- Use `EXARP_WATCH=1` in production
- Run unoptimized debug builds
- Spawn multiple MCP servers unnecessarily

### 3. Security

✅ **Do**:
- Use absolute paths (prevents injection)
- Validate `PROJECT_ROOT` points to correct project
- Keep exarp-go binary updated
- Review tool permissions

❌ **Don't**:
- Use shell expansion in `command` (e.g., `~/path`)
- Set `PROJECT_ROOT` to sensitive directories
- Run as root unnecessarily

---

## Examples

### Example 1: Go Web Application

```bash
cd ~/Projects/my-web-app
cat > opencode.json <<EOF
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "exarp-go": {
      "type": "local",
      "command": ["/Users/davidl/go/bin/run_exarp_go.sh"],
      "enabled": true,
      "environment": {
        "PROJECT_ROOT": "$(pwd)",
        "EXARP_MIGRATIONS_DIR": "/Users/davidl/Projects/mcp/exarp-go/migrations"
      }
    }
  }
}
EOF
opencode
```

### Example 2: Python Data Science Project

```bash
cd ~/Projects/ml-project
cat > opencode.json <<EOF
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "exarp-go": {
      "type": "local",
      "command": ["/Users/davidl/go/bin/run_exarp_go.sh"],
      "enabled": true,
      "environment": {
        "PROJECT_ROOT": "$(pwd)"
      }
    }
  }
}
EOF
opencode
```

### Example 3: Multi-Repo Monorepo

```bash
cd ~/Projects/monorepo
cat > opencode.json <<EOF
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "exarp-go": {
      "type": "local",
      "command": ["/Users/davidl/go/bin/run_exarp_go.sh"],
      "enabled": true,
      "environment": {
        "PROJECT_ROOT": "$(pwd)",
        "EXARP_WATCH": "0"
      }
    }
  }
}
EOF
opencode
```

---

## FAQ

### Q: Do I need to rebuild exarp-go for each project?

**A**: No! Build once, use everywhere with the wrapper script.

### Q: Can I use exarp-go without OpenCode?

**A**: Yes! exarp-go works standalone via CLI:
```bash
exarp-go task list
exarp-go -tool health -args '{"action":"server"}'
```

### Q: What's the difference between `command` and `cmd`?

**A**: OpenCode uses `command` (array). Some tools use `cmd` (string). Always use `command` for OpenCode.

### Q: Can I use exarp-go with VS Code?

**A**: Not directly. Use Cursor (VS Code fork) or OpenCode for MCP support.

### Q: How do I update exarp-go?

**A**:
```bash
cd /Users/davidl/Projects/mcp/exarp-go
git pull
make go-build
```

### Q: What if I don't have Go installed?

**A**: You need Go to build exarp-go. Install from https://go.dev/dl/

---

## Next Steps

1. **Read Documentation**:
   - `docs/AI_LLM_INTEGRATION.md` - AI/LLM features
   - `docs/EXARP_ABILITIES_AUDIT.md` - Complete tool reference
   - `docs/TASK_TOOLS_GUIDE.md` - Task management

2. **Try Examples**:
   - List tasks: `exarp-go task list`
   - Run health check: `exarp-go -tool health -args '{"action":"server"}'`
   - Generate with AI: `exarp-go -tool text_generate -args '{"provider":"auto","prompt":"..."}'`

3. **Join Community**:
   - GitHub Issues: Report bugs, request features
   - Discussions: Ask questions, share tips

---

## Support

- **Documentation**: `docs/` directory
- **Issues**: https://github.com/davidl71/exarp-go/issues
- **Template**: `opencode.json.template`
- **Validation**: `docs/OPENCODE_VALIDATION.md`

**For help**: Include your `opencode.json` (without sensitive data) and error output.
