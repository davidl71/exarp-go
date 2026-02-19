// todo2_protobuf.go — Protobuf serialization/deserialization for Todo2Task.
package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/davidl71/exarp-go/proto"
	"google.golang.org/protobuf/encoding/protojson"
	protobuf "google.golang.org/protobuf/proto"
)

// Todo2TaskToProto converts a models.Todo2Task to protobuf Todo2Task.
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

// ProtoToTodo2Task converts a protobuf Todo2Task to models.Todo2Task.
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
		ProjectID:       pbTask.ProjectId,
		AssignedTo:      pbTask.AssignedTo,
		Host:            pbTask.Host,
		Agent:           pbTask.Agent,
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
