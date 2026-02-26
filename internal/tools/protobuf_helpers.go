// protobuf_helpers.go — Protobuf helpers: Memory, Context, Report, Health, and Metrics conversions.
// See also: protobuf_helpers_report.go, protobuf_helpers_tools.go
package tools

import (
	"encoding/json"
	"fmt"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/proto"
	protobuf "google.golang.org/protobuf/proto"
)

// ─── Contents ───────────────────────────────────────────────────────────────
//   ParseMemoryRequest — ParseMemoryRequest parses a memory tool request (protobuf or JSON)
//   MemoryRequestToParams — MemoryRequestToParams converts a protobuf MemoryRequest to params map for compatibility
//   MemoryToProto — MemoryToProto converts a Memory to protobuf Memory.
//   ProtoToMemory — ProtoToMemory converts a protobuf Memory to Memory.
//   SerializeMemoryToProtobuf — SerializeMemoryToProtobuf marshals a Memory to its protobuf binary representation.
//   DeserializeMemoryFromProtobuf — DeserializeMemoryFromProtobuf unmarshals protobuf binary data into a Memory.
//   MemoryResponseToMap — MemoryResponseToMap converts MemoryResponse proto to map for response.FormatResult.
//   ParseContextRequest — ParseContextRequest parses a context tool request (protobuf or JSON)
//   ContextRequestToParams — ContextRequestToParams converts a protobuf ContextRequest to params map for compatibility
//   ParseContextItems — ParseContextItems parses items from protobuf request or JSON string
//   ContextItemToDataString — ContextItemToDataString extracts data string from ContextItem
//   ParseReportRequest — ParseReportRequest parses a report tool request (protobuf or JSON).
//   ReportRequestToParams — ReportRequestToParams converts a protobuf ReportRequest to params map
//   ProjectInfoToProto — ProjectInfoToProto converts map[string]interface{} to proto.ProjectInfo.
//   ProtoToProjectInfo — ProtoToProjectInfo converts proto.ProjectInfo to map[string]interface{}.
//   HealthDataToProto — HealthDataToProto converts map[string]interface{} to proto.HealthData.
//   ProtoToHealthData — ProtoToHealthData converts proto.HealthData to map[string]interface{}.
//   CodebaseMetricsToProto — CodebaseMetricsToProto converts map[string]interface{} to proto.CodebaseMetrics.
//   ProtoToCodebaseMetrics — ProtoToCodebaseMetrics converts proto.CodebaseMetrics to map[string]interface{}.
//   TaskMetricsToProto — TaskMetricsToProto converts map[string]interface{} to proto.TaskMetrics.
//   ProtoToTaskMetrics — ProtoToTaskMetrics converts proto.TaskMetrics to map[string]interface{}.
// ────────────────────────────────────────────────────────────────────────────

// ─── ParseMemoryRequest ─────────────────────────────────────────────────────
// ParseMemoryRequest parses a memory tool request (protobuf or JSON)
// Returns protobuf request if protobuf format, or nil with JSON params map.
func ParseMemoryRequest(args json.RawMessage) (*proto.MemoryRequest, map[string]interface{}, error) {
	req, params, err := framework.ParseRequest(args, func() *proto.MemoryRequest {
		return &proto.MemoryRequest{}
	})
	if err != nil {
		return nil, nil, err
	}

	if req != nil {
		return req, nil, nil
	}

	return nil, params, nil
}

