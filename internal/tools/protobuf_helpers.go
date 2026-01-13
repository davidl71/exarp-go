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

// ParseHealthRequest parses a health tool request (protobuf or JSON)
func ParseHealthRequest(args json.RawMessage) (*proto.HealthRequest, map[string]interface{}, error) {
	var req proto.HealthRequest

	if err := protobuf.Unmarshal(args, &req); err == nil {
		return &req, nil, nil
	}

	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	return nil, params, nil
}

// HealthRequestToParams converts a protobuf HealthRequest to params map
func HealthRequestToParams(req *proto.HealthRequest) map[string]interface{} {
	params := make(map[string]interface{})

	if req.Action != "" {
		params["action"] = req.Action
	}
	if req.AgentName != "" {
		params["agent_name"] = req.AgentName
	}
	params["check_remote"] = req.CheckRemote
	if req.OutputPath != "" {
		params["output_path"] = req.OutputPath
	}
	params["create_tasks"] = req.CreateTasks
	if req.TaskId != "" {
		params["task_id"] = req.TaskId
	}
	if req.ChangedFiles != "" {
		params["changed_files"] = req.ChangedFiles
	}
	params["auto_check"] = req.AutoCheck
	if req.WorkflowPath != "" {
		params["workflow_path"] = req.WorkflowPath
	}
	params["check_runners"] = req.CheckRunners

	return params
}

// ParseSecurityRequest parses a security tool request (protobuf or JSON)
func ParseSecurityRequest(args json.RawMessage) (*proto.SecurityRequest, map[string]interface{}, error) {
	var req proto.SecurityRequest

	if err := protobuf.Unmarshal(args, &req); err == nil {
		return &req, nil, nil
	}

	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	return nil, params, nil
}

// SecurityRequestToParams converts a protobuf SecurityRequest to params map
func SecurityRequestToParams(req *proto.SecurityRequest) map[string]interface{} {
	params := make(map[string]interface{})

	if req.Action != "" {
		params["action"] = req.Action
	}
	if req.Repo != "" {
		params["repo"] = req.Repo
	}
	if len(req.Languages) > 0 {
		langsJSON, _ := json.Marshal(req.Languages)
		params["languages"] = string(langsJSON)
	}
	if req.ConfigPath != "" {
		params["config_path"] = req.ConfigPath
	}
	if req.State != "" {
		params["state"] = req.State
	}
	params["include_dismissed"] = req.IncludeDismissed

	return params
}

// ParseInferSessionModeRequest parses an infer_session_mode request (protobuf or JSON)
func ParseInferSessionModeRequest(args json.RawMessage) (*proto.InferSessionModeRequest, map[string]interface{}, error) {
	var req proto.InferSessionModeRequest

	if err := protobuf.Unmarshal(args, &req); err == nil {
		return &req, nil, nil
	}

	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	return nil, params, nil
}

// InferSessionModeRequestToParams converts a protobuf InferSessionModeRequest to params map
func InferSessionModeRequestToParams(req *proto.InferSessionModeRequest) map[string]interface{} {
	params := make(map[string]interface{})

	params["force_recompute"] = req.ForceRecompute

	return params
}

// ParseToolCatalogRequest parses a tool_catalog request (protobuf or JSON)
func ParseToolCatalogRequest(args json.RawMessage) (*proto.ToolCatalogRequest, map[string]interface{}, error) {
	var req proto.ToolCatalogRequest

	if err := protobuf.Unmarshal(args, &req); err == nil {
		return &req, nil, nil
	}

	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	return nil, params, nil
}

// ToolCatalogRequestToParams converts a protobuf ToolCatalogRequest to params map
func ToolCatalogRequestToParams(req *proto.ToolCatalogRequest) map[string]interface{} {
	params := make(map[string]interface{})

	if req.Action != "" {
		params["action"] = req.Action
	}
	if req.ToolName != "" {
		params["tool_name"] = req.ToolName
	}

	return params
}

