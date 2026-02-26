// handlers_ai.go — MCP tool adapters: AI, session, cursor, and recommendation tools.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/proto"
	"github.com/spf13/cast"
	"strings"
)

// handleCursorCloudAgent handles the cursor_cloud_agent tool (Cursor Cloud Agents API).
// JSON-only args; no protobuf. See docs/CURSOR_API_AND_CLI_INTEGRATION.md §3.4.
func handleCursorCloudAgent(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if len(args) > 0 {
		if err := json.Unmarshal(args, &params); err != nil {
			return nil, fmt.Errorf("cursor_cloud_agent: invalid JSON: %w", err)
		}
	}
	if params == nil {
		params = make(map[string]interface{})
	}
	framework.ApplyDefaults(params, map[string]interface{}{"action": "list"})
	return handleCursorCloudAgentNative(ctx, params)
}

// Batch 3 Tool Handlers (T-37 through T-44)

// handleAutomation handles the automation tool
// Uses native Go implementation for all actions (daily, nightly, sprint, discover) - fully native Go with no fallback.
func handleAutomation(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParseAutomationRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed
	if req != nil {
		params = AutomationRequestToParams(req)
		framework.ApplyDefaults(params, map[string]interface{}{
			"action": "daily",
		})
	}

	// Use native Go implementation - all actions are native
	return handleAutomationNative(ctx, params)
}

// handleToolCatalog handles the tool_catalog tool
// Uses native Go implementation (migrated from Python bridge).
func handleToolCatalog(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParseToolCatalogRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed
	if req != nil {
		params = ToolCatalogRequestToParams(req)
		framework.ApplyDefaults(params, map[string]interface{}{
			"action": "help",
		})
	}

	// Use native Go implementation
	return handleToolCatalogNative(ctx, params)
}

// handleWorkflowMode handles the workflow_mode tool
// Uses native Go implementation (migrated from Python bridge).
func handleWorkflowMode(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParseWorkflowModeRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed
	if req != nil {
		params = WorkflowModeRequestToParams(req)
		framework.ApplyDefaults(params, map[string]interface{}{
			"action": "focus",
		})
	}

	// Use native Go implementation
	return handleWorkflowModeNative(ctx, params)
}

// handleLint handles the lint tool.
func handleLint(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParseLintRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed
	if req != nil {
		params = LintRequestToParams(req)
		framework.ApplyDefaults(params, map[string]interface{}{
			"action": "run",
			"linter": "golangci-lint",
		})
	}

	// Extract parameters
	linter := "golangci-lint" // Default to golangci-lint for Go projects
	if l := strings.TrimSpace(cast.ToString(params["linter"])); l != "" {
		linter = l
	}

	path := ""
	if p := cast.ToString(params["path"]); p != "" {
		path = p
	}

	fix := cast.ToBool(params["fix"])

	// Check if this is a native linter - use native Go implementation
	nativeLinters := map[string]bool{
		"golangci-lint":    true,
		"golangcilint":     true,
		"go-vet":           true,
		"govet":            true,
		"go vet":           true,
		"gofmt":            true,
		"goimports":        true,
		"deadcode":         true,
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

		m, err := framework.ConvertToMap(result)
		if err != nil {
			return nil, fmt.Errorf("failed to convert lint result: %w", err)
		}

		return framework.FormatResult(m, "")
	}

	// Unsupported linter: native only (no bridge fallback)
	return nil, fmt.Errorf("lint linter %q not supported; supported: golangci-lint, go-vet, gofmt, goimports, markdownlint, shellcheck, auto", linter)
}

// handleEstimation handles the estimation tool
// EstimationResultToMap converts EstimationResult proto to map for response.FormatResult (unmarshals result_json).
func EstimationResultToMap(resp *proto.EstimationResult) map[string]interface{} {
	if resp == nil {
		return nil
	}

	out := map[string]interface{}{"action": resp.GetAction()}

	if resp.GetResultJson() != "" {
		var payload map[string]interface{}
		if json.Unmarshal([]byte(resp.GetResultJson()), &payload) == nil {
			for k, v := range payload {
				out[k] = v
			}
		}
	}

	return out
}

// Native Go only (stats, estimate, analyze when Apple FM available); no Python fallback.
func handleEstimation(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
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
		framework.ApplyDefaults(params, map[string]interface{}{
			"action":   "estimate",
			"priority": "medium",
		})
	}

	result, err := handleEstimationNative(ctx, projectRoot, params)
	if err != nil {
		return nil, err
	}

	action := "estimate"
	if a := strings.TrimSpace(cast.ToString(params["action"])); a != "" {
		action = a
	}

	resp := &proto.EstimationResult{Action: action, ResultJson: result}

	return framework.FormatResult(EstimationResultToMap(resp), "")
}

