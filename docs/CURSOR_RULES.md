# Cursor Rules (exarp-go)

This project uses [Cursor rules](https://docs.cursor.com/context/rules-for-ai) in `.cursor/rules/` to guide the AI when editing code, plans, and config. Rules apply when matching files are open or when marked always-apply.

## Rule index

| Rule | Description | When it applies |
|------|-------------|-----------------|
| `agent-locking.mdc` | Task locking and agent ID for parallel execution | Always |
| `agentic-ci.mdc` | Agentic CI workflow and validation | Always |
| `code-and-planning-tag-hints.mdc` | **Tag hints in plans and code** | Plans (`.plan.md`, `docs/*_PLAN*.md`) and Go (`**/*.go`) |
| `plan-todos-required.mdc` | **Todos with task IDs required in plans** | Plans (`**/*.plan.md`, `.cursor/plans/**`) |
| `go-development.mdc` | Go style, Makefile, Todo2 DB, testing | Go files, always for key sections |
| `llm-tools.mdc` | LLM backend discovery and tool choice | Always |
| `mcp-configuration.mdc` | MCP server config and Context7 vs web search | Always |
| `session-prime.mdc` | Session priming at conversation start | Always |
| `todo2.mdc` | Todo2 workflow, research, task lifecycle | Always |
| `todo2-overview.mdc` | Auto-generated task overview | Always |

## Code and planning tag hints (use this)

**Rule file:** `.cursor/rules/code-and-planning-tag-hints.mdc`

When you (or the AI) create or edit **planning markdown** or **Go code** that maps to Todo2 work, add **tag hints** so tasks stay consistent and discoverable (e.g. for `task_analysis`, `task_workflow`, and filters like `--tag`).

### Planning .md files

- **Where:** `.cursor/plans/*.plan.md`, `docs/*_PLAN*.md`, `docs/*_STRATEGY*.md`, and similar planning docs.
- **What to do:** At the top of the doc (after the title or in frontmatter), add suggested Todo2 tags:
  - Inline: `**Tag hints:** \`#migration\` \`#cli\` \`#refactor\``
  - Or frontmatter: `tag_hints: [migration, cli, refactor]`
- **Optional:** For sections that map to distinct tasks, add a short tag hint under the section heading (e.g. `**Tag hints:** \`#cli\` \`#refactor\``).

### Go code

- **Where:** New or meaningfully changed packages/files under `internal/` (or other Go) that implement a tracked feature.
- **What to do:** Add **one** file-level or package-level tag hint, e.g.:
  - In package doc: `// Package foo implements X. Tag hints for Todo2: #migration #cli`
  - Or: `// exarp-tags: #migration #cli`
- Use 1–3 tags from the project’s canonical set; skip for trivial edits or generated code.

### Canonical tags (prefer these)

Use tags from the project set so filtering and analysis stay consistent:

`#migration` `#refactor` `#cli` `#mcp` `#testing` `#docs` `#performance` `#database` `#bug` `#feature` `#config` `#security` `#build` `#linting` `#concurrency` `#git` `#planning` `#workflow` `#research` `#analysis`

### Quick reference

| Artifact | Where to put tag hints | Example |
|----------|------------------------|---------|
| Planning .md | Top of doc or frontmatter; per section optional | `#migration` `#cli` |
| New/refactored Go package or file | One file-level or package comment | `#refactor` `#mcp` |

Using this rule keeps plans and code aligned with Todo2 and improves task discovery and `task_analysis` results.

## Verifying rules

1. In Cursor: **Settings → Rules** (or project Rules).
2. Confirm the `.cursor/rules/*.mdc` files you expect are listed.
3. Open a plan file or Go file and ask the AI to add tag hints; it should follow the rule when the rule is in scope.

## See also

- **Cursor skills guide:** [docs/CURSOR_SKILLS_GUIDE.md](CURSOR_SKILLS_GUIDE.md) — How to use skills (task-workflow, locking, git_tools, conflict detection)
- Task management: `.cursorrules` (convenience commands vs tool calls)
- Configuration: `docs/CONFIGURABLE_PARAMETERS_RECOMMENDATIONS.md` (default task tags, etc.)
- Task analysis: `task_analysis` tool and `getCanonicalTagsList()` in codebase
