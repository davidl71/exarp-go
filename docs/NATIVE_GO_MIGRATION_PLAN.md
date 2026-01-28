# Native Go Migration Plan for exarp-go

**Date:** 2026-01-08  
**Last Updated:** 2026-01-28 (Migration complete except mlx; context and task_workflow full native)  
**Status:** ✅ Migration complete for all tools with native implementations; mlx hybrid (native status/hardware on darwin+CGO, bridge for models/generate)  
**Goal:** Migrate all remaining Python bridge tools, resources, and prompts to native Go implementation

---

## Executive Summary

This plan outlines a comprehensive migration strategy to convert all remaining Python bridge implementations to native Go, with strategic use of hybrid approaches where appropriate. This plan represents an **intentional architectural decision** to move away from Python bridge dependencies.

**Supersedes:** `MIGRATION_STRATEGY.md` (2026-01-07), which recommended Python Bridge for most tools. This plan instead prioritizes **Native Go** and is the current source of truth.

**Key Insight:** The superseded strategy prioritized **speed of migration** (Python Bridge); this plan prioritizes **performance, maintainability, and reducing external dependencies** (Native Go).

---

## Strategy Divergence and Rationale

### Existing Strategy vs This Plan

**Existing `MIGRATION_STRATEGY.md` Recommendation:**
- **Pattern 1: Python Bridge (Recommended for Most Tools)**
- Rationale: Faster migration, preserves existing Python logic
- Decision Framework: Use Python Bridge unless tool is simple or performance-critical

**This Plan's Approach:**
- **Pattern 1: Native Go (Recommended for Most Tools)**
- Rationale: Better performance, single language codebase, reduced dependencies, easier maintenance
- Decision Framework: Use Native Go unless tool has complex ML/AI dependencies or lacks Go equivalents

### When Hybrid Approach Was Used (Historical)

**As of 2026-01-28:** One hybrid tool: **mlx** — native status/hardware via luxfi/mlx when built with CGO on darwin; bridge for models/generate. All other tools with native implementations are **full native** (no bridge fallback).

**Previously,** the hybrid pattern (Native Go with Python Bridge Fallback) was used for progressive migration. Fallbacks were removed in favor of returning clear errors when native fails or a feature is unsupported (e.g. context summarize without FM, task_workflow external=true).

---

## Current Status (Updated 2026-01-28)

**Audit Reference:** See `docs/MIGRATION_AUDIT_2026-01-12.md` for comprehensive audit findings (and 2026-01-28 note below).

**Native Go Coverage (Actual):**
- **Full Native (no bridge):** 28 tools — All tools with native implementations; no Python fallback. Includes: server_status, tool_catalog, workflow_mode, infer_session_mode, git_tools, generate_config, add_external_tool_hints, setup_hooks, check_attribution, session, memory_maint, memory, analyze_alignment, estimation, task_analysis, task_discovery, prompt_tracking, health, report, recommend, security, testing, lint, ollama, **context**, **task_workflow**, and others. See `docs/NATIVE_GO_HANDLER_STATUS.md` for the full list.
- **Hybrid Native:** 1 tool — `mlx` (native status/hardware via luxfi/mlx on darwin+CGO; bridge for models/generate — see § mlx below).
- **Python Bridge Only:** 0 tools — mlx is hybrid; bridge used only for models/generate when native unavailable.
- **Total Tools:** 28 (plus 1 conditional Apple FM tool = 28–29)
- **Native Coverage:** 100% of tools with native implementations; 1 tool remains bridge-only by design/choice.

**2026-01-28:** **Context** and **task_workflow** fully native: context (summarize, budget, batch; unknown action returns error); task_workflow (sync SQLite↔JSON, approve, clarify, clarity, cleanup, create; external=true returns error). No hybrid tools remain. See `docs/NATIVE_GO_HANDLER_STATUS.md`, `docs/TASK_WORKFLOW_FULL_NATIVE_PLAN.md`.

**2026-01-28:** Report briefing/scorecard shrink (briefing native-only; non-Go scorecard returns clear error). Estimation native-only, no Python fallback; estimation handler removed from `bridge/execute_tool.py`. Tests updated.

**2026-01-27:** Four more tools fully native: `setup_hooks`, `check_attribution`, `session`, `memory_maint`. Bridge cleanup: removed dead `generate_config` and `add_external_tool_hints` branches.

**Resources (Actual - from Audit):**
- **Native:** 21 resources (100%) — All resources use native Go as primary implementation
- **Hybrid:** 0 resources (0%)
- **Python Bridge Only:** 0 resources (0%)
- **Total Resources:** 21
- **Native Coverage:** 100% (21/21 resources have native implementations)

**Prompts:**
- **Native:** 19 prompts (100%)
- **Status:** ✅ All prompts migrated to native Go

