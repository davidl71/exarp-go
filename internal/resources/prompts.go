package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/prompts"
)

// handleAllPrompts handles the stdio://prompts resource
// Returns all prompts in compact format (name + description only)
func handleAllPrompts(ctx context.Context, uri string) ([]byte, string, error) {
	// All prompt names in registry order (35 prompts; see internal/prompts/registry.go)
	promptNames := []string{
		// Original prompts (9)
		"align", "discover", "config", "scan", "scorecard", "overview", "plan",
		"dashboard", "remember",
		// High-value workflow prompts (7)
		"daily_checkin", "sprint_start", "sprint_end", "pre_sprint",
		"post_impl", "sync", "dups",
		// mcp-generic-tools prompts
		"context", "mode",
		// Task management prompts
		"task_update",
		// Migrated from Python (16): docs, automation, workflow, advisor, personas
		"docs", "automation_discover", "weekly_maintenance", "task_review", "project_health",
		"automation_setup", "advisor_consult", "advisor_briefing",
		"persona_developer", "persona_project_manager", "persona_code_reviewer", "persona_executive",
		"persona_security", "persona_architect", "persona_qa", "persona_tech_writer",
	}

	// Build compact format (name + description only)
	promptList := make([]map[string]interface{}, 0, len(promptNames))
	for _, name := range promptNames {
		template, err := prompts.GetPromptTemplate(name)
		if err != nil {
			// Skip if prompt not found (shouldn't happen, but handle gracefully)
			continue
		}

		// Extract first line as description (before first newline or first period)
		description := extractPromptDescription(template)

		promptList = append(promptList, map[string]interface{}{
			"name":        name,
			"description": description,
		})
	}

	result := map[string]interface{}{
		"prompts":   promptList,
		"count":     len(promptList),
		"timestamp": time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal prompts: %w", err)
	}

	return jsonData, "application/json", nil
}

// handlePromptsByMode handles the stdio://prompts/mode/{mode} resource
// Returns prompts filtered by workflow mode
func handlePromptsByMode(ctx context.Context, uri string) ([]byte, string, error) {
	// Parse mode from URI: stdio://prompts/mode/{mode}
	parts := strings.Split(uri, "/")
	if len(parts) < 4 {
		return nil, "", fmt.Errorf("invalid URI format: %s (expected stdio://prompts/mode/{mode})", uri)
	}
	mode := parts[3]

	// Get mode mapping
	modeMap := getPromptModeMapping()

	// Find prompts matching this mode
	matchingPrompts := []string{}
	for promptName, promptMode := range modeMap {
		if promptMode == mode {
			matchingPrompts = append(matchingPrompts, promptName)
		}
	}

	// Build response
	promptList := make([]map[string]interface{}, 0, len(matchingPrompts))
	for _, name := range matchingPrompts {
		template, err := prompts.GetPromptTemplate(name)
		if err != nil {
			continue
		}

		description := extractPromptDescription(template)

		promptList = append(promptList, map[string]interface{}{
			"name":        name,
			"description": description,
			"mode":        mode,
		})
	}

	result := map[string]interface{}{
		"mode":      mode,
		"prompts":   promptList,
		"count":     len(promptList),
		"timestamp": time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal prompts: %w", err)
	}

	return jsonData, "application/json", nil
}

// handlePromptsByPersona handles the stdio://prompts/persona/{persona} resource
// Returns prompts filtered by persona
func handlePromptsByPersona(ctx context.Context, uri string) ([]byte, string, error) {
	// Parse persona from URI: stdio://prompts/persona/{persona}
	parts := strings.Split(uri, "/")
	if len(parts) < 4 {
		return nil, "", fmt.Errorf("invalid URI format: %s (expected stdio://prompts/persona/{persona})", uri)
	}
	persona := parts[3]

	// Get persona mapping
	personaMap := getPromptPersonaMapping()

	// Find prompts matching this persona
	matchingPrompts := []string{}
	for promptName, promptPersona := range personaMap {
		if promptPersona == persona {
			matchingPrompts = append(matchingPrompts, promptName)
		}
	}

	// Build response
	promptList := make([]map[string]interface{}, 0, len(matchingPrompts))
	for _, name := range matchingPrompts {
		template, err := prompts.GetPromptTemplate(name)
		if err != nil {
			continue
		}

		description := extractPromptDescription(template)

		promptList = append(promptList, map[string]interface{}{
			"name":        name,
			"description": description,
			"persona":     persona,
		})
	}

	result := map[string]interface{}{
		"persona":   persona,
		"prompts":   promptList,
		"count":     len(promptList),
		"timestamp": time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal prompts: %w", err)
	}

	return jsonData, "application/json", nil
}

// handlePromptsByCategory handles the stdio://prompts/category/{category} resource
// Returns prompts filtered by category
func handlePromptsByCategory(ctx context.Context, uri string) ([]byte, string, error) {
	// Parse category from URI: stdio://prompts/category/{category}
	parts := strings.Split(uri, "/")
	if len(parts) < 4 {
		return nil, "", fmt.Errorf("invalid URI format: %s (expected stdio://prompts/category/{category})", uri)
	}
	category := parts[3]

	// Get category mapping
	categoryMap := getPromptCategoryMapping()

	// Find prompts matching this category
	matchingPrompts := []string{}
	for promptName, promptCategory := range categoryMap {
		if promptCategory == category {
			matchingPrompts = append(matchingPrompts, promptName)
		}
	}

	// Build response
	promptList := make([]map[string]interface{}, 0, len(matchingPrompts))
	for _, name := range matchingPrompts {
		template, err := prompts.GetPromptTemplate(name)
		if err != nil {
			continue
		}

		description := extractPromptDescription(template)

		promptList = append(promptList, map[string]interface{}{
			"name":        name,
			"description": description,
			"category":    category,
		})
	}

	result := map[string]interface{}{
		"category":  category,
		"prompts":   promptList,
		"count":     len(promptList),
		"timestamp": time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal prompts: %w", err)
	}

	return jsonData, "application/json", nil
}

