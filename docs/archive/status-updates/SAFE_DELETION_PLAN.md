# Safe Deletion Plan for Leftover Python Code

**Date:** 2026-01-07  
**Status:** Ready for Review

## Summary

Based on the Python code audit, here's what can be safely deleted:

## ✅ SAFE TO DELETE IMMEDIATELY

### 1. `mcp_stdio_tools/research_helpers.py` (476 lines)
**Confidence:** ✅ **100% SAFE**

**Why:**
- NOT imported by any bridge scripts
- NOT used by Go code
- Only referenced in documentation (historical)
- No active dependencies

**Action:** Delete immediately

### 2. `mcp_stdio_tools/__init__.py` (4 lines)
**Confidence:** ✅ **95% SAFE** (after server.py removal)

**Why:**
- Only needed if `mcp_stdio_tools` package is imported
- If `server.py` is removed, this package structure is unnecessary
- Bridge scripts don't import this package

**Action:** Delete after confirming `server.py` removal

## ⚠️ VERIFY BEFORE DELETING

### 3. `mcp_stdio_tools/server.py` (1,114 lines)
**Confidence:** ⚠️ **VERIFY FIRST**

**Why Verify:**
- Still referenced in multiple files (Makefile, run_server.sh, dev.sh, README.md)
- May still be used if Go server isn't fully active
- Need to confirm Go server (`bin/exarp-go`) is working

**Verification Steps:**
1. ✅ Check if `bin/exarp-go` exists and is executable
2. ✅ Verify Go server registers all 24 tools, 8 prompts, 6 resources
3. ✅ Confirm MCP config uses Go binary, not Python server
4. ✅ Test that tools work via Go server

**If Verified:**
- Delete `mcp_stdio_tools/server.py`
- Update all references in:
  - `README.md` (line 69)
  - `Makefile` (line 6, and all SERVER_MODULE references)
  - `run_server.sh` (line 15)
  - `dev.sh` (line 28, and all SERVER_MODULE references)

## 📦 CONDITIONAL CLEANUP

### 4. `pyproject.toml` Dependencies
**Confidence:** ⚠️ **CONDITIONAL**

**Current Dependencies:**
- `mcp>=1.0.0` - Only needed for `server.py`
- `mlx>=0.20.0` - Only needed for `server.py`
- `mlx-lm>=0.20.0` - Only needed for `server.py`

**Keep:**
- Dev dependencies: `pytest`, `black`, `ruff` (needed for tests)

**Action After server.py Removal:**
```toml
[project]
name = "mcp-stdio-tools"
version = "0.1.0"
description = "Python bridge scripts for Go MCP server"
requires-python = ">=3.10"
# No runtime dependencies - bridge scripts use standard library only

[project.optional-dependencies]
dev = [
    "pytest>=7.0.0",
    "black>=23.0.0",
    "ruff>=0.1.0",
]
```

## 🗑️ Deletion Order (Safest First)

### Phase 1: Definitely Safe (Do Now)
```bash
# 1. Delete unused research helpers
rm mcp_stdio_tools/research_helpers.py

# 2. Verify no imports break
python3 -c "import sys; sys.path.insert(0, '.'); from bridge import execute_tool; print('✅ Bridge scripts OK')"
```

### Phase 2: After Verification (Verify Go Server First)
```bash
# 1. Verify Go server works
./bin/exarp-go --help  # or test MCP connection

# 2. If Go server works, delete Python server
rm mcp_stdio_tools/server.py
rm mcp_stdio_tools/__init__.py

# 3. Update all references (see below)
```

### Phase 3: Update References
After deleting `server.py`, update:

**README.md:**
```markdown
## Running

```bash
./bin/exarp-go
```
```

**Makefile:**
```makefile
# Remove SERVER_MODULE references
# Update run target:
run: ## Run the MCP server
	@echo "$(BLUE)Running $(PROJECT_NAME) server...$(NC)"
	@./bin/exarp-go
```

**run_server.sh:**
```bash
#!/bin/bash
# Wrapper script to run exarp-go server
exec "$(dirname "$0")/bin/exarp-go"
```

**dev.sh:**
- Remove all `SERVER_MODULE` references
- Update to use `./bin/exarp-go` instead

**pyproject.toml:**
- Remove `mcp`, `mlx`, `mlx-lm` dependencies
- Keep dev dependencies

## ✅ Files to KEEP (Do NOT Delete)

### Bridge Scripts (3 files) - ACTIVE
- ✅ `bridge/execute_tool.py`
- ✅ `bridge/get_prompt.py`
- ✅ `bridge/execute_resource.py`

### Test Files (8 files) - INTENTIONAL
- ✅ All files in `tests/` directory

### Configuration
- ✅ `pyproject.toml` (simplified, but keep)
- ✅ `go.mod` / `go.sum` (Go dependencies)

## 📋 Pre-Deletion Checklist

Before deleting `server.py`:

- [ ] Verify `bin/exarp-go` exists and is executable
- [ ] Test Go server: `./bin/exarp-go` responds to MCP requests
- [ ] Verify all 24 tools work via Go server
- [ ] Verify all 8 prompts work via Go server
- [ ] Verify all 6 resources work via Go server
- [ ] Check MCP config uses Go binary (not Python server)
- [ ] Backup old files (optional): `mkdir -p docs/archive && mv mcp_stdio_tools/server.py docs/archive/`

## 🚨 Rollback Plan

If deletion causes issues:

1. Restore from git: `git checkout mcp_stdio_tools/server.py`
2. Restore references in updated files
3. Revert `pyproject.toml` changes

## Summary

**Immediate Deletions (100% Safe):**
- `mcp_stdio_tools/research_helpers.py` ✅

**After Verification:**
- `mcp_stdio_tools/server.py` ⚠️
- `mcp_stdio_tools/__init__.py` ⚠️
- Simplify `pyproject.toml` ⚠️

**Keep:**
- All bridge scripts ✅
- All test files ✅
- Simplified `pyproject.toml` ✅

