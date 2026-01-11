package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/davidl71/exarp-go/internal/framework"
)

// ToolCatalogEntry represents a tool in the catalog
type ToolCatalogEntry struct {
	Tool             string   `json:"tool"`
	Hint             string   `json:"hint"`
	Category         string   `json:"category"`
	Description      string   `json:"description"`
	RecommendedModel string   `json:"recommended_model,omitempty"`
	Examples         []string `json:"examples,omitempty"`
}

// ToolCatalogResponse represents the catalog response
type ToolCatalogResponse struct {
	Tools               []ToolCatalogEntry     `json:"tools"`
	Count               int                    `json:"count"`
	AvailableCategories []string               `json:"available_categories"`
	AvailablePersonas   []string               `json:"available_personas"`
	FiltersApplied      map[string]interface{} `json:"filters_applied"`
}

// GetToolCatalog returns the static tool catalog
// This is a simplified version that can be enhanced to read from registry dynamically
func GetToolCatalog() map[string]ToolCatalogEntry {
	return map[string]ToolCatalogEntry{
		// Project Health
		"project_scorecard": {
			Tool:             "project_scorecard",
			Hint:             "Scorecard. Overall score, component scores, production readiness.",
			Category:         "Project Health",
			Description:      "Comprehensive project health assessment with scores across multiple dimensions",
			RecommendedModel: "claude-haiku",
		},
		"project_overview": {
			Tool:             "project_overview",
			Hint:             "Overview. One-page summary with scores, metrics, tasks, risks.",
			Category:         "Project Health",
			Description:      "Executive summary of project status for stakeholders",
			RecommendedModel: "claude-haiku",
		},

		// Task Management
		"analyze_alignment": {
			Tool:             "analyze_alignment",
			Hint:             "Alignment analysis. action=todo2|prd. Unified alignment analysis tool.",
			Category:         "Task Management",
			Description:      "Analyzes task-project alignment and generates alignment reports",
			RecommendedModel: "claude-haiku",
		},
		"task_analysis": {
			Tool:             "task_analysis",
			Hint:             "Task analysis. action=duplicates|tags|hierarchy|dependencies|parallelization. Task quality and structure.",
			Category:         "Task Management",
			Description:      "Analyzes tasks for duplicates, tags, hierarchy, dependencies, and parallelization opportunities",
			RecommendedModel: "claude-haiku",
		},
		"task_discovery": {
			Tool:             "task_discovery",
			Hint:             "Task discovery. action=comments|markdown|orphans|all. Find tasks from various sources.",
			Category:         "Task Management",
			Description:      "Discovers tasks from code comments, markdown files, and other sources",
			RecommendedModel: "claude-haiku",
		},
		"task_workflow": {
			Tool:             "task_workflow",
			Hint:             "Task workflow. action=sync|approve|clarify|clarity|cleanup. Manage task lifecycle. ⚠️ CRITICAL: ALWAYS use this tool for task updates - NEVER edit .todo2/state.todo2.json directly.",
			Category:         "Task Management",
			Description:      "Manages task workflow: sync, approve, clarify, and cleanup operations. Use action=approve with task_ids for batch status updates. Automatically calculates actualHours, normalizes status, and tracks history.",
			RecommendedModel: "claude-haiku",
			Examples: []string{
				"Batch update tasks to Done: task_workflow(action=\"approve\", status=\"Todo\", new_status=\"Done\", task_ids='[\"T-0\", \"T-1\"]')",
				"Never edit .todo2/state.todo2.json directly - use this tool instead",
			},
		},

		// Code Quality
		"lint": {
			Tool:             "lint",
			Hint:             "Linting tool. action=run|analyze. Run linter or analyze problems.",
			Category:         "Code Quality",
			Description:      "Runs linters and analyzes code quality issues",
			RecommendedModel: "claude-haiku",
		},
		"testing": {
			Tool:             "testing",
			Hint:             "Testing tool. action=run|coverage|suggest|validate. Execute tests, analyze coverage, suggest test cases, or validate test structure.",
			Category:         "Code Quality",
			Description:      "Unified testing: run tests, coverage analysis, test suggestions, validation",
			RecommendedModel: "claude-haiku",
		},

		// Security
		"security": {
			Tool:             "security",
			Hint:             "Security. action=scan|alerts|report. Vulnerabilities, remediation.",
			Category:         "Security",
			Description:      "Security scanning and vulnerability assessment",
			RecommendedModel: "claude-haiku",
		},
		"check_attribution": {
			Tool:             "check_attribution",
			Hint:             "Attribution compliance check. Verify proper attribution for all third-party components.",
			Category:         "Security",
			Description:      "Checks for proper attribution of third-party code and licenses",
			RecommendedModel: "claude-haiku",
		},

		// Workflow
		"context": {
			Tool:             "context",
			Hint:             "Context management. action=summarize|budget|batch. Unified context operations.",
			Category:         "Workflow",
			Description:      "Manages context: summarize, estimate budget, batch operations",
			RecommendedModel: "claude-haiku",
		},
		"tool_catalog": {
			Tool:             "tool_catalog",
			Hint:             "Tool catalog. action=list|help. Unified tool catalog and help.",
			Category:         "Workflow",
			Description:      "Browses tool catalog and provides help for available tools",
			RecommendedModel: "claude-haiku",
		},
		"workflow_mode": {
			Tool:             "workflow_mode",
			Hint:             "Workflow mode management. action=focus|suggest|stats. Unified workflow operations.",
			Category:         "Workflow",
			Description:      "Manages workflow modes and operational states",
			RecommendedModel: "claude-haiku",
		},
		"server_status": {
			Tool:             "server_status",
			Hint:             "Server status. Get the current status of the project management automation server.",
			Category:         "Workflow",
			Description:      "Returns server operational status and version information",
			RecommendedModel: "claude-haiku",
		},

		// Configuration
		"generate_config": {
			Tool:             "generate_config",
			Hint:             "Config generation. action=rules|ignore|simplify. Creates IDE config files.",
			Category:         "Configuration",
			Description:      "Generates IDE configuration files (.cursorrules, .cursorignore)",
			RecommendedModel: "claude-haiku",
		},
		"setup_hooks": {
			Tool:             "setup_hooks",
			Hint:             "Hooks setup. action=git|patterns. Install automation hooks.",
			Category:         "Configuration",
			Description:      "Sets up Git hooks and automation triggers",
			RecommendedModel: "claude-haiku",
		},

		// Memory & Learning
		"memory": {
			Tool:             "memory",
			Hint:             "Memory tool. action=save|recall|search. Persist and retrieve AI discoveries.",
			Category:         "Memory & Learning",
			Description:      "Manages AI memory: save, recall, and search discoveries",
			RecommendedModel: "claude-haiku",
		},
		"memory_maint": {
			Tool:             "memory_maint",
			Hint:             "Memory maintenance. action=health|gc|prune|consolidate|dream. Lifecycle management.",
			Category:         "Memory & Learning",
			Description:      "Maintains memory system: health checks, garbage collection, consolidation",
			RecommendedModel: "claude-haiku",
		},

		// Reporting
		"report": {
			Tool:             "report",
			Hint:             "Report generation. action=overview|scorecard|briefing|prd. Project reports.",
			Category:         "Reporting",
			Description:      "Generates project reports: overview, scorecard, briefing, PRD",
			RecommendedModel: "claude-haiku",
		},

		// Automation
		"automation": {
			Tool:             "automation",
			Hint:             "Automation. action=daily|nightly|sprint|discover. Unified automation tool.",
			Category:         "Automation",
			Description:      "Unified automation: daily, nightly, sprint, and discovery workflows",
			RecommendedModel: "claude-haiku",
		},

		// AI & ML
		"ollama": {
			Tool:             "ollama",
			Hint:             "Ollama. action=status|models|generate|pull|hardware|docs|quality|summary. Unified Ollama tool.",
			Category:         "AI & ML",
			Description:      "Ollama integration for local LLM operations",
			RecommendedModel: "claude-haiku",
		},
		"mlx": {
			Tool:             "mlx",
			Hint:             "MLX. action=status|hardware|models|generate. Unified MLX tool.",
			Category:         "AI & ML",
			Description:      "MLX integration for Apple Silicon ML operations",
			RecommendedModel: "claude-haiku",
		},
		"recommend": {
			Tool:             "recommend",
			Hint:             "Recommend. action=model|workflow|advisor. Unified recommendation tool.",
			Category:         "AI & ML",
			Description:      "Recommends models, workflows, and advisors based on task context",
			RecommendedModel: "claude-haiku",
		},

		// Utilities
		"health": {
			Tool:             "health",
			Hint:             "Health check. action=server|git|docs|dod|cicd. Status and health metrics.",
			Category:         "Utilities",
			Description:      "Health checks for server, git, docs, definition of done, CI/CD",
			RecommendedModel: "claude-haiku",
		},
		"git_tools": {
			Tool:             "git_tools",
			Hint:             "Git tools. action=commits|branches|tasks|diff|graph|merge|set_branch. Unified git-inspired tools.",
			Category:         "Utilities",
			Description:      "Git-inspired task management and version control operations",
			RecommendedModel: "claude-haiku",
		},
		"session": {
			Tool:             "session",
			Hint:             "Session. action=prime|handoff|prompts|assignee. Unified session management tools.",
			Category:         "Utilities",
			Description:      "Session management: prime, handoff, prompts, assignee operations",
			RecommendedModel: "claude-haiku",
		},
		"infer_session_mode": {
			Tool:             "infer_session_mode",
			Hint:             "Session mode inference. Returns AGENT/ASK/MANUAL with confidence.",
			Category:         "Utilities",
			Description:      "Infers session mode (AGENT/ASK/MANUAL) based on task patterns",
			RecommendedModel: "claude-haiku",
		},
		"estimation": {
			Tool:             "estimation",
			Hint:             "Estimation. action=estimate|analyze|stats. Unified task duration estimation tool.",
			Category:         "Utilities",
			Description:      "Estimates task duration using historical data and ML models",
			RecommendedModel: "claude-haiku",
		},
		"prompt_tracking": {
			Tool:             "prompt_tracking",
			Hint:             "Prompt tracking. action=log|analyze. Track and analyze prompts.",
			Category:         "Utilities",
			Description:      "Tracks and analyzes prompt usage patterns",
			RecommendedModel: "claude-haiku",
		},
		"add_external_tool_hints": {
			Tool:             "add_external_tool_hints",
			Hint:             "Tool hints. Files scanned, modified, hints added.",
			Category:         "Utilities",
			Description:      "Adds external tool hints to documentation",
			RecommendedModel: "claude-haiku",
		},
	}
}