// ─── MemoryRequestToParams ──────────────────────────────────────────────────
// MemoryRequestToParams converts a protobuf MemoryRequest to params map for compatibility
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func MemoryRequestToParams(req *proto.MemoryRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := framework.ProtobufToParams(req, &framework.ProtobufToParamsOptions{
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

// ─── MemoryToProto ──────────────────────────────────────────────────────────
// MemoryToProto converts a Memory to protobuf Memory.
func MemoryToProto(memory *Memory) (*proto.Memory, error) {
	if memory == nil {
		return nil, fmt.Errorf("memory is nil")
	}

	pbMemory := &proto.Memory{
		Id:          memory.ID,
		Title:       memory.Title,
		Content:     memory.Content,
		Category:    memory.Category,
		LinkedTasks: memory.LinkedTasks,
		CreatedAt:   memory.CreatedAt,
		SessionDate: memory.SessionDate,
	}

	// Convert metadata from map[string]interface{} to map[string]string
	// Complex values are serialized to JSON strings
	if memory.Metadata != nil && len(memory.Metadata) > 0 {
		pbMemory.Metadata = make(map[string]string)

		for k, v := range memory.Metadata {
			switch val := v.(type) {
			case string:
				pbMemory.Metadata[k] = val
			case int, int32, int64, float32, float64, bool:
				pbMemory.Metadata[k] = fmt.Sprintf("%v", val)
			default:
				// Marshal complex types to JSON string
				jsonBytes, err := json.Marshal(val)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal metadata value for key %s: %w", k, err)
				}

				pbMemory.Metadata[k] = string(jsonBytes)
			}
		}
	}

	return pbMemory, nil
}

// ─── ProtoToMemory ──────────────────────────────────────────────────────────
// ProtoToMemory converts a protobuf Memory to Memory.
func ProtoToMemory(pbMemory *proto.Memory) (*Memory, error) {
	if pbMemory == nil {
		return nil, fmt.Errorf("protobuf memory is nil")
	}

	memory := &Memory{
		ID:          pbMemory.Id,
		Title:       pbMemory.Title,
		Content:     pbMemory.Content,
		Category:    pbMemory.Category,
		LinkedTasks: pbMemory.LinkedTasks,
		CreatedAt:   pbMemory.CreatedAt,
		SessionDate: pbMemory.SessionDate,
	}

	// Convert metadata from map[string]string back to map[string]interface{}
	if pbMemory.Metadata != nil && len(pbMemory.Metadata) > 0 {
		memory.Metadata = make(map[string]interface{})

		for k, v := range pbMemory.Metadata {
			// Attempt to unmarshal if it looks like a JSON string, otherwise keep as string
			var unmarshaled interface{}
			if err := json.Unmarshal([]byte(v), &unmarshaled); err == nil {
				memory.Metadata[k] = unmarshaled
			} else {
				memory.Metadata[k] = v
			}
		}
	}

	return memory, nil
}

// ─── SerializeMemoryToProtobuf ──────────────────────────────────────────────
// SerializeMemoryToProtobuf marshals a Memory to its protobuf binary representation.
func SerializeMemoryToProtobuf(memory *Memory) ([]byte, error) {
	pbMemory, err := MemoryToProto(memory)
	if err != nil {
		return nil, err
	}

	return protobuf.Marshal(pbMemory)
}

// ─── DeserializeMemoryFromProtobuf ──────────────────────────────────────────
// DeserializeMemoryFromProtobuf unmarshals protobuf binary data into a Memory.
func DeserializeMemoryFromProtobuf(data []byte) (*Memory, error) {
	pbMemory := &proto.Memory{}
	if err := protobuf.Unmarshal(data, pbMemory); err != nil {
		return nil, fmt.Errorf("failed to unmarshal protobuf data: %w", err)
	}

	return ProtoToMemory(pbMemory)
}

// ─── MemoryResponseToMap ────────────────────────────────────────────────────
// MemoryResponseToMap converts MemoryResponse proto to map for response.FormatResult.
func MemoryResponseToMap(resp *proto.MemoryResponse) map[string]interface{} {
	if resp == nil {
		return nil
	}

	out := map[string]interface{}{
		"success": resp.GetSuccess(),
		"method":  resp.GetMethod(),
		"count":   resp.GetCount(),
	}
	if resp.MemoryId != "" {
		out["memory_id"] = resp.MemoryId
	}

	if resp.Message != "" {
		out["message"] = resp.Message
	}

	if len(resp.Memories) > 0 {
		memories := make([]Memory, 0, len(resp.Memories))

		for _, pm := range resp.Memories {
			if m, err := ProtoToMemory(pm); err == nil {
				memories = append(memories, *m)
			}
		}

		out["memories"] = formatMemories(memories)
	}

	if len(resp.Categories) > 0 {
		cat := make(map[string]int)
		for k, v := range resp.Categories {
			cat[k] = int(v)
		}

		out["categories"] = cat
	}

	if len(resp.AvailableCategories) > 0 {
		out["available_categories"] = resp.AvailableCategories
	}

	if resp.TotalFound != 0 {
		out["total_found"] = resp.TotalFound
	}

	if resp.TaskId != "" {
		out["task_id"] = resp.TaskId
	}

	if resp.IncludeRelated {
		out["include_related"] = true
	}

	if resp.Query != "" {
		out["query"] = resp.Query
	}

	if resp.Total != 0 {
		out["total"] = resp.Total
	}

	if resp.Returned != 0 {
		out["returned"] = resp.Returned
	}

	return out
}

// ─── ParseContextRequest ────────────────────────────────────────────────────
// ParseContextRequest parses a context tool request (protobuf or JSON)
// Returns protobuf request if protobuf format, or nil with JSON params map.
func ParseContextRequest(args json.RawMessage) (*proto.ContextRequest, map[string]interface{}, error) {
	req, params, err := framework.ParseRequest(args, func() *proto.ContextRequest {
		return &proto.ContextRequest{}
	})
	if err != nil {
		return nil, nil, err
	}

	if req != nil {
		return req, nil, nil
	}

	return nil, params, nil
}

// ─── ContextRequestToParams ─────────────────────────────────────────────────
// ContextRequestToParams converts a protobuf ContextRequest to params map for compatibility
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func ContextRequestToParams(req *proto.ContextRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := framework.ProtobufToParams(req, &framework.ProtobufToParamsOptions{
		FilterEmptyStrings:  true,
		StringifyArrays:     false,
		ConvertFloat64ToInt: true,
		Float64ToIntFields:  []string{"max_tokens", "budget_tokens"},
	})
	if err != nil {
		return make(map[string]interface{})
	}

	return params
}

// ─── ParseContextItems ──────────────────────────────────────────────────────
// ParseContextItems parses items from protobuf request or JSON string
// Returns array of ContextItem-like structures for easier processing.
func ParseContextItems(itemsRaw interface{}) ([]*proto.ContextItem, error) {
	var itemsJSON string

	// Convert itemsRaw to JSON string
	switch v := itemsRaw.(type) {
	case string:
		itemsJSON = v
	case []interface{}:
		// Marshal array to JSON string
		bytes, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal items array: %w", err)
		}

		itemsJSON = string(bytes)
	default:
		return nil, fmt.Errorf("items must be a JSON string or array")
	}

	// Parse JSON array
	var itemsArray []interface{}
	if err := json.Unmarshal([]byte(itemsJSON), &itemsArray); err != nil {
		return nil, fmt.Errorf("failed to parse items JSON: %w", err)
	}

	// Convert to ContextItem messages
	contextItems := make([]*proto.ContextItem, 0, len(itemsArray))

	for _, itemRaw := range itemsArray {
		item := &proto.ContextItem{}

		// Handle different item formats
		switch v := itemRaw.(type) {
		case map[string]interface{}:
			// Extract data and tool_type from map
			if dataRaw, ok := v["data"]; ok {
				// Convert data to JSON string
				dataBytes, err := json.Marshal(dataRaw)
				if err == nil {
					item.Data = string(dataBytes)
				} else {
					item.Data = fmt.Sprintf("%v", dataRaw)
				}
			} else {
				// If no "data" field, use entire map as data
				dataBytes, err := json.Marshal(v)
				if err == nil {
					item.Data = string(dataBytes)
				} else {
					item.Data = fmt.Sprintf("%v", v)
				}
			}

			if toolType, ok := v["tool_type"].(string); ok {
				item.ToolType = toolType
			}
		case string:
			// Item is already a string
			item.Data = v
		default:
			// Convert to JSON string
			dataBytes, err := json.Marshal(v)
			if err == nil {
				item.Data = string(dataBytes)
			} else {
				item.Data = fmt.Sprintf("%v", v)
			}
		}

		contextItems = append(contextItems, item)
	}

	return contextItems, nil
}

// ─── ContextItemToDataString ────────────────────────────────────────────────
// ContextItemToDataString extracts data string from ContextItem
// Helper function to simplify data extraction.
func ContextItemToDataString(item *proto.ContextItem) string {
	return item.Data
}

// ─── ParseReportRequest ─────────────────────────────────────────────────────
// ParseReportRequest parses a report tool request (protobuf or JSON).
func ParseReportRequest(args json.RawMessage) (*proto.ReportRequest, map[string]interface{}, error) {
	req, params, err := framework.ParseRequest(args, func() *proto.ReportRequest {
		return &proto.ReportRequest{}
	})
	if err != nil {
		return nil, nil, err
	}

	if req != nil {
		return req, nil, nil
	}

	return nil, params, nil
}

// ─── ReportRequestToParams ──────────────────────────────────────────────────
// ReportRequestToParams converts a protobuf ReportRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func ReportRequestToParams(req *proto.ReportRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := framework.ProtobufToParams(req, &framework.ProtobufToParamsOptions{
		FilterEmptyStrings:  true,
		StringifyArrays:     false,
		ConvertFloat64ToInt: true,
		Float64ToIntFields:  []string{"overall_score", "completion_score", "documentation_score", "testing_score", "security_score", "alignment_score"},
	})
	if err != nil {
		return make(map[string]interface{})
	}

	return params
}

