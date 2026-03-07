// protobuf_helpers_tools.go — Protobuf helpers for WrapHandler-based tool handlers.
// Implements ParseXxx / XxxRequestToParams pairs for all tools using the WrapHandler pattern.
// See also: protobuf_helpers.go, protobuf_helpers_report.go
package tools

import (
	"encoding/json"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/proto"
)

// ─── Contents ───────────────────────────────────────────────────────────────
//   ParseAnalyzeAlignmentRequest — parses analyze_alignment tool request (protobuf or JSON)
//   AnalyzeAlignmentRequestToParams — converts proto to params map
//   ParseSetupHooksRequest — parses setup_hooks tool request (protobuf or JSON)
//   SetupHooksRequestToParams — converts proto to params map
//   ParseCheckAttributionRequest — parses check_attribution tool request (protobuf or JSON)
//   CheckAttributionRequestToParams — converts proto to params map
//   ParseAddExternalToolHintsRequest — parses add_external_tool_hints request (protobuf or JSON)
//   AddExternalToolHintsRequestToParams — converts proto to params map
//   ParseTaskAnalysisRequest — parses task_analysis tool request (protobuf or JSON)
//   TaskAnalysisRequestToParams — converts proto to params map
//   ParseTaskDiscoveryRequest — parses task_discovery tool request (protobuf or JSON)
//   TaskDiscoveryRequestToParams — converts proto to params map
// ────────────────────────────────────────────────────────────────────────────

// ─── ParseAnalyzeAlignmentRequest ───────────────────────────────────────────
// ParseAnalyzeAlignmentRequest parses an analyze_alignment request (protobuf or JSON).
// Returns (*proto.AnalyzeAlignmentRequest, nil, nil) on proto success,
// or (nil, params, nil) on JSON fallback.
func ParseAnalyzeAlignmentRequest(args json.RawMessage) (*proto.AnalyzeAlignmentRequest, map[string]interface{}, error) {
	return framework.ParseRequest(args, func() *proto.AnalyzeAlignmentRequest {
		return &proto.AnalyzeAlignmentRequest{}
	})
}

// ─── AnalyzeAlignmentRequestToParams ────────────────────────────────────────
// AnalyzeAlignmentRequestToParams converts a protobuf AnalyzeAlignmentRequest to params map.
func AnalyzeAlignmentRequestToParams(req *proto.AnalyzeAlignmentRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}
	params, err := framework.ProtobufToParams(req, &framework.ProtobufToParamsOptions{
		FilterEmptyStrings: true,
	})
	if err != nil {
		return make(map[string]interface{})
	}
	return params
}

// ─── ParseSetupHooksRequest ──────────────────────────────────────────────────
// ParseSetupHooksRequest parses a setup_hooks request (protobuf or JSON).
func ParseSetupHooksRequest(args json.RawMessage) (*proto.SetupHooksRequest, map[string]interface{}, error) {
	return framework.ParseRequest(args, func() *proto.SetupHooksRequest {
		return &proto.SetupHooksRequest{}
	})
}

// ─── SetupHooksRequestToParams ───────────────────────────────────────────────
// SetupHooksRequestToParams converts a protobuf SetupHooksRequest to params map.
func SetupHooksRequestToParams(req *proto.SetupHooksRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}
	params, err := framework.ProtobufToParams(req, &framework.ProtobufToParamsOptions{
		FilterEmptyStrings: true,
	})
	if err != nil {
		return make(map[string]interface{})
	}
	return params
}

// ─── ParseCheckAttributionRequest ────────────────────────────────────────────
// ParseCheckAttributionRequest parses a check_attribution request (protobuf or JSON).
func ParseCheckAttributionRequest(args json.RawMessage) (*proto.CheckAttributionRequest, map[string]interface{}, error) {
	return framework.ParseRequest(args, func() *proto.CheckAttributionRequest {
		return &proto.CheckAttributionRequest{}
	})
}

// ─── CheckAttributionRequestToParams ─────────────────────────────────────────
// CheckAttributionRequestToParams converts a protobuf CheckAttributionRequest to params map.
func CheckAttributionRequestToParams(req *proto.CheckAttributionRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}
	params, err := framework.ProtobufToParams(req, &framework.ProtobufToParamsOptions{
		FilterEmptyStrings: true,
	})
	if err != nil {
		return make(map[string]interface{})
	}
	return params
}

// ─── ParseAddExternalToolHintsRequest ────────────────────────────────────────
// ParseAddExternalToolHintsRequest parses an add_external_tool_hints request (protobuf or JSON).
func ParseAddExternalToolHintsRequest(args json.RawMessage) (*proto.AddExternalToolHintsRequest, map[string]interface{}, error) {
	return framework.ParseRequest(args, func() *proto.AddExternalToolHintsRequest {
		return &proto.AddExternalToolHintsRequest{}
	})
}

