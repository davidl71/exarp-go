// registry_batch1.go — MCP tool registration batch.
package tools

import (
	"fmt"

	"github.com/davidl71/exarp-go/internal/framework"
)

func registerBatch1Tools(server framework.MCPServer) error {
	// T-22: analyze_alignment
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

	// T-23: generate_config
	if err := server.RegisterTool(
		"generate_config",
		"[HINT: action=rules|ignore|simplify. Generate Cursor config files (.cursor/rules/*.mdc, .cursorignore). Use when setting up Cursor IDE. Not for Claude Code.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"rules", "ignore", "simplify"},
					"default": "rules",
				},
				"rules": map[string]interface{}{
					"type": "string",
				},
				"overwrite": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
				"analyze_only": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
				"include_indexing": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"analyze_project": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"rule_files": map[string]interface{}{
					"type": "string",
				},
				"output_dir": map[string]interface{}{
					"type": "string",
				},
				"dry_run": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
			},
		},
		handleGenerateConfig,
	); err != nil {
		return fmt.Errorf("failed to register generate_config: %w", err)
	}

	// T-24: health
	if err := server.RegisterTool(
		"health",
		"[HINT: action=server|git|docs|dod|cicd|tools|ctags. Check project health and component status. Use when diagnosing issues or before releases. action=ctags reports if universal-ctags and tags file are present.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"server", "git", "docs", "dod", "cicd", "tools", "ctags"},
					"default": "server",
				},
				"agent_name": map[string]interface{}{
					"type": "string",
				},
				"check_remote": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"output_path": map[string]interface{}{
					"type": "string",
				},
				"create_tasks": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"task_id": map[string]interface{}{
					"type": "string",
				},
				"changed_files": map[string]interface{}{
					"type": "string",
				},
				"auto_check": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"workflow_path": map[string]interface{}{
					"type": "string",
				},
				"check_runners": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
			},
		},
		handleHealth,
	); err != nil {
		return fmt.Errorf("failed to register health: %w", err)
	}

	// T-25: setup_hooks
	if err := server.RegisterTool(
		"setup_hooks",
		"[HINT: action=git|patterns. Install git hooks and automation patterns. Use when setting up dev environment or CI hooks.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"git", "patterns"},
					"default": "git",
				},
				"hooks": map[string]interface{}{
					"type":  "array",
					"items": map[string]interface{}{"type": "string"},
				},
				"patterns": map[string]interface{}{
					"type": "string",
				},
				"config_path": map[string]interface{}{
					"type": "string",
				},
				"install": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"dry_run": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
			},
		},
		handleSetupHooks,
	); err != nil {
		return fmt.Errorf("failed to register setup_hooks: %w", err)
	}

	// T-26: check_attribution
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

	// T-27: add_external_tool_hints
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

	return nil
}