// handleToolCatalogNative handles the tool_catalog tool with native Go implementation
// Note: "list" action converted to stdio://tools and stdio://tools/{category} resources
// This tool now only handles the "help" action
func handleToolCatalogNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	action, _ := params["action"].(string)
	if action == "" {
		action = "help"
	}

	switch action {
	case "help":
		return handleToolCatalogHelp(ctx, params)
	default:
		return nil, fmt.Errorf("unknown tool_catalog action: %s. Use 'help' (list action converted to stdio://tools resources)", action)
	}
}

// handleToolCatalogList handles the list action
func handleToolCatalogList(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	catalog := GetToolCatalog()

	// Get filters
	category, _ := params["category"].(string)
	persona, _ := params["persona"].(string)
	includeExamples := true
	if inc, ok := params["include_examples"].(bool); ok {
		includeExamples = inc
	}

	// Filter tools
	var filtered []ToolCatalogEntry
	categories := make(map[string]bool)

	for toolID, tool := range catalog {
		// Apply category filter (case-insensitive)
		if category != "" && !strings.EqualFold(tool.Category, category) {
			continue
		}

		// Persona filter would need to be added to catalog entries if needed
		// For now, we'll skip persona filtering as it's not in our catalog structure

		// Create entry
		entry := tool
		entry.Tool = toolID

		// Add examples if requested
		if !includeExamples {
			entry.Examples = nil
		}

		filtered = append(filtered, entry)
		categories[tool.Category] = true
	}

	// Build response
	catList := make([]string, 0, len(categories))
	for cat := range categories {
		catList = append(catList, cat)
	}

	// Sort categories (simple alphabetical)
	for i := 0; i < len(catList)-1; i++ {
		for j := i + 1; j < len(catList); j++ {
			if catList[i] > catList[j] {
				catList[i], catList[j] = catList[j], catList[i]
			}
		}
	}

	response := ToolCatalogResponse{
		Tools:               filtered,
		Count:               len(filtered),
		AvailableCategories: catList,
		AvailablePersonas:   []string{}, // Not currently supported
		FiltersApplied: map[string]interface{}{
			"category": category,
			"persona":  persona,
		},
	}

	result, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal catalog response: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: string(result)},
	}, nil
}

