// registry_ai.go — Tool registrations: AI/LLM backends, memory, estimation, generation.
package tools

import (
	"fmt"

	"github.com/davidl71/exarp-go/internal/framework"
)

func registerAITools(server framework.MCPServer) error {
	// memory
	if err := server.RegisterTool(
		"memory",
		"[HINT: action=save|recall|search. Persist and retrieve AI discoveries. Use when saving learnings or searching past context.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"save", "recall", "search"},
					"default": "search",
				},
				"title": map[string]interface{}{
					"type": "string",
				},
				"content": map[string]interface{}{
					"type": "string",
				},
				"category": map[string]interface{}{
					"type":    "string",
					"default": "insight",
				},
				"task_id": map[string]interface{}{
					"type": "string",
				},
				"metadata": map[string]interface{}{
					"type": "string",
				},
				"include_related": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"query": map[string]interface{}{
					"type": "string",
				},
				"limit": map[string]interface{}{
					"type":    "integer",
					"default": 10,
				},
			},
		},
		handleMemory,
	); err != nil {
		return fmt.Errorf("failed to register memory: %w", err)
	}

	// memory_maint
	if err := server.RegisterTool(
		"memory_maint",
		"[HINT: action=health|gc|prune|consolidate|dream. Memory lifecycle management. Use when cleaning up old memories or consolidating insights.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"health", "gc", "prune", "consolidate", "dream"},
					"default": "health",
				},
				"max_age_days": map[string]interface{}{
					"type":    "integer",
					"default": 90,
				},
				"delete_orphaned": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"delete_duplicates": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"scorecard_max_age_days": map[string]interface{}{
					"type":    "integer",
					"default": 7,
				},
				"value_threshold": map[string]interface{}{
					"type":    "number",
					"default": 0.3,
				},
				"keep_minimum": map[string]interface{}{
					"type":    "integer",
					"default": 50,
				},
				"similarity_threshold": map[string]interface{}{
					"type":    "number",
					"default": 0.85,
				},
				"merge_strategy": map[string]interface{}{
					"type":    "string",
					"default": "newest",
				},
				"scope": map[string]interface{}{
					"type":    "string",
					"default": "week",
				},
				"advisors": map[string]interface{}{
					"type": "string",
				},
				"generate_insights": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"save_dream": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"dry_run": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"interactive": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
			},
		},
		handleMemoryMaint,
	); err != nil {
		return fmt.Errorf("failed to register memory_maint: %w", err)
	}

	// estimation
	if err := server.RegisterTool(
		"estimation",
		"[HINT: action=estimate|analyze|stats|estimate_batch. Task duration estimation. Use when planning work or estimating backlog. Supports FM/MLX/Ollama backends.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"estimate", "analyze", "stats", "estimate_batch"},
					"default": "estimate",
				},
				"task_ids": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "For estimate_batch: list of task IDs to estimate (or omit with status_filter for all matching)",
				},
				"status_filter": map[string]interface{}{
					"type":        "string",
					"description": "For estimate_batch: e.g. 'Todo' to estimate all Todo tasks (max 50)",
				},
				"name": map[string]interface{}{
					"type": "string",
				},
				"details": map[string]interface{}{
					"type":    "string",
					"default": "",
				},
				"tags": map[string]interface{}{
					"type": "string",
				},
				"tag_list": map[string]interface{}{
					"type":  "array",
					"items": map[string]interface{}{"type": "string"},
				},
				"priority": map[string]interface{}{
					"type":    "string",
					"default": "medium",
				},
				"use_historical": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"detailed": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
				"use_mlx": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"mlx_weight": map[string]interface{}{
					"type":    "number",
					"default": 0.3,
				},
				"local_ai_backend": map[string]interface{}{
					"type":        "string",
					"description": "Preferred local LLM for estimation: fm (Apple), mlx, ollama. Overrides use_apple_fm when set.",
					"enum":        []string{"", "fm", "mlx", "ollama"},
				},
			},
		},
		handleEstimation,
	); err != nil {
		return fmt.Errorf("failed to register estimation: %w", err)
	}

	// ollama
	if err := server.RegisterTool(
		"ollama",
		"[HINT: action=status|models|generate|pull|hardware|docs|quality|summary. Ollama LLM backend. Use for local text generation via Ollama server. Related: mlx, text_generate.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"status", "models", "generate", "pull", "hardware", "docs", "quality", "summary"},
					"default": "status",
				},
				"host": map[string]interface{}{
					"type": "string",
				},
				"prompt": map[string]interface{}{
					"type": "string",
				},
				"model": map[string]interface{}{
					"type":    "string",
					"default": "llama3.2",
				},
				"stream": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
				"options": map[string]interface{}{
					"type": "string",
				},
				"num_gpu": map[string]interface{}{
					"type": "integer",
				},
				"num_threads": map[string]interface{}{
					"type": "integer",
				},
				"context_size": map[string]interface{}{
					"type": "integer",
				},
				"file_path": map[string]interface{}{
					"type": "string",
				},
				"output_path": map[string]interface{}{
					"type": "string",
				},
				"style": map[string]interface{}{
					"type":    "string",
					"default": "google",
				},
				"include_suggestions": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"data": map[string]interface{}{
					"type": "string",
				},
				"level": map[string]interface{}{
					"type":    "string",
					"default": "brief",
				},
			},
		},
		handleOllama,
	); err != nil {
		return fmt.Errorf("failed to register ollama: %w", err)
	}

	// mlx
	if err := server.RegisterTool(
		"mlx",
		"[HINT: action=status|hardware|models|generate. MLX LLM backend for Apple Silicon. Use for on-device generation via MLX. Related: ollama, text_generate.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"status", "hardware", "models", "generate"},
					"default": "status",
				},
				"prompt": map[string]interface{}{
					"type": "string",
				},
				"model": map[string]interface{}{
					"type":    "string",
					"default": "mlx-community/Phi-3.5-mini-instruct-4bit",
				},
				"max_tokens": map[string]interface{}{
					"type":    "integer",
					"default": 512,
				},
				"temperature": map[string]interface{}{
					"type":    "number",
					"default": 0.7,
				},
				"verbose": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
			},
		},
		handleMlx,
	); err != nil {
		return fmt.Errorf("failed to register mlx: %w", err)
	}

	// llamacpp — local GGUF inference via llama.cpp bindings
	if err := registerLlamaCppTool(server); err != nil {
		return fmt.Errorf("failed to register llamacpp: %w", err)
	}

	// Apple Foundation Models tool (platform-specific, conditional compilation)
	if err := registerAppleFoundationModelsTool(server); err != nil {
		return err
	}

	// text_generate
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

	// context
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

	// prompt_tracking
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

	// recommend
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

	// cursor_cloud_agent
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

	// fm_plan_and_execute
	if err := server.RegisterTool(
		"fm_plan_and_execute",
		"[HINT: Plan-and-execute with FM/Ollama. Breaks task into subtasks (planner), runs workers in parallel, combines. Use for complex single-shot tasks.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"task": map[string]interface{}{
					"type":        "string",
					"description": "The high-level task to plan and execute",
				},
				"max_subtasks": map[string]interface{}{
					"type":        "integer",
					"default":     5,
					"description": "Maximum number of subtasks (1-20)",
				},
				"plan_max_tokens": map[string]interface{}{
					"type":        "integer",
					"default":     512,
					"description": "Max tokens for planner response",
				},
				"worker_max_tokens": map[string]interface{}{
					"type":        "integer",
					"default":     512,
					"description": "Max tokens per worker response",
				},
			},
			Required: []string{"task"},
		},
		handleFMPlanAndExecute,
	); err != nil {
		return fmt.Errorf("failed to register fm_plan_and_execute: %w", err)
	}

	// task_execute
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

	// research_aggregator
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

	return nil
}
