# Go SDK Migration Tasks Summary

**Date:** 2025-01-01  
**Status:** Tasks Created (Todo2 system issue - manual tracking recommended)

---

## MCP Server Configuration Status

### Current Status: ⚠️ **NOT CONFIGURED**

**Finding:**
- ❌ No `.cursor/mcp.json` in `/Users/davidl/Projects/mcp-stdio-tools`
- ❌ MCP server not configured in parent project (`project-management-automation`)
- ✅ README shows expected configuration format

**Required Configuration:**
```json
{
  "mcpServers": {
    "mcp-stdio-tools": {
      "command": "{{PROJECT_ROOT}}/../mcp-stdio-tools/run_server.sh",
      "args": [],
      "env": {
        "PROJECT_ROOT": "{{PROJECT_ROOT}}"
      },
      "description": "Stdio-based MCP tools (broken in FastMCP) - 24 tools, 8 prompts, 6 resources"
    }
  }
}
```

**Action Required:**
- Create `.cursor/mcp.json` in parent project or this project
- Update command path after Go migration to point to Go binary

---

## Migration Tasks Created

### Phase 1: Foundation (High Priority)

#### T-1: Phase 1: Go Project Setup & Foundation
**Priority:** High  
**Tags:** migration, go-sdk, foundation, phase-1  
**Dependencies:** None

**Objectives:**
- Set up Go project structure
- Install Go SDK dependencies
- Create basic server skeleton
- Implement Python bridge mechanism
- Framework abstraction interfaces

**Key Deliverables:**
- `exarp-go/go.mod`
- `cmd/server/main.go`
- `internal/framework/server.go` (interfaces)
- `internal/bridge/python.go`
- `bridge/execute_tool.py`

**Acceptance Criteria:**
- ✅ Go module initialized with go-sdk
- ✅ Server starts successfully
- ✅ Python bridge executes test tool
- ✅ STDIO transport works

---

#### T-2: Phase 1: Framework-Agnostic Design Implementation
**Priority:** High  
**Tags:** migration, go-sdk, framework-agnostic, architecture, phase-1  
**Dependencies:** T-1

**Objectives:**
- Implement framework abstraction layer
- Create adapters for go-sdk, mcp-go, go-mcp
- Factory pattern for framework selection
- Configuration-based switching
- Error handling (per MLX analysis)

**Key Deliverables:**
- `internal/framework/server.go` (interfaces)
- `internal/framework/adapters/gosdk/adapter.go`
- `internal/framework/adapters/mcpgo/adapter.go` (optional)
- `internal/framework/factory.go`
- `internal/config/config.go`
- `config.yaml`

**Acceptance Criteria:**
- ✅ Interface compliance tests pass
- ✅ Can switch frameworks via config
- ✅ Error handling implemented
- ✅ Transport types implemented

---

### Phase 2: Tool Migration - Batch 1 (High Priority)

#### T-3: Phase 2: Batch 1 Tool Migration (6 Simple Tools)
**Priority:** High  
**Tags:** migration, go-sdk, tools, phase-2  
**Dependencies:** T-1, T-2

**Tools to Migrate:**
1. `analyze_alignment`
2. `generate_config`
3. `health`
4. `setup_hooks`
5. `check_attribution`
6. `add_external_tool_hints`

**Key Deliverables:**
- `internal/tools/registry.go`
- `internal/tools/handlers.go`
- Updated `bridge/execute_tool.py`

**Acceptance Criteria:**
- ✅ All 6 tools registered and working
- ✅ Tools execute via Python bridge
- ✅ All tools testable in Cursor

---

### Phase 3: Tool Migration - Batch 2 (High Priority)

#### T-4: Phase 3: Batch 2 Tool Migration (8 Medium Tools)
**Priority:** High  
**Tags:** migration, go-sdk, tools, prompts, phase-3  
**Dependencies:** T-3

**Tools to Migrate:**
1. `memory`
2. `memory_maint`
3. `report`
4. `security`
5. `task_analysis`
6. `task_discovery`
7. `task_workflow`
8. `testing`

**Prompts to Migrate:**
- `align`, `discover`, `config`, `scan`, `scorecard`, `overview`, `dashboard`, `remember`

**Key Deliverables:**
- `internal/prompts/registry.go`
- Updated tool registry
- Prompt template loading

**Acceptance Criteria:**
- ✅ All 8 tools working
- ✅ All 8 prompts accessible
- ✅ Cursor integration tested

---

### Phase 4: Tool Migration - Batch 3 (High Priority)

#### T-5: Phase 4: Batch 3 Tool Migration (8 Advanced Tools)
**Priority:** High  
**Tags:** migration, go-sdk, tools, resources, phase-4  
**Dependencies:** T-4

**Tools to Migrate:**
1. `automation`
2. `tool_catalog`
3. `workflow_mode`
4. `lint`
5. `estimation`
6. `git_tools`
7. `session`
8. `infer_session_mode`