// handleToolCatalogHelp handles the help action
func handleToolCatalogHelp(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	toolName, _ := params["tool_name"].(string)
	if toolName == "" {
		errorResponse := map[string]interface{}{
			"status": "error",
			"error":  "tool_name parameter required for help action",
		}
		result, _ := json.MarshalIndent(errorResponse, "", "  ")
		return []framework.TextContent{
			{Type: "text", Text: string(result)},
		}, nil
	}

	catalog := GetToolCatalog()
	tool, exists := catalog[toolName]
	if !exists {
		errorResponse := map[string]interface{}{
			"status": "error",
			"error":  fmt.Sprintf("Tool '%s' not found in catalog", toolName),
		}
		result, _ := json.MarshalIndent(errorResponse, "", "  ")
		return []framework.TextContent{
			{Type: "text", Text: string(result)},
		}, nil
	}

	// Build help response
	helpResponse := map[string]interface{}{
		"tool":              tool.Tool,
		"hint":              tool.Hint,
		"category":          tool.Category,
		"description":       tool.Description,
		"recommended_model": tool.RecommendedModel,
	}
	if len(tool.Examples) > 0 {
		helpResponse["examples"] = tool.Examples
	}

	result, err := json.MarshalIndent(helpResponse, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal help response: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: string(result)},
	}, nil
}
