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
| **Suggested next tasks / what to work on** | `session` with `action=prime`, `include_tasks=true`, `include_hints=true`. Returns `suggested_next` (backlog in dependency order) plus `lazy_context` pointers for task/skill resources. |
| Task list/update/create/show/delete | Prefer `task_workflow` MCP tool when exarp-go MCP is available; fallback: `exarp-go task` CLI (see task-workflow skill). |
| Project overview, scorecard, or briefing | `report` with `action=overview`, `action=scorecard`, or `action=briefing`. |
| Docs health, CI, repo status, or SQLite maintenance | `health` with appropriate `action` (e.g. docs, git, cicd, database). |
| Task branches, merge task changes, task commit history, diff tasks | `git_tools` with `action=commits|branches|tasks|diff|graph|merge|set_branch`. |
| **Task analysis (deps, duplicates, plan)** | `task_analysis` with `action=parallelization|dependencies|duplicates|conflicts|execution_plan|tags|suggest_deps|stale|completable`. |
| **Run a task (execute)** | `task_execute` — execute a task by ID (single action). |
| **Agent startup briefing** | Resource `stdio://agent/briefing` — compact session prime + orchestration lanes + ledger in one fetch. |
| **Per-task execution pack** | Resource `stdio://agent/task/{task_id}/execution-pack` — workflow contract, safe actions, preconditions, recent runs. |
| **Execution alerts** | Resource `stdio://agent/alerts` — stale locks, long-running runs, review-ready tasks. |
| **Broken references / link check in docs** | `lint` with `path` set to `docs` (or a `.md` file) and `linter=markdownlint` or `auto`. gomarklint link check is enabled in `.gomarklint.json`. See **lint-docs** skill. |
| **Task discovery (TODO/markdown/orphans)** | `task_discovery` with `action=comments|markdown|planning_links|orphans|all`; optional `create_tasks=true`. Deprecated items (strikethrough, "(removed)") are never created as tasks — see `.cursor/rules/task-discovery.mdc`. |
| Session context at conversation start | `session` with `action=prime`, `include_hints=true`, `include_tasks=true`. |
| Test structure or runs | `testing` with `action=validate`, `action=run`, or `action=coverage`. |
| Tool-specific help | `tool_catalog` with `action=help` and `tool_name`, or stdio://tools resources. |
| **Bulk remove one-off/performance tasks** | Use **task-cleanup** skill (batch delete via `task_workflow` with `task_ids`). See `.cursor/skills/task-cleanup/SKILL.md`. |
| **Which skills to use** | Read resource **stdio://agent/skills** (legacy alias: **stdio://cursor/skills**) or **.cursor/skills/README.md** for task-workflow, use-exarp-tools, report-scorecard, task-cleanup, lint-docs, tractatus-decompose. |
| **Available prompts (workflow, persona, category)** | Resource **stdio://prompts**; **stdio://prompts/mode/{mode}**, **stdio://prompts/persona/{persona}**, **stdio://prompts/category/{category}** for filtered lists. |
| **Models / LLM backends (local AI)** | Resource **stdio://models** — returns `data.models` (recommend catalog) and **data.backends** (fm_available, ollama_reachable, localai_available, gateway_available). Use to choose backend before calling `apple_foundation_models`, `ollama`, or `text_generate`. |
| **Task list / suggested tasks** | **stdio://tasks**, **stdio://tasks/status/{status}**, **stdio://suggested-tasks** for dependency-ready tasks. |
| **Docs/code for a GitHub repo** | **GitMCP** for a specific repo’s docs/code (e.g. this repo); **Context7** for library/framework docs. See `.cursor/rules/mcp-configuration.mdc` (Context7 vs GitMCP vs web search). **GitHub MCP** for issues/PRs/repo API. |

## Resources and prompts (quick reference)

- **stdio://agent/skills** — Aggregated agent-facing skill guide for exarp-go.
- **stdio://agent/skills/{name}** — Load one skill lazily (for example `stdio://agent/skills/task-workflow`) when `suggested_next[].lazy_context` points to it.
- **stdio://cursor/skills** and **stdio://cursor/skills/{name}** — legacy aliases kept for compatibility.
- **stdio://tools** — Full tool catalog; **stdio://tools/{category}** for category filter (e.g. "Task Management", "AI & ML").
- **stdio://prompts** — All prompt names and short descriptions; use **/mode/{mode}**, **/persona/{persona}**, **/category/{category}** for filtered lists.
- **stdio://models** — Model catalog and `backends` (fm_available, ollama_tool, localai_tool, gateway_tool). Check before using LLM tools (see .cursor/rules/llm-tools.mdc).
- **stdio://tasks**, **stdio://suggested-tasks** — Task list and dependency-ready suggestions.

## Discovering exarp-go usage (do not run --help)

**Do not run `exarp-go --help`, `exarp-go help`, or `./bin/exarp-go` to discover tools or usage.** The binary is either an MCP server (stdio) or shows only flag usage (-tool, -list, -args). For discovery use:

- **Tools and capabilities:** MCP resource **stdio://tools** or **stdio://tools/{category}**, or `tool_catalog` with `action=help` and `tool_name`.
- **CLI subcommands (task, config, tui):** **make help** in the repo, or read `.cursorrules` / task-workflow skill for task commands.

Using MCP resources avoids unnecessary process spawns and gives full tool/prompt lists.

## General Rules

1. **PROJECT_ROOT** – exarp-go uses the project root from its config (e.g. `PROJECT_ROOT` in `~/.cursor/mcp.json` env). Do not pass project root in tool args unless the tool schema asks for it.
2. **Prefer convenience** – Use high-level flows (e.g. `exarp-go task ...`, `report` actions) before raw tool JSON when the skill or docs say to.
3. **Errors** – If a tool fails, check that exarp-go is running and that PROJECT_ROOT matches the workspace you mean.
4. **Claude Code — load schemas first** – MCP tool schemas are deferred in Claude Code. Use `ToolSearch` to load a tool's schema before invoking it. Common order: `session` → `task_workflow` → `report` → `health`. Resources (`read_resource`, `list_resources`) are always available without pre-loading.

## Examples

- *User: "What’s the project status?"* → Use `report` with `action=overview` or `action=scorecard`.
- *User: "What should I work on next?" or "Suggest next task"* → Use `session` with `action=prime`, `include_tasks=true`, `include_hints=true`. Response includes `suggested_next` and `lazy_context`; load the pointed task/skill resources only for the task you actually start.
- *User: "List my Todo tasks"* → Use task-workflow patterns: `exarp-go task list --status Todo` or `task_workflow` with `action=list`, `status=Todo`, `output_format=json`.
- *User: "Is the docs setup ok?"* → Use `health` with `action=docs`.
- *User: "How big is the Todo2 database?" or "Run vacuum/checkpoint"* → Use `health` with `action=database` and `operation=status|checkpoint|vacuum|analyze`.
- *User: "Show task change history" or "Merge my task branch"* → Use `git_tools` with `action=commits`, `action=graph`, or `action=merge` (task branches/versioning).
