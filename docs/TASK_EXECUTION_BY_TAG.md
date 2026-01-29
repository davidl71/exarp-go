# Task Execution by Tag

This doc describes how to plan and execute Todo2 tasks scoped by tag(s), using the CLI and the exarp-go MCP server (task_analysis and task_workflow tools).

## Quick list by tag in execution order

Use the CLI to list backlog tasks for a given tag in dependency order:

```bash
exarp-go task list --tag <tag> --order execution
```

Optional: add `--limit <n>` to cap the number of tasks.

Examples:

```bash
exarp-go task list --tag migration --order execution
exarp-go task list --tag bug --order execution --limit 10
```

## Written plan for one tag (MCP)

Call the `task_analysis` tool with `action=execution_plan` and optional `filter_tag` to get a backlog plan restricted to tasks with that tag. The plan includes a **Tags** column per task.

Example (MCP client):

- `action`: `execution_plan`
- `filter_tag`: `migration`
- `output_path`: `docs/execution-migration.md` (optional; if present and path ends in `.md`, the plan is written as Markdown)

The JSON result includes `ordered_task_ids`, `waves`, and `details` (each detail has `id`, `content`, `priority`, `status`, `level`, `tags`).

## Written plan for multiple tags (OR)

To restrict the backlog to tasks that have **any** of several tags, use `filter_tags` (comma-separated):

- `action`: `execution_plan`
- `filter_tags`: `migration,bug`
- `output_path`: optional

Only backlog tasks (Todo or In Progress) that have at least one of the given tags are included; dependency order and levels use the full task set so ordering stays correct.

## Full backlog plan (no tag filter)

To get a plan for the full backlog with no tag filter:

- `action`: `execution_plan`
- Do not set `filter_tag` or `filter_tags`.

The plan now includes a **Tags** column so you can see each task’s tags in the Markdown table and in the JSON `details`.

## Canonical tag set

Prefer tags from the project canonical set so filtering and discovery stay consistent. See [.cursor/rules/code-and-planning-tag-hints.mdc](../.cursor/rules/code-and-planning-tag-hints.mdc) for the full list. Common examples:

- `#migration` `#refactor` `#cli` `#mcp` `#testing` `#docs` `#bug` `#feature` `#config` `#performance` `#database`

## Summary

| Goal                         | CLI                                      | MCP (task_analysis)                          |
|-----------------------------|------------------------------------------|-----------------------------------------------|
| List by tag in exec order   | `exarp-go task list --tag X --order execution` | N/A (use task_workflow list with filter_tag) |
| Plan for one tag            | N/A                                      | `action=execution_plan`, `filter_tag=X`       |
| Plan for multiple tags (OR) | N/A                                      | `action=execution_plan`, `filter_tags=X,Y`   |
| Full backlog plan with tags | N/A                                      | `action=execution_plan` (no filter)           |
