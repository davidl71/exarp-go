# Python Code Audit Report

**Date:** 2026-01-07  
**Purpose:** Identify leftover Python code after Go migration  
**Status:** ✅ Complete

## Executive Summary

**Total Python Files Found:** 14 files

- ✅ **Intentional (11 files):** Bridge scripts and test files - KEEP
- ⚠️ **Potentially Leftover (3 files):** Old server and unused helpers - REVIEW

## Detailed Findings

### ✅ Intentional Python Files (KEEP)

#### Bridge Scripts (3 files)
These files are actively used by the Go server via `internal/bridge/python.go`:

1. **`bridge/execute_tool.py`**
   - Purpose: Executes Python tools from Go server
   - Dependencies: Standard library only (json, sys, os, pathlib)
   - Status: ✅ Active - Used by Go handlers

2. **`bridge/get_prompt.py`**
   - Purpose: Retrieves prompt templates from Python
   - Dependencies: Standard library only
   - Status: ✅ Active - Used by Go prompt registry

3. **`bridge/execute_resource.py`**
   - Purpose: Executes Python resource handlers
   - Dependencies: Standard library only
   - Status: ✅ Active - Used by Go resource handlers

#### Test Files (8 files)
All test files are intentional and should be kept:

- `tests/integration/bridge/test_bridge_integration.py`
- `tests/integration/mcp/test_mcp_server.py`
- `tests/integration/mcp/test_server_startup.py`
- `tests/unit/python/test_execute_resource.py`
- `tests/unit/python/test_get_prompt.py`
- `tests/unit/python/test_execute_tool.py`
- `tests/fixtures/test_helpers.py`
- `tests/fixtures/mock_python.py`

### ⚠️ Potentially Leftover Files (REVIEW)

#### 1. `mcp_stdio_tools/server.py` (1,114 lines)
**Status:** ⚠️ **LIKELY LEFTOVER**

**Description:**
- Old Python MCP server from before Go migration
- Contains full MCP server implementation with 24 tools, 8 prompts, 6 resources
- Uses `mcp.server.stdio` framework

**Current References:**
- `README.md` line 69: `uv run python -m mcp_stdio_tools.server`
- `Makefile` line 6: `SERVER_MODULE := mcp_stdio_tools.server`
- `run_server.sh` line 15: `exec uv run python -m mcp_stdio_tools.server`
- `dev.sh` line 28: `SERVER_MODULE="mcp_stdio_tools.server"`
- Various documentation files (historical references)

**Analysis:**
- The Go server (`bin/exarp-go`) is now the main server
- Tools are executed via bridge scripts, not this old server
- This file appears to be leftover from pre-migration

**Recommendation:**
1. Verify Go migration is complete and all functionality is working
2. If confirmed, remove this file
3. Update all references in `README.md`, `Makefile`, `run_server.sh`, `dev.sh`
4. Consider archiving in `docs/archive/` if historical reference is needed

#### 2. `mcp_stdio_tools/research_helpers.py` (476 lines)
**Status:** ⚠️ **LIKELY LEFTOVER**

**Description:**
- Helper module for parallel research execution
- Contains functions for CodeLlama, Context7, and Tractatus integration
- Designed for research phase of migration

**Current References:**
- Only in documentation files (historical references)
- NOT imported by bridge scripts
- NOT used by Go code
- NOT referenced in any active code

**Analysis:**
- Appears to be unused code from migration planning phase
- No active usage found in codebase

**Recommendation:**
1. Verify if this is used by any external tools or scripts
2. If unused, remove this file
3. Consider archiving if it might be useful for future research tasks

#### 3. `mcp_stdio_tools/__init__.py` (4 lines)
**Status:** ⚠️ **CONDITIONAL**

**Description:**
- Package initialization file
- Defines `__version__ = "0.1.0"`

**Analysis:**
- Only needed if `mcp_stdio_tools` package is imported
- If `server.py` is removed, this becomes unnecessary

**Recommendation:**
- Remove if `server.py` is removed
- Keep if package structure is needed for other purposes

### 📦 Configuration Files

#### `pyproject.toml`
**Status:** ⚠️ **CONDITIONAL**

**Current Dependencies:**
- `mcp>=1.0.0` - MCP framework (only needed for `server.py`)
- `mlx>=0.20.0` - MLX library (only needed for `server.py`)
- `mlx-lm>=0.20.0` - MLX language models (only needed for `server.py`)

**Analysis:**
- Bridge scripts only use standard library (json, sys, os, pathlib)
- Dependencies are only needed if `server.py` is still used
- Test files may need pytest, but that's in dev dependencies

**Recommendation:**
1. If `server.py` is removed, simplify `pyproject.toml`:
   - Remove `mcp`, `mlx`, `mlx-lm` from dependencies
   - Keep dev dependencies (pytest, black, ruff) if tests are kept
2. If `server.py` is kept, maintain current dependencies

## Go Code Analysis

**Python References in Go Files:** 65 matches

All references are **intentional**:
- `internal/bridge/python.go` - Bridge functions that call Python scripts
- `internal/tools/handlers.go` - Tool handlers that use Python bridge
- `internal/resources/handlers.go` - Resource handlers that use Python bridge
- `internal/prompts/registry.go` - Prompt registry that uses Python bridge

**Conclusion:** No leftover Python code in Go files. All references are part of the intentional bridge architecture.

## Recommendations

### Immediate Actions

1. **Verify Go Migration Status**
   - Confirm all 24 tools work via Go server
   - Verify all 8 prompts work via Go server
   - Verify all 6 resources work via Go server

2. **If Migration Complete:**
   - ✅ Remove `mcp_stdio_tools/server.py`
   - ✅ Remove `mcp_stdio_tools/research_helpers.py`
   - ✅ Remove `mcp_stdio_tools/__init__.py` (if not needed)
   - ✅ Update `README.md` to remove Python server instructions
   - ✅ Update `Makefile` to remove Python server references
   - ✅ Update `run_server.sh` to point to Go binary
   - ✅ Update `dev.sh` to remove Python server references
   - ✅ Simplify `pyproject.toml` (remove unused dependencies)

3. **If Migration Incomplete:**
   - Keep files until migration is complete
   - Document which features still need migration

### Files to Keep

- ✅ All bridge scripts (`bridge/*.py`)
- ✅ All test files (`tests/**/*.py`)
- ✅ `pyproject.toml` (simplified if server.py removed)

### Files to Remove (if migration complete)

- ❌ `mcp_stdio_tools/server.py`
- ❌ `mcp_stdio_tools/research_helpers.py`
- ❌ `mcp_stdio_tools/__init__.py` (conditional)

## Next Steps

1. Review this audit report
2. Verify Go migration completion status
3. Make decision on leftover files
4. Update documentation and scripts accordingly
5. Remove or archive leftover files

## Notes

- Bridge scripts are intentionally minimal (standard library only)
- Test files are all intentional and should be kept
- Old Python server appears to be complete leftover from migration
- Research helpers appear unused but may have been useful during migration planning

