package protocol

import (
	"encoding/json"
	"testing"
)

func TestJSONRPCRequest_Marshal(t *testing.T) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      "123",
		Method:  "tools/list",
		Params:  json.RawMessage(`{"key":"value"}`),
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	var unmarshaled JSONRPCRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	if unmarshaled.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC 2.0, got %s", unmarshaled.JSONRPC)
	}
	if unmarshaled.ID != "123" {
		t.Errorf("Expected ID 123, got %v", unmarshaled.ID)
	}
	if unmarshaled.Method != "tools/list" {
		t.Errorf("Expected method tools/list, got %s", unmarshaled.Method)
	}
}

func TestJSONRPCResponse_Success(t *testing.T) {
	resp := NewSuccessResponse("123", map[string]interface{}{"result": "success"})

	if resp.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC 2.0, got %s", resp.JSONRPC)
	}
	if resp.ID != "123" {
		t.Errorf("Expected ID 123, got %v", resp.ID)
	}
	if resp.Error != nil {
		t.Error("Success response should not have error")
	}
	if resp.Result == nil {
		t.Error("Success response should have result")
	}
}

func TestJSONRPCResponse_Error(t *testing.T) {
	resp := NewErrorResponse("123", ErrCodeMethodNotFound, "Method not found", nil)

	if resp.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC 2.0, got %s", resp.JSONRPC)
	}
	if resp.ID != "123" {
		t.Errorf("Expected ID 123, got %v", resp.ID)
	}
	if resp.Error == nil {
		t.Error("Error response should have error")
	}
	if resp.Error.Code != ErrCodeMethodNotFound {
		t.Errorf("Expected error code %d, got %d", ErrCodeMethodNotFound, resp.Error.Code)
	}
	if resp.Result != nil {
		t.Error("Error response should not have result")
	}
}

func TestNewMethodNotFoundError(t *testing.T) {
	resp := NewMethodNotFoundError("123", "invalid_method")

	if resp.Error == nil {
		t.Fatal("Expected error response")
	}
	if resp.Error.Code != ErrCodeMethodNotFound {
		t.Errorf("Expected error code %d, got %d", ErrCodeMethodNotFound, resp.Error.Code)
	}
	if resp.Error.Message != "Method not found: invalid_method" {
		t.Errorf("Expected error message 'Method not found: invalid_method', got %s", resp.Error.Message)
	}
}

func TestNewInvalidParamsError(t *testing.T) {
	resp := NewInvalidParamsError("123", "Invalid parameter")

	if resp.Error == nil {
		t.Fatal("Expected error response")
	}
	if resp.Error.Code != ErrCodeInvalidParams {
		t.Errorf("Expected error code %d, got %d", ErrCodeInvalidParams, resp.Error.Code)
	}
	if resp.Error.Message != "Invalid parameter" {
		t.Errorf("Expected error message 'Invalid parameter', got %s", resp.Error.Message)
	}
}

func TestNewInternalError(t *testing.T) {
	resp := NewInternalError("123", "Internal server error")

	if resp.Error == nil {
		t.Fatal("Expected error response")
	}
	if resp.Error.Code != ErrCodeInternalError {
		t.Errorf("Expected error code %d, got %d", ErrCodeInternalError, resp.Error.Code)
	}
	if resp.Error.Message != "Internal server error" {
		t.Errorf("Expected error message 'Internal server error', got %s", resp.Error.Message)
	}
}

func TestInitializeParams_Marshal(t *testing.T) {
	params := InitializeParams{
		ProtocolVersion: "2024-11-05",
		Capabilities: ClientCapabilities{
			Experimental: map[string]interface{}{"feature": true},
		},
		ClientInfo: ClientInfo{
			Name:    "test-client",
			Version: "1.0.0",
		},
	}

	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Failed to marshal InitializeParams: %v", err)
	}

	var unmarshaled InitializeParams
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal InitializeParams: %v", err)
	}

	if unmarshaled.ProtocolVersion != "2024-11-05" {
		t.Errorf("Expected protocol version 2024-11-05, got %s", unmarshaled.ProtocolVersion)
	}
	if unmarshaled.ClientInfo.Name != "test-client" {
		t.Errorf("Expected client name test-client, got %s", unmarshaled.ClientInfo.Name)
	}
}

func TestToolCallParams_Marshal(t *testing.T) {
	params := ToolCallParams{
		Name:      "test_tool",
		Arguments: map[string]interface{}{"arg1": "value1"},
	}

	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Failed to marshal ToolCallParams: %v", err)
	}

	var unmarshaled ToolCallParams
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal ToolCallParams: %v", err)
	}

	if unmarshaled.Name != "test_tool" {
		t.Errorf("Expected tool name test_tool, got %s", unmarshaled.Name)
	}
	if unmarshaled.Arguments["arg1"] != "value1" {
		t.Errorf("Expected argument value1, got %v", unmarshaled.Arguments["arg1"])
	}
}

func TestFormatRequestID(t *testing.T) {
	tests := []struct {
		name     string
		id       interface{}
		expected string
	}{
		{"nil", nil, "null"},
		{"string", "req-1", "req-1"},
		{"int", 42, "42"},
		{"int64", int64(99), "99"},
		{"float64 integer", float64(46), "46"},
		{"float64 truncated", 3.14, "3"}, // %.0f for non-integer float64
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatRequestID(tt.id)
			if got != tt.expected {
				t.Errorf("FormatRequestID(%v) = %q, want %q", tt.id, got, tt.expected)
			}
		})
	}
}