// ─── AddExternalToolHintsRequestToParams ─────────────────────────────────────
// AddExternalToolHintsRequestToParams converts a protobuf AddExternalToolHintsRequest to params map.
func AddExternalToolHintsRequestToParams(req *proto.AddExternalToolHintsRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}
	params, err := framework.ProtobufToParams(req, &framework.ProtobufToParamsOptions{
		FilterEmptyStrings: true,
	})
	if err != nil {
		return make(map[string]interface{})
	}
	return params
}

// ─── ParseTaskAnalysisRequest ─────────────────────────────────────────────────
// ParseTaskAnalysisRequest parses a task_analysis request (protobuf or JSON).
func ParseTaskAnalysisRequest(args json.RawMessage) (*proto.TaskAnalysisRequest, map[string]interface{}, error) {
	return framework.ParseRequest(args, func() *proto.TaskAnalysisRequest {
		return &proto.TaskAnalysisRequest{}
	})
}

// ─── TaskAnalysisRequestToParams ──────────────────────────────────────────────
// TaskAnalysisRequestToParams converts a protobuf TaskAnalysisRequest to params map.
func TaskAnalysisRequestToParams(req *proto.TaskAnalysisRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}
	params, err := framework.ProtobufToParams(req, &framework.ProtobufToParamsOptions{
		FilterEmptyStrings:  true,
		ConvertFloat64ToInt: true,
		Float64ToIntFields:  []string{"limit", "llm_batch_size"},
	})
	if err != nil {
		return make(map[string]interface{})
	}
	return params
}

// ─── ParseTaskDiscoveryRequest ────────────────────────────────────────────────
// ParseTaskDiscoveryRequest parses a task_discovery request (protobuf or JSON).
// Returns (*proto.TaskDiscoveryRequest, nil, nil) on proto success,
// or (nil, params, nil) on JSON fallback.
func ParseTaskDiscoveryRequest(args json.RawMessage) (*proto.TaskDiscoveryRequest, map[string]interface{}, error) {
	return framework.ParseRequest(args, func() *proto.TaskDiscoveryRequest {
		return &proto.TaskDiscoveryRequest{}
	})
}

// ─── TaskDiscoveryRequestToParams ────────────────────────────────────────────
// TaskDiscoveryRequestToParams converts a protobuf TaskDiscoveryRequest to params map.
func TaskDiscoveryRequestToParams(req *proto.TaskDiscoveryRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}
	params, err := framework.ProtobufToParams(req, &framework.ProtobufToParamsOptions{
		FilterEmptyStrings: true,
	})
	if err != nil {
		return make(map[string]interface{})
	}
	return params
}

// ─── ParseGenerateConfigRequest ──────────────────────────────────────────────
// ParseGenerateConfigRequest parses a generate_config request (protobuf or JSON).
func ParseGenerateConfigRequest(args json.RawMessage) (*proto.GenerateConfigRequest, map[string]interface{}, error) {
	return framework.ParseRequest(args, func() *proto.GenerateConfigRequest {
		return &proto.GenerateConfigRequest{}
	})
}

// ─── GenerateConfigRequestToParams ───────────────────────────────────────────
func GenerateConfigRequestToParams(req *proto.GenerateConfigRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}
	params, err := framework.ProtobufToParams(req, &framework.ProtobufToParamsOptions{
		FilterEmptyStrings: true,
	})
	if err != nil {
		return make(map[string]interface{})
	}
	return params
}

// ─── ParseMemoryMaintRequest ─────────────────────────────────────────────────
// ParseMemoryMaintRequest parses a memory_maint request (protobuf or JSON).
func ParseMemoryMaintRequest(args json.RawMessage) (*proto.MemoryMaintRequest, map[string]interface{}, error) {
	return framework.ParseRequest(args, func() *proto.MemoryMaintRequest {
		return &proto.MemoryMaintRequest{}
	})
}

// ─── MemoryMaintRequestToParams ──────────────────────────────────────────────
func MemoryMaintRequestToParams(req *proto.MemoryMaintRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}
	params, err := framework.ProtobufToParams(req, &framework.ProtobufToParamsOptions{
		FilterEmptyStrings: true,
	})
	if err != nil {
		return make(map[string]interface{})
	}
	return params
}

// ─── ParseTestingRequest ─────────────────────────────────────────────────────
// ParseTestingRequest parses a testing request (protobuf or JSON).
func ParseTestingRequest(args json.RawMessage) (*proto.TestingRequest, map[string]interface{}, error) {
	return framework.ParseRequest(args, func() *proto.TestingRequest {
		return &proto.TestingRequest{}
	})
}

