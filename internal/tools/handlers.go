package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/davidl71/exarp-go/internal/bridge"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/platform"
	"github.com/davidl71/exarp-go/internal/security"
	"github.com/davidl71/mcp-go-core/pkg/mcp/request"
)

// handleAnalyzeAlignment handles the analyze_alignment tool
// Uses native Go implementation for "todo2" action (fully native with followup task creation)
// Falls back to Python bridge for "prd" action (complex PRD analysis)
func handleAnalyzeAlignment(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParseAnalyzeAlignmentRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed
	if req != nil {
		params = AnalyzeAlignmentRequestToParams(req)
		request.ApplyDefaults(params, map[string]interface{}{
			"action": "todo2",
		})
	}

	// Try native Go implementation first
	result, err := handleAnalyzeAlignmentNative(ctx, params)
	if err == nil {
		return result, nil
	}

	// If native implementation doesn't support the action (e.g., "prd"), fall back to Python bridge
	bridgeResult, err := bridge.ExecutePythonTool(ctx, "analyze_alignment", params)
	if err != nil {
		return nil, fmt.Errorf("analyze_alignment failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: bridgeResult},
	}, nil
}

// handleGenerateConfig handles the generate_config tool
// Uses native Go implementation for all actions (rules, ignore, simplify) - fully native Go
func handleGenerateConfig(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParseGenerateConfigRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed
	if req != nil {
		params = GenerateConfigRequestToParams(req)
		request.ApplyDefaults(params, map[string]interface{}{
			"action": "rules",
		})
	}

	// Convert params to JSON for native handler (it expects json.RawMessage)
	argsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params: %w", err)
	}
	return handleGenerateConfigNative(ctx, argsJSON)
}

// handleHealth handles the health tool
// Uses native Go implementation for all actions (server, git, docs, dod, cicd) - fully native Go with no fallback
func handleHealth(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParseHealthRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed
	if req != nil {
		params = HealthRequestToParams(req)
		request.ApplyDefaults(params, map[string]interface{}{
			"action": "server",
		})
	}

	// Use native Go implementation - all actions are native
	return handleHealthNative(ctx, params)
}

