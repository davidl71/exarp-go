package fixtures

import (
	"context"
	"encoding/json"
	"time"
)

// TestContext creates a context with timeout for testing
func TestContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// JSONRPCRequest builds a JSON-RPC 2.0 request
func JSONRPCRequest(method string, params interface{}) map[string]interface{} {
	return map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  method,
		"params":  params,
	}
}

// JSONRPCResponse builds a JSON-RPC 2.0 response
func JSONRPCResponse(id interface{}, result interface{}) map[string]interface{} {
	return map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	}
}

// JSONRPCError builds a JSON-RPC 2.0 error response
func JSONRPCError(id interface{}, code int, message string) map[string]interface{} {
	return map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}
}

// ToJSONRawMessage converts a value to json.RawMessage
func ToJSONRawMessage(v interface{}) (json.RawMessage, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// MustToJSONRawMessage converts a value to json.RawMessage, panics on error
func MustToJSONRawMessage(v interface{}) json.RawMessage {
	data, err := ToJSONRawMessage(v)
	if err != nil {
		panic(err)
	}
	return data
}
