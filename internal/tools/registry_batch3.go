// registry_batch3.go — MCP tool registration batch.
package tools

import (
	"fmt"

	"github.com/davidl71/exarp-go/internal/framework"
)

// registerBatch3Tools registers Batch 3 tools (9 advanced tools).
func registerBatch3Tools(server framework.MCPServer) error {
	// T-37: automation
	if err := server.RegisterTool(
		"automation",
		"[HINT: action=daily|nightly|sprint|discover. Scheduled automation workflows. Use for routine maintenance, sprint automation, or discovering actionable tasks.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"daily", "nightly", "sprint", "discover"},
					"default": "daily",
				},
				"tasks": map[string]interface{}{
					"type":  "array",
					"items": map[string]interface{}{"type": "string"},
				},
				"include_slow": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
				"max_tasks_per_host": map[string]interface{}{
					"type":    "integer",
					"default": 5,
				},
				"max_parallel_tasks": map[string]interface{}{
					"type":    "integer",
					"default": 10,
				},
				"priority_filter": map[string]interface{}{
					"type": "string",
				},
				"tag_filter": map[string]interface{}{
					"type":  "array",
					"items": map[string]interface{}{"type": "string"},
				},
				"max_iterations": map[string]interface{}{
					"type":    "integer",
					"default": 10,
				},
				"auto_approve": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"extract_subtasks": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"run_analysis_tools": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"run_testing_tools": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"min_value_score": map[string]interface{}{
					"type":    "number",
					"default": 0.7,
				},
				"dry_run": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
				"output_path": map[string]interface{}{
					"type": "string",
				},
				"use_cursor_agent": map[string]interface{}{
					"type":        "boolean",
					"default":     false,
					"description": "When true, run Cursor CLI agent -p in project root and attach output to result (daily/nightly/sprint). Requires agent on PATH.",
				},
				"cursor_agent_prompt": map[string]interface{}{
					"type":        "string",
					"description": "Prompt for Cursor agent step when use_cursor_agent is true. Default: \"Review the backlog and suggest which task to do next\".",
				},
			},
		},
		handleAutomation,
	); err != nil {
		return fmt.Errorf("failed to register automation: %w", err)
	}

	// T-38: tool_catalog (help action only - list action converted to stdio://tools resources)
	if err := server.RegisterTool(
		"tool_catalog",
		"[HINT: action=help. Get detailed help for a specific tool by name. Use stdio://tools resource for listing all available tools.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"help"},
					"default": "help",
				},
				"tool_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the tool to get help for (required)",
				},
			},
			Required: []string{"tool_name"},
		},
		handleToolCatalog,
	); err != nil {
		return fmt.Errorf("failed to register tool_catalog: %w", err)
	}

	// T-39: workflow_mode
	if err := server.RegisterTool(
		"workflow_mode",
		"[HINT: action=focus|suggest|stats. Manage workflow modes and focus. Use when switching between dev/review/planning modes.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"focus", "suggest", "stats"},
					"default": "focus",
				},
				"mode": map[string]interface{}{
					"type": "string",
				},
				"enable_group": map[string]interface{}{
					"type": "string",
				},
				"disable_group": map[string]interface{}{
					"type": "string",
				},
				"status": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
				"text": map[string]interface{}{
					"type": "string",
				},
				"auto_switch": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
			},
		},
		handleWorkflowMode,
	); err != nil {
		return fmt.Errorf("failed to register workflow_mode: %w", err)
	}

	// T-40: lint
	if err := server.RegisterTool(
		"lint",
		"[HINT: action=run|analyze. Run linters or analyze results. Supports golangci-lint, gofmt, gomarklint (with link check). Use when checking code quality.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"run", "analyze"},
					"default": "run",
				},
				"path": map[string]interface{}{
					"type": "string",
				},
				"linter": map[string]interface{}{
					"type":    "string",
					"default": "golangci-lint",
				},
				"fix": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
				"analyze": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"select": map[string]interface{}{
					"type": "string",
				},
				"ignore": map[string]interface{}{
					"type": "string",
				},
				"problems_json": map[string]interface{}{
					"type": "string",
				},
				"include_hints": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"output_path": map[string]interface{}{
					"type": "string",
				},
			},
		},
		handleLint,
	); err != nil {
		return fmt.Errorf("failed to register lint: %w", err)
	}

	// T-41: estimation
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

	// T-42: git_tools
	if err := server.RegisterTool(
		"git_tools",
		"[HINT: action=commits|local_commits|branches|tasks|diff|graph|merge|set_branch. Git operations and task-commit linking. Use when reviewing changes or linking commits to tasks.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"commits", "local_commits", "branches", "tasks", "diff", "graph", "merge", "set_branch"},
					"default": "commits",
				},
				"task_id": map[string]interface{}{
					"type": "string",
				},
				"branch": map[string]interface{}{
					"type": "string",
				},
				"limit": map[string]interface{}{
					"type":    "integer",
					"default": 50,
				},
				"commit1": map[string]interface{}{
					"type": "string",
				},
				"commit2": map[string]interface{}{
					"type": "string",
				},
				"time1": map[string]interface{}{
					"type": "string",
				},
				"time2": map[string]interface{}{
					"type": "string",
				},
				"format": map[string]interface{}{
					"type":    "string",
					"default": "text",
				},
				"output_path": map[string]interface{}{
					"type": "string",
				},
				"max_commits": map[string]interface{}{
					"type":    "integer",
					"default": 50,
				},
				"source_branch": map[string]interface{}{
					"type": "string",
				},
				"target_branch": map[string]interface{}{
					"type": "string",
				},
				"conflict_strategy": map[string]interface{}{
					"type":    "string",
					"default": "newer",
				},
				"author": map[string]interface{}{
					"type":    "string",
					"default": "system",
				},
				"dry_run": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
			},
		},
		handleGitTools,
	); err != nil {
		return fmt.Errorf("failed to register git_tools: %w", err)
	}

	// T-43: session
	if err := server.RegisterTool(
		"session",
		"[HINT: action=prime|handoff|prompts|assignee. Session management. Call prime at start; handoff to save/resume context across sessions. Returns suggested_next tasks.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"prime", "handoff", "prompts", "assignee"},
					"default": "prime",
				},
				"include_hints": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"include_tasks": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"compact": map[string]interface{}{
					"type":        "boolean",
					"default":     false,
					"description": "When true (e.g. for prime), return compact JSON to reduce context size",
				},
				"include_cli_command": map[string]interface{}{
					"type":        "boolean",
					"default":     false,
					"description": "When true, include cursor_cli_suggestion (runnable agent -p command) in the response. Default false so chat only gets suggested_next_action (text); set true for CLI/TUI/scripts that may execute the command.",
				},
				"ask_preferences": map[string]interface{}{
					"type":        "boolean",
					"default":     false,
					"description": "When true and client supports elicitation, prompt user for include_tasks/include_hints preferences at prime time",
				},
				"override_mode": map[string]interface{}{
					"type": "string",
				},
				"task_id": map[string]interface{}{
					"type": "string",
				},
				"summary": map[string]interface{}{
					"type": "string",
				},
				"blockers": map[string]interface{}{
					"type": "string",
				},
				"next_steps": map[string]interface{}{
					"type": "string",
				},
				"unassign_my_tasks": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"include_git_status": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"include_point_in_time_snapshot": map[string]interface{}{
					"type":        "boolean",
					"default":     false,
					"description": "When true (handoff end), attach full task list as gzip+base64 snapshot in handoff",
				},
				"task_journal": map[string]interface{}{
					"type":        "string",
					"description": "Optional JSON array of {id, action?, ...} for tasks modified this session (handoff end)",
				},
				"modified_task_ids": map[string]interface{}{
					"type":        "array",
					"description": "Optional list of task IDs modified this session; stored as task_journal (handoff end)",
				},
				"limit": map[string]interface{}{
					"type":    "integer",
					"default": 5,
				},
				"dry_run": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
				"direction": map[string]interface{}{
					"type":    "string",
					"default": "both",
				},
				"prefer_agentic_tools": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"auto_commit": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"mode": map[string]interface{}{
					"type": "string",
				},
				"category": map[string]interface{}{
					"type": "string",
				},
				"keywords": map[string]interface{}{
					"type": "string",
				},
				"assignee_name": map[string]interface{}{
					"type": "string",
				},
				"assignee_type": map[string]interface{}{
					"type":    "string",
					"default": "agent",
				},
				"hostname": map[string]interface{}{
					"type": "string",
				},
				"status_filter": map[string]interface{}{
					"type": "string",
				},
				"sub_action": map[string]interface{}{
					"type": "string",
				},
				"output_path": map[string]interface{}{
					"type": "string",
				},
				"export_latest": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"priority_filter": map[string]interface{}{
					"type": "string",
				},
				"include_unassigned": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
				"max_tasks_per_agent": map[string]interface{}{
					"type":    "integer",
					"default": 5,
				},
			},
		},
		handleSession,
	); err != nil {
		return fmt.Errorf("failed to register session: %w", err)
	}

	// T-44: infer_session_mode
	if err := server.RegisterTool(
		"infer_session_mode",
		"[HINT: Infer session mode (AGENT/ASK/MANUAL) with confidence. Use when auto-detecting optimal interaction mode.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"force_recompute": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
			},
		},
		handleInferSessionMode,
	); err != nil {
		return fmt.Errorf("failed to register infer_session_mode: %w", err)
	}

	// T-6: MLX Integration tools (ollama and mlx)
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

	// Plan-and-execute flow (uses DefaultFMProvider: Apple FM or Ollama)
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

	return nil
}
