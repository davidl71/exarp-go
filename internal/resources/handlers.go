// handlers.go — MCP resource registry: registers all stdio:// resource handlers.
//
// Package resources provides MCP resource handlers for stdio:// URIs (tasks, tools, session, etc.).
package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/tools"
)

// registerAndTrack registers a resource with the server and tracks it for ResourcesAsTools.
func registerAndTrack(server framework.MCPServer, uri, name, description, mimeType string, handler framework.ResourceHandler) error {
	if err := server.RegisterResource(uri, name, description, mimeType, handler); err != nil {
		return err
	}
	tools.TrackResource(uri, name, description, mimeType, handler)
	return nil
}

// registerTemplate registers a resource template for parameterized URI discovery.
func registerTemplate(server framework.MCPServer, uriTemplate, name, description, mimeType string, handler framework.ResourceHandler) error {
	registrar, ok := any(server).(framework.ResourceTemplateRegistrar)
	if !ok {
		return nil
	}
	return registrar.RegisterResourceTemplate(uriTemplate, name, description, mimeType, handler)
}

// handleConfig handles the stdio://config resource
// Returns current configuration as JSON.
func handleConfig(ctx context.Context, uri string) ([]byte, string, error) {
	projectRoot, err := tools.FindProjectRoot()
	if err != nil {
		return nil, "", fmt.Errorf("failed to find project root: %w", err)
	}

	cfg, err := config.LoadConfig(projectRoot)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load config: %w", err)
	}

	jsonData, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, "", err
	}

	return jsonData, "application/json", nil
}

// handleConfigSchema handles the stdio://config/schema resource
// Returns available config fields and their types.
func handleConfigSchema(ctx context.Context, uri string) ([]byte, string, error) {
	schema := map[string]interface{}{
		"version": map[string]string{
			"type":        "string",
			"description": "Config version",
		},
		"timeouts": map[string]interface{}{
			"type": "object",
			"fields": map[string]string{
				"task_lock_lease":      "duration",
				"task_lock_renewal":    "duration",
				"stale_lock_threshold": "duration",
				"tool_default":         "duration",
				"tool_scorecard":       "duration",
				"tool_linting":         "duration",
				"tool_testing":         "duration",
				"tool_report":          "duration",
				"ollama_download":      "duration",
				"ollama_generate":      "duration",
				"http_client":          "duration",
				"database_retry":       "duration",
				"context_summarize":    "duration",
				"context_budget":       "duration",
			},
		},
		"thresholds": map[string]interface{}{
			"type": "object",
			"fields": map[string]string{
				"similarity_threshold":   "float",
				"min_coverage":           "int",
				"min_task_confidence":    "float",
				"min_test_confidence":    "float",
				"min_description_length": "int",
			},
		},
		"tasks": map[string]interface{}{
			"type": "object",
			"fields": map[string]string{
				"default_status":   "string",
				"default_priority": "string",
			},
		},
		"project": map[string]interface{}{
			"type": "object",
			"fields": map[string]string{
				"name":                        "string",
				"type":                        "string",
				"language":                    "string",
				"root":                        "string",
				"task_discovery_ignore_paths": "string[]",
			},
		},
		"database": map[string]interface{}{
			"type": "object",
			"fields": map[string]string{
				"sqlite_path":        "string",
				"json_fallback_path": "string",
				"backup_path":        "string",
				"max_connections":    "int",
				"connection_timeout": "duration",
				"query_timeout":      "duration",
				"retry_attempts":     "int",
				"auto_vacuum":        "bool",
				"wal_mode":           "bool",
			},
		},
		"security": map[string]interface{}{
			"type": "object",
			"fields": map[string]string{
				"rate_limit":      "object",
				"path_validation": "object",
				"file_limits":     "object",
				"access_control":  "object",
			},
		},
		"logging": map[string]interface{}{
			"type": "object",
			"fields": map[string]string{
				"level":          "string",
				"format":         "string",
				"log_dir":        "string",
				"log_file":       "string",
				"color_output":   "bool",
				"retention_days": "int",
			},
		},
		"tools": map[string]interface{}{
			"type": "object",
			"fields": map[string]string{
				"scorecard": "object",
				"report":    "object",
				"linting":   "object",
				"testing":   "object",
				"mlx":       "object",
				"ollama":    "object",
				"context":   "object",
			},
		},
		"workflow": map[string]interface{}{
			"type": "object",
			"fields": map[string]string{
				"default_mode":     "string",
				"auto_detect_mode": "bool",
			},
		},
		"memory": map[string]interface{}{
			"type": "object",
			"fields": map[string]string{
				"storage_path":     "string",
				"session_log_path": "string",
				"retention_days":   "int",
				"auto_cleanup":     "bool",
				"max_memories":     "int",
			},
		},
	}

	jsonData, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return nil, "", err
	}

	return jsonData, "application/json", nil
}

