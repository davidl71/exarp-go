# task_discovery Tool

Discover tasks from code comments, markdown files, and orphaned Todo2 tasks.

## Usage

```json
{
  "action": "all",
  "file_patterns": "[\"**/*.go\", \"**/*.py\"]",
  "ignore_paths": ".cache,vendor,third_party",
  "include_fixme": true,
  "create_tasks": false,
  "use_llm": true
}
```

## Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `action` | string | "all" | Action: comments, markdown, orphans, git_json, planning_links, all |
| `file_patterns` | string | ["**/*.go", "**/*.py", "**/*.js", "**/*.ts"] | File patterns to scan (JSON array) |
| `ignore_paths` | string | .cache, third_party, mcp-servers, native | Comma-separated paths to exclude from scanning |
| `include_fixme` | boolean | true | Include FIXME comments |
| `doc_path` | string | "docs" | Path to scan for markdown files |
| `json_pattern` | string | "**/.todo2/state.todo2.json" | Pattern for orphaned task detection |
| `create_tasks` | boolean | false | Auto-create discovered tasks in Todo2 |
| `use_llm` | boolean | true | Use Apple FM for semantic enhancement (Darwin/arm64) |

## Default Ignore Patterns

The following paths are automatically excluded from scanning to avoid noise:

**Version control & dependencies:**
- `.git`, `node_modules`, `vendor`

**Python virtual environments:**
- `__pycache__`, `.venv`

**IDE & editors:**
- `.idea`, `.vscode`

**Build outputs:**
- `dist`, `build`, `target`, `archive`, `bin`

**Common noise (external projects, caches):**
- `.cache`, `third_party`, `mcp-servers`, `native`

## Customizing Ignore Paths

Use the `ignore_paths` parameter to add custom exclusions:

```json
{
  "action": "comments",
  "ignore_paths": "vendor,third_party,my-custom-dir"
}
```

The `ignore_paths` parameter adds to (not replaces) the default patterns above.

## Actions

### comments
Scan source code files for TODO and FIXME comments.

### markdown
Scan markdown files for task lists (e.g., `- [ ] Task name`).

### orphans
Find Todo2 tasks that exist in JSON but have no corresponding code/docs references.

### git_json
Scan git history for commit messages containing task IDs.

### planning_links
Scan planning documents (.plan.md) for task references.

### all
Run all discovery actions and combine results.

## Examples

```json
{
  "action": "comments",
  "file_patterns": "[\"**/*.go\", \"**/*.ts\"]",
  "ignore_paths": "vendor,third_party"
}
```

```json
{
  "action": "markdown",
  "doc_path": "docs",
  "create_tasks": true
}
```
