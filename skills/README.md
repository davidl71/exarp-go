# exarp-go Cursor Skills

Bundled [Cursor Agent Skills](https://www.cursor.com/changelog) that teach the AI when and how to use exarp-go MCP tools. Copy these into your Cursor skills path to use them.

## Installation

**Option A: Personal skills (all projects)**  
Copy each skill directory into your user skills folder:

```bash
cp -r skills/use-exarp-tools ~/.cursor/skills/
cp -r skills/task-workflow ~/.cursor/skills/
cp -r skills/report-scorecard ~/.cursor/skills/
```

**Option B: Project skills (single repo)**  
Copy into the project’s Cursor skills folder:

```bash
mkdir -p .cursor/skills
cp -r skills/use-exarp-tools skills/task-workflow skills/report-scorecard .cursor/skills/
```

**Paths:** Cursor discovers skills from `~/.cursor/skills/` (personal) and `.cursor/skills/` (project). Do not put these in `~/.cursor/skills-cursor/`; that directory is reserved for Cursor’s built-in skills.

## Bundled skills

| Skill | Purpose |
|-------|--------|
| **use-exarp-tools** | When and how to use exarp-go MCP tools; PROJECT_ROOT; key tools (report, health, session, testing). |
| **task-workflow** | Todo2 via exarp-go: prefer `exarp-go task` commands, fallback to `task_workflow` tool; never edit `.todo2/state.todo2.json` directly. |
| **report-scorecard** | Using the `report` tool (overview, scorecard, briefing) and when to run it. |

## Requirements

- exarp-go MCP server configured in `.cursor/mcp.json` with `PROJECT_ROOT` (or equivalent) set for the workspace you use these skills in.

## Future enhancement

A later release may add an exarp-go **tool** (e.g. `create_skill` or an option in `generate_config`) that writes SKILL.md files into your Cursor skills path, so you can install or update skills via the MCP server instead of copying by hand.
