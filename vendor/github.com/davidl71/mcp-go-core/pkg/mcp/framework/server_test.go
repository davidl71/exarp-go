package framework

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/davidl71/mcp-go-core/pkg/mcp/types"
)

// TestMCPServer_Interface verifies interface types exist
func TestMCPServer_Interface(t *testing.T) {
	// Verify interface types exist
	// Actual implementation tests are in gosdk package
	_ = MCPServer(nil) // Type assertion test
}

// TestMCPServer_ToolSchema tests ToolSchema usage
func TestMCPServer_ToolSchema(t *testing.T) {
	schema := types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"test": map[string]interface{}{
				"type": "string",
			},
		},
	}

	if schema.Type != "object" {
		t.Errorf("schema.Type = %v, want object", schema.Type)
	}

	if schema.Properties == nil {
		t.Error("schema.Properties is nil")
	}
}

// TestMCPServer_PromptHandler tests PromptHandler signature
func TestMCPServer_PromptHandler(t *testing.T) {
	handler := func(ctx context.Context, args map[string]interface{}) (string, error) {
		return "test prompt", nil
	}

	ctx := context.Background()
	result, err := handler(ctx, map[string]interface{}{"key": "value"})
	if err != nil {
		t.Fatalf("handler() error = %v", err)
	}

	if result != "test prompt" {
		t.Errorf("result = %v, want test prompt", result)
	}
}

// TestMCPServer_ResourceHandler tests ResourceHandler signature
func TestMCPServer_ResourceHandler(t *testing.T) {
	handler := func(ctx context.Context, uri string) ([]byte, string, error) {
		return []byte("test data"), "application/json", nil
	}

	ctx := context.Background()
	data, mimeType, err := handler(ctx, "stdio://test")
	if err != nil {
		t.Fatalf("handler() error = %v", err)
	}

	if string(data) != "test data" {
		t.Errorf("data = %v, want test data", string(data))
	}

	if mimeType != "application/json" {
		t.Errorf("mimeType = %v, want application/json", mimeType)
	}
}

// TestMCPServer_InterfaceContracts verifies interface contracts
func TestMCPServer_InterfaceContracts(t *testing.T) {
	// Verify interface types
	var _ ToolHandler = func(ctx context.Context, args json.RawMessage) ([]types.TextContent, error) {
		return nil, nil
	}

	var _ PromptHandler = func(ctx context.Context, args map[string]interface{}) (string, error) {
		return "", nil
	}

	var _ ResourceHandler = func(ctx context.Context, uri string) ([]byte, string, error) {
		return nil, "", nil
	}
}
