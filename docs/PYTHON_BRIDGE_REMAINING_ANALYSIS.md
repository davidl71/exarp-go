# Python Bridge Remaining Tools Analysis

**Date:** 2026-01-12  
**Status:** Analysis Complete  
**Goal:** Identify remaining Python bridge tools and prioritize migration targets

---

## Executive Summary

**Total Python Bridge Calls:** 28 occurrences across 5 files

**Files with Python Bridge:**
- `handlers.go`: 20 occurrences (most tools)
- `testing.go`: 3 occurrences
- `security.go`: 3 occurrences
- `report_mlx.go`: 1 occurrence
- `automation_native.go`: 1 occurrence

---

## Tool Categorization by Priority

### 🔴 High Priority - Quick Wins (Simple Actions)

These are single actions or simple implementations that would be quick to migrate:

#### 1. **health/docs** action
- **Status:** `server` action already native, `docs` action uses Python bridge
- **Complexity:** Low-Medium
- **Implementation:** Document generation/health check
- **File:** `internal/tools/health.go`
- **Estimated Effort:** 2-3 days
- **Priority:** High (completes health tool migration)

#### 2. **analyze_alignment/prd** action
- **Status:** `todo2` action already native, `prd` action uses Python bridge
- **Complexity:** Medium
- **Implementation:** PRD analysis and alignment checking
- **File:** `internal/tools/alignment_analysis.go`
- **Estimated Effort:** 3-4 days
- **Priority:** High (completes analyze_alignment tool migration)

#### 3. **automation/nightly** action
- **Status:** Currently stubbed (returns error)
- **Complexity:** Medium
- **Implementation:** Nightly automation workflow
- **File:** `internal/tools/automation_native.go`
- **Estimated Effort:** 4-5 days
- **Priority:** High (completes automation daily/nightly/sprint trio)

#### 4. **automation/sprint** action
- **Status:** Currently stubbed (returns error)
- **Complexity:** Medium
- **Implementation:** Sprint automation workflow
- **File:** `internal/tools/automation_native.go`
- **Estimated Effort:** 4-5 days
- **Priority:** High (completes automation daily/nightly/sprint trio)

---

### 🟡 Medium Priority - Partial Native Tools

These tools have some native implementations but still use Python bridge for certain actions:

#### 5. **security** tool
- **Status:** Some actions native (scan, alerts, report), others may use bridge
- **Complexity:** Medium-High
- **Implementation:** Security scanning and vulnerability detection
- **File:** `internal/tools/security.go`
- **Python Bridge Calls:** 3 occurrences
- **Estimated Effort:** 5-7 days
- **Priority:** Medium (already partially native)

#### 6. **testing** tool
- **Status:** Some actions may be native, others use bridge
- **Complexity:** Medium
- **Implementation:** Test execution and coverage analysis
- **File:** `internal/tools/testing.go`
- **Python Bridge Calls:** 3 occurrences
- **Estimated Effort:** 4-6 days
- **Priority:** Medium

#### 7. **report** tool
- **Status:** Most actions native (overview, scorecard, briefing, prd), MLX enhancement uses bridge
- **Complexity:** Low (just MLX integration)
- **Implementation:** MLX enhancement for reports
- **File:** `internal/tools/report_mlx.go`
- **Python Bridge Calls:** 1 occurrence (MLX only)
- **Estimated Effort:** 1-2 days (if migrating MLX, otherwise keep as-is)
- **Priority:** Low (MLX is Python-native, may keep bridge)

---

### 🟢 Lower Priority - Complex/ML-Dependent Tools

These tools are complex or have ML/AI dependencies that may require keeping Python bridge:

#### 8. **memory** tool
- **Status:** Basic operations native, semantic search may use bridge
- **Complexity:** High (semantic search requires embeddings)
- **Implementation:** Memory storage and retrieval
- **File:** `internal/tools/memory.go`
- **Python Bridge Calls:** Fallback only
- **Estimated Effort:** 10-14 days (if migrating semantic search)
- **Priority:** Low (semantic search may require Python/ML libraries)

