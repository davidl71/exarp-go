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
// Returns the full content of workspace Cursor skills so Cursor can learn them from one fetch.
// Reads .cursor/skills/*/SKILL.md from project root and concatenates them.
func handleCursorSkills(ctx context.Context, uri string) ([]byte, string, error) {
	projectRoot, err := tools.FindProjectRoot()
	if err != nil {
		// Fallback: return static hint table so Cursor still gets guidance
		return staticSkillHints(), "text/markdown", nil
	}

	var parts []string
	parts = append(parts, "# Cursor skills (exarp-go)\n\nFetched from workspace. Apply when using exarp-go MCP.\n")

	for _, rel := range cursorSkillPaths {
		full := filepath.Join(projectRoot, rel)
		body, err := os.ReadFile(full)
		if err != nil {
			continue // skip missing skills
		}
		name := filepath.Base(filepath.Dir(rel))
		parts = append(parts, fmt.Sprintf("## Skill: %s\n\n%s\n", name, strings.TrimSpace(string(body))))
	}

	if len(parts) <= 1 {
		return staticSkillHints(), "text/markdown", nil
	}

	return []byte(strings.Join(parts, "\n")), "text/markdown", nil
}

// staticSkillHints returns a fallback hint table when skill files are not found.
func staticSkillHints() []byte {
	return []byte(strings.TrimSpace(`
# Cursor skill hints (exarp-go)

When using exarp-go MCP, consider these skills (read from .cursor/skills/ if present):

| User intent | Skills |
|-------------|--------|
| Tasks, Todo2, list/update/create/show/delete, next task | task-workflow, use-exarp-tools |
| Suggested next task, what to work on | use-exarp-tools (session prime) |
| End session, handoff, list handoffs | session-handoff |
| Project overview, scorecard, briefing | report-scorecard, use-exarp-tools |
| Health, docs, CI | use-exarp-tools (health tool) |
| Broken references, validate doc links, lint markdown | lint-docs, use-exarp-tools (lint tool) |
| Bulk remove one-off/performance tasks | task-cleanup |
| Logical decomposition, complex concepts | tractatus-decompose |

Paths: .cursor/skills/use-exarp-tools/SKILL.md, .cursor/skills/task-workflow/SKILL.md, .cursor/skills/session-handoff/SKILL.md, .cursor/skills/report-scorecard/SKILL.md, .cursor/skills/task-cleanup/SKILL.md, .cursor/skills/lint-docs/SKILL.md, .cursor/skills/tractatus-decompose/SKILL.md
`))
}
