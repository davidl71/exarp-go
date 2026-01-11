package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/davidl71/exarp-go/internal/bridge"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/tools"
)

// RegisterAllResources registers all resources with the server
func RegisterAllResources(server framework.MCPServer) error {
	// Register 6 resources (T-45)

	// stdio://scorecard
	if err := server.RegisterResource(
		"stdio://scorecard",
		"Project Scorecard",
		"Get current project scorecard with all health metrics.",
		"application/json",
		handleScorecard,
	); err != nil {
		return fmt.Errorf("failed to register scorecard resource: %w", err)
	}

	// stdio://memories
	if err := server.RegisterResource(
		"stdio://memories",
		"All Memories",
		"Get all AI session memories - browsable context for session continuity.",
		"application/json",
		handleMemories,
	); err != nil {
		return fmt.Errorf("failed to register memories resource: %w", err)
	}

	// stdio://memories/category/{category}
	if err := server.RegisterResource(
		"stdio://memories/category/{category}",
		"Memories by Category",
		"Get memories filtered by category (debug, research, architecture, preference, insight).",
		"application/json",
		handleMemoriesByCategory,
	); err != nil {
		return fmt.Errorf("failed to register memories by category resource: %w", err)
	}

	// stdio://memories/task/{task_id}
	if err := server.RegisterResource(
		"stdio://memories/task/{task_id}",
		"Memories for Task",
		"Get memories linked to a specific task.",
		"application/json",
		handleMemoriesByTask,
	); err != nil {
		return fmt.Errorf("failed to register memories by task resource: %w", err)
	}

	// stdio://memories/recent
	if err := server.RegisterResource(
		"stdio://memories/recent",
		"Recent Memories",
		"Get memories from the last 24 hours.",
		"application/json",
		handleRecentMemories,
	); err != nil {
		return fmt.Errorf("failed to register recent memories resource: %w", err)
	}

	// stdio://memories/session/{date}
	if err := server.RegisterResource(
		"stdio://memories/session/{date}",
		"Session Memories",
		"Get memories from a specific session date (YYYY-MM-DD format).",
		"application/json",
		handleSessionMemories,
	); err != nil {
		return fmt.Errorf("failed to register session memories resource: %w", err)
	}

	// stdio://prompts
	if err := server.RegisterResource(
		"stdio://prompts",
		"All Prompts",
		"Get all prompts in compact format for discovery.",
		"application/json",
		handleAllPrompts,
	); err != nil {
		return fmt.Errorf("failed to register all prompts resource: %w", err)
	}

	// stdio://prompts/mode/{mode}
	if err := server.RegisterResource(
		"stdio://prompts/mode/{mode}",
		"Prompts by Mode",
		"Get prompts for a workflow mode (daily_checkin, security_review, task_management, etc.).",
		"application/json",
		handlePromptsByMode,
	); err != nil {
		return fmt.Errorf("failed to register prompts by mode resource: %w", err)
	}

	// stdio://prompts/persona/{persona}
	if err := server.RegisterResource(
		"stdio://prompts/persona/{persona}",
		"Prompts by Persona",
		"Get prompts for a persona (developer, pm, qa, reviewer, etc.).",
		"application/json",
		handlePromptsByPersona,
	); err != nil {
		return fmt.Errorf("failed to register prompts by persona resource: %w", err)
	}

	// stdio://prompts/category/{category}
	if err := server.RegisterResource(
		"stdio://prompts/category/{category}",
		"Prompts by Category",
		"Get prompts in a category (workflow, persona, analysis, etc.).",
		"application/json",
		handlePromptsByCategory,
	); err != nil {
		return fmt.Errorf("failed to register prompts by category resource: %w", err)
	}

	// stdio://session/mode
	if err := server.RegisterResource(
		"stdio://session/mode",
		"Session Mode",
		"Get current inferred session mode (AGENT/ASK/MANUAL) with confidence.",
		"application/json",
		handleSessionMode,
	); err != nil {
		return fmt.Errorf("failed to register session mode resource: %w", err)
	}

	// stdio://server/status
	if err := server.RegisterResource(
		"stdio://server/status",
		"Server Status",
		"Get server operational status, version, and project root information.",
		"application/json",
		handleServerStatus,
	); err != nil {
		return fmt.Errorf("failed to register server status resource: %w", err)
	}

	// stdio://models
	if err := server.RegisterResource(
		"stdio://models",
		"AI Models",
		"Get all available AI models with capabilities.",
		"application/json",
		handleModels,
	); err != nil {
		return fmt.Errorf("failed to register models resource: %w", err)
	}

	// stdio://tools
	if err := server.RegisterResource(
		"stdio://tools",
		"All Tools",
		"Get all tools in the catalog.",
		"application/json",
		handleAllTools,
	); err != nil {
		return fmt.Errorf("failed to register all tools resource: %w", err)
	}

	// stdio://tools/{category}
	if err := server.RegisterResource(
		"stdio://tools/{category}",
		"Tools by Category",
		"Get tools filtered by category.",
		"application/json",
		handleToolsByCategory,
	); err != nil {
		return fmt.Errorf("failed to register tools by category resource: %w", err)
	}

	// stdio://tasks
	if err := server.RegisterResource(
		"stdio://tasks",
		"All Tasks",
		"Get all tasks (read-only, paginated with limit of 50).",
		"application/json",
		handleAllTasks,
	); err != nil {
		return fmt.Errorf("failed to register all tasks resource: %w", err)
	}

	// stdio://tasks/{task_id}
	if err := server.RegisterResource(
		"stdio://tasks/{task_id}",
		"Task by ID",
		"Get a specific task by ID with full details.",
		"application/json",
		handleTaskByID,
	); err != nil {
		return fmt.Errorf("failed to register task by ID resource: %w", err)
	}

	// stdio://tasks/status/{status}
	if err := server.RegisterResource(
		"stdio://tasks/status/{status}",
		"Tasks by Status",
		"Get tasks filtered by status (Todo, In Progress, Done).",
		"application/json",
		handleTasksByStatus,
	); err != nil {
		return fmt.Errorf("failed to register tasks by status resource: %w", err)
	}

	// stdio://tasks/priority/{priority}
	if err := server.RegisterResource(
		"stdio://tasks/priority/{priority}",
		"Tasks by Priority",
		"Get tasks filtered by priority (low, medium, high, critical).",
		"application/json",
		handleTasksByPriority,
	); err != nil {
		return fmt.Errorf("failed to register tasks by priority resource: %w", err)
	}

	// stdio://tasks/tag/{tag}
	if err := server.RegisterResource(
		"stdio://tasks/tag/{tag}",
		"Tasks by Tag",
		"Get tasks filtered by tag (any tag value).",
		"application/json",
		handleTasksByTag,
	); err != nil {
		return fmt.Errorf("failed to register tasks by tag resource: %w", err)
	}

	// stdio://tasks/summary
	if err := server.RegisterResource(
		"stdio://tasks/summary",
		"Tasks Summary",
		"Get task statistics and overview (counts by status, priority, tags).",
		"application/json",
		handleTasksSummary,
	); err != nil {
		return fmt.Errorf("failed to register tasks summary resource: %w", err)
	}

	return nil
}

