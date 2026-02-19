# TUI MCP Adapter Design

**Created:** 2026-02-20  
**Status:** Design Phase  
**Related Task:** T-1771543666094356000

## Overview

Design for routing TUI through MCP tools instead of direct database access. This establishes a single, consistent data access layer across CLI, TUI, and MCP interfaces.

---

## Current Architecture (Problem)

```
┌─────────┐     ┌──────────┐     ┌──────────┐
│ MCP CLI │────▶│ MCP Tools│────▶│ Database │
└─────────┘     └──────────┘     └──────────┘

┌─────────┐                      ┌──────────┐
│   CLI   │─────────────────────▶│ Database │
└─────────┘                      └──────────┘

┌─────────┐                      ┌──────────┐
│   TUI   │─────────────────────▶│ Database │
└─────────┘                      └──────────┘
```

**Issues:**
- 3 independent data paths
- Logic duplication across surfaces
- Hard to add features (must update 3 places)
- Testing complexity
- Inconsistent behavior possible

---

## Proposed Architecture (Solution)

```
┌─────────┐     ┌──────────┐     ┌──────────┐
│ MCP CLI │────▶│          │     │          │
└─────────┘     │          │     │          │
                │   MCP    │────▶│ Database │
┌─────────┐     │  Tools   │     │          │
│   CLI   │────▶│  (Single │     │          │
└─────────┘     │  Source  │     │          │
                │    of    │     │          │
┌─────────┐     │  Truth)  │     │          │
│   TUI   │────▶│          │     │          │
└─────────┘     └──────────┘     └──────────┘
         │              ▲
         │              │
         └──────────────┘
         TUI MCP Adapter
         (Typed helpers)
```

**Benefits:**
- Single capability API
- Add features in one place
- Consistent behavior guaranteed
- Easier testing (mock tool layer)
- Clean separation of concerns

---

## Current TUI Database Access

### Analysis Results

**Direct database calls in TUI:** 38 instances

**Operations by type:**

| Operation | Count | Files |
|-----------|------:|-------|
| `GetTasksByStatus` | 6 | tui_commands.go, tui3270.go |
| `GetTask` | 4 | tui_commands.go, tui_update.go |
| `UpdateTask` | 4 | tui_commands.go |
| `Close` | 2 | tui.go, tui3270.go |
| Type refs | ~20 | All TUI files |

**Tools package calls:** ~20 instances
- `ComputeWavesForTUI` (4)
- `GenerateGoScorecard` (2)
- `GetOverviewText` (2)
- Others

**Config package calls:** ~15 instances
- `LoadConfig`, `WriteConfigToProtobufFile`, etc.

---

## Design: TUI MCP Adapter Layer

### File: `internal/cli/tui_mcp_adapter.go`

Typed helper functions that wrap `server.CallTool()` with TUI-friendly interfaces.

### Core Principle

**Before:**
```go
tasks, err := database.GetTasksByStatus(ctx, "Todo")
```

**After:**
```go
tasks, err := m.loadTasksByStatus(ctx, "Todo")
```

**Implementation:**
```go
func (m *model) loadTasksByStatus(ctx context.Context, status string) ([]*database.Todo2Task, error) {
    result, err := m.server.CallTool(ctx, "task_workflow", map[string]interface{}{
        "action": "sync",
        "sub_action": "list",
        "status_filter": status,
        "output_format": "json",
    })
    // Parse result and return typed tasks
}
```

---

## Implementation Plan

### Phase 1: Adapter Layer Foundation

**File:** `internal/cli/tui_mcp_adapter.go`

