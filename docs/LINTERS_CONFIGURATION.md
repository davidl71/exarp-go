# Linters Configuration

**Date:** 2026-01-09  
**Status:** ✅ Configured

## Tools Configured

### 1. golangci-lint ✅

**Status:** Configured and working  
**Configuration:** `.golangci.yml`

**Enabled Linters:**
- `errcheck` - Check for unchecked errors
- `gosimple` - Simplify code
- `govet` - Go vet
- `ineffassign` - Detect unused assignments
- `staticcheck` - Static analysis
- `unused` - Unused code detection
- `gofmt` - Format checking
- `goimports` - Import formatting
- `misspell` - Spelling checker
- `gocritic` - Go code critic
- `revive` - Fast, configurable linter
- `gosec` - Security issues
- And 30+ more linters

**Usage:**
```bash
# Run on entire project
golangci-lint run ./...

# Run on specific package
golangci-lint run ./internal/database

# Run with timeout
golangci-lint run --timeout 30s ./...
```

### 2. govulncheck ✅

**Status:** Installed and working  
**Installation:** `go install golang.org/x/vuln/cmd/govulncheck@latest`

**Usage:**
```bash
# Check entire project
govulncheck ./...

# Check specific package
govulncheck ./internal/database
```

**Current Status:** ✅ No vulnerabilities found

## Ansible Integration

### Development Environment

**Location:** `ansible/inventories/development/group_vars/all.yml`

**Linters configured:**
- golangci-lint
- govulncheck (newly added)
- shellcheck
- shfmt
- gomarklint
- markdownlint-cli
- cspell

**Installation:**
```bash
cd ansible
ansible-playbook playbooks/development.yml --tags linters
```

### Production Environment

**Location:** `ansible/inventories/production/group_vars/all.yml`

**Linters:** Not installed (production doesn't need linters)

## Configuration Files

1. **`.golangci.yml`** - golangci-lint configuration
   - Timeout: 5 minutes
   - Includes test files
   - Skips vendor, .git, bin, .exarp, .todo2 directories
   - 40+ linters enabled

2. **`ansible/roles/linters/tasks/main.yml`** - Ansible linter installation
   - Added govulncheck installation task
   - Added govulncheck to verification list

## Known Issues

### golangci-lint Warnings

1. **Line length** - Some lines exceed 140 characters
   - Action: Consider breaking long lines or adjusting limit

2. **Error handling** - Some error returns not checked
   - Action: Check all error returns (e.g., `tx.Rollback()`)

3. **Dynamic errors** - Some errors use `fmt.Errorf` instead of wrapped static errors
   - Action: Consider using `errors.New` with `fmt.Errorf` wrapping

### Fixed Issues

- ✅ Removed deprecated linter names (goerr113 → err113, gomnd → mnd, etc.)
- ✅ Updated deprecated configuration options
- ✅ Fixed `tx.Rollback()` error handling
- ✅ Updated Ansible to include govulncheck

## Next Steps

1. Fix remaining golangci-lint warnings
2. Add pre-commit hook to run linters
3. Integrate into CI/CD pipeline
4. Set up automated linting in development workflow

## References

- [golangci-lint Documentation](https://golangci-lint.run/)
- [govulncheck Documentation](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck)
- [Ansible Linters Role](../ansible/roles/linters/)

