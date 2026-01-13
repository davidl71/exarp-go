# Task Discovery: Git JSON Support

**Date:** 2026-01-12  
**Status:** ✅ Implemented

## Overview

Added `action="git_json"` to the `task_discovery` tool to discover tasks from JSON files committed in the git repository. This allows you to find tasks from historical commits, different branches, and multiple JSON files across the repository.

## New Action: `git_json`

### Purpose

Discover tasks from JSON files (typically `.todo2/state.todo2.json`) that are committed in the git repository. This is useful for:

- Finding tasks from historical commits
- Discovering tasks from different branches
- Extracting tasks from multiple JSON files across the repo
- Recovering tasks that were committed but not yet in the current database

### Usage

```bash
# Discover tasks from git JSON files
@exarp-go task_discovery action=git_json

# Discover with custom JSON pattern
@exarp-go task_discovery action=git_json json_pattern="**/*.json"

# Include in "all" discovery
@exarp-go task_discovery action=all
```

### Parameters

- **`action`**: Set to `"git_json"` or include in `"all"`
- **`json_pattern`**: (Optional) Glob pattern to match JSON files. Default: `"**/.todo2/state.todo2.json"`
- **`create_tasks`**: (Optional) If `true`, creates discovered tasks in the database. Default: `false`

### How It Works

1. **Find JSON Files**: Uses `git ls-files` to find all tracked JSON files matching the pattern
2. **Scan Git History**: For each JSON file, uses `git log` to find all commits that modified it
3. **Extract Tasks**: For each commit, extracts tasks from the JSON file at that commit
4. **Deduplicate**: Tracks unique tasks across commits to avoid duplicates
5. **Return Results**: Returns discovered tasks with metadata (file, commit, status, priority)

### Example Output

```json
{
  "action": "git_json",
  "discoveries": [
    {
      "type": "JSON_TASK",
      "text": "Migrate health tool to native Go",
      "task_id": "T-123",
      "status": "Done",
      "priority": "high",
      "file": ".todo2/state.todo2.json",
      "commit": "4df2051",
      "source": "git_json",
      "completed": true
    }
  ],
  "summary": {
    "total": 1,
    "by_source": {
      "git_json": 1
    },
    "by_type": {
      "JSON_TASK": 1
    }
  },
  "method": "native_go"
}
```

## Implementation Details

### Files Modified

1. **`internal/tools/registry.go`**
   - Added `"git_json"` to action enum
   - Added `json_pattern` parameter

2. **`internal/tools/task_discovery_native.go`** (Apple FM version)
   - Added `scanGitJSON()` function
   - Added `loadJSONStateFromFile()` and `loadJSONStateFromContent()` helpers
   - Integrated into `handleTaskDiscoveryNative()`

3. **`internal/tools/task_discovery_native_nocgo.go`** (Non-Apple FM version)
   - Added same functions for cross-platform support

### Functions Added

#### `scanGitJSON(projectRoot, jsonPattern)`

Scans git repository for JSON files and extracts tasks:
- Uses `git ls-files` to find tracked JSON files
- Uses `git log` to find all commits modifying each file
- Uses `git show` to get file content from each commit
- Parses JSON and extracts tasks
- Deduplicates tasks across commits

#### `loadJSONStateFromFile(jsonPath)`

Loads tasks from a JSON file (reused from migration tool):
- Reads JSON file from filesystem
- Parses Todo2 format
- Returns tasks and comments

#### `loadJSONStateFromContent(data)`

Loads tasks from JSON content bytes:
- Parses JSON from byte array
- Handles both Todo2 formats (top-level comments, nested comments)
- Returns tasks and comments

## Use Cases

### 1. Discover Historical Tasks

Find tasks from previous commits that might have been lost or forgotten:

```bash
@exarp-go task_discovery action=git_json create_tasks=true
```

### 2. Find Tasks from Different Branches

Discover tasks committed in other branches:

```bash
# Git log scans all branches by default
@exarp-go task_discovery action=git_json
```

### 3. Recover Lost Tasks

If tasks were deleted or lost, recover them from git history:

```bash
@exarp-go task_discovery action=git_json create_tasks=true
```

### 4. Audit Task History

See how tasks evolved over time across commits:

```bash
@exarp-go task_discovery action=git_json
# Review discoveries with commit hashes
```

## Integration with Other Actions

The `git_json` action works seamlessly with other discovery actions:

```bash
# Discover from all sources
@exarp-go task_discovery action=all

# This will discover from:
# - Comments (TODO/FIXME in code)
# - Markdown (task lists in .md files)
# - Orphans (invalid task structure)
# - Git JSON (tasks from committed JSON files)
```

## Limitations

1. **Git Required**: Only works in git repositories
2. **Tracked Files Only**: Only finds files tracked by git (not untracked files)
3. **Performance**: Scanning all commits can be slow for large repositories
4. **Memory**: Large JSON files or many commits may use significant memory

## Future Enhancements

Potential improvements:

1. **Branch Filtering**: Add parameter to scan specific branches only
2. **Commit Range**: Add parameters for date/commit range filtering
3. **Incremental Discovery**: Track last scanned commit to only scan new commits
4. **Caching**: Cache parsed results to speed up repeated scans
5. **Parallel Processing**: Process multiple files/commits in parallel

## Related Tools

- **Migration Tool** (`cmd/migrate`): Migrates tasks from single JSON file to SQLite
- **Task Discovery**: Discovers tasks from codebase (comments, markdown, orphans, git_json)
- **Task Workflow**: Manages task lifecycle (create, update, approve)

## See Also

- `docs/MIGRATION_VS_DISCOVERY_COMPARISON.md` - Comparison of migration and discovery tools
- `docs/NATIVE_GO_MIGRATION_PLAN.md` - Native Go migration plan
