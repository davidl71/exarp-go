# Phase 3: Complex Tools Migration - Execution Plan

**Date:** 2026-01-09  
**Status:** Planning  
**Goal:** Migrate complex tools with hybrid approach or document Python bridge retention  
**Progress:** 1/13 tools complete (8%) - `git_tools` ✅

---

## Executive Summary

Phase 3 focuses on migrating complex tools that require hybrid approaches, platform-specific implementations, or completing partial implementations. The phase is organized into 3 sub-phases based on complexity and dependencies.

**Key Strategy:**
- **Hybrid First** - Native Go for orchestration/CRUD, Python bridge for complex ML/AI
- **Complete Partial** - Finish remaining actions for task tools
- **Document Retention** - Clearly document tools that remain Python bridge

**Total Tools:** 13 tools (1 complete, 12 remaining)  
**Estimated Timeline:** 102-134 days (~15-19 weeks)  
**Current Progress:** 8% (1/13 complete)

---

## Phase 3 Structure

### Sub-Phase 3A: Complete Partial Implementations (Priority 1)
**Goal:** Finish remaining actions for partially migrated tools  
**Timeline:** 18-23 days  
**Dependencies:** None (can start immediately)

### Sub-Phase 3B: Medium Complexity Tools (Priority 2)
**Goal:** Migrate tools with manageable complexity  
**Timeline:** 23-31 days  
**Dependencies:** None (can parallelize with 3A)

### Sub-Phase 3C: High Complexity Tools (Priority 3)
**Goal:** Migrate or document retention for very complex tools  
**Timeline:** 61-80 days  
**Dependencies:** Some tools may depend on 3A/3B completion

---

## Sub-Phase 3A: Complete Partial Implementations

**Priority:** 🔴 **HIGHEST** - Unblocks other work, builds on existing patterns  
**Timeline:** 18-23 days  
**Status:** 0/3 complete

### Why Start Here?

1. **Builds on Existing Work** - Extends proven patterns
2. **Unblocks Dependencies** - `task_analysis` blocks `task_discovery` orphans
3. **Quick Wins** - Partial implementations mean less new code
4. **Pattern Validation** - Validates hybrid approach works

### 3A.1: Complete `task_analysis` (8-10 days)

**Current Status:**
- ✅ `hierarchy` action - Native Go with Apple FM (complete)
- ⏳ `duplicates` action - Python bridge
- ⏳ `tags` action - Python bridge
- ⏳ `dependencies` action - Python bridge
- ⏳ `parallelization` action - Python bridge

**Implementation Plan:**

#### Step 1: `duplicates` Action (2-3 days)
- **Approach:** Native Go with optional MLX for semantic similarity
- **Dependencies:** Todo2 utilities (already available)
- **Implementation:**
  ```go
  // internal/tools/task_analysis_native.go
  func handleTaskAnalysisDuplicates(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
      // 1. Load all tasks using todo2_utils.go
      // 2. Compare tasks (exact match first)
      // 3. Optional: Use MLX for semantic similarity (if available)
      // 4. Return duplicate groups
  }
  ```
- **Files:** Extend `internal/tools/task_analysis_native.go`
- **Testing:** Unit tests for duplicate detection logic

#### Step 2: `tags` Action (2-3 days)
- **Approach:** Pure Native Go (no ML needed)
- **Dependencies:** Todo2 utilities
- **Implementation:**
  ```go
  func handleTaskAnalysisTags(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
      // 1. Load all tasks
      // 2. Extract all tags
      // 3. Analyze tag usage (frequency, co-occurrence)
      // 4. Suggest tag consolidation
  }
  ```
- **Files:** Extend `internal/tools/task_analysis_native.go`
- **Testing:** Test tag analysis and consolidation logic

#### Step 3: `dependencies` Action (2-3 days)
- **Approach:** Native Go with graph analysis
- **Dependencies:** Todo2 utilities, graph library (e.g., `gonum/graph`)
- **Implementation:**
  ```go
  func handleTaskAnalysisDependencies(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
      // 1. Load all tasks with dependencies
      // 2. Build dependency graph
      // 3. Detect cycles, missing dependencies
      // 4. Suggest dependency optimization
  }
  ```
