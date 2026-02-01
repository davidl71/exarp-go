# SQLite Migration vs Task Discovery Tool Comparison

**Date:** 2026-01-12  
**Status:** Analysis Complete

## Overview

Both tools work with Todo2 tasks and interact with the SQLite database, but serve different purposes. This document compares them and evaluates whether they can be unified.

---

## Tool Comparison

### SQLite Migration Tool (`cmd/migrate/main.go`)

**Purpose:** One-time migration of existing tasks from JSON to SQLite database

**Functionality:**
- ✅ Loads tasks from `.todo2/state.todo2.json`
- ✅ Loads comments from JSON (supports two formats)
- ✅ Creates/updates tasks in SQLite database
- ✅ Migrates comments to database
- ✅ Dry-run mode (preview without changes)
- ✅ Backup creation before migration
- ✅ Duplicate detection (skips existing comments)
- ✅ Conflict resolution (updates existing tasks)

**Input:**
- JSON file: `.todo2/state.todo2.json`

**Output:**
- SQLite database: `.todo2/todo2.db`
- Backup file: `.todo2/state.todo2.json.backup`

**Use Cases:**
- Initial migration from JSON to SQLite
- Re-migration after JSON file updates
- Recovery from JSON backup

**Key Features:**
- One-time operation (not meant for ongoing use)
- Data preservation (backup before migration)
- Conflict handling (update vs create)
- Comment deduplication

---

### Task Discovery Tool (`task_discovery`)

**Purpose:** Ongoing discovery of new tasks from codebase sources

**Functionality:**
- ✅ Scans code files for TODO/FIXME comments
- ✅ Scans markdown files for task lists (`- [ ]`)
- ✅ Finds orphan tasks (invalid structure, missing deps, cycles)
- ✅ Optional task creation (`create_tasks` parameter)
- ✅ AI enhancement (Apple FM for semantic extraction)
- ✅ Multiple actions: `comments`, `markdown`, `orphans`, `all`

**Input:**
- Codebase files (`.go`, `.py`, `.js`, `.ts`, `.md`)
- Existing Todo2 tasks (for orphan detection)

**Output:**
- JSON report of discoveries
- Optional: Creates tasks in Todo2 database

**Use Cases:**
- Regular task discovery from codebase
- Finding incomplete tasks
- Detecting task structure issues
- Populating task database from code

**Key Features:**
- Ongoing operation (run regularly)
- Multiple source types
- Optional task creation
- AI-enhanced extraction

---

## Similarities

1. **Both interact with Todo2 database**
   - Use `database.CreateTask()` / `database.UpdateTask()`
   - Handle comments via `database.AddComments()`

2. **Both can create tasks**
   - Migration: Creates tasks from JSON
   - Discovery: Creates tasks from codebase (optional)

3. **Both handle task data structures**
   - Work with `models.Todo2Task`
   - Handle task IDs, content, status, etc.

4. **Both support project root detection**
   - Find `.todo2` directory
   - Use `FindProjectRoot()` or similar

---

## Differences

| Aspect | Migration Tool | Task Discovery |
|--------|---------------|----------------|
| **Primary Purpose** | One-time data migration | Ongoing task discovery |
| **Input Source** | JSON file | Codebase files |
| **Frequency** | Rare (one-time or recovery) | Regular (ongoing) |
| **Backup** | Creates backup before migration | No backup needed |
| **Conflict Resolution** | Updates existing tasks | Creates new tasks (or skips) |
| **Dry Run** | Yes (preview migration) | No (but can skip `create_tasks`) |
| **Comment Handling** | Migrates existing comments | Doesn't create comments |
| **Validation** | Validates JSON structure | Validates task structure (orphans) |
| **AI Enhancement** | No | Yes (optional Apple FM) |

---

## Can They Be United?

### ✅ **Arguments FOR Unification:**

1. **Shared Infrastructure**
   - Both use same database functions
   - Both handle Todo2 tasks
   - Could share common code

2. **Similar Operations**
   - Both create/update tasks
   - Both handle task data
   - Both work with project root

3. **Logical Extension**
   - Migration could be a "source type" in discovery
   - Discovery could add "json" action
   - Unified interface for all task sources

4. **Code Reduction**
   - Eliminate duplicate code
   - Single tool for all task operations
   - Easier maintenance

### ❌ **Arguments AGAINST Unification:**

1. **Different Use Cases**
   - Migration: One-time, data preservation focus
   - Discovery: Ongoing, codebase scanning focus

2. **Different Input Formats**
   - Migration: Structured JSON file
   - Discovery: Unstructured codebase files

3. **Different Validation Needs**
   - Migration: JSON schema validation, backup
   - Discovery: Code parsing, orphan detection

4. **Different User Expectations**
   - Migration: CLI tool (explicit, one-time)
   - Discovery: MCP tool (regular, automated)

