// protobuf_helpers_report.go — Protobuf helpers: Scorecard, BriefingData, TaskWorkflow, and Session conversions.
// See also: protobuf_helpers.go, protobuf_helpers_tools.go
package tools

import (
	"encoding/json"
	"github.com/davidl71/exarp-go/proto"
	"github.com/davidl71/mcp-go-core/pkg/mcp/request"
)

// ─── Contents ───────────────────────────────────────────────────────────────
//   GoScorecardResultToProto
//   BriefingDataToMap — BriefingDataToMap converts proto.BriefingData to map for JSON output (same shape as legacy briefing map).
//   ProtoToScorecardMap — ProtoToScorecardMap converts proto.ScorecardData to map for MLX/JSON (same shape as GoScorecardToMap).
//   ProjectOverviewDataToProto — ProjectOverviewDataToProto converts map[string]interface{} to proto.ProjectOverviewData
//   ProtoToProjectOverviewData — ProtoToProjectOverviewData converts proto.ProjectOverviewData to map[string]interface{}
//   ParseTaskWorkflowRequest — ParseTaskWorkflowRequest parses a task_workflow tool request (protobuf or JSON).
//   TaskWorkflowRequestToParams — TaskWorkflowRequestToParams converts a protobuf TaskWorkflowRequest to params map
//   ParseHealthRequest — ParseHealthRequest parses a health tool request (protobuf or JSON).
//   HealthRequestToParams — HealthRequestToParams converts a protobuf HealthRequest to params map
//   ParseSecurityRequest — ParseSecurityRequest parses a security tool request (protobuf or JSON).
//   SecurityRequestToParams — SecurityRequestToParams converts a protobuf SecurityRequest to params map
//   ParseInferSessionModeRequest — ParseInferSessionModeRequest parses an infer_session_mode request (protobuf or JSON).
//   InferSessionModeRequestToParams — InferSessionModeRequestToParams converts a protobuf InferSessionModeRequest to params map
//   ParseToolCatalogRequest — ParseToolCatalogRequest parses a tool_catalog request (protobuf or JSON).
//   ToolCatalogRequestToParams — ToolCatalogRequestToParams converts a protobuf ToolCatalogRequest to params map
//   ParseWorkflowModeRequest — ParseWorkflowModeRequest parses a workflow_mode request (protobuf or JSON).
//   WorkflowModeRequestToParams — WorkflowModeRequestToParams converts a protobuf WorkflowModeRequest to params map
//   ParseEstimationRequest — ParseEstimationRequest parses an estimation request (protobuf or JSON).
//   EstimationRequestToParams — EstimationRequestToParams converts a protobuf EstimationRequest to params map
//   ParseSessionRequest — ParseSessionRequest parses a session request (protobuf or JSON).
//   SessionRequestToParams — SessionRequestToParams converts a protobuf SessionRequest to params map
//   ParseGitToolsRequest — ParseGitToolsRequest parses a git_tools request (protobuf or JSON).
//   GitToolsRequestToParams — GitToolsRequestToParams converts a protobuf GitToolsRequest to params map
// ────────────────────────────────────────────────────────────────────────────

// ─── GoScorecardResultToProto ───────────────────────────────────────────────
func GoScorecardResultToProto(scorecard *GoScorecardResult) *proto.ScorecardData {
	if scorecard == nil {
		return nil
	}

	pb := &proto.ScorecardData{
		Score:           scorecard.Score,
		Recommendations: scorecard.Recommendations,
		Blockers:        ExtractBlockers(scorecard),
		TestCoverage:    scorecard.Health.GoTestCoverage,
		ComponentScores: map[string]float64{
			"testing":       calculateTestingScore(scorecard),
			"security":      calculateSecurityScore(scorecard),
			"documentation": calculateDocumentationScore(scorecard),
			"completion":    calculateCompletionScore(scorecard),
			"ci_cd":         calculateCICDScore(scorecard),
		},
		MetricsCounts: map[string]int32{
			"go_files":      int32(scorecard.Metrics.GoFiles),
			"go_lines":      int32(scorecard.Metrics.GoLines),
			"go_test_files": int32(scorecard.Metrics.GoTestFiles),
			"go_test_lines": int32(scorecard.Metrics.GoTestLines),
			"mcp_tools":     int32(scorecard.Metrics.MCPTools),
			"mcp_prompts":   int32(scorecard.Metrics.MCPPrompts),
			"mcp_resources": int32(scorecard.Metrics.MCPResources),
		},
	}

	return pb
}

