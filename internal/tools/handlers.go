package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/davidl71/exarp-go/internal/bridge"
	"github.com/davidl71/exarp-go/internal/framework"
)

// handleAnalyzeAlignment handles the analyze_alignment tool
func handleAnalyzeAlignment(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Parse arguments
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Execute via Python bridge
	result, err := bridge.ExecutePythonTool(ctx, "analyze_alignment", params)
	if err != nil {
		return nil, fmt.Errorf("analyze_alignment failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleGenerateConfig handles the generate_config tool
func handleGenerateConfig(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Parse arguments
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Execute via Python bridge
	result, err := bridge.ExecutePythonTool(ctx, "generate_config", params)
	if err != nil {
		return nil, fmt.Errorf("generate_config failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleHealth handles the health tool
func handleHealth(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Parse arguments
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Execute via Python bridge
	result, err := bridge.ExecutePythonTool(ctx, "health", params)
	if err != nil {
		return nil, fmt.Errorf("health failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleSetupHooks handles the setup_hooks tool
func handleSetupHooks(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Parse arguments
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Execute via Python bridge
	result, err := bridge.ExecutePythonTool(ctx, "setup_hooks", params)
	if err != nil {
		return nil, fmt.Errorf("setup_hooks failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleCheckAttribution handles the check_attribution tool
func handleCheckAttribution(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Parse arguments
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Execute via Python bridge
	result, err := bridge.ExecutePythonTool(ctx, "check_attribution", params)
	if err != nil {
		return nil, fmt.Errorf("check_attribution failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleAddExternalToolHints handles the add_external_tool_hints tool
func handleAddExternalToolHints(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Parse arguments
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Execute via Python bridge
	result, err := bridge.ExecutePythonTool(ctx, "add_external_tool_hints", params)
	if err != nil {
		return nil, fmt.Errorf("add_external_tool_hints failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// Batch 2 Tool Handlers (T-28 through T-35)

// handleMemory handles the memory tool
func handleMemory(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	result, err := bridge.ExecutePythonTool(ctx, "memory", params)
	if err != nil {
		return nil, fmt.Errorf("memory failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleMemoryMaint handles the memory_maint tool
func handleMemoryMaint(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	result, err := bridge.ExecutePythonTool(ctx, "memory_maint", params)
	if err != nil {
		return nil, fmt.Errorf("memory_maint failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleReport handles the report tool
func handleReport(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Check if this is a Go project and scorecard action
	action, _ := params["action"].(string)
	if action == "scorecard" {
		// Check if go.mod exists (Go project)
		if isGoProject() {
			// Use Go-specific scorecard
			projectRoot := getProjectRoot()
			scorecard, err := GenerateGoScorecard(ctx, projectRoot)
			if err == nil {
				// Format as text output
				result := FormatGoScorecard(scorecard)
				return []framework.TextContent{
					{Type: "text", Text: result},
				}, nil
			}
			// Fall through to Python bridge if Go scorecard fails
		}
	}

	// Execute via Python bridge (for non-Go projects or other actions)
	result, err := bridge.ExecutePythonTool(ctx, "report", params)
	if err != nil {
		return nil, fmt.Errorf("report failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleSecurity handles the security tool
func handleSecurity(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	result, err := bridge.ExecutePythonTool(ctx, "security", params)
	if err != nil {
		return nil, fmt.Errorf("security failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleTaskAnalysis handles the task_analysis tool
func handleTaskAnalysis(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	result, err := bridge.ExecutePythonTool(ctx, "task_analysis", params)
	if err != nil {
		return nil, fmt.Errorf("task_analysis failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleTaskDiscovery handles the task_discovery tool
func handleTaskDiscovery(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	result, err := bridge.ExecutePythonTool(ctx, "task_discovery", params)
	if err != nil {
		return nil, fmt.Errorf("task_discovery failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleTaskWorkflow handles the task_workflow tool
func handleTaskWorkflow(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	result, err := bridge.ExecutePythonTool(ctx, "task_workflow", params)
	if err != nil {
		return nil, fmt.Errorf("task_workflow failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleTesting handles the testing tool
func handleTesting(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	result, err := bridge.ExecutePythonTool(ctx, "testing", params)
	if err != nil {
		return nil, fmt.Errorf("testing failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// Batch 3 Tool Handlers (T-37 through T-44)

// handleAutomation handles the automation tool
func handleAutomation(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	result, err := bridge.ExecutePythonTool(ctx, "automation", params)
	if err != nil {
		return nil, fmt.Errorf("automation failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleToolCatalog handles the tool_catalog tool
func handleToolCatalog(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	result, err := bridge.ExecutePythonTool(ctx, "tool_catalog", params)
	if err != nil {
		return nil, fmt.Errorf("tool_catalog failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleWorkflowMode handles the workflow_mode tool
func handleWorkflowMode(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	result, err := bridge.ExecutePythonTool(ctx, "workflow_mode", params)
	if err != nil {
		return nil, fmt.Errorf("workflow_mode failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleLint handles the lint tool
func handleLint(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Extract parameters
	linter := "golangci-lint" // Default to golangci-lint for Go projects
	if l, ok := params["linter"].(string); ok && l != "" {
		linter = l
	}

	path := ""
	if p, ok := params["path"].(string); ok {
		path = p
	}

	fix := false
	if f, ok := params["fix"].(bool); ok {
		fix = f
	}

	// Check if this is a native linter - use native Go implementation
	nativeLinters := map[string]bool{
		"golangci-lint": true,
		"golangcilint":  true,
		"go-vet":        true,
		"govet":         true,
		"go vet":        true,
		"gofmt":         true,
		"goimports":     true,
		"markdownlint":  true,
		"markdownlint-cli": true,
		"mdl":           true,
		"markdown":      true,
		"shellcheck":    true,
		"shfmt":         true,
		"shell":         true,
		"auto":          true, // Auto-detection uses native implementation
	}

	if nativeLinters[linter] {
		// Use native Go implementation
		result, err := runLinter(ctx, linter, path, fix)
		if err != nil {
			return nil, fmt.Errorf("lint failed: %w", err)
		}

		// Convert result to JSON with indentation for readability
		resultJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}

		return []framework.TextContent{
			{Type: "text", Text: string(resultJSON)},
		}, nil
	}

	// Fallback to Python bridge for non-Go linters (e.g., ruff)
	result, err := bridge.ExecutePythonTool(ctx, "lint", params)
	if err != nil {
		return nil, fmt.Errorf("lint failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleEstimation handles the estimation tool
func handleEstimation(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	result, err := bridge.ExecutePythonTool(ctx, "estimation", params)
	if err != nil {
		return nil, fmt.Errorf("estimation failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleGitTools handles the git_tools tool
func handleGitTools(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	result, err := bridge.ExecutePythonTool(ctx, "git_tools", params)
	if err != nil {
		return nil, fmt.Errorf("git_tools failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleSession handles the session tool
func handleSession(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	result, err := bridge.ExecutePythonTool(ctx, "session", params)
	if err != nil {
		return nil, fmt.Errorf("session failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleInferSessionMode handles the infer_session_mode tool
func handleInferSessionMode(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	result, err := bridge.ExecutePythonTool(ctx, "infer_session_mode", params)
	if err != nil {
		return nil, fmt.Errorf("infer_session_mode failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleOllama handles the ollama tool
func handleOllama(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	result, err := bridge.ExecutePythonTool(ctx, "ollama", params)
	if err != nil {
		return nil, fmt.Errorf("ollama failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleMlx handles the mlx tool
func handleMlx(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	result, err := bridge.ExecutePythonTool(ctx, "mlx", params)
	if err != nil {
		return nil, fmt.Errorf("mlx failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