**Resources to Migrate:**
- `stdio://scorecard`
- `stdio://memories`
- `stdio://memories/category/{category}`
- `stdio://memories/task/{task_id}`
- `stdio://memories/recent`
- `stdio://memories/session/{date}`

**Key Deliverables:**
- `internal/resources/handlers.go`
- Updated tool registry
- Resource URI handling

**Acceptance Criteria:**
- ✅ All 8 tools working
- ✅ All 6 resources accessible
- ✅ Cursor integration tested

---

### Phase 5: MLX Integration (Medium Priority)

#### T-6: Phase 5: MLX Integration & Special Tools
**Priority:** Medium  
**Tags:** migration, go-sdk, mlx, apple-silicon, phase-5  
**Dependencies:** T-5

**Tools to Handle:**
- `ollama` - May work via HTTP API
- `mlx` - Requires Apple Silicon, Python bridge

**Key Deliverables:**
- MLX Python bridge support
- Ollama HTTP client (if using API)
- Apple Silicon detection

**Acceptance Criteria:**
- ✅ MLX tools execute
- ✅ Apple Silicon compatibility verified
- ✅ Fallback for non-Apple Silicon

---

### Phase 6: Testing & Documentation (High Priority)

#### T-7: Phase 6: Testing, Optimization & Documentation
**Priority:** High  
**Tags:** migration, go-sdk, testing, documentation, phase-6  
**Dependencies:** T-6

**Key Deliverables:**
- Unit tests (>80% coverage)
- Integration tests
- Performance benchmarks
- Complete documentation
- Migration guide
- Deployment guide

**Acceptance Criteria:**
- ✅ All components tested
- ✅ Performance meets targets
- ✅ Documentation complete
- ✅ Production deployment ready

---

### Configuration Setup (Medium Priority)

#### T-8: MCP Server Configuration Setup
**Priority:** Medium  
**Tags:** migration, mcp-config, cursor, setup  
**Dependencies:** T-1 (for binary path)

**Key Deliverables:**
- `.cursor/mcp.json` configuration
- Updated README with config instructions
- Cursor integration verified

**Acceptance Criteria:**
- ✅ MCP config file created
- ✅ Server connects in Cursor
- ✅ All tools/prompts/resources accessible

---

## Task Dependencies Graph

```
T-1 (Foundation Setup)
  ├── T-2 (Framework-Agnostic Design)
  │     └── T-3 (Batch 1 Tools)
  │           └── T-4 (Batch 2 Tools)
  │                 └── T-5 (Batch 3 Tools)
  │                       └── T-6 (MLX Integration)
  │                             └── T-7 (Testing & Docs)
  └── T-8 (MCP Configuration)
```

---

## Timeline Estimate

| Phase | Tasks | Duration | Priority |
|-------|-------|----------|----------|
| Phase 1 | T-1, T-2 | 3-5 days | High |
| Phase 2 | T-3 | 2-3 days | High |
| Phase 3 | T-4 | 3-4 days | High |
| Phase 4 | T-5 | 3-4 days | High |
| Phase 5 | T-6 | 2-3 days | Medium |
| Phase 6 | T-7 | 3-4 days | High |
| Config | T-8 | 1 day | Medium |
| **Total** | **8 tasks** | **17-24 days** | |

---

## Next Steps

1. ✅ **Review tasks** - Verify all tasks are correctly defined
2. ✅ **Set up MCP configuration** - Create `.cursor/mcp.json` (T-8)
3. 🔬 **Start Phase 1** - Begin with T-1 (Go project setup)
4. 📝 **Add research comments** - Document research before implementation
5. 🚀 **Begin implementation** - Start with foundation tasks

---

## MCP Configuration Action Items

### Immediate Action Required:

1. **Create MCP Configuration File:**
   ```bash
   mkdir -p /Users/davidl/Projects/mcp-stdio-tools/.cursor
   # Or in parent project:
   mkdir -p /Users/davidl/Projects/project-management-automation/.cursor
   ```

2. **Add Configuration:**
   ```json
   {
     "mcpServers": {
       "mcp-stdio-tools": {
         "command": "{{PROJECT_ROOT}}/../mcp-stdio-tools/run_server.sh",
         "args": [],
         "env": {
           "PROJECT_ROOT": "{{PROJECT_ROOT}}"
         },
         "description": "Stdio-based MCP tools - 24 tools, 8 prompts, 6 resources"
       }
     }
   }
   ```

3. **After Go Migration:**
   - Update command to: `{{PROJECT_ROOT}}/../exarp-go/bin/exarp-go`
   - Remove Python dependencies from description

---

## Notes

- **Todo2 System:** Tasks created but system showing ID issues (T-null, T-NaN)
- **Manual Tracking:** May need to track tasks manually or fix Todo2 system
- **MCP Config:** Not currently configured - needs setup (Task T-8)
- **Research Required:** Each task needs research_with_links comment before "In Progress"

---

**Status:** Ready to begin Phase 1 implementation  
**Blockers:** None (MCP config can be set up in parallel)

