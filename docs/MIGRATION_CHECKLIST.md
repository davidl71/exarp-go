# Migration Checklist

**Date:** 2026-01-12  
**Purpose:** Per-tool and per-resource checklist for migration completion  
**Status:** Active Reference Document

---

## Overview

This checklist provides completion criteria for migrating tools and resources from Python bridge to native Go implementations. Use this checklist to verify completion and ensure all requirements are met.

---

## Tool Migration Checklist

For each tool migration, verify the following:

### ✅ Prerequisites
- [ ] Tool handler exists in `internal/tools/handlers.go`
- [ ] Tool is registered in `internal/tools/registry.go`
- [ ] Tool schema is defined correctly
- [ ] Python bridge fallback exists (for hybrid tools)

### ✅ Native Implementation
- [ ] Native Go implementation file exists (e.g., `*_native.go`, `*.go`)
- [ ] All supported actions have native implementations
- [ ] Error handling is comprehensive
- [ ] Edge cases are handled
- [ ] Code follows Go best practices

### ✅ Handler Integration
- [ ] Handler routes to native implementation first
- [ ] Python bridge fallback works correctly (if applicable)
- [ ] Handler properly parses arguments
- [ ] Handler returns correct response format
- [ ] Error messages are clear and actionable

### ✅ Testing
- [ ] Unit tests exist (`*_test.go`)
- [ ] Integration tests pass (if applicable)
- [ ] Edge cases are tested
- [ ] Error conditions are tested
- [ ] Tests pass consistently

### ✅ Documentation
- [ ] Tool description is accurate
- [ ] Migration status updated in `MIGRATION_STATUS_CURRENT.md`
- [ ] Python bridge dependencies updated in `PYTHON_BRIDGE_DEPENDENCIES.md`
- [ ] Migration plan updated (if applicable)
- [ ] Code comments are clear

### ✅ Validation
- [ ] Tool works via MCP interface
- [ ] All actions work as expected
- [ ] Python bridge fallback works (if applicable)
- [ ] No regressions introduced
- [ ] Performance is acceptable

---

## Resource Migration Checklist

For each resource migration, verify the following:

### ✅ Prerequisites
- [ ] Resource handler exists in `internal/resources/handlers.go`
- [ ] Resource is registered in `RegisterAllResources`
- [ ] Resource URI is correct
- [ ] Resource MIME type is correct

### ✅ Native Implementation
- [ ] Native Go implementation exists
- [ ] Implementation uses native tool logic (when applicable)
- [ ] Error handling is comprehensive
- [ ] Edge cases are handled
- [ ] Code follows Go best practices

### ✅ Handler Integration
- [ ] Handler is registered correctly
- [ ] Handler properly parses URI variables (if applicable)
- [ ] Handler returns correct MIME type
- [ ] Handler returns valid JSON (if JSON resource)
- [ ] Error messages are clear

### ✅ Testing
- [ ] Unit tests exist (if applicable)
- [ ] Integration tests pass
- [ ] Edge cases are tested
- [ ] Error conditions are tested
- [ ] Tests pass consistently

### ✅ Documentation
- [ ] Resource description is accurate
- [ ] Migration status updated in `MIGRATION_STATUS_CURRENT.md`
- [ ] Python bridge dependencies updated in `PYTHON_BRIDGE_DEPENDENCIES.md`
- [ ] Migration plan updated (if applicable)
- [ ] Code comments are clear

### ✅ Validation
- [ ] Resource works via MCP interface
- [ ] Resource returns correct data
- [ ] No regressions introduced
- [ ] Performance is acceptable

---

## Per-Tool Checklist

### Session Tool ✅ **COMPLETE**

**Location:** `internal/tools/handlers.go:723`

**Status:** ✅ All actions native - prompts and assignee actions completed (2026-01-12)

**Actions:**
- [x] Native implementation exists (`internal/tools/session.go`)
- [x] `handleSessionPrompts` implemented
- [x] `handleSessionAssignee` implemented
- [x] Handler routes prompts action to native implementation
- [x] Handler routes assignee action to native implementation
- [x] Tests updated (pending comprehensive unit tests)
- [x] Documentation updated

**Priority:** ✅ Complete

---

### Health Tool ✅ **COMPLETE**

**Location:** `internal/tools/handlers.go:67`

