// handlers.go — Top-level MCP tool dispatch: core tool adapters (analyzeAlignment → testing).
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
	"github.com/davidl71/exarp-go/internal/cache"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/models"
	"github.com/davidl71/mcp-go-core/pkg/mcp/request"
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
