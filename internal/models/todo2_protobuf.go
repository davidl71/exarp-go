// todo2_protobuf.go — Protobuf serialization/deserialization for Todo2Task.
package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/proto"
	"google.golang.org/protobuf/encoding/protojson"
	protobuf "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func todo2StatusStringToEnum(status string) proto.Todo2TaskStatus {
	s := strings.TrimSpace(strings.ToLower(status))
	if s == "" {
		return proto.Todo2TaskStatus_TODO2_TASK_STATUS_UNSPECIFIED
	}
	switch s {
	case "todo", "pending", "not started", "new":
		return proto.Todo2TaskStatus_TODO2_TASK_STATUS_TODO
	case "in progress", "in_progress", "in-progress", "working", "active", "inprogress":
		return proto.Todo2TaskStatus_TODO2_TASK_STATUS_IN_PROGRESS
	case "review", "needs review", "awaiting review":
		return proto.Todo2TaskStatus_TODO2_TASK_STATUS_REVIEW
	case "done", "completed", "finished", "closed":
		return proto.Todo2TaskStatus_TODO2_TASK_STATUS_DONE
	case "blocked", "waiting":
		return proto.Todo2TaskStatus_TODO2_TASK_STATUS_BLOCKED
	case "cancelled", "canceled", "abandoned":
		return proto.Todo2TaskStatus_TODO2_TASK_STATUS_CANCELLED
	default:
		return proto.Todo2TaskStatus_TODO2_TASK_STATUS_UNSPECIFIED
	}
}

func todo2StatusEnumToTitle(status proto.Todo2TaskStatus) string {
	switch status {
	case proto.Todo2TaskStatus_TODO2_TASK_STATUS_TODO:
		return StatusTodo
	case proto.Todo2TaskStatus_TODO2_TASK_STATUS_IN_PROGRESS:
		return StatusInProgress
	case proto.Todo2TaskStatus_TODO2_TASK_STATUS_REVIEW:
		return StatusReview
	case proto.Todo2TaskStatus_TODO2_TASK_STATUS_DONE:
		return StatusDone
	case proto.Todo2TaskStatus_TODO2_TASK_STATUS_BLOCKED:
		return StatusBlocked
	case proto.Todo2TaskStatus_TODO2_TASK_STATUS_CANCELLED:
		return StatusCancelled
	default:
		return ""
	}
}

func todo2PriorityStringToEnum(priority string) proto.Todo2TaskPriority {
	s := strings.TrimSpace(strings.ToLower(priority))
	if s == "" {
		return proto.Todo2TaskPriority_TODO2_TASK_PRIORITY_UNSPECIFIED
	}
	switch s {
	case "low", "lowest":
		return proto.Todo2TaskPriority_TODO2_TASK_PRIORITY_LOW
	case "medium", "normal", "standard":
		return proto.Todo2TaskPriority_TODO2_TASK_PRIORITY_MEDIUM
	case "high":
		return proto.Todo2TaskPriority_TODO2_TASK_PRIORITY_HIGH
	case "critical", "urgent", "highest":
		return proto.Todo2TaskPriority_TODO2_TASK_PRIORITY_CRITICAL
	default:
		return proto.Todo2TaskPriority_TODO2_TASK_PRIORITY_UNSPECIFIED
	}
}

func todo2PriorityEnumToCanonical(priority proto.Todo2TaskPriority) string {
	switch priority {
	case proto.Todo2TaskPriority_TODO2_TASK_PRIORITY_LOW:
		return PriorityLow
	case proto.Todo2TaskPriority_TODO2_TASK_PRIORITY_MEDIUM:
		return PriorityMedium
	case proto.Todo2TaskPriority_TODO2_TASK_PRIORITY_HIGH:
		return PriorityHigh
	case proto.Todo2TaskPriority_TODO2_TASK_PRIORITY_CRITICAL:
		return PriorityCritical
	default:
		return ""
	}
}

