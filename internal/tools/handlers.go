// handlers.go — Top-level MCP tool dispatch: routes tool calls to per-tool handlers.
//
// Package tools implements all MCP tool handlers for the exarp-go server.
// Each tool is registered in registry.go and dispatched via this file.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/cache"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/models"
	"github.com/davidl71/exarp-go/proto"
	"github.com/davidl71/mcp-go-core/pkg/mcp/request"
	"github.com/davidl71/mcp-go-core/pkg/mcp/response"
	"github.com/spf13/cast"
)

// handleAnalyzeAlignment handles the analyze_alignment tool
// Fully native Go for both "todo2" and "prd" actions; no Python bridge.
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

	result, err := handleAnalyzeAlignmentNative(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("analyze_alignment failed: %w", err)
	}

	return result, nil
}

// handleGenerateConfig handles the generate_config tool
// Uses native Go implementation for all actions (rules, ignore, simplify) - fully native Go.
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
// Uses native Go implementation for all actions (server, git, docs, dod, cicd) - fully native Go with no fallback.
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
// Uses native Go implementation for both "git" and "patterns" actions - fully native Go.
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

	// Use native Go implementation only (git + patterns actions)
	return handleSetupHooksNative(ctx, params)
}

// handleCheckAttribution handles the check_attribution tool
// Uses native Go implementation only - fully native Go, no Python fallback.
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

	// Use native Go implementation only
	return handleCheckAttributionNative(ctx, params)
}

// handleAddExternalToolHints handles the add_external_tool_hints tool
// Uses native Go implementation - fully native Go, no Python bridge needed.
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
// Uses native Go for all actions (save, recall, search, list) - fully native Go, no Python bridge
// Note: Basic text search is native; semantic search enhancement can be added later in Go if needed.
func handleMemory(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Note: handleMemoryNative already handles protobuf parsing, so we just pass args through
	result, err := handleMemoryNative(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("memory failed: %w", err)
	}

	return result, nil
}

// handleMemoryMaint handles the memory_maint tool
// Uses native Go implementation only (health, gc, prune, consolidate, dream) - fully native Go, no Python fallback.
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

	// Use native Go implementation only (health, gc, prune, consolidate, dream)
	return handleMemoryMaintNative(ctx, args)
}

// handleReport handles the report tool.
// Fully native Go: overview, scorecard (Go projects only), briefing, prd. No Python bridge fallback.
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

	action := cast.ToString(params["action"])
	if action == "" {
		action = "overview"
	}

	// Route to native implementations based on action (no Python bridge fallback)
	switch action {
	case "scorecard":
		if !IsGoProject() {
			return nil, fmt.Errorf("scorecard action is only supported for Go projects (go.mod); use action=overview for other project types")
		}

		projectRoot, err := FindProjectRoot()
		if err != nil {
			return nil, fmt.Errorf("report scorecard: %w", err)
		}

		fastMode := true
		if _, ok := params["fast_mode"]; ok {
			fastMode = cast.ToBool(params["fast_mode"])
		}

		skipCache := cast.ToBool(params["skip_scorecard_cache"])
		cacheKey := "scorecard:" + projectRoot + "|" + strconv.FormatBool(fastMode)

		var scorecard *GoScorecardResult
		if !skipCache {
			if cached, ok := cache.GetScorecardCache().Get(cacheKey); ok {
				var decoded GoScorecardResult
				if err := json.Unmarshal(cached, &decoded); err == nil {
					scorecard = &decoded
				}
			}
		}
		if scorecard == nil {
			opts := &ScorecardOptions{FastMode: fastMode}
			var err error
			scorecard, err = GenerateGoScorecard(ctx, projectRoot, opts)
			if err != nil {
				return nil, fmt.Errorf("report scorecard: %w", err)
			}
			if !skipCache {
				if data, err := json.Marshal(scorecard); err == nil {
					cache.GetScorecardCache().Set(cacheKey, data, 5*time.Minute)
				}
			}
		}
		// Use proto for type-safe scorecard data (report/MLX path)
		scorecardProto := GoScorecardResultToProto(scorecard)
		scorecardMap := ProtoToScorecardMap(scorecardProto)

		outputFormat := cast.ToString(params["output_format"])
		if outputFormat == "json" {
			// Return JSON for Python/script consumers (e.g. project_overview, consolidated_reporting)
			blockers := scorecardProto.GetBlockers()
			productionReady := scorecard.Score >= 70.0 && len(blockers) == 0
			out := map[string]interface{}{
				"overall_score":    scorecard.Score,
				"production_ready": productionReady,
				"scores":           scorecardMap["scores"],
				"blockers":         blockers,
				"recommendations":  scorecardProto.GetRecommendations(),
				"metrics":          scorecardMap["metrics"],
			}
			AddTokenEstimateToResult(out)
			compact := cast.ToBool(params["compact"])
			return FormatResultOptionalCompact(out, "", compact)
		}
		// Use proto-derived map for MLX enhancement
		enhanced, err := enhanceReportWithMLX(ctx, scorecardMap, action)
		if err == nil && enhanced != nil {
			if insights, ok := enhanced["ai_insights"].(map[string]interface{}); ok {
				result := FormatGoScorecardWithMLX(scorecard, insights)
				wisdomResult := addWisdomToScorecard(result, scorecard)

				return []framework.TextContent{
					{Type: "text", Text: wisdomResult},
				}, nil
			}
		}

		result := FormatGoScorecardWithWisdom(scorecard)

		return []framework.TextContent{
			{Type: "text", Text: result},
		}, nil

	case "overview":
		result, err := handleReportOverview(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("report overview: %w", err)
		}

		return result, nil

	case "briefing":
		result, err := handleReportBriefing(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("report briefing: %w", err)
		}

		return result, nil

	case "prd":
		result, err := handleReportPRD(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("report prd: %w", err)
		}

		return result, nil

	case "plan":
		result, err := handleReportPlan(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("report plan: %w", err)
		}

		return result, nil

	case "scorecard_plans":
		result, err := handleReportScorecardPlans(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("report scorecard_plans: %w", err)
		}

		return result, nil

	case "parallel_execution_plan":
		result, err := handleReportParallelExecutionPlan(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("report parallel_execution_plan: %w", err)
		}

		return result, nil

	case "update_waves_from_plan":
		result, err := handleReportUpdateWavesFromPlan(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("report update_waves_from_plan: %w", err)
		}

		return result, nil
	}

	return nil, fmt.Errorf("report action %q not supported; supported: overview, scorecard, briefing, prd, plan, scorecard_plans, parallel_execution_plan, update_waves_from_plan", action)
}

