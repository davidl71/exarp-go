---
name: use-exarp-tools
description: When and how to use exarp-go MCP tools. Use when the workspace has exarp-go configured, when the user asks about tasks, suggested next tasks, reports, health, testing, or project automation, or when you need PROJECT_ROOT-aware tool calls.
---

# Using exarp-go MCP Tools

Apply this skill when the workspace uses the exarp-go MCP server and you need to run project automation, tasks, reports, or health checks.

## Configuration

- exarp-go must be in MCP config (`~/.cursor/mcp.json`) with `PROJECT_ROOT` (or equivalent) set for the current workspace.
- Tools and prompts are invoked via the exarp-go server; do not assume paths or project root. Use the project root the server is configured with.

## When to Use Key Tools

| Need | Tool or pattern |
|------|------------------|
| **Suggested next tasks / what to work on** | `session` with `action=prime`, `include_tasks=true`, `include_hints=true`. Returns `suggested_next` (backlog in dependency order). |
| Task list/update/create/show/delete | Prefer `task_workflow` MCP tool when exarp-go MCP is available; fallback: `exarp-go task` CLI (see task-workflow skill). |
| Project overview, scorecard, or briefing | `report` with `action=overview`, `action=scorecard`, or `action=briefing`. |
| Docs health, CI, or repo status | `health` with appropriate `action` (e.g. docs, git, cicd). |
| Task branches, merge task changes, task commit history, diff tasks | `git_tools` with `action=commits|branches|tasks|diff|graph|merge|set_branch`. |
| **Broken references / link check in docs** | `lint` with `path` set to `docs` (or a `.md` file) and `linter=markdownlint` or `auto`. gomarklint link check is enabled in `.gomarklint.json`. See **lint-docs** skill. |
| Session context at conversation start | `session` with `action=prime`, `include_hints=true`, `include_tasks=true`. |
| Test structure or runs | `testing` with `action=validate`, `action=run`, or `action=coverage`. |
| Tool-specific help | `tool_catalog` with `action=help` and `tool_name`, or stdio://tools resources. |
| **Bulk remove one-off/performance tasks** | Use **task-cleanup** skill (batch delete via `task_workflow` with `task_ids`). See `.cursor/skills/task-cleanup/SKILL.md`. |
| **Cursor: which skills to use** | Read resource **stdio://cursor/skills** or **.cursor/skills/README.md** for task-workflow, use-exarp-tools, report-scorecard, task-cleanup, lint-docs. |

## General Rules

1. **PROJECT_ROOT** – exarp-go uses the project root from its config (e.g. `PROJECT_ROOT` in `~/.cursor/mcp.json` env). Do not pass project root in tool args unless the tool schema asks for it.
2. **Prefer convenience** – Use high-level flows (e.g. `exarp-go task ...`, `report` actions) before raw tool JSON when the skill or docs say to.
3. **Errors** – If a tool fails, check that exarp-go is running and that PROJECT_ROOT matches the workspace you mean.

## Examples

- *User: "What’s the project status?"* → Use `report` with `action=overview` or `action=scorecard`.
- *User: "What should I work on next?" or "Suggest next task"* → Use `session` with `action=prime`, `include_tasks=true`, `include_hints=true`. Response includes `suggested_next` (tasks in dependency order).
- *User: "List my Todo tasks"* → Use task-workflow patterns: `exarp-go task list --status Todo` or `task_workflow` with `action=sync`, `sub_action=list`, `status=Todo`.
- *User: "Is the docs setup ok?"* → Use `health` with `action=docs`.
- *User: "Show task change history" or "Merge my task branch"* → Use `git_tools` with `action=commits`, `action=graph`, or `action=merge` (task branches/versioning).