// Todo2TaskToProto converts a models.Todo2Task to protobuf Todo2Task.
func Todo2TaskToProto(task *Todo2Task) (*proto.Todo2Task, error) {
	if task == nil {
		return nil, fmt.Errorf("task is nil")
	}

	statusEnum := todo2StatusStringToEnum(task.Status)
	priorityEnum := todo2PriorityStringToEnum(task.Priority)

	// Keep internal typed fields in sync for callers that rely on them.
	task.StatusEnum = ParseTaskStatus(task.Status)
	task.PriorityEnum = ParseTaskPriority(task.Priority)

	pbTask := &proto.Todo2Task{
		Id:              task.ID,
		Name:            task.Name,
		Content:         task.Content,
		LongDescription: task.LongDescription,
		Status:          task.Status,
		StatusEnum:      statusEnum,
		Priority:        task.Priority,
		PriorityEnum:    priorityEnum,
		Tags:            task.Tags,
		Dependencies:    task.Dependencies,
		Completed:       task.Completed,
		ProjectId:       task.ProjectID,
		AssignedTo:      task.AssignedTo,
		Host:            task.Host,
		Agent:           task.Agent,
	}

	// Convert metadata from map[string]interface{} to map[string]string
	// Complex values are serialized to JSON strings
	if task.Metadata != nil && len(task.Metadata) > 0 {
		pbTask.Metadata = make(map[string]string, len(task.Metadata))

		for k, v := range task.Metadata {
			switch val := v.(type) {
			case string:
				pbTask.Metadata[k] = val
			case nil:
				pbTask.Metadata[k] = ""
			default:
				// Serialize complex types to JSON
				jsonBytes, err := json.Marshal(val)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal metadata value for key %s: %w", k, err)
				}

				pbTask.Metadata[k] = string(jsonBytes)
			}
		}
	}

	// Set timestamps (use current time if missing/unparseable)
	now := timestamppb.Now()

	task.EnsureName()
	if pbTask.Name == "" {
		pbTask.Name = task.Name
	}

	// Keep legacy string fields populated for JSON/MCP edge consumers.
	if pbTask.Status == "" {
		if title := todo2StatusEnumToTitle(statusEnum); title != "" {
			pbTask.Status = title
		} else {
			pbTask.Status = StatusTodo
		}
	}
	if pbTask.Priority == "" {
		if canon := todo2PriorityEnumToCanonical(priorityEnum); canon != "" {
			pbTask.Priority = canon
		} else {
			pbTask.Priority = PriorityMedium
		}
	}

	pbTask.CreatedAt = parseRFC3339ToTimestamp(task.CreatedAt)
	if pbTask.CreatedAt == nil {
		pbTask.CreatedAt = now
	}

	pbTask.UpdatedAt = parseRFC3339ToTimestamp(task.LastModified)
	if pbTask.UpdatedAt == nil {
		pbTask.UpdatedAt = now
	}

	return pbTask, nil
}