// handleSetupHooks handles the setup_hooks tool
// Uses native Go implementation for both "git" and "patterns" actions - fully native Go
func handleSetupHooks(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParseSetupHooksRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed
	if req != nil {
		params = SetupHooksRequestToParams(req)
		request.ApplyDefaults(params, map[string]interface{}{
			"action": "git",
		})
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
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParseCheckAttributionRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed
	if req != nil {
		params = CheckAttributionRequestToParams(req)
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
// Uses native Go implementation - fully native Go, no Python bridge needed
func handleAddExternalToolHints(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParseAddExternalToolHintsRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed
	if req != nil {
		params = AddExternalToolHintsRequestToParams(req)
		request.ApplyDefaults(params, map[string]interface{}{
			"min_file_size": 50,
		})
	}

	// Use native Go implementation
	return handleAddExternalToolHintsNative(ctx, params)
}

// Batch 2 Tool Handlers (T-28 through T-35)

// handleMemory handles the memory tool
// Uses native Go for all actions (save, recall, search, list) - fully native Go
// Note: Basic text search is native; semantic search enhancement can be added later if needed
func handleMemory(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Note: handleMemoryNative already handles protobuf parsing, so we just pass args through
	// Try native Go implementation for all actions
	result, err := handleMemoryNative(ctx, args)
	if err == nil {
		return result, nil
	}

	// If native fails, fall back to Python bridge (for semantic search or error recovery)
	// Parse args to params map for Python bridge
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}
	resultText, err := bridge.ExecutePythonTool(ctx, "memory", params)
	if err != nil {
		return nil, fmt.Errorf("memory failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: resultText},
	}, nil
}

// handleMemoryMaint handles the memory_maint tool
// Uses native Go for health, gc, prune, consolidate actions
// Falls back to Python bridge for dream action (requires advisor integration via devwisdom-go MCP)
func handleMemoryMaint(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParseMemoryMaintRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed
	if req != nil {
		params = MemoryMaintRequestToParams(req)
		request.ApplyDefaults(params, map[string]interface{}{
			"action":         "health",
			"merge_strategy": "newest",
			"scope":          "week",
		})
	}

	action, _ := params["action"].(string)
	if action == "" {
		action = "health"
	}

	// Try native Go implementation first (for health, gc, prune, consolidate, dream)
	// All actions are now native (dream uses devwisdom-go wisdom engine directly)
	if action == "health" || action == "gc" || action == "prune" || action == "consolidate" || action == "dream" {
		result, err := handleMemoryMaintNative(ctx, args)
		if err == nil {
			return result, nil
		}
		// If native fails, fall through to Python bridge
	}

	// For unsupported actions or if native fails, use Python bridge
	result, err := bridge.ExecutePythonTool(ctx, "memory_maint", params)
	if err != nil {
		return nil, fmt.Errorf("memory_maint failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleReport handles the report tool
// Uses native Go for overview, scorecard (Go projects), and prd actions
// Falls back to Python bridge for briefing (requires devwisdom-go MCP client) and non-Go projects
func handleReport(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParseReportRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed (for compatibility with existing functions)
	if req != nil {
		params = ReportRequestToParams(req)
		// Set defaults for protobuf request
		request.ApplyDefaults(params, map[string]interface{}{
			"action":        "overview",
			"output_format": "text",
		})
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
			projectRoot, err := security.GetProjectRoot(".")
			if err == nil {
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
							// Format with MLX insights (wisdom is added separately after MLX)
							result := FormatGoScorecardWithMLX(scorecard, insights)
							// Add wisdom to MLX-enhanced scorecard
							wisdomResult := addWisdomToScorecard(result, scorecard)
							return []framework.TextContent{
								{Type: "text", Text: wisdomResult},
							}, nil
						}
					}

					// Format as text output with wisdom (without MLX if enhancement failed)
					result := FormatGoScorecardWithWisdom(scorecard)
					return []framework.TextContent{
						{Type: "text", Text: result},
					}, nil
				}
				// Fall through to Python bridge if Go scorecard fails
			}
			// Fall through to Python bridge if project root not found
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
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParseSecurityRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed
	if req != nil {
		params = SecurityRequestToParams(req)
		request.ApplyDefaults(params, map[string]interface{}{
			"action": "report",
			"repo":   "davidl71/exarp-go",
			"state":  "open",
		})
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
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParseTaskAnalysisRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed
	if req != nil {
		params = TaskAnalysisRequestToParams(req)
		request.ApplyDefaults(params, map[string]interface{}{
			"output_format": "text",
		})
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
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParseTaskDiscoveryRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed
	if req != nil {
		params = TaskDiscoveryRequestToParams(req)
		request.ApplyDefaults(params, map[string]interface{}{
			"action":       "all",
			"json_pattern": "**/.todo2/state.todo2.json",
		})
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
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParseTaskWorkflowRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed (for compatibility with existing functions)
	if req != nil {
		params = TaskWorkflowRequestToParams(req)
		// Set defaults for protobuf request
		request.ApplyDefaults(params, map[string]interface{}{
			"action":        "sync",
			"sub_action":    "list",
			"output_format": "text",
			"status":        "Review",
		})
	}

	// Try native Go implementation first for all actions
	// Most actions (approve, create, sync, clarity, cleanup) work on all platforms
	// Only clarify action requires Apple FM (checked inside native implementation)
	result, err := handleTaskWorkflowNative(ctx, params)
	if err == nil {
		return result, nil
	}

	// Only fallback to Python bridge for specific errors that indicate native isn't available
	// Don't fallback for sync/database errors - those should be reported, not hidden
	errStr := err.Error()
	shouldFallback := strings.Contains(errStr, "Apple Foundation Models not supported") ||
		strings.Contains(errStr, "not available on this platform") ||
		strings.Contains(errStr, "requires Apple Foundation Models")

	if !shouldFallback {
		// Native implementation failed with a real error (e.g., database save failed)
		// Return the error instead of falling back to Python bridge
		return nil, err
	}

	// If native fails due to platform limitations (e.g., clarify without Apple FM), fall back to Python bridge
	bridgeResult, bridgeErr := bridge.ExecutePythonTool(ctx, "task_workflow", params)
	if bridgeErr != nil {
		return nil, fmt.Errorf("task_workflow failed (native: %v, bridge: %w)", err, bridgeErr)
	}

	return []framework.TextContent{
		{Type: "text", Text: bridgeResult},
	}, nil
}

// handleTesting handles the testing tool
func handleTesting(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParseTestingRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed
	if req != nil {
		params = TestingRequestToParams(req)
		request.ApplyDefaults(params, map[string]interface{}{
			"action":         "run",
			"test_framework": "auto",
			"format":         "html",
		})
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
// Uses native Go implementation for all actions (daily, nightly, sprint, discover) - fully native Go with no fallback
func handleAutomation(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParseAutomationRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed
	if req != nil {
		params = AutomationRequestToParams(req)
		request.ApplyDefaults(params, map[string]interface{}{
			"action": "daily",
		})
	}

	// Use native Go implementation - all actions are native
	return handleAutomationNative(ctx, params)
}

// handleToolCatalog handles the tool_catalog tool
// Uses native Go implementation (migrated from Python bridge)
func handleToolCatalog(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParseToolCatalogRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed
	if req != nil {
		params = ToolCatalogRequestToParams(req)
		request.ApplyDefaults(params, map[string]interface{}{
			"action": "help",
		})
	}

	// Use native Go implementation
	return handleToolCatalogNative(ctx, params)
}

// handleWorkflowMode handles the workflow_mode tool
// Uses native Go implementation (migrated from Python bridge)
func handleWorkflowMode(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParseWorkflowModeRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed
	if req != nil {
		params = WorkflowModeRequestToParams(req)
		request.ApplyDefaults(params, map[string]interface{}{
			"action": "focus",
		})
	}

	// Use native Go implementation
	return handleWorkflowModeNative(ctx, params)
}

// handleLint handles the lint tool
func handleLint(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParseLintRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed
	if req != nil {
		params = LintRequestToParams(req)
		request.ApplyDefaults(params, map[string]interface{}{
			"action": "run",
			"linter": "golangci-lint",
		})
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
	projectRoot, err := security.GetProjectRoot(".")
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParseEstimationRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed
	if req != nil {
		params = EstimationRequestToParams(req)
		request.ApplyDefaults(params, map[string]interface{}{
			"action":   "estimate",
			"priority": "medium",
		})
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
	// Try protobuf first, fall back to JSON for backward compatibility
	req, paramsMap, err := ParseGitToolsRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	var params GitToolsParams
	if req != nil {
		// Convert protobuf request to GitToolsParams struct
		params = GitToolsParams{
			Action:           req.Action,
			TaskID:           req.TaskId,
			Branch:           req.Branch,
			Limit:            int(req.Limit),
			Commit1:          req.Commit1,
			Commit2:          req.Commit2,
			Time1:            req.Time1,
			Time2:            req.Time2,
			Format:           req.Format,
			OutputPath:       req.OutputPath,
			MaxCommits:       int(req.MaxCommits),
			SourceBranch:     req.SourceBranch,
			TargetBranch:     req.TargetBranch,
			ConflictStrategy: req.ConflictStrategy,
			Author:           req.Author,
			DryRun:           req.DryRun,
		}
		// Apply defaults to struct fields (special case - struct not map, can't use request.ApplyDefaults)
		// Note: This is acceptable as a special case since GitToolsParams is a struct, not a map
		if params.Format == "" {
			params.Format = "text"
		}
		if params.ConflictStrategy == "" {
			params.ConflictStrategy = "newer"
		}
		if params.Author == "" {
			params.Author = "system"
		}
	} else {
		// Convert params map to GitToolsParams struct
		paramsJSON, _ := json.Marshal(paramsMap)
		if err := json.Unmarshal(paramsJSON, &params); err != nil {
			return nil, fmt.Errorf("failed to convert params: %w", err)
		}
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
// Uses native Go implementation for all actions (prime, handoff, prompts, assignee) - fully native Go
func handleSession(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParseSessionRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed
	if req != nil {
		params = SessionRequestToParams(req)
		request.ApplyDefaults(params, map[string]interface{}{
			"action":    "prime",
			"direction": "both",
		})
	}

	// Try native Go implementation for all actions
	result, err := handleSessionNative(ctx, params)
	if err == nil {
		return result, nil
	}

	// If native fails, fall back to Python bridge
	resultText, err := bridge.ExecutePythonTool(ctx, "session", params)
	if err != nil {
		return nil, fmt.Errorf("session failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: resultText},
	}, nil
}

// handleInferSessionMode handles the infer_session_mode tool
// Uses native Go implementation (migrated from Python bridge)
func handleInferSessionMode(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParseInferSessionModeRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed
	if req != nil {
		params = InferSessionModeRequestToParams(req)
	}

	// Use native Go implementation
	return HandleInferSessionModeNative(ctx, params)
}

// handleOllama handles the ollama tool
// Uses native Go HTTP client implementation
func handleOllama(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParseOllamaRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed
	if req != nil {
		params = OllamaRequestToParams(req)
		request.ApplyDefaults(params, map[string]interface{}{
			"model": "llama3.2",
		})
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
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParseMlxRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed
	if req != nil {
		params = MlxRequestToParams(req)
		request.ApplyDefaults(params, map[string]interface{}{
			"model": "mlx-community/Phi-3.5-mini-instruct-4bit",
		})
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
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParseContextRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed (for compatibility with existing functions)
	if req != nil {
		params = ContextRequestToParams(req)
		// Set defaults for protobuf request
		request.ApplyDefaults(params, map[string]interface{}{
			"action":     "summarize",
			"level":      "brief",
			"max_tokens": 512,
		})
		if !req.Combine {
			params["combine"] = true // Default is true
		}
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
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParsePromptTrackingRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed
	if req != nil {
		params = PromptTrackingRequestToParams(req)
		request.ApplyDefaults(params, map[string]interface{}{
			"action":    "analyze",
			"iteration": 1,
			"days":      7,
		})
	}

	// Use native Go implementation
	return handlePromptTrackingNative(ctx, params)
}

// handleRecommend handles the recommend tool (unified wrapper)
// Uses native Go implementation for "model" and "workflow" actions, falls back to Python bridge for "advisor" (requires MCP client)
func handleRecommend(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParseRecommendRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed
	if req != nil {
		params = RecommendRequestToParams(req)
		request.ApplyDefaults(params, map[string]interface{}{
			"action":       "model",
			"optimize_for": "quality",
		})
	}

	// Get action (default: "model")
	action := "model"
	if actionRaw, ok := params["action"].(string); ok && actionRaw != "" {
		action = actionRaw
	}

	// Try native Go implementation for "model", "workflow", and "advisor" actions
	switch action {
	case "model":
		result, err := handleRecommendModelNative(ctx, params)
		if err == nil {
			return result, nil
		}
		// If native fails, fall through to Python bridge
	case "workflow":
		result, err := handleRecommendWorkflowNative(ctx, params)
		if err == nil {
			return result, nil
		}
		// If native fails, fall through to Python bridge
	case "advisor":
		result, err := handleRecommendAdvisorNative(ctx, params)
		if err == nil {
			return result, nil
		}
		// If native fails, fall through to Python bridge
	}

	// For unsupported actions or if native fails, use Python bridge
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
