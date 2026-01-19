package tools

import (
	"encoding/json"
	"fmt"

	"github.com/davidl71/exarp-go/proto"
	"github.com/davidl71/mcp-go-core/pkg/mcp/request"
	protobuf "google.golang.org/protobuf/proto"
)

// ParseMemoryRequest parses a memory tool request (protobuf or JSON)
// Returns protobuf request if protobuf format, or nil with JSON params map
func ParseMemoryRequest(args json.RawMessage) (*proto.MemoryRequest, map[string]interface{}, error) {
	req, params, err := request.ParseRequest(args, func() *proto.MemoryRequest {
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

// MemoryRequestToParams converts a protobuf MemoryRequest to params map for compatibility
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func MemoryRequestToParams(req *proto.MemoryRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := request.ProtobufToParams(req, &request.ProtobufToParamsOptions{
		FilterEmptyStrings:    true,
		StringifyArrays:       false,
		ConvertFloat64ToInt:   true,
		Float64ToIntFields:    []string{"limit"},
	})
	if err != nil {
		return make(map[string]interface{})
	}
	
	return params
}

// MemoryToProto converts a Memory to protobuf Memory
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

// ProtoToMemory converts a protobuf Memory to Memory
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

// SerializeMemoryToProtobuf marshals a Memory to its protobuf binary representation
func SerializeMemoryToProtobuf(memory *Memory) ([]byte, error) {
	pbMemory, err := MemoryToProto(memory)
	if err != nil {
		return nil, err
	}
	return protobuf.Marshal(pbMemory)
}

// DeserializeMemoryFromProtobuf unmarshals protobuf binary data into a Memory
func DeserializeMemoryFromProtobuf(data []byte) (*Memory, error) {
	pbMemory := &proto.Memory{}
	if err := protobuf.Unmarshal(data, pbMemory); err != nil {
		return nil, fmt.Errorf("failed to unmarshal protobuf data: %w", err)
	}
	return ProtoToMemory(pbMemory)
}

// ParseContextRequest parses a context tool request (protobuf or JSON)
// Returns protobuf request if protobuf format, or nil with JSON params map
func ParseContextRequest(args json.RawMessage) (*proto.ContextRequest, map[string]interface{}, error) {
	req, params, err := request.ParseRequest(args, func() *proto.ContextRequest {
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

// ContextRequestToParams converts a protobuf ContextRequest to params map for compatibility
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func ContextRequestToParams(req *proto.ContextRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := request.ProtobufToParams(req, &request.ProtobufToParamsOptions{
		FilterEmptyStrings:    true,
		StringifyArrays:       false,
		ConvertFloat64ToInt:   true,
		Float64ToIntFields:    []string{"max_tokens", "budget_tokens"},
	})
	if err != nil {
		return make(map[string]interface{})
	}

	return params
}

// ParseContextItems parses items from protobuf request or JSON string
// Returns array of ContextItem-like structures for easier processing
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

// ContextItemToDataString extracts data string from ContextItem
// Helper function to simplify data extraction
func ContextItemToDataString(item *proto.ContextItem) string {
	return item.Data
}

// ParseReportRequest parses a report tool request (protobuf or JSON)
func ParseReportRequest(args json.RawMessage) (*proto.ReportRequest, map[string]interface{}, error) {
	req, params, err := request.ParseRequest(args, func() *proto.ReportRequest {
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

// ReportRequestToParams converts a protobuf ReportRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func ReportRequestToParams(req *proto.ReportRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := request.ProtobufToParams(req, &request.ProtobufToParamsOptions{
		FilterEmptyStrings:    true,
		StringifyArrays:       false,
		ConvertFloat64ToInt:   true,
		Float64ToIntFields:    []string{"overall_score", "completion_score", "documentation_score", "testing_score", "security_score", "alignment_score"},
	})
	if err != nil {
		return make(map[string]interface{})
	}

	return params
}

// ProjectInfoToProto converts map[string]interface{} to proto.ProjectInfo
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

// ProtoToProjectInfo converts proto.ProjectInfo to map[string]interface{}
func ProtoToProjectInfo(pb *proto.ProjectInfo) map[string]interface{} {
	return map[string]interface{}{
		"name":        pb.Name,
		"version":     pb.Version,
		"description": pb.Description,
		"type":        pb.Type,
		"status":      pb.Status,
	}
}

// HealthDataToProto converts map[string]interface{} to proto.HealthData
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

// ProtoToHealthData converts proto.HealthData to map[string]interface{}
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

// CodebaseMetricsToProto converts map[string]interface{} to proto.CodebaseMetrics
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

// ProtoToCodebaseMetrics converts proto.CodebaseMetrics to map[string]interface{}
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

// TaskMetricsToProto converts map[string]interface{} to proto.TaskMetrics
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

// ProtoToTaskMetrics converts proto.TaskMetrics to map[string]interface{}
func ProtoToTaskMetrics(pb *proto.TaskMetrics) map[string]interface{} {
	return map[string]interface{}{
		"total":           int(pb.Total),
		"pending":         int(pb.Pending),
		"completed":       int(pb.Completed),
		"completion_rate": pb.CompletionRate,
		"remaining_hours": pb.RemainingHours,
	}
}

// ProjectOverviewDataToProto converts map[string]interface{} to proto.ProjectOverviewData
// NOTE: Currently unused - kept for potential future protobuf-based report processing
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

// ProtoToProjectOverviewData converts proto.ProjectOverviewData to map[string]interface{}
// NOTE: Currently unused - kept for potential future protobuf-based report processing
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

// ParseTaskWorkflowRequest parses a task_workflow tool request (protobuf or JSON)
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

// TaskWorkflowRequestToParams converts a protobuf TaskWorkflowRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core
// to eliminate repetitive field-by-field conversion code.
func TaskWorkflowRequestToParams(req *proto.TaskWorkflowRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	// Use generic converter with options matching existing behavior
	params, err := request.ProtobufToParams(req, &request.ProtobufToParamsOptions{
		FilterEmptyStrings:    true,  // Filter empty strings and zero numbers
		StringifyArrays:       true,  // Convert arrays to JSON strings
		ConvertFloat64ToInt:   true,  // Convert stale_threshold_hours from float64 to int
		Float64ToIntFields:    []string{"stale_threshold_hours"},
	})
	if err != nil {
		// Fallback to empty map if conversion fails (shouldn't happen in practice)
		return make(map[string]interface{})
	}

	return params
}

// ParseHealthRequest parses a health tool request (protobuf or JSON)
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

// ParseSecurityRequest parses a security tool request (protobuf or JSON)
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

// ParseInferSessionModeRequest parses an infer_session_mode request (protobuf or JSON)
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

// ParseToolCatalogRequest parses a tool_catalog request (protobuf or JSON)
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

// ParseWorkflowModeRequest parses a workflow_mode request (protobuf or JSON)
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

// ParseEstimationRequest parses an estimation request (protobuf or JSON)
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

// EstimationRequestToParams converts a protobuf EstimationRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func EstimationRequestToParams(req *proto.EstimationRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := request.ProtobufToParams(req, &request.ProtobufToParamsOptions{
		FilterEmptyStrings:    true,
		StringifyArrays:       true, // TagList is an array
		ConvertFloat64ToInt:   false, // Keep mlx_weight as float64 (it's a weight, not a count)
	})
	if err != nil {
		return make(map[string]interface{})
	}

	return params
}

// ParseSessionRequest parses a session request (protobuf or JSON)
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

// SessionRequestToParams converts a protobuf SessionRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func SessionRequestToParams(req *proto.SessionRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := request.ProtobufToParams(req, &request.ProtobufToParamsOptions{
		FilterEmptyStrings:    true,
		StringifyArrays:       false,
		ConvertFloat64ToInt:   true,
		Float64ToIntFields:    []string{"limit"},
	})
	if err != nil {
		return make(map[string]interface{})
	}

	return params
}

