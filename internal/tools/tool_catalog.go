// tool_catalog.go — MCP "tool_catalog" tool: help and discovery for available tools.
package tools

import (
	"context"
	"fmt"
	"sort"

	"github.com/davidl71/exarp-go/internal/framework"
)

// ToolCatalogEntry represents a tool in the catalog.
type ToolCatalogEntry struct {
	Tool             string   `json:"tool"`
	Hint             string   `json:"hint"`
	Category         string   `json:"category"`
	Description      string   `json:"description"`
	Aliases          []string `json:"aliases,omitempty"`
	Class            string   `json:"class,omitempty"`
	PreferredTool    string   `json:"preferred_tool,omitempty"`
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
			Class:            "specialist",
			RecommendedModel: "claude-haiku",
		},
		"task_analysis": {
			Tool:             "task_analysis",
			Hint:             "TRIGGER: 'analyze tasks', 'dependencies', 'duplicates', 'conflicts', 'parallel', 'execution plan'. Advanced task analysis.",
			Category:         "Task Management",
			Description:      "Analyzes tasks for duplicates, tags, hierarchy, dependencies, and parallelization opportunities. Use when user wants to understand task relationships or find conflicts.",
			Class:            "primary",
			RecommendedModel: "claude-haiku",
		},
		"task_discovery": {
			Tool:             "task_discovery",
			Hint:             "TRIGGER: 'find tasks', 'discover', 'missing tasks', 'scan for todos'. Searches code and docs for tasks.",
			Category:         "Task Management",
			Description:      "Discovers tasks from code comments, markdown files, and other sources. Use when user wants to find tasks that aren't in the task list yet.",
			Class:            "primary",
			RecommendedModel: "claude-haiku",
		},
		"task_workflow": {
			Tool:             "task_workflow",
			Hint:             "TRIGGER: 'task workflow', 'create task', 'update task', 'list tasks', 'T-xxx', 'todo', 'triaging', 'start run', 'verify'. Use for lifecycle plus execution-cockpit actions. Prefer exarp-go task CLI aliases for simple list/show/update/create flows.",
			Category:         "Task Management",
			Description:      "Manages task workflow and execution state: list/create/update, claim, start_run/end_run, verify, add_progress, split, and approvals. Prefer CLI aliases for simple list/show/update/create flows. Never edit .todo2/state.todo2.json directly.",
			Aliases:          []string{"task list", "task show", "task update", "task create", "stdio://agent/task/{task_id}/execution-pack"},
			Class:            "primary",
			RecommendedModel: "claude-haiku",
			Examples: []string{
				"Simple: exarp-go task list, exarp-go task update T-123 --new-status Done",
				"Execution: task_workflow(action='start_run', task_id='T-123')",
				"Advanced: task_workflow(action='clarity', task_id='T-123')",
				"Never edit .todo2/state.todo2.json directly",
			},
		},
		"task_runs": {
			Tool:             "task_runs",
			Hint:             "Alias for execution-run operations. Use when the intent is start_run, end_run, list_runs, or show_run.",
			Category:         "Task Management",
			Description:      "Alias entry for execution-run workflows under task_workflow.",
			Class:            "alias",
			PreferredTool:    "task_workflow",
			RecommendedModel: "claude-haiku",
			Examples: []string{
				"task_workflow(action='list_runs', task_id='T-123')",
				"task_workflow(action='start_run', task_id='T-123', summary='Implement feature')",
			},
		},
		"task_verify": {
			Tool:             "task_verify",
			Hint:             "Alias for recording verification evidence on a task or execution run.",
			Category:         "Task Management",
			Description:      "Alias entry for task_workflow(action='verify').",
			Class:            "alias",
			PreferredTool:    "task_workflow",
			RecommendedModel: "claude-haiku",
		},
		"task_progress": {
			Tool:             "task_progress",
			Hint:             "Alias for recording partial progress slices on a task or execution run.",
			Category:         "Task Management",
			Description:      "Alias entry for task_workflow(action='add_progress').",
			Class:            "alias",
			PreferredTool:    "task_workflow",
			RecommendedModel: "claude-haiku",
		},
		"task_claim": {
			Tool:             "task_claim",
			Hint:             "Alias for active-work coordination: claim, release, and agent_status.",
			Category:         "Task Management",
			Description:      "Alias entry for task claiming and active-work coordination under task_workflow.",
			Class:            "alias",
			PreferredTool:    "task_workflow",
			RecommendedModel: "claude-haiku",
		},
		"ready_tasks": {
			Tool:             "ready_tasks",
			Hint:             "Alias for dependency-ready task discovery; use suggested-tasks or tasks/ready resources.",
			Category:         "Task Management",
			Description:      "Alias entry for dependency-ready work selection.",
			Class:            "alias",
			PreferredTool:    "task_workflow",
			RecommendedModel: "claude-haiku",
			Aliases:          []string{"stdio://tasks/ready", "stdio://ready-tasks"},
		},

		// Code Quality
		"lint": {
			Tool:             "lint",
			Hint:             "TRIGGER: 'lint', 'format', 'gofmt', 'style', 'analyze code'. Code linting and formatting.",
			Category:         "Code Quality",
			Description:      "Runs linters and analyzes code quality issues. Use when user asks to lint, format, or check code style.",
			Class:            "primary",
			RecommendedModel: "claude-haiku",
		},
		"testing": {
			Tool:             "testing",
			Hint:             "TRIGGER: 'test', 'coverage', 'run tests', 'test failure', 'validate'. Testing workflows; run|coverage|validate are Go-project flows today.",
			Category:         "Code Quality",
			Description:      "Runs Go test/coverage/validation flows and offers test suggestions. Use when user mentions tests, coverage, or test failures; non-Go repos should treat run/coverage/validate as Go-specific until framework-aware runners are added.",
			Class:            "primary",
			RecommendedModel: "claude-haiku",
		},

		// Security
		"security": {
			Tool:             "security",
			Hint:             "TRIGGER: 'security', 'vulnerabilities', 'scan', 'safe', 'security check', 'govulncheck'. Security scanning.",
			Category:         "Security",
			Description:      "Security scanning and vulnerability assessment. Use when user asks about security, vulnerabilities, or wants to check for issues.",
			Class:            "primary",
			RecommendedModel: "claude-haiku",
		},
		"check_attribution": {
			Tool:             "check_attribution",
			Hint:             "Attribution compliance check. Verify proper attribution for all third-party components.",
			Category:         "Security",
			Description:      "Checks for proper attribution of third-party code and licenses",
			Class:            "specialist",
			RecommendedModel: "claude-haiku",
		},

		// Workflow
		"context": {
			Tool:             "context",
			Hint:             "TRIGGER: 'context', 'too much context', 'compact', 'summarize', 'context budget'. Manages conversation context size.",
			Category:         "Workflow",
			Description:      "Manages context: summarize, estimate budget, batch operations. Use when context is getting large or user asks about context limits.",
			Class:            "specialist",
			RecommendedModel: "claude-haiku",
		},
		"tool_catalog": {
			Tool:             "tool_catalog",
			Hint:             "Tool catalog. action=help (requires tool_name). For listing tools use stdio://tools or stdio://tools/names resources.",
			Category:         "Workflow",
			Description:      "Provides per-tool help from the static tool catalog. Tool listing is exposed via stdio://tools resources.",
			Aliases:          []string{"stdio://tools", "stdio://tool_catalog"},
			Class:            "primary",
			RecommendedModel: "claude-haiku",
		},
		"workflow_mode": {
			Tool:             "workflow_mode",
			Hint:             "Workflow mode management. action=focus|suggest|stats. Unified workflow operations.",
			Category:         "Workflow",
			Description:      "Manages workflow modes and operational states",
			Class:            "primary",
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
			Class:            "primary",
			RecommendedModel: "claude-haiku",
		},
		"setup_hooks": {
			Tool:             "setup_hooks",
			Hint:             "Hooks setup. action=git|patterns. Install automation hooks.",
			Category:         "Configuration",
			Description:      "Sets up Git hooks and automation triggers",
			Class:            "primary",
			RecommendedModel: "claude-haiku",
		},

		// Memory & Learning
		"memory": {
			Tool:             "memory",
			Hint:             "TRIGGER: 'remember', 'recall', 'memory', 'what did we decide', 'save this', 'look up'. AI memory storage.",
			Category:         "Memory & Learning",
			Description:      "Manages AI memory: save, recall, and search discoveries. Use when user wants to remember something or recall past decisions.",
			Class:            "primary",
			RecommendedModel: "claude-haiku",
		},
		"memory_maint": {
			Tool:             "memory_maint",
			Hint:             "Memory maintenance. action=health|gc|prune|consolidate|dream. Lifecycle management.",
			Category:         "Memory & Learning",
			Description:      "Maintains memory system: health checks, garbage collection, consolidation",
			Class:            "primary",
			RecommendedModel: "claude-haiku",
		},

		// Reporting
		"report": {
			Tool:             "report",
			Hint:             "TRIGGER: 'report', 'briefing', 'plan', 'PRD', 'scorecard', 'overview', 'execution briefing'. Use action=execution_briefing for active work and execution-cockpit status.",
			Category:         "Reporting",
			Description:      "Generates project reports: overview, scorecard, briefing, execution_briefing, PRD, and plan (.plan.md). Use execution_briefing for active work and execution-cockpit status.",
			Aliases:          []string{"execution_briefing"},
			Class:            "primary",
			RecommendedModel: "claude-haiku",
		},
		"execution_briefing": {
			Tool:             "execution_briefing",
			Hint:             "Alias for report(action='execution_briefing') and active-work status discovery.",
			Category:         "Reporting",
			Description:      "Alias entry for execution-focused status reporting.",
			Class:            "alias",
			PreferredTool:    "report",
			RecommendedModel: "claude-haiku",
			Aliases:          []string{"stdio://active-work"},
		},
		"active_work": {
			Tool:             "active_work",
			Hint:             "Alias for active execution visibility; use report(action='execution_briefing') or stdio://active-work.",
			Category:         "Reporting",
			Description:      "Alias entry for active claims and active execution runs.",
			Class:            "alias",
			PreferredTool:    "report",
			RecommendedModel: "claude-haiku",
			Aliases:          []string{"execution_briefing", "stdio://active-work"},
		},
		// Automation
		"automation": {
			Tool:             "automation",
			Hint:             "Automation. action=daily|nightly|sprint|discover|schedule|unschedule. Unified automation tool.",
			Category:         "Automation",
			Description:      "Unified automation: daily, nightly, sprint, discovery, and OS-native schedule workflows",
			Class:            "primary",
			RecommendedModel: "claude-haiku",
		},

		// AI & ML (LLM abstraction: FM, Ollama, MLX)
		"fm_plan_and_execute": {
			Tool:             "fm_plan_and_execute",
			Hint:             "Plan-and-execute with FM/Ollama. Breaks task into subtasks (planner), runs workers in parallel, combines. Use for complex single-shot tasks.",
			Category:         "AI & ML",
			Description:      "Plan-and-execute flow: planner breaks task into subtasks, workers run in parallel, results combined (uses DefaultFMProvider)",
			Class:            "specialist",
			RecommendedModel: "claude-haiku",
		},
		"ollama": {
			Tool:             "ollama",
			Hint:             "LLM abstraction. ollama. action=status|models|generate|pull|hardware|docs|quality|summary. Native then bridge (DefaultOllama()).",
			Category:         "AI & ML",
			Description:      "Ollama local LLM; part of LLM abstraction (OllamaProvider, native then bridge)",
			Class:            "specialist",
			RecommendedModel: "claude-haiku",
		},
		"text_generate": {
			Tool:             "text_generate",
			Hint:             "Unified generate-text dispatcher. provider=fm|ollama|insight|mlx|localai|gateway|auto. Single entry point for all LLM text generation.",
			Category:         "AI & ML",
			Description:      "Unified text generation across all backends (FM, Ollama, MLX, LocalAI, gateway, auto model selection)",
			Class:            "primary",
			RecommendedModel: "claude-haiku",
		},
		"task_execute": {
			Tool:             "task_execute",
			Hint:             "Run execution flow for a Todo2 task. task_execution template + model + ApplyChanges + result comment.",
			Category:         "Workflow",
			Description:      "Model-assisted task execution: load task, run task_execution prompt, parse response, optionally apply file changes, add result comment (T-215)",
			Class:            "alias",
			PreferredTool:    "task_workflow",
			RecommendedModel: "claude-haiku",
		},
		"mlx": {
			Tool:             "mlx",
			Hint:             "LLM abstraction (MLX). action=status|hardware|models|generate. Bridge-only; report insights use DefaultReportInsight() (MLX then FM).",
			Category:         "AI & ML",
			Description:      "MLX on Apple Silicon; part of LLM abstraction (report insights: MLX then FM)",
			Class:            "specialist",
			RecommendedModel: "claude-haiku",
		},
		"recommend": {
			Tool:             "recommend",
			Hint:             "Recommend. action=model|workflow|advisor. Unified recommendation tool.",
			Category:         "AI & ML",
			Description:      "Recommends models, workflows, and advisors based on task context",
			Class:            "primary",
			RecommendedModel: "claude-haiku",
		},

		// Utilities
		"health": {
			Tool:             "health",
			Hint:             "TRIGGER: 'health check', 'server status', 'git status', 'docs status'. OpenCode/agent: use action=docs|git|cicd|tools for component status. Checks project health.",
			Category:         "Utilities",
			Description:      "Health checks for server, git, docs, definition of done, CI/CD. OpenCode/agent: use for component status (docs, git, cicd, tools).",
			Class:            "primary",
			RecommendedModel: "claude-haiku",
		},
		"git_tools": {
			Tool:             "git_tools",
			Hint:             "Git tools. action=commits|branches|tasks|diff|graph|merge|set_branch. Unified git-inspired tools.",
			Category:         "Utilities",
			Description:      "Git-inspired task management and version control operations",
			Class:            "primary",
			RecommendedModel: "claude-haiku",
		},
		"session": {
			Tool:             "session",
			Hint:             "TRIGGER: 'session start', 'handoff', 'context', 'resume', 'what should I do next', 'prime'. OpenCode/agent: call action=prime at session start (include_tasks, include_hints); use handoff to leave/resume notes.",
			Category:         "Utilities",
			Description:      "Session management: prime, handoff, prompts, assignee. OpenCode/agent: prime at start for task/hint context; handoff for leave/resume notes.",
			Class:            "primary",
			RecommendedModel: "claude-haiku",
			Aliases:          []string{"stdio://agent/briefing", "stdio://agent/alerts"},
		},
		"infer_session_mode": {
			Tool:             "infer_session_mode",
			Hint:             "Session mode inference. Returns AGENT/ASK/MANUAL with confidence.",
			Category:         "Utilities",
			Description:      "Infers session mode (AGENT/ASK/MANUAL) based on task patterns",
			Class:            "alias",
			PreferredTool:    "session",
			RecommendedModel: "claude-haiku",
		},
		"estimation": {
			Tool:             "estimation",
			Hint:             "TRIGGER: 'estimate', 'how long', 'duration', 'time estimate', 'effort'. Task duration estimation.",
			Category:         "Utilities",
			Description:      "Estimates task duration using historical data and ML models. Use when user asks how long a task will take.",
			Class:            "specialist",
			RecommendedModel: "claude-haiku",
		},
		"prompt_tracking": {
			Tool:             "prompt_tracking",
			Hint:             "Prompt tracking. action=log|analyze. Track and analyze prompts.",
			Category:         "Utilities",
			Description:      "Tracks and analyzes prompt usage patterns",
			Class:            "specialist",
			RecommendedModel: "claude-haiku",
		},
		"add_external_tool_hints": {
			Tool:             "add_external_tool_hints",
			Hint:             "Tool hints. Files scanned, modified, hints added.",
			Category:         "Utilities",
			Description:      "Adds external tool hints to documentation",
			Class:            "specialist",
			RecommendedModel: "claude-haiku",
		},
	}
}