**Migration Progress:**
- **Overall Tools:** 28/28 (100%) have native implementations **or** are intentionally bridge-only (mlx). No hybrid tools.
- **Overall Resources:** 21/21 (100%) have native implementations
- **Phase 1–5:** All phases complete. See `docs/NATIVE_GO_HANDLER_STATUS.md` for current handler status.

**Recent Completions:**
- **2026-01-28:** **Context** and **task_workflow** full native (no bridge). Migration complete for all tools except mlx (bridge-only). See `NATIVE_GO_HANDLER_STATUS.md`, `TASK_WORKFLOW_FULL_NATIVE_PLAN.md`.
- **2026-01-27:** **4 removed Python fallbacks** — `setup_hooks`, `check_attribution`, `session`, `memory_maint` full native. Bridge cleanup.
- **2026-01-12:** Stream 1–3 — hybrid tool actions, full tool migrations, resource migrations (see audit doc)

---

## mlx: Options for Native Go

**Current:** The `mlx` tool is **bridge-only** — `handleMlx` and `insight_provider.executeMLXViaBridge` call the Python bridge for status, hardware, models, and generate. Used by report/scorecard AI insights (DefaultReportInsight: FM then MLX).

**Options:**

| Option | Description | Effort |
|--------|-------------|--------|
| **A. Keep bridge-only** | Leave mlx on Python; document as intentional. Simplest; no CGO/MLX dependency in Go. | None |
| **B. Evaluate Go MLX bindings** | Use **luxfi/mlx** (`github.com/luxfi/mlx`) — Go bindings for Apple MLX, MIT license, mirrors Python API; includes Metal bindings. Would allow a native `handleMlx` for status/hardware/models/generate on Apple Silicon. Requires CGO, Apple Silicon; LLM inference patterns can be adapted from mlx-lm (Python) or luxfi docs. | Medium: add dependency, implement native handler, keep bridge as fallback for non-Apple or if native fails |
| **C. Alternative: GoMLX** | **GoMLX** (`github.com/gomlx`) — accelerated ML framework for Go (740+ stars); go-huggingface for models/tokenizers. Different API than MLX Python; would need adapter layer. | Higher: different API, more integration work |

**Recommendation:** **A** for now (keep bridge-only). If you want to remove the last Python tool dependency, **B** (luxfi/mlx) is the most direct path for a native MLX tool on Apple Silicon; do a small spike (add `github.com/luxfi/mlx`, implement one action e.g. status or generate) before committing.

**See:** `docs/MLX_NATIVE_OPTIONS.md` for a dedicated mlx options doc (options A/B/C, references, handler locations).

---

## Migration Strategy

### Decision Framework

**Tool Complexity Assessment:**

1. **Simple Tools** → Full Native Go
   - Minimal dependencies
   - Straightforward logic
   - Examples: `server_status`, `tool_catalog`, `workflow_mode`

2. **Medium Complexity** → Native Go with possible hybrid
   - Some dependencies but manageable
   - May need Go library alternatives
   - Examples: `generate_config`, `setup_hooks`, `check_attribution`

3. **Complex Tools** → Hybrid or Python Bridge
   - Heavy ML/AI dependencies
   - Complex Python ecosystem integration
   - May require keeping Python bridge
   - Examples: `memory` (semantic search), `automation` (complex workflow engine), `ollama/mlx` (no Go bindings)

4. **Platform-Specific** → Hybrid (Apple FM on macOS, Python bridge elsewhere)
   - Uses Apple Foundation Models
   - Requires CGO on macOS
   - Examples: `context` summarize, `task_discovery`

### Implementation Patterns

**Pattern 1: Pure Native Go**
```go
// internal/tools/server_status.go
func handleServerStatus(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
    // Direct Go implementation, no Python bridge
    status := map[string]interface{}{
        "status": "operational",
        "version": "0.1.0",
        "project_root": os.Getenv("PROJECT_ROOT"),
    }
    result, _ := json.MarshalIndent(status, "", "  ")
    return []framework.TextContent{{Type: "text", Text: string(result)}}, nil
}
```

**Pattern 2: Hybrid with Fallback** (Recommended for complex tools)
```go
// internal/tools/handlers.go
func handleToolName(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
    // Try native Go first
    result, err := handleToolNameNative(ctx, params)
    if err == nil {
        return result, nil
    }
    
    // Fall back to Python bridge
    return bridge.ExecutePythonTool(ctx, "tool_name", params)
}
```

**Pattern 3: Platform-Specific Native**
```go
// internal/tools/tool_native.go (darwin && arm64 && cgo)
func handleToolNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
    // Uses Apple Foundation Models
    support := platform.CheckAppleFoundationModelsSupport()
    if !support.Supported {
        return nil, fmt.Errorf("Apple FM not available")
    }
    // Native implementation using Apple FM
}

// internal/tools/tool_native_nocgo.go (!(darwin && arm64 && cgo))
func handleToolNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
    return nil, fmt.Errorf("not implemented on this platform")
}
```