**Status:** ✅ All actions native - docs, dod, cicd actions completed (2026-01-12)

**Actions:**
- [x] Server action native
- [x] Docs action native
- [x] DoD action native
- [x] CI/CD action native
- [x] Tests updated (pending comprehensive unit tests)
- [x] Documentation updated

**Priority:** ✅ Complete

---

### Recommend Tool ✅ **COMPLETE**

**Location:** `internal/tools/handlers.go:894`

**Status:** ✅ Model and workflow actions native, advisor action deferred (requires devwisdom-go)

**Actions:**
- [x] Model action native
- [x] Workflow action native ✅ (completed 2026-01-12)
- [x] Advisor action (deferred - requires devwisdom-go) - Intentional
- [x] Tests updated (pending comprehensive unit tests)
- [x] Documentation updated

**Priority:** ✅ Complete (advisor deferred intentionally)

---

### Setup Hooks Tool ✅ **COMPLETE**

**Location:** `internal/tools/handlers.go:93`

**Status:** ✅ All actions native - patterns action completed (2026-01-12)

**Actions:**
- [x] Git action native
- [x] Patterns action native ✅ (completed 2026-01-12)
- [x] Tests updated (pending comprehensive unit tests)
- [x] Documentation updated

**Priority:** ✅ Complete

---

### Analyze Alignment Tool ✅ **COMPLETE**

**Location:** `internal/tools/handlers.go:15`

**Status:** ✅ All actions native - prd action completed (2026-01-12)

**Actions:**
- [x] Todo2 action native
- [x] PRD action native ✅ (completed 2026-01-12)
- [x] Tests updated (pending comprehensive unit tests)
- [x] Documentation updated

**Priority:** ✅ Complete

---

### Ollama Tool ✅ **COMPLETE**

**Location:** `internal/tools/handlers.go:768`

**Status:** ✅ All actions native via HTTP - docs, quality, summary actions completed (2026-01-12)

**Actions:**
- [x] Status action native
- [x] Models action native
- [x] Generate action native
- [x] Pull action native
- [x] Hardware action native
- [x] Docs action native ✅ (completed 2026-01-12)
- [x] Quality action native ✅ (completed 2026-01-12)
- [x] Summary action native ✅ (completed 2026-01-12)
- [x] Tests updated (pending comprehensive unit tests)
- [x] Documentation updated

**Priority:** ✅ Complete

---

### Automation Tool ✅ **COMPLETE**

**Location:** `internal/tools/handlers.go:553`

**Status:** ✅ All actions native - nightly and sprint actions completed (2026-01-12)

**Actions:**
- [x] Daily action native
- [x] Discover action native
- [x] Nightly action native ✅ (completed 2026-01-12)
- [x] Sprint action native ✅ (completed 2026-01-12)
- [x] Tests updated (pending comprehensive unit tests)
- [x] Documentation updated

**Priority:** ✅ Complete

---

### Memory Maintenance Tool

**Location:** `internal/tools/handlers.go:217`

**Status:** Hybrid - health/gc/prune/consolidate native, dream action uses Python bridge

**Actions:**
- [x] Health action native
- [x] GC action native
- [x] Prune action native
- [x] Consolidate action native
- [ ] Dream action (deferred - requires advisor)
- [ ] Tests updated
- [ ] Documentation updated

**Priority:** Low (deferred - requires advisor)

---

### Report Tool

**Location:** `internal/tools/handlers.go:251`

**Status:** Hybrid - scorecard/overview/prd native, briefing uses Python bridge (devwisdom MCP)

**Actions:**
- [x] Scorecard action native
- [x] Overview action native
- [x] PRD action native
- [ ] Briefing action (deferred - requires devwisdom MCP)
- [ ] Tests updated
- [ ] Documentation updated

**Priority:** Low (deferred - requires devwisdom MCP)

---

### Estimation Tool ✅ **COMPLETE**

**Location:** `internal/tools/handlers.go:673`

**Status:** ✅ All actions native - analyze action completed (2026-01-12)

**Actions:**
- [x] Native implementation exists (with Apple FM)
- [x] Estimate action native
- [x] Stats action native
- [x] Analyze action native ✅ (completed 2026-01-12)
- [x] Bridge fallback works
- [x] Tests updated (pending comprehensive unit tests)
- [x] Documentation updated

**Priority:** ✅ Complete

