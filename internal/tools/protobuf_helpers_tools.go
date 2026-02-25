// protobuf_helpers_tools.go — Protobuf helpers: GitTools, Testing, MemoryMaint, TaskAnalysis, and Ollama conversions.
// See also: protobuf_helpers.go, protobuf_helpers_report.go
package tools

import (
	"encoding/json"
	"github.com/davidl71/exarp-go/proto"
	"github.com/davidl71/mcp-go-core/pkg/mcp/request"
)

// ─── Contents ───────────────────────────────────────────────────────────────
//   GitToolsResponseToMap
//   TestingResponseToMap — TestingResponseToMap converts TestingResponse proto to map for response.FormatResult (unmarshals result_json).
//   ParseMemoryMaintRequest — ParseMemoryMaintRequest parses a memory_maint request (protobuf or JSON).
//   MemoryMaintRequestToParams — MemoryMaintRequestToParams converts a protobuf MemoryMaintRequest to params map
//   ParseTaskAnalysisRequest — ParseTaskAnalysisRequest parses a task_analysis request (protobuf or JSON).
//   TaskAnalysisRequestToParams — TaskAnalysisRequestToParams converts a protobuf TaskAnalysisRequest to params map
//   ParseTaskDiscoveryRequest — ParseTaskDiscoveryRequest parses a task_discovery request (protobuf or JSON).
//   TaskDiscoveryRequestToParams — TaskDiscoveryRequestToParams converts a protobuf TaskDiscoveryRequest to params map
//   ParseOllamaRequest — ParseOllamaRequest parses an ollama request (protobuf or JSON).
//   OllamaRequestToParams — OllamaRequestToParams converts a protobuf OllamaRequest to params map
//   ParseMlxRequest — ParseMlxRequest parses an mlx request (protobuf or JSON).
//   MlxRequestToParams — MlxRequestToParams converts a protobuf MlxRequest to params map
//   ParsePromptTrackingRequest — ParsePromptTrackingRequest parses a prompt_tracking request (protobuf or JSON).
//   PromptTrackingRequestToParams — PromptTrackingRequestToParams converts a protobuf PromptTrackingRequest to params map
//   ParseRecommendRequest — ParseRecommendRequest parses a recommend request (protobuf or JSON).
//   RecommendRequestToParams — RecommendRequestToParams converts a protobuf RecommendRequest to params map
//   ParseAnalyzeAlignmentRequest — ParseAnalyzeAlignmentRequest parses an analyze_alignment request (protobuf or JSON).
//   AnalyzeAlignmentRequestToParams — AnalyzeAlignmentRequestToParams converts a protobuf AnalyzeAlignmentRequest to params map
//   ParseGenerateConfigRequest — ParseGenerateConfigRequest parses a generate_config request (protobuf or JSON).
//   GenerateConfigRequestToParams — GenerateConfigRequestToParams converts a protobuf GenerateConfigRequest to params map
//   ParseSetupHooksRequest — ParseSetupHooksRequest parses a setup_hooks request (protobuf or JSON).
//   SetupHooksRequestToParams — SetupHooksRequestToParams converts a protobuf SetupHooksRequest to params map
//   ParseCheckAttributionRequest — ParseCheckAttributionRequest parses a check_attribution request (protobuf or JSON).
//   CheckAttributionRequestToParams — CheckAttributionRequestToParams converts a protobuf CheckAttributionRequest to params map
//   ParseAddExternalToolHintsRequest — ParseAddExternalToolHintsRequest parses an add_external_tool_hints request (protobuf or JSON).
//   AddExternalToolHintsRequestToParams — AddExternalToolHintsRequestToParams converts a protobuf AddExternalToolHintsRequest to params map
//   ParseTestingRequest — ParseTestingRequest parses a testing tool request (protobuf or JSON).
//   TestingRequestToParams — TestingRequestToParams converts a protobuf TestingRequest to params map
//   ParseAutomationRequest — ParseAutomationRequest parses an automation tool request (protobuf or JSON).
//   AutomationRequestToParams — AutomationRequestToParams converts a protobuf AutomationRequest to params map
//   ParseLintRequest — ParseLintRequest parses a lint tool request (protobuf or JSON).
//   LintRequestToParams — LintRequestToParams converts a protobuf LintRequest to params map
// ────────────────────────────────────────────────────────────────────────────

