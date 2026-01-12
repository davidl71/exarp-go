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
// Uses native Go implementation for "todo2" action, falls back to Python bridge for "prd" and complex analysis
func handleAnalyzeAlignment(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Parse arguments
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Try native Go implementation first
	result, err := handleAnalyzeAlignmentNative(ctx, params)
	if err == nil {
		return result, nil
	}

	// If native implementation doesn't support the action, fall back to Python bridge
	bridgeResult, err := bridge.ExecutePythonTool(ctx, "analyze_alignment", params)
	if err != nil {
		return nil, fmt.Errorf("analyze_alignment failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: bridgeResult},
	}, nil
}

// handleGenerateConfig handles the generate_config tool
// Uses native Go implementation for all actions (rules, ignore, simplify) - fully native
func handleGenerateConfig(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Try native Go implementation first
	result, err := handleGenerateConfigNative(ctx, args)
	if err == nil {
		return result, nil
	}

	// If native implementation doesn't support the action, fall back to Python bridge
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Execute via Python bridge
	bridgeResult, err := bridge.ExecutePythonTool(ctx, "generate_config", params)
	if err != nil {
		return nil, fmt.Errorf("generate_config failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: bridgeResult},
	}, nil
}

// handleHealth handles the health tool
// Uses native Go implementation for "server" action, falls back to Python bridge for others
func handleHealth(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Parse arguments
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Try native Go implementation first
	result, err := handleHealthNative(ctx, params)
	if err == nil {
		return result, nil
	}

	// If native implementation doesn't support the action, fall back to Python bridge
	bridgeResult, err := bridge.ExecutePythonTool(ctx, "health", params)
	if err != nil {
		return nil, fmt.Errorf("health failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: bridgeResult},
	}, nil
}

// handleSetupHooks handles the setup_hooks tool
// Uses native Go implementation for both "git" and "patterns" actions - fully native Go
func handleSetupHooks(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Parse arguments
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Try native Go implementation first
	result, err := handleSetupHooksNative(ctx, params)
	if err == nil {
		return result, nil
	}

	// If native implementation doesn't support the action, fall back to Python bridge
	bridgeResult, err := bridge.ExecutePythonTool(ctx, "setup_hooks", params)
	if err != nil {
		return nil, fmt.Errorf("setup_hooks failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: bridgeResult},
	}, nil
}

// handleCheckAttribution handles the check_attribution tool
// Uses native Go implementation, falls back to Python bridge for complex analysis
func handleCheckAttribution(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Parse arguments
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Try native Go implementation first
	result, err := handleCheckAttributionNative(ctx, params)
	if err == nil {
		return result, nil
	}

	// If native implementation fails, fall back to Python bridge
	bridgeResult, err := bridge.ExecutePythonTool(ctx, "check_attribution", params)
	if err != nil {
		return nil, fmt.Errorf("check_attribution failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: bridgeResult},
	}, nil
}

// handleAddExternalToolHints handles the add_external_tool_hints tool
// Uses native Go implementation, falls back to Python bridge for complex analysis
func handleAddExternalToolHints(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Parse arguments
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Try native Go implementation first
	result, err := handleAddExternalToolHintsNative(ctx, params)
	if err == nil {
		return result, nil
	}

	// If native implementation fails, fall back to Python bridge
	bridgeResult, err := bridge.ExecutePythonTool(ctx, "add_external_tool_hints", params)
	if err != nil {
		return nil, fmt.Errorf("add_external_tool_hints failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: bridgeResult},
	}, nil
}

// Batch 2 Tool Handlers (T-28 through T-35)

// handleMemory handles the memory tool
// Uses native Go for CRUD operations, falls back to Python bridge for semantic search
func handleMemory(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	action, _ := params["action"].(string)
	if action == "" {
		action = "search"
	}

	// Try native Go implementation first (for save, recall, list, basic search)
	if action == "save" || action == "recall" || action == "list" {
		result, err := handleMemoryNative(ctx, args)
		if err == nil {
			return result, nil
		}
		// If native fails, fall through to Python bridge
	}

	// For search, try basic text search in Go first
	if action == "search" {
		result, err := handleMemoryNative(ctx, args)
		if err == nil {
			// Basic search succeeded - return results
			// For semantic search, user can explicitly request it or we can enhance later
			return result, nil
		}
		// If native search fails, fall through to Python bridge for semantic search
	}

	// Fall back to Python bridge for semantic search or unknown actions
	result, err := bridge.ExecutePythonTool(ctx, "memory", params)
	if err != nil {
		return nil, fmt.Errorf("memory failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleMemoryMaint handles the memory_maint tool
// Uses native Go for health, gc, prune; falls back to Python bridge for consolidate, dream
func handleMemoryMaint(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	action, _ := params["action"].(string)
	if action == "" {
		action = "health"
	}

	// Try native Go implementation first (for health, gc, prune, consolidate)
	// Dream still requires advisor integration, so falls back to Python bridge
	if action == "health" || action == "gc" || action == "prune" || action == "consolidate" {
		result, err := handleMemoryMaintNative(ctx, args)
		if err == nil {
			return result, nil
		}
		// If native fails, fall through to Python bridge
	}

	// For dream and other actions, use Python bridge
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

	// Route to native implementations based on action
	switch action {
	case "scorecard":
		// Check if go.mod exists (Go project)
		if IsGoProject() {
			// Use Go-specific scorecard with fast mode by default
			projectRoot := getProjectRoot()
			opts := &ScorecardOptions{FastMode: true}
			scorecard, err := GenerateGoScorecard(ctx, projectRoot, opts)
			if err == nil {
				// Convert to map for MLX enhancement
				scorecardMap := GoScorecardToMap(scorecard)

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
		// Fall through to Python bridge for non-Go projects

	case "overview":
		// Try native Go implementation
		result, err := handleReportOverview(ctx, params)
		if err == nil {
			return result, nil
		}
		// Fall through to Python bridge if native fails

	case "briefing":
		// Briefing uses devwisdom-go MCP server via Python bridge for now
		// TODO: Implement native Go MCP client for devwisdom-go
		result, err := handleReportBriefing(ctx, params)
		if err == nil {
			return result, nil
		}
		// Fall through to Python bridge if native fails

	case "prd":
		// Try native Go implementation
		result, err := handleReportPRD(ctx, params)
		if err == nil {
			return result, nil
		}
		// Fall through to Python bridge if native fails
	}

	// Execute via Python bridge (for unsupported actions or when native fails)
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

	action, _ := params["action"].(string)
	if action == "" {
		action = "report"
	}

	// Route to native implementations based on action
	switch action {
	case "scan":
		// Try native Go implementation first
		result, err := handleSecurityScan(ctx, params)
		if err == nil {
			return result, nil
		}
		// Fall through to Python bridge if native fails

	case "alerts":
		// Try native Go implementation first
		result, err := handleSecurityAlerts(ctx, params)
		if err == nil {
			return result, nil
		}
		// Fall through to Python bridge if native fails

	case "report":
		// Try native Go implementation first
		result, err := handleSecurityReport(ctx, params)
		if err == nil {
			return result, nil
		}
		// Fall through to Python bridge if native fails
	}

	// Execute via Python bridge (for unsupported actions or when native fails)
	result, err := bridge.ExecutePythonTool(ctx, "security", params)
	if err != nil {
		return nil, fmt.Errorf("security failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleTaskAnalysis handles the task_analysis tool
// Uses native Go implementation for all actions (duplicates, tags, dependencies, parallelization work on all platforms)
// Hierarchy action requires Apple FM (only available on macOS with CGO)
func handleTaskAnalysis(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Try native Go implementation first for all actions
	// Native implementations exist for duplicates, tags, dependencies, parallelization on all platforms
	// Hierarchy action requires Apple FM (checked inside native implementation)
	result, err := handleTaskAnalysisNative(ctx, params)
	if err == nil {
		return result, nil
	}

	// If native fails (e.g., hierarchy without Apple FM, or other errors), fall back to Python bridge
	bridgeResult, err := bridge.ExecutePythonTool(ctx, "task_analysis", params)
	if err != nil {
		return nil, fmt.Errorf("task_analysis failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: bridgeResult},
	}, nil
}

// handleTaskDiscovery handles the task_discovery tool
// Uses native Go implementation for basic scanning (comments, markdown, orphans)
// Apple FM is optional and only enhances semantic extraction when available
func handleTaskDiscovery(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Try native Go implementation first
	// Basic scanning (comments, markdown, orphans) works on all platforms
	// Apple FM is optional and only enhances results when available
	result, err := handleTaskDiscoveryNative(ctx, params)
	if err == nil {
		return result, nil
	}

	// If native fails, fall back to Python bridge
	bridgeResult, err := bridge.ExecutePythonTool(ctx, "task_discovery", params)
	if err != nil {
		return nil, fmt.Errorf("task_discovery failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: bridgeResult},
	}, nil
}

// handleTaskWorkflow handles the task_workflow tool
// Uses native Go implementation for all actions except clarify (which requires Apple FM)
// approve, create, sync, clarity, cleanup all work on all platforms
func handleTaskWorkflow(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Try native Go implementation first for all actions
	// Most actions (approve, create, sync, clarity, cleanup) work on all platforms
	// Only clarify action requires Apple FM (checked inside native implementation)
	result, err := handleTaskWorkflowNative(ctx, params)
	if err == nil {
		return result, nil
	}

	// If native fails (e.g., clarify without Apple FM, or other errors), fall back to Python bridge
	bridgeResult, err := bridge.ExecutePythonTool(ctx, "task_workflow", params)
	if err != nil {
		return nil, fmt.Errorf("task_workflow failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: bridgeResult},
	}, nil
}

// handleTesting handles the testing tool
func handleTesting(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	action, _ := params["action"].(string)
	if action == "" {
		action = "run"
	}

	// Route to native implementations based on action
	switch action {
	case "run":
		// Try native Go implementation first
		result, err := handleTestingRun(ctx, params)
		if err == nil {
			return result, nil
		}
		// Fall through to Python bridge if native fails

	case "coverage":
		// Try native Go implementation first
		result, err := handleTestingCoverage(ctx, params)
		if err == nil {
			return result, nil
		}
		// Fall through to Python bridge if native fails

	case "validate":
		// Try native Go implementation first
		result, err := handleTestingValidate(ctx, params)
		if err == nil {
			return result, nil
		}
		// Fall through to Python bridge if native fails

	case "suggest", "generate":
		// ML/AI features - use Python bridge
		// These require ML capabilities that are better handled in Python
	}

	// Execute via Python bridge (for unsupported actions or when native fails)
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
// Uses native Go implementation for "daily" and "discover" actions, falls back to Python bridge for "nightly" and "sprint"
func handleAutomation(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Try native Go implementation first (for daily and discover actions)
	result, err := handleAutomationNative(ctx, params)
	if err == nil {
		return result, nil
	}

	// If native implementation doesn't support the action, fall back to Python bridge
	resultText, err := bridge.ExecutePythonTool(ctx, "automation", params)
	if err != nil {
		return nil, fmt.Errorf("automation failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: resultText},
	}, nil
}

// handleToolCatalog handles the tool_catalog tool
// Uses native Go implementation (migrated from Python bridge)
func handleToolCatalog(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Use native Go implementation
	return handleToolCatalogNative(ctx, params)
}

// handleWorkflowMode handles the workflow_mode tool
// Uses native Go implementation (migrated from Python bridge)
func handleWorkflowMode(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Use native Go implementation
	return handleWorkflowModeNative(ctx, params)
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
// Uses native Go with Apple Foundation Models when available, falls back to Python bridge
func handleEstimation(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Try native Go implementation first
	result, err := handleEstimationNative(ctx, projectRoot, params)
	if err == nil {
		return []framework.TextContent{
			{Type: "text", Text: result},
		}, nil
	}

	// Fall back to Python bridge if native implementation fails or not available
	result, err = bridge.ExecutePythonTool(ctx, "estimation", params)
	if err != nil {
		return nil, fmt.Errorf("estimation failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleGitTools handles the git_tools tool using native Go implementation
func handleGitTools(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params GitToolsParams
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Use native Go implementation
	result, err := HandleGitToolsNative(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("git_tools failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleSession handles the session tool
// Uses native Go implementation for prime and handoff actions, falls back to Python bridge for prompts and assignee
func handleSession(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	action, _ := params["action"].(string)
	if action == "" {
		action = "prime"
	}

	// Try native Go implementation for prime and handoff actions
	if action == "prime" || action == "handoff" {
		result, err := handleSessionNative(ctx, params)
		if err == nil {
			return result, nil
		}
		// If native fails, fall through to Python bridge
	}

	// For prompts and assignee actions, or if native fails, use Python bridge
	result, err := bridge.ExecutePythonTool(ctx, "session", params)
	if err != nil {
		return nil, fmt.Errorf("session failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleInferSessionMode handles the infer_session_mode tool
// Uses native Go implementation (migrated from Python bridge)
func handleInferSessionMode(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Use native Go implementation
	return handleInferSessionModeNative(ctx, params)
}

// handleOllama handles the ollama tool
// Uses native Go HTTP client implementation
func handleOllama(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Try native Go implementation first
	result, err := handleOllamaNative(ctx, params)
	if err == nil {
		return result, nil
	}

	// Fall back to Python bridge for actions not yet implemented (e.g., docs, quality, summary)
	resultStr, err := bridge.ExecutePythonTool(ctx, "ollama", params)
	if err != nil {
		return nil, fmt.Errorf("ollama failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: resultStr},
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
		// Try native Go implementation for batch action
		result, err := handleContextBatchNative(ctx, params)
		if err == nil {
			return result, nil
		}
		// If native implementation fails, fall through to Python bridge

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
// Uses native Go implementation
func handlePromptTracking(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Use native Go implementation
	return handlePromptTrackingNative(ctx, params)
}

// handleRecommend handles the recommend tool (unified wrapper)
// Uses native Go implementation for "model" action, falls back to Python bridge for "workflow" and "advisor"
func handleRecommend(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Get action (default: "model")
	action := "model"
	if actionRaw, ok := params["action"].(string); ok && actionRaw != "" {
		action = actionRaw
	}

	// Try native Go implementation for "model" action
	if action == "model" {
		result, err := handleRecommendModelNative(ctx, params)
		if err == nil {
			return result, nil
		}
		// If native fails, fall through to Python bridge
	}

	// For "workflow" and "advisor" actions, or if native fails, use Python bridge
	bridgeResult, err := bridge.ExecutePythonTool(ctx, "recommend", params)
	if err != nil {
		return nil, fmt.Errorf("recommend failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: bridgeResult},
	}, nil
}

// Note: handleServerStatus removed - server_status tool converted to stdio://server/status resource
// See internal/resources/server.go for resource implementation

// Note: handleDemonstrateElicit and handleInteractiveTaskCreate removed
// These tools required FastMCP Context (not available in stdio mode)
// They were demonstration tools that don't work in exarp-go's primary stdio mode