---

## Phase 1: Simple Tools (High Priority) ✅ **COMPLETE**

**Goal:** Establish patterns and build momentum with straightforward migrations  
**Status:** ✅ **COMPLETE** (Completed 2026-01-09)  
**Timeline:** 5-7 days estimated, **ACTUALLY COMPLETED** in ~1 week

### Tools Migrated (3 tools, `infer_session_mode` moved to Phase 2)

#### 1.1 `server_status` ✅ **COMPLETE**
- **Complexity:** Very Low
- **Dependencies:** None
- **Implementation:** Pure Native Go
- **File:** `internal/tools/server_status.go`
- **Timeline:** 1 day
- **Status:** ✅ **COMPLETED** - Native Go implementation in place

#### 1.2 `tool_catalog` ✅ **COMPLETE**
- **Complexity:** Low
- **Dependencies:** Tool registry (already in Go)
- **Implementation:** Pure Native Go
- **File:** `internal/tools/tool_catalog.go`
- **Timeline:** 2-3 days
- **Status:** ✅ **COMPLETED** - Native Go implementation in place

#### 1.3 `workflow_mode` ✅ **COMPLETE**
- **Complexity:** Low
- **Dependencies:** Workflow state file (JSON-based)
- **Implementation:** Pure Native Go
- **File:** `internal/tools/workflow_mode.go`
- **Timeline:** 2-3 days
- **Status:** ✅ **COMPLETED** - Native Go implementation in place

**Phase 1 Total:** 5-7 days estimated, **ACTUALLY COMPLETED** ✅

---

## Phase 2: Medium Complexity Tools (Medium Priority)

**Goal:** Migrate tools with manageable dependencies, establish hybrid patterns

### Tools to Migrate (9 tools, including `infer_session_mode` moved from Phase 1)

#### 2.1 `infer_session_mode` ✅ **COMPLETE** ⚠️ MOVED FROM PHASE 1
- **Complexity:** Low-Medium
- **Dependencies:** Todo2 file parsing, session data analysis
- **Implementation:** Native Go (uses existing `todo2_utils.go`)
- **File:** `internal/tools/session_mode_inference.go`
- **Timeline:** 3-4 days
- **Status:** ✅ **COMPLETED** - Native Go implementation in place
- **Note:** Moved to Phase 2 due to dependency on Todo2 parsing utilities

#### 2.2 `generate_config`
- **Complexity:** Medium
- **Dependencies:** File system, template rendering
- **Implementation:** Native Go with Go templates
- **File:** `internal/tools/config_generator.go`
- **Challenges:** Template logic porting
- **Timeline:** 4-5 days

#### 2.3 `setup_hooks`
- **Complexity:** Medium
- **Dependencies:** Git hooks, file system
- **Implementation:** Native Go using `os/exec`
- **File:** `internal/tools/hooks_setup.go`
- **Challenges:** Python hook scripts may need shell script conversion
- **Timeline:** 3-4 days

#### 2.4 `check_attribution`
- **Complexity:** Medium
- **Dependencies:** File scanning, license detection
- **Implementation:** Native Go
- **File:** `internal/tools/attribution_check.go`
- **Timeline:** 4-5 days

#### 2.5 `add_external_tool_hints`
- **Complexity:** Medium
- **Dependencies:** AST manipulation (`go/ast`)
- **Implementation:** Native Go using Go AST
- **File:** `internal/tools/external_tool_hints.go`
- **Timeline:** 4-5 days

#### 2.6 `analyze_alignment`
- **Complexity:** Medium-High
- **Dependencies:** Todo2 parsing, alignment analysis
- **Implementation:** Native Go (reuse `todo2_utils.go`)
- **File:** `internal/tools/alignment_analysis.go`
- **Timeline:** 5-6 days

#### 2.7 `health`
- **Complexity:** Medium
- **Dependencies:** Multiple health check sources
- **Implementation:** Native Go using `os/exec` for external checks
- **File:** `internal/tools/health_check.go`
- **Timeline:** 4-5 days

#### 2.8 `context` (batch action)
- **Complexity:** Medium
- **Dependencies:** Context summarization (reuse existing native implementation)
- **Implementation:** Extend existing `internal/tools/context.go`
- **Timeline:** 2-3 days
- **Note:** Reuses `context` summarize logic

#### 2.9 `recommend`
- **Complexity:** Medium-High
- **Dependencies:** Model/workflow/advisor recommendation logic
- **Implementation:** Hybrid approach - Native Go where possible, Python bridge for complex ML
- **File:** `internal/tools/recommend.go`
- **Timeline:** 5-7 days
- **Note:** May require keeping Python bridge for ML-based recommendations