// ─── GitToolsResponseToMap ──────────────────────────────────────────────────
func GitToolsResponseToMap(resp *proto.GitToolsResponse) map[string]interface{} {
	if resp == nil {
		return nil
	}

	out := map[string]interface{}{
		"success": resp.GetSuccess(),
		"action":  resp.GetAction(),
	}

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

// ─── TestingResponseToMap ───────────────────────────────────────────────────
// TestingResponseToMap converts TestingResponse proto to map for response.FormatResult (unmarshals result_json).
func TestingResponseToMap(resp *proto.TestingResponse) map[string]interface{} {
	if resp == nil {
		return nil
	}

	out := map[string]interface{}{
		"success": resp.GetSuccess(),
		"action":  resp.GetAction(),
	}

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

// ─── ParseMemoryMaintRequest ────────────────────────────────────────────────
// ParseMemoryMaintRequest parses a memory_maint request (protobuf or JSON).
func ParseMemoryMaintRequest(args json.RawMessage) (*proto.MemoryMaintRequest, map[string]interface{}, error) {
	req, params, err := request.ParseRequest(args, func() *proto.MemoryMaintRequest {
		return &proto.MemoryMaintRequest{}
	})
	if err != nil {
		return nil, nil, err
	}

	if req != nil {
		return req, nil, nil
	}

	return nil, params, nil
}

// ─── MemoryMaintRequestToParams ─────────────────────────────────────────────
// MemoryMaintRequestToParams converts a protobuf MemoryMaintRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func MemoryMaintRequestToParams(req *proto.MemoryMaintRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := request.ProtobufToParams(req, &request.ProtobufToParamsOptions{
		FilterEmptyStrings:  true,
		StringifyArrays:     false,
		ConvertFloat64ToInt: true,
		Float64ToIntFields:  []string{"max_age_days", "scorecard_max_age_days", "keep_minimum"},
		// Note: value_threshold and similarity_threshold remain as float64 (they're thresholds, not counts)
	})
	if err != nil {
		return make(map[string]interface{})
	}

	return params
}

// ─── ParseTaskAnalysisRequest ───────────────────────────────────────────────
// ParseTaskAnalysisRequest parses a task_analysis request (protobuf or JSON).
func ParseTaskAnalysisRequest(args json.RawMessage) (*proto.TaskAnalysisRequest, map[string]interface{}, error) {
	req, params, err := request.ParseRequest(args, func() *proto.TaskAnalysisRequest {
		return &proto.TaskAnalysisRequest{}
	})
	if err != nil {
		return nil, nil, err
	}

	if req != nil {
		return req, nil, nil
	}

	return nil, params, nil
}

// ─── TaskAnalysisRequestToParams ────────────────────────────────────────────
// TaskAnalysisRequestToParams converts a protobuf TaskAnalysisRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func TaskAnalysisRequestToParams(req *proto.TaskAnalysisRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := request.ProtobufToParams(req, &request.ProtobufToParamsOptions{
		FilterEmptyStrings:  true,
		StringifyArrays:     false,
		ConvertFloat64ToInt: false, // Keep similarity_threshold as float64 (it's a threshold, not a count)
	})
	if err != nil {
		return make(map[string]interface{})
	}

	return params
}

// ─── ParseTaskDiscoveryRequest ──────────────────────────────────────────────
// ParseTaskDiscoveryRequest parses a task_discovery request (protobuf or JSON).
func ParseTaskDiscoveryRequest(args json.RawMessage) (*proto.TaskDiscoveryRequest, map[string]interface{}, error) {
	req, params, err := request.ParseRequest(args, func() *proto.TaskDiscoveryRequest {
		return &proto.TaskDiscoveryRequest{}
	})
	if err != nil {
		return nil, nil, err
	}

	if req != nil {
		return req, nil, nil
	}

	return nil, params, nil
}