// ─── ProjectInfoToProto ─────────────────────────────────────────────────────
// ProjectInfoToProto converts map[string]interface{} to proto.ProjectInfo.
func ProjectInfoToProto(info map[string]interface{}) *proto.ProjectInfo {
	pb := &proto.ProjectInfo{}
	if name, ok := info["name"].(string); ok {
		pb.Name = name
	}

	if version, ok := info["version"].(string); ok {
		pb.Version = version
	}

	if desc, ok := info["description"].(string); ok {
		pb.Description = desc
	}

	if typ, ok := info["type"].(string); ok {
		pb.Type = typ
	}

	if status, ok := info["status"].(string); ok {
		pb.Status = status
	}

	return pb
}

// ─── ProtoToProjectInfo ─────────────────────────────────────────────────────
// ProtoToProjectInfo converts proto.ProjectInfo to map[string]interface{}.
func ProtoToProjectInfo(pb *proto.ProjectInfo) map[string]interface{} {
	return map[string]interface{}{
		"name":        pb.Name,
		"version":     pb.Version,
		"description": pb.Description,
		"type":        pb.Type,
		"status":      pb.Status,
	}
}

// ─── HealthDataToProto ──────────────────────────────────────────────────────
// HealthDataToProto converts map[string]interface{} to proto.HealthData.
func HealthDataToProto(health map[string]interface{}) *proto.HealthData {
	pb := &proto.HealthData{}
	if score, ok := health["overall_score"].(float64); ok {
		pb.OverallScore = score
	}

	if ready, ok := health["production_ready"].(bool); ok {
		pb.ProductionReady = ready
	}

	if scores, ok := health["scores"].(map[string]interface{}); ok {
		pb.Scores = make(map[string]float64)

		for k, v := range scores {
			if score, ok := v.(float64); ok {
				pb.Scores[k] = score
			}
		}
	}

	return pb
}