**Phase 2 Total:** 34-46 days (~5-7 weeks)

---

## Phase 3: Complex Tools (Lower Priority, Hybrid Recommended)

**Goal:** Migrate complex tools with hybrid approach or document Python bridge retention

### Tools Requiring Hybrid Approach (8 tools)

#### 3.1 `memory` and `memory_maint`
- **Complexity:** High
- **Dependencies:** Semantic search, embedding models
- **Implementation:** **HYBRID RECOMMENDED**
  - Native Go for CRUD operations and file management
  - Python bridge for semantic search (unless Go embedding library available)
- **Files:** `internal/tools/memory.go`, `internal/tools/memory_maint.go`
- **Timeline:** 10-14 days
- **Decision:** Evaluate Go embedding libraries (e.g., `github.com/qdrant/go-client`) before full migration

#### 3.2 `report`
- **Complexity:** High
- **Dependencies:** Multiple data sources, report generation
- **Implementation:** Native Go for aggregation, Python bridge for complex templates
- **File:** `internal/tools/report.go`
- **Timeline:** 8-10 days
- **Note:** Can migrate data aggregation to Go, keep Python for report formatting if complex

#### 3.3 `security`
- **Complexity:** High
- **Dependencies:** Security scanning tools, vulnerability databases
- **Implementation:** **HYBRID RECOMMENDED**
  - Native Go for orchestration
  - Python bridge for security scanner integration
- **File:** `internal/tools/security.go`
- **Timeline:** 6-8 days

#### 3.4 `testing`
- **Complexity:** Medium-High
- **Dependencies:** Test framework detection, coverage parsing
- **Implementation:** Native Go for orchestration, Python bridge for test case suggestion (ML)
- **File:** `internal/tools/testing.go`
- **Timeline:** 7-9 days

#### 3.5 `automation` ✅ **FULLY NATIVE (2026-01-27)**
- **Complexity:** Very High
- **Dependencies:** Complex workflow engine, parallel execution
- **Implementation:** **FULLY NATIVE GO** — no Python bridge
  - **daily:** analyze_alignment, task_analysis, health (all via `runDailyTask` native)
  - **nightly:** memory_maint, task_analysis, health (all via `runDailyTask` native)
  - **sprint:** analyze_alignment, task_analysis, report overview (all via `runDailyTask` native)
  - **discover:** task_discovery native
- **File:** `internal/tools/automation_native.go`
- **Status:** `runDailyTaskPython` removed; all sub-tasks use native handlers (health, memory_maint, report, etc.)

#### 3.6 `estimation`
- **Complexity:** Medium-High
- **Dependencies:** ML estimation (MLX), historical data
- **Implementation:** **HYBRID RECOMMENDED**
  - Native Go for historical data analysis
  - Python bridge for MLX integration (if no Go bindings)
- **File:** `internal/tools/estimation.go`
- **Timeline:** 6-8 days

#### 3.7 `git_tools` ✅ **COMPLETE**
- **Complexity:** Medium
- **Dependencies:** Git operations (`go-git` library)
- **Implementation:** Native Go using `go-git`
- **File:** `internal/tools/git_tools.go`
- **Timeline:** 5-7 days estimated
- **Status:** ✅ **COMPLETED** 2026-01-09 (ahead of schedule!)
- **Note:** Go has excellent Git libraries. All 7 actions implemented (commits, branches, tasks, diff, graph, merge, set_branch)

#### 3.8 `session`
- **Complexity:** Medium-High
- **Dependencies:** Session management, Todo2, Git status
- **Implementation:** Native Go (reuse `todo2_utils.go`)
- **File:** `internal/tools/session.go`
- **Timeline:** 7-9 days

### Tools Requiring Platform-Specific Native (2 tools)

#### 3.9 `ollama` and `mlx`
- **Complexity:** Medium-High
- **Dependencies:** Ollama API (HTTP), MLX bindings
- **Implementation:** **HYBRID RECOMMENDED**
  - Native Go HTTP client for Ollama API
  - Python bridge for MLX (unless Go bindings available)
- **Files:** `internal/tools/ollama.go`, `internal/tools/mlx.go`
- **Timeline:** 6-8 days
- **Note:** Ollama can be native Go (HTTP API), MLX may require Python bridge

#### 3.10 `prompt_tracking`
- **Complexity:** Medium
- **Dependencies:** Prompt log storage, analysis
- **Implementation:** Native Go for storage, Python bridge for complex analysis
- **File:** `internal/tools/prompt_tracking.go`
- **Timeline:** 4-6 days

### Complete Partial Implementations (3 tools)

