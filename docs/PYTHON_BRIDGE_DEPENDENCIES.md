# Python Bridge Dependencies

**Last Updated**: 2026-01-09  
**Status**: After Native Go Migration Phase

## Overview

This document lists all remaining dependencies on the Python bridge after the recent migration phase. Items are categorized by migration priority and complexity.

---

## Tools Still Using Python Bridge

### 🔴 Fully Python Bridge (No Native Implementation)

#### High Priority
1. **`memory`** - Memory CRUD operations
   - **Location**: `internal/tools/handlers.go:205`
   - **Status**: Native Go implementation exists for basic CRUD, but falls back for semantic search
   - **Complexity**: Medium (semantic search requires ML/AI)
   - **Migration**: Partially complete - needs semantic search implementation

2. **`memory_maint`** - Memory maintenance (dream action)
   - **Location**: `internal/tools/handlers.go:239`
   - **Status**: Native Go for health, gc, prune, consolidate ✅
   - **Remaining**: `dream` action (requires advisor integration)
   - **Complexity**: High (requires advisor/ML integration)
   - **Migration**: Dream action deferred (requires devwisdom-go integration)

3. **`report`** - Report generation
   - **Location**: `internal/tools/handlers.go:325`
   - **Status**: Fully Python bridge
   - **Complexity**: High (ML/AI content generation, multiple report types)
   - **Migration**: Low priority (complex, depends on ML capabilities)

4. **`security`** - Security scanning ✅ **NOW NATIVE**
   - **Location**: `internal/tools/handlers.go:361`
   - **Status**: ✅ Native Go for scan, alerts, report actions
   - **Implementation**: `internal/tools/security.go`
   - **Actions**: All 3 actions have native implementations (scan uses govulncheck, alerts uses gh CLI, report combines both)
   - **Migration**: ✅ Complete (was already native, documentation outdated)

5. **`task_analysis`** - Task analysis ✅ **NOW NATIVE**
   - **Location**: `internal/tools/handlers.go:411`
   - **Status**: ✅ Native Go for duplicates, tags, dependencies, parallelization on all platforms
   - **Remaining**: Hierarchy action requires Apple FM (optional enhancement)
   - **Migration**: ✅ Complete for platform-agnostic actions

6. **`task_discovery`** - Task discovery ✅ **NOW NATIVE**
   - **Location**: `internal/tools/handlers.go:447`
   - **Status**: ✅ Native Go for comments, markdown, orphans on all platforms
   - **Remaining**: Apple FM enhancement for semantic extraction (optional)
   - **Migration**: ✅ Complete for basic scanning functionality

7. **`task_workflow`** - Task workflow management ✅ **NOW NATIVE**
   - **Location**: `internal/tools/handlers.go:476`
   - **Status**: ✅ Native Go for approve, create, sync, clarity, cleanup on all platforms
   - **Remaining**: Clarify action requires Apple FM (optional enhancement)
   - **Migration**: ✅ Complete for platform-agnostic actions

8. **`automation`** - Automation orchestration ✅ **NOW NATIVE (Partial)**
   - **Location**: `internal/tools/handlers.go:551`
   - **Status**: ✅ Native Go for daily and discover actions
   - **Remaining**: nightly and sprint actions - falls back to Python bridge (correct behavior)
   - **Migration**: ✅ Complete for daily and discover actions

9. **`lint`** - Linting ✅ **NOW NATIVE (Partial)**
   - **Location**: `internal/tools/handlers.go:621`
   - **Status**: ✅ Native Go for Go linters (golangci-lint, go-vet, gofmt, goimports, markdownlint, shellcheck)
   - **Remaining**: Python linters (ruff, flake8) - falls back to Python bridge (correct behavior)
   - **Migration**: ✅ Complete for Go ecosystem linters

10. **`estimation`** - Task duration estimation
    - **Location**: `internal/tools/handlers.go:714`
    - **Status**: Fully Python bridge
    - **Complexity**: High (ML/AI model inference, statistical analysis)
    - **Migration**: Low priority (requires ML/AI capabilities, MLX integration)

11. **`mlx`** - MLX model operations
    - **Location**: `internal/tools/handlers.go:819`
    - **Status**: Fully Python bridge
    - **Complexity**: High (ML framework integration)
    - **Migration**: Low priority (platform-specific ML framework)