5. **Complexity Risk**
   - Combining could make tool harder to use
   - Different parameters for different modes
   - Risk of breaking one use case while fixing another

---

## Recommendation: **Hybrid Approach** ✅

### Option 1: Keep Separate, Share Code (Recommended)

**Keep tools separate but extract shared functionality:**

```go
// internal/tools/task_import.go (new shared module)
package tools

// ImportTasksFromJSON imports tasks from JSON file
func ImportTasksFromJSON(jsonPath string, dryRun bool) ([]models.Todo2Task, []database.Comment, error)

// ImportTasksFromCodebase discovers tasks from codebase
func ImportTasksFromCodebase(projectRoot string, action string) ([]map[string]interface{}, error)

// CreateTasksFromDiscoveries creates tasks from discovery results
func CreateTasksFromDiscoveries(discoveries []map[string]interface{}) error
```

**Benefits:**
- ✅ Clear separation of concerns
- ✅ Shared code reduces duplication
- ✅ Each tool maintains its specific purpose
- ✅ Easier to test and maintain

**Migration Tool:**
- Uses `ImportTasksFromJSON()`
- Handles backup, validation, conflict resolution
- CLI-focused

**Discovery Tool:**
- Uses `ImportTasksFromCodebase()`
- Handles AI enhancement, orphan detection
- MCP-focused

---

### Option 2: Add Migration Action to Discovery

**Add `action="json"` to task_discovery:**

```go
// In task_discovery handler
if action == "json" || action == "all" {
    jsonPath := filepath.Join(projectRoot, ".todo2", "state.todo2.json")
    jsonTasks, jsonComments, err := ImportTasksFromJSON(jsonPath, false)
    // Convert to discovery format
    discoveries = append(discoveries, convertJSONTasksToDiscoveries(jsonTasks)...)
}
```

**Benefits:**
- ✅ Single tool for all task sources
- ✅ Unified interface
- ✅ Can discover from JSON + codebase in one call

**Drawbacks:**
- ❌ Migration-specific features (backup, dry-run) need to be added
- ❌ Tool becomes more complex
- ❌ CLI migration tool still needed for explicit migration

---

### Option 3: Unified Task Import Tool

**Create new unified tool that handles all sources:**

```go
// New tool: task_import
// Actions: json, comments, markdown, orphans, all
// Supports: dry-run, backup, create_tasks
```

**Benefits:**
- ✅ Single tool for all import/discovery operations
- ✅ Consistent interface
- ✅ Can combine sources

**Drawbacks:**
- ❌ Requires refactoring both tools
- ❌ More complex tool
- ❌ May break existing workflows

---

## Final Recommendation: **Option 1** ✅

**Keep tools separate, extract shared code:**

1. **Create shared module** (`internal/tools/task_import.go`)
   - `ImportTasksFromJSON()` - JSON loading logic
   - `ImportTasksFromCodebase()` - Codebase scanning logic
   - `CreateTasksFromDiscoveries()` - Task creation logic

2. **Refactor migration tool**
   - Use shared `ImportTasksFromJSON()`
   - Keep migration-specific features (backup, dry-run, conflict resolution)
   - Maintain CLI interface

3. **Refactor discovery tool**
   - Use shared `ImportTasksFromCodebase()`
   - Keep discovery-specific features (AI enhancement, orphan detection)
   - Maintain MCP interface

4. **Optional: Add JSON source to discovery**
   - Add `action="json"` that uses shared `ImportTasksFromJSON()`
   - Allows discovery from JSON without full migration
   - Useful for preview/analysis

**Benefits:**
- ✅ Clear separation of concerns
- ✅ Code reuse without complexity
- ✅ Each tool maintains its purpose
- ✅ Easy to extend (add new sources to discovery)
- ✅ Backward compatible

---

## Implementation Plan

### Phase 1: Extract Shared Code
1. Create `internal/tools/task_import.go`
2. Move JSON loading logic from migration tool
3. Move codebase scanning logic from discovery tool
4. Create shared task creation function

### Phase 2: Refactor Tools
1. Update migration tool to use shared code
2. Update discovery tool to use shared code
3. Add tests for shared functions

### Phase 3: Optional Enhancement
1. Add `action="json"` to discovery tool
2. Document unified interface
3. Update documentation

---

## Conclusion

**Tools serve different purposes but share common functionality.**

**Best approach:** Keep tools separate for clarity, but extract shared code to reduce duplication. This maintains clear separation of concerns while enabling code reuse.

**Migration tool:** One-time, data-focused, CLI-based  
**Discovery tool:** Ongoing, codebase-focused, MCP-based  
**Shared code:** Common import/creation logic

This approach provides the benefits of unification (code reuse) without the drawbacks (complexity, breaking changes).
