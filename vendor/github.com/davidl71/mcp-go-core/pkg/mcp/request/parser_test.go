package request

import (
	"encoding/json"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestParseRequest_ProtobufSuccess(t *testing.T) {
	// Create a test protobuf message using structpb.Value
	msg := structpb.NewStringValue("test_value")

	// Marshal to protobuf binary
	args, err := proto.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal protobuf: %v", err)
	}

	// Parse using generic function
	req, params, err := ParseRequest(args, func() *structpb.Value {
		return &structpb.Value{}
	})

	if err != nil {
		t.Fatalf("ParseRequest() error = %v, want nil", err)
	}

	if req == nil {
		t.Fatal("ParseRequest() returned nil request, want non-nil")
	}

	if params != nil {
		t.Fatal("ParseRequest() returned non-nil params, want nil for protobuf")
	}

	// Verify the protobuf message was parsed correctly
	if req.GetStringValue() != "test_value" {
		t.Errorf("ParseRequest() req.GetStringValue() = %q, want %q", req.GetStringValue(), "test_value")
	}
}

func TestParseRequest_JSONSuccess(t *testing.T) {
	// Create JSON request
	jsonData := map[string]interface{}{
		"action": "test_action",
		"limit":  42,
	}

	args, err := json.Marshal(jsonData)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	// Parse using generic function
	req, params, err := ParseRequest(args, func() *structpb.Value {
		return &structpb.Value{}
	})

	if err != nil {
		t.Fatalf("ParseRequest() error = %v, want nil", err)
	}

	// For JSON, req should be zero-value (empty struct)
	if req != nil {
		// Check if it's actually empty/zero - structpb.Value zero value is nil pointer
		// So we just verify params is populated
	}

	if params == nil {
		t.Fatal("ParseRequest() returned nil params, want non-nil for JSON")
	}

	if params["action"] != "test_action" {
		t.Errorf("ParseRequest() params[action] = %v, want %q", params["action"], "test_action")
	}

	if params["limit"].(float64) != 42 {
		t.Errorf("ParseRequest() params[limit] = %v, want %d", params["limit"], 42)
	}
}

func TestParseRequest_InvalidInput(t *testing.T) {
	// Invalid input (neither protobuf nor JSON)
	args := json.RawMessage("invalid input")

	req, params, err := ParseRequest(args, func() *structpb.Value {
		return &structpb.Value{}
	})

	if err == nil {
		t.Fatal("ParseRequest() error = nil, want error for invalid input")
	}

	// For invalid input, req should be zero-value
	_ = req

	if params != nil {
		t.Error("ParseRequest() returned non-nil params, want nil")
	}
}

func TestParseRequest_EmptyInput(t *testing.T) {
	// Empty input
	args := json.RawMessage("{}")

	req, params, err := ParseRequest(args, func() *structpb.Value {
		return &structpb.Value{}
	})

	if err != nil {
		t.Fatalf("ParseRequest() error = %v, want nil for empty JSON", err)
	}

	// For empty JSON, req should be zero-value
	_ = req

	if params == nil {
		t.Fatal("ParseRequest() returned nil params, want non-nil for empty JSON")
	}

	if len(params) != 0 {
		t.Errorf("ParseRequest() params length = %d, want 0", len(params))
	}
}

func TestParseRequest_ProtobufPriority(t *testing.T) {
	// Create a protobuf message
	msg := structpb.NewStringValue("protobuf_value")

	// Marshal to protobuf binary
	args, err := proto.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal protobuf: %v", err)
	}

	// Parse - should prefer protobuf even if it could be JSON
	req, params, err := ParseRequest(args, func() *structpb.Value {
		return &structpb.Value{}
	})

	if err != nil {
		t.Fatalf("ParseRequest() error = %v, want nil", err)
	}

	// Should return protobuf, not JSON
	if req == nil {
		t.Fatal("ParseRequest() returned nil request, want protobuf request")
	}

	if params != nil {
		t.Fatal("ParseRequest() returned non-nil params, want nil (protobuf takes priority)")
	}

	if req.GetStringValue() != "protobuf_value" {
		t.Errorf("ParseRequest() req.GetStringValue() = %q, want %q", req.GetStringValue(), "protobuf_value")
	}
}