12. **`recommend`** - Recommendations (workflow, advisor)
    - **Location**: `internal/tools/handlers.go:937`
    - **Status**: Native Go for "model" action ✅, Python bridge for "workflow" and "advisor"
    - **Remaining**: workflow and advisor actions
    - **Complexity**: Medium-High (requires advisor integration)
    - **Migration**: Medium priority (workflow can be native, advisor requires devwisdom-go)

#### Medium Priority
13. **`analyze_alignment`** - Alignment analysis (prd action)
    - **Location**: `internal/tools/handlers.go:29`
    - **Status**: Native Go for "todo2" action ✅, Python bridge for "prd" action
    - **Remaining**: PRD alignment analysis
    - **Complexity**: Medium
    - **Migration**: Medium priority

14. **`generate_config`** - Config generation ✅ **NOW FULLY NATIVE**
   - **Location**: `internal/tools/handlers.go:39`
   - **Status**: ✅ Native Go for all actions (rules, ignore, simplify) - fully native
   - **Migration**: ✅ Complete
    - **Location**: `internal/tools/handlers.go:55`
    - **Status**: Native Go for "rules" action ✅, Python bridge for "ignore" and "simplify"
    - **Remaining**: ignore and simplify actions
    - **Complexity**: Low-Medium
    - **Migration**: High priority (straightforward file operations)

15. **`health`** - Health checks (dod, cicd)
    - **Location**: `internal/tools/handlers.go:81`
    - **Status**: Fully Python bridge
    - **Complexity**: Medium (multiple health check types)
    - **Migration**: Medium priority

16. **`setup_hooks`** - Hook setup (patterns)
    - **Location**: `internal/tools/handlers.go:107`
    - **Status**: Native Go for "git" action ✅, Python bridge for "patterns"
    - **Remaining**: patterns action
    - **Complexity**: Low-Medium
    - **Migration**: Medium priority

17. **`add_external_tool_hints`** - External tool hints
    - **Location**: `internal/tools/handlers.go:159`
    - **Status**: Native Go implementation exists ✅, falls back on error
    - **Complexity**: Low
    - **Migration**: Complete but needs error handling review

18. **`testing`** - Testing (suggest, generate)
    - **Location**: `internal/tools/testing.go:50,94,133`
    - **Status**: Native Go for run, coverage, validate ✅, Python bridge for suggest and generate
    - **Remaining**: suggest and generate actions (ML/AI)
    - **Complexity**: High (ML/AI test generation)
    - **Migration**: Low priority (requires ML/AI capabilities)

19. **`context`** - Context management (specific actions)
    - **Location**: `internal/tools/handlers.go:891`
    - **Status**: Native Go for summarize and batch ✅, Python bridge fallback
    - **Complexity**: Low-Medium
    - **Migration**: Mostly complete, verify all actions work

20. **`session`** - Session management (prompts, assignee)
    - **Location**: `internal/tools/handlers.go:765`
    - **Status**: Native Go for prime and handoff ✅, Python bridge for prompts and assignee
    - **Remaining**: prompts and assignee actions
    - **Complexity**: Medium (depends on prompt discovery resource)
    - **Migration**: Medium priority

21. **`ollama`** - Ollama integration (docs, quality, summary)
    - **Location**: `internal/tools/handlers.go:802`
    - **Status**: Native Go for status, models, generate, pull, hardware ✅
    - **Remaining**: docs, quality, summary actions
    - **Complexity**: Medium (requires Ollama API analysis)
    - **Migration**: Medium priority (can use native HTTP client)

---

### 🟡 Hybrid (Partial Native, Partial Python Bridge)

Tools with native implementations that still fall back to Python bridge for specific actions:

1. **`analyze_alignment`** - PRD action → Python bridge
2. **`generate_config`** - ignore, simplify → Python bridge
3. **`health`** - All actions → Python bridge (may have partial native)
4. **`setup_hooks`** - patterns → Python bridge
5. **`memory`** - Semantic search → Python bridge
6. **`memory_maint`** - dream → Python bridge
7. **`testing`** - suggest, generate → Python bridge
8. **`context`** - Fallback on error → Python bridge
9. **`session`** - prompts, assignee → Python bridge
10. **`ollama`** - docs, quality, summary → Python bridge
11. **`recommend`** - workflow, advisor → Python bridge