// ─── TaskDiscoveryRequestToParams ───────────────────────────────────────────
// TaskDiscoveryRequestToParams converts a protobuf TaskDiscoveryRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func TaskDiscoveryRequestToParams(req *proto.TaskDiscoveryRequest) map[string]interface{} {
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

// ─── ParseOllamaRequest ─────────────────────────────────────────────────────
// ParseOllamaRequest parses an ollama request (protobuf or JSON).
func ParseOllamaRequest(args json.RawMessage) (*proto.OllamaRequest, map[string]interface{}, error) {
	req, params, err := request.ParseRequest(args, func() *proto.OllamaRequest {
		return &proto.OllamaRequest{}
	})
	if err != nil {
		return nil, nil, err
	}

	if req != nil {
		return req, nil, nil
	}

	return nil, params, nil
}

// ─── OllamaRequestToParams ──────────────────────────────────────────────────
// OllamaRequestToParams converts a protobuf OllamaRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func OllamaRequestToParams(req *proto.OllamaRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := request.ProtobufToParams(req, &request.ProtobufToParamsOptions{
		FilterEmptyStrings:  true,
		StringifyArrays:     false,
		ConvertFloat64ToInt: true,
		Float64ToIntFields:  []string{"num_gpu", "num_threads", "context_size"},
	})
	if err != nil {
		return make(map[string]interface{})
	}

	return params
}

// ─── ParseMlxRequest ────────────────────────────────────────────────────────
// ParseMlxRequest parses an mlx request (protobuf or JSON).
func ParseMlxRequest(args json.RawMessage) (*proto.MLXRequest, map[string]interface{}, error) {
	req, params, err := request.ParseRequest(args, func() *proto.MLXRequest {
		return &proto.MLXRequest{}
	})
	if err != nil {
		return nil, nil, err
	}

	if req != nil {
		return req, nil, nil
	}

	return nil, params, nil
}

// ─── MlxRequestToParams ─────────────────────────────────────────────────────
// MlxRequestToParams converts a protobuf MlxRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func MlxRequestToParams(req *proto.MLXRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := request.ProtobufToParams(req, &request.ProtobufToParamsOptions{
		FilterEmptyStrings:  true,
		StringifyArrays:     false,
		ConvertFloat64ToInt: true,
		Float64ToIntFields:  []string{"max_tokens"},
		// Note: temperature remains as float64 (it's a threshold, not a count)
	})
	if err != nil {
		return make(map[string]interface{})
	}

	return params
}

// ─── ParsePromptTrackingRequest ─────────────────────────────────────────────
// ParsePromptTrackingRequest parses a prompt_tracking request (protobuf or JSON).
func ParsePromptTrackingRequest(args json.RawMessage) (*proto.PromptTrackingRequest, map[string]interface{}, error) {
	req, params, err := request.ParseRequest(args, func() *proto.PromptTrackingRequest {
		return &proto.PromptTrackingRequest{}
	})
	if err != nil {
		return nil, nil, err
	}

	if req != nil {
		return req, nil, nil
	}

	return nil, params, nil
}

// ─── PromptTrackingRequestToParams ──────────────────────────────────────────
// PromptTrackingRequestToParams converts a protobuf PromptTrackingRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func PromptTrackingRequestToParams(req *proto.PromptTrackingRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := request.ProtobufToParams(req, &request.ProtobufToParamsOptions{
		FilterEmptyStrings:  true,
		StringifyArrays:     false,
		ConvertFloat64ToInt: true,
		Float64ToIntFields:  []string{"iteration", "days"},
	})
	if err != nil {
		return make(map[string]interface{})
	}

	return params
}

// ─── ParseRecommendRequest ──────────────────────────────────────────────────
// ParseRecommendRequest parses a recommend request (protobuf or JSON).
func ParseRecommendRequest(args json.RawMessage) (*proto.RecommendRequest, map[string]interface{}, error) {
	req, params, err := request.ParseRequest(args, func() *proto.RecommendRequest {
		return &proto.RecommendRequest{}
	})
	if err != nil {
		return nil, nil, err
	}

	if req != nil {
		return req, nil, nil
	}

	return nil, params, nil
}

