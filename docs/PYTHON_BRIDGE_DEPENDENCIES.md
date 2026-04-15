# Python Bridge Dependencies

**Last Updated**: 2026-01-12  
**Status**: After Comprehensive Audit and Documentation Updates (Stream 4)  
**Audit Reference**: See `docs/MIGRATION_AUDIT_2026-01-12.md`

## Overview

This document lists all remaining dependencies on the Python bridge after the comprehensive audit. Items are categorized by migration priority and complexity. **Note:** Most tools use a hybrid approach (native first, Python bridge fallback).

---

## Tools Using Python Bridge

### 🔴 Fully Python Bridge Only (1 tool)

Tools with no native implementation (intentional retention):

1. **`mlx`** - MLX model operations
   - **Location**: `internal/tools/handlers.go:791`
   - **Status**: Fully Python bridge (no Go bindings available)
   - **Rationale**: MLX is a Python ML framework with no Go bindings - intentional retention
   - **Complexity**: High (ML framework integration)
   - **Migration**: ❌ Not recommended (no Go bindings available)
   - **Decision**: Documented as intentional Python bridge retention

---

### 🟡 Hybrid Tools (22 tools - Native + Python Bridge Fallback)

Tools that try native Go first, fallback to Python bridge for specific actions or errors:

#### ✅ Completed - All Core Actions Native (Recent Completions)

1. **`analyze_alignment`** - Alignment analysis
   - **Location**: `internal/tools/handlers.go:15`
   - **Native Actions**: `todo2`, `prd` ✅
   - **Python Bridge Actions**: Fallback on error
   - **Status**: Hybrid - Native for all actions, bridge fallback
   - **Completion**: ✅ Complete (prd action migrated 2026-01-12)
   - **Migration**: ✅ Complete

2. **`health`** - Health checks
   - **Location**: `internal/tools/handlers.go:67`
   - **Native Actions**: `server`, `git`, `docs`, `dod`, `cicd` ✅
   - **Python Bridge Actions**: Fallback on error
   - **Status**: Hybrid - Native for all actions, bridge fallback
   - **Completion**: ✅ Complete (docs, dod, cicd actions migrated 2026-01-12)
   - **Migration**: ✅ Complete

3. **`setup_hooks`** - Hook setup
   - **Location**: `internal/tools/handlers.go:93`
   - **Native Actions**: `git`, `patterns` ✅
   - **Python Bridge Actions**: Fallback on error
   - **Status**: Hybrid - Native for all actions, bridge fallback
   - **Completion**: ✅ Complete (patterns action migrated 2026-01-12)
   - **Migration**: ✅ Complete

4. **`session`** - Session management
   - **Location**: `internal/tools/handlers.go:721`
   - **Native Actions**: `prime`, `handoff`, `prompts`, `assignee` ✅
   - **Python Bridge Actions**: Fallback on error
   - **Status**: Hybrid - Native for all actions, bridge fallback
   - **Completion**: ✅ Complete (prompts and assignee actions migrated 2026-01-12)
   - **Migration**: ✅ Complete


6. **`report`** - Report generation
   - **Location**: `internal/tools/handlers.go:249`
   - **Native Actions**: `scorecard`, `overview`, `prd` ✅
   - **Python Bridge Actions**: `briefing` (devwisdom MCP)
   - **Status**: Hybrid - Native for scorecard/overview/prd, bridge for briefing
   - **Remaining**: briefing action (requires devwisdom-go MCP server)
   - **Complexity**: High (requires devwisdom-go integration)
   - **Migration**: Low priority (briefing requires devwisdom-go MCP server)

#### Medium Priority - Most Actions Native, Fallback on Error

7. **`memory`** - Memory CRUD operations
   - **Location**: `internal/tools/handlers.go:171`
   - **Native Actions**: `save`, `recall`, `list` ✅
   - **Python Bridge Actions**: Semantic search (advanced search action)
   - **Status**: Hybrid - Native for CRUD, bridge for semantic search
   - **Remaining**: Semantic search (ML/AI feature)
   - **Complexity**: High (requires ML/AI semantic search)
   - **Migration**: Low priority (requires ML/AI capabilities)

8. **`memory_maint`** - Memory maintenance
   - **Location**: `internal/tools/handlers.go:215`
   - **Native Actions**: `health`, `gc`, `prune`, `consolidate` ✅
   - **Python Bridge Actions**: `dream` (requires advisor integration)
   - **Status**: Hybrid - Native for health/gc/prune/consolidate, bridge for dream
   - **Remaining**: dream action (requires advisor integration)
   - **Complexity**: High (requires advisor/ML integration)
   - **Migration**: Low priority (dream requires devwisdom-go integration)