// ListToolNames returns the alphabetically sorted tool IDs from the catalog.
func ListToolNames() []string {
	catalog := GetToolCatalog()
	names := make([]string, 0, len(catalog))
	for name := range catalog {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// handleToolCatalogNative handles the tool_catalog tool with native Go implementation
// Note: "list" action converted to stdio://tools and stdio://tools/{category} resources
// This tool now only handles the "help" action.
func handleToolCatalogNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	action := ParamString(params, "action")
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

// handleToolCatalogHelp handles the help action.
func handleToolCatalogHelp(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	toolName := ParamString(params, "tool_name")
	if toolName == "" {
		errorResponse := map[string]interface{}{
			"status": "error",
			"error":  "tool_name parameter required for help action",
		}

		return framework.FormatResult(errorResponse, "")
	}

	catalog := GetToolCatalog()

	tool, exists := catalog[toolName]
	if !exists {
		errorResponse := map[string]interface{}{
			"status": "error",
			"error":  fmt.Sprintf("Tool '%s' not found in catalog", toolName),
		}

		return framework.FormatResult(errorResponse, "")
	}

	// Build help response
	helpResponse := map[string]interface{}{
		"tool":              tool.Tool,
		"hint":              tool.Hint,
		"category":          tool.Category,
		"description":       tool.Description,
		"class":             tool.Class,
		"preferred_tool":    tool.PreferredTool,
		"recommended_model": tool.RecommendedModel,
	}
	if len(tool.Examples) > 0 {
		helpResponse["examples"] = tool.Examples
	}

	return framework.FormatResult(helpResponse, "")
}
