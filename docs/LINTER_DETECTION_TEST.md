# Linter Detection Test Results

**Date:** 2026-01-07  
**Purpose:** Verify exarp-go can detect linters installed via Ansible

## Installation Status

### Installed Linters

| Linter | Status | Path | Version |
|--------|--------|------|---------|
| golangci-lint | ✅ Installed | `/Users/davidl/go/bin/golangci-lint` | v1.64.8 |
| shellcheck | ✅ Installed | `/opt/homebrew/bin/shellcheck` | Latest |
| shfmt | ✅ Installed | `/Users/davidl/go/bin/shfmt` | v3.12.0 |
| gomarklint | ✅ Installed | `/Users/davidl/go/bin/gomarklint` | Latest |
| markdownlint | ✅ Installed | `/Users/davidl/.nvm/versions/node/v20.19.5/bin/markdownlint` | 0.47.0 |
| cspell | ✅ Installed | `/Users/davidl/.nvm/versions/node/v20.19.5/bin/cspell` | 9.4.0 |

## exarp-go Detection

exarp-go uses `exec.LookPath()` to detect linters, which searches the system PATH.

### Detection Method

From `internal/tools/linting.go`:

```go
// Check if golangci-lint is available
if _, err := exec.LookPath("golangci-lint"); err != nil {
    // Linter not found
}
```

### PATH Requirements

For exarp-go to detect linters, they must be in the system PATH:

- **Go-installed linters** (golangci-lint, shfmt, gomarklint): Should be in `$GOPATH/bin` or `$HOME/go/bin`
- **System-installed linters** (shellcheck): Should be in standard system paths
- **npm-installed linters** (markdownlint, cspell): Should be in npm global bin path

### Testing Detection

Run the detection test:

```bash
cd /Users/davidl/Projects/exarp-go
go run /tmp/test-linter-detection.go
```

Or test via exarp-go:

```bash
# Test golangci-lint
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"lint","arguments":{"action":"run","linter":"golangci-lint","path":"cmd/server"}}}' | ./bin/exarp-go

# Test shellcheck
echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"lint","arguments":{"action":"run","linter":"shellcheck","path":"scripts/check-go-health.sh"}}}' | ./bin/exarp-go

# Test markdownlint
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"lint","arguments":{"action":"run","linter":"markdownlint","path":"README.md"}}}' | ./bin/exarp-go
```

## Ansible Installation

The Ansible playbook installs linters to standard locations:

- **Go linters**: `$GOPATH/bin` (typically `~/go/bin`)
- **System linters**: System package manager paths
- **npm linters**: npm global bin path

### Ensuring PATH Access

The Ansible playbook should ensure linters are in PATH:

1. **Go linters**: Add `$GOPATH/bin` to PATH in shell profile
2. **npm linters**: Add npm global bin to PATH
3. **System linters**: Already in standard PATH

## Recommendations

1. **Verify PATH**: Ensure all linter paths are in system PATH
2. **Test Detection**: Run detection test after Ansible installation
3. **Update Shell Profile**: Add Go and npm bin paths if needed
4. **Restart Shell**: Reload shell profile after PATH changes

---

**Last Updated:** 2026-01-07

