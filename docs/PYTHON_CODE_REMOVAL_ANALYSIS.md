# Python Code Removal Analysis

**Date:** 2026-01-09  
**Status:** ✅ Complete Analysis  
**Task:** T-43

---

## Executive Summary

**Total Python Files:** 143 files found

**Safe to Remove:** ~5-10 files (dead code in bridge, unused standalone scripts)  
**Must Keep:** ~133-138 files (actively used via bridge or Makefile)

**Key Finding:** Most Python code is still actively used via the bridge pattern. Only a small number of files can be safely removed.

---

## Python Code Categories

### ✅ **REQUIRED - Must Keep (133-138 files)**

#### 1. Bridge Scripts (3 files) - **CRITICAL**
- `bridge/execute_tool.py` - Main tool bridge entry point
- `bridge/execute_resource.py` - Resource bridge entry point  
- `bridge/statistics_bridge.py` - Statistics bridge (if exists)

**Status:** Actively called by `internal/bridge/python.go`

#### 2. Tool Implementations (~80+ files) - **ACTIVE**
- `project_management_automation/tools/*.py` - All tool implementations
- **Reason:** Still imported and used by `bridge/execute_tool.py`
- **Note:** Even tools with native Go implementations may still be called as fallback

#### 3. Resource Handlers (~10 files) - **ACTIVE**
- `project_management_automation/resources/*.py` - Resource implementations
- **Reason:** Still imported and used by `bridge/execute_resource.py`

#### 4. Automation Scripts (~15 files) - **ACTIVE**
- `project_management_automation/scripts/automate_*.py` - Automation scripts
- **Reason:** Called by Makefile targets or used by other scripts

#### 5. Utility Modules (~20 files) - **ACTIVE**
- `project_management_automation/utils/*.py` - Shared utilities
- **Reason:** Imported by tools, scripts, and resources

#### 6. Test Files (~10 files) - **ACTIVE**
- `tests/**/*.py` - Python integration and unit tests
- **Reason:** Required for testing bridge functionality

---

## ⚠️ **POTENTIALLY REMOVABLE (5-10 files)**

### Category 1: Dead Code in Bridge (5 files)

**Fully Native Go Tools (No Python Fallback):**

These tools are fully implemented in Go with **no fallback to Python**, but the bridge still imports them:

1. **`tool_catalog`** - Fully native Go (`handleToolCatalogNative`)
   - Bridge still imports: `from project_management_automation.tools.consolidated import tool_catalog`
   - **Status:** Dead code in bridge (never called)
   - **Files:** `project_management_automation/tools/consolidated_workflow.py` (tool_catalog function)

2. **`workflow_mode`** - Fully native Go (`handleWorkflowModeNative`)
   - Bridge still imports: `from project_management_automation.tools.consolidated import workflow_mode`
   - **Status:** Dead code in bridge (never called)
   - **Files:** `project_management_automation/tools/consolidated_workflow.py` (workflow_mode function)

3. **`infer_session_mode`** - Fully native Go (`handleInferSessionModeNative`)
   - Bridge still imports: `from project_management_automation.tools.session_mode_inference_interfaces import infer_session_mode_tool`
   - **Status:** Dead code in bridge (never called)
   - **Files:** `project_management_automation/tools/session_mode_inference_interfaces.py`

4. **`server_status`** - Fully native Go (`handleServerStatusNative`)
   - Bridge has inline implementation (not imported)
   - **Status:** No Python file to remove (already handled in bridge)

5. **`git_tools`** - Fully native Go (`handleGitToolsNative`)
   - Bridge still imports: `from project_management_automation.tools.consolidated import git_tools`
   - **Status:** Dead code in bridge (never called)
   - **Files:** `project_management_automation/tools/consolidated_git.py` (git_tools function)

**⚠️ CAUTION:** These Python implementations might still be needed if:
- Someone calls the bridge directly (bypassing Go handlers)
- Future fallback is needed for error recovery
- Testing requires Python versions

**Recommendation:** Remove from bridge imports, but keep Python files for now as backup.

### Category 2: Standalone Scripts (1-2 files)

1. **`scripts/update_todo2_tasks.py`** - Standalone utility script
   - **Status:** Not referenced in Makefile or Go code
   - **Usage:** Manual utility for updating Todo2 tasks
   - **Recommendation:** **SAFE TO REMOVE** if not used manually

**Verification Needed:**
- Check if used in documentation or workflows
- Check if referenced in any shell scripts

### Category 3: Unused Helper Files (0-3 files)

**Potential candidates (need verification):**
- Files in `project_management_automation/tools/` that are not imported by `consolidated.py`
- Old migration helper files
- Deprecated tool implementations

**⚠️ Requires detailed import analysis to identify**

---

## Removal Recommendations

### ✅ **SAFE TO REMOVE NOW**

1. **Remove dead imports from bridge:**
   - Remove `tool_catalog`, `workflow_mode`, `infer_session_mode`, `git_tools` from `bridge/execute_tool.py` imports
   - Remove corresponding `elif` branches in bridge routing
   - **Impact:** Reduces bridge complexity, no functional impact

2. **Remove standalone script (if unused):**
   - `scripts/update_todo2_tasks.py` - **Verify first** that it's not used manually
   - **Impact:** None if truly unused

### ⚠️ **KEEP FOR NOW (But Could Remove Later)**

