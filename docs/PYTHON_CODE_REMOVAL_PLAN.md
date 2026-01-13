# Python Code Removal Plan

**Date:** 2026-01-12  
**Status:** 📋 Planning  
**Related Task:** T-1768223765685

---

## Executive Summary

This document provides a detailed removal plan for Python code that can be safely removed after the Go migration. The analysis identified **5 fully native Go tools** whose Python implementations are no longer called by the bridge, plus one stub implementation.

**Total Files to Remove:** 6 Python function implementations  
**Total Lines to Remove:** ~500-600 lines (estimated)  
**Risk Level:** 🟢 **LOW** - These functions are not imported by bridge scripts

---

## Removal Categories

### 🟢 **Category 1: Fully Native Go Tools (LOW RISK)**

These tools are **fully implemented in native Go** and their Python implementations are **NOT imported** by `bridge/execute_tool.py`. The Python code is marked as "LEGACY: Removed from bridge, kept for testing/backup" but is never actually called.

#### 1. `tool_catalog` Tool

**Status:** ✅ Fully native Go (`internal/tools/tool_catalog.go`)  
**Python Location:** `project_management_automation/tools/consolidated_workflow.py`  
**Lines:** ~45 lines (lines 198-242)  
**Bridge Status:** ❌ NOT imported (commented out in `bridge/execute_tool.py` line 29)

**Removal Action:**
```python
# Remove function definition (lines 198-242)
def tool_catalog(...) -> str:
    # ... entire function ...
```

**Verification:**
```bash
# Verify not imported in bridge
grep -n "tool_catalog" bridge/execute_tool.py
# Should show only comment: "# Note: tool_catalog, workflow_mode, git_tools, and infer_session_mode removed"

# Verify Go implementation exists
ls -la internal/tools/tool_catalog.go
```

**Impact:** None - function is never called

---

#### 2. `workflow_mode` Tool

**Status:** ✅ Fully native Go (`internal/tools/workflow_mode.go`)  
**Python Location:** `project_management_automation/tools/consolidated_workflow.py`  
**Lines:** ~45 lines (lines 50-94)  
**Bridge Status:** ❌ NOT imported (commented out in `bridge/execute_tool.py` line 29)

**Removal Action:**
```python
# Remove function definition (lines 50-94)
def workflow_mode(...) -> str:
    # ... entire function ...
```

**Verification:**
```bash
# Verify not imported in bridge
grep -n "workflow_mode" bridge/execute_tool.py
# Should show only comment

# Verify Go implementation exists
ls -la internal/tools/workflow_mode.go
```

**Impact:** None - function is never called

---

#### 3. `git_tools` Tool

**Status:** ✅ Fully native Go (`internal/tools/git_tools.go`)  
**Python Location:** `project_management_automation/tools/consolidated_git.py`  
**Lines:** ~85 lines (lines 50-134)  
**Bridge Status:** ❌ NOT imported (commented out in `bridge/execute_tool.py` line 29)

**Removal Action:**
```python
# Remove function definition (lines 50-134)
def git_tools(...) -> str:
    # ... entire function ...
```

**Verification:**
```bash
# Verify not imported in bridge
grep -n "git_tools" bridge/execute_tool.py
# Should show only comment

# Verify Go implementation exists
ls -la internal/tools/git_tools.go
```

**Impact:** None - function is never called

---

#### 4. `prompt_tracking` Tool

**Status:** ✅ Fully native Go (`internal/tools/prompt_tracking.go`)  
**Python Location:** `project_management_automation/tools/consolidated_config.py`  
**Lines:** ~41 lines (lines 152-192)  
**Bridge Status:** ❌ NOT imported (commented out in `bridge/execute_tool.py` line 61)

**Removal Action:**
```python
# Remove function definition (lines 152-192)
def prompt_tracking(...) -> str:
    # ... entire function ...
```

**Verification:**
```bash
# Verify not imported in bridge
grep -n "prompt_tracking" bridge/execute_tool.py
# Should show only comment: "# Note: prompt_tracking removed - fully native Go"

# Verify Go implementation exists
ls -la internal/tools/prompt_tracking.go
```

**Impact:** None - function is never called

---

#### 5. `infer_session_mode` Tool