#### 3.11 `task_analysis` (complete)
- **Status:** Done — fully native Go (all actions: hierarchy, duplicates, tags, dependencies, parallelization)
- **Implementation:** Single native path in `task_analysis_shared.go`; hierarchy uses FM provider abstraction (`DefaultFM`), Apple FM when available via `internal/tools/apple_foundation.GenerateWithOptions`
- **File:** `internal/tools/task_analysis_shared.go`; FM abstraction in `fm_provider.go`, `fm_apple.go`, `fm_stub.go`
- **Bridge:** No Python fallback; `task_analysis` removed from `bridge/execute_tool.py`

#### 3.12 `task_discovery` (complete remaining actions)
- **Status:** Partial (comments/markdown native)
- **Remaining:** `orphans` (requires task_analysis)
- **Implementation:** Native Go, depends on task_analysis completion
- **File:** Extend `internal/tools/task_discovery_native.go`
- **Timeline:** 4-5 days (after task_analysis)

#### 3.13 `task_workflow` (complete remaining actions)
- **Status:** Partial (clarify action native)
- **Remaining:** `sync`, `approve`, `clarity`, `cleanup`
- **Implementation:** Native Go where possible
- **File:** Extend `internal/tools/task_workflow_native.go`
- **Timeline:** 6-8 days

**Phase 3 Total:** 102-134 days (~15-19 weeks)
**Note:** Timeline assumes some tools remain hybrid or Python bridge

---

## Phase 4: Resources Migration ✅ **COMPLETE**

**Status:** All 21 resources use native Go as primary implementation. Python bridge via `bridge/execute_resource.py` is no longer the primary path; native implementations live in `internal/resources/`.

### Resources (21 total — all native)

| Resource | Implementation | Status |
|----------|----------------|--------|
| `stdio://scorecard` | `internal/resources/scorecard.go` | ✅ Native |
| `stdio://memories` (all variants) | `internal/resources/` (memories) | ✅ Native |
| `stdio://prompts` (all variants) | `internal/prompts/templates.go` | ✅ Native |
| `stdio://session/mode` | infer_session_mode tool | ✅ Native |
| `stdio://server/status` | server status resource | ✅ Native |
| `stdio://models`, `stdio://tools`, `stdio://tasks` (all variants) | `internal/resources/` | ✅ Native |

**Phase 4 Total:** ✅ **COMPLETE** (2026-01-12) — 21/21 resources migrated. Historical migration scope (scorecard, memories, etc.) is documented in `MIGRATION_AUDIT_2026-01-12.md` and `MIGRATION_STATUS_CURRENT.md`.

---

## Phase 5: Prompts Migration (Optional/Deferred)

### Decision: Evaluate Case-by-Case

**Recommendation:** Keep prompts as Python bridge unless there's a specific need to migrate.

**Rationale:**
- Prompts are simple template strings
- No performance impact (templates are lightweight)
- Python template rendering is flexible
- Low maintenance burden

**If migrating:**
- Use Go `text/template` or `html/template`
- Store templates as Go string constants or files
- File: `internal/prompts/templates.go`
- Timeline: 10-15 days for all prompts

**Prompts List (15 total):**
1. `align`, `discover`, `config`, `scan`, `scorecard`, `overview`, `dashboard`, `remember`
2. `daily_checkin`, `sprint_start`, `sprint_end`, `pre_sprint`, `post_impl`, `sync`, `dups`

**Phase 5 Total:** 10-15 days (if migrating, otherwise 0)

---

## Tool Dependency Analysis

### Dependency Graph

```
Phase 1 (Simple - No Dependencies):
  server_status → (none)
  tool_catalog → tool registry (internal)
  workflow_mode → (none)

Phase 2 (Medium):
  infer_session_mode → todo2_utils.go
  generate_config → (none)
  setup_hooks → (none)
  check_attribution → (none)
  add_external_tool_hints → go/ast
  analyze_alignment → todo2_utils.go
  health → (none, uses os/exec)
  context (batch) → context.go (existing)
  recommend → (may depend on model registry)

Phase 3 (Complex):
  memory → (may depend on embedding libraries)
  task_analysis (complete) → todo2_utils.go
  task_discovery (complete) → task_analysis (must complete first)
  task_workflow (complete) → todo2_utils.go
  git_tools → go-git library
  session → todo2_utils.go
```

### Critical Paths

1. **Todo2 utilities** (`todo2_utils.go`) - Already exists, used by:
   - `infer_session_mode` (Phase 2)
   - `analyze_alignment` (Phase 2)
   - `task_analysis` (Phase 3)
   - `task_discovery` (Phase 3)
   - `task_workflow` (Phase 3)
   - `session` (Phase 3)

2. **Task analysis completion** - Blocks:
   - `task_discovery` orphans action (depends on task_analysis)

3. **Memory system** - Blocks:
   - Memory resources (Phase 4)

---

## Implementation Guidelines