```go
package cli

import (
    "context"
    "encoding/json"
    
    "github.com/davidl71/exarp-go/internal/database"
    "github.com/davidl71/exarp-go/internal/framework"
)

// tuiMCPAdapter provides typed wrappers around MCP tool calls for TUI.
type tuiMCPAdapter struct {
    server framework.MCPServer
    ctx    context.Context
}

// Task operations
func (a *tuiMCPAdapter) ListTasks(status string) ([]*database.Todo2Task, error)
func (a *tuiMCPAdapter) GetTask(taskID string) (*database.Todo2Task, error)
func (a *tuiMCPAdapter) UpdateTask(task *database.Todo2Task) error
func (a *tuiMCPAdapter) UpdateTaskStatus(taskID, newStatus string) error
func (a *tuiMCPAdapter) BulkUpdateTaskStatus(taskIDs []string, newStatus string) (int, error)

// Config operations
func (a *tuiMCPAdapter) LoadConfig() (*config.FullConfig, error)
func (a *tuiMCPAdapter) SaveConfig(cfg *config.FullConfig) error

// Scorecard operations
func (a *tuiMCPAdapter) GetScorecard(projectRoot string, fullMode bool) (string, []string, error)

// Handoff operations
func (a *tuiMCPAdapter) ListHandoffs() ([]map[string]interface{}, error)
func (a *tuiMCPAdapter) CloseHandoffs(ids []string) (int, error)
func (a *tuiMCPAdapter) ApproveHandoffs(ids []string) (int, error)

// Wave operations
func (a *tuiMCPAdapter) ComputeWaves(projectRoot string, tasks []database.Todo2Task) (map[int][]string, error)

// Analysis operations
func (a *tuiMCPAdapter) RunTaskAnalysis(action string) (string, error)
```

### Phase 2: Replace Task Operations

**Files to update:**
- `tui_commands.go` - loadTasks, updateTaskStatusCmd, bulkUpdateStatusCmd
- `tui_update.go` - Any direct task operations

**Pattern:**
```go
// Old
func loadTasks(status string) tea.Cmd {
    return func() tea.Msg {
        ctx := context.Background()
        tasks, err := database.GetTasksByStatus(ctx, status)
        // ...
    }
}

// New
func (m model) loadTasksViaMCP(status string) tea.Cmd {
    return func() tea.Msg {
        adapter := &tuiMCPAdapter{server: m.server, ctx: context.Background()}
        tasks, err := adapter.ListTasks(status)
        // ...
    }
}
```

### Phase 3: Replace Config Operations

**Files:**
- `tui_config.go` - Config loading/saving
- `tui.go` - Initial config load

### Phase 4: Replace Scorecard Operations

**Files:**
- `tui_scorecard.go` - Scorecard generation

### Phase 5: Replace Handoff Operations

**Files:**
- `tui_handoffs.go` - Handoff list/close/approve

### Phase 6: Replace Wave Operations

**Files:**
- `tui_waves.go` - Wave computation

---

## Detailed Method Signatures

### Task Operations

```go
// ListTasks retrieves tasks filtered by status via task_workflow tool.
// status: "Todo", "In Progress", "Done", "Review", or "" for all
func (a *tuiMCPAdapter) ListTasks(status string) ([]*database.Todo2Task, error) {
    args := map[string]interface{}{
        "action": "sync",
        "sub_action": "list",
        "output_format": "json",
    }
    if status != "" {
        args["status_filter"] = status
    }
    
    result, err := a.server.CallTool(a.ctx, "task_workflow", args)
    if err != nil {
        return nil, fmt.Errorf("task_workflow call failed: %w", err)
    }
    
    // Parse JSON result
    var response struct {
        Tasks []*database.Todo2Task `json:"tasks"`
    }
    if err := json.Unmarshal([]byte(result.Content[0].Text), &response); err != nil {
        return nil, fmt.Errorf("parse task list: %w", err)
    }
    
    return response.Tasks, nil
}

// UpdateTaskStatus updates a task's status via task_workflow tool.
func (a *tuiMCPAdapter) UpdateTaskStatus(taskID, newStatus string) error {
    args := map[string]interface{}{
        "action": "update",
        "task_id": taskID,
        "new_status": newStatus,
    }
    
    _, err := a.server.CallTool(a.ctx, "task_workflow", args)
    return err
}

// BulkUpdateTaskStatus updates multiple tasks' status.
// Returns count of successfully updated tasks.
func (a *tuiMCPAdapter) BulkUpdateTaskStatus(taskIDs []string, newStatus string) (int, error) {
    args := map[string]interface{}{
        "action": "approve",  // Uses approve flow for batch updates
        "task_ids": strings.Join(taskIDs, ","),
        "new_status": newStatus,
    }
    
    result, err := a.server.CallTool(a.ctx, "task_workflow", args)
    if err != nil {
        return 0, err
    }
    
    // Parse result for count
    // Implementation depends on task_workflow response format
    return len(taskIDs), nil  // Simplified
}
```

### Config Operations

