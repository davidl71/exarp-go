package tools

import "fmt"

// buildLazyTaskContext returns task-scoped resource pointers that clients can
// load on demand instead of preloading all skills and workflow guidance.
func buildLazyTaskContext(task Todo2Task) map[string]interface{} {
	resourceURIs := []string{
		fmt.Sprintf("stdio://tasks/%s", task.ID),
	}
	skillURIs := skillResourceURIsForTask(task)
	if len(skillURIs) > 0 {
		resourceURIs = append(resourceURIs, skillURIs...)
	}

	return map[string]interface{}{
		"load_strategy":       "on_demand",
		"task_resource_uri":   fmt.Sprintf("stdio://tasks/%s", task.ID),
		"skill_resource_uris": skillURIs,
		"resource_uris":       resourceURIs,
		"reason":              "Load these only if you start or inspect this task to keep prime/session context small.",
	}
}

func skillResourceURIsForTask(task Todo2Task) []string {
	seen := map[string]bool{
		"stdio://cursor/skills/use-exarp-tools": true,
	}
	out := []string{"stdio://cursor/skills/use-exarp-tools"}

	for _, toolID := range GetRecommendedTools(task.Metadata) {
		if uri := skillResourceURIForTool(toolID); uri != "" && !seen[uri] {
			seen[uri] = true
			out = append(out, uri)
		}
	}

	return out
}

func skillResourceURIForTool(toolID string) string {
	switch toolID {
	case "task_workflow", "task_analysis", "task_discovery", "task_execute":
		return "stdio://cursor/skills/task-workflow"
	case "report":
		return "stdio://cursor/skills/report-scorecard"
	case "health":
		return "stdio://cursor/skills/database-maintenance"
	case "session":
		return "stdio://cursor/skills/session-handoff"
	case "lint":
		return "stdio://cursor/skills/lint-docs"
	case "tractatus_thinking":
		return "stdio://cursor/skills/tractatus-decompose"
	case "text_generate":
		return "stdio://cursor/skills/text-generate"
	default:
		return ""
	}
}