// ParseWorkflowModeRequest parses a workflow_mode request (protobuf or JSON)
func ParseWorkflowModeRequest(args json.RawMessage) (*proto.WorkflowModeRequest, map[string]interface{}, error) {
	var req proto.WorkflowModeRequest

	if err := protobuf.Unmarshal(args, &req); err == nil {
		return &req, nil, nil
	}

	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	return nil, params, nil
}

// WorkflowModeRequestToParams converts a protobuf WorkflowModeRequest to params map
func WorkflowModeRequestToParams(req *proto.WorkflowModeRequest) map[string]interface{} {
	params := make(map[string]interface{})

	if req.Action != "" {
		params["action"] = req.Action
	}
	if req.Mode != "" {
		params["mode"] = req.Mode
	}
	if req.EnableGroup != "" {
		params["enable_group"] = req.EnableGroup
	}
	if req.DisableGroup != "" {
		params["disable_group"] = req.DisableGroup
	}
	params["status"] = req.Status
	if req.Text != "" {
		params["text"] = req.Text
	}
	params["auto_switch"] = req.AutoSwitch

	return params
}

// ParseEstimationRequest parses an estimation request (protobuf or JSON)
func ParseEstimationRequest(args json.RawMessage) (*proto.EstimationRequest, map[string]interface{}, error) {
	var req proto.EstimationRequest

	if err := protobuf.Unmarshal(args, &req); err == nil {
		return &req, nil, nil
	}

	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	return nil, params, nil
}

// EstimationRequestToParams converts a protobuf EstimationRequest to params map
func EstimationRequestToParams(req *proto.EstimationRequest) map[string]interface{} {
	params := make(map[string]interface{})

	if req.Action != "" {
		params["action"] = req.Action
	}
	if req.Name != "" {
		params["name"] = req.Name
	}
	if req.Details != "" {
		params["details"] = req.Details
	}
	if req.Tags != "" {
		params["tags"] = req.Tags
	}
	if len(req.TagList) > 0 {
		tagsJSON, _ := json.Marshal(req.TagList)
		params["tag_list"] = string(tagsJSON)
	}
	if req.Priority != "" {
		params["priority"] = req.Priority
	}
	params["use_historical"] = req.UseHistorical
	params["detailed"] = req.Detailed
	params["use_mlx"] = req.UseMlx
	if req.MlxWeight > 0 {
		params["mlx_weight"] = req.MlxWeight
	}

	return params
}

// ParseSessionRequest parses a session request (protobuf or JSON)
func ParseSessionRequest(args json.RawMessage) (*proto.SessionRequest, map[string]interface{}, error) {
	var req proto.SessionRequest

	if err := protobuf.Unmarshal(args, &req); err == nil {
		return &req, nil, nil
	}

	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	return nil, params, nil
}

// SessionRequestToParams converts a protobuf SessionRequest to params map
func SessionRequestToParams(req *proto.SessionRequest) map[string]interface{} {
	params := make(map[string]interface{})

	if req.Action != "" {
		params["action"] = req.Action
	}
	params["include_hints"] = req.IncludeHints
	params["include_tasks"] = req.IncludeTasks
	if req.OverrideMode != "" {
		params["override_mode"] = req.OverrideMode
	}
	if req.TaskId != "" {
		params["task_id"] = req.TaskId
	}
	if req.Summary != "" {
		params["summary"] = req.Summary
	}
	if req.Blockers != "" {
		params["blockers"] = req.Blockers
	}
	if req.NextSteps != "" {
		params["next_steps"] = req.NextSteps
	}
	params["unassign_my_tasks"] = req.UnassignMyTasks
	params["include_git_status"] = req.IncludeGitStatus
	if req.Limit > 0 {
		params["limit"] = int(req.Limit)
	}
	params["dry_run"] = req.DryRun
	if req.Direction != "" {
		params["direction"] = req.Direction
	}
	params["prefer_agentic_tools"] = req.PreferAgenticTools
	params["auto_commit"] = req.AutoCommit

	return params
}

