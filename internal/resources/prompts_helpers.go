package resources

import (
	"strings"

	"github.com/davidl71/exarp-go/internal/prompts"
)

// getAllPromptsNative retrieves all prompts from native Go templates.
func getAllPromptsNative() map[string]string {
	// All 34 prompt names (same order as internal/prompts/registry.go)
	promptNames := []string{
		"align", "discover", "config", "scan", "scorecard", "overview", "dashboard", "remember",
		"daily_checkin", "sprint_start", "sprint_end", "pre_sprint", "post_impl", "sync", "dups",
		"context", "mode", "task_update",
		"docs", "automation_discover", "weekly_maintenance", "task_review", "project_health",
		"automation_setup", "advisor_consult", "advisor_briefing",
		"persona_developer", "persona_project_manager", "persona_code_reviewer", "persona_executive",
		"persona_security", "persona_architect", "persona_qa", "persona_tech_writer",
	}

	result := make(map[string]string)

	for _, name := range promptNames {
		template, err := prompts.GetPromptTemplate(name)
		if err == nil {
			// Extract first line or short description
			desc := extractDescription(template)
			result[name] = desc
		}
	}

	return result
}

// extractDescription extracts a short description from a prompt template.
func extractDescription(template string) string {
	lines := strings.Split(template, "\n")
	if len(lines) > 0 {
		// Get first meaningful line (skip empty lines and markdown headers)
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "**") {
				// Remove markdown formatting
				line = strings.TrimPrefix(line, "**")
				line = strings.TrimSuffix(line, "**")

				if len(line) > 100 {
					line = line[:100] + "..."
				}

				return line
			}
		}
	}

	return "Prompt template"
}

// categorizePrompt categorizes a prompt based on its name and description.
func categorizePrompt(name, desc string) string {
	lowerName := strings.ToLower(name)
	lowerDesc := strings.ToLower(desc)

	// Workflow prompts
	if strings.Contains(lowerName, "checkin") || strings.Contains(lowerName, "sprint") ||
		strings.Contains(lowerName, "pre_sprint") || strings.Contains(lowerName, "post_impl") {
		return "workflow"
	}

	// Analysis prompts
	if strings.Contains(lowerName, "align") || strings.Contains(lowerName, "discover") ||
		strings.Contains(lowerName, "sync") || strings.Contains(lowerName, "dups") {
		return "analysis"
	}

	// Configuration prompts
	if strings.Contains(lowerName, "config") || strings.Contains(lowerName, "mode") {
		return "configuration"
	}

	// Reporting prompts
	if strings.Contains(lowerName, "scorecard") || strings.Contains(lowerName, "overview") ||
		strings.Contains(lowerName, "dashboard") {
		return "reporting"
	}

	// Security prompts
	if strings.Contains(lowerName, "scan") || strings.Contains(lowerDesc, "security") ||
		strings.Contains(lowerDesc, "vulnerability") {
		return "security"
	}

	// Memory prompts
	if strings.Contains(lowerName, "remember") || strings.Contains(lowerDesc, "memory") {
		return "memory"
	}

	// Context management prompts
	if strings.Contains(lowerName, "context") {
		return "context"
	}

	// Task management prompts
	if strings.Contains(lowerName, "task") || strings.Contains(lowerDesc, "task") {
		return "task_management"
	}

	// Default category
	return "general"
}

// getPromptsForMode returns prompts for a specific workflow mode.
func getPromptsForMode(mode string) map[string]string {
	allPrompts := getAllPromptsNative()
	result := make(map[string]string)

	// Mode mappings based on prompt names
	modeMappings := map[string][]string{
		"daily_checkin":   {"daily_checkin", "remember", "scorecard", "task_update"},
		"sprint_start":    {"sprint_start", "align", "dups", "sync"},
		"sprint_end":      {"sprint_end", "scorecard", "overview", "dashboard"},
		"pre_sprint":      {"pre_sprint", "dups", "align", "config"},
		"post_impl":       {"post_impl", "remember", "task_update", "config"},
		"security_review": {"scan", "scorecard", "overview"},
		"task_management": {"discover", "sync", "dups", "task_update", "align"},
		"development":     {"discover", "config", "remember", "mode", "context"},
	}

	promptNames, ok := modeMappings[mode]
	if !ok {
		// Unknown mode - return all prompts
		return allPrompts
	}

	for _, name := range promptNames {
		if desc, ok := allPrompts[name]; ok {
			result[name] = desc
		}
	}

	// If no specific mappings found, try "development" as fallback
	if len(result) == 0 {
		fallbackNames := modeMappings["development"]
		for _, name := range fallbackNames {
			if desc, ok := allPrompts[name]; ok {
				result[name] = desc
			}
		}
	}

	return result
}

// getPromptsForPersona returns prompts for a specific persona.
func getPromptsForPersona(persona string) map[string]string {
	allPrompts := getAllPromptsNative()
	result := make(map[string]string)

	// Persona mappings
	personaMappings := map[string][]string{
		"developer":   {"discover", "config", "remember", "context", "mode"},
		"pm":          {"scorecard", "overview", "dashboard", "align", "sync"},
		"qa":          {"sprint_end", "scorecard", "scan", "task_update"},
		"reviewer":    {"align", "dups", "scorecard", "overview"},
		"security":    {"scan", "scorecard", "overview"},
		"architect":   {"align", "overview", "dashboard", "context"},
		"executive":   {"scorecard", "overview", "dashboard"},
		"tech_writer": {"config", "remember", "overview", "dashboard"},
	}

	promptNames, ok := personaMappings[strings.ToLower(persona)]
	if !ok {
		// Unknown persona - return all prompts
		return allPrompts
	}

	for _, name := range promptNames {
		if desc, ok := allPrompts[name]; ok {
			result[name] = desc
		}
	}

	return result
}

// getPromptsForCategory returns prompts for a specific category.
func getPromptsForCategory(category string) map[string]string {
	allPrompts := getAllPromptsNative()
	result := make(map[string]string)

	for name, desc := range allPrompts {
		if categorizePrompt(name, desc) == category {
			result[name] = desc
		}
	}

	return result
}

// getPromptCategories returns all available prompt categories.
func getPromptCategories() []string {
	return []string{
		"workflow", "analysis", "configuration", "reporting",
		"security", "memory", "context", "task_management", "general",
	}
}

// getAvailableModes returns all available workflow modes.
func getAvailableModes() []string {
	return []string{
		"daily_checkin", "sprint_start", "sprint_end", "pre_sprint",
		"post_impl", "security_review", "task_management", "development",
	}
}

// getAvailablePersonas returns all available personas.
func getAvailablePersonas() []string {
	return []string{
		"developer", "pm", "qa", "reviewer", "security",
		"architect", "executive", "tech_writer",
	}
}

// formatPromptsForResource formats prompts for resource output.
func formatPromptsForResource(prompts map[string]string) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(prompts))
	for name, desc := range prompts {
		result = append(result, map[string]interface{}{
			"name":        name,
			"description": desc,
		})
	}

	return result
}