### File Structure

```
internal/tools/
  ├── server_status.go              (Phase 1)
  ├── tool_catalog.go               (Phase 1)
  ├── workflow_mode.go              (Phase 1)
  ├── session_mode_inference.go     (Phase 2)
  ├── config_generator.go           (Phase 2)
  ├── hooks_setup.go                (Phase 2)
  ├── attribution_check.go          (Phase 2)
  ├── external_tool_hints.go        (Phase 2)
  ├── alignment_analysis.go         (Phase 2)
  ├── health_check.go               (Phase 2)
  ├── context.go                    (Phase 2 - extend existing)
  ├── recommend.go                  (Phase 2)
  ├── memory.go                     (Phase 3 - hybrid)
  ├── memory_maint.go               (Phase 3 - hybrid)
  ├── report.go                     (Phase 3 - hybrid)
  ├── security.go                   (Phase 3 - hybrid)
  ├── testing.go                    (Phase 3 - hybrid)
  ├── automation_native.go          (Phase 3 - fully native, no Python bridge, 2026-01-27)
  ├── estimation.go                 (Phase 3 - hybrid)
  ├── git_tools.go                  (Phase 3)
  ├── session.go                    (Phase 3)
  ├── ollama.go                     (Phase 3 - hybrid)
  ├── mlx.go                        (Phase 3 - hybrid)
  ├── prompt_tracking.go            (Phase 3)
  ├── task_analysis_shared.go       (Phase 3 - fully native, FM provider abstraction)
  ├── task_discovery_native.go      (Phase 3 - extend existing)
  └── task_workflow_native.go       (Phase 3 - extend existing)

internal/resources/
  ├── scorecard.go                  (Phase 4)
  ├── memories.go                   (Phase 4)
  └── memory_health.go              (Phase 4)
```

### Testing Strategy

**Unit Tests:**
- Test each tool function independently
- Mock external dependencies
- Test error handling
- File: `internal/tools/{tool}_test.go`

**Integration Tests:**
- Test via MCP interface
- Verify tool registration
- Test parameter parsing
- File: `tests/integration/mcp/test_{tool}.go`

**Migration Checklist (per tool):**
- [ ] Create native Go implementation file
- [ ] Implement tool logic (or hybrid pattern)
- [ ] Update handler in `internal/tools/handlers.go`
- [ ] Register tool in `internal/tools/registry.go` (if new registration needed)
- [ ] Write unit tests
- [ ] Write integration tests
- [ ] Update documentation
- [ ] Keep Python bridge as fallback (if hybrid)
- [ ] Verify tool works via MCP interface
- [ ] Update migration status docs

---

## Updated Timeline Estimate

**Phase 1:** 5-7 days (~1 week)
- 3 simple tools

**Phase 2:** 34-46 days (~5-7 weeks)
- 9 medium complexity tools
- Includes `infer_session_mode` (moved from Phase 1)

**Phase 3:** 102-134 days (~15-19 weeks)
- 13 complex tools
- Some may remain hybrid or Python bridge
- Includes completing partial implementations

**Phase 4:** 6-9 days (~1 week)
- 6 resources

**Phase 5:** 10-15 days (optional) or 0 days (if keeping Python bridge)
- 15 prompts

**Total Estimated Duration:** 
- **With prompts migration:** 157-211 days (~22-30 weeks, ~5.5-7.5 months)
- **Without prompts migration:** 147-196 days (~21-28 weeks, ~5-7 months)

**Note:** Timeline assumes sequential migration. Parallel work possible for independent tools. Actual timeline may vary based on complexity discovered during implementation.

---

## Risk Mitigation

### High-Risk Items

1. **Memory System** - Complex semantic search
   - **Mitigation:** Hybrid approach - Native Go for CRUD, Python bridge for semantic search
   - **Alternative:** Evaluate Go embedding libraries (e.g., Qdrant Go client)

2. **Automation Engine** - Very complex workflow system
   - **Mitigation:** Document as intentional Python bridge retention
   - **Alternative:** Partial native Go for simple workflows only

3. **MLX/Ollama** - May not have Go bindings
   - **Mitigation:** Hybrid - Native Go for Ollama (HTTP API), Python bridge for MLX
   - **Alternative:** Keep Python bridge if Go bindings unavailable

4. **Semantic Search** - Requires embedding models
   - **Mitigation:** Hybrid approach or evaluate Go embedding libraries
   - **Alternative:** Keep Python bridge for semantic operations

### Mitigation Strategies

1. **Hybrid Approach** - Keep Python bridge as fallback for complex tools
2. **Gradual Migration** - Migrate incrementally, test thoroughly at each step
3. **Document Intentional Retention** - Clearly document which tools remain Python bridge and why
4. **Pattern Reuse** - Leverage existing hybrid patterns (`lint`, `context`, `task_analysis`)
5. **Dependency Analysis** - Complete tools in dependency order (e.g., task_analysis before task_discovery)

