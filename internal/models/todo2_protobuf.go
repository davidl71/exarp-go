package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/davidl71/exarp-go/proto"
	protobuf "google.golang.org/protobuf/proto"
)

// Todo2TaskToProto converts a models.Todo2Task to protobuf Todo2Task
func Todo2TaskToProto(task *Todo2Task) (*proto.Todo2Task, error) {
	if task == nil {
		return nil, fmt.Errorf("task is nil")
	}

	pbTask := &proto.Todo2Task{
		Id:              task.ID,
		Content:         task.Content,
		LongDescription: task.LongDescription,
		Status:          task.Status,
		Priority:        task.Priority,
		Tags:            task.Tags,
		Dependencies:    task.Dependencies,
		Completed:       task.Completed,
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

	// Set timestamps (use current time if not set)
	now := time.Now().Unix()
	if pbTask.CreatedAt == 0 {
		pbTask.CreatedAt = now
	}
	if pbTask.UpdatedAt == 0 {
		pbTask.UpdatedAt = now
	}

	return pbTask, nil
}

// ProtoToTodo2Task converts a protobuf Todo2Task to models.Todo2Task
func ProtoToTodo2Task(pbTask *proto.Todo2Task) (*Todo2Task, error) {
	if pbTask == nil {
		return nil, fmt.Errorf("protobuf task is nil")
	}

	task := &Todo2Task{
		ID:              pbTask.Id,
		Content:         pbTask.Content,
		LongDescription: pbTask.LongDescription,
		Status:          pbTask.Status,
		Priority:        pbTask.Priority,
		Tags:            pbTask.Tags,
		Dependencies:    pbTask.Dependencies,
		Completed:       pbTask.Completed,
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

// SerializeTaskToProtobuf serializes a Todo2Task to protobuf binary format
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

// DeserializeTaskFromProtobuf deserializes a Todo2Task from protobuf binary format
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

// SerializeTaskToProtobufJSON serializes a Todo2Task to protobuf JSON format
// This is useful for debugging and human-readable storage
func SerializeTaskToProtobufJSON(task *Todo2Task) ([]byte, error) {
	pbTask, err := Todo2TaskToProto(task)
	if err != nil {
		return nil, fmt.Errorf("failed to convert task to protobuf: %w", err)
	}

	// Use protojson for JSON serialization
	// Note: This requires importing encoding/protojson
	// For now, we'll use standard JSON as fallback
	// TODO: Use protojson.Marshal for proper protobuf JSON format
	return json.Marshal(pbTask)
}

// DeserializeTaskFromProtobufJSON deserializes a Todo2Task from protobuf JSON format
func DeserializeTaskFromProtobufJSON(data []byte) (*Todo2Task, error) {
	pbTask := &proto.Todo2Task{}
	if err := json.Unmarshal(data, pbTask); err != nil {
		return nil, fmt.Errorf("failed to unmarshal protobuf JSON: %w", err)
	}

	task, err := ProtoToTodo2Task(pbTask)
	if err != nil {
		return nil, fmt.Errorf("failed to convert protobuf to task: %w", err)
	}

	return task, nil
}