// ParseGitToolsRequest parses a git_tools request (protobuf or JSON)
func ParseGitToolsRequest(args json.RawMessage) (*proto.GitToolsRequest, map[string]interface{}, error) {
	var req proto.GitToolsRequest

	if err := protobuf.Unmarshal(args, &req); err == nil {
		return &req, nil, nil
	}

	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	return nil, params, nil
}

// GitToolsRequestToParams converts a protobuf GitToolsRequest to params map
func GitToolsRequestToParams(req *proto.GitToolsRequest) map[string]interface{} {
	params := make(map[string]interface{})

	if req.Action != "" {
		params["action"] = req.Action
	}
	if req.TaskId != "" {
		params["task_id"] = req.TaskId
	}
	if req.Branch != "" {
		params["branch"] = req.Branch
	}
	if req.Limit > 0 {
		params["limit"] = int(req.Limit)
	}
	if req.Commit1 != "" {
		params["commit1"] = req.Commit1
	}
	if req.Commit2 != "" {
		params["commit2"] = req.Commit2
	}
	if req.Time1 != "" {
		params["time1"] = req.Time1
	}
	if req.Time2 != "" {
		params["time2"] = req.Time2
	}
	if req.Format != "" {
		params["format"] = req.Format
	}
	if req.OutputPath != "" {
		params["output_path"] = req.OutputPath
	}
	if req.MaxCommits > 0 {
		params["max_commits"] = int(req.MaxCommits)
	}
	if req.SourceBranch != "" {
		params["source_branch"] = req.SourceBranch
	}
	if req.TargetBranch != "" {
		params["target_branch"] = req.TargetBranch
	}
	if req.ConflictStrategy != "" {
		params["conflict_strategy"] = req.ConflictStrategy
	}
	if req.Author != "" {
		params["author"] = req.Author
	}
	params["dry_run"] = req.DryRun

	return params
}

// ParseMemoryMaintRequest parses a memory_maint request (protobuf or JSON)
func ParseMemoryMaintRequest(args json.RawMessage) (*proto.MemoryMaintRequest, map[string]interface{}, error) {
	var req proto.MemoryMaintRequest

	if err := protobuf.Unmarshal(args, &req); err == nil {
		return &req, nil, nil
	}

	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	return nil, params, nil
}

// MemoryMaintRequestToParams converts a protobuf MemoryMaintRequest to params map
func MemoryMaintRequestToParams(req *proto.MemoryMaintRequest) map[string]interface{} {
	params := make(map[string]interface{})

	if req.Action != "" {
		params["action"] = req.Action
	}
	if req.MaxAgeDays > 0 {
		params["max_age_days"] = int(req.MaxAgeDays)
	}
	params["delete_orphaned"] = req.DeleteOrphaned
	params["delete_duplicates"] = req.DeleteDuplicates
	if req.ScorecardMaxAgeDays > 0 {
		params["scorecard_max_age_days"] = int(req.ScorecardMaxAgeDays)
	}
	if req.ValueThreshold > 0 {
		params["value_threshold"] = req.ValueThreshold
	}
	if req.KeepMinimum > 0 {
		params["keep_minimum"] = int(req.KeepMinimum)
	}
	if req.SimilarityThreshold > 0 {
		params["similarity_threshold"] = req.SimilarityThreshold
	}
	if req.MergeStrategy != "" {
		params["merge_strategy"] = req.MergeStrategy
	}
	if req.Scope != "" {
		params["scope"] = req.Scope
	}
	params["dry_run"] = req.DryRun
	params["interactive"] = req.Interactive
	params["generate_insights"] = req.GenerateInsights
	params["save_dream"] = req.SaveDream
	if req.Advisors != "" {
		params["advisors"] = req.Advisors
	}

	return params
}

// ParseTaskAnalysisRequest parses a task_analysis request (protobuf or JSON)
func ParseTaskAnalysisRequest(args json.RawMessage) (*proto.TaskAnalysisRequest, map[string]interface{}, error) {
	var req proto.TaskAnalysisRequest

	if err := protobuf.Unmarshal(args, &req); err == nil {
		return &req, nil, nil
	}

	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	return nil, params, nil
}

