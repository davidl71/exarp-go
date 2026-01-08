package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/davidl71/exarp-go/internal/bridge"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/platform"
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
// Uses MLX to enhance reports with AI-generated insights when available
func handleReport(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	action, _ := params["action"].(string)
	if action == "" {
		action = "overview"
	}

	// Check if this is a Go project and scorecard action
	if action == "scorecard" {
		// Check if go.mod exists (Go project)
		if isGoProject() {
			// Use Go-specific scorecard
			projectRoot := getProjectRoot()
			scorecard, err := GenerateGoScorecard(ctx, projectRoot)
			if err == nil {
				// Convert to map for MLX enhancement
				scorecardMap := goScorecardToMap(scorecard)
				
				// Enhance with MLX if available
				enhanced, err := enhanceReportWithMLX(ctx, scorecardMap, action)
				if err == nil && enhanced != nil {
					// Check if MLX insights were added
					if insights, ok := enhanced["ai_insights"].(map[string]interface{}); ok {
						// Format with MLX insights
						result := FormatGoScorecardWithMLX(scorecard, insights)
						return []framework.TextContent{
							{Type: "text", Text: result},
						}, nil
					}
				}

				// Format as text output (without MLX if enhancement failed)
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

	// Try to enhance with MLX insights
	// Parse the result to extract data
	var reportData map[string]interface{}
	if err := json.Unmarshal([]byte(result), &reportData); err == nil {
		// Enhance with MLX
		enhanced, err := enhanceReportWithMLX(ctx, reportData, action)
		if err == nil && enhanced != nil {
			// Check if we should include MLX insights
			includeMLX := true
			if mlxParam, ok := params["use_mlx"].(bool); ok {
				includeMLX = mlxParam
			} else {
				// Default to true for scorecard and overview
				includeMLX = action == "scorecard" || action == "overview"
			}

			if includeMLX {
				// Re-marshal with MLX insights
				enhancedJSON, err := json.MarshalIndent(enhanced, "", "  ")
				if err == nil {
					result = string(enhancedJSON)
				}
			}
		}
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
// Uses native Go with Apple Foundation Models when available
func handleTaskAnalysis(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Try native Go implementation first (for hierarchy action with Apple FM)
	action, _ := params["action"].(string)
	if action == "" {
		action = "duplicates"
	}

	if action == "hierarchy" {
		support := platform.CheckAppleFoundationModelsSupport()
		if support.Supported {
			result, err := handleTaskAnalysisNative(ctx, params)
			if err == nil {
				return result, nil
			}
			// If native fails, fall through to Python bridge
		}
	}

	// For other actions or when Apple FM unavailable, use Python bridge
	result, err := bridge.ExecutePythonTool(ctx, "task_analysis", params)
	if err != nil {
		return nil, fmt.Errorf("task_analysis failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleTaskDiscovery handles the task_discovery tool
// Uses native Go with Apple Foundation Models for semantic extraction when available
func handleTaskDiscovery(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Try native Go implementation first (uses Apple FM for semantic extraction)
	support := platform.CheckAppleFoundationModelsSupport()
	if support.Supported {
		result, err := handleTaskDiscoveryNative(ctx, params)
		if err == nil {
			return result, nil
		}
		// If native fails, fall through to Python bridge
	}

	// When Apple FM unavailable, use Python bridge
	result, err := bridge.ExecutePythonTool(ctx, "task_discovery", params)
	if err != nil {
		return nil, fmt.Errorf("task_discovery failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleTaskWorkflow handles the task_workflow tool
// Uses native Go with Apple Foundation Models for clarify action when available
func handleTaskWorkflow(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Try native Go implementation first (for clarify action with Apple FM)
	action, _ := params["action"].(string)
	if action == "" {
		action = "sync"
	}

	if action == "clarify" {
		support := platform.CheckAppleFoundationModelsSupport()
		if support.Supported {
			result, err := handleTaskWorkflowNative(ctx, params)
			if err == nil {
				return result, nil
			}
			// If native fails, fall through to Python bridge
		}
	}

	// For other actions or when Apple FM unavailable, use Python bridge
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
		"golangci-lint":    true,
		"golangcilint":     true,
		"go-vet":           true,
		"govet":            true,
		"go vet":           true,
		"gofmt":            true,
		"goimports":        true,
		"markdownlint":     true,
		"markdownlint-cli": true,
		"mdl":              true,
		"markdown":         true,
		"shellcheck":       true,
		"shfmt":            true,
		"shell":            true,
		"auto":             true, // Auto-detection uses native implementation
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

// mcp-generic-tools: Context management tools

// Phase 3 Migration: Unified tools
// Note: Individual tool handlers (handleContextSummarize, handleContextBatch, handlePromptLog,
// handlePromptAnalyze, handleRecommendModel, handleRecommendWorkflow) were removed in favor of
// unified handlers below that use action parameters.

// handleContext handles the context tool (unified wrapper)
// Uses native Go with Apple Foundation Models for summarization when available
func handleContext(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Get action (default: "summarize")
	action := "summarize"
	if actionRaw, ok := params["action"].(string); ok && actionRaw != "" {
		action = actionRaw
	}

	// Route to native Go implementations when available
	switch action {
	case "summarize":
		// Try native Go with Apple FM first
		support := platform.CheckAppleFoundationModelsSupport()
		if support.Supported {
			// Use native Go implementation with Apple FM
			result, err := handleContextSummarizeNative(ctx, params)
			if err == nil {
				return result, nil
			}
			// If native implementation fails, fall through to Python bridge
		}
		// If Apple FM not available, fall through to Python bridge

	case "budget":
		// Use native Go implementation for budget analysis
		// Convert params to json.RawMessage for handleContextBudget
		budgetArgs, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal budget arguments: %w", err)
		}
		result, err := handleContextBudget(ctx, budgetArgs)
		if err == nil {
			return result, nil
		}
		// If native implementation fails, fall through to Python bridge

	case "batch":
		// Batch action still uses Python bridge (not yet migrated to native Go)
		// Fall through to Python bridge

	default:
		// Unknown action, fall through to Python bridge
	}

	// For actions not handled natively or when native implementations fail, use Python bridge
	result, err := bridge.ExecutePythonTool(ctx, "context", params)
	if err != nil {
		return nil, fmt.Errorf("context failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handlePromptTracking handles the prompt_tracking tool (unified wrapper)
func handlePromptTracking(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	result, err := bridge.ExecutePythonTool(ctx, "prompt_tracking", params)
	if err != nil {
		return nil, fmt.Errorf("prompt_tracking failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleRecommend handles the recommend tool (unified wrapper)
func handleRecommend(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	result, err := bridge.ExecutePythonTool(ctx, "recommend", params)
	if err != nil {
		return nil, fmt.Errorf("recommend failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleServerStatus handles the server_status tool
func handleServerStatus(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	result, err := bridge.ExecutePythonTool(ctx, "server_status", params)
	if err != nil {
		return nil, fmt.Errorf("server_status failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// Note: handleDemonstrateElicit and handleInteractiveTaskCreate removed
// These tools required FastMCP Context (not available in stdio mode)
// They were demonstration tools that don't work in exarp-go's primary stdio mode