// ─── BriefingDataToMap ──────────────────────────────────────────────────────
// BriefingDataToMap converts proto.BriefingData to map for JSON output (same shape as legacy briefing map).
func BriefingDataToMap(pb *proto.BriefingData) map[string]interface{} {
	if pb == nil {
		return make(map[string]interface{})
	}

	quotes := make([]interface{}, 0, len(pb.Quotes))

	for _, q := range pb.Quotes {
		if q == nil {
			continue
		}

		quotes = append(quotes, map[string]interface{}{
			"quote":         q.Quote,
			"source":        q.Source,
			"encouragement": q.Encouragement,
			"wisdom_source": q.WisdomSource,
			"wisdom_icon":   q.WisdomIcon,
		})
	}

	return map[string]interface{}{
		"date":    pb.Date,
		"score":   pb.Score,
		"quotes":  quotes,
		"sources": pb.Sources,
	}
}

// ─── ProtoToScorecardMap ────────────────────────────────────────────────────
// ProtoToScorecardMap converts proto.ScorecardData to map for MLX/JSON (same shape as GoScorecardToMap).
func ProtoToScorecardMap(pb *proto.ScorecardData) map[string]interface{} {
	if pb == nil {
		return make(map[string]interface{})
	}

	scores := make(map[string]interface{})
	for k, v := range pb.ComponentScores {
		scores[k] = v
	}

	metrics := make(map[string]interface{})
	for k, v := range pb.MetricsCounts {
		metrics[k] = int(v)
	}

	metrics["test_coverage"] = pb.TestCoverage

	return map[string]interface{}{
		"overall_score":   pb.Score,
		"scores":          scores,
		"blockers":        pb.Blockers,
		"recommendations": pb.Recommendations,
		"metrics":         metrics,
	}
}

// ─── ProjectOverviewDataToProto ─────────────────────────────────────────────
// ProjectOverviewDataToProto converts map[string]interface{} to proto.ProjectOverviewData
// NOTE: Currently unused - kept for potential future protobuf-based report processing.
func ProjectOverviewDataToProto(data map[string]interface{}) *proto.ProjectOverviewData {
	pb := &proto.ProjectOverviewData{}

	if project, ok := data["project"].(map[string]interface{}); ok {
		pb.Project = ProjectInfoToProto(project)
	}

	if health, ok := data["health"].(map[string]interface{}); ok {
		pb.Health = HealthDataToProto(health)
	}

	if codebase, ok := data["codebase"].(map[string]interface{}); ok {
		pb.Codebase = CodebaseMetricsToProto(codebase)
	}

	if tasks, ok := data["tasks"].(map[string]interface{}); ok {
		pb.Tasks = TaskMetricsToProto(tasks)
	}

	if phases, ok := data["phases"].(map[string]interface{}); ok {
		// Convert phases map to repeated ProjectPhase
		for key, phaseRaw := range phases {
			if phase, ok := phaseRaw.(map[string]interface{}); ok {
				pbPhase := &proto.ProjectPhase{
					Name: key,
				}
				if status, ok := phase["status"].(string); ok {
					pbPhase.Status = status
				}

				if progress, ok := phase["progress"].(int); ok {
					pbPhase.Progress = int32(progress)
				} else if progress, ok := phase["progress"].(float64); ok {
					pbPhase.Progress = int32(progress)
				}

				pb.Phases = append(pb.Phases, pbPhase)
			}
		}
	}

	if risks, ok := data["risks"].([]interface{}); ok {
		for _, riskRaw := range risks {
			if risk, ok := riskRaw.(map[string]interface{}); ok {
				pbRisk := &proto.RiskOrBlocker{}
				if typ, ok := risk["type"].(string); ok {
					pbRisk.Type = typ
				}

				if desc, ok := risk["description"].(string); ok {
					pbRisk.Description = desc
				}

				if taskID, ok := risk["task_id"].(string); ok {
					pbRisk.TaskId = taskID
				}

				if priority, ok := risk["priority"].(string); ok {
					pbRisk.Priority = priority
				}

				pb.Risks = append(pb.Risks, pbRisk)
			}
		}
	}

	if actions, ok := data["next_actions"].([]interface{}); ok {
		for _, actionRaw := range actions {
			if action, ok := actionRaw.(map[string]interface{}); ok {
				pbAction := &proto.NextAction{}
				if taskID, ok := action["task_id"].(string); ok {
					pbAction.TaskId = taskID
				}

				if name, ok := action["name"].(string); ok {
					pbAction.Name = name
				}

				if priority, ok := action["priority"].(string); ok {
					pbAction.Priority = priority
				}

				if hours, ok := action["estimated_hours"].(float64); ok {
					pbAction.EstimatedHours = hours
				}

				pb.NextActions = append(pb.NextActions, pbAction)
			}
		}
	}

	if generatedAt, ok := data["generated_at"].(string); ok {
		pb.GeneratedAt = generatedAt
	}

	return pb
}