- **Files:** Extend `internal/tools/task_analysis_native.go`
- **Dependencies:** May need `go get gonum.org/v1/gonum/graph`
- **Testing:** Test cycle detection, dependency validation

#### Step 4: `parallelization` Action (2-3 days)
- **Approach:** Native Go with graph analysis
- **Dependencies:** Todo2 utilities, graph library
- **Implementation:**
  ```go
  func handleTaskAnalysisParallelization(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
      // 1. Load tasks with dependencies
      // 2. Build dependency graph
      // 3. Identify parallel execution opportunities
      // 4. Suggest task grouping for parallel work
  }
  ```
- **Files:** Extend `internal/tools/task_analysis_native.go`
- **Testing:** Test parallelization detection logic

**Acceptance Criteria:**
- [ ] All 4 remaining actions implemented in native Go
- [ ] Handler routes to native implementation for all actions
- [ ] Python bridge fallback maintained for error cases
- [ ] Unit tests for each action
- [ ] Integration tests via MCP interface

### 3A.2: Complete `task_discovery` (4-5 days)

**Current Status:**
- ✅ `comments` action - Native Go with Apple FM (complete)
- ✅ `markdown` action - Native Go (complete)
- ⏳ `orphans` action - Python bridge (depends on task_analysis)

**Implementation Plan:**

#### Step 1: `orphans` Action (4-5 days)
- **Approach:** Native Go using task_analysis
- **Dependencies:** `task_analysis` must be complete first
- **Implementation:**
  ```go
  func handleTaskDiscoveryOrphans(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
      // 1. Load all tasks
      // 2. Use task_analysis to find tasks with no dependencies
      // 3. Check if tasks are referenced in code/docs
      // 4. Identify truly orphaned tasks
  }
  ```
- **Files:** Extend `internal/tools/task_discovery_native.go`
- **Testing:** Test orphan detection logic

**Acceptance Criteria:**
- [ ] `orphans` action implemented in native Go
- [ ] Uses task_analysis for dependency checking
- [ ] Integration tests pass

### 3A.3: Complete `task_workflow` (6-8 days)

**Current Status:**
- ✅ `clarify` action - Native Go with Apple FM (complete)
- ✅ `approve` action - Native Go (complete)
- ⏳ `sync` action - Python bridge
- ⏳ `clarity` action - Python bridge
- ⏳ `cleanup` action - Python bridge

**Implementation Plan:**

#### Step 1: `sync` Action (2-3 days)
- **Approach:** Native Go with Todo2 utilities
- **Dependencies:** Todo2 utilities
- **Implementation:**
  ```go
  func handleTaskWorkflowSync(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
      // 1. Load tasks from Todo2
      // 2. Sync with external systems (if any)
      // 3. Update task status
      // 4. Return sync results
  }
  ```
- **Files:** Extend `internal/tools/task_workflow_native.go`

#### Step 2: `clarity` Action (2-3 days)
- **Approach:** Native Go (may use Apple FM for improvement suggestions)
- **Dependencies:** Todo2 utilities, optional Apple FM
- **Implementation:**
  ```go
  func handleTaskWorkflowClarity(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
      // 1. Load task
      // 2. Analyze clarity (length, specificity, completeness)
      // 3. Optional: Use Apple FM for improvement suggestions
      // 4. Return clarity score and suggestions
  }
  ```
- **Files:** Extend `internal/tools/task_workflow_native.go`

#### Step 3: `cleanup` Action (2-3 days)
- **Approach:** Pure Native Go
- **Dependencies:** Todo2 utilities
- **Implementation:**
  ```go
  func handleTaskWorkflowCleanup(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
      // 1. Load all tasks
      // 2. Identify stale tasks (old, no activity)
      // 3. Suggest cleanup actions
      // 4. Optionally auto-cleanup if dry_run=false
  }
  ```
- **Files:** Extend `internal/tools/task_workflow_native.go`

**Acceptance Criteria:**
- [ ] All 3 remaining actions implemented
- [ ] Handler routes to native for all actions
- [ ] Tests pass

**Sub-Phase 3A Total:** 18-23 days

---

## Sub-Phase 3B: Medium Complexity Tools

**Priority:** 🟡 **MEDIUM** - Can parallelize with 3A  
**Timeline:** 23-31 days  
**Status:** 0/4 complete

### 3B.1: `session` Tool (7-9 days)