---

## Success Criteria

- [x] All simple tools (Phase 1) migrated to native Go ✅ **COMPLETE** (2026-01-09)
- [x] All medium complexity tools (Phase 2) migrated (hybrid where appropriate) ✅ **COMPLETE** (2026-01-12)
- [x] Complex tools evaluated and migrated where practical (Phase 3) ✅ **COMPLETE** (2026-01-12)
  - [x] Document intentional Python bridge retention for `mlx` (no Go bindings)
  - [x] All other tools have native implementations (hybrid pattern)
- [x] All resources migrated to native Go (Phase 4) ✅ **COMPLETE** (2026-01-12) - 21/21 resources
- [x] Prompts evaluated and migrated (Phase 5) ✅ **COMPLETE** (2026-01-12) - 19/19 prompts
- [ ] All migrated tools have unit and integration tests (ongoing)
- [x] Migration documentation updated ✅ (2026-01-12 - comprehensive audit)
- [x] No regressions in functionality ✅ (verified)
- [ ] Performance improvements measured and documented (future work)
- [x] Hybrid patterns documented for future tools ✅ (examples in plan and code)

**Overall Achievement:** 96% tool coverage (27/28), 100% resource coverage (21/21), 100% prompt coverage (19/19)

---

## Architectural Decisions

### ADR 1: Native Go vs Python Bridge Strategy
**Decision:** Prioritize native Go migration over Python bridge retention
**Rationale:** Better performance, single language codebase, reduced dependencies
**Consequence:** Longer migration timeline, but better long-term maintainability

### ADR 2: Hybrid Approach for Complex Tools
**Decision:** Use hybrid pattern (native Go with Python bridge fallback) for tools with ML/AI dependencies
**Rationale:** Best of both worlds - native performance where possible, Python ecosystem for complex ML
**Consequence:** Some tools remain partially dependent on Python bridge

### ADR 3: Intentional Python Bridge Retention
**Decision:** Document tools that intentionally remain as Python bridge
**Rationale:** Some tools (e.g., `automation`) are too complex or benefit too much from Python ecosystem
**Consequence:** Reduced migration scope, but more realistic timeline

### ADR 4: Prompts Migration (Optional)
**Decision:** Evaluate prompts case-by-case, default to keeping Python bridge
**Rationale:** Prompts are lightweight templates, no performance benefit to migration
**Consequence:** May keep prompts as Python bridge indefinitely

---

## Next Steps (Updated 2026-01-09)

1. ✅ **Phase 1 Complete** - All simple tools migrated successfully
2. **Continue Phase 2** - Focus on medium complexity tools:
   - `generate_config` or `health` (simpler, good next targets)
   - `analyze_alignment` (uses Todo2 utilities, can leverage existing patterns)
3. **Complete task tool actions** - Finish remaining actions for:
   - `task_analysis` (duplicates, tags, dependencies, parallelization)
   - `task_discovery` (orphans - depends on task_analysis)
   - `task_workflow` (sync, clarity, cleanup)
4. **Document hybrid patterns** - Create guide for future hybrid implementations
5. **Update progress tracking** - Keep status section current as tools are migrated

---

## Lessons Learned (Updated 2026-01-12)

### Successful Patterns

1. **Pure Native Go** - Works excellently for simple tools:
   - `tool_catalog`, `workflow_mode`, `git_tools`, `infer_session_mode`, `prompt_tracking` - All completed quickly
   - Pattern: Direct Go implementation, minimal dependencies
   - Result: 5 tools fully native (19%)

2. **Hybrid Pattern (Native Primary + Fallback)** - Dominant pattern, highly effective:
   - 22 tools use hybrid pattern (79% of all tools)
   - Pattern: Try native Go first, fallback to Python bridge on error
   - Benefits: Best of both worlds - native performance where possible, Python ecosystem for complex cases
   - Result: 96% of tools have native implementations

3. **Hybrid with Apple FM** - Proven effective for platform-specific features:
   - `context`, `task_discovery`, `task_workflow` - Use this pattern; `task_analysis` is now fully native using FM provider abstraction
   - Pattern: Native Go with Apple FM on macOS (via FM provider), Python bridge fallback elsewhere where applicable
   - Result: Platform-specific features work seamlessly

4. **Resource Migration Strategy** - Highly successful:
   - All 21 resources migrated to native Go (100%)
   - Pattern: Direct native implementation using existing Go packages (prompts, session_mode, etc.)
   - Result: Zero Python bridge dependencies for resources

5. **Todo2 Utilities Reuse** - Highly effective:
   - `infer_session_mode`, `task_analysis`, `task_discovery`, `session` - All use `todo2_utils.go`; `task_analysis` also uses FM provider for hierarchy
   - Pattern: Create reusable utilities, use across multiple tools
   - Result: Faster development, consistent behavior

