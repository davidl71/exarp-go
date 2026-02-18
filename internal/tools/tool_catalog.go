package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/davidl71/exarp-go/internal/framework"
	mcpresponse "github.com/davidl71/mcp-go-core/pkg/mcp/response"
)

// ToolCatalogEntry represents a tool in the catalog.
type ToolCatalogEntry struct {
	Tool             string   `json:"tool"`
	Hint             string   `json:"hint"`
	Category         string   `json:"category"`
	Description      string   `json:"description"`
	RecommendedModel string   `json:"recommended_model,omitempty"`
	Examples         []string `json:"examples,omitempty"`
}

// ToolCatalogResponse represents the catalog response.
type ToolCatalogResponse struct {
	Tools               []ToolCatalogEntry     `json:"tools"`
	Count               int                    `json:"count"`
	AvailableCategories []string               `json:"available_categories"`
	AvailablePersonas   []string               `json:"available_personas"`
	FiltersApplied      map[string]interface{} `json:"filters_applied"`
}

// GetToolCatalog returns the static tool catalog
// This is a simplified version that can be enhanced to read from registry dynamically.
func GetToolCatalog() map[string]ToolCatalogEntry {
	return map[string]ToolCatalogEntry{
		// Project Health
		"project_scorecard": {
			Tool:             "project_scorecard",
			Hint:             "TRIGGER: 'how is the project', 'scorecard', 'health check', 'status report', 'production ready'. Provides overall health score and metrics.",
			Category:         "Project Health",
			Description:      "Comprehensive project health assessment with scores across multiple dimensions. Use when user asks about project status, health, or readiness.",
			RecommendedModel: "claude-haiku",
		},
		"project_overview": {
			Tool:             "project_overview",
			Hint:             "TRIGGER: 'overview', 'summary', 'brief', 'what is this project', 'executive summary'. Quick high-level status.",
			Category:         "Project Health",
			Description:      "Executive summary of project status for stakeholders. Use when user wants a quick overview without detailed metrics.",
			RecommendedModel: "claude-haiku",
		},

		// Task Management
		"analyze_alignment": {
			Tool:             "analyze_alignment",
			Hint:             "TRIGGER: 'alignment', 'PRD', 'requirements', 'does this match'. Checks if tasks align with project requirements.",
			Category:         "Task Management",
			Description:      "Analyzes task-project alignment and generates alignment reports. Use when user asks about PRD alignment or requirements compliance.",
			RecommendedModel: "claude-haiku",
		},
		"task_analysis": {
			Tool:             "task_analysis",
			Hint:             "TRIGGER: 'analyze tasks', 'dependencies', 'duplicates', 'conflicts', 'parallel', 'execution plan'. Advanced task analysis.",
			Category:         "Task Management",
			Description:      "Analyzes tasks for duplicates, tags, hierarchy, dependencies, and parallelization opportunities. Use when user wants to understand task relationships or find conflicts.",
			RecommendedModel: "claude-haiku",
		},
		"task_discovery": {
			Tool:             "task_discovery",
			Hint:             "TRIGGER: 'find tasks', 'discover', 'missing tasks', 'scan for todos'. Searches code and docs for tasks.",
			Category:         "Task Management",
			Description:      "Discovers tasks from code comments, markdown files, and other sources. Use when user wants to find tasks that aren't in the task list yet.",
			RecommendedModel: "claude-haiku",
		},
		"task_workflow": {
			Tool:             "task_workflow",
			Hint:             "TRIGGER: 'task workflow', 'create task', 'update task', 'list tasks', 'T-xxx', 'todo', 'triaging'. OpenCode/agent: use for list/update/create when user asks for backlog or task status; use action=sync, sub_action=list with status/filter_tag. PREFER exarp-go task CLI for simple ops; use this tool for clarify, cleanup, sync_approvals.",
			Category:         "Task Management",
			Description:      "Manages task workflow: sync, approve, clarify, cleanup, create. OpenCode/agent: use for listing/updating tasks (action=sync, sub_action=list). PREFER CLI (exarp-go task list/update/create) for simple ops. Never edit .todo2/state.todo2.json directly.",
			RecommendedModel: "claude-haiku",
			Examples: []string{
				"Simple: exarp-go task list, exarp-go task update T-123 --new-status Done",
				"Advanced: task_workflow(action='clarity', task_id='T-123')",
				"Never edit .todo2/state.todo2.json directly",
			},
		},

		// Code Quality
		"lint": {
			Tool:             "lint",
			Hint:             "TRIGGER: 'lint', 'format', 'gofmt', 'style', 'analyze code'. Code linting and formatting.",
			Category:         "Code Quality",
			Description:      "Runs linters and analyzes code quality issues. Use when user asks to lint, format, or check code style.",
			RecommendedModel: "claude-haiku",
		},
		"testing": {
			Tool:             "testing",
			Hint:             "TRIGGER: 'test', 'coverage', 'run tests', 'test failure', 'validate'. Test execution and analysis.",
			Category:         "Code Quality",
			Description:      "Unified testing: run tests, coverage analysis, test suggestions, validation. Use when user mentions tests, coverage, or test failures.",
			RecommendedModel: "claude-haiku",
		},

		// Security
		"security": {
			Tool:             "security",
			Hint:             "TRIGGER: 'security', 'vulnerabilities', 'scan', 'safe', 'security check', 'govulncheck'. Security scanning.",
			Category:         "Security",
			Description:      "Security scanning and vulnerability assessment. Use when user asks about security, vulnerabilities, or wants to check for issues.",
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
			Hint:             "TRIGGER: 'context', 'too much context', 'compact', 'summarize', 'context budget'. Manages conversation context size.",
			Category:         "Workflow",
			Description:      "Manages context: summarize, estimate budget, batch operations. Use when context is getting large or user asks about context limits.",
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
			Hint:             "TRIGGER: 'remember', 'recall', 'memory', 'what did we decide', 'save this', 'look up'. AI memory storage.",
			Category:         "Memory & Learning",
			Description:      "Manages AI memory: save, recall, and search discoveries. Use when user wants to remember something or recall past decisions.",
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
			Hint:             "TRIGGER: 'report', 'briefing', 'plan', 'PRD', 'scorecard', 'overview'. OpenCode/agent: use action=overview for quick status, action=scorecard for health, action=briefing for standup. Creates project reports.",
			Category:         "Reporting",
			Description:      "Generates project reports: overview, scorecard, briefing, PRD, plan (.plan.md). OpenCode/agent: use overview/scorecard/briefing for project status.",
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

		// AI & ML (LLM abstraction: FM, Ollama, MLX)
		"apple_foundation_models": {
			Tool:             "apple_foundation_models",
			Hint:             "LLM abstraction (FM). action=generate|respond|summarize|classify. On-device Apple Silicon. Uses DefaultFMProvider().",
			Category:         "AI & ML",
			Description:      "Apple Foundation Models on-device; part of LLM abstraction (FMProvider)",
			RecommendedModel: "claude-haiku",
		},
		"ollama": {
			Tool:             "ollama",
			Hint:             "LLM abstraction. ollama. action=status|models|generate|pull|hardware|docs|quality|summary. Native then bridge (DefaultOllama()).",
			Category:         "AI & ML",
			Description:      "Ollama local LLM; part of LLM abstraction (OllamaProvider, native then bridge)",
			RecommendedModel: "claude-haiku",
		},
		"text_generate": {
			Tool:             "text_generate",
			Hint:             "Unified generate-text dispatcher. provider=fm|ollama|insight|mlx|localai|auto. Single entry point for all LLM text generation.",
			Category:         "AI & ML",
			Description:      "Unified text generation across all backends (FM, Ollama, MLX, LocalAI, auto model selection)",
			RecommendedModel: "claude-haiku",
		},
		"task_execute": {
			Tool:             "task_execute",
			Hint:             "Run execution flow for a Todo2 task. task_execution template + model + ApplyChanges + result comment.",
			Category:         "Workflow",
			Description:      "Model-assisted task execution: load task, run task_execution prompt, parse response, optionally apply file changes, add result comment (T-215)",
			RecommendedModel: "claude-haiku",
		},
		"mlx": {
			Tool:             "mlx",
			Hint:             "LLM abstraction (MLX). action=status|hardware|models|generate. Bridge-only; report insights use DefaultReportInsight() (MLX then FM).",
			Category:         "AI & ML",
			Description:      "MLX on Apple Silicon; part of LLM abstraction (report insights: MLX then FM)",
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
			Hint:             "TRIGGER: 'health check', 'server status', 'git status', 'docs status'. OpenCode/agent: use action=docs|git|cicd|tools for component status. Checks project health.",
			Category:         "Utilities",
			Description:      "Health checks for server, git, docs, definition of done, CI/CD. OpenCode/agent: use for component status (docs, git, cicd, tools).",
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
			Hint:             "TRIGGER: 'session start', 'handoff', 'context', 'resume', 'what should I do next', 'prime'. OpenCode/agent: call action=prime at session start (include_tasks, include_hints); use handoff to leave/resume notes.",
			Category:         "Utilities",
			Description:      "Session management: prime, handoff, prompts, assignee. OpenCode/agent: prime at start for task/hint context; handoff for leave/resume notes.",
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
			Hint:             "TRIGGER: 'estimate', 'how long', 'duration', 'time estimate', 'effort'. Task duration estimation.",
			Category:         "Utilities",
			Description:      "Estimates task duration using historical data and ML models. Use when user asks how long a task will take.",
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
// This tool now only handles the "help" action.
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

// handleToolCatalogList handles the list action.
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

	m, err := mcpresponse.ConvertToMap(response)
	if err != nil {
		return nil, fmt.Errorf("failed to convert catalog response: %w", err)
	}

	return mcpresponse.FormatResult(m, "")
}

// handleToolCatalogHelp handles the help action.
func handleToolCatalogHelp(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	toolName, _ := params["tool_name"].(string)
	if toolName == "" {
		errorResponse := map[string]interface{}{
			"status": "error",
			"error":  "tool_name parameter required for help action",
		}

		return mcpresponse.FormatResult(errorResponse, "")
	}

	catalog := GetToolCatalog()

	tool, exists := catalog[toolName]
	if !exists {
		errorResponse := map[string]interface{}{
			"status": "error",
			"error":  fmt.Sprintf("Tool '%s' not found in catalog", toolName),
		}

		return mcpresponse.FormatResult(errorResponse, "")
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

	return mcpresponse.FormatResult(helpResponse, "")
}
