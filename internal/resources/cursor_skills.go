// cursor_skills.go — Resource handler for .cursor/skills/ SKILL.md files.
package resources

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/davidl71/exarp-go/internal/tools"
)

type cursorSkill struct {
	Name string
	Path string
}

// cursorSkills are relative to project root; order determines aggregated output order.
var cursorSkills = []cursorSkill{
	{Name: "use-exarp-tools", Path: ".cursor/skills/use-exarp-tools/SKILL.md"},
	{Name: "task-workflow", Path: ".cursor/skills/task-workflow/SKILL.md"},
	{Name: "session-handoff", Path: ".cursor/skills/session-handoff/SKILL.md"},
	{Name: "report-scorecard", Path: ".cursor/skills/report-scorecard/SKILL.md"},
	{Name: "task-cleanup", Path: ".cursor/skills/task-cleanup/SKILL.md"},
	{Name: "lint-docs", Path: ".cursor/skills/lint-docs/SKILL.md"},
	{Name: "database-maintenance", Path: ".cursor/skills/database-maintenance/SKILL.md"},
	{Name: "text-generate", Path: ".cursor/skills/text-generate/SKILL.md"},
	{Name: "thinking-workflow", Path: ".cursor/skills/thinking-workflow/SKILL.md"},
	{Name: "tractatus-decompose", Path: ".cursor/skills/tractatus-decompose/SKILL.md"},
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

	for _, skill := range cursorSkills {
		full := filepath.Join(projectRoot, skill.Path)

		body, err := os.ReadFile(full)
		if err != nil {
			continue // skip missing skills
		}

		parts = append(parts, fmt.Sprintf("## %s\n\n%s\n", skill.Name, strings.TrimSpace(string(body))))
	}

	if len(parts) <= 1 {
		return staticSkillHints(), "text/markdown", nil
	}

	return []byte(strings.Join(parts, "\n")), "text/markdown", nil
}

// handleCursorSkillByName handles stdio://cursor/skills/{name}.
// Returns one skill body so clients can load only the relevant skill on demand.
func handleCursorSkillByName(ctx context.Context, uri string) ([]byte, string, error) {
	skillName, err := parseURIVariableByIndexWithValidation(uri, 4, "name", "stdio://cursor/skills/{name}")
	if err != nil {
		return nil, "", err
	}

	projectRoot, err := tools.FindProjectRoot()
	if err != nil {
		return nil, "", fmt.Errorf("find project root: %w", err)
	}

	skill, ok := findCursorSkill(skillName)
	if !ok {
		return nil, "", fmt.Errorf("unknown skill: %s", skillName)
	}

	body, err := os.ReadFile(filepath.Join(projectRoot, skill.Path))
	if err != nil {
		return nil, "", fmt.Errorf("read skill %s: %w", skillName, err)
	}

	return body, "text/markdown", nil
}

func findCursorSkill(name string) (cursorSkill, bool) {
	for _, skill := range cursorSkills {
		if strings.EqualFold(skill.Name, name) {
			return skill, true
		}
	}

	return cursorSkill{}, false
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
