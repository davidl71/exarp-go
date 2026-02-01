# Resource & Prompt Migration Plan

**Date:** 2026-01-12  
**Status:** Ready for Implementation

## Executive Summary

**Prompts:** ✅ **Already fully native!**
- All 18 prompts are stored in Go map in `internal/prompts/templates.go`
- `GetPromptTemplate()` retrieves prompts from Go (no Python bridge)

**Resources:** ⚠️ **Mostly native, 5 remaining**
- 6 memory resources: ✅ Native (with Python fallback)
- Scorecard resource: ✅ Native for Go projects (with Python fallback)
- Server, models, tools, tasks resources: ✅ All native
- **5 prompt-related resources: ❌ Still Python bridge**

---

## Migration Status

### ✅ Fully Native Resources (16 resources)

**Memory Resources (6):**
1. `stdio://memories` - Native Go
2. `stdio://memories/category/{category}` - Native Go
3. `stdio://memories/task/{task_id}` - Native Go
4. `stdio://memories/recent` - Native Go
5. `stdio://memories/session/{date}` - Native Go
6. `stdio://scorecard` - Native Go (for Go projects)

**Other Resources (10):**
7. `stdio://server/status` - Native Go
8. `stdio://models` - Native Go
9. `stdio://tools` - Native Go
10. `stdio://tools/{category}` - Native Go
11. `stdio://tasks` - Native Go
12. `stdio://tasks/{task_id}` - Native Go
13. `stdio://tasks/status/{status}` - Native Go
14. `stdio://tasks/priority/{priority}` - Native Go
15. `stdio://tasks/tag/{tag}` - Native Go
16. `stdio://tasks/summary` - Native Go

### ❌ Still Python Bridge Resources (5 resources)

**Prompt Resources (4):**
1. `stdio://prompts` - Python bridge (should use `prompts.GetPromptTemplate()`)
2. `stdio://prompts/mode/{mode}` - Python bridge (needs mode mapping)
3. `stdio://prompts/persona/{persona}` - Python bridge (needs persona mapping)
4. `stdio://prompts/category/{category}` - Python bridge (needs category mapping)

**Session Resource (1):**
5. `stdio://session/mode` - Python bridge (should use `infer_session_mode` tool)

---

## Implementation Plan

### Phase 1: Prompt Resources (Quick Win)

**1. Migrate `handleAllPrompts`** - Easiest (no filtering needed)

**Implementation:**
```go
func handleAllPrompts(ctx context.Context, uri string) ([]byte, string, error) {
    // Get all prompt names from templates map
    promptNames := []string{
        "align", "discover", "config", "scan", "scorecard", "overview",
        "dashboard", "remember", "daily_checkin", "sprint_start", "sprint_end",
        "pre_sprint", "post_impl", "sync", "dups", "context", "mode", "task_update",
    }
    
    // Build compact format (name + description only)
    prompts := make([]map[string]interface{}, len(promptNames))
    for i, name := range promptNames {
        template, _ := prompts.GetPromptTemplate(name)
        // Extract first line as description (before first newline)
        description := extractFirstLine(template)
        prompts[i] = map[string]interface{}{
            "name": name,
            "description": description,
        }
    }
    
    result := map[string]interface{}{
        "prompts": prompts,
        "count": len(prompts),
        "timestamp": time.Now().Format(time.RFC3339),
    }
    
    jsonData, err := json.Marshal(result)
    return jsonData, "application/json", err
}
```

**2. Migrate `handlePromptsByMode`** - Requires mode mapping

**Prompt Mode Mapping:**
- `daily_checkin` → mode: "daily_checkin"
- `sprint_start`, `sprint_end`, `pre_sprint`, `post_impl` → mode: "sprint_management"
- `sync`, `dups` → mode: "task_management"
- `align`, `discover`, `task_update` → mode: "task_management"
- `config` → mode: "configuration"
- `scan` → mode: "security_review"
- `scorecard`, `overview`, `dashboard` → mode: "reporting"
- `remember` → mode: "memory"
- `context`, `mode` → mode: "utility"

**3. Migrate `handlePromptsByPersona`** - Requires persona mapping

**Prompt Persona Mapping:**
- `daily_checkin`, `sprint_start`, `sprint_end`, `pre_sprint`, `post_impl` → persona: "pm"
- `align`, `discover`, `task_update`, `sync`, `dups` → persona: "developer"
- `config` → persona: "developer"
- `scan` → persona: "security"
- `scorecard`, `overview`, `dashboard` → persona: "pm"
- `remember` → persona: "developer"
- `context`, `mode` → persona: "developer"

**4. Migrate `handlePromptsByCategory`** - Requires category mapping

