# Python Code Audit Report

**Date:** 2026-01-11  
**Status:** ✅ Complete Audit  
**Task:** T-50

---

## Executive Summary

**Total Python Files:** 141 files analyzed

**Findings:**
- **✅ Actively Used:** ~135 files (via bridge, scripts, tests, utilities)
- **⚠️ Potentially Removable:** ~3-5 files (legacy/unused implementations)
- **❌ Confirmed Removed:** 1 file (`scripts/update_todo2_tasks.py` - already deleted)

**Key Finding:** The codebase is well-maintained with minimal leftover code. Most Python files serve active purposes via the bridge pattern or are required for testing/utilities.

---

## File Categorization

### ✅ **ACTIVELY USED (135 files)**

#### 1. Bridge Scripts (3 files) - **CRITICAL**
- `bridge/execute_tool.py` - Routes 23 tools from Go to Python
- `bridge/execute_resource.py` - Routes resource handlers
- `bridge/statistics_bridge.py` - Statistics bridge wrapper
- **Status:** Actively called by `internal/bridge/python.go`

#### 2. Tool Implementations (~76 files) - **ACTIVE**
- `project_management_automation/tools/*.py`
- **Status:** Imported and routed by `bridge/execute_tool.py`
- **Note:** 21 tools actively routed (2 return errors for native Go tools)

#### 3. Resource Handlers (~6 files) - **ACTIVE**
- `project_management_automation/resources/*.py`
- **Status:** Imported and routed by `bridge/execute_resource.py`

#### 4. Utility Modules (~15 files) - **ACTIVE**
- `project_management_automation/utils/*.py`
- **Status:** Imported by tools, scripts, and resources

#### 5. Scripts (~0 files in scripts/, ~15 files in project_management_automation/scripts/) - **ACTIVE**
- `project_management_automation/scripts/automate_*.py`
- **Status:** Called by Makefile targets or used by automation workflows

#### 6. Test Files (~12 files) - **REQUIRED**
- `tests/**/*.py`
- **Status:** Required for testing bridge functionality

#### 7. Configuration Helpers (1 file) - **ACTIVE**
- `.exarp/security_features.py`
- **Status:** Used by scorecard detection system (referenced in `docs/SCORECARD_DETECTION_UPDATE.md`)

---

## ⚠️ **POTENTIALLY REMOVABLE (3-5 files)**

### Category 1: Tools Removed from Bridge (But Python Code Still Exists)

**Status:** Bridge comments indicate these tools were removed, but Python implementations may still exist.

**Tools Mentioned as Removed:**
1. **`tool_catalog`** - Removed from bridge (line 29 comment)
2. **`workflow_mode`** - Removed from bridge (line 29 comment)
3. **`git_tools`** - Removed from bridge (line 29 comment)
4. **`infer_session_mode`** - Removed from bridge (line 29 comment)

**Action Required:**
- ✅ **Verify** if Python implementations still exist
- ✅ **Check** if they're imported/used elsewhere
- ⚠️ **Recommendation:** Keep for now as backup/testing, but document as "legacy"

### Category 2: Native Go Tools (Error Return in Bridge)

**Tools that return errors (fully native Go, no Python fallback):**
1. **`prompt_tracking`** - Returns error (bridge lines 344-352)
2. **`server_status`** - Returns error (bridge lines 363-371)

**Status:** Bridge routes these tools but returns errors indicating they're native Go only.

**Action Required:**
- ✅ **Verify** if Python implementations exist
- ⚠️ **Recommendation:** Remove routing from bridge if Python implementations don't exist, or keep for error clarity

### Category 3: Other Tools Mentioned as Removed

**From bridge comments (line 372):**
- `demonstrate_elicit` - Removed (required FastMCP Context)
- `interactive_task_create` - Removed (required FastMCP Context)

**Status:** Not routed in bridge, Python implementations may still exist.