// ─── ProtoToHealthData ──────────────────────────────────────────────────────
// ProtoToHealthData converts proto.HealthData to map[string]interface{}.
func ProtoToHealthData(pb *proto.HealthData) map[string]interface{} {
	result := map[string]interface{}{
		"overall_score":    pb.OverallScore,
		"production_ready": pb.ProductionReady,
	}

	if len(pb.Scores) > 0 {
		scores := make(map[string]interface{})
		for k, v := range pb.Scores {
			scores[k] = v
		}

		result["scores"] = scores
	}

	return result
}

// ─── CodebaseMetricsToProto ─────────────────────────────────────────────────
// CodebaseMetricsToProto converts map[string]interface{} to proto.CodebaseMetrics.
func CodebaseMetricsToProto(metrics map[string]interface{}) *proto.CodebaseMetrics {
	pb := &proto.CodebaseMetrics{}
	if v, ok := metrics["go_files"].(int); ok {
		pb.GoFiles = int32(v)
	} else if v, ok := metrics["go_files"].(float64); ok {
		pb.GoFiles = int32(v)
	}

	if v, ok := metrics["go_lines"].(int); ok {
		pb.GoLines = int32(v)
	} else if v, ok := metrics["go_lines"].(float64); ok {
		pb.GoLines = int32(v)
	}

	if v, ok := metrics["python_files"].(int); ok {
		pb.PythonFiles = int32(v)
	} else if v, ok := metrics["python_files"].(float64); ok {
		pb.PythonFiles = int32(v)
	}

	if v, ok := metrics["python_lines"].(int); ok {
		pb.PythonLines = int32(v)
	} else if v, ok := metrics["python_lines"].(float64); ok {
		pb.PythonLines = int32(v)
	}

	if v, ok := metrics["total_files"].(int); ok {
		pb.TotalFiles = int32(v)
	} else if v, ok := metrics["total_files"].(float64); ok {
		pb.TotalFiles = int32(v)
	}

	if v, ok := metrics["total_lines"].(int); ok {
		pb.TotalLines = int32(v)
	} else if v, ok := metrics["total_lines"].(float64); ok {
		pb.TotalLines = int32(v)
	}

	if v, ok := metrics["tools"].(int); ok {
		pb.Tools = int32(v)
	} else if v, ok := metrics["tools"].(float64); ok {
		pb.Tools = int32(v)
	}

	if v, ok := metrics["prompts"].(int); ok {
		pb.Prompts = int32(v)
	} else if v, ok := metrics["prompts"].(float64); ok {
		pb.Prompts = int32(v)
	}

	if v, ok := metrics["resources"].(int); ok {
		pb.Resources = int32(v)
	} else if v, ok := metrics["resources"].(float64); ok {
		pb.Resources = int32(v)
	}

	return pb
}

