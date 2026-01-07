# Ansible Linter Installation Summary

**Date:** 2026-01-07  
**Status:** ✅ Complete

## Installation Results

### ✅ All Linters Installed

| Linter | Method | Location | Status |
|--------|--------|---------|--------|
| golangci-lint | Already installed | `$HOME/go/bin` | ✅ |
| shellcheck | Already installed | `/opt/homebrew/bin` | ✅ |
| shfmt | `go install` | `$HOME/go/bin` | ✅ **Newly Installed** |
| gomarklint | Already installed | `$HOME/go/bin` | ✅ |
| markdownlint | `npm install -g` | `~/.nvm/versions/node/*/bin` | ✅ **Newly Installed** |
| cspell | `npm install -g` | `~/.nvm/versions/node/*/bin` | ✅ **Newly Installed** |

## Detection Verification

### Using `exec.LookPath()` (Same as exarp-go)

All linters are detectable:

```
✅ golangci-lint: /Users/davidl/go/bin/golangci-lint
✅ shellcheck: /opt/homebrew/bin/shellcheck
✅ shfmt: /Users/davidl/go/bin/shfmt
✅ gomarklint: /Users/davidl/go/bin/gomarklint
✅ markdownlint: /Users/davidl/.nvm/versions/node/v20.19.5/bin/markdownlint
✅ cspell: /Users/davidl/.nvm/versions/node/v20.19.5/bin/cspell
```

## exarp-go Detection

exarp-go uses `exec.LookPath()` which searches the system PATH.

### PATH Configuration

For exarp-go to detect linters, PATH must include:

1. **Go bin**: `$HOME/go/bin` or `$GOPATH/bin`
2. **npm bin**: `$HOME/.nvm/versions/node/*/bin` or npm global bin
3. **System paths**: Already included

### MCP Configuration Update

Update `.cursor/mcp.json` to include PATH:

```json
{
  "mcpServers": {
    "exarp-go": {
      "command": "/Users/davidl/Projects/exarp-go/bin/exarp-go",
      "env": {
        "PROJECT_ROOT": "{{PROJECT_ROOT}}",
        "PATH": "/Users/davidl/go/bin:/Users/davidl/.nvm/versions/node/v20.19.5/bin:{{PATH}}"
      }
    }
  }
}
```

**Note:** Replace paths with your actual Go and npm bin paths.

## Ansible Playbook Updates

### ✅ Completed

1. ✅ Fixed Ansible callback configuration
2. ✅ Added npm bin path to shell profiles
3. ✅ Installed missing linters (shfmt, markdownlint, cspell)

### 📋 Recommended

1. Update `.cursor/mcp.json` with PATH (see above)
2. Restart Cursor to apply changes
3. Test linter detection via exarp-go

## Testing

After updating MCP config and restarting Cursor:

```bash
# In Cursor chat, test linters:
@gotoHuman list-forms  # Test gotoHuman
@exarp-go lint  # Test lint tool
```

Or via CLI:

```bash
export PATH="$HOME/go/bin:$HOME/.nvm/versions/node/v20.19.5/bin:$PATH"
./bin/exarp-go -tool lint -args '{"action":"run","linter":"golangci-lint","path":"cmd/server"}'
```

## Summary

✅ **All linters installed**  
✅ **All linters detectable** (when PATH is configured)  
✅ **Ansible playbook updated** (includes PATH setup)  
⚠️ **MCP config needs PATH update** (for Cursor integration)

---

**Last Updated:** 2026-01-07