// ─── TestingRequestToParams ───────────────────────────────────────────────────
func TestingRequestToParams(req *proto.TestingRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}
	params, err := framework.ProtobufToParams(req, &framework.ProtobufToParamsOptions{
		FilterEmptyStrings: true,
	})
	if err != nil {
		return make(map[string]interface{})
	}
	return params
}

// ─── ParseAutomationRequest ───────────────────────────────────────────────────
// ParseAutomationRequest parses an automation request (protobuf or JSON).
func ParseAutomationRequest(args json.RawMessage) (*proto.AutomationRequest, map[string]interface{}, error) {
	return framework.ParseRequest(args, func() *proto.AutomationRequest {
		return &proto.AutomationRequest{}
	})
}

// ─── AutomationRequestToParams ────────────────────────────────────────────────
func AutomationRequestToParams(req *proto.AutomationRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}
	params, err := framework.ProtobufToParams(req, &framework.ProtobufToParamsOptions{
		FilterEmptyStrings: true,
	})
	if err != nil {
		return make(map[string]interface{})
	}
	return params
}

// ─── ParseLintRequest ────────────────────────────────────────────────────────
// ParseLintRequest parses a lint request (protobuf or JSON).
func ParseLintRequest(args json.RawMessage) (*proto.LintRequest, map[string]interface{}, error) {
	return framework.ParseRequest(args, func() *proto.LintRequest {
		return &proto.LintRequest{}
	})
}

// ─── LintRequestToParams ─────────────────────────────────────────────────────
func LintRequestToParams(req *proto.LintRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}
	params, err := framework.ProtobufToParams(req, &framework.ProtobufToParamsOptions{
		FilterEmptyStrings: true,
	})
	if err != nil {
		return make(map[string]interface{})
	}
	return params
}

// ─── ParseOllamaRequest ──────────────────────────────────────────────────────
// ParseOllamaRequest parses an ollama request (protobuf or JSON).
func ParseOllamaRequest(args json.RawMessage) (*proto.OllamaRequest, map[string]interface{}, error) {
	return framework.ParseRequest(args, func() *proto.OllamaRequest {
		return &proto.OllamaRequest{}
	})
}

// ─── OllamaRequestToParams ───────────────────────────────────────────────────
func OllamaRequestToParams(req *proto.OllamaRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}
	params, err := framework.ProtobufToParams(req, &framework.ProtobufToParamsOptions{
		FilterEmptyStrings: true,
	})
	if err != nil {
		return make(map[string]interface{})
	}
	return params
}

// ─── ParseMlxRequest ─────────────────────────────────────────────────────────
// ParseMlxRequest parses an mlx request (JSON-only; no dedicated proto type — reuses OllamaRequest shape).
func ParseMlxRequest(args json.RawMessage) (*proto.OllamaRequest, map[string]interface{}, error) {
	return framework.ParseRequest(args, func() *proto.OllamaRequest {
		return &proto.OllamaRequest{}
	})
}

// ─── MlxRequestToParams ──────────────────────────────────────────────────────
func MlxRequestToParams(req *proto.OllamaRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}
	params, err := framework.ProtobufToParams(req, &framework.ProtobufToParamsOptions{
		FilterEmptyStrings: true,
	})
	if err != nil {
		return make(map[string]interface{})
	}
	return params
}

// ─── ParsePromptTrackingRequest ──────────────────────────────────────────────
// ParsePromptTrackingRequest parses a prompt_tracking request (protobuf or JSON).
func ParsePromptTrackingRequest(args json.RawMessage) (*proto.PromptTrackingRequest, map[string]interface{}, error) {
	return framework.ParseRequest(args, func() *proto.PromptTrackingRequest {
		return &proto.PromptTrackingRequest{}
	})
}

// ─── PromptTrackingRequestToParams ───────────────────────────────────────────
func PromptTrackingRequestToParams(req *proto.PromptTrackingRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}
	params, err := framework.ProtobufToParams(req, &framework.ProtobufToParamsOptions{
		FilterEmptyStrings: true,
	})
	if err != nil {
		return make(map[string]interface{})
	}
	return params
}

// ─── ParseRecommendRequest ───────────────────────────────────────────────────
// ParseRecommendRequest parses a recommend request (protobuf or JSON).
func ParseRecommendRequest(args json.RawMessage) (*proto.RecommendRequest, map[string]interface{}, error) {
	return framework.ParseRequest(args, func() *proto.RecommendRequest {
		return &proto.RecommendRequest{}
	})
}

// ─── RecommendRequestToParams ────────────────────────────────────────────────
func RecommendRequestToParams(req *proto.RecommendRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}
	params, err := framework.ProtobufToParams(req, &framework.ProtobufToParamsOptions{
		FilterEmptyStrings: true,
	})
	if err != nil {
		return make(map[string]interface{})
	}
	return params
}