// ─── ProtoToCodebaseMetrics ─────────────────────────────────────────────────
// ProtoToCodebaseMetrics converts proto.CodebaseMetrics to map[string]interface{}.
func ProtoToCodebaseMetrics(pb *proto.CodebaseMetrics) map[string]interface{} {
	return map[string]interface{}{
		"go_files":     int(pb.GoFiles),
		"go_lines":     int(pb.GoLines),
		"python_files": int(pb.PythonFiles),
		"python_lines": int(pb.PythonLines),
		"total_files":  int(pb.TotalFiles),
		"total_lines":  int(pb.TotalLines),
		"tools":        int(pb.Tools),
		"prompts":      int(pb.Prompts),
		"resources":    int(pb.Resources),
	}
}

// ─── TaskMetricsToProto ─────────────────────────────────────────────────────
// TaskMetricsToProto converts map[string]interface{} to proto.TaskMetrics.
func TaskMetricsToProto(metrics map[string]interface{}) *proto.TaskMetrics {
	pb := &proto.TaskMetrics{}
	if v, ok := metrics["total"].(int); ok {
		pb.Total = int32(v)
	} else if v, ok := metrics["total"].(float64); ok {
		pb.Total = int32(v)
	}

	if v, ok := metrics["pending"].(int); ok {
		pb.Pending = int32(v)
	} else if v, ok := metrics["pending"].(float64); ok {
		pb.Pending = int32(v)
	}

	if v, ok := metrics["completed"].(int); ok {
		pb.Completed = int32(v)
	} else if v, ok := metrics["completed"].(float64); ok {
		pb.Completed = int32(v)
	}

	if v, ok := metrics["completion_rate"].(float64); ok {
		pb.CompletionRate = v
	}

	if v, ok := metrics["remaining_hours"].(float64); ok {
		pb.RemainingHours = v
	}

	return pb
}

// ─── ProtoToTaskMetrics ─────────────────────────────────────────────────────
// ProtoToTaskMetrics converts proto.TaskMetrics to map[string]interface{}.
func ProtoToTaskMetrics(pb *proto.TaskMetrics) map[string]interface{} {
	return map[string]interface{}{
		"total":           int(pb.Total),
		"pending":         int(pb.Pending),
		"completed":       int(pb.Completed),
		"completion_rate": pb.CompletionRate,
		"remaining_hours": pb.RemainingHours,
	}
}

// GoScorecardResultToProto converts GoScorecardResult to proto.ScorecardData for type-safe report/MLX path.