// handleSecurity handles the security tool.
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

	action := cast.ToString(params["action"])
	if action == "" {
		action = "report"
	}

	// Route to native implementations (no Python bridge fallback)
	switch action {
	case "scan":
		result, err := handleSecurityScan(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("security scan: %w", err)
		}

		return result, nil
	case "alerts":
		result, err := handleSecurityAlerts(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("security alerts: %w", err)
		}

		return result, nil
	case "report":
		result, err := handleSecurityReport(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("security report: %w", err)
		}

		return result, nil
	}

	return nil, fmt.Errorf("security action %q not supported; supported: scan, alerts, report", action)
}

// handleTaskAnalysis handles the task_analysis tool (native Go only, no Python fallback).
// All actions (duplicates, tags, dependencies, parallelization, hierarchy) use native Go.
// Hierarchy uses the FM provider abstraction (Apple FM when available; clear error otherwise).
func handleTaskAnalysis(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	req, params, err := ParseTaskAnalysisRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	if req != nil {
		params = TaskAnalysisRequestToParams(req)
		request.ApplyDefaults(params, map[string]interface{}{
			"output_format": "text",
		})
	}

	return handleTaskAnalysisNative(ctx, params)
}

// handleTaskDiscovery handles the task_discovery tool
// Uses native Go implementation only (comments, markdown, orphans, create_tasks); no Python bridge.
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

	result, err := handleTaskDiscoveryNative(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("task_discovery failed: %w", err)
	}

	return result, nil
}

// handleInferTaskProgress handles the infer_task_progress tool (native Go).
// Evaluates Todo/In Progress tasks against codebase evidence and infers completion.
func handleInferTaskProgress(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	if params == nil {
		params = make(map[string]interface{})
	}

	return handleInferTaskProgressNative(ctx, params)
}

// handleTaskWorkflow handles the task_workflow tool
// Uses native Go only. Clarify uses DefaultFMProvider() (FM chain: Apple → Ollama → stub).
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
			"status":        models.StatusReview,
		})
	}

	result, err := handleTaskWorkflowNative(ctx, params)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// handleTesting handles the testing tool.
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

	action := cast.ToString(params["action"])
	if action == "" {
		action = "run"
	}

	// Route to native implementations (no Python bridge fallback)
	switch action {
	case "run":
		result, err := handleTestingRun(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("testing run: %w", err)
		}

		return result, nil
	case "coverage":
		result, err := handleTestingCoverage(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("testing coverage: %w", err)
		}

		return result, nil
	case "validate":
		result, err := handleTestingValidate(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("testing validate: %w", err)
		}

		return result, nil
	}

	return nil, fmt.Errorf("testing action %q not supported; supported: run, coverage, validate", action)
}

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
	request.ApplyDefaults(params, map[string]interface{}{"action": "list"})
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
		request.ApplyDefaults(params, map[string]interface{}{
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
		request.ApplyDefaults(params, map[string]interface{}{
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
		request.ApplyDefaults(params, map[string]interface{}{
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
		request.ApplyDefaults(params, map[string]interface{}{
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

		m, err := response.ConvertToMap(result)
		if err != nil {
			return nil, fmt.Errorf("failed to convert lint result: %w", err)
		}

		return response.FormatResult(m, "")
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
		request.ApplyDefaults(params, map[string]interface{}{
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

	return response.FormatResult(EstimationResultToMap(resp), "")
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

	return response.FormatResult(GitToolsResponseToMap(resp), "")
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
		request.ApplyDefaults(params, map[string]interface{}{
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
		request.ApplyDefaults(params, map[string]interface{}{
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
		request.ApplyDefaults(params, map[string]interface{}{
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
		request.ApplyDefaults(params, map[string]interface{}{
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