// RegisterAllResources registers all resources with the server.
func RegisterAllResources(server framework.MCPServer) error {
	// Register config resources first
	// stdio://config
	if err := registerAndTrack(server,
		"stdio://config",
		"Current Configuration",
		"Get current configuration values as JSON.",
		"application/json",
		handleConfig,
	); err != nil {
		return fmt.Errorf("failed to register config resource: %w", err)
	}

	// stdio://config/schema
	if err := registerAndTrack(server,
		"stdio://config/schema",
		"Configuration Schema",
		"Get configuration schema with available fields and their types.",
		"application/json",
		handleConfigSchema,
	); err != nil {
		return fmt.Errorf("failed to register config schema resource: %w", err)
	}

	// Register 24 resources (all native Go; no Python bridge)
	// stdio://scorecard
	if err := registerAndTrack(server,
		"stdio://scorecard",
		"Project Scorecard",
		"Get current project scorecard with all health metrics.",
		"application/json",
		handleScorecard,
	); err != nil {
		return fmt.Errorf("failed to register scorecard resource: %w", err)
	}

	// stdio://memories
	if err := registerAndTrack(server,
		"stdio://memories",
		"All Memories",
		"Get all AI session memories - browsable context for session continuity.",
		"application/json",
		handleMemories,
	); err != nil {
		return fmt.Errorf("failed to register memories resource: %w", err)
	}

	// stdio://memories/category/{category}
	if err := registerAndTrack(server,
		"stdio://memories/category/{category}",
		"Memories by Category",
		"Get memories filtered by category (debug, research, architecture, preference, insight).",
		"application/json",
		handleMemoriesByCategory,
	); err != nil {
		return fmt.Errorf("failed to register memories by category resource: %w", err)
	}
	if err := registerTemplate(server,
		"stdio://memories/category/{category}",
		"Memories by Category",
		"Get memories filtered by category (debug, research, architecture, preference, insight).",
		"application/json",
		handleMemoriesByCategory,
	); err != nil {
		return fmt.Errorf("failed to register memories by category template: %w", err)
	}

	// stdio://memories/task/{task_id}
	if err := registerAndTrack(server,
		"stdio://memories/task/{task_id}",
		"Memories for Task",
		"Get memories linked to a specific task.",
		"application/json",
		handleMemoriesByTask,
	); err != nil {
		return fmt.Errorf("failed to register memories by task resource: %w", err)
	}
	if err := registerTemplate(server,
		"stdio://memories/task/{task_id}",
		"Memories for Task",
		"Get memories linked to a specific task.",
		"application/json",
		handleMemoriesByTask,
	); err != nil {
		return fmt.Errorf("failed to register memories by task template: %w", err)
	}

	// stdio://memories/recent
	if err := registerAndTrack(server,
		"stdio://memories/recent",
		"Recent Memories",
		"Get memories from the last 24 hours.",
		"application/json",
		handleRecentMemories,
	); err != nil {
		return fmt.Errorf("failed to register recent memories resource: %w", err)
	}

	// stdio://memories/session/{date}
	if err := registerAndTrack(server,
		"stdio://memories/session/{date}",
		"Session Memories",
		"Get memories from a specific session date (YYYY-MM-DD format).",
		"application/json",
		handleSessionMemories,
	); err != nil {
		return fmt.Errorf("failed to register session memories resource: %w", err)
	}
	if err := registerTemplate(server,
		"stdio://memories/session/{date}",
		"Session Memories",
		"Get memories from a specific session date (YYYY-MM-DD format).",
		"application/json",
		handleSessionMemories,
	); err != nil {
		return fmt.Errorf("failed to register session memories template: %w", err)
	}

	// stdio://prompts
	if err := registerAndTrack(server,
		"stdio://prompts",
		"All Prompts",
		"Get all prompts in compact format for discovery.",
		"application/json",
		handleAllPrompts,
	); err != nil {
		return fmt.Errorf("failed to register all prompts resource: %w", err)
	}

	// stdio://prompts/mode/{mode}
	if err := registerAndTrack(server,
		"stdio://prompts/mode/{mode}",
		"Prompts by Mode",
		"Get prompts for a workflow mode (daily_checkin, security_review, task_management, etc.).",
		"application/json",
		handlePromptsByMode,
	); err != nil {
		return fmt.Errorf("failed to register prompts by mode resource: %w", err)
	}
	if err := registerTemplate(server,
		"stdio://prompts/mode/{mode}",
		"Prompts by Mode",
		"Get prompts for a workflow mode (daily_checkin, security_review, task_management, etc.).",
		"application/json",
		handlePromptsByMode,
	); err != nil {
		return fmt.Errorf("failed to register prompts by mode template: %w", err)
	}

	// stdio://prompts/persona/{persona}
	if err := registerAndTrack(server,
		"stdio://prompts/persona/{persona}",
		"Prompts by Persona",
		"Get prompts for a persona (developer, pm, qa, reviewer, etc.).",
		"application/json",
		handlePromptsByPersona,
	); err != nil {
		return fmt.Errorf("failed to register prompts by persona resource: %w", err)
	}
	if err := registerTemplate(server,
		"stdio://prompts/persona/{persona}",
		"Prompts by Persona",
		"Get prompts for a persona (developer, pm, qa, reviewer, etc.).",
		"application/json",
		handlePromptsByPersona,
	); err != nil {
		return fmt.Errorf("failed to register prompts by persona template: %w", err)
	}

	// stdio://prompts/category/{category}
	if err := registerAndTrack(server,
		"stdio://prompts/category/{category}",
		"Prompts by Category",
		"Get prompts in a category (workflow, persona, analysis, etc.).",
		"application/json",
		handlePromptsByCategory,
	); err != nil {
		return fmt.Errorf("failed to register prompts by category resource: %w", err)
	}
	if err := registerTemplate(server,
		"stdio://prompts/category/{category}",
		"Prompts by Category",
		"Get prompts in a category (workflow, persona, analysis, etc.).",
		"application/json",
		handlePromptsByCategory,
	); err != nil {
		return fmt.Errorf("failed to register prompts by category template: %w", err)
	}

	// stdio://session/mode
	if err := registerAndTrack(server,
		"stdio://session/mode",
		"Session Mode",
		"Get current inferred session mode (AGENT/ASK/MANUAL) with confidence.",
		"application/json",
		handleSessionMode,
	); err != nil {
		return fmt.Errorf("failed to register session mode resource: %w", err)
	}

	// stdio://session/status — dynamic context for UI labels (handoff, task, dashboard)
	if err := registerAndTrack(server,
		"stdio://session/status",
		"Session Status",
		"Get current session context (handoff, current task, project dashboard) for dynamic UI labels.",
		"application/json",
		handleSessionStatus,
	); err != nil {
		return fmt.Errorf("failed to register session status resource: %w", err)
	}

	// stdio://server/status
	if err := registerAndTrack(server,
		"stdio://server/status",
		"Server Status",
		"Get server operational status, version, and project root information.",
		"application/json",
		handleServerStatus,
	); err != nil {
		return fmt.Errorf("failed to register server status resource: %w", err)
	}

	// stdio://models
	if err := registerAndTrack(server,
		"stdio://models",
		"AI Models",
		"Get all available AI models with capabilities.",
		"application/json",
		handleModels,
	); err != nil {
		return fmt.Errorf("failed to register models resource: %w", err)
	}

	// stdio://cursor/skills — workflow guide for all MCP clients (Cursor skills + Claude Code commands)
	if err := registerAndTrack(server,
		"stdio://cursor/skills",
		"exarp-go workflow guide",
		"Workflow guide for exarp-go MCP tools: task management, session, reports, health. Usable by Cursor (skills) and Claude Code (CLAUDE.md + commands).",
		"text/markdown",
		handleCursorSkills,
	); err != nil {
		return fmt.Errorf("failed to register cursor skills resource: %w", err)
	}

	// stdio://tools
	if err := registerAndTrack(server,
		"stdio://tools",
		"All Tools",
		"Get all tools in the catalog.",
		"application/json",
		handleAllTools,
	); err != nil {
		return fmt.Errorf("failed to register all tools resource: %w", err)
	}

	if err := registerAndTrack(server,
		"stdio://tool_catalog",
		"Tool Catalog",
		"Get the full tool catalog with descriptions, categories, and usage hints.",
		"application/json",
		handleAllTools,
	); err != nil {
		return fmt.Errorf("failed to register tool catalog resource: %w", err)
	}

	// stdio://tools/names
	if err := registerAndTrack(server,
		"stdio://tools/names",
		"Tool Names",
		"Get all registered tool names as a JSON array.",
		"application/json",
		handleToolNames,
	); err != nil {
		return fmt.Errorf("failed to register tool names resource: %w", err)
	}

	// stdio://resources/uris
	if err := registerAndTrack(server,
		"stdio://resources/uris",
		"Resource URIs",
		"Get all tracked resource URIs as a JSON array.",
		"application/json",
		handleResourceURIs,
	); err != nil {
		return fmt.Errorf("failed to register resource URIs resource: %w", err)
	}

	// stdio://tools/{category}
	if err := registerAndTrack(server,
		"stdio://tools/{category}",
		"Tools by Category",
		"Get tools filtered by category.",
		"application/json",
		handleToolsByCategory,
	); err != nil {
		return fmt.Errorf("failed to register tools by category resource: %w", err)
	}
	if err := registerTemplate(server,
		"stdio://tools/{category}",
		"Tools by Category",
		"Get tools filtered by category.",
		"application/json",
		handleToolsByCategory,
	); err != nil {
		return fmt.Errorf("failed to register tools by category template: %w", err)
	}

	// stdio://tasks
	if err := registerAndTrack(server,
		"stdio://tasks",
		"All Tasks",
		"Get all tasks (read-only, paginated with limit of 50).",
		"application/json",
		handleAllTasks,
	); err != nil {
		return fmt.Errorf("failed to register all tasks resource: %w", err)
	}

	// stdio://tasks/{task_id}
	if err := registerAndTrack(server,
		"stdio://tasks/{task_id}",
		"Task by ID",
		"Get a specific task by ID with full details.",
		"application/json",
		handleTaskByID,
	); err != nil {
		return fmt.Errorf("failed to register task by ID resource: %w", err)
	}
	if err := registerTemplate(server,
		"stdio://tasks/{task_id}",
		"Task by ID",
		"Get a specific task by ID with full details.",
		"application/json",
		handleTaskByID,
	); err != nil {
		return fmt.Errorf("failed to register task by ID template: %w", err)
	}

	// stdio://tasks/status/{status}
	if err := registerAndTrack(server,
		"stdio://tasks/status/{status}",
		"Tasks by Status",
		"Get tasks filtered by status (Todo, In Progress, Done).",
		"application/json",
		handleTasksByStatus,
	); err != nil {
		return fmt.Errorf("failed to register tasks by status resource: %w", err)
	}
	if err := registerTemplate(server,
		"stdio://tasks/status/{status}",
		"Tasks by Status",
		"Get tasks filtered by status (Todo, In Progress, Done).",
		"application/json",
		handleTasksByStatus,
	); err != nil {
		return fmt.Errorf("failed to register tasks by status template: %w", err)
	}

	// stdio://tasks/priority/{priority}
	if err := registerAndTrack(server,
		"stdio://tasks/priority/{priority}",
		"Tasks by Priority",
		"Get tasks filtered by priority (low, medium, high, critical).",
		"application/json",
		handleTasksByPriority,
	); err != nil {
		return fmt.Errorf("failed to register tasks by priority resource: %w", err)
	}
	if err := registerTemplate(server,
		"stdio://tasks/priority/{priority}",
		"Tasks by Priority",
		"Get tasks filtered by priority (low, medium, high, critical).",
		"application/json",
		handleTasksByPriority,
	); err != nil {
		return fmt.Errorf("failed to register tasks by priority template: %w", err)
	}

	// stdio://tasks/tag/{tag}
	if err := registerAndTrack(server,
		"stdio://tasks/tag/{tag}",
		"Tasks by Tag",
		"Get tasks filtered by tag (any tag value).",
		"application/json",
		handleTasksByTag,
	); err != nil {
		return fmt.Errorf("failed to register tasks by tag resource: %w", err)
	}
	if err := registerTemplate(server,
		"stdio://tasks/tag/{tag}",
		"Tasks by Tag",
		"Get tasks filtered by tag (any tag value).",
		"application/json",
		handleTasksByTag,
	); err != nil {
		return fmt.Errorf("failed to register tasks by tag template: %w", err)
	}

	// stdio://tasks/summary
	if err := registerAndTrack(server,
		"stdio://tasks/summary",
		"Tasks Summary",
		"Get task statistics and overview (counts by status, priority, tags).",
		"application/json",
		handleTasksSummary,
	); err != nil {
		return fmt.Errorf("failed to register tasks summary resource: %w", err)
	}

	if err := registerAndTrack(server,
		"stdio://tasks/ready",
		"Ready Tasks",
		"Get dependency-ready tasks that can be started immediately.",
		"application/json",
		handleReadyTasks,
	); err != nil {
		return fmt.Errorf("failed to register ready tasks resource: %w", err)
	}

	if err := registerAndTrack(server,
		"stdio://ready-tasks",
		"Ready Tasks Alias",
		"Alias for dependency-ready tasks that can be started immediately.",
		"application/json",
		handleReadyTasks,
	); err != nil {
		return fmt.Errorf("failed to register ready-tasks alias resource: %w", err)
	}

	// stdio://suggested-tasks — dependency-ready tasks for Cursor/Todo2 hints
	if err := registerAndTrack(server,
		"stdio://suggested-tasks",
		"Suggested Tasks",
		"Get dependency-ordered tasks ready to start (all dependencies Done). For Cursor and MCP clients.",
		"application/json",
		handleSuggestedTasks,
	); err != nil {
		return fmt.Errorf("failed to register suggested tasks resource: %w", err)
	}

	if err := registerAndTrack(server,
		"stdio://active-work",
		"Active Work",
		"Get active task claims and active execution runs for execution-cockpit discovery.",
		"application/json",
		handleActiveWork,
	); err != nil {
		return fmt.Errorf("failed to register active work resource: %w", err)
	}

	if err := registerAndTrack(server,
		"stdio://task-runs/{task_id}",
		"Task Runs",
		"Get recent execution runs for a specific task.",
		"application/json",
		handleTaskRuns,
	); err != nil {
		return fmt.Errorf("failed to register task runs resource: %w", err)
	}
	if err := registerTemplate(server,
		"stdio://task-runs/{task_id}",
		"Task Runs",
		"Get recent execution runs for a specific task.",
		"application/json",
		handleTaskRuns,
	); err != nil {
		return fmt.Errorf("failed to register task runs template: %w", err)
	}

	// agent://card — A2A/ACP agent discovery card
	if err := registerAndTrack(server,
		"agent://card",
		"Agent Card",
		"A2A/ACP agent discovery card: identity, transports, capabilities, and skills grouped by category.",
		"application/json",
		handleAgentCard,
	); err != nil {
		return fmt.Errorf("failed to register agent card resource: %w", err)
	}

	// prime://context — session bootstrap for OpenCode and other MCP clients
	if err := registerAndTrack(server,
		"prime://context",
		"Session Prime Context",
		"Compact session prime data (tasks, hints, mode) for OpenCode and Cursor session bootstrap.",
		"application/json",
		handlePrimeContext,
	); err != nil {
		return fmt.Errorf("failed to register prime context resource: %w", err)
	}

	return nil
}

// Resource handlers
// Note: handleScorecard and memory handlers are implemented in scorecard.go and memories.go
// They are in the same package, so they're automatically available here

// Resource handlers for prompts and session mode
// Note: handleAllPrompts, handlePromptsByMode, handlePromptsByPersona, handlePromptsByCategory
// are implemented in prompts.go (native Go using prompts.GetPromptTemplate)
// handleSessionMode is implemented in session.go (native Go using infer_session_mode)

// handleToolNames handles the stdio://tools/names resource
// Returns all registered tool names as a JSON array.
func handleToolNames(ctx context.Context, uri string) ([]byte, string, error) {
	names := tools.ListToolNames()
	jsonData, err := json.MarshalIndent(names, "", "  ")
	if err != nil {
		return nil, "", err
	}
	return jsonData, "application/json", nil
}

// handleResourceURIs handles the stdio://resources/uris resource
// Returns all tracked resource URIs as a JSON array.
func handleResourceURIs(ctx context.Context, uri string) ([]byte, string, error) {
	uris := tools.ListTrackedResourceURIs()
	jsonData, err := json.MarshalIndent(uris, "", "  ")
	if err != nil {
		return nil, "", err
	}
	return jsonData, "application/json", nil
}
