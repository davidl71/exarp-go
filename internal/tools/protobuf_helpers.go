package tools

import (
	"encoding/json"
	"fmt"

	"github.com/davidl71/exarp-go/proto"
	protobuf "google.golang.org/protobuf/proto"
)

// ParseMemoryRequest parses a memory tool request (protobuf or JSON)
// Returns protobuf request if protobuf format, or nil with JSON params map
func ParseMemoryRequest(args json.RawMessage) (*proto.MemoryRequest, map[string]interface{}, error) {
	var req proto.MemoryRequest
	
	// Try protobuf binary first
	if err := protobuf.Unmarshal(args, &req); err == nil {
		// Successfully parsed as protobuf
		return &req, nil, nil
	}

	// Fall back to JSON
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	return nil, params, nil
}

// MemoryRequestToParams converts a protobuf MemoryRequest to params map for compatibility
func MemoryRequestToParams(req *proto.MemoryRequest) map[string]interface{} {
	params := make(map[string]interface{})
	
	if req.Action != "" {
		params["action"] = req.Action
	}
	if req.Title != "" {
		params["title"] = req.Title
	}
	if req.Content != "" {
		params["content"] = req.Content
	}
	if req.Category != "" {
		params["category"] = req.Category
	}
	if req.TaskId != "" {
		params["task_id"] = req.TaskId
	}
	if req.Metadata != "" {
		params["metadata"] = req.Metadata
	}
	params["include_related"] = req.IncludeRelated
	if req.Query != "" {
		params["query"] = req.Query
	}
	if req.Limit > 0 {
		params["limit"] = int(req.Limit)
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
	var req proto.ContextRequest

	// Try protobuf binary first
	if err := protobuf.Unmarshal(args, &req); err == nil {
		// Successfully parsed as protobuf
		return &req, nil, nil
	}

	// Fall back to JSON
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	return nil, params, nil
}

// ContextRequestToParams converts a protobuf ContextRequest to params map for compatibility
func ContextRequestToParams(req *proto.ContextRequest) map[string]interface{} {
	params := make(map[string]interface{})

	if req.Action != "" {
		params["action"] = req.Action
	}
	if req.Data != "" {
		params["data"] = req.Data
	}
	if req.Level != "" {
		params["level"] = req.Level
	}
	if req.MaxTokens > 0 {
		params["max_tokens"] = int(req.MaxTokens)
	}
	params["include_raw"] = req.IncludeRaw
	if req.ToolType != "" {
		params["tool_type"] = req.ToolType
	}
	// For batch action
	if req.Items != "" {
		params["items"] = req.Items
	}
	params["combine"] = req.Combine
	// For budget action
	if req.BudgetTokens > 0 {
		params["budget_tokens"] = int(req.BudgetTokens)
	}

	return params
}

// ParseReportRequest parses a report tool request (protobuf or JSON)
func ParseReportRequest(args json.RawMessage) (*proto.ReportRequest, map[string]interface{}, error) {
	var req proto.ReportRequest

	// Try protobuf binary first
	if err := protobuf.Unmarshal(args, &req); err == nil {
		return &req, nil, nil
	}

	// Fall back to JSON
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	return nil, params, nil
}

// ReportRequestToParams converts a protobuf ReportRequest to params map
func ReportRequestToParams(req *proto.ReportRequest) map[string]interface{} {
	params := make(map[string]interface{})

	if req.Action != "" {
		params["action"] = req.Action
	}
	if req.ProjectName != "" {
		params["project_name"] = req.ProjectName
	}
	params["include_metrics"] = req.IncludeMetrics
	params["include_architecture"] = req.IncludeArchitecture
	params["include_tasks"] = req.IncludeTasks
	params["include_recommendations"] = req.IncludeRecommendations
	if req.OutputFormat != "" {
		params["output_format"] = req.OutputFormat
	}
	if req.OutputPath != "" {
		params["output_path"] = req.OutputPath
	}
	// Score fields
	if req.OverallScore > 0 {
		params["overall_score"] = int(req.OverallScore)
	}
	if req.CompletionScore > 0 {
		params["completion_score"] = int(req.CompletionScore)
	}
	if req.DocumentationScore > 0 {
		params["documentation_score"] = int(req.DocumentationScore)
	}
	if req.TestingScore > 0 {
		params["testing_score"] = int(req.TestingScore)
	}
	if req.SecurityScore > 0 {
		params["security_score"] = int(req.SecurityScore)
	}
	if req.AlignmentScore > 0 {
		params["alignment_score"] = int(req.AlignmentScore)
	}

	return params
}

// ParseTaskWorkflowRequest parses a task_workflow tool request (protobuf or JSON)
func ParseTaskWorkflowRequest(args json.RawMessage) (*proto.TaskWorkflowRequest, map[string]interface{}, error) {
	var req proto.TaskWorkflowRequest

	// Try protobuf binary first
	if err := protobuf.Unmarshal(args, &req); err == nil {
		return &req, nil, nil
	}

	// Fall back to JSON
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	return nil, params, nil
}

// TaskWorkflowRequestToParams converts a protobuf TaskWorkflowRequest to params map
func TaskWorkflowRequestToParams(req *proto.TaskWorkflowRequest) map[string]interface{} {
	params := make(map[string]interface{})

	if req.Action != "" {
		params["action"] = req.Action
	}
	if req.TaskId != "" {
		params["task_id"] = req.TaskId
	}
	if req.TaskIds != "" {
		params["task_ids"] = req.TaskIds
	}
	if req.Name != "" {
		params["name"] = req.Name
	}
	if req.LongDescription != "" {
		params["long_description"] = req.LongDescription
	}
	if req.NewStatus != "" {
		params["new_status"] = req.NewStatus
	}
	if req.Status != "" {
		params["status"] = req.Status
	}
	if req.FilterTag != "" {
		params["filter_tag"] = req.FilterTag
	}
	// Tags and Dependencies are repeated string (arrays)
	if len(req.Tags) > 0 {
		// Convert to comma-separated string or JSON array string for compatibility
		tagsJSON, _ := json.Marshal(req.Tags)
		params["tags"] = string(tagsJSON)
	}
	if len(req.Dependencies) > 0 {
		// Convert to comma-separated string or JSON array string for compatibility
		depsJSON, _ := json.Marshal(req.Dependencies)
		params["dependencies"] = string(depsJSON)
	}
	params["auto_estimate"] = req.AutoEstimate
	params["auto_apply"] = req.AutoApply
	params["dry_run"] = req.DryRun
	params["move_to_todo"] = req.MoveToTodo
	params["clarification_none"] = req.ClarificationNone
	if req.ClarificationText != "" {
		params["clarification_text"] = req.ClarificationText
	}
	if req.Decision != "" {
		params["decision"] = req.Decision
	}
	if req.DecisionsJson != "" {
		params["decisions_json"] = req.DecisionsJson
	}
	if req.SubAction != "" {
		params["sub_action"] = req.SubAction
	}
	if req.OutputFormat != "" {
		params["output_format"] = req.OutputFormat
	}
	if req.OutputPath != "" {
		params["output_path"] = req.OutputPath
	}
	if req.StaleThresholdHours > 0 {
		params["stale_threshold_hours"] = int(req.StaleThresholdHours)
	}
	params["include_legacy"] = req.IncludeLegacy
	params["external"] = req.External

	return params
}
