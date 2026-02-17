package resources

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/davidl71/exarp-go/internal/tools"
)

// skillPaths are relative to project root; order determines output order.
var cursorSkillPaths = []string{
	".cursor/skills/use-exarp-tools/SKILL.md",
	".cursor/skills/task-workflow/SKILL.md",
	".cursor/skills/session-handoff/SKILL.md",
	".cursor/skills/report-scorecard/SKILL.md",
	".cursor/skills/task-cleanup/SKILL.md",
	".cursor/skills/lint-docs/SKILL.md",
	".cursor/skills/tractatus-decompose/SKILL.md",
}

// handleCursorSkills handles the stdio://cursor/skills resource.
// Returns workflow guidance for all MCP clients: Cursor (skills) and Claude Code (commands/CLAUDE.md).
// Reads .cursor/skills/*/SKILL.md from project root and concatenates them.
func handleCursorSkills(ctx context.Context, uri string) ([]byte, string, error) {
	projectRoot, err := tools.FindProjectRoot()
	if err != nil {
		return staticSkillHints(), "text/markdown", nil
	}

	var parts []string
	parts = append(parts, "# exarp-go workflow guide\n\nApply when using exarp-go MCP tools. Works with Cursor (skills) and Claude Code (CLAUDE.md + commands).\n")

	for _, rel := range cursorSkillPaths {
		full := filepath.Join(projectRoot, rel)

		body, err := os.ReadFile(full)
		if err != nil {
			continue // skip missing skills
		}

		name := filepath.Base(filepath.Dir(rel))
		parts = append(parts, fmt.Sprintf("## %s\n\n%s\n", name, strings.TrimSpace(string(body))))
	}

	if len(parts) <= 1 {
		return staticSkillHints(), "text/markdown", nil
	}

	return []byte(strings.Join(parts, "\n")), "text/markdown", nil
}

// staticSkillHints returns a fallback hint table when skill files are not found.
func staticSkillHints() []byte {
	return []byte(strings.TrimSpace(`
# exarp-go workflow guide

When using exarp-go MCP, apply the following patterns:

| User intent | Tool / pattern |
|-------------|----------------|
| Tasks, Todo2, list/update/create/show/delete, next task | task_workflow tool; prefer exarp-go task CLI |
| Suggested next task, what to work on | session(action=prime, include_tasks=true) |
| End session, handoff, list handoffs | session(action=handoff, sub_action=end|list|resume) |
| Project overview, scorecard, briefing | report(action=overview|scorecard|briefing) |
| Health, docs, CI | health(action=docs|git|cicd) |
| Broken references, validate doc links, lint markdown | lint tool with markdownlint |
| Bulk remove one-off/performance tasks | task_workflow(action=delete, task_ids=...) |
| Logical decomposition, complex concepts | tractatus_thinking MCP (operation=start, add, export) |

Cursor: skills in .cursor/skills/. Claude Code: see CLAUDE.md and .claude/commands/.
`))
}