// TaskAnalysisRequestToParams converts a protobuf TaskAnalysisRequest to params map
func TaskAnalysisRequestToParams(req *proto.TaskAnalysisRequest) map[string]interface{} {
	params := make(map[string]interface{})

	if req.Action != "" {
		params["action"] = req.Action
	}
	if req.SimilarityThreshold > 0 {
		params["similarity_threshold"] = req.SimilarityThreshold
	}
	params["auto_fix"] = req.AutoFix
	params["dry_run"] = req.DryRun
	if req.CustomRules != "" {
		params["custom_rules"] = req.CustomRules
	}
	if req.RemoveTags != "" {
		params["remove_tags"] = req.RemoveTags
	}
	if req.OutputFormat != "" {
		params["output_format"] = req.OutputFormat
	}
	params["include_recommendations"] = req.IncludeRecommendations
	if req.OutputPath != "" {
		params["output_path"] = req.OutputPath
	}

	return params
}

// ParseTaskDiscoveryRequest parses a task_discovery request (protobuf or JSON)
func ParseTaskDiscoveryRequest(args json.RawMessage) (*proto.TaskDiscoveryRequest, map[string]interface{}, error) {
	var req proto.TaskDiscoveryRequest

	if err := protobuf.Unmarshal(args, &req); err == nil {
		return &req, nil, nil
	}

	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	return nil, params, nil
}

// TaskDiscoveryRequestToParams converts a protobuf TaskDiscoveryRequest to params map
func TaskDiscoveryRequestToParams(req *proto.TaskDiscoveryRequest) map[string]interface{} {
	params := make(map[string]interface{})

	if req.Action != "" {
		params["action"] = req.Action
	}
	if req.FilePatterns != "" {
		params["file_patterns"] = req.FilePatterns
	}
	params["include_fixme"] = req.IncludeFixme
	if req.DocPath != "" {
		params["doc_path"] = req.DocPath
	}
	if req.JsonPattern != "" {
		params["json_pattern"] = req.JsonPattern
	}
	if req.OutputPath != "" {
		params["output_path"] = req.OutputPath
	}
	params["create_tasks"] = req.CreateTasks

	return params
}

// ParseOllamaRequest parses an ollama request (protobuf or JSON)
func ParseOllamaRequest(args json.RawMessage) (*proto.OllamaRequest, map[string]interface{}, error) {
	var req proto.OllamaRequest

	if err := protobuf.Unmarshal(args, &req); err == nil {
		return &req, nil, nil
	}

	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	return nil, params, nil
}

// OllamaRequestToParams converts a protobuf OllamaRequest to params map
func OllamaRequestToParams(req *proto.OllamaRequest) map[string]interface{} {
	params := make(map[string]interface{})

	if req.Action != "" {
		params["action"] = req.Action
	}
	if req.Host != "" {
		params["host"] = req.Host
	}
	if req.Prompt != "" {
		params["prompt"] = req.Prompt
	}
	if req.Model != "" {
		params["model"] = req.Model
	}
	params["stream"] = req.Stream
	if req.Options != "" {
		params["options"] = req.Options
	}
	if req.NumGpu > 0 {
		params["num_gpu"] = int(req.NumGpu)
	}
	if req.NumThreads > 0 {
		params["num_threads"] = int(req.NumThreads)
	}
	if req.ContextSize > 0 {
		params["context_size"] = int(req.ContextSize)
	}
	if req.FilePath != "" {
		params["file_path"] = req.FilePath
	}
	if req.OutputPath != "" {
		params["output_path"] = req.OutputPath
	}
	if req.Data != "" {
		params["data"] = req.Data
	}
	if req.Style != "" {
		params["style"] = req.Style
	}
	if req.Level != "" {
		params["level"] = req.Level
	}
	params["include_suggestions"] = req.IncludeSuggestions

	return params
}

// ParseMlxRequest parses an mlx request (protobuf or JSON)
func ParseMlxRequest(args json.RawMessage) (*proto.MLXRequest, map[string]interface{}, error) {
	var req proto.MLXRequest

	if err := protobuf.Unmarshal(args, &req); err == nil {
		return &req, nil, nil
	}

	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	return nil, params, nil
}

