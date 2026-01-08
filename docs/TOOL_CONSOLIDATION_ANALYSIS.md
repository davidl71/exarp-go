# Tool Consolidation Analysis

**Date:** 2026-01-07  
**Status:** 📋 Analysis Complete - Ready for Implementation

---

## Executive Summary

**Current State:** 38 tools registered across 5 batches  
**Consolidation Opportunity:** Remove 6 duplicate Python bridge tools that are covered by unified tools in Batch 5  
**Proposed Reduction:** 38 → 32 tools (15.8% reduction)

---

## Duplicate Tools Identified

### Batch 4 vs Batch 5 Overlap

**Batch 4 (Individual Tools):**
1. `context_summarize` - Python bridge
2. `context_batch` - Python bridge  
3. `prompt_log` - Python bridge
4. `prompt_analyze` - Python bridge
5. `recommend_model` - Python bridge
6. `recommend_workflow` - Python bridge

**Batch 5 (Unified Tools):**
1. `context` - Unified wrapper (action=summarize|budget|batch) ✅ Covers context_summarize + context_batch
2. `prompt_tracking` - Unified wrapper (action=log|analyze) ✅ Covers prompt_log + prompt_analyze
3. `recommend` - Unified wrapper (action=model|workflow|advisor) ✅ Covers recommend_model + recommend_workflow

**Conclusion:** The 6 individual tools in Batch 4 are redundant and can be removed.

---

## Tools to Keep (Not Consolidated)

### Native Go Tools (Keep)
- ✅ `context_budget` - **KEEP** (Native Go implementation, more efficient than Python bridge)
- ✅ `list_models` - **KEEP** (Native Go, standalone tool)

**Reasoning:** These are native Go implementations that provide better performance than Python bridge alternatives. The unified `context` tool can still call `context_budget` via action parameter, but keeping the native version provides direct access.

---

## Consolidation Plan

### Phase 1: Remove Duplicate Tools (6 tools)

**Remove from Batch 4:**
1. ❌ `context_summarize` → Use `context(action="summarize")`
2. ❌ `context_batch` → Use `context(action="batch")`
3. ❌ `prompt_log` → Use `prompt_tracking(action="log")`
4. ❌ `prompt_analyze` → Use `prompt_tracking(action="analyze")`
5. ❌ `recommend_model` → Use `recommend(action="model")`
6. ❌ `recommend_workflow` → Use `recommend(action="workflow")`

**Keep in Batch 4:**
- ✅ `context_budget` (Native Go)
- ✅ `list_models` (Native Go)

### Phase 2: Update Documentation

- Update `docs/BRIDGE_ANALYSIS_TABLE.md` to reflect new tool count
- Update tool catalog to show unified tools as primary interface
- Add migration notes for any existing code using individual tools

---

## Implementation Details

### Files to Modify

1. **internal/tools/registry.go**
   - Remove 6 tool registrations from `registerBatch4Tools()`
   - Keep `context_budget` and `list_models` registrations
   - Update batch comment: "Batch 4: mcp-generic-tools migration (2 native Go tools)"

2. **internal/tools/handlers.go**
   - Remove handler functions:
     - `handleContextSummarize`
     - `handleContextBatch`
     - `handlePromptLog`
     - `handlePromptAnalyze`
     - `handleRecommendModel`
     - `handleRecommendWorkflow`
   - Keep unified handlers:
     - `handleContext` (already handles all context actions)
     - `handlePromptTracking` (already handles all prompt actions)
     - `handleRecommend` (already handles all recommend actions)

3. **bridge/execute_tool.py** (if needed)
   - Verify Python bridge routing still works for unified tools
   - Remove any special handling for individual tools

### Testing Requirements

- ✅ Verify all unified tools work correctly
- ✅ Test backward compatibility (if keeping individual tools temporarily)
- ✅ Update tool count in tests
- ✅ Verify no broken references in documentation

---

## Tool Count Summary

| Category | Before | After | Change |
|----------|--------|-------|--------|
| **Total Tools** | 38 | 32 | -6 (-15.8%) |
| **Batch 1** | 6 | 6 | 0 |
| **Batch 2** | 8 | 8 | 0 |
| **Batch 3** | 10 | 10 | 0 |
| **Batch 4** | 8 | 2 | -6 |
| **Batch 5** | 6 | 6 | 0 |

---

## Migration Path

### For Users/Code Using Individual Tools

**Before:**
```go
// Old way
context_summarize(data="...")
prompt_log(prompt="...")
recommend_model(task_description="...")
```

**After:**
```go
// New unified way
context(action="summarize", data="...")
prompt_tracking(action="log", prompt="...")
recommend(action="model", task_description="...")
```

### Backward Compatibility Options

**Option 1: Remove Immediately (Recommended)**
- Clean break, forces migration to unified tools
- Simpler codebase, less maintenance
- Users must update to unified tools

**Option 2: Deprecate First**
- Mark individual tools as deprecated
- Keep for 1-2 releases
- Remove after deprecation period
- More user-friendly but requires maintenance

**Recommendation:** Option 1 - Remove immediately since:
- Unified tools are already available (Batch 5)
- Migration is straightforward (just add action parameter)
- Reduces code complexity immediately

---

## Benefits

✅ **Reduced Tool Count:** 38 → 32 tools (15.8% reduction)  
✅ **Simplified API:** Fewer tools to learn and maintain  
✅ **Consistent Interface:** All related tools use action parameter pattern  
✅ **Less Code:** Remove 6 handler functions and registrations  
✅ **Better Organization:** Clear separation between individual and unified tools  

---

## Risks & Considerations

⚠️ **Breaking Change:** Removing tools will break any code using individual tools  
✅ **Mitigation:** Unified tools already exist, migration is simple  
⚠️ **Native Go Tools:** Keeping `context_budget` as native provides better performance  
✅ **Documentation:** Need to update all references to removed tools  

---

## Next Steps

1. ✅ **Analysis Complete** (this document)
2. ⏳ **Review & Approval** - Get stakeholder approval for consolidation
3. ⏳ **Implementation** - Remove duplicate tools from registry
4. ⏳ **Testing** - Verify all unified tools work correctly
5. ⏳ **Documentation** - Update tool catalog and migration guide
6. ⏳ **Release** - Deploy consolidated tool set

---

## Related Documents

- `docs/BRIDGE_ANALYSIS_TABLE.md` - Current tool implementation status
- `internal/tools/registry.go` - Tool registration code
- `internal/tools/handlers.go` - Tool handler implementations

---

**Status:** ✅ Analysis complete, ready for implementation review

