# exarp-go Cursor skills

Skills in this folder extend the AI with exarp-go workflows. **Read the SKILL.md in each subfolder** when the user's request matches the description.

**For humans:** See [docs/CURSOR_SKILLS_GUIDE.md](../../docs/CURSOR_SKILLS_GUIDE.md) for a guide on how to use these skills (example prompts, locking, git_tools, conflict detection).

**Resources:** When using exarp-go MCP, fetch **stdio://cursor/skills** for skill text, **stdio://tools** or **stdio://tools/{category}** for the tool catalog, **stdio://prompts** for prompt list, and **stdio://models** for model/backend discovery. See use-exarp-tools skill for full list.

| Skill | When to use |
|-------|-------------|
| **report-scorecard** | Project overview, scorecard, briefing, or status; after big changes; before reviews. |
| **task-workflow** | List, update, create, show, or delete Todo2 tasks; task status; avoid editing `.todo2` files directly. |
| **session-handoff** | End session (create handoff note), list all handoffs, resume from handoff, export handoff data. |
| **task-cleanup** | Bulk remove one-off or performance tasks; when those tasks "reappeared." Use batch delete (`task_ids`) for speed. |
| **lint-docs** | Check broken references, validate doc links, lint markdown; gomarklint link check is built-in via lint tool. |
| **text-generate** | Quick local LLM text generation; use for fast on-device generation, classification, summarization, or when other AI backends unavailable. |
| **tractatus-decompose** | Use Tractatus Thinking MCP for logical decomposition of complex concepts (operation=start, add, export). |
| **use-exarp-tools** | When to use which exarp-go MCP tool; tasks, reports, health, testing, project automation. |

**Cursor rules (task discovery):** `.cursor/rules/task-discovery.mdc` — task_discovery tool behavior and deprecated-item filter (strikethrough / "(removed)" never create Todo2 tasks).

Path format: `.cursor/skills/<name>/SKILL.md`.