**Status:** ✅ Fully native Go (`internal/tools/session_mode_inference.go`)  
**Python Location:** `project_management_automation/tools/session_mode_inference_interfaces.py`  
**Lines:** ~16 lines (lines 311-325)  
**Bridge Status:** ❌ NOT imported (commented out in `bridge/execute_tool.py` line 29)  
**Special Note:** This is a **stub** that raises `NotImplementedError`

**Removal Action:**
```python
# Remove function definition (lines 311-325)
def infer_session_mode_tool(force_recompute: bool = False) -> str:
    """
    MCP Tool: infer_session_mode
    ...
    """
    raise NotImplementedError("Agent C must implement this")
```

**Verification:**
```bash
# Verify not imported in bridge
grep -n "infer_session_mode" bridge/execute_tool.py
# Should show only comment

# Verify Go implementation exists
ls -la internal/tools/session_mode_inference.go
```

**Impact:** None - function is a stub that's never called

---

### 🟡 **Category 2: Consolidated Module Exports (MEDIUM RISK)**

These tools are exported from `consolidated.py` but the functions themselves are in separate files. We need to remove the exports and imports.

#### 6. Remove Exports from `consolidated.py`

**File:** `project_management_automation/tools/consolidated.py`  
**Lines to Remove:**
- Line 36: `from .consolidated_config import generate_config, setup_hooks, prompt_tracking`
  - Change to: `from .consolidated_config import generate_config, setup_hooks`
- Line 42: `from .consolidated_workflow import workflow_mode, recommend, tool_catalog`
  - Change to: `from .consolidated_workflow import recommend`
- Line 45: `from .consolidated_git import git_tools, session`
  - Change to: `from .consolidated_git import session`
- Lines 71, 77, 79, 81: Remove from `__all__` list:
  - Remove `"prompt_tracking"` (line 71)
  - Remove `"workflow_mode"` (line 77)
  - Remove `"tool_catalog"` (line 79)
  - Remove `"git_tools"` (line 81)

**Verification:**
```bash
# Verify consolidated.py still imports other tools correctly
grep -n "from .consolidated" project_management_automation/tools/consolidated.py

# Verify __all__ list is correct
grep -A 20 "__all__" project_management_automation/tools/consolidated.py
```

**Impact:** Low - only affects imports, not actual functionality

---

## Files Already Removed ✅

The following files were already removed in previous cleanup:

- ✅ `mcp_stdio_tools/server.py` - Old Python MCP server (removed)
- ✅ `mcp_stdio_tools/research_helpers.py` - Unused research helpers (removed)
- ✅ `mcp_stdio_tools/__init__.py` - Package init (removed)

**Verification:**
```bash
# Confirm directory doesn't exist
ls -la mcp_stdio_tools/ 2>&1
# Should show: "No such file or directory"
```

---

## Files That Must Be Kept ✅

### Bridge Scripts (REQUIRED)
- ✅ `bridge/execute_tool.py` - Executes Python tools from Go
- ✅ `bridge/execute_resource.py` - Executes Python resources from Go
- ✅ `bridge/get_prompt.py` - Retrieves Python prompts (if still used)
- ✅ `bridge/context/` - Context tool bridge modules
- ✅ `bridge/recommend/` - Recommend tool bridge modules
- ✅ `bridge/agent_evaluation.py` - Agent evaluation script
- ✅ `bridge/statistics_bridge.py` - Statistics bridge script

### Test Files (REQUIRED)
- ✅ All files in `tests/` directory

### Tools Still Using Python Bridge
- ✅ All tools imported in `bridge/execute_tool.py` lines 31-60
- ✅ All resources imported in `bridge/execute_resource.py`
- ✅ `mlx` tool - Intentionally Python-only (no Go bindings)

---

## Removal Procedure

### Phase 1: Pre-Removal Verification

1. **Verify Go implementations exist:**
   ```bash
   ls -la internal/tools/tool_catalog.go
   ls -la internal/tools/workflow_mode.go
   ls -la internal/tools/git_tools.go
   ls -la internal/tools/prompt_tracking.go
   ls -la internal/tools/session_mode_inference.go
   ```

2. **Verify bridge doesn't import these tools:**
   ```bash
   grep -n "tool_catalog\|workflow_mode\|git_tools\|prompt_tracking\|infer_session_mode" bridge/execute_tool.py
   # Should only show comments, not actual imports
   ```

3. **Run tests to ensure current functionality works:**
   ```bash
   make test
   ```