// Resource handlers
// Note: handleScorecard and memory handlers are implemented in scorecard.go and memories.go
// They are in the same package, so they're automatically available here

// handleAllPrompts handles the stdio://prompts resource
// Uses native Go prompts from internal/prompts/templates.go
func handleAllPrompts(ctx context.Context, uri string) ([]byte, string, error) {
	// Get all prompts from native Go templates
	allPrompts := getAllPromptsNative()
	
	// Build categories mapping
	byCategory := make(map[string][]map[string]interface{})
	for name, desc := range allPrompts {
		category := categorizePrompt(name, desc)
		if byCategory[category] == nil {
			byCategory[category] = []map[string]interface{}{}
		}
		byCategory[category] = append(byCategory[category], map[string]interface{}{
			"name":        name,
			"description": desc,
		})
	}

	result := map[string]interface{}{
		"prompts":       formatPromptsForResource(allPrompts),
		"total":         len(allPrompts),
		"by_category":   byCategory,
		"categories":    getPromptCategories(),
		"timestamp":     time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal prompts: %w", err)
	}

	return jsonData, "application/json", nil
}

// handlePromptsByMode handles the stdio://prompts/mode/{mode} resource
// Uses native Go prompts filtered by mode
func handlePromptsByMode(ctx context.Context, uri string) ([]byte, string, error) {
	// Parse mode from URI: stdio://prompts/mode/{mode}
	mode, err := parseURIVariableByIndexWithValidation(uri, 3, "mode", "stdio://prompts/mode/{mode}")
	if err != nil {
		return nil, "", err
	}

	// Get prompts for mode
	modePrompts := getPromptsForMode(mode)
	
	result := map[string]interface{}{
		"mode":       mode,
		"prompts":    formatPromptsForResource(modePrompts),
		"count":      len(modePrompts),
		"available_modes": getAvailableModes(),
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal prompts: %w", err)
	}

	return jsonData, "application/json", nil
}