// ProtoToTodo2Task converts a protobuf Todo2Task to models.Todo2Task.
func ProtoToTodo2Task(pbTask *proto.Todo2Task) (*Todo2Task, error) {
	if pbTask == nil {
		return nil, fmt.Errorf("protobuf task is nil")
	}

	status := pbTask.Status
	if pbTask.StatusEnum != proto.Todo2TaskStatus_TODO2_TASK_STATUS_UNSPECIFIED {
		if title := todo2StatusEnumToTitle(pbTask.StatusEnum); title != "" {
			status = title
		}
	}

	priority := pbTask.Priority
	if pbTask.PriorityEnum != proto.Todo2TaskPriority_TODO2_TASK_PRIORITY_UNSPECIFIED {
		if canon := todo2PriorityEnumToCanonical(pbTask.PriorityEnum); canon != "" {
			priority = canon
		}
	}

	task := &Todo2Task{
		ID:              pbTask.Id,
		Name:            pbTask.Name,
		Content:         pbTask.Content,
		LongDescription: pbTask.LongDescription,
		Status:          status,
		StatusEnum:      ParseTaskStatus(status),
		Priority:        priority,
		PriorityEnum:    ParseTaskPriority(priority),
		Tags:            pbTask.Tags,
		Dependencies:    pbTask.Dependencies,
		Completed:       pbTask.Completed,
		ProjectID:       pbTask.ProjectId,
		AssignedTo:      pbTask.AssignedTo,
		Host:            pbTask.Host,
		Agent:           pbTask.Agent,
		CreatedAt:       timestampToRFC3339(pbTask.CreatedAt),
		LastModified:    timestampToRFC3339(pbTask.UpdatedAt),
	}

	// Convert metadata from map[string]string to map[string]interface{}
	// Try to deserialize JSON strings back to their original types
	if pbTask.Metadata != nil && len(pbTask.Metadata) > 0 {
		task.Metadata = make(map[string]interface{}, len(pbTask.Metadata))

		for k, v := range pbTask.Metadata {
			// Try to parse as JSON first
			var jsonVal interface{}
			if err := json.Unmarshal([]byte(v), &jsonVal); err == nil {
				// Successfully parsed as JSON - use the parsed value
				task.Metadata[k] = jsonVal
			} else {
				// Not JSON or parse failed - treat as plain string
				task.Metadata[k] = v
			}
		}
	}

	return task, nil
}

func parseRFC3339ToTimestamp(s string) *timestamppb.Timestamp {
	s = strings.TrimSpace(s)
	if s == "" || IsEpochDate(s) {
		return nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil
	}
	return timestamppb.New(t)
}

func timestampToRFC3339(pb *timestamppb.Timestamp) string {
	if pb == nil {
		return ""
	}
	t := pb.AsTime().UTC()
	// If proto had zero value, treat as unset.
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

// SerializeTaskToProtobuf serializes a Todo2Task to protobuf binary format.
func SerializeTaskToProtobuf(task *Todo2Task) ([]byte, error) {
	pbTask, err := Todo2TaskToProto(task)
	if err != nil {
		return nil, fmt.Errorf("failed to convert task to protobuf: %w", err)
	}

	data, err := protobuf.Marshal(pbTask)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal protobuf: %w", err)
	}

	return data, nil
}

// DeserializeTaskFromProtobuf deserializes a Todo2Task from protobuf binary format.
func DeserializeTaskFromProtobuf(data []byte) (*Todo2Task, error) {
	pbTask := &proto.Todo2Task{}
	if err := protobuf.Unmarshal(data, pbTask); err != nil {
		return nil, fmt.Errorf("failed to unmarshal protobuf: %w", err)
	}

	task, err := ProtoToTodo2Task(pbTask)
	if err != nil {
		return nil, fmt.Errorf("failed to convert protobuf to task: %w", err)
	}

	return task, nil
}

// SerializeTaskToProtobufJSON serializes a Todo2Task to protobuf JSON format using protojson.
// Produces canonical proto3 JSON (camelCase field names, proper enum/timestamp handling).
func SerializeTaskToProtobufJSON(task *Todo2Task) ([]byte, error) {
	pbTask, err := Todo2TaskToProto(task)
	if err != nil {
		return nil, fmt.Errorf("failed to convert task to protobuf: %w", err)
	}

	opts := protojson.MarshalOptions{
		EmitUnpopulated: false,
	}
	return opts.Marshal(pbTask)
}

// DeserializeTaskFromProtobufJSON deserializes a Todo2Task from protobuf JSON format using protojson.
// Accepts canonical proto3 JSON (camelCase) and the original proto field names.
func DeserializeTaskFromProtobufJSON(data []byte) (*Todo2Task, error) {
	pbTask := &proto.Todo2Task{}
	opts := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
	if err := opts.Unmarshal(data, pbTask); err != nil {
		return nil, fmt.Errorf("failed to unmarshal protobuf JSON: %w", err)
	}

	task, err := ProtoToTodo2Task(pbTask)
	if err != nil {
		return nil, fmt.Errorf("failed to convert protobuf to task: %w", err)
	}

	return task, nil
}
