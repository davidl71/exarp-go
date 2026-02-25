// registry_batch5.go — MCP tool registration batch.
package tools

import (
	"fmt"

	"github.com/davidl71/exarp-go/internal/framework"
)

// Note: demonstrate_elicit and interactive_task_create were removed (required FastMCP Context).
func registerBatch5Tools(server framework.MCPServer) error {
	// context - Unified context management (summarize/budget/batch actions)
	if err := server.RegisterTool(
		"context",
		"[HINT: action=summarize|budget|batch. Context management and summarization. Use when reducing context size or summarizing tool outputs.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"summarize", "budget", "batch"},
					"default":     "summarize",
					"description": "Action to perform",
				},
				"data": map[string]interface{}{
					"type":        "string",
					"description": "JSON string to summarize (summarize action)",
				},
				"level": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"brief", "detailed", "key_metrics", "actionable"},
					"default":     "brief",
					"description": "Summarization level (summarize action)",
				},
				"tool_type": map[string]interface{}{
					"type":        "string",
					"description": "Tool type hint for smarter summarization (summarize action)",
				},
				"max_tokens": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum tokens for output (summarize action)",
				},
				"include_raw": map[string]interface{}{
					"type":        "boolean",
					"default":     false,
					"description": "Include original data in response (summarize action)",
				},
				"items": map[string]interface{}{
					"type":        "string",
					"description": "JSON array of items to analyze (budget/batch actions)",
				},
				"budget_tokens": map[string]interface{}{
					"type":        "integer",
					"default":     4000,
					"description": "Target token budget (budget action)",
				},
				"combine": map[string]interface{}{
					"type":        "boolean",
					"default":     true,
					"description": "Merge summaries into combined view (batch action)",
				},
			},
		},
		handleContext,
	); err != nil {
		return fmt.Errorf("failed to register context: %w", err)
	}

	// text_generate - Unified generate-text dispatcher for all LLM backends (FM, Ollama, MLX, LocalAI, ReportInsight)
	// When task_type or task_description is provided, uses ResolveModelForTask (recommend + router) for model selection (T-207).
	if err := server.RegisterTool(
		"text_generate",
		"[HINT: provider=fm|ollama|insight|mlx|localai|auto. Unified text generation across all LLM backends. Single entry point for generate-text. task_type enables auto model selection.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"provider": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"fm", "ollama", "insight", "mlx", "localai", "llamacpp", "auto"},
					"default":     "fm",
					"description": "Backend: fm (Apple/chain), ollama (native Ollama), insight (report), mlx (Apple Silicon), localai (OpenAI-compatible), llamacpp (GGUF via llama.cpp), or auto (model selection from task_type/task_description)",
				},
				"prompt": map[string]interface{}{
					"type":        "string",
					"description": "Prompt for text generation (required)",
				},
				"task_type": map[string]interface{}{
					"type":        "string",
					"description": "Task type hint for model selection (e.g. code_analysis, quick_fix). Used with provider=auto.",
				},
				"task_description": map[string]interface{}{
					"type":        "string",
					"description": "Task description for model selection. Used with provider=auto.",
				},
				"optimize_for": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"quality", "speed", "cost"},
					"default":     "quality",
					"description": "Optimization target for model selection (when provider=auto)",
				},
				"max_tokens": map[string]interface{}{
					"type":        "integer",
					"default":     512,
					"description": "Maximum tokens to generate",
				},
				"temperature": map[string]interface{}{
					"type":        "number",
					"default":     0.7,
					"description": "Sampling temperature",
				},
			},
			Required: []string{"prompt"},
		},
		handleTextGenerate,
	); err != nil {
		return fmt.Errorf("failed to register text_generate: %w", err)
	}

	// task_execute - Run model-assisted execution flow for a Todo2 task (T-215; MODEL_ASSISTED_WORKFLOW Phase 4)
	if err := server.RegisterTool(
		"task_execute",
		"[HINT: Model-assisted task execution. Loads Todo2 task, generates plan via LLM, optionally applies file changes. Use when auto-executing tasks.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"task_id": map[string]interface{}{
					"type":        "string",
					"description": "Todo2 task ID to execute (required)",
				},
				"project_root": map[string]interface{}{
					"type":        "string",
					"description": "Project root for task store and file changes (default: detected)",
				},
				"apply": map[string]interface{}{
					"type":        "boolean",
					"default":     true,
					"description": "If true, apply parsed file changes when confidence >= min_confidence",
				},
				"min_confidence": map[string]interface{}{
					"type":        "number",
					"default":     0.5,
					"description": "Minimum confidence (0–1) to apply changes",
				},
			},
			Required: []string{"task_id"},
		},
		handleTaskExecute,
	); err != nil {
		return fmt.Errorf("failed to register task_execute: %w", err)
	}

	// prompt_tracking - Unified prompt tracking (log/analyze actions)
	if err := server.RegisterTool(
		"prompt_tracking",
		"[HINT: action=log|analyze. Track and analyze prompt usage patterns. Use when optimizing prompt strategies over time.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"log", "analyze"},
					"default":     "analyze",
					"description": "Action to perform",
				},
				"prompt": map[string]interface{}{
					"type":        "string",
					"description": "Prompt text to log (log action)",
				},
				"task_id": map[string]interface{}{
					"type":        "string",
					"description": "Optional task ID",
				},
				"mode": map[string]interface{}{
					"type":        "string",
					"description": "Optional mode",
				},
				"outcome": map[string]interface{}{
					"type":        "string",
					"description": "Optional outcome",
				},
				"iteration": map[string]interface{}{
					"type":        "integer",
					"default":     1,
					"description": "Iteration number (log action)",
				},
				"days": map[string]interface{}{
					"type":        "integer",
					"default":     7,
					"description": "Number of days to analyze (analyze action)",
				},
			},
		},
		handlePromptTracking,
	); err != nil {
		return fmt.Errorf("failed to register prompt_tracking: %w", err)
	}

	// recommend - Unified recommendation tool (model/workflow/advisor actions)
	if err := server.RegisterTool(
		"recommend",
		"[HINT: action=model|workflow|advisor. Get recommendations for models, workflows, or advisors. Use when choosing the right approach for a task.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"model", "workflow", "advisor"},
					"default":     "model",
					"description": "Action to perform",
				},
				"task_description": map[string]interface{}{
					"type":        "string",
					"description": "Description of the task",
				},
				"tags": map[string]interface{}{
					"type":        "string",
					"description": "Optional JSON string of tags to consider",
				},
				"include_rationale": map[string]interface{}{
					"type":        "boolean",
					"default":     true,
					"description": "Whether to include detailed reasoning",
				},
				"task_type": map[string]interface{}{
					"type":        "string",
					"description": "Optional explicit task type (model action)",
				},
				"optimize_for": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"quality", "speed", "cost"},
					"default":     "quality",
					"description": "Optimization target (model action)",
				},
				"include_alternatives": map[string]interface{}{
					"type":        "boolean",
					"default":     true,
					"description": "Include alternative recommendations (model action)",
				},
			},
		},
		handleRecommend,
	); err != nil {
		return fmt.Errorf("failed to register recommend: %w", err)
	}

	// T-224: research_aggregator - runs multiple research tools and combines outputs
	if err := server.RegisterTool(
		"research_aggregator",
		"[HINT: Run multiple analysis tools and combine outputs. Use when you need a comprehensive research overview. Related: task_analysis, analyze_alignment.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"tools": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"oneOf": []interface{}{
							map[string]interface{}{"type": "string"},
							map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"tool":   map[string]interface{}{"type": "string"},
									"action": map[string]interface{}{"type": "string"},
								},
								"required": []string{"tool"},
							},
						},
					},
					"description": "Tool configs: [{tool, action}] or tool names. Default: duplicates, dependencies, todo2.",
				},
			},
		},
		handleResearchAggregator,
	); err != nil {
		return fmt.Errorf("failed to register research_aggregator: %w", err)
	}

	// Note: server_status tool removed - converted to stdio://server/status resource
	// See internal/resources/server.go for resource implementation

	// Note: demonstrate_elicit and interactive_task_create removed
	// These tools require FastMCP Context (not available in stdio mode)
	// They were demonstration tools that don't work in exarp-go's primary stdio mode

	// cursor_cloud_agent — Cursor Cloud Agents API (Beta). T-1771164550717
	if err := server.RegisterTool(
		"cursor_cloud_agent",
		"[HINT: action=launch|status|list|follow_up|delete. Cursor Cloud Agents API (Beta). Requires CURSOR_API_KEY. Use when running remote Cursor agents.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"launch", "status", "list", "follow_up", "delete"},
					"default": "list",
				},
				"agent_id": map[string]interface{}{
					"type":        "string",
					"description": "Required for status, follow_up, delete.",
				},
				"prompt": map[string]interface{}{
					"type":        "string",
					"description": "Required for launch and follow_up.",
				},
				"repo": map[string]interface{}{
					"type":        "string",
					"description": "Optional for launch.",
				},
				"model": map[string]interface{}{
					"type":        "string",
					"description": "Optional for launch.",
				},
				"compact": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
			},
		},
		handleCursorCloudAgent,
	); err != nil {
		return fmt.Errorf("failed to register cursor_cloud_agent: %w", err)
	}

	return nil
}