9. **`automation`** - Automation orchestration
   - **Location**: `internal/tools/handlers.go:551`
   - **Native Actions**: `daily`, `discover`, `nightly`, `sprint` ✅
   - **Python Bridge Actions**: Fallback on error
   - **Status**: Hybrid - Native for all actions, bridge fallback
   - **Completion**: ✅ Complete (nightly and sprint actions migrated 2026-01-12)
   - **Migration**: ✅ Complete

10. **`testing`** - Testing operations
    - **Location**: `internal/tools/handlers.go:495`
    - **Native Actions**: `run`, `coverage`, `validate` (Go projects) ✅
    - **Python Bridge Actions**: Other languages, ML features
    - **Status**: Hybrid - Native for Go projects, bridge for other languages/ML
    - **Remaining**: Non-Go language support, ML test generation
    - **Complexity**: Medium-High (requires multi-language support, ML for test generation)
    - **Migration**: Low priority (non-Go support requires language-specific tooling)

11. **`security`** - Security scanning
    - **Location**: `internal/tools/handlers.go:362`
    - **Native Actions**: `scan`, `alerts`, `report` (Go projects) ✅
    - **Python Bridge Actions**: Other languages
    - **Status**: Hybrid - Native for Go projects, bridge for other languages
    - **Remaining**: Non-Go language support
    - **Complexity**: Medium (requires language-specific security scanners)
    - **Migration**: Low priority (non-Go support requires language-specific tooling)

12. **`lint`** - Linting
    - **Location**: `internal/tools/handlers.go:600`
    - **Native Actions**: Go linters (golangci-lint, go-vet, gofmt, goimports, markdownlint, shellcheck) ✅
    - **Python Bridge Actions**: Python linters (ruff, flake8)
    - **Status**: Hybrid - Native for Go linters, bridge for Python linters
    - **Remaining**: Python linter support
    - **Complexity**: Low-Medium (subprocess execution)
    - **Migration**: Medium priority (can add Python linter support if needed)

#### Low Priority - Native with Optional Bridge Fallback

13. **`check_attribution`** - Attribution checking
    - **Location**: `internal/tools/handlers.go:119`
    - **Native Actions**: All actions ✅
    - **Python Bridge Actions**: Fallback on error
    - **Status**: Hybrid - Native primary, bridge fallback
    - **Complexity**: Low
    - **Migration**: ✅ Complete (fallback is safety measure)

14. **`add_external_tool_hints`** - External tool hints
    - **Location**: `internal/tools/handlers.go:143`
    - **Native Actions**: All actions ✅
    - **Python Bridge Actions**: Fallback on error
    - **Status**: Hybrid - Native primary, bridge fallback
    - **Complexity**: Low
    - **Migration**: ✅ Complete (fallback is safety measure)

15. **`generate_config`** - Config generation
    - **Location**: `internal/tools/handlers.go:39`
    - **Native Actions**: All actions (rules, ignore, simplify) ✅
    - **Python Bridge Actions**: Fallback on error (likely unused)
    - **Status**: Hybrid - Native for all actions, bridge fallback
    - **Complexity**: Low
    - **Migration**: ✅ Complete (fallback is safety measure)

16. **`task_analysis`** - Task analysis
    - **Location**: `internal/tools/handlers.go:411`
    - **Native Actions**: All actions (duplicates, tags, dependencies, parallelization, hierarchy) ✅
    - **Python Bridge Actions**: Fallback on error (hierarchy requires Apple FM, falls back if unavailable)
    - **Status**: Hybrid - Native for all actions, bridge fallback
    - **Complexity**: Low
    - **Migration**: ✅ Complete (fallback for Apple FM unavailability)

17. **`task_discovery`** - Task discovery
    - **Location**: `internal/tools/handlers.go:439`
    - **Native Actions**: All actions (comments, markdown, orphans) ✅
    - **Python Bridge Actions**: Fallback on error (Apple FM enhancement optional)
    - **Status**: Hybrid - Native for all actions, bridge fallback
    - **Complexity**: Low
    - **Migration**: ✅ Complete (fallback is safety measure)

18. **`task_workflow`** - Task workflow management
    - **Location**: `internal/tools/handlers.go:467`
    - **Native Actions**: All actions (sync, approve, clarify, clarity, cleanup) ✅
    - **Python Bridge Actions**: Fallback on error (clarify requires Apple FM, falls back if unavailable)
    - **Status**: Hybrid - Native for all actions, bridge fallback
    - **Complexity**: Low
    - **Migration**: ✅ Complete (fallback for Apple FM unavailability)