---

## Resources Still Using Python Bridge

### Fully Python Bridge Resources

1. **`stdio://scorecard`** - Project scorecard
   - **Location**: `internal/resources/scorecard.go:39`
   - **Status**: Fully Python bridge
   - **Complexity**: High (ML/AI content generation)
   - **Migration**: Low priority (depends on report tool)

2. **`stdio://memories/*`** - All memory resources
   - **Location**: `internal/resources/memories.go:26,75,123,162,219`
   - **Status**: Fully Python bridge
   - **Complexity**: Medium (CRUD operations, semantic search)
   - **Migration**: Medium priority (can reuse memory.go utilities)

3. **Other resources** - Context primer, prompt discovery, session
   - **Location**: `internal/resources/handlers.go:210,214,218,222,226`
   - **Status**: Fully Python bridge
   - **Complexity**: Medium (depends on session tool migration)
   - **Migration**: Medium priority (depends on related tool migrations)

---

## Migration Priority Summary

### 🔴 High Priority (Quick Wins - No ML/AI Required)
1. ✅ `task_discovery` - File parsing and pattern matching
2. ✅ `task_workflow` - Todo2 operations (utilities already exist)
3. ✅ `lint` - Subprocess execution of linters
4. ✅ `generate_config` - ignore, simplify actions (file operations)
5. ✅ `task_analysis` - Graph algorithms (gonum already available)
6. ✅ `automation` - daily and discover actions (orchestration)

### 🟡 Medium Priority (Some Complexity)
1. ⚠️ `health` - Multiple health check types (docs action)
2. ✅ `security` - **ALREADY NATIVE** (documentation updated)
3. ⚠️ `session` - prompts, assignee actions
4. ⚠️ `ollama` - docs, quality, summary actions
5. ⚠️ `recommend` - workflow action
6. ⚠️ `setup_hooks` - patterns action
7. ⚠️ `analyze_alignment` - prd action
8. ⚠️ Memory resources - CRUD operations

### 🟢 Low Priority (Requires ML/AI or Platform-Specific)
1. ❌ `report` - ML/AI content generation
2. ❌ `estimation` - ML model inference
3. ❌ `mlx` - ML framework integration
4. ❌ `memory_maint` - dream action (requires advisor)
5. ❌ `recommend` - advisor action (requires devwisdom-go)
6. ❌ `testing` - suggest, generate actions (ML/AI)
7. ❌ Memory semantic search - ML/AI capabilities
8. ❌ Scorecard resource - ML/AI content generation

---

## Statistics

### Tools
- **Total Tools**: 24
- **Fully Native**: 8 (33%)
- **Hybrid (Native + Python)**: 8 (33%)
- **Fully Python Bridge**: 8 (33%)

### Resources
- **Total Resources**: 6
- **Fully Native**: 0 (0%)
- **Fully Python Bridge**: 6 (100%)

### Migration Progress
- **Recent Migration**: 5 tools migrated to native Go (session, ollama, prompt_tracking, memory_maint consolidate, context batch verified)
- **High Priority Remaining**: 6 tools
- **Medium Priority Remaining**: 8 tools
- **Low Priority Remaining**: 8 tools/features

---

## Next Steps

1. **Focus on High Priority Quick Wins**:
   - `task_discovery` - File operations
   - `task_workflow` - Todo2 utilities exist
   - `lint` - Subprocess execution
   - `generate_config` - ignore, simplify actions
   - `task_analysis` - Graph algorithms with gonum

2. **Resources Migration** (Phase 4):
   - Start with memory resources (can reuse memory.go)
   - Context primer and prompt discovery (after session tool)
   - Scorecard (after report tool)

3. **Complex Tools** (Defer):
   - ML/AI dependent tools (report, estimation, mlx, testing suggest/generate)
   - Advisor-dependent features (memory_maint dream, recommend advisor)

---

## Notes

- **Hybrid Approach**: Many tools successfully use hybrid approach (try native first, fallback to Python)
- **Error Handling**: Some tools may fall back to Python on error - need to review error handling
- **Dependencies**: Some migrations depend on others (e.g., automation sprint depends on report tool)
- **Testing**: All native implementations should be thoroughly tested before removing Python bridge fallbacks