// MlxRequestToParams converts a protobuf MlxRequest to params map
func MlxRequestToParams(req *proto.MLXRequest) map[string]interface{} {
	params := make(map[string]interface{})

	if req.Action != "" {
		params["action"] = req.Action
	}
	if req.Prompt != "" {
		params["prompt"] = req.Prompt
	}
	if req.Model != "" {
		params["model"] = req.Model
	}
	if req.MaxTokens > 0 {
		params["max_tokens"] = int(req.MaxTokens)
	}
	if req.Temperature > 0 {
		params["temperature"] = req.Temperature
	}
	params["verbose"] = req.Verbose

	return params
}

// ParsePromptTrackingRequest parses a prompt_tracking request (protobuf or JSON)
func ParsePromptTrackingRequest(args json.RawMessage) (*proto.PromptTrackingRequest, map[string]interface{}, error) {
	var req proto.PromptTrackingRequest

	if err := protobuf.Unmarshal(args, &req); err == nil {
		return &req, nil, nil
	}

	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	return nil, params, nil
}

// PromptTrackingRequestToParams converts a protobuf PromptTrackingRequest to params map
func PromptTrackingRequestToParams(req *proto.PromptTrackingRequest) map[string]interface{} {
	params := make(map[string]interface{})

	if req.Action != "" {
		params["action"] = req.Action
	}
	if req.Prompt != "" {
		params["prompt"] = req.Prompt
	}
	if req.TaskId != "" {
		params["task_id"] = req.TaskId
	}
	if req.Mode != "" {
		params["mode"] = req.Mode
	}
	if req.Outcome != "" {
		params["outcome"] = req.Outcome
	}
	if req.Iteration > 0 {
		params["iteration"] = int(req.Iteration)
	}
	if req.Days > 0 {
		params["days"] = int(req.Days)
	}

	return params
}

// ParseRecommendRequest parses a recommend request (protobuf or JSON)
func ParseRecommendRequest(args json.RawMessage) (*proto.RecommendRequest, map[string]interface{}, error) {
	var req proto.RecommendRequest

	if err := protobuf.Unmarshal(args, &req); err == nil {
		return &req, nil, nil
	}

	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	return nil, params, nil
}

// RecommendRequestToParams converts a protobuf RecommendRequest to params map
func RecommendRequestToParams(req *proto.RecommendRequest) map[string]interface{} {
	params := make(map[string]interface{})

	if req.Action != "" {
		params["action"] = req.Action
	}
	if req.TaskDescription != "" {
		params["task_description"] = req.TaskDescription
	}
	if req.Tags != "" {
		params["tags"] = req.Tags
	}
	params["include_rationale"] = req.IncludeRationale
	if req.TaskType != "" {
		params["task_type"] = req.TaskType
	}
	if req.OptimizeFor != "" {
		params["optimize_for"] = req.OptimizeFor
	}
	params["include_alternatives"] = req.IncludeAlternatives

	return params
}

// ParseAnalyzeAlignmentRequest parses an analyze_alignment request (protobuf or JSON)
func ParseAnalyzeAlignmentRequest(args json.RawMessage) (*proto.AnalyzeAlignmentRequest, map[string]interface{}, error) {
	var req proto.AnalyzeAlignmentRequest

	if err := protobuf.Unmarshal(args, &req); err == nil {
		return &req, nil, nil
	}

	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	return nil, params, nil
}

// AnalyzeAlignmentRequestToParams converts a protobuf AnalyzeAlignmentRequest to params map
func AnalyzeAlignmentRequestToParams(req *proto.AnalyzeAlignmentRequest) map[string]interface{} {
	params := make(map[string]interface{})

	if req.Action != "" {
		params["action"] = req.Action
	}
	params["create_followup_tasks"] = req.CreateFollowupTasks
	if req.OutputPath != "" {
		params["output_path"] = req.OutputPath
	}

	return params
}

