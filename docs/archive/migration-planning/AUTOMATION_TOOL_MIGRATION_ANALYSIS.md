# Automation Tool Migration Analysis

**Date**: 2026-01-09  
**Task**: T-104  
**Tool**: `automation`  
**Status**: Analysis Complete

## Tool Overview

The `automation` tool is a **high-level orchestration tool** that coordinates multiple other tools to perform complex workflows. It supports 4 actions:

1. **daily** - Daily maintenance workflow
2. **nightly** - Nightly task processing workflow
3. **sprint** - Sprint automation workflow
4. **discover** - Task discovery automation

## Action Complexity Analysis

### 1. `daily` Action

**Purpose**: Daily maintenance and health checks

**Orchestration** (from Python implementation):
- Runs `analyze_alignment` (todo2 alignment)
- Runs `task_analysis` (duplicate detection)
- Runs `health` (documentation health)
- Executes tools in sequence with error handling
- Collects and reports results

**Complexity**: **Medium**
- ✅ Most tools it calls are already native Go (or have native implementations)
- ✅ Orchestration logic is straightforward (sequential execution)
- ✅ No ML/AI dependencies
- ⚠️ Requires error aggregation and reporting

**Migration Feasibility**: **HIGH** ✅
- Can be implemented natively in Go
- Uses existing native Go tools (task_analysis, health, analyze_alignment)
- Simple orchestration pattern

### 2. `nightly` Action

**Purpose**: Nightly task processing and automation

**Orchestration**:
- Processes background tasks
- Runs automated workflows
- Task completion detection
- May involve task updates and status changes

**Complexity**: **Medium-High**
- ✅ Uses Todo2 utilities (already native Go)
- ✅ Uses task_workflow tool (already native Go)
- ⚠️ May involve complex task filtering and batch operations
- ⚠️ Requires careful error handling for batch operations

**Migration Feasibility**: **MEDIUM-HIGH** ✅
- Most dependencies are native Go
- Batch operations are well-understood
- Requires careful testing of task update workflows

### 3. `sprint` Action

**Purpose**: Sprint automation workflow

**Orchestration**:
- Coordinates multiple tools for sprint planning/execution
- Task alignment, discovery, estimation
- May involve report generation

**Complexity**: **High**
- ✅ Uses existing native tools
- ⚠️ Complex orchestration logic
- ⚠️ May involve report generation (requires report tool)
- ⚠️ Cross-tool coordination

**Migration Feasibility**: **MEDIUM** ⚠️
- Depends on report tool migration status
- Complex orchestration but feasible
- Requires thorough testing

### 4. `discover` Action

**Purpose**: Task discovery automation

**Orchestration**:
- Uses `task_discovery` tool
- Finds tasks from various sources (comments, markdown, etc.)
- Creates tasks automatically

**Complexity**: **Medium**
- ✅ Uses task_discovery tool (may already have native implementation)
- ✅ Task creation is straightforward
- ⚠️ File parsing and pattern matching

**Migration Feasibility**: **HIGH** ✅
- Mostly depends on task_discovery tool (which is already native Go)
- Simple orchestration

## Migration Strategy Recommendation

### Recommendation: **HYBRID APPROACH** (Recommended)

**Rationale**:
- **High-value actions** (daily, discover) can be migrated to native Go quickly
- **Complex actions** (sprint) may benefit from keeping Python bridge for now
- **Nightly action** is feasible but requires careful testing

### Implementation Plan

#### Phase 1: Quick Wins (Native Go)
1. ✅ **daily** action - Migrate to native Go
   - Simple orchestration
   - All dependencies are native Go
   - Low risk, high value

2. ✅ **discover** action - Migrate to native Go
   - Simple orchestration
   - Depends on task_discovery (native Go)
   - Low risk

#### Phase 2: Medium Complexity (Native Go with Fallback)
3. ⚠️ **nightly** action - Migrate with fallback
   - Implement native Go version
   - Keep Python bridge as fallback
   - Test thoroughly before removing fallback

#### Phase 3: Complex Actions (Retain Python Bridge or Full Migration)
4. ❓ **sprint** action - Decision needed
   - **Option A**: Migrate to native Go (requires report tool)
   - **Option B**: Keep Python bridge for now (lower priority)
   - **Recommendation**: Migrate after report tool is native Go

### Code Structure

```go
// internal/tools/automation.go

func handleAutomationNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
    action := getAction(params)
    
    switch action {
    case "daily":
        return handleAutomationDaily(ctx, params)
    case "nightly":
        return handleAutomationNightly(ctx, params)
    case "sprint":
        // Fall back to Python bridge (complex, depends on report tool)
        return nil, fmt.Errorf("sprint action requires report tool, falling back to Python bridge")
    case "discover":
        return handleAutomationDiscover(ctx, params)
    default:
        return nil, fmt.Errorf("unknown action: %s", action)
    }
}
```

## Dependencies

### Native Go Tools (Available):
- ✅ `task_analysis` - Native Go
- ✅ `health` - Native Go (docs health check)
- ✅ `analyze_alignment` - Native Go
- ✅ `task_discovery` - Native Go
- ✅ `task_workflow` - Native Go

### Still Using Python Bridge:
- ⚠️ `report` - Used by sprint action (not yet migrated)
- ⚠️ `automation` itself - Still Python bridge

## Risk Assessment

| Action | Risk Level | Migration Priority | Estimated Effort |
|--------|------------|-------------------|------------------|
| daily  | Low        | High              | 2-4 hours        |
| discover | Low      | High              | 2-4 hours        |
| nightly | Medium    | Medium            | 4-8 hours        |
| sprint | High      | Low (after report) | 8-16 hours      |

## Conclusion

**Recommendation**: **Migrate daily and discover actions first** (quick wins), then evaluate nightly and sprint based on report tool migration status.

**Benefits**:
- ✅ Removes Python bridge dependency for most common use case (daily)
- ✅ Maintains functionality while reducing complexity
- ✅ Allows incremental migration approach

**Next Steps**:
1. Implement native Go `daily` action
2. Implement native Go `discover` action
3. Test thoroughly
4. Update handler to use native implementation with Python fallback for sprint
5. Migrate sprint after report tool is native Go
