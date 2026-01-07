# Coordinator vs exarp-go Duplication Analysis

**Date:** 2026-01-07  
**Status:** ⚠️ **DUPLICATION DETECTED**

---

## Summary

There is **significant duplication** between the `coordinator` (exarp_pma) and `exarp-go` servers. The exarp-go server now contains **all tools** from coordinator plus the original 12 migrated tools.

---

## Tool Comparison

### Coordinator (exarp_pma - FastMCP Server)
**Location:** `/Users/davidl/Projects/project-management-automation`  
**Tools:** 14 total

1. `add_external_tool_hints`
2. `automation`
3. `check_attribution`
4. `demonstrate_elicit` ⭐ **UNIQUE**
5. `estimation`
6. `git_tools`
7. `infer_session_mode`
8. `interactive_task_create` ⭐ **UNIQUE**
9. `lint`
10. `mlx`
11. `ollama`
12. `session`
13. `tool_catalog`
14. `workflow_mode`

### exarp-go (mcp-stdio-tools - Go Server)
**Location:** `/Users/davidl/Projects/mcp-stdio-tools`  
**Tools:** 24 total

**Originally Migrated (12):**
1. `analyze_alignment` ⭐ **UNIQUE**
2. `generate_config` ⭐ **UNIQUE**
3. `health` ⭐ **UNIQUE**
4. `memory` ⭐ **UNIQUE**
5. `memory_maint` ⭐ **UNIQUE**
6. `report` ⭐ **UNIQUE**
7. `security` ⭐ **UNIQUE**
8. `setup_hooks` ⭐ **UNIQUE**
9. `task_analysis` ⭐ **UNIQUE**
10. `task_discovery` ⭐ **UNIQUE**
11. `task_workflow` ⭐ **UNIQUE**
12. `testing` ⭐ **UNIQUE**

**Newly Migrated (12) - DUPLICATED:**
13. `add_external_tool_hints` ⚠️ **DUPLICATE**
14. `automation` ⚠️ **DUPLICATE**
15. `check_attribution` ⚠️ **DUPLICATE**
16. `estimation` ⚠️ **DUPLICATE**
17. `git_tools` ⚠️ **DUPLICATE**
18. `infer_session_mode` ⚠️ **DUPLICATE**
19. `lint` ⚠️ **DUPLICATE**
20. `mlx` ⚠️ **DUPLICATE**
21. `ollama` ⚠️ **DUPLICATE**
22. `session` ⚠️ **DUPLICATE**
23. `tool_catalog` ⚠️ **DUPLICATE**
24. `workflow_mode` ⚠️ **DUPLICATE**

---

## Duplication Summary

### Duplicated Tools: **12 tools**
- `add_external_tool_hints`
- `automation`
- `check_attribution`
- `estimation`
- `git_tools`
- `infer_session_mode`
- `lint`
- `mlx`
- `ollama`
- `session`
- `tool_catalog`
- `workflow_mode`

### Unique to Coordinator: **2 tools**
- `demonstrate_elicit`
- `interactive_task_create`

### Unique to exarp-go: **12 tools**
- All originally migrated tools (analyze_alignment, generate_config, health, memory, memory_maint, report, security, setup_hooks, task_analysis, task_discovery, task_workflow, testing)

---

## Impact

### Issues:
1. **Tool Name Conflicts** - Same tool names in both servers may cause confusion
2. **Redundant Execution** - Both servers provide identical functionality for 12 tools
3. **Maintenance Overhead** - Changes need to be made in both places
4. **Resource Usage** - Running both servers uses more memory/CPU

### Benefits (Current State):
1. **Backward Compatibility** - Coordinator still works for existing workflows
2. **Migration Path** - exarp-go provides Go-native implementation
3. **Fallback Option** - Can use coordinator if exarp-go has issues

---

## Recommendations

### Option 1: Disable Coordinator (Recommended)
**If exarp-go is fully functional:**
- Disable `coordinator` server in `.cursor/mcp.json`
- Keep only `exarp-go` active
- **Pros:** Cleaner setup, no duplication, single source of truth
- **Cons:** Lose `demonstrate_elicit` and `interactive_task_create` (unless migrated)

### Option 2: Keep Both (Current State)
**If you need coordinator's unique tools:**
- Keep both servers active
- Accept duplication for migration period
- **Pros:** All tools available, gradual migration
- **Cons:** Duplication, potential conflicts

### Option 3: Migrate Unique Tools
**Complete the migration:**
- Migrate `demonstrate_elicit` and `interactive_task_create` to exarp-go
- Then disable coordinator
- **Pros:** Complete migration, no duplication
- **Cons:** Requires additional migration work

---

## Action Items

1. **Decide on approach** - Choose one of the options above
2. **If Option 1:** Disable coordinator in `.cursor/mcp.json`
3. **If Option 3:** Create tasks to migrate unique tools
4. **Update documentation** - Reflect chosen approach

---

**Current Status:** Both servers active, 12 tools duplicated

