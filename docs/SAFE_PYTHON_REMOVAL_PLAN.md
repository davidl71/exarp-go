# Safe Python Code Removal Plan

**Date:** 2026-01-12  
**Status:** Ready for Execution  
**Based on:** `SAFE_PYTHON_REMOVAL_ANALYSIS.md`, `PYTHON_CODE_REMOVAL_ANALYSIS.md`, `SAFE_DELETION_PLAN.md`

## Executive Summary

**Total Python Files to Remove:** 5-7 files  
**Total Lines to Remove:** ~1,600 lines  
**Risk Level:** ✅ **LOW** - All identified code is either dead or unused

---

## ✅ Phase 1: Safe to Remove Immediately (100% Safe)

### 1. `bridge/prompt_tracking/` Directory
**Confidence:** ✅ **100% SAFE**

**Files:**
- `bridge/prompt_tracking/tool.py`
- `bridge/prompt_tracking/tracker.py`
- `bridge/prompt_tracking/__init__.py`

**Why:**
- Tool is fully native Go (`handlePromptTrackingNative`)
- Handler has NO Python fallback
- Bridge code returns error (lines 344-353 in `execute_tool.py`)
- Never executed

**Action:** Delete entire directory

---

### 2. `bridge/execute_tool.py` - Dead Code Blocks

**Lines to Remove:**
- Line 61: Comment about prompt_tracking (already removed)
- Lines 344-353: `prompt_tracking` routing block (returns error)
- Lines 363-371: `server_status` routing block (returns error)

**Why:**
- Both tools are fully native Go
- Bridge code just returns errors
- Never successfully executed

**Action:** Remove error-returning blocks

---

### 3. `mcp_stdio_tools/research_helpers.py` (476 lines)
**Confidence:** ✅ **100% SAFE**

**Why:**
- NOT imported by any bridge scripts
- NOT used by Go code
- Only referenced in documentation (historical)
- No active dependencies

**Action:** Delete immediately

---

## ⚠️ Phase 2: Verify Before Removing (95% Safe)

### 4. `mcp_stdio_tools/server.py` (1,114 lines)
**Confidence:** ⚠️ **VERIFY FIRST**

**Why Verify:**
- Old Python MCP server from before Go migration
- May still be referenced in docs/scripts (but not used)

**Verification Steps:**
1. ✅ Check if `bin/exarp-go` exists and works
2. ✅ Verify MCP config uses Go binary
3. ✅ Confirm no active references in code

**If Verified:**
- Delete `mcp_stdio_tools/server.py`
- Update references in docs (if any)

---

### 5. `mcp_stdio_tools/__init__.py` (4 lines)
**Confidence:** ✅ **95% SAFE** (after server.py removal)

**Why:**
- Only needed if `mcp_stdio_tools` package is imported
- If `server.py` is removed, package structure is unnecessary
- Bridge scripts don't import this package

**Action:** Delete after confirming `server.py` removal

---

## 📦 Phase 3: Dependency Cleanup (After Phase 2)

### 6. `pyproject.toml` Dependencies

**Dependencies to Remove (if server.py removed):**
- `mcp>=1.0.0` - Only needed for `server.py`
- `mlx>=0.20.0` - Only needed for `server.py`
- `mlx-lm>=0.20.0` - Only needed for `server.py`

**Keep:**
- Dev dependencies: `pytest`, `black`, `ruff` (needed for tests)

---

## ✅ Files to KEEP (Do NOT Delete)

### Bridge Scripts (3 files) - ACTIVE
- ✅ `bridge/execute_tool.py` (will be cleaned, not deleted)
- ✅ `bridge/get_prompt.py`
- ✅ `bridge/execute_resource.py`

### Tool Implementations - ACTIVE
- ✅ All files in `project_management_automation/tools/` (still used by bridge)
- ✅ Even tools with native Go may have Python fallbacks

### Test Files - INTENTIONAL
- ✅ All files in `tests/` directory

---

## Removal Order

### Step 1: Remove Dead Bridge Code (Safest)
```bash
# 1. Remove prompt_tracking directory
rm -rf bridge/prompt_tracking/

# 2. Remove dead routing blocks from execute_tool.py
# (Lines 344-353 and 363-371)
```

### Step 2: Remove Unused Files
```bash
# 1. Remove research helpers
rm mcp_stdio_tools/research_helpers.py

# 2. Verify Go server works
./bin/exarp-go --help

# 3. If verified, remove old Python server
rm mcp_stdio_tools/server.py
rm mcp_stdio_tools/__init__.py
```

### Step 3: Clean Dependencies
```bash
# Update pyproject.toml to remove mcp, mlx, mlx-lm
# (Keep dev dependencies)
```

---

## Impact Summary

**Files to Remove:**
- `bridge/prompt_tracking/` (3 files)
- `mcp_stdio_tools/research_helpers.py` (1 file)
- `mcp_stdio_tools/server.py` (1 file, after verification)
- `mcp_stdio_tools/__init__.py` (1 file, after verification)
- Dead code blocks in `bridge/execute_tool.py` (~20 lines)

**Total:** 5-7 files, ~1,600 lines

**Risk Level:** ✅ **LOW**  
**Impact:** None (code is never called)  
**Rollback:** Easy (can restore from git)

---

## Verification Checklist

Before removing `mcp_stdio_tools/server.py`:
- [x] `bin/exarp-go` exists and is executable
- [x] Go server registers all tools/prompts/resources
- [x] MCP config uses Go binary (not Python server)
- [x] No active code references to `mcp_stdio_tools.server`

---

## Next Steps

1. ✅ Execute Phase 1 removals (100% safe)
2. ⚠️ Verify and execute Phase 2 removals
3. 📦 Clean up dependencies in Phase 3
4. 📝 Update documentation if needed