// handleGitTools handles the git_tools tool using native Go implementation.
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

	action := params.Action
	if action == "" {
		action = "commits"
	}

	resp := &proto.GitToolsResponse{Success: true, Action: action, ResultJson: result}

	return framework.FormatResult(GitToolsResponseToMap(resp), "")
}

// handleSession handles the session tool
// Uses native Go implementation for all actions (prime, handoff, prompts, assignee) - fully native Go.
func handleSession(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParseSessionRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed
	if req != nil {
		params = SessionRequestToParams(req)
		framework.ApplyDefaults(params, map[string]interface{}{
			"action":    "prime",
			"direction": "both",
		})
	}

	// Use native Go implementation only (prime, handoff, prompts, assignee)
	return handleSessionNative(ctx, params)
}

// handleInferSessionMode handles the infer_session_mode tool
// Uses native Go implementation (migrated from Python bridge).
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

// handleOllama handles the ollama tool via DefaultOllama (native only; no bridge fallback).
func handleOllama(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	req, params, err := ParseOllamaRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	if req != nil {
		params = OllamaRequestToParams(req)
		framework.ApplyDefaults(params, map[string]interface{}{
			"model": "llama3.2",
		})
	}

	result, err := DefaultOllama().Invoke(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("ollama failed: %w", err)
	}

	return result, nil
}

// handleMlx handles the mlx tool. Native-only: models (static list); status/hardware return unavailable message; generate returns error (use ollama or apple_foundation_models). Python bridge removed.
func handleMlx(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	req, params, err := ParseMlxRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	if req != nil {
		params = MlxRequestToParams(req)
		framework.ApplyDefaults(params, map[string]interface{}{
			"model": "mlx-community/Phi-3.5-mini-instruct-4bit",
		})
	}

	return handleMlxNative(ctx, params)
}

// mcp-generic-tools: Context management tools

// Phase 3 Migration: Unified tools
// Note: Individual tool handlers (handleContextSummarize, handleContextBatch, handlePromptLog,
// handlePromptAnalyze, handleRecommendModel, handleRecommendWorkflow) were removed in favor of
// unified handlers below that use action parameters.

// handleContext handles the context tool (unified wrapper)
// Uses native Go with Apple Foundation Models for summarization when available.
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
		framework.ApplyDefaults(params, map[string]interface{}{
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
		// Native summarization requires Apple Foundation Models
		if !FMAvailable() {
			return nil, fmt.Errorf("context summarize requires Apple Foundation Models (darwin/arm64 with CGO); use action=budget or action=batch for other operations")
		}

		result, err := handleContextSummarizeNative(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("context summarize: %w", err)
		}

		return result, nil

	case "budget":
		budgetArgs, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal budget arguments: %w", err)
		}

		result, err := handleContextBudget(ctx, budgetArgs)
		if err != nil {
			return nil, fmt.Errorf("context budget: %w", err)
		}

		return result, nil

	case "batch":
		result, err := handleContextBatchNative(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("context batch: %w", err)
		}

		return result, nil

	default:
		return nil, fmt.Errorf("unknown context action %q; use summarize, budget, or batch", action)
	}
}

// handlePromptTracking handles the prompt_tracking tool (unified wrapper)
// Uses native Go implementation.
func handlePromptTracking(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParsePromptTrackingRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed
	if req != nil {
		params = PromptTrackingRequestToParams(req)
		framework.ApplyDefaults(params, map[string]interface{}{
			"action":    "analyze",
			"iteration": 1,
			"days":      7,
		})
	}

	// Use native Go implementation
	return handlePromptTrackingNative(ctx, params)
}

// handleRecommend handles the recommend tool (unified wrapper).
// Fully native Go: model, workflow, advisor (devwisdom-go in-process). No Python bridge fallback.
func handleRecommend(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParseRecommendRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed
	if req != nil {
		params = RecommendRequestToParams(req)
		framework.ApplyDefaults(params, map[string]interface{}{
			"action":       "model",
			"optimize_for": "quality",
		})
	}

	// Get action (default: "model")
	action := "model"
	if actionRaw, ok := params["action"].(string); ok && actionRaw != "" {
		action = actionRaw
	}

	// Route to native implementations (no Python bridge fallback)
	switch action {
	case "model":
		result, err := handleRecommendModelNative(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("recommend model: %w", err)
		}

		return result, nil
	case "workflow":
		result, err := handleRecommendWorkflowNative(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("recommend workflow: %w", err)
		}

		return result, nil
	case "advisor":
		result, err := handleRecommendAdvisorNative(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("recommend advisor: %w", err)
		}

		return result, nil
	}

	return nil, fmt.Errorf("recommend action %q not supported; supported: model, workflow, advisor", action)
}

// Note: handleServerStatus removed - server_status tool converted to stdio://server/status resource
// See internal/resources/server.go for resource implementation

// Note: handleDemonstrateElicit and handleInteractiveTaskCreate removed
// These tools required FastMCP Context (not available in stdio mode)
// They were demonstration tools that don't work in exarp-go's primary stdio mode