// ─── ProtoToProjectOverviewData ─────────────────────────────────────────────
// ProtoToProjectOverviewData converts proto.ProjectOverviewData to map[string]interface{}
// NOTE: Currently unused - kept for potential future protobuf-based report processing.
func ProtoToProjectOverviewData(pb *proto.ProjectOverviewData) map[string]interface{} {
	data := make(map[string]interface{})

	if pb.Project != nil {
		data["project"] = ProtoToProjectInfo(pb.Project)
	}

	if pb.Health != nil {
		data["health"] = ProtoToHealthData(pb.Health)
	}

	if pb.Codebase != nil {
		data["codebase"] = ProtoToCodebaseMetrics(pb.Codebase)
	}

	if pb.Tasks != nil {
		data["tasks"] = ProtoToTaskMetrics(pb.Tasks)
	}

	if len(pb.Phases) > 0 {
		phases := make(map[string]interface{})
		for _, phase := range pb.Phases {
			phases[phase.Name] = map[string]interface{}{
				"name":     phase.Name,
				"status":   phase.Status,
				"progress": int(phase.Progress),
			}
		}

		data["phases"] = phases
	}

	if len(pb.Risks) > 0 {
		risks := make([]interface{}, 0, len(pb.Risks))
		for _, risk := range pb.Risks {
			risks = append(risks, map[string]interface{}{
				"type":        risk.Type,
				"description": risk.Description,
				"task_id":     risk.TaskId,
				"priority":    risk.Priority,
			})
		}

		data["risks"] = risks
	}

	if len(pb.NextActions) > 0 {
		actions := make([]interface{}, 0, len(pb.NextActions))
		for _, action := range pb.NextActions {
			actions = append(actions, map[string]interface{}{
				"task_id":         action.TaskId,
				"name":            action.Name,
				"priority":        action.Priority,
				"estimated_hours": action.EstimatedHours,
			})
		}

		data["next_actions"] = actions
	}

	if pb.GeneratedAt != "" {
		data["generated_at"] = pb.GeneratedAt
	}

	return data
}

// ─── ParseTaskWorkflowRequest ───────────────────────────────────────────────
// ParseTaskWorkflowRequest parses a task_workflow tool request (protobuf or JSON).
func ParseTaskWorkflowRequest(args json.RawMessage) (*proto.TaskWorkflowRequest, map[string]interface{}, error) {
	req, params, err := request.ParseRequest(args, func() *proto.TaskWorkflowRequest {
		return &proto.TaskWorkflowRequest{}
	})
	if err != nil {
		return nil, nil, err
	}

	if req != nil {
		return req, nil, nil
	}

	return nil, params, nil
}

// ─── TaskWorkflowRequestToParams ────────────────────────────────────────────
// TaskWorkflowRequestToParams converts a protobuf TaskWorkflowRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core
// to eliminate repetitive field-by-field conversion code.
func TaskWorkflowRequestToParams(req *proto.TaskWorkflowRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	// Use generic converter with options matching existing behavior
	params, err := request.ProtobufToParams(req, &request.ProtobufToParamsOptions{
		FilterEmptyStrings:  true, // Filter empty strings and zero numbers
		StringifyArrays:     true, // Convert arrays to JSON strings
		ConvertFloat64ToInt: true, // Convert stale_threshold_hours from float64 to int
		Float64ToIntFields:  []string{"stale_threshold_hours"},
	})
	if err != nil {
		// Fallback to empty map if conversion fails (shouldn't happen in practice)
		return make(map[string]interface{})
	}

	return params
}