#### 9. **ollama** tool
- **Status:** Some actions native, others use bridge (docs, quality, summary)
- **Complexity:** Medium-High
- **Implementation:** Ollama API integration
- **File:** `internal/tools/ollama.go`
- **Python Bridge Calls:** For docs/quality/summary actions
- **Estimated Effort:** 5-7 days
- **Priority:** Medium (Ollama has HTTP API, can be fully native)

#### 10. **mlx** tool
- **Status:** Fully Python bridge (MLX is Python-native)
- **Complexity:** High
- **Implementation:** Apple Silicon ML acceleration
- **File:** `internal/tools/mlx.go`
- **Python Bridge Calls:** All actions
- **Estimated Effort:** N/A (MLX is Python-native, no Go bindings)
- **Priority:** Very Low (keep Python bridge - MLX has no Go equivalent)

#### 11. **automation** tool (complex workflows)
- **Status:** `daily` and `discover` native, `nightly`/`sprint` stubbed
- **Complexity:** Very High
- **Implementation:** Complex workflow engine
- **File:** `internal/tools/automation_native.go`
- **Python Bridge Calls:** 1 occurrence (for Python tool execution in daily workflow)
- **Estimated Effort:** 15-20 days (if fully migrating)
- **Priority:** Medium (nightly/sprint are quick wins, full migration is complex)

---

## Recommended Migration Order

### Phase 1: Quick Wins (1-2 weeks)
1. ✅ **health/docs** - Complete health tool migration
2. ✅ **analyze_alignment/prd** - Complete analyze_alignment tool migration
3. ✅ **automation/nightly** - Complete nightly automation
4. ✅ **automation/sprint** - Complete sprint automation

**Total Estimated Effort:** 13-17 days

### Phase 2: Medium Complexity (2-3 weeks)
5. **security** - Complete remaining security actions
6. **testing** - Complete remaining testing actions
7. **ollama/docs, quality, summary** - Complete ollama tool migration

**Total Estimated Effort:** 14-20 days

### Phase 3: Complex/Deferred (Optional)
8. **memory** semantic search - Evaluate Go embedding libraries first
9. **mlx** - Keep Python bridge (no Go equivalent)
10. **automation** full migration - Evaluate if needed

---

## Tools Already Fully Native ✅

These tools no longer use Python bridge (or use it only as fallback):

- ✅ `generate_config` - Fully native
- ✅ `setup_hooks` - Fully native
- ✅ `add_external_tool_hints` - Fully native
- ✅ `memory_maint` - All actions native (including dream)
- ✅ `report/briefing` - Now native (uses devwisdom-go)
- ✅ `recommend/advisor` - Now native (uses devwisdom-go)
- ✅ `task_analysis` - Mostly native (hierarchy requires Apple FM)
- ✅ `task_discovery` - Mostly native (basic scanning)
- ✅ `task_workflow` - Mostly native (clarify requires Apple FM)
- ✅ `git_tools` - Fully native
- ✅ `session` - Mostly native
- ✅ `context` - Partially native (summarize/budget)
- ✅ `lint` - Hybrid (Go linters native, others bridge)
- ✅ `estimation` - Partially native (with Apple FM)

---

## Summary Statistics

**Total Tools:** 24
**Fully Native:** ~8 tools (33%)
**Partially Native:** ~10 tools (42%)
**Python Bridge Only:** ~6 tools (25%)

**Python Bridge Calls Remaining:** 28 occurrences
- Most are fallbacks or specific actions
- Many tools are already mostly native

---

## Next Steps

1. **Start with Quick Wins** - Complete health/docs, analyze_alignment/prd, automation/nightly/sprint
2. **Evaluate Complex Tools** - Assess if full migration is worth it for memory, mlx, automation
3. **Document Intentional Bridge Retention** - For tools that should stay Python (mlx, complex ML features)

---

**Last Updated:** 2026-01-12