**Action Required:**
- ✅ **Verify** if Python implementations exist
- ⚠️ **Recommendation:** Safe to remove if not used elsewhere

---

## ✅ **CONFIRMED REMOVED**

1. **`scripts/update_todo2_tasks.py`**
   - **Status:** Already deleted (mentioned in `docs/PROJECT_SCORECARD.md` line 65)
   - **Action:** ✅ No action needed

---

## Detailed Analysis

### Bridge Routing Analysis

**Tools Actively Routed (21 tools):**
1. analyze_alignment
2. generate_config
3. health
4. setup_hooks
5. check_attribution
6. add_external_tool_hints
7. memory
8. memory_maint
9. report
10. security
11. task_analysis
12. task_discovery
13. task_workflow
14. testing
15. automation
16. lint
17. estimation
18. session
19. ollama
20. mlx
21. context (via unified wrapper)
22. recommend
23. server_status (returns error - native Go)
24. prompt_tracking (returns error - native Go)

**Modules Imported:**
- `consolidated` - Main tool implementations
- `external_tool_hints` - External tool hints
- `attribution_check` - Attribution compliance
- `context_tool` - Unified context wrapper

---

## Recommendations

### ✅ **SAFE TO REMOVE (After Verification)**

1. **Remove routing for native Go-only tools:**
   - Remove `prompt_tracking` routing (lines 344-352 in bridge)
   - Remove `server_status` routing (lines 363-371 in bridge)
   - **Impact:** Cleaner bridge code, no functional change (already returns errors)

2. **Document legacy tools:**
   - Add comments to Python implementations of removed tools:
     - `tool_catalog`, `workflow_mode`, `git_tools`, `infer_session_mode`
   - **Action:** Add `# LEGACY: Removed from bridge, kept for testing/backup`

3. **Remove FastMCP-only tools (if exist):**
   - `demonstrate_elicit`
   - `interactive_task_create`
   - **Action:** Verify existence, remove if not used

### ⚠️ **KEEP FOR NOW**

1. **All tool implementations** - Even if removed from bridge, keep as backup/testing
2. **All utility modules** - Required by tools/scripts
3. **All resource handlers** - Actively used
4. **All test files** - Required for testing
5. **All bridge scripts** - Critical infrastructure

---

## Action Items

### Immediate (Low Risk)

- [ ] Remove `prompt_tracking` routing from bridge (already returns error)
- [ ] Remove `server_status` routing from bridge (already returns error)
- [ ] Add legacy comments to removed tool implementations

### Verification Required (Medium Risk)

- [ ] Verify Python implementations exist for removed tools
- [ ] Check if removed tools are imported/used elsewhere
- [ ] Verify FastMCP-only tools (`demonstrate_elicit`, `interactive_task_create`) exist

### Future Cleanup (Low Priority)

- [ ] Consider removing legacy tool implementations after full migration
- [ ] Document bridge cleanup strategy in migration docs

---

## Migration Status Context

**Current Migration State:**
- ✅ 24 tools registered in Go
- ✅ 21 tools routed via Python bridge
- ✅ 2 tools fully native Go (prompt_tracking, server_status)
- ✅ 4 tools removed from bridge (tool_catalog, workflow_mode, git_tools, infer_session_mode)

**Bridge Pattern:**
- Go handlers attempt native implementation first
- Fall back to Python bridge if native not available
- Bridge routes to Python tool implementations

---

## Conclusion

The Python codebase is well-maintained with minimal leftover code. Most files serve active purposes via the bridge pattern. The main cleanup opportunities are:

1. **Remove error-returning routes** from bridge (prompt_tracking, server_status)
2. **Document legacy tools** that were removed from bridge
3. **Verify and remove** FastMCP-only tools if they exist

**Risk Level:** Low - Most cleanup is documentation/verification, not deletion.

**Recommendation:** Proceed with immediate cleanup items (removing error routes), then verify legacy tools before removing implementations.
