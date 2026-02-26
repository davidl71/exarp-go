// registry_misc.go — Tool registrations: alignment, attribution, hints, catalog, workflow mode, session mode, context budget.
package tools

import (
	"fmt"

	"github.com/davidl71/exarp-go/internal/framework"
)

func registerMiscTools(server framework.MCPServer) error {
	// analyze_alignment
	if err := server.RegisterTool(
		"analyze_alignment",
		"[HINT: action=todo2|prd. Check task/PRD alignment. Use when validating backlog against requirements. Related: task_analysis.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"todo2", "prd"},
					"default": "todo2",
				},
				"create_followup_tasks": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"output_path": map[string]interface{}{
					"type": "string",
				},
			},
		},
		handleAnalyzeAlignment,
	); err != nil {
		return fmt.Errorf("failed to register analyze_alignment: %w", err)
	}

	// check_attribution
	if err := server.RegisterTool(
		"check_attribution",
		"[HINT: Verify third-party attribution compliance. Use when auditing licenses or before releases.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"output_path": map[string]interface{}{
					"type": "string",
				},
				"create_tasks": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
			},
		},
		handleCheckAttribution,
	); err != nil {
		return fmt.Errorf("failed to register check_attribution: %w", err)
	}

	// add_external_tool_hints
	if err := server.RegisterTool(
		"add_external_tool_hints",
		"[HINT: Scan source files and add tool-usage hints. Use when onboarding or enriching tool documentation in code.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"dry_run": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
				"output_path": map[string]interface{}{
					"type": "string",
				},
				"min_file_size": map[string]interface{}{
					"type":    "integer",
					"default": 50,
				},
			},
		},
		handleAddExternalToolHints,
	); err != nil {
		return fmt.Errorf("failed to register add_external_tool_hints: %w", err)
	}

	// tool_catalog
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

	// workflow_mode
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

	// infer_session_mode
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

	// context_budget
	// Note: Individual Python bridge tools (context_summarize, context_batch, prompt_log, prompt_analyze,
	// recommend_model, recommend_workflow) were removed in favor of unified tools in registry_ai.go:
	// - context(action=summarize|budget|batch)
	// - prompt_tracking(action=log|analyze)
	// - recommend(action=model|workflow|advisor).
	// Note: list_models tool removed - converted to stdio://models resource
	// See internal/resources/models.go for resource implementation
	if err := server.RegisterTool(
		"context_budget",
		"[HINT: Estimate token usage and suggest context reduction. Use when managing context window limits. Related: context.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"items": map[string]interface{}{
					"type":        "string",
					"description": "JSON array of items to analyze",
				},
				"budget_tokens": map[string]interface{}{
					"type":        "integer",
					"default":     4000,
					"description": "Target token budget",
				},
			},
			Required: []string{"items"},
		},
		handleContextBudget,
	); err != nil {
		return fmt.Errorf("failed to register context_budget: %w", err)
	}

	if err := RegisterResourcesAsTools(server); err != nil {
		return fmt.Errorf("failed to register resources-as-tools: %w", err)
	}

	return nil
}