// ParseGenerateConfigRequest parses a generate_config request (protobuf or JSON)
func ParseGenerateConfigRequest(args json.RawMessage) (*proto.GenerateConfigRequest, map[string]interface{}, error) {
	var req proto.GenerateConfigRequest

	if err := protobuf.Unmarshal(args, &req); err == nil {
		return &req, nil, nil
	}

	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	return nil, params, nil
}

// GenerateConfigRequestToParams converts a protobuf GenerateConfigRequest to params map
func GenerateConfigRequestToParams(req *proto.GenerateConfigRequest) map[string]interface{} {
	params := make(map[string]interface{})

	if req.Action != "" {
		params["action"] = req.Action
	}
	if req.Rules != "" {
		params["rules"] = req.Rules
	}
	params["overwrite"] = req.Overwrite
	params["analyze_only"] = req.AnalyzeOnly
	params["include_indexing"] = req.IncludeIndexing
	params["analyze_project"] = req.AnalyzeProject
	if req.RuleFiles != "" {
		params["rule_files"] = req.RuleFiles
	}
	if req.OutputDir != "" {
		params["output_dir"] = req.OutputDir
	}
	params["dry_run"] = req.DryRun

	return params
}

// ParseSetupHooksRequest parses a setup_hooks request (protobuf or JSON)
func ParseSetupHooksRequest(args json.RawMessage) (*proto.SetupHooksRequest, map[string]interface{}, error) {
	var req proto.SetupHooksRequest

	if err := protobuf.Unmarshal(args, &req); err == nil {
		return &req, nil, nil
	}

	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	return nil, params, nil
}

// SetupHooksRequestToParams converts a protobuf SetupHooksRequest to params map
func SetupHooksRequestToParams(req *proto.SetupHooksRequest) map[string]interface{} {
	params := make(map[string]interface{})

	if req.Action != "" {
		params["action"] = req.Action
	}
	if len(req.Hooks) > 0 {
		hooksJSON, _ := json.Marshal(req.Hooks)
		params["hooks"] = string(hooksJSON)
	}
	if req.Patterns != "" {
		params["patterns"] = req.Patterns
	}
	if req.ConfigPath != "" {
		params["config_path"] = req.ConfigPath
	}
	params["install"] = req.Install
	params["dry_run"] = req.DryRun

	return params
}

// ParseCheckAttributionRequest parses a check_attribution request (protobuf or JSON)
func ParseCheckAttributionRequest(args json.RawMessage) (*proto.CheckAttributionRequest, map[string]interface{}, error) {
	var req proto.CheckAttributionRequest

	if err := protobuf.Unmarshal(args, &req); err == nil {
		return &req, nil, nil
	}

	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	return nil, params, nil
}

// CheckAttributionRequestToParams converts a protobuf CheckAttributionRequest to params map
func CheckAttributionRequestToParams(req *proto.CheckAttributionRequest) map[string]interface{} {
	params := make(map[string]interface{})

	if req.OutputPath != "" {
		params["output_path"] = req.OutputPath
	}
	params["create_tasks"] = req.CreateTasks

	return params
}

// ParseAddExternalToolHintsRequest parses an add_external_tool_hints request (protobuf or JSON)
func ParseAddExternalToolHintsRequest(args json.RawMessage) (*proto.AddExternalToolHintsRequest, map[string]interface{}, error) {
	var req proto.AddExternalToolHintsRequest

	if err := protobuf.Unmarshal(args, &req); err == nil {
		return &req, nil, nil
	}

	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	return nil, params, nil
}

// AddExternalToolHintsRequestToParams converts a protobuf AddExternalToolHintsRequest to params map
func AddExternalToolHintsRequestToParams(req *proto.AddExternalToolHintsRequest) map[string]interface{} {
	params := make(map[string]interface{})

	params["dry_run"] = req.DryRun
	if req.OutputPath != "" {
		params["output_path"] = req.OutputPath
	}
	if req.MinFileSize > 0 {
		params["min_file_size"] = int(req.MinFileSize)
	}

	return params
}