// ─── RecommendRequestToParams ───────────────────────────────────────────────
// RecommendRequestToParams converts a protobuf RecommendRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func RecommendRequestToParams(req *proto.RecommendRequest) map[string]interface{} {
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

// ─── ParseAnalyzeAlignmentRequest ───────────────────────────────────────────
// ParseAnalyzeAlignmentRequest parses an analyze_alignment request (protobuf or JSON).
func ParseAnalyzeAlignmentRequest(args json.RawMessage) (*proto.AnalyzeAlignmentRequest, map[string]interface{}, error) {
	req, params, err := request.ParseRequest(args, func() *proto.AnalyzeAlignmentRequest {
		return &proto.AnalyzeAlignmentRequest{}
	})
	if err != nil {
		return nil, nil, err
	}

	if req != nil {
		return req, nil, nil
	}

	return nil, params, nil
}

// ─── AnalyzeAlignmentRequestToParams ────────────────────────────────────────
// AnalyzeAlignmentRequestToParams converts a protobuf AnalyzeAlignmentRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func AnalyzeAlignmentRequestToParams(req *proto.AnalyzeAlignmentRequest) map[string]interface{} {
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

// ─── ParseGenerateConfigRequest ─────────────────────────────────────────────
// ParseGenerateConfigRequest parses a generate_config request (protobuf or JSON).
func ParseGenerateConfigRequest(args json.RawMessage) (*proto.GenerateConfigRequest, map[string]interface{}, error) {
	req, params, err := request.ParseRequest(args, func() *proto.GenerateConfigRequest {
		return &proto.GenerateConfigRequest{}
	})
	if err != nil {
		return nil, nil, err
	}

	if req != nil {
		return req, nil, nil
	}

	return nil, params, nil
}

// ─── GenerateConfigRequestToParams ──────────────────────────────────────────
// GenerateConfigRequestToParams converts a protobuf GenerateConfigRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func GenerateConfigRequestToParams(req *proto.GenerateConfigRequest) map[string]interface{} {
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

// ─── ParseSetupHooksRequest ─────────────────────────────────────────────────
// ParseSetupHooksRequest parses a setup_hooks request (protobuf or JSON).
func ParseSetupHooksRequest(args json.RawMessage) (*proto.SetupHooksRequest, map[string]interface{}, error) {
	req, params, err := request.ParseRequest(args, func() *proto.SetupHooksRequest {
		return &proto.SetupHooksRequest{}
	})
	if err != nil {
		return nil, nil, err
	}

	if req != nil {
		return req, nil, nil
	}

	return nil, params, nil
}

// ─── SetupHooksRequestToParams ──────────────────────────────────────────────
// SetupHooksRequestToParams converts a protobuf SetupHooksRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func SetupHooksRequestToParams(req *proto.SetupHooksRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := request.ProtobufToParams(req, &request.ProtobufToParamsOptions{
		FilterEmptyStrings: true,
		StringifyArrays:    true, // Hooks is an array
	})
	if err != nil {
		return make(map[string]interface{})
	}

	return params
}

// ─── ParseCheckAttributionRequest ───────────────────────────────────────────
// ParseCheckAttributionRequest parses a check_attribution request (protobuf or JSON).
func ParseCheckAttributionRequest(args json.RawMessage) (*proto.CheckAttributionRequest, map[string]interface{}, error) {
	req, params, err := request.ParseRequest(args, func() *proto.CheckAttributionRequest {
		return &proto.CheckAttributionRequest{}
	})
	if err != nil {
		return nil, nil, err
	}

	if req != nil {
		return req, nil, nil
	}

	return nil, params, nil
}

// ─── CheckAttributionRequestToParams ────────────────────────────────────────
// CheckAttributionRequestToParams converts a protobuf CheckAttributionRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func CheckAttributionRequestToParams(req *proto.CheckAttributionRequest) map[string]interface{} {
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

