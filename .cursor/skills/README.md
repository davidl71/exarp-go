# exarp-go Cursor skills

Skills in this folder extend the AI with exarp-go workflows. **Read the SKILL.md in each subfolder** when the user's request matches the description.

**For humans:** See [docs/CURSOR_SKILLS_GUIDE.md](../../docs/CURSOR_SKILLS_GUIDE.md) for a guide on how to use these skills (example prompts, locking, git_tools, conflict detection).

| Skill | When to use |
|-------|-------------|
| **report-scorecard** | Project overview, scorecard, briefing, or status; after big changes; before reviews. |
| **task-workflow** | List, update, create, show, or delete Todo2 tasks; task status; avoid editing `.todo2` files directly. |
| **session-handoff** | End session (create handoff note), list all handoffs, resume from handoff, export handoff data. |
| **task-cleanup** | Bulk remove one-off or performance tasks; when those tasks "reappeared." Use batch delete (`task_ids`) for speed. |
| **lint-docs** | Check broken references, validate doc links, lint markdown; gomarklint link check is built-in via lint tool. |
| **tractatus-decompose** | Use Tractatus Thinking MCP for logical decomposition of complex concepts (operation=start, add, export). |
| **use-exarp-tools** | When to use which exarp-go MCP tool; tasks, reports, health, testing, project automation. |

Path format: `.cursor/skills/<name>/SKILL.md`.