// ParseGitToolsRequest parses a git_tools request (protobuf or JSON)
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

// GitToolsRequestToParams converts a protobuf GitToolsRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func GitToolsRequestToParams(req *proto.GitToolsRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := request.ProtobufToParams(req, &request.ProtobufToParamsOptions{
		FilterEmptyStrings:    true,
		StringifyArrays:       false,
		ConvertFloat64ToInt:   true,
		Float64ToIntFields:    []string{"limit", "max_commits"},
	})
	if err != nil {
		return make(map[string]interface{})
	}

	return params
}

// ParseMemoryMaintRequest parses a memory_maint request (protobuf or JSON)
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

// MemoryMaintRequestToParams converts a protobuf MemoryMaintRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func MemoryMaintRequestToParams(req *proto.MemoryMaintRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := request.ProtobufToParams(req, &request.ProtobufToParamsOptions{
		FilterEmptyStrings:    true,
		StringifyArrays:       false,
		ConvertFloat64ToInt:   true,
		Float64ToIntFields:    []string{"max_age_days", "scorecard_max_age_days", "keep_minimum"},
		// Note: value_threshold and similarity_threshold remain as float64 (they're thresholds, not counts)
	})
	if err != nil {
		return make(map[string]interface{})
	}

	return params
}

// ParseTaskAnalysisRequest parses a task_analysis request (protobuf or JSON)
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