6. **Go Library Maturity** - Go libraries are excellent:
   - `go-git/v5` - Mature, well-documented, easy to use
   - `gonum.org/v1/gonum/graph` - Excellent for graph algorithms
   - Result: Complex operations (Git, graph algorithms) migrate smoothly

7. **Incremental Migration (Action-by-Action)** - Effective for multi-action tools:
   - Tools like `session`, `ollama`, `health`, `automation` migrated action-by-action
   - Pattern: Migrate one action at a time, keep fallback for unmigrated actions
   - Result: Lower risk, faster delivery, easier testing

### Timeline Insights

- **Simple tools:** Complete faster than estimated (1-2 days vs 5-7 days)
- **Medium tools:** Timeline varies based on dependencies, but hybrid pattern reduces complexity
- **Complex tools:** Can complete faster if good Go libraries available (Git, graph algorithms)
- **Resources:** Complete faster than tools (no complex logic, direct implementation)
- **Action-by-action migration:** Faster than full tool migration, lower risk

### Key Success Factors

1. **Hybrid Pattern Adoption** - Using hybrid pattern (native first, Python fallback) reduces risk and enables incremental migration
2. **Go Ecosystem Maturity** - Excellent libraries available for most use cases (Git, graph, HTTP, etc.)
3. **Incremental Approach** - Migrating action-by-action for multi-action tools proved effective
4. **Resource Migration Priority** - Migrating resources first was strategic (simpler, faster, provides building blocks)
5. **Reusable Utilities** - Creating reusable utilities (todo2_utils, etc.) accelerated development

---

## Next Steps (Updated 2026-01-12)

### ✅ Completed Work (Stream 1, 2, 3)

1. ✅ **Stream 1: Hybrid Tool Actions** - All high-priority actions completed:
   - ✅ `session` - `prompts` and `assignee` actions now native
   - ✅ `ollama` - `docs`, `quality`, `summary` actions now native
   - ✅ `recommend` - `workflow` action now native
   - ✅ `setup_hooks` - `patterns` action now native
   - ✅ `analyze_alignment` - `prd` action now native
   - ✅ `health` - `docs`, `dod`, `cicd` actions now native

2. ✅ **Stream 2: Full Tool Migrations** - All migrations completed:
   - ✅ `estimation` - `analyze` action now native (stats and estimate were already native)
   - ✅ `automation` - daily, nightly, sprint, discover all native; `runDailyTaskPython` removed (2026-01-27)
   - ⚠️ `mlx` - Evaluated, documented as intentional Python bridge retention

3. ✅ **Stream 3: Resource Migrations** - All resources migrated:
   - ✅ All prompt resources migrated to native Go
   - ✅ Session mode resource migrated to native Go
   - ✅ Memory resources verified native Go implementation
   - ✅ Scorecard resource verified native Go implementation

4. ✅ **Stream 4: Documentation** - Audit and documentation updates:
   - ✅ Comprehensive migration audit completed
   - ✅ Migration status documents updated
   - ✅ Python bridge dependencies document updated
   - ✅ Migration plan updated with current status

### Remaining High-Priority Actions

1. **Testing and Validation:**
   - Write comprehensive unit tests for all native implementations
   - Write integration tests for all tools and resources via MCP interface
   - Benchmark native vs Python bridge implementations
   - Document performance improvements
   - Perform regression testing to verify feature parity

2. **Documentation Maintenance:**
   - ✅ Migration status documents updated
   - ✅ Python bridge dependencies updated
   - ✅ Migration plan updated
   - Create migration checklist for future migrations
   - Document hybrid patterns and when to use them

### Low-Priority Actions (ML/AI or External Dependencies)

1. **ML/AI Features (Deferred):**
   - `memory` semantic search (requires ML/AI) - Already hybrid, Python fallback works well
   - `memory_maint` dream action (requires devwisdom-go advisor) - Already hybrid
   - `testing` non-Go language support, ML test generation - Already hybrid
   - `security` non-Go language support - Already hybrid

2. **External Dependencies (Deferred):**
   - `recommend` advisor action (requires devwisdom-go) - Already hybrid
   - `report` briefing action (requires devwisdom-go) - Already hybrid

3. **Intentional Python Bridge Retention:**
   - `mlx` - No Go bindings available (intentional retention) - ✅ Documented

---

**Last Updated:** 2026-01-12 (Comprehensive Audit - Stream 4)  
**Status:** Excellent Progress - 96% Tool Coverage, 100% Resource Coverage ✅  
**Migration Status:** Core migration complete - All tools and resources have native implementations  
**Next Steps:** Focus on testing, validation, and documentation maintenance