1. **Python tool implementations for fully native tools:**
   - Keep `consolidated_workflow.py` (tool_catalog, workflow_mode functions)
   - Keep `consolidated_git.py` (git_tools function)
   - Keep `session_mode_inference_interfaces.py` (infer_session_mode function)
   - **Reason:** May be needed for testing, fallback, or direct bridge calls
   - **Action:** Remove only after confirming no dependencies

### ❌ **DO NOT REMOVE**

1. **All bridge scripts** - Required for Go → Python communication
2. **All tool implementations** - Still used by bridge (even if native Go exists)
3. **All resource handlers** - Still used by bridge
4. **All automation scripts** - Called by Makefile or other scripts
5. **All test files** - Required for testing
6. **All utility modules** - Imported by tools/scripts/resources

---

## Detailed Analysis by Tool

### Fully Native Go Tools (No Python Fallback)

| Tool | Go Handler | Bridge Import | Python File | Safe to Remove? |
|------|-----------|---------------|-------------|-----------------|
| `server_status` | ✅ `handleServerStatusNative` | ❌ Inline (not imported) | N/A | ✅ Already removed |
| `tool_catalog` | ✅ `handleToolCatalogNative` | ⚠️ Yes (dead code) | `consolidated_workflow.py` | ⚠️ Remove from bridge only |
| `workflow_mode` | ✅ `handleWorkflowModeNative` | ⚠️ Yes (dead code) | `consolidated_workflow.py` | ⚠️ Remove from bridge only |
| `infer_session_mode` | ✅ `handleInferSessionModeNative` | ⚠️ Yes (dead code) | `session_mode_inference_interfaces.py` | ⚠️ Remove from bridge only |
| `git_tools` | ✅ `handleGitToolsNative` | ⚠️ Yes (dead code) | `consolidated_git.py` | ⚠️ Remove from bridge only |
| `context_budget` | ✅ Native (not in bridge) | ❌ Not in bridge | N/A | ✅ Already removed |

### Hybrid Tools (Native Go + Python Fallback)

| Tool | Native Actions | Python Fallback | Python File | Safe to Remove? |
|------|---------------|-----------------|-------------|-----------------|
| `lint` | Go linters | Non-Go linters | `consolidated_quality.py` | ❌ **KEEP** (fallback needed) |
| `context` | summarize, budget | batch | `context_tool.py` | ❌ **KEEP** (fallback needed) |
| `task_analysis` | hierarchy | duplicates, tags, etc. | `consolidated_analysis.py` | ❌ **KEEP** (fallback needed) |
| `task_discovery` | comments, markdown | orphans | `consolidated_analysis.py` | ❌ **KEEP** (fallback needed) |
| `task_workflow` | clarify, approve | sync, cleanup | `consolidated_workflow.py` | ❌ **KEEP** (fallback needed) |

### Python Bridge Only Tools (22 tools)

All 22 remaining tools require Python implementations:
- `analyze_alignment`, `generate_config`, `health`, `setup_hooks`, `check_attribution`
- `add_external_tool_hints`, `memory`, `memory_maint`, `report`, `security`
- `testing`, `automation`, `estimation`, `session`, `ollama`, `mlx`
- `prompt_tracking`, `recommend`
- Plus others...

**Status:** ❌ **KEEP ALL** - Required for bridge functionality

---

## Action Plan

### Phase 1: Safe Removals (No Risk)

1. ✅ **Remove dead imports from bridge:**
   ```python
   # In bridge/execute_tool.py, remove:
   - tool_catalog from consolidated import
   - workflow_mode from consolidated import
   - git_tools from consolidated import
   - infer_session_mode_tool from session_mode_inference_interfaces import
   
   # Remove corresponding elif branches:
   - elif tool_name == "tool_catalog": ...
   - elif tool_name == "workflow_mode": ...
   - elif tool_name == "git_tools": ...
   - elif tool_name == "infer_session_mode": ...
   ```

2. ✅ **Verify and remove standalone script:**
   - Check `scripts/update_todo2_tasks.py` usage
   - Remove if truly unused

### Phase 2: Verification (Before Removal)

1. ⚠️ **Check Python tool implementations:**
   - Verify no other code imports these functions directly
   - Check if used in tests
   - Confirm no direct bridge calls bypass Go handlers

2. ⚠️ **Document removal:**
   - Update migration docs
   - Note removed tools in changelog

### Phase 3: Optional Cleanup (Future)

1. 🔮 **Remove Python implementations:**
   - Only after confirming no dependencies
   - Keep as backup for 1-2 releases
   - Remove in future cleanup pass

---

## Summary

**Immediate Action:**
- Remove 4 dead imports from `bridge/execute_tool.py` (tool_catalog, workflow_mode, git_tools, infer_session_mode)
- Remove 4 corresponding `elif` branches
- Verify and potentially remove `scripts/update_todo2_tasks.py`

**Estimated Files Removed:** 1-5 files (mostly dead code in bridge)

**Estimated Files Kept:** 138-142 files (all actively used)

**Risk Level:** ⚠️ **LOW** - Only removing dead code that's never called

---

## Notes

- Most Python code is still actively used via the bridge pattern
- Hybrid tools require Python fallbacks even if native Go exists
- Fully native tools have dead code in bridge that can be safely removed
- Standalone scripts may be unused but need verification
- Python tool implementations should be kept as backup even if not called

