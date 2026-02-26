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
	"github.com/davidl71/exarp-go/proto"
	"github.com/spf13/cast"
)

// handleAnalyzeAlignment handles the analyze_alignment tool.
// Fully native Go for both "todo2" and "prd" actions; no Python bridge.
var handleAnalyzeAlignment = WrapHandler(
	"analyze_alignment",
	func(args json.RawMessage) (any, map[string]interface{}, error) { return ParseAnalyzeAlignmentRequest(args) },
	func(req any) map[string]interface{} { return AnalyzeAlignmentRequestToParams(req.(*proto.AnalyzeAlignmentRequest)) },
	map[string]interface{}{"action": "todo2"},
	handleAnalyzeAlignmentNative,
)

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
		framework.ApplyDefaults(params, map[string]interface{}{
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

// handleHealth handles the health tool.
// Uses native Go implementation for all actions (server, git, docs, dod, cicd) - fully native Go with no fallback.
var handleHealth = WrapHandler(
	"health",
	func(args json.RawMessage) (any, map[string]interface{}, error) { return ParseHealthRequest(args) },
	func(req any) map[string]interface{} { return HealthRequestToParams(req.(*proto.HealthRequest)) },
	map[string]interface{}{"action": "server"},
	handleHealthNative,
)

// handleSetupHooks handles the setup_hooks tool.
// Uses native Go implementation for both "git" and "patterns" actions - fully native Go.
var handleSetupHooks = WrapHandler(
	"setup_hooks",
	func(args json.RawMessage) (any, map[string]interface{}, error) { return ParseSetupHooksRequest(args) },
	func(req any) map[string]interface{} { return SetupHooksRequestToParams(req.(*proto.SetupHooksRequest)) },
	map[string]interface{}{"action": "git"},
	handleSetupHooksNative,
)

// handleCheckAttribution handles the check_attribution tool.
// Uses native Go implementation only - fully native Go, no Python fallback.
var handleCheckAttribution = WrapHandler(
	"check_attribution",
	func(args json.RawMessage) (any, map[string]interface{}, error) { return ParseCheckAttributionRequest(args) },
	func(req any) map[string]interface{} { return CheckAttributionRequestToParams(req.(*proto.CheckAttributionRequest)) },
	nil,
	handleCheckAttributionNative,
)

// handleAddExternalToolHints handles the add_external_tool_hints tool.
// Uses native Go implementation - fully native Go, no Python bridge needed.
var handleAddExternalToolHints = WrapHandler(
	"add_external_tool_hints",
	func(args json.RawMessage) (any, map[string]interface{}, error) {
		return ParseAddExternalToolHintsRequest(args)
	},
	func(req any) map[string]interface{} {
		return AddExternalToolHintsRequestToParams(req.(*proto.AddExternalToolHintsRequest))
	},
	map[string]interface{}{"min_file_size": 50},
	handleAddExternalToolHintsNative,
)

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
		framework.ApplyDefaults(params, map[string]interface{}{
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
		framework.ApplyDefaults(params, map[string]interface{}{
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
			v, err, _ := scorecardFlight.Do(cacheKey, func() (interface{}, error) {
				return GenerateGoScorecard(ctx, projectRoot, opts)
			})
			if err != nil {
				return nil, fmt.Errorf("report scorecard: %w", err)
			}
			scorecard = v.(*GoScorecardResult)
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
		framework.ApplyDefaults(params, map[string]interface{}{
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
var handleTaskAnalysis = WrapHandler(
	"task_analysis",
	func(args json.RawMessage) (any, map[string]interface{}, error) { return ParseTaskAnalysisRequest(args) },
	func(req any) map[string]interface{} { return TaskAnalysisRequestToParams(req.(*proto.TaskAnalysisRequest)) },
	map[string]interface{}{"output_format": "text"},
	handleTaskAnalysisNative,
)

// handleTaskDiscovery handles the task_discovery tool.
// Uses native Go implementation only (comments, markdown, orphans, create_tasks); no Python bridge.
var handleTaskDiscovery = WrapHandler(
	"task_discovery",
	func(args json.RawMessage) (any, map[string]interface{}, error) { return ParseTaskDiscoveryRequest(args) },
	func(req any) map[string]interface{} { return TaskDiscoveryRequestToParams(req.(*proto.TaskDiscoveryRequest)) },
	map[string]interface{}{"action": "all", "json_pattern": "**/.todo2/state.todo2.json"},
	handleTaskDiscoveryNative,
)

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
		framework.ApplyDefaults(params, map[string]interface{}{
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
		framework.ApplyDefaults(params, map[string]interface{}{
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