// handlePromptsByPersona handles the stdio://prompts/persona/{persona} resource
// Uses native Go prompts filtered by persona
func handlePromptsByPersona(ctx context.Context, uri string) ([]byte, string, error) {
	// Parse persona from URI: stdio://prompts/persona/{persona}
	persona, err := parseURIVariableByIndexWithValidation(uri, 3, "persona", "stdio://prompts/persona/{persona}")
	if err != nil {
		return nil, "", err
	}

	// Get prompts for persona
	personaPrompts := getPromptsForPersona(persona)
	
	result := map[string]interface{}{
		"persona":        persona,
		"prompts":        formatPromptsForResource(personaPrompts),
		"count":          len(personaPrompts),
		"available_personas": getAvailablePersonas(),
		"timestamp":      time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal prompts: %w", err)
	}

	return jsonData, "application/json", nil
}

// handlePromptsByCategory handles the stdio://prompts/category/{category} resource
// Uses native Go prompts filtered by category
func handlePromptsByCategory(ctx context.Context, uri string) ([]byte, string, error) {
	// Parse category from URI: stdio://prompts/category/{category}
	category, err := parseURIVariableByIndexWithValidation(uri, 3, "category", "stdio://prompts/category/{category}")
	if err != nil {
		return nil, "", err
	}

	// Get prompts for category
	categoryPrompts := getPromptsForCategory(category)
	
	result := map[string]interface{}{
		"category":         category,
		"prompts":          formatPromptsForResource(categoryPrompts),
		"count":            len(categoryPrompts),
		"available_categories": getPromptCategories(),
		"timestamp":        time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal prompts: %w", err)
	}

	return jsonData, "application/json", nil
}

// handleSessionMode handles the stdio://session/mode resource
// Uses native Go infer_session_mode tool logic
func handleSessionMode(ctx context.Context, uri string) ([]byte, string, error) {
	// Use native Go implementation from infer_session_mode tool
	params := map[string]interface{}{
		"force_recompute": false,
	}
	
	// Call the exported function from tools package
	result, err := tools.HandleInferSessionModeNative(ctx, params)
	if err != nil {
		// Fallback to Python bridge if native fails
		return bridge.ExecutePythonResource(ctx, uri)
	}

	// Extract text from TextContent response
	if len(result) > 0 && result[0].Type == "text" {
		return []byte(result[0].Text), "application/json", nil
	}

	// Fallback to Python bridge if response format is unexpected
	return bridge.ExecutePythonResource(ctx, uri)
}