**Complexity:** Medium-High  
**Dependencies:** Todo2 utilities, Git status  
**Implementation:** Native Go

**Actions:**
- `prime` - Prepare session context
- `handoff` - Transfer session to another agent
- `prompts` - Get session prompts
- `assignee` - Manage task assignments

**Implementation Plan:**

```go
// internal/tools/session.go
func HandleSessionNative(ctx context.Context, params SessionParams) (string, error) {
    action := params.Action
    switch action {
    case "prime":
        return handleSessionPrime(ctx, params)
    case "handoff":
        return handleSessionHandoff(ctx, params)
    case "prompts":
        return handleSessionPrompts(ctx, params)
    case "assignee":
        return handleSessionAssignee(ctx, params)
    }
}
```

**Steps:**
1. Create `internal/tools/session.go` (2-3 days)
2. Implement `prime` action (2 days)
3. Implement `handoff` action (1-2 days)
4. Implement `prompts` action (1 day)
5. Implement `assignee` action (1-2 days)
6. Update handler (1 day)

**Dependencies:**
- Todo2 utilities (available)
- Git status (use `os/exec` for `git status`)

### 3B.2: `prompt_tracking` Tool (4-6 days)

**Complexity:** Medium  
**Dependencies:** Prompt log storage  
**Implementation:** Native Go for storage, Python bridge for complex analysis

**Actions:**
- `log` - Log prompt iteration
- `analyze` - Analyze prompt iterations

**Implementation Plan:**

```go
// internal/tools/prompt_tracking.go
func HandlePromptTrackingNative(ctx context.Context, params PromptTrackingParams) (string, error) {
    action := params.Action
    switch action {
    case "log":
        return handlePromptTrackingLog(ctx, params)
    case "analyze":
        // Try native first, fallback to Python for complex analysis
        return handlePromptTrackingAnalyze(ctx, params)
    }
}
```

**Steps:**
1. Create `internal/tools/prompt_tracking.go` (1 day)
2. Implement `log` action (1-2 days)
3. Implement basic `analyze` action (1-2 days)
4. Keep Python bridge for advanced analysis (1 day)
5. Update handler (1 day)

### 3B.3: `ollama` Tool (3-4 days)

**Complexity:** Medium  
**Dependencies:** Ollama HTTP API  
**Implementation:** Native Go HTTP client

**Actions:**
- `status` - Check Ollama status
- `models` - List available models
- `generate` - Generate text
- `pull` - Pull model
- `hardware` - Get hardware info
- `docs` - Get documentation
- `quality` - Quality check
- `summary` - Generate summary

**Implementation Plan:**

```go
// internal/tools/ollama.go
func HandleOllamaNative(ctx context.Context, params OllamaParams) (string, error) {
    // Use HTTP client to call Ollama API
    // No Python bridge needed - pure HTTP API
}
```

**Steps:**
1. Create `internal/tools/ollama.go` (1 day)
2. Implement HTTP client wrapper (1 day)
3. Implement all actions (1-2 days)
4. Update handler (1 day)

**Dependencies:**
- `net/http` (standard library)
- Ollama API documentation

### 3B.4: `mlx` Tool (3-4 days)

**Complexity:** Medium-High  
**Dependencies:** MLX bindings  
**Implementation:** Hybrid - Native Go if bindings available, Python bridge otherwise

**Actions:**
- `status` - Check MLX status
- `hardware` - Get hardware info
- `models` - List models
- `generate` - Generate text

**Implementation Plan:**

```go
// internal/tools/mlx.go
func HandleMlxNative(ctx context.Context, params MlxParams) (string, error) {
    // Check if Go bindings available
    // If yes: use native Go
    // If no: fallback to Python bridge
}
```

**Steps:**
1. Research MLX Go bindings (1 day)
2. If available: implement native Go (2-3 days)
3. If not: document Python bridge retention (1 day)
4. Update handler (1 day)

**Sub-Phase 3B Total:** 23-31 days

---

## Sub-Phase 3C: High Complexity Tools

**Priority:** 🟢 **LOW** - Complex, may require Python bridge retention  
**Timeline:** 61-80 days  
**Status:** 0/6 complete

### 3C.1: `memory` and `memory_maint` Tools (10-14 days)

**Complexity:** High  
**Dependencies:** Semantic search, embedding models  
**Implementation:** **HYBRID RECOMMENDED**