// ─── ParseHealthRequest ─────────────────────────────────────────────────────
// ParseHealthRequest parses a health tool request (protobuf or JSON).
func ParseHealthRequest(args json.RawMessage) (*proto.HealthRequest, map[string]interface{}, error) {
	req, params, err := request.ParseRequest(args, func() *proto.HealthRequest {
		return &proto.HealthRequest{}
	})
	if err != nil {
		return nil, nil, err
	}

	if req != nil {
		return req, nil, nil
	}

	return nil, params, nil
}

// ─── HealthRequestToParams ──────────────────────────────────────────────────
// HealthRequestToParams converts a protobuf HealthRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func HealthRequestToParams(req *proto.HealthRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := request.ProtobufToParams(req, &request.ProtobufToParamsOptions{
		FilterEmptyStrings: true,
		StringifyArrays:    false,
	})
	if err != nil {
		return make(map[string]interface{})
	}

	return params
}

// ─── ParseSecurityRequest ───────────────────────────────────────────────────
// ParseSecurityRequest parses a security tool request (protobuf or JSON).
func ParseSecurityRequest(args json.RawMessage) (*proto.SecurityRequest, map[string]interface{}, error) {
	req, params, err := request.ParseRequest(args, func() *proto.SecurityRequest {
		return &proto.SecurityRequest{}
	})
	if err != nil {
		return nil, nil, err
	}

	if req != nil {
		return req, nil, nil
	}

	return nil, params, nil
}

// ─── SecurityRequestToParams ────────────────────────────────────────────────
// SecurityRequestToParams converts a protobuf SecurityRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func SecurityRequestToParams(req *proto.SecurityRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := request.ProtobufToParams(req, &request.ProtobufToParamsOptions{
		FilterEmptyStrings: true,
		StringifyArrays:    true, // Languages is an array
	})
	if err != nil {
		return make(map[string]interface{})
	}

	return params
}

// ─── ParseInferSessionModeRequest ───────────────────────────────────────────
// ParseInferSessionModeRequest parses an infer_session_mode request (protobuf or JSON).
func ParseInferSessionModeRequest(args json.RawMessage) (*proto.InferSessionModeRequest, map[string]interface{}, error) {
	req, params, err := request.ParseRequest(args, func() *proto.InferSessionModeRequest {
		return &proto.InferSessionModeRequest{}
	})
	if err != nil {
		return nil, nil, err
	}

	if req != nil {
		return req, nil, nil
	}

	return nil, params, nil
}

// ─── InferSessionModeRequestToParams ────────────────────────────────────────
// InferSessionModeRequestToParams converts a protobuf InferSessionModeRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func InferSessionModeRequestToParams(req *proto.InferSessionModeRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := request.ProtobufToParams(req, &request.ProtobufToParamsOptions{
		FilterEmptyStrings: true,
		StringifyArrays:    false,
	})
	if err != nil {
		return make(map[string]interface{})
	}

	return params
}

// ─── ParseToolCatalogRequest ────────────────────────────────────────────────
// ParseToolCatalogRequest parses a tool_catalog request (protobuf or JSON).
func ParseToolCatalogRequest(args json.RawMessage) (*proto.ToolCatalogRequest, map[string]interface{}, error) {
	req, params, err := request.ParseRequest(args, func() *proto.ToolCatalogRequest {
		return &proto.ToolCatalogRequest{}
	})
	if err != nil {
		return nil, nil, err
	}

	if req != nil {
		return req, nil, nil
	}

	return nil, params, nil
}

// ─── ToolCatalogRequestToParams ─────────────────────────────────────────────
// ToolCatalogRequestToParams converts a protobuf ToolCatalogRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func ToolCatalogRequestToParams(req *proto.ToolCatalogRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := request.ProtobufToParams(req, &request.ProtobufToParamsOptions{
		FilterEmptyStrings: true,
		StringifyArrays:    false,
	})
	if err != nil {
		return make(map[string]interface{})
	}

	return params
}