// TaskAnalysisRequestToParams converts a protobuf TaskAnalysisRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func TaskAnalysisRequestToParams(req *proto.TaskAnalysisRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := request.ProtobufToParams(req, &request.ProtobufToParamsOptions{
		FilterEmptyStrings:    true,
		StringifyArrays:       false,
		ConvertFloat64ToInt:   false, // Keep similarity_threshold as float64 (it's a threshold, not a count)
	})
	if err != nil {
		return make(map[string]interface{})
	}

	return params
}

// ParseTaskDiscoveryRequest parses a task_discovery request (protobuf or JSON)
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

// ParseOllamaRequest parses an ollama request (protobuf or JSON)
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

// OllamaRequestToParams converts a protobuf OllamaRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func OllamaRequestToParams(req *proto.OllamaRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := request.ProtobufToParams(req, &request.ProtobufToParamsOptions{
		FilterEmptyStrings:    true,
		StringifyArrays:       false,
		ConvertFloat64ToInt:   true,
		Float64ToIntFields:    []string{"num_gpu", "num_threads", "context_size"},
	})
	if err != nil {
		return make(map[string]interface{})
	}

	return params
}

// ParseMlxRequest parses an mlx request (protobuf or JSON)
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

// MlxRequestToParams converts a protobuf MlxRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func MlxRequestToParams(req *proto.MLXRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := request.ProtobufToParams(req, &request.ProtobufToParamsOptions{
		FilterEmptyStrings:    true,
		StringifyArrays:       false,
		ConvertFloat64ToInt:   true,
		Float64ToIntFields:    []string{"max_tokens"},
		// Note: temperature remains as float64 (it's a threshold, not a count)
	})
	if err != nil {
		return make(map[string]interface{})
	}

	return params
}

// ParsePromptTrackingRequest parses a prompt_tracking request (protobuf or JSON)
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

// PromptTrackingRequestToParams converts a protobuf PromptTrackingRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func PromptTrackingRequestToParams(req *proto.PromptTrackingRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := request.ProtobufToParams(req, &request.ProtobufToParamsOptions{
		FilterEmptyStrings:    true,
		StringifyArrays:       false,
		ConvertFloat64ToInt:   true,
		Float64ToIntFields:    []string{"iteration", "days"},
	})
	if err != nil {
		return make(map[string]interface{})
	}

	return params
}

// ParseRecommendRequest parses a recommend request (protobuf or JSON)
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

// ParseAnalyzeAlignmentRequest parses an analyze_alignment request (protobuf or JSON)
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

// ParseGenerateConfigRequest parses a generate_config request (protobuf or JSON)
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

// ParseSetupHooksRequest parses a setup_hooks request (protobuf or JSON)
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

// ParseCheckAttributionRequest parses a check_attribution request (protobuf or JSON)
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

// ParseAddExternalToolHintsRequest parses an add_external_tool_hints request (protobuf or JSON)
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

// AddExternalToolHintsRequestToParams converts a protobuf AddExternalToolHintsRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func AddExternalToolHintsRequestToParams(req *proto.AddExternalToolHintsRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := request.ProtobufToParams(req, &request.ProtobufToParamsOptions{
		FilterEmptyStrings:    true,
		StringifyArrays:       false,
		ConvertFloat64ToInt:   true,
		Float64ToIntFields:    []string{"min_file_size"},
	})
	if err != nil {
		return make(map[string]interface{})
	}

	return params
}

// ParseTestingRequest parses a testing tool request (protobuf or JSON)
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

// TestingRequestToParams converts a protobuf TestingRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func TestingRequestToParams(req *proto.TestingRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := request.ProtobufToParams(req, &request.ProtobufToParamsOptions{
		FilterEmptyStrings:    true,
		StringifyArrays:       false,
		ConvertFloat64ToInt:   true,
		Float64ToIntFields:    []string{"min_coverage"},
		// Note: min_confidence remains as float64 (it's a threshold, not a count)
	})
	if err != nil {
		return make(map[string]interface{})
	}

	return params
}

// ParseAutomationRequest parses an automation tool request (protobuf or JSON)
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

// AutomationRequestToParams converts a protobuf AutomationRequest to params map
// This function now uses the generic ProtobufToParams converter from mcp-go-core.
func AutomationRequestToParams(req *proto.AutomationRequest) map[string]interface{} {
	if req == nil {
		return make(map[string]interface{})
	}

	params, err := request.ProtobufToParams(req, &request.ProtobufToParamsOptions{
		FilterEmptyStrings:    true,
		StringifyArrays:       true, // Tasks and TagFilter are arrays
		ConvertFloat64ToInt:   true,
		Float64ToIntFields:    []string{"max_tasks_per_host", "max_parallel_tasks", "max_iterations"},
		// Note: min_value_score remains as float64 (it's a threshold, not a count)
	})
	if err != nil {
		return make(map[string]interface{})
	}

	return params
}

// ParseLintRequest parses a lint tool request (protobuf or JSON)
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