**Actions:**
- `memory`: `save`, `recall`, `search`
- `memory_maint`: `health`, `gc`, `prune`, `consolidate`, `dream`

**Implementation Strategy:**

#### Phase 1: CRUD Operations (Native Go) - 3-4 days
```go
// internal/tools/memory.go
func HandleMemoryNative(ctx context.Context, params MemoryParams) (string, error) {
    action := params.Action
    switch action {
    case "save":
        return handleMemorySave(ctx, params) // Native Go
    case "recall":
        return handleMemoryRecall(ctx, params) // Native Go
    case "search":
        return handleMemorySearch(ctx, params) // Hybrid: keyword native, semantic Python
    }
}
```

#### Phase 2: Semantic Search (Hybrid) - 4-5 days
- Native Go for keyword search
- Python bridge for semantic search (unless Go embedding library found)
- Evaluate: `github.com/qdrant/go-client` or similar

#### Phase 3: Maintenance Operations (Native Go) - 3-5 days
```go
// internal/tools/memory_maint.go
func HandleMemoryMaintNative(ctx context.Context, params MemoryMaintParams) (string, error) {
    // health, gc, prune, consolidate - all native Go
    // dream - may need Python bridge for AI generation
}
```

**Decision Point:** After Phase 1, evaluate Go embedding libraries. If good library found, implement semantic search in Go. Otherwise, keep Python bridge.

### 3C.2: `report` Tool (8-10 days)

**Complexity:** High  
**Dependencies:** Multiple data sources, report generation  
**Implementation:** Hybrid - Native Go for aggregation, Python bridge for complex templates

**Actions:**
- `overview` - Project overview
- `scorecard` - Project scorecard
- `briefing` - Daily briefing
- `prd` - PRD generation

**Implementation Strategy:**

#### Phase 1: Data Aggregation (Native Go) - 4-5 days
```go
// internal/tools/report.go
func HandleReportNative(ctx context.Context, params ReportParams) (string, error) {
    // Aggregate data from multiple sources
    // Use existing tools: git_tools, task_analysis, etc.
}
```

#### Phase 2: Report Formatting (Hybrid) - 4-5 days
- Native Go for simple text reports
- Python bridge for complex templates (if needed)
- Consider: Go templates for simple formatting

### 3C.3: `security` Tool (6-8 days)

**Complexity:** High  
**Dependencies:** Security scanning tools, vulnerability databases  
**Implementation:** **HYBRID RECOMMENDED**

**Actions:**
- `scan` - Scan dependencies
- `alerts` - Fetch Dependabot alerts
- `report` - Generate security report

**Implementation Strategy:**

```go
// internal/tools/security.go
func HandleSecurityNative(ctx context.Context, params SecurityParams) (string, error) {
    // Native Go for orchestration
    // Use os/exec for security scanners
    // Python bridge for complex vulnerability analysis
}
```

**Steps:**
1. Implement orchestration in Go (2-3 days)
2. Integrate security scanners via `os/exec` (2-3 days)
3. Keep Python bridge for complex analysis (2 days)

### 3C.4: `testing` Tool (7-9 days)

**Complexity:** Medium-High  
**Dependencies:** Test framework detection, coverage parsing  
**Implementation:** Hybrid - Native Go for orchestration, Python bridge for ML suggestions

**Actions:**
- `run` - Run tests
- `coverage` - Analyze coverage
- `suggest` - Suggest test cases (ML)
- `validate` - Validate test structure

**Implementation Strategy:**

```go
// internal/tools/testing.go
func HandleTestingNative(ctx context.Context, params TestingParams) (string, error) {
    action := params.Action
    switch action {
    case "run", "coverage", "validate":
        return handleTestingNativeActions(ctx, params) // Native Go
    case "suggest":
        return handleTestingSuggest(ctx, params) // Hybrid: Python bridge for ML
    }
}
```

### 3C.5: `estimation` Tool (6-8 days)

**Complexity:** Medium-High  
**Dependencies:** ML estimation (MLX), historical data  
**Implementation:** **HYBRID RECOMMENDED**

**Actions:**
- `estimate` - Estimate task duration
- `analyze` - Analyze estimation accuracy
- `stats` - Get estimation statistics

**Implementation Strategy:**