// ─── ParseWorkflowModeRequest ───────────────────────────────────────────────
// ParseWorkflowModeRequest parses a workflow_mode request (protobuf or JSON).
func ParseWorkflowModeRequest(args json.RawMessage) (*proto.WorkflowModeRequest, map[string]interface{}, error) {
	req, params, err := request.ParseRequest(args, func() *proto.WorkflowModeRequest {
		return &proto.WorkflowModeRequest{}
	})
	if err != nil {
		return nil, nil, err
	}

	if req != nil {
		return req, nil, nil
	}

	return nil, params, nil
}

// ─── WorkflowModeRequestToParams ────────────────────────────────────────────
// WorkflowModeRequestToParams converts a protobuf WorkflowModeRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func WorkflowModeRequestToParams(req *proto.WorkflowModeRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := request.ProtobufToParams(req, &request.ProtobufToParamsOptions{
		FilterEmptyStrings: true,
		StringifyArrays:    false,
	})
	if err != nil {
		return make(map[string]interface{})
	}

	return params
}

// ─── ParseEstimationRequest ─────────────────────────────────────────────────
// ParseEstimationRequest parses an estimation request (protobuf or JSON).
func ParseEstimationRequest(args json.RawMessage) (*proto.EstimationRequest, map[string]interface{}, error) {
	req, params, err := request.ParseRequest(args, func() *proto.EstimationRequest {
		return &proto.EstimationRequest{}
	})
	if err != nil {
		return nil, nil, err
	}

	if req != nil {
		return req, nil, nil
	}

	return nil, params, nil
}

// ─── EstimationRequestToParams ──────────────────────────────────────────────
// EstimationRequestToParams converts a protobuf EstimationRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func EstimationRequestToParams(req *proto.EstimationRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := request.ProtobufToParams(req, &request.ProtobufToParamsOptions{
		FilterEmptyStrings:  true,
		StringifyArrays:     true,  // TagList is an array
		ConvertFloat64ToInt: false, // Keep mlx_weight as float64 (it's a weight, not a count)
	})
	if err != nil {
		return make(map[string]interface{})
	}

	// Ensure local_ai_backend is in params when set (proto field 11)
	if b := req.GetLocalAiBackend(); b != "" {
		params["local_ai_backend"] = b
	}
	return params
}

// ─── ParseSessionRequest ────────────────────────────────────────────────────
// ParseSessionRequest parses a session request (protobuf or JSON).
func ParseSessionRequest(args json.RawMessage) (*proto.SessionRequest, map[string]interface{}, error) {
	req, params, err := request.ParseRequest(args, func() *proto.SessionRequest {
		return &proto.SessionRequest{}
	})
	if err != nil {
		return nil, nil, err
	}

	if req != nil {
		return req, nil, nil
	}

	return nil, params, nil
}

// ─── SessionRequestToParams ─────────────────────────────────────────────────
// SessionRequestToParams converts a protobuf SessionRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func SessionRequestToParams(req *proto.SessionRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := request.ProtobufToParams(req, &request.ProtobufToParamsOptions{
		FilterEmptyStrings:  true,
		StringifyArrays:     false,
		ConvertFloat64ToInt: true,
		Float64ToIntFields:  []string{"limit"},
	})
	if err != nil {
		return make(map[string]interface{})
	}

	return params
}

// ─── ParseGitToolsRequest ───────────────────────────────────────────────────
// ParseGitToolsRequest parses a git_tools request (protobuf or JSON).
func ParseGitToolsRequest(args json.RawMessage) (*proto.GitToolsRequest, map[string]interface{}, error) {
	req, params, err := request.ParseRequest(args, func() *proto.GitToolsRequest {
		return &proto.GitToolsRequest{}
	})
	if err != nil {
		return nil, nil, err
	}

	if req != nil {
		return req, nil, nil
	}

	return nil, params, nil
}

// ─── GitToolsRequestToParams ────────────────────────────────────────────────
// GitToolsRequestToParams converts a protobuf GitToolsRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func GitToolsRequestToParams(req *proto.GitToolsRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := request.ProtobufToParams(req, &request.ProtobufToParamsOptions{
		FilterEmptyStrings:  true,
		StringifyArrays:     false,
		ConvertFloat64ToInt: true,
		Float64ToIntFields:  []string{"limit", "max_commits"},
	})
	if err != nil {
		return make(map[string]interface{})
	}

	return params
}

// GitToolsResponseToMap converts GitToolsResponse proto to map for response.FormatResult (unmarshals result_json).
