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
	// Get all prompt names from registry (matching the order in registry.go)
	promptNames := []string{
		// Original prompts (8)
		"align", "discover", "config", "scan", "scorecard", "overview",
		"dashboard", "remember",
		// High-value workflow prompts (7)
		"daily_checkin", "sprint_start", "sprint_end", "pre_sprint",
		"post_impl", "sync", "dups",
		// mcp-generic-tools prompts
		"context", "mode",
		// Task management prompts
		"task_update",
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
		"daily_checkin": "daily_checkin",
		"sprint_start":  "sprint_management",
		"sprint_end":    "sprint_management",
		"pre_sprint":    "sprint_management",
		"post_impl":     "sprint_management",
		// Task management prompts
		"align":       "task_management",
		"discover":    "task_management",
		"sync":        "task_management",
		"dups":        "task_management",
		"task_update": "task_management",
		// Configuration prompts
		"config": "configuration",
		// Security prompts
		"scan": "security_review",
		// Reporting prompts
		"scorecard": "reporting",
		"overview":  "reporting",
		"dashboard": "reporting",
		// Memory prompts
		"remember": "memory",
		// Utility prompts
		"context": "utility",
		"mode":    "utility",
	}
}

// getPromptPersonaMapping returns a map of prompt names to their personas
func getPromptPersonaMapping() map[string]string {
	return map[string]string{
		// PM/Management prompts
		"daily_checkin": "pm",
		"sprint_start":  "pm",
		"sprint_end":    "pm",
		"pre_sprint":    "pm",
		"post_impl":     "pm",
		"scorecard":     "pm",
		"overview":      "pm",
		"dashboard":     "pm",
		// Developer prompts
		"align":       "developer",
		"discover":    "developer",
		"task_update": "developer",
		"sync":        "developer",
		"dups":        "developer",
		"config":      "developer",
		"remember":    "developer",
		"context":     "developer",
		"mode":        "developer",
		// Security prompts
		"scan": "security",
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
		// Workflow
		"daily_checkin": "workflow",
		"sprint_start":  "workflow",
		"sprint_end":    "workflow",
		"pre_sprint":    "workflow",
		"post_impl":     "workflow",
		// Configuration
		"config": "configuration",
		// Security
		"scan": "security",
		// Reporting
		"scorecard": "reporting",
		"overview":  "reporting",
		"dashboard": "reporting",
		// Memory
		"remember": "memory",
		// Utility
		"context": "utility",
		"mode":    "utility",
	}
}