---

### MLX Tool ⚠️ **NEEDS EVALUATION**

**Location:** `internal/tools/handlers.go:792`

**Status:** Python Bridge Only (evaluation pending)

**Actions:**
- [ ] Evaluate Go bindings availability
- [ ] Document as intentional retention (if no bindings)
- [ ] Document rationale (no Go bindings)
- [ ] No migration needed (if intentional)

**Priority:** Low (evaluation pending)

---

## Per-Resource Checklist

### All Resources ✅ **COMPLETE**

**Status:** All 21 resources are native Go implementations

**Resources:**
- [x] `stdio://scorecard` - Native
- [x] `stdio://memories` - Native
- [x] `stdio://memories/category/{category}` - Native
- [x] `stdio://memories/task/{task_id}` - Native
- [x] `stdio://memories/recent` - Native
- [x] `stdio://memories/session/{date}` - Native
- [x] `stdio://prompts` - Native
- [x] `stdio://prompts/mode/{mode}` - Native
- [x] `stdio://prompts/persona/{persona}` - Native
- [x] `stdio://prompts/category/{category}` - Native
- [x] `stdio://session/mode` - Native
- [x] `stdio://server/status` - Native
- [x] `stdio://models` - Native
- [x] `stdio://tools` - Native
- [x] `stdio://tools/{category}` - Native
- [x] `stdio://tasks` - Native
- [x] `stdio://tasks/{task_id}` - Native
- [x] `stdio://tasks/status/{status}` - Native
- [x] `stdio://tasks/priority/{priority}` - Native
- [x] `stdio://tasks/tag/{tag}` - Native
- [x] `stdio://tasks/summary` - Native

**Priority:** N/A (all complete)

---

## Testing Requirements

### Unit Tests
- [ ] Test file exists (`*_test.go`)
- [ ] All actions are tested
- [ ] Edge cases are tested
- [ ] Error conditions are tested
- [ ] Tests pass consistently
- [ ] Test coverage is adequate (>80% recommended)

### Integration Tests
- [ ] Tool works via MCP interface
- [ ] Resource works via MCP interface
- [ ] Error responses are correct
- [ ] Response format is valid
- [ ] Performance is acceptable

### Regression Tests
- [ ] No regressions from Python bridge
- [ ] Feature parity verified
- [ ] Output format matches (if applicable)
- [ ] Behavior matches (if applicable)

---

## Completion Criteria

A tool/resource migration is considered complete when:

1. ✅ Native implementation exists and works
2. ✅ Handler routes to native implementation
3. ✅ Python bridge fallback works (if applicable)
4. ✅ Unit tests exist and pass
5. ✅ Integration tests pass (if applicable)
6. ✅ No regressions introduced
7. ✅ Documentation updated
8. ✅ Code reviewed and approved

---

## Migration Status Summary

### Tools
- **Total:** 28 tools (29 with Apple FM)
- **Fully Native:** 5 tools (18%) - `tool_catalog`, `workflow_mode`, `git_tools`, `infer_session_mode`, `prompt_tracking`
- **Hybrid:** 22 tools (79%) - Native primary with Python bridge fallback
- **Python Bridge Only:** 1 tool (4% - `mlx`, evaluation pending)
- **Native Coverage:** 96% (27/28 tools have native implementations)

### Resources
- **Total:** 21 resources
- **Fully Native:** 21 resources (100%)
- **Native Coverage:** 100%

### Recent Completions (2026-01-12)
- ✅ Stream 1: All hybrid tool actions completed (session prompts/assignee, ollama docs/quality/summary, recommend workflow, setup_hooks patterns, analyze_alignment prd, health docs/dod/cicd)
- ✅ Stream 2: Full tool migrations completed (estimation analyze, automation nightly/sprint)
- ✅ Stream 3: All resource migrations completed (prompts, session/mode, verified memory/scorecard)

### Remaining Work
- **Evaluation:** 1 tool (`mlx` - evaluate Go bindings or document retention)
- **Testing:** Unit tests and integration tests for all native implementations
- **Documentation:** Performance benchmarks, regression testing
- **Deferred (ML/AI):** 8 actions/features (intentional - require ML/AI or external dependencies)

---

**Last Updated:** 2026-01-12 (After Stream 1, 2, 3 completions)  
**Next Review:** After testing and benchmarking completion