// Helper functions

// extractPromptDescription extracts a short description from a prompt template
// Returns the first line (before first newline) or first sentence (before first period)
func extractPromptDescription(template string) string {
	// Remove leading/trailing whitespace
	template = strings.TrimSpace(template)

	// Try to get first line (before first newline)
	firstLine := strings.Split(template, "\n")[0]
	firstLine = strings.TrimSpace(firstLine)

	// If first line is too long (>100 chars), truncate at first sentence
	if len(firstLine) > 100 {
		// Look for first period followed by space
		periodIdx := strings.Index(firstLine, ". ")
		if periodIdx > 0 && periodIdx < 100 {
			return firstLine[:periodIdx+1]
		}
		// Otherwise truncate at 100 chars with ellipsis
		if len(firstLine) > 100 {
			return firstLine[:100] + "..."
		}
	}

	return firstLine
}

// getPromptModeMapping returns a map of prompt names to their workflow modes
func getPromptModeMapping() map[string]string {
	return map[string]string{
		// Workflow prompts
		"daily_checkin":      "daily_checkin",
		"sprint_start":       "sprint_management",
		"sprint_end":         "sprint_management",
		"pre_sprint":         "sprint_management",
		"post_impl":          "sprint_management",
		"weekly_maintenance": "sprint_management",
		"task_review":        "sprint_management",
		// Task management prompts
		"align":       "task_management",
		"discover":    "task_management",
		"sync":        "task_management",
		"dups":        "task_management",
		"task_update": "task_management",
		// Configuration / automation
		"config":              "configuration",
		"automation_setup":    "configuration",
		"automation_discover": "configuration",
		// Security prompts
		"scan": "security_review",
		// Reporting / health
		"scorecard":      "reporting",
		"overview":       "reporting",
		"plan":           "reporting",
		"dashboard":      "reporting",
		"project_health": "reporting",
		"docs":           "reporting",
		// Advisor
		"advisor_consult":  "advisor",
		"advisor_briefing": "advisor",
		// Memory prompts
		"remember": "memory",
		// Utility prompts
		"context": "utility",
		"mode":    "utility",
		// Personas (by role)
		"persona_developer":       "developer",
		"persona_project_manager": "pm",
		"persona_code_reviewer":   "developer",
		"persona_executive":       "executive",
		"persona_security":        "security",
		"persona_architect":       "architecture",
		"persona_qa":              "qa",
		"persona_tech_writer":     "documentation",
	}
}

// getPromptPersonaMapping returns a map of prompt names to their personas
func getPromptPersonaMapping() map[string]string {
	return map[string]string{
		// PM/Management prompts
		"daily_checkin":      "pm",
		"sprint_start":       "pm",
		"sprint_end":         "pm",
		"pre_sprint":         "pm",
		"post_impl":          "pm",
		"scorecard":          "pm",
		"overview":           "pm",
		"plan":               "pm",
		"dashboard":          "pm",
		"weekly_maintenance": "pm",
		"task_review":        "pm",
		"project_health":     "pm",
		"docs":               "pm",
		// Developer prompts
		"align":                 "developer",
		"discover":              "developer",
		"task_update":           "developer",
		"sync":                  "developer",
		"dups":                  "developer",
		"config":                "developer",
		"remember":              "developer",
		"context":               "developer",
		"mode":                  "developer",
		"automation_setup":      "developer",
		"automation_discover":   "developer",
		"persona_developer":     "developer",
		"persona_code_reviewer": "developer",
		// Security prompts
		"scan":             "security",
		"persona_security": "security",
		// Advisor
		"advisor_consult":         "pm",
		"advisor_briefing":        "pm",
		"persona_project_manager": "pm",
		"persona_executive":       "executive",
		"persona_architect":       "architect",
		"persona_qa":              "qa",
		"persona_tech_writer":     "tech_writer",
	}
}

// getPromptCategoryMapping returns a map of prompt names to their categories
func getPromptCategoryMapping() map[string]string {
	return map[string]string{
		// Task management
		"align":       "task_management",
		"discover":    "task_management",
		"sync":        "task_management",
		"dups":        "task_management",
		"task_update": "task_management",
		"task_review": "task_management",
		// Workflow
		"daily_checkin":      "workflow",
		"sprint_start":       "workflow",
		"sprint_end":         "workflow",
		"pre_sprint":         "workflow",
		"post_impl":          "workflow",
		"weekly_maintenance": "workflow",
		// Configuration / automation
		"config":              "configuration",
		"automation_setup":    "configuration",
		"automation_discover": "configuration",
		// Security
		"scan": "security",
		// Reporting / health
		"scorecard":      "reporting",
		"overview":       "reporting",
		"plan":           "reporting",
		"dashboard":      "reporting",
		"project_health": "reporting",
		"docs":           "reporting",
		// Advisor
		"advisor_consult":  "advisor",
		"advisor_briefing": "advisor",
		// Memory
		"remember": "memory",
		// Utility
		"context": "utility",
		"mode":    "utility",
		// Personas
		"persona_developer":       "persona",
		"persona_project_manager": "persona",
		"persona_code_reviewer":   "persona",
		"persona_executive":       "persona",
		"persona_security":        "persona",
		"persona_architect":       "persona",
		"persona_qa":              "persona",
		"persona_tech_writer":     "persona",
	}
}