19. **`ollama`** - Ollama integration
   - **Location**: `internal/tools/handlers.go:766`
   - **Native Actions**: All actions (status, models, generate, pull, hardware, docs, quality, summary) ✅
   - **Python Bridge Actions**: Fallback on error
   - **Status**: Hybrid - Native for all actions via HTTP client, bridge fallback
   - **Completion**: ✅ Complete (docs, quality, summary actions migrated 2026-01-12)
   - **Migration**: ✅ Complete (fallback is safety measure)

20. **`context`** - Context management
   - **Location**: `internal/tools/handlers.go:815`
   - **Native Actions**: `summarize` (with Apple FM), `budget`, `batch` ✅
   - **Python Bridge Actions**: Fallback on error or when Apple FM unavailable
   - **Status**: Hybrid - Native for all actions, bridge fallback
   - **Complexity**: Low-Medium
   - **Migration**: ✅ Complete (fallback for Apple FM unavailability)

21. **`estimation`** - Task duration estimation
   - **Location**: `internal/tools/handlers.go:671`
   - **Native Actions**: `estimate`, `stats`, `analyze` ✅
   - **Python Bridge Actions**: MLX inference, fallback on error
   - **Status**: Hybrid - Native for all actions (estimate, stats, analyze), bridge for MLX inference (optional)
   - **Completion**: ✅ Complete (analyze action migrated 2026-01-12)
   - **Migration**: ✅ Complete (MLX inference is optional enhancement)

22. **`recommend`** - Recommendations
   - **Location**: `internal/tools/handlers.go:892`
   - **Native Actions**: `model`, `workflow` ✅
   - **Python Bridge Actions**: `advisor` (requires devwisdom-go), fallback on error
   - **Status**: Hybrid - Native for model/workflow, bridge for advisor
   - **Completion**: ✅ Complete (workflow action migrated 2026-01-12)
   - **Migration**: Low priority (advisor requires devwisdom-go MCP server - intentional bridge usage)

---

## Resources Using Python Bridge

### 🟡 Hybrid Resources (1 resource)

Resources with native implementation that fallback to Python bridge:

1. **`stdio://session/mode`** - Session mode resource
   - **Location**: `internal/resources/handlers.go:377`
   - **Status**: Native Go (primary), Python bridge (fallback)
   - **Implementation**: Uses `tools.HandleInferSessionModeNative`
   - **Fallback**: Python bridge if native fails
   - **Migration**: ✅ Complete (fallback is safety measure)

---

### ✅ Native Resources (20 resources - No Python Bridge)

All other resources are fully native (no Python bridge calls):

**Memory Resources (5):**
- `stdio://memories` - Native Go
- `stdio://memories/category/{category}` - Native Go
- `stdio://memories/recent` - Native Go
- `stdio://memories/session/{date}` - Native Go
- `stdio://memories/task/{task_id}` - Native Go

**Prompt Resources (4):**
- `stdio://prompts` - Native Go (uses getAllPromptsNative)
- `stdio://prompts/category/{category}` - Native Go
- `stdio://prompts/mode/{mode}` - Native Go
- `stdio://prompts/persona/{persona}` - Native Go

**Task Resources (6):**
- `stdio://tasks` - Native Go
- `stdio://tasks/priority/{priority}` - Native Go
- `stdio://tasks/status/{status}` - Native Go
- `stdio://tasks/summary` - Native Go
- `stdio://tasks/tag/{tag}` - Native Go
- `stdio://tasks/{task_id}` - Native Go

**Tool Resources (2):**
- `stdio://tools` - Native Go
- `stdio://tools/{category}` - Native Go

**Other Resources (3):**
- `stdio://scorecard` - Native Go (uses GenerateGoScorecard for Go projects, bridge fallback for non-Go)
- `stdio://models` - Native Go
- `stdio://server/status` - Native Go

---

## Migration Priority Summary

### ✅ Completed (All High-Priority Actions Complete!)

**Tools:**
- All tools have native implementations (5 fully native, 22 hybrid)
- Only 1 tool (`mlx`) is intentionally Python bridge only
- All high-priority actions completed (Stream 1, 2, 3)

**Resources:**
- All 21 resources have native implementations
- 21 resources are native (100%)
- All resources use native Go as primary implementation

### ✅ Recent Completions (2026-01-12)

**Stream 1, 2, 3 Actions Completed:**
1. ✅ **`session`** - `prompts`, `assignee` actions now native
2. ✅ **`ollama`** - `docs`, `quality`, `summary` actions now native
3. ✅ **`recommend`** - `workflow` action now native
4. ✅ **`setup_hooks`** - `patterns` action now native
5. ✅ **`analyze_alignment`** - `prd` action now native
6. ✅ **`health`** - `docs`, `dod`, `cicd` actions now native
7. ✅ **`estimation`** - `analyze` action now native
8. ✅ **`automation`** - `nightly`, `sprint` actions now native