```go
// internal/tools/estimation.go
func HandleEstimationNative(ctx context.Context, params EstimationParams) (string, error) {
    // Native Go for historical data analysis
    // Python bridge for MLX integration (if no Go bindings)
}
```

### 3C.6: `automation` Tool (15-20 days OR Document Retention)

**Complexity:** Very High  
**Dependencies:** Complex workflow engine, parallel execution  
**Implementation:** **PYTHON BRIDGE RECOMMENDED**

**Decision:** **Document as intentional Python bridge retention**

**Rationale:**
- Core automation engine
- Very complex workflow system
- Benefits significantly from Python ecosystem
- Low ROI for migration

**Alternative (if migrating):**
- Partial native Go for simple workflows only
- Python bridge for complex workflows
- Timeline: 15-20 days

**Recommendation:** **Document retention, skip migration**

**Sub-Phase 3C Total:** 61-80 days (or 52-67 days if automation skipped)

---

## Execution Order & Dependencies

### Critical Path

```
3A.1 (task_analysis) → 3A.2 (task_discovery orphans)
     ↓
3A.3 (task_workflow) [parallel with 3B]
     ↓
3B (medium complexity) [parallel with 3C]
     ↓
3C (high complexity)
```

### Parallelization Opportunities

1. **3A.1 and 3A.3** - Can work in parallel (different tools)
2. **3B.1, 3B.2, 3B.3, 3B.4** - All independent, can parallelize
3. **3C tools** - Most are independent, can parallelize

### Recommended Execution Order

**Week 1-3: Sub-Phase 3A**
- Focus: Complete partial implementations
- Deliverables: All task tools fully native

**Week 4-7: Sub-Phase 3B (parallel with 3C start)**
- Focus: Medium complexity tools
- Deliverables: session, prompt_tracking, ollama, mlx

**Week 8-19: Sub-Phase 3C**
- Focus: High complexity tools
- Deliverables: memory, report, security, testing, estimation
- Decision: Document automation retention

---

## Risk Mitigation

### High-Risk Items

1. **Memory Semantic Search**
   - **Risk:** No good Go embedding library
   - **Mitigation:** Hybrid approach - keyword native, semantic Python
   - **Fallback:** Keep Python bridge for semantic search

2. **MLX Go Bindings**
   - **Risk:** No Go bindings available
   - **Mitigation:** Research first, document retention if none
   - **Fallback:** Python bridge for MLX

3. **Automation Complexity**
   - **Risk:** Too complex to migrate
   - **Mitigation:** Document as intentional retention
   - **Fallback:** Skip migration, keep Python bridge

4. **Report Templates**
   - **Risk:** Complex template logic
   - **Mitigation:** Use Go templates for simple, Python for complex
   - **Fallback:** Hybrid approach

### Success Criteria

- [ ] All partial implementations completed (3A)
- [ ] All medium complexity tools migrated (3B)
- [ ] High complexity tools evaluated and migrated where practical (3C)
- [ ] Automation tool documented as intentional Python bridge retention
- [ ] All migrated tools have tests
- [x] Hybrid patterns documented
- [ ] No regressions in functionality

---

## Timeline Summary

| Sub-Phase | Tools | Timeline | Status |
|-----------|-------|----------|--------|
| 3A: Complete Partial | 3 tools | 18-23 days | 0/3 (0%) |
| 3B: Medium Complexity | 4 tools | 23-31 days | 0/4 (0%) |
| 3C: High Complexity | 6 tools | 52-67 days* | 0/6 (0%) |
| **Total** | **13 tools** | **93-121 days** | **1/13 (8%)** |

*Excludes automation (15-20 days) if documented as retention

**With automation:** 108-141 days  
**Without automation:** 93-121 days

---

## Next Steps

1. **Start Sub-Phase 3A** - Begin with `task_analysis` duplicates action
2. **Research MLX Go Bindings** - Determine if native Go possible
3. **Evaluate Go Embedding Libraries** - For memory semantic search
4. **Document Automation Retention** - Create ADR for intentional Python bridge
5. **Set up Parallel Work** - Assign different tools to different developers (if team)

---

**Last Updated:** 2026-01-09  
**Status:** Ready for Execution  
**Next Action:** Start Sub-Phase 3A.1 - Complete task_analysis duplicates action