```go
// LoadConfig loads configuration via generate_config tool.
func (a *tuiMCPAdapter) LoadConfig() (*config.FullConfig, error) {
    // Config loading might need to stay as direct call
    // since it's not exposed via MCP tool currently
    // OR: extend generate_config tool to support "get" action
    
    // Option 1: Direct call (if no tool support)
    return config.LoadConfig(a.projectRoot)
    
    // Option 2: Add "get" action to generate_config tool
    result, err := a.server.CallTool(a.ctx, "generate_config", map[string]interface{}{
        "action": "get",
        "output_format": "json",
    })
    // Parse and return
}
```

### Scorecard Operations

```go
// GetScorecard retrieves project scorecard via report tool.
func (a *tuiMCPAdapter) GetScorecard(projectRoot string, fullMode bool) (string, []string, error) {
    args := map[string]interface{}{
        "action": "scorecard",
        "output_format": "text",
        "fast_mode": !fullMode,
    }
    
    result, err := a.server.CallTool(a.ctx, "report", args)
    if err != nil {
        return "", nil, err
    }
    
    // Parse scorecard text and recommendations
    text := result.Content[0].Text
    
    // Extract recommendations (implementation specific)
    recs := []string{}
    // ... parse recommendations from text
    
    return text, recs, nil
}
```

---

## Migration Strategy

### Step-by-Step Approach

1. ✅ **Create adapter file** - `tui_mcp_adapter.go`
2. **Implement adapter methods** - One operation type at a time
3. **Update commands** - Replace in `tui_commands.go` first
4. **Update handlers** - Replace in `tui_update*.go`
5. **Update views** - Replace in view-specific files
6. **Test incrementally** - After each operation type
7. **Remove direct imports** - Clean up unused database imports

### Testing Strategy

**Unit tests for adapter:**
```go
func TestTuiMCPAdapter_ListTasks(t *testing.T) {
    mockServer := &mockMCPServer{
        response: `{"tasks": [{"id": "T-123", "content": "Test"}]}`,
    }
    
    adapter := &tuiMCPAdapter{server: mockServer, ctx: context.Background()}
    tasks, err := adapter.ListTasks("Todo")
    
    assert.NoError(t, err)
    assert.Len(t, tasks, 1)
    assert.Equal(t, "T-123", tasks[0].ID)
}
```

**Integration tests:**
- Test TUI with real MCP server
- Verify all operations work end-to-end

---

## Risk Mitigation

### Risks

1. **MCP tool API changes** - Tool interfaces might not match TUI needs
2. **Performance** - Extra serialization overhead
3. **Error handling** - Tool errors vs database errors
4. **Breaking changes** - Large refactor touching many files

### Mitigations

1. **Incremental migration** - One operation type at a time
2. **Keep old code** - Comment out, don't delete initially
3. **Comprehensive testing** - Test after each phase
4. **Feature flags** - Allow switching between paths during migration
5. **Rollback plan** - Git branches, easy revert

---

## Success Criteria

- ✅ Zero direct `database.*` calls from TUI files
- ✅ All operations go through MCP adapter
- ✅ All 13 TUI tests still passing
- ✅ TUI functionality unchanged (user-facing)
- ✅ Clean separation of concerns
- ✅ Easy to add new features (via MCP tools)

---

## Effort Estimate

| Phase | Effort | Risk |
|-------|--------|------|
| 1. Adapter foundation | 2 hours | Low |
| 2. Task operations | 3 hours | Medium |
| 3. Config operations | 1 hour | Low |
| 4. Scorecard operations | 1 hour | Low |
| 5. Handoff operations | 2 hours | Medium |
| 6. Wave operations | 1 hour | Low |
| 7. Testing & cleanup | 2 hours | Low |
| **Total** | **12 hours** | **Medium** |

**Recommended:** Dedicate 1.5-2 days for careful implementation and testing.

---

## Next Steps

1. **Review this design** with team
2. **Create adapter file** with method stubs
3. **Start with task operations** (highest impact)
4. **Test thoroughly** after each phase
5. **Document any tool API gaps** found during implementation
6. **Update this document** with learnings

---

## References

- **Task**: T-1771543666094356000
- **TUI Review**: `docs/TUI_REVIEW.md` Section 3
- **MCP Tools**: `internal/tools/registry.go`
- **Current TUI**: `internal/cli/tui*.go`

## Notes

- Config and some operations might need tool API extensions
- Performance impact should be measured
- Consider caching strategy for frequently accessed data
- May expose gaps in current MCP tool coverage