### 🟢 Remaining Actions (Low Priority - ML/AI or External Dependencies)

### 🟢 Low Priority (ML/AI or External Dependencies Required)

1. **`mlx`** - ❌ Intentional Python bridge (no Go bindings)
2. **`memory`** - Semantic search (ML/AI)
3. **`memory_maint`** - `dream` action (requires devwisdom-go advisor)
4. **`recommend`** - `advisor` action (requires devwisdom-go)
5. **`report`** - `briefing` action (requires devwisdom-go)
6. **`testing`** - Non-Go language support, ML test generation
7. **`security`** - Non-Go language support
8. **`estimation`** - MLX inference (intentional bridge usage)

---

## Statistics

### Tools
- **Total Tools**: 28 (plus 1 conditional Apple FM tool = 28-29)
- **Fully Native**: 5 tools (18%) - `tool_catalog`, `workflow_mode`, `git_tools`, `infer_session_mode`, `prompt_tracking`
- **Hybrid (Native + Python)**: 22 tools (79%) - Native primary with Python bridge fallback
- **Fully Python Bridge**: 1 tool (4% - `mlx` only, intentional - no Go bindings)
- **Native Coverage**: 96% (27/28 tools have native implementations)

### Resources
- **Total Resources**: 21
- **Fully Native**: 21 resources (100%) - All resources use native Go as primary
- **Hybrid (Native + Python)**: 0 resources (0%)
- **Fully Python Bridge**: 0 resources (0%)
- **Native Coverage**: 100% (21/21 resources have native implementations)

---

## Key Findings from Audit (2026-01-12)

1. ✅ **Excellent Progress**: 96% tool coverage, 100% resource coverage
2. ✅ **Hybrid Pattern Dominant**: 79% of tools use hybrid pattern (native first, bridge fallback)
3. ✅ **All High-Priority Actions Complete**: Stream 1, 2, 3 actions all migrated (session, ollama, recommend, setup_hooks, analyze_alignment, health, automation, estimation)
4. ✅ **All Resources Native**: All 21 resources have native implementations (100%)
5. ✅ **Only 1 Intentional Python Bridge Tool**: `mlx` is the only tool intentionally Python bridge only (no Go bindings)
6. ✅ **Remaining Work Is Low Priority**: Most remaining actions are ML/AI features or require external dependencies (devwisdom-go, MLX)

---

## Next Steps

### ✅ High-Priority Actions Complete!

All high-priority actions from Stream 1, 2, and 3 have been completed (2026-01-12):
1. ✅ **`session` tool** - `prompts` and `assignee` actions
2. ✅ **`health` tool** - `docs`, `dod`, `cicd` actions
3. ✅ **`setup_hooks` tool** - `patterns` action
4. ✅ **`analyze_alignment` tool** - `prd` action
5. ✅ **`automation` tool** - `nightly`, `sprint` actions
6. ✅ **`ollama` tool** - `docs`, `quality`, `summary` actions
7. ✅ **`recommend` tool** - `workflow` action
8. ✅ **`estimation` tool** - `analyze` action

### Deferred Actions (ML/AI or External Dependencies)

1. **`mlx`** - ❌ Keep Python bridge (no Go bindings)
2. **`memory` semantic search** - Requires ML/AI
3. **`memory_maint` dream** - Requires devwisdom-go advisor
4. **`recommend` advisor** - Requires devwisdom-go
5. **`report` briefing** - Requires devwisdom-go
6. **`testing` non-Go/ML** - Requires language-specific tooling/ML
7. **`security` non-Go** - Requires language-specific scanners
8. **`estimation` MLX** - Intentional bridge usage for MLX inference

---

## Notes

- **Hybrid Approach Works Well**: 75% of tools successfully use hybrid approach (try native first, fallback to Python)
- **Fallback is Safety Measure**: Many "Python Bridge" entries are actually fallbacks that rarely trigger
- **External Dependencies**: Some features (advisor, briefing) require devwisdom-go MCP server - intentional bridge usage
- **ML/AI Features**: Semantic search, test generation, etc. intentionally use Python bridge for ML capabilities
- **Platform-Specific**: MLX intentionally uses Python bridge (no Go bindings)
- **Testing**: All native implementations should be thoroughly tested, but Python bridge fallback provides safety net

---

**Last Audit**: 2026-01-12 (See `docs/MIGRATION_AUDIT_2026-01-12.md`)  
**Last Updated**: 2026-01-12 (After Stream 1, 2, 3 completions)  
**Next Review**: After completing testing and documentation updates
