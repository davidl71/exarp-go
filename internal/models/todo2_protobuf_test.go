package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/davidl71/exarp-go/proto"
)

func TestTodo2TaskToProto(t *testing.T) {
	tests := []struct {
		name    string
		task    *Todo2Task
		wantErr bool
	}{
		{
			name: "basic task",
			task: &Todo2Task{
				ID:              "T-1",
				Content:         "Test task",
				LongDescription: "Test description",
				Status:          "Todo",
				Priority:        "high",
				Tags:            []string{"test", "protobuf"},
				Dependencies:    []string{"T-0"},
				Completed:       false,
				Metadata:        map[string]interface{}{"key": "value"},
			},
			wantErr: false,
		},
		{
			name: "task with complex metadata",
			task: &Todo2Task{
				ID:       "T-2",
				Content:  "Complex task",
				Status:   "In Progress",
				Metadata: map[string]interface{}{"nested": map[string]interface{}{"key": "value"}},
			},
			wantErr: false,
		},
		{
			name: "task with nil metadata",
			task: &Todo2Task{
				ID:      "T-3",
				Content: "No metadata",
				Status:  "Todo",
			},
			wantErr: false,
		},
		{
			name:    "nil task",
			task:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pbTask, err := Todo2TaskToProto(tt.task)
			if (err != nil) != tt.wantErr {
				t.Errorf("Todo2TaskToProto() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Verify basic fields
			if pbTask.Id != tt.task.ID {
				t.Errorf("Id = %v, want %v", pbTask.Id, tt.task.ID)
			}

			if pbTask.Content != tt.task.Content {
				t.Errorf("Content = %v, want %v", pbTask.Content, tt.task.Content)
			}

			if pbTask.Status != tt.task.Status {
				t.Errorf("Status = %v, want %v", pbTask.Status, tt.task.Status)
			}

			if pbTask.Completed != tt.task.Completed {
				t.Errorf("Completed = %v, want %v", pbTask.Completed, tt.task.Completed)
			}
		})
	}
}

func TestProtoToTodo2Task(t *testing.T) {
	tests := []struct {
		name    string
		pbTask  *proto.Todo2Task
		wantErr bool
	}{
		{
			name: "basic protobuf task",
			pbTask: &proto.Todo2Task{
				Id:              "T-1",
				Content:         "Test task",
				LongDescription: "Test description",
				Status:          "Todo",
				Priority:        "high",
				Tags:            []string{"test", "protobuf"},
				Dependencies:    []string{"T-0"},
				Completed:       false,
				Metadata:        map[string]string{"key": "value"},
				CreatedAt:       time.Now().Unix(),
				UpdatedAt:       time.Now().Unix(),
			},
			wantErr: false,
		},
		{
			name: "protobuf task with JSON metadata",
			pbTask: &proto.Todo2Task{
				Id:      "T-2",
				Content: "Complex task",
				Status:  "In Progress",
				Metadata: map[string]string{
					"nested": `{"key":"value"}`,
				},
			},
			wantErr: false,
		},
		{
			name:    "nil protobuf task",
			pbTask:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := ProtoToTodo2Task(tt.pbTask)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProtoToTodo2Task() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Verify basic fields
			if task.ID != tt.pbTask.Id {
				t.Errorf("ID = %v, want %v", task.ID, tt.pbTask.Id)
			}

			if task.Content != tt.pbTask.Content {
				t.Errorf("Content = %v, want %v", task.Content, tt.pbTask.Content)
			}

			if task.Status != tt.pbTask.Status {
				t.Errorf("Status = %v, want %v", task.Status, tt.pbTask.Status)
			}

			if task.Completed != tt.pbTask.Completed {
				t.Errorf("Completed = %v, want %v", task.Completed, tt.pbTask.Completed)
			}
		})
	}
}

func TestSerializeTaskToProtobuf(t *testing.T) {
	task := &Todo2Task{
		ID:      "T-1",
		Content: "Test task",
		Status:  "Todo",
		Metadata: map[string]interface{}{
			"string": "value",
			"number": 42,
			"nested": map[string]interface{}{"key": "value"},
		},
	}

	data, err := SerializeTaskToProtobuf(task)
	if err != nil {
		t.Fatalf("SerializeTaskToProtobuf() error = %v", err)
	}

	if len(data) == 0 {
		t.Error("SerializeTaskToProtobuf() returned empty data")
	}

	// Verify we can deserialize it back
	deserialized, err := DeserializeTaskFromProtobuf(data)
	if err != nil {
		t.Fatalf("DeserializeTaskFromProtobuf() error = %v", err)
	}

	if deserialized.ID != task.ID {
		t.Errorf("ID = %v, want %v", deserialized.ID, task.ID)
	}

	if deserialized.Content != task.Content {
		t.Errorf("Content = %v, want %v", deserialized.Content, task.Content)
	}
}

