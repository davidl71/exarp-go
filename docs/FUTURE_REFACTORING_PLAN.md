# Future Refactoring Plan

**Date:** 2026-01-12  
**Status:** 📋 Planning  
**Last Refactoring:** Committed in 3d01d18 - Eliminated ~180-200 lines of duplication

---

## Overview

This document outlines identified code duplication and refactoring opportunities for the exarp-go project. The plan is prioritized by impact and complexity.

---

## Completed Refactoring ✅

### 1. GetComments Functions (2026-01-12)
- **Status:** ✅ Complete
- **Impact:** -107 lines (-34% in comments.go)
- **Changes:** Created `queryComments()` helper function
- **Files:** `internal/database/comments.go`

### 2. Tag/Dependency Loading (2026-01-12)
- **Status:** ✅ Complete
- **Impact:** -80-100 lines estimated
- **Changes:** Created `loadTaskTags()` and `loadTaskDependencies()` helper functions
- **Files:** `internal/database/tasks.go`, `internal/database/tasks_lock.go`

---

## Identified Refactoring Opportunities

### 🔴 High Priority

#### 1. Duplicate Task Loading Logic in tasks.go
- **Location:** `internal/database/tasks.go:213-242` and `246-275`
- **Issue:** Similar logic for loading task data (tags/dependencies pattern)
- **Impact:** ~30 lines duplicated
- **Solution:** Already partially addressed with helper functions, but there may be additional patterns to extract
- **Complexity:** Low
- **Estimated Effort:** 1-2 hours

#### 2. Linting Tool Duplication
- **Location:** `internal/tools/linting.go:374-468` and `471-563`
- **Issue:** ~95 lines duplicated in mergeDuplicateTasks function
- **Impact:** Significant duplication in task analysis logic
- **Solution:** Extract common merge/update logic into helper functions
- **Complexity:** Medium
- **Estimated Effort:** 2-3 hours

### 🟡 Medium Priority

#### 3. Database Retry Pattern Consolidation
- **Location:** `internal/database/retry.go`, `internal/database/comments.go`, `internal/database/tasks.go`
- **Issue:** `retryWithBackoff()` pattern used extensively (58 matches across 5 files)
- **Impact:** Pattern is already centralized but could benefit from:
  - Standardizing error messages
  - Consistent timeout handling
  - Better logging integration
- **Solution:** Enhance retry utilities with standardized patterns
- **Complexity:** Low-Medium
- **Estimated Effort:** 2-3 hours

#### 4. Context Management Pattern
- **Location:** Multiple database functions
- **Issue:** Pattern of `ensureContext()` → `withQueryTimeout()` → `defer cancel()` repeated
- **Impact:** ~58 occurrences of context handling patterns
- **Solution:** Create context helpers or wrapper functions
- **Complexity:** Low
- **Estimated Effort:** 1-2 hours

#### 5. Error Handling Consistency
- **Location:** Throughout codebase (10+ files with fmt.Errorf patterns)
- **Issue:** Inconsistent error wrapping and formatting
- **Impact:** Harder to debug, inconsistent error messages
- **Solution:** Create error handling utilities with standardized patterns
- **Complexity:** Medium
- **Estimated Effort:** 3-4 hours

### 🟢 Low Priority

#### 6. Rows.Close() Pattern
- **Location:** `internal/database/*.go` (7 defer .Close() patterns remaining)
- **Issue:** Standard Go pattern, but error handling could be more consistent
- **Impact:** Low - mostly cosmetic
- **Solution:** Create helper function for rows cleanup with consistent error handling
- **Complexity:** Low
- **Estimated Effort:** 1 hour

#### 7. SQL Query Building
- **Location:** `internal/database/*.go`
- **Issue:** SQL queries scattered throughout code, some with string concatenation
- **Impact:** Harder to maintain, potential SQL injection risks (though mitigated by parameterization)
- **Solution:** Consider query builder or at least SQL query constants
- **Complexity:** Medium-High
- **Estimated Effort:** 4-6 hours

#### 8. Metadata JSON Handling
- **Location:** `internal/database/tasks.go`
- **Issue:** JSON marshal/unmarshal pattern repeated
- **Impact:** Moderate duplication
- **Solution:** Create helper functions for task metadata serialization
- **Complexity:** Low
- **Estimated Effort:** 1-2 hours

---

## Large Files (Context Fit)

Files that exceed ~400 lines or ~4k tokens can hurt AI/editor context fit. Consider splitting or extracting by responsibility. **Source:** scorecard/lint or context-budget tooling.

