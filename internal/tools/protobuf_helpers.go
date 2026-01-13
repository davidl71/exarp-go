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