### Phase 2: Removal Execution

1. **Remove function definitions:**
   - Remove `tool_catalog()` from `consolidated_workflow.py` (lines 198-242)
   - Remove `workflow_mode()` from `consolidated_workflow.py` (lines 50-94)
   - Remove `git_tools()` from `consolidated_git.py` (lines 50-134)
   - Remove `prompt_tracking()` from `consolidated_config.py` (lines 152-192)
   - Remove `infer_session_mode_tool()` from `session_mode_inference_interfaces.py` (lines 311-325)

2. **Update consolidated.py exports:**
   - Remove imports of removed tools
   - Remove from `__all__` list

3. **Update comments in consolidated files:**
   - Remove or update "LEGACY" comments if functions are removed

### Phase 3: Post-Removal Verification

1. **Verify Python syntax:**
   ```bash
   python3 -m py_compile project_management_automation/tools/consolidated*.py
   python3 -m py_compile project_management_automation/tools/session_mode_inference_interfaces.py
   ```

2. **Run Go tests:**
   ```bash
   make test-go
   ```

3. **Run integration tests:**
   ```bash
   make test-integration
   ```

4. **Verify bridge scripts still work:**
   ```bash
   # Test a tool that still uses Python bridge
   ./bin/exarp-go -tool memory -args '{"action":"list"}'
   ```

### Phase 4: Documentation Update

1. Update `docs/PYTHON_CODE_AUDIT_REPORT.md` to reflect removals
2. Update any documentation that references these Python implementations

---

## Risk Assessment

### 🟢 **LOW RISK Removals**

**Rationale:**
- Functions are **not imported** by bridge scripts
- Go implementations are **fully functional** and tested
- Functions are marked as "LEGACY" in comments
- No active code paths call these Python functions

**Mitigation:**
- Pre-removal verification ensures Go implementations exist
- Post-removal tests verify no regressions
- Easy rollback (git revert) if issues occur

### 🟡 **MEDIUM RISK Removals**

**Rationale:**
- Removing exports from `consolidated.py` could break imports in other Python files
- Need to verify no other Python code imports these functions directly

**Mitigation:**
- Search codebase for direct imports before removal
- Update imports in any files that use these functions
- Test Python syntax after changes

---

## Rollback Plan

If issues occur after removal:

1. **Immediate Rollback:**
   ```bash
   git revert <commit-hash>
   ```

2. **Partial Rollback:**
   - Restore specific function if only one causes issues
   - Keep other removals intact

3. **Verification:**
   ```bash
   make test
   make test-integration
   ```

---

## Impact Analysis

### Code Reduction
- **Lines Removed:** ~500-600 lines
- **Files Modified:** 5 files
- **Functions Removed:** 5 functions

### Dependencies
- **No breaking changes** - functions are not called
- **No test failures expected** - tests use Go implementations
- **No bridge script changes needed** - already don't import these

### Performance
- **No performance impact** - code was never executed
- **Slight reduction in Python module load time** (negligible)

---

## Verification Commands Summary

```bash
# 1. Verify Go implementations exist
ls -la internal/tools/{tool_catalog,workflow_mode,git_tools,prompt_tracking,session_mode_inference}.go

# 2. Verify bridge doesn't import removed tools
grep -n "tool_catalog\|workflow_mode\|git_tools\|prompt_tracking\|infer_session_mode" bridge/execute_tool.py

# 3. Verify no other Python code imports these
grep -r "from.*consolidated.*import.*tool_catalog\|workflow_mode\|git_tools\|prompt_tracking\|infer_session_mode" project_management_automation/

# 4. Run tests
make test
make test-go
make test-integration

# 5. Test bridge scripts
./bin/exarp-go -tool memory -args '{"action":"list"}'
```

---

## Next Steps

1. ✅ **Analysis Complete** - This document created
2. ⏳ **Review Plan** - Human review of removal plan
3. ⏳ **Execute Removal** - Create separate task for actual removal
4. ⏳ **Verify Results** - Run all verification commands
5. ⏳ **Update Documentation** - Update audit report

---

## Notes

- All removed functions are marked as "LEGACY" in their source files
- Bridge scripts explicitly comment that these tools are removed
- Go implementations are fully tested and functional
- Removal is safe because functions are never called

---

**Last Updated:** 2026-01-12  
**Status:** Ready for Review
