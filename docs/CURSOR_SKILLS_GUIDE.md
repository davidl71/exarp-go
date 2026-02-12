# Cursor Skills Guide (exarp-go)

This guide explains how to use exarp-go's Cursor skills and related rules. Skills extend Cursor's AI with exarp-go workflows so you get the right tool at the right time.

## What Are Skills?

Skills live in `.cursor/skills/<name>/SKILL.md`. When your request matches a skill's description, Cursor reads it and follows its guidance (which tool to use, MCP vs CLI, etc.). You don't install skills manually—they're applied automatically when the AI detects a match.

## Skill Index (Quick Reference)

| Skill | When to ask for it |
|-------|--------------------|
| **task-workflow** | List/update/create/show/delete Todo2 tasks; move tasks between statuses |
| **report-scorecard** | Project overview, scorecard, briefing, or status; after big changes |
| **session-handoff** | End a session, list handoffs, resume from a handoff |
| **task-cleanup** | Bulk remove one-off or performance tasks; when those tasks "reappeared" |
| **lint-docs** | Check broken references, validate doc links, lint markdown |
| **tractatus-decompose** | Logical decomposition of complex concepts before implementation |
| **use-exarp-tools** | Choosing the right exarp-go tool (tasks, reports, health, git, etc.) |

## Example Prompts for Each Skill

Ask naturally; the AI will pick the right skill. Examples:

| What you want | Example prompt |
|---------------|----------------|
| Task management | "List my Todo tasks" / "Move T-123 to Done" / "Create a new task for X" |
| Project status | "What's the project scorecard?" / "Give me a briefing" |
| Session handoff | "End my session" / "List handoffs" / "Resume from latest handoff" |
| Bulk cleanup | "Remove the one-off performance tasks" |
| Doc quality | "Check for broken links in docs" / "Lint the markdown" |
| Tool selection | "Show task change history" / "Merge my task branch" / "What should I work on next?" |

## Multi-Agent: Locking and Task Assignment

When multiple agents (or Cursor instances) may work on the same project:

- **Use exarp-go** for moving tasks from Todo → In Progress. exarp-go uses database locking (`ClaimTaskForAgent`) so only one agent can claim a task at a time.
- **Avoid Todo2 MCP** `update_todos` for Todo → In Progress; it bypasses locking and can cause conflicting assignments.

**Rule:** `.cursor/rules/agent-locking.mdc`  
**Skill:** `.cursor/skills/task-workflow/SKILL.md`

**Example prompt:** "Assign task T-123 to me" or "Start working on T-123" — the AI should use exarp-go `task_workflow` or `exarp-go task update`, not Todo2 MCP.

## Git Tools (Task Branches, Merge, History)

Use the **use-exarp-tools** skill when you need git-related workflows tied to tasks:

| Need | Example prompt |
|------|----------------|
| Task change history | "Show commits for task T-123" |
| Task branches | "List branches for my tasks" |
| Merge task branch | "Merge my task branch into main" |
| Diff between tasks | "Show diff between T-1 and T-2" |
| Commit history | "Show recent commits for this task" |

**Tool:** `git_tools` with `action=commits|branches|tasks|diff|graph|merge|set_branch`

## Task Analysis: Conflict Detection

When multiple agents work in parallel, you may have **task-overlap conflicts**: task A blocks task B, but both are In Progress.

Use the `task_analysis` tool with `action=conflicts`:

**Example prompt:** "Check for task conflicts" / "Are there overlapping In Progress tasks?"

The tool returns:
- `conflict`: boolean
- `conflicts`: list of overlapping pairs
- `overlapping`: task IDs involved
- `reasons`: human-readable explanations

**When to use:** Before assigning more tasks in multi-agent mode, or when you suspect two agents are working on dependent tasks.

## Where to Learn More

- **Skills index:** `.cursor/skills/README.md` — when each skill applies (AI-focused)
- **Locking rule:** `.cursor/rules/agent-locking.mdc` — agent ID, claiming, releasing, AI guidance
- **Cursor rules overview:** `docs/CURSOR_RULES.md` — rules index and tag hints
- **Task cleanup:** `docs/TASKS_AS_CI_AND_AUTOMATION.md` — batch delete usage