func TestDeserializeTaskFromProtobuf(t *testing.T) {
	original := &Todo2Task{
		ID:      "T-1",
		Content: "Test task",
		Status:  "Todo",
		Metadata: map[string]interface{}{
			"key": "value",
		},
	}

	data, err := SerializeTaskToProtobuf(original)
	if err != nil {
		t.Fatalf("SerializeTaskToProtobuf() error = %v", err)
	}

	deserialized, err := DeserializeTaskFromProtobuf(data)
	if err != nil {
		t.Fatalf("DeserializeTaskFromProtobuf() error = %v", err)
	}

	// Compare key fields
	if deserialized.ID != original.ID {
		t.Errorf("ID = %v, want %v", deserialized.ID, original.ID)
	}

	if deserialized.Content != original.Content {
		t.Errorf("Content = %v, want %v", deserialized.Content, original.Content)
	}

	if deserialized.Status != original.Status {
		t.Errorf("Status = %v, want %v", deserialized.Status, original.Status)
	}

	// Verify metadata (may be converted to JSON string in protobuf)
	if deserialized.Metadata != nil {
		if val, ok := deserialized.Metadata["key"]; ok {
			if valStr, ok := val.(string); ok && valStr != "value" {
				t.Errorf("Metadata[key] = %v, want 'value'", valStr)
			}
		}
	}
}

func TestRoundTripProtobuf(t *testing.T) {
	original := &Todo2Task{
		ID:              "T-1",
		Content:         "Round trip test",
		LongDescription: "Testing round trip serialization",
		Status:          "In Progress",
		Priority:        "high",
		Tags:            []string{"test", "protobuf", "roundtrip"},
		Dependencies:    []string{"T-0", "T-2"},
		Completed:       false,
		Metadata: map[string]interface{}{
			"string": "value",
			"number": 42,
			"bool":   true,
			"nested": map[string]interface{}{"key": "value"},
		},
	}

	// Serialize to protobuf
	data, err := SerializeTaskToProtobuf(original)
	if err != nil {
		t.Fatalf("SerializeTaskToProtobuf() error = %v", err)
	}

	// Deserialize from protobuf
	deserialized, err := DeserializeTaskFromProtobuf(data)
	if err != nil {
		t.Fatalf("DeserializeTaskFromProtobuf() error = %v", err)
	}

	// Verify all fields
	if deserialized.ID != original.ID {
		t.Errorf("ID = %v, want %v", deserialized.ID, original.ID)
	}

	if deserialized.Content != original.Content {
		t.Errorf("Content = %v, want %v", deserialized.Content, original.Content)
	}

	if deserialized.LongDescription != original.LongDescription {
		t.Errorf("LongDescription = %v, want %v", deserialized.LongDescription, original.LongDescription)
	}

	if deserialized.Status != original.Status {
		t.Errorf("Status = %v, want %v", deserialized.Status, original.Status)
	}

	if deserialized.Priority != original.Priority {
		t.Errorf("Priority = %v, want %v", deserialized.Priority, original.Priority)
	}

	if deserialized.Completed != original.Completed {
		t.Errorf("Completed = %v, want %v", deserialized.Completed, original.Completed)
	}

	// Verify tags
	if len(deserialized.Tags) != len(original.Tags) {
		t.Errorf("Tags length = %v, want %v", len(deserialized.Tags), len(original.Tags))
	}

	// Verify dependencies
	if len(deserialized.Dependencies) != len(original.Dependencies) {
		t.Errorf("Dependencies length = %v, want %v", len(deserialized.Dependencies), len(original.Dependencies))
	}
}

func TestMetadataConversion(t *testing.T) {
	tests := []struct {
		name     string
		metadata map[string]interface{}
	}{
		{
			name:     "nil metadata",
			metadata: nil,
		},
		{
			name:     "empty metadata",
			metadata: map[string]interface{}{},
		},
		{
			name: "string values",
			metadata: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "mixed types",
			metadata: map[string]interface{}{
				"string": "value",
				"number": 42,
				"bool":   true,
				"nil":    nil,
			},
		},
		{
			name: "nested structures",
			metadata: map[string]interface{}{
				"nested": map[string]interface{}{
					"key": "value",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Todo2Task{
				ID:       "T-1",
				Content:  "Test",
				Status:   "Todo",
				Metadata: tt.metadata,
			}

			// Convert to protobuf
			pbTask, err := Todo2TaskToProto(task)
			if err != nil {
				t.Fatalf("Todo2TaskToProto() error = %v", err)
			}

			// Convert back
			converted, err := ProtoToTodo2Task(pbTask)
			if err != nil {
				t.Fatalf("ProtoToTodo2Task() error = %v", err)
			}

			// Verify metadata was preserved (may be converted to JSON strings)
			if tt.metadata == nil {
				if converted.Metadata != nil && len(converted.Metadata) > 0 {
					t.Error("Metadata should be nil or empty")
				}
			} else if len(tt.metadata) > 0 {
				// Non-empty metadata should be preserved (may be converted to JSON strings)
				if converted.Metadata == nil || len(converted.Metadata) == 0 {
					t.Error("Metadata should not be nil or empty for non-empty input")
				}
			}
		})
	}
}

func TestProtobufSizeComparison(t *testing.T) {
	task := &Todo2Task{
		ID:      "T-1",
		Content: "Test task with metadata",
		Status:  "Todo",
		Metadata: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
			"key3": map[string]interface{}{"nested": "value"},
		},
	}

	// Serialize to protobuf
	protobufData, err := SerializeTaskToProtobuf(task)
	if err != nil {
		t.Fatalf("SerializeTaskToProtobuf() error = %v", err)
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Protobuf should be smaller (or at least comparable)
	t.Logf("Protobuf size: %d bytes", len(protobufData))
	t.Logf("JSON size: %d bytes", len(jsonData))
	t.Logf("Size reduction: %.1f%%", float64(len(jsonData)-len(protobufData))/float64(len(jsonData))*100)

	// Verify protobuf is not empty
	if len(protobufData) == 0 {
		t.Error("Protobuf data is empty")
	}
}