| File | Lines | ~Tokens | Suggested split / refactor |
|------|-------|---------|---------------------------|
| `internal/cli/child_agent.go` | 507 | ~3.5k | Extract agent invocation/streaming into `child_agent_invoke.go`, keep types/opts in main file. |
| `internal/cli/cli.go` | 910 | ~7.5k | Split by subcommand: `cli_dispatch.go` (root), `cli_task.go`, `cli_report.go`, `cli_health.go`, etc.; keep shared flags in `cli.go`. |
| `internal/cli/config.go` | 542 | ~3.8k | Move config load/save vs validation into `config_load.go` / `config_validate.go` if needed. |
| `internal/cli/tui3270_transactions.go` | 985 | ~7.5k | One file per transaction group: `tui3270_menu.go`, `tui3270_tasklist.go`, `tui3270_editor.go`, `tui3270_scorecard.go`, etc. |
| `internal/cli/tui_test.go` | 542 | ~3.7k | Split by area: `tui_test_basic.go`, `tui_test_commands.go`, `tui_test_detail.go` (or similar). |
| `internal/config/protobuf.go` | 1072 | ~8.3k | Split by domain: load/save, validation, defaults, and proto↔internal mapping in separate files. |
| `internal/config/protobuf_test.go` | 526 | ~4.1k | Mirror protobuf split: one test file per prod file. |
| `internal/database/tasks.go` | 1550 | ~10.3k | Extract: `tasks_crud.go`, `tasks_list.go`, `tasks_filters.go`, `tasks_helpers.go`; keep core types in `tasks.go`. |
| `internal/database/tasks_lock.go` | 549 | ~3.6k | Consider `tasks_lock_claim.go` vs `tasks_lock_cleanup.go` if logical. |
| `internal/database/tasks_lock_test.go` | 523 | ~3.3k | Split by test group (claim, renew, cleanup, stale). |
| `internal/database/tasks_test.go` | 731 | ~4.4k | Split by area: CRUD tests, list/filter tests, migration tests. |
| `internal/prompts/templates.go` | 803 | ~8.6k | Group prompts by tool or domain; e.g. `templates_task.go`, `templates_report.go`, `templates_session.go`. |
| `internal/prompts/templates_test.go` | 502 | ~3.4k | Mirror template split. |
| `internal/tools/alignment_analysis.go` | 640 | ~4.5k | Extract: alignment scoring vs PRD/todo2 comparison into separate files. |
| `internal/tools/automation_native.go` | 1049 | ~8.4k | Split by workflow: sprint_start, sprint_end, pre_sprint, or by phase (setup / run / report). |
| `internal/tools/config_generator.go` | 1029 | ~6.6k | Split: init, export, convert (and shared helpers). |
| `internal/tools/config_generator_test.go` | 538 | ~3.2k | Mirror generator split. |
| `internal/tools/estimation_shared.go` | 1072 | ~7.4k | Split: estimation types/helpers, batch logic, single-task flow. |
| `internal/tools/git_tools.go` | 931 | ~6.1k | Split by action: commits, branches, tasks, diff, graph, merge. |
| `internal/tools/graph_helpers.go` | 724 | ~4.5k | Extract: graph build, cycle detection, critical path, level calculation. |

**Principles for splits:**
- Preserve package API; avoid changing callers where possible.
- Keep types and interfaces in one place; move implementation blocks.
- After split: run `make test`, `make lint`, update `.cursor/rules/code-map.mdc` if tool/file mapping changes.

---

## Refactoring Principles

### 1. DRY (Don't Repeat Yourself)
- Extract common patterns into reusable functions
- Centralize logic to single source of truth
- Reduce maintenance burden

### 2. Single Responsibility
- Each function should have one clear purpose
- Helper functions should be focused and testable

### 3. Consistency
- Use consistent error handling patterns
- Standardize naming conventions
- Maintain consistent code style

### 4. Testability
- Refactored code should be easier to test
- Extract functions should have clear inputs/outputs
- Maintain or improve test coverage

### 5. Backward Compatibility
- Ensure refactoring doesn't break existing functionality
- Run all tests after refactoring
- Maintain API compatibility where possible

---

## Implementation Guidelines

### Before Refactoring
1. ✅ Identify duplication with `golangci-lint --enable=dupl`
2. ✅ Measure code reduction potential
3. ✅ Review dependencies and test coverage
4. ✅ Create test cases if needed

### During Refactoring
1. ✅ Extract helper functions incrementally
2. ✅ Update all call sites
3. ✅ Run tests frequently
4. ✅ Verify builds succeed

### After Refactoring
1. ✅ Run full test suite
2. ✅ Verify golangci-lint passes
3. ✅ Measure code reduction
4. ✅ Update documentation if needed
5. ✅ Commit with clear message

---

## Prioritization Criteria

### High Priority (Do First)
- Duplication > 50 lines
- Used in 3+ locations
- Affects critical paths (database, core functionality)
- Easy to refactor (low risk)

### Medium Priority (Do Soon)
- Duplication 20-50 lines
- Used in 2-3 locations
- Improves maintainability
- Moderate complexity

### Low Priority (Backlog)
- Duplication < 20 lines
- Used in 1-2 locations
- Cosmetic improvements
- High complexity or low impact

---

## Metrics & Tracking

### Code Quality Metrics
- **Lines of Code:** Track reduction in total LOC
- **Duplication Rate:** Measure with `golangci-lint --enable=dupl`
- **Cyclomatic Complexity:** Monitor with `gocyclo` linter
- **Test Coverage:** Ensure coverage doesn't decrease

### Success Criteria
- ✅ Duplication warnings reduced
- ✅ Test coverage maintained or improved
- ✅ Builds succeed
- ✅ No performance regressions
- ✅ Code easier to maintain

---

## Next Steps

### Immediate (Next Sprint)
1. ✅ Fix remaining duplication in `tasks.go` (lines 213-242, 246-275)
2. ✅ Refactor `linting.go` mergeDuplicateTasks duplication
3. ✅ Create context management helpers

### Short Term (Next Month)
1. ⏳ Enhance retry pattern utilities
2. ⏳ Standardize error handling
3. ⏳ Create rows cleanup helpers

### Long Term (Future)
1. ⏳ Consider SQL query builder
2. ⏳ Refactor metadata handling
3. ⏳ Review architecture for additional opportunities
4. ⏳ Large-file splits for context fit (see **Large Files (Context Fit)** table above); prioritize `internal/database/tasks.go`, `internal/cli/cli.go`, `internal/config/protobuf.go`, `internal/tools/automation_native.go`.

---

## References

- **Completed Refactoring:** Commit `3d01d18` - "Refactor duplicate code: eliminate ~180-200 lines of duplication"
- **Linting Tool:** `golangci-lint --enable=dupl`
- **Code Quality:** See `docs/TODO2_ALIGNMENT_REPORT.md`

---

**Last Updated:** 2026-02-21  
**Next Review:** After next refactoring session