// ─── ParseAddExternalToolHintsRequest ───────────────────────────────────────
// ParseAddExternalToolHintsRequest parses an add_external_tool_hints request (protobuf or JSON).
func ParseAddExternalToolHintsRequest(args json.RawMessage) (*proto.AddExternalToolHintsRequest, map[string]interface{}, error) {
	req, params, err := request.ParseRequest(args, func() *proto.AddExternalToolHintsRequest {
		return &proto.AddExternalToolHintsRequest{}
	})
	if err != nil {
		return nil, nil, err
	}

	if req != nil {
		return req, nil, nil
	}

	return nil, params, nil
}

// ─── AddExternalToolHintsRequestToParams ────────────────────────────────────
// AddExternalToolHintsRequestToParams converts a protobuf AddExternalToolHintsRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func AddExternalToolHintsRequestToParams(req *proto.AddExternalToolHintsRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := request.ProtobufToParams(req, &request.ProtobufToParamsOptions{
		FilterEmptyStrings:  true,
		StringifyArrays:     false,
		ConvertFloat64ToInt: true,
		Float64ToIntFields:  []string{"min_file_size"},
	})
	if err != nil {
		return make(map[string]interface{})
	}

	return params
}

// ─── ParseTestingRequest ────────────────────────────────────────────────────
// ParseTestingRequest parses a testing tool request (protobuf or JSON).
func ParseTestingRequest(args json.RawMessage) (*proto.TestingRequest, map[string]interface{}, error) {
	req, params, err := request.ParseRequest(args, func() *proto.TestingRequest {
		return &proto.TestingRequest{}
	})
	if err != nil {
		return nil, nil, err
	}

	if req != nil {
		return req, nil, nil
	}

	return nil, params, nil
}

// ─── TestingRequestToParams ─────────────────────────────────────────────────
// TestingRequestToParams converts a protobuf TestingRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func TestingRequestToParams(req *proto.TestingRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := request.ProtobufToParams(req, &request.ProtobufToParamsOptions{
		FilterEmptyStrings:  true,
		StringifyArrays:     false,
		ConvertFloat64ToInt: true,
		Float64ToIntFields:  []string{"min_coverage"},
		// Note: min_confidence remains as float64 (it's a threshold, not a count)
	})
	if err != nil {
		return make(map[string]interface{})
	}

	return params
}

// ─── ParseAutomationRequest ─────────────────────────────────────────────────
// ParseAutomationRequest parses an automation tool request (protobuf or JSON).
func ParseAutomationRequest(args json.RawMessage) (*proto.AutomationRequest, map[string]interface{}, error) {
	req, params, err := request.ParseRequest(args, func() *proto.AutomationRequest {
		return &proto.AutomationRequest{}
	})
	if err != nil {
		return nil, nil, err
	}

	if req != nil {
		return req, nil, nil
	}

	return nil, params, nil
}

// ─── AutomationRequestToParams ──────────────────────────────────────────────
// AutomationRequestToParams converts a protobuf AutomationRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func AutomationRequestToParams(req *proto.AutomationRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := request.ProtobufToParams(req, &request.ProtobufToParamsOptions{
		FilterEmptyStrings:  true,
		StringifyArrays:     true, // Tasks and TagFilter are arrays
		ConvertFloat64ToInt: true,
		Float64ToIntFields:  []string{"max_tasks_per_host", "max_parallel_tasks", "max_iterations"},
		// Note: min_value_score remains as float64 (it's a threshold, not a count)
	})
	if err != nil {
		return make(map[string]interface{})
	}

	return params
}

// ─── ParseLintRequest ───────────────────────────────────────────────────────
// ParseLintRequest parses a lint tool request (protobuf or JSON).
func ParseLintRequest(args json.RawMessage) (*proto.LintRequest, map[string]interface{}, error) {
	req, params, err := request.ParseRequest(args, func() *proto.LintRequest {
		return &proto.LintRequest{}
	})
	if err != nil {
		return nil, nil, err
	}

	if req != nil {
		return req, nil, nil
	}

	return nil, params, nil
}

// ─── LintRequestToParams ────────────────────────────────────────────────────
// LintRequestToParams converts a protobuf LintRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func LintRequestToParams(req *proto.LintRequest) map[string]interface{} {
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