**Prompt Category Mapping:**
- `align`, `discover`, `task_update`, `sync`, `dups` → category: "task_management"
- `daily_checkin`, `sprint_start`, `sprint_end`, `pre_sprint`, `post_impl` → category: "workflow"
- `config` → category: "configuration"
- `scan` → category: "security"
- `scorecard`, `overview`, `dashboard` → category: "reporting"
- `remember` → category: "memory"
- `context`, `mode` → category: "utility"

### Phase 2: Session Mode Resource

**5. Migrate `handleSessionMode`** - Use `infer_session_mode` tool

**Implementation:**
```go
func handleSessionMode(ctx context.Context, uri string) ([]byte, string, error) {
    // Call infer_session_mode tool (which is already native Go)
    // This requires access to the tool registry or direct function call
    
    // Option 1: Call infer_session_mode native function directly
    mode, confidence, err := inferSessionMode(ctx)
    if err != nil {
        return bridge.ExecutePythonResource(ctx, uri) // Fallback
    }
    
    result := map[string]interface{}{
        "mode": mode,
        "confidence": confidence,
        "timestamp": time.Now().Format(time.RFC3339),
    }
    
    jsonData, err := json.Marshal(result)
    return jsonData, "application/json", err
}
```

**Note:** `infer_session_mode` is already a native Go tool. We need to either:
- Export the underlying function for direct use
- Or create a shared function that both the tool and resource can use

---

## File Structure

### New File: `internal/resources/prompts.go`

Contains all prompt resource handlers:
- `handleAllPrompts` - List all prompts
- `handlePromptsByMode` - Filter by mode
- `handlePromptsByPersona` - Filter by persona
- `handlePromptsByCategory` - Filter by category
- Helper functions for mode/persona/category mapping

### Modify: `internal/resources/handlers.go`

Update handlers to call native implementations:
```go
// Change from:
func handleAllPrompts(ctx context.Context, uri string) ([]byte, string, error) {
    return bridge.ExecutePythonResource(ctx, uri)
}

// To:
func handleAllPrompts(ctx context.Context, uri string) ([]byte, string, error) {
    return handleAllPromptsNative(ctx, uri)
}
```

### New File: `internal/resources/session.go`

Contains session mode resource handler:
- `handleSessionMode` - Uses native `infer_session_mode` implementation

---

## Testing Strategy

1. **Unit Tests:**
   - Test `handleAllPrompts` returns all 18 prompts
   - Test filtering by mode/persona/category
   - Test session mode inference

2. **Integration Tests:**
   - Test resource responses match Python bridge format
   - Test URI parsing for parameterized resources
   - Test error handling and fallbacks

3. **Comparison Tests:**
   - Compare Go output with Python bridge output
   - Ensure JSON format compatibility

---

## Dependencies

**Current:**
- `internal/prompts/templates.go` - Prompt template storage ✅
- `internal/prompts/registry.go` - Prompt registration ✅
- `internal/tools/infer_session_mode.go` - Session mode inference (need to check if function is exported)

**New:**
- Prompt metadata mapping (mode/persona/category)
- Helper functions for prompt filtering

---

## Migration Order

1. ✅ **Phase 1a:** Migrate `handleAllPrompts` (easiest, no filtering)
2. ⚠️ **Phase 1b:** Migrate prompt filtering resources (requires metadata mapping)
3. ⚠️ **Phase 2:** Migrate `handleSessionMode` (requires access to infer_session_mode function)

---

## Expected Results

**After Migration:**
- ✅ All 21 resources will be native Go
- ✅ All 18 prompts remain native (already are)
- ✅ Zero Python bridge calls for resources
- ✅ Zero Python bridge calls for prompts

**Python Bridge Remaining:**
- Tool execution (`bridge/execute_tool.py`) - Still needed for 19 tools
- Resource execution (`bridge/execute_resource.py`) - Can be removed or simplified

---

## Notes

1. **Prompt Templates:** Already fully native in Go map - no migration needed
2. **Prompt Resources:** Need to build metadata mappings for filtering
3. **Session Mode:** Can leverage existing native `infer_session_mode` tool
4. **Backward Compatibility:** Ensure JSON format matches Python bridge output

---

## Priority

**High Priority:** 
- `handleAllPrompts` - Quick win, no dependencies
- `handleSessionMode` - Simple, uses existing native tool

**Medium Priority:**
- Prompt filtering resources - Requires metadata mapping design

---

## Completion Criteria

- [ ] All prompt resources migrated to native Go
- [ ] Session mode resource migrated to native Go
- [ ] All resources have no Python bridge dependency
- [ ] Tests pass for all resources
- [ ] JSON output format matches Python bridge (for compatibility)
- [ ] Documentation updated
