package gosdk

import (
	"testing"

	"github.com/davidl71/mcp-go-core/pkg/mcp/types"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestTextContentToMCP(t *testing.T) {
	tests := []struct {
		name     string
		contents []types.TextContent
		wantLen  int
	}{
		{
			name: "single content",
			contents: []types.TextContent{
				{Type: "text", Text: "Hello, world!"},
			},
			wantLen: 1,
		},
		{
			name: "multiple contents",
			contents: []types.TextContent{
				{Type: "text", Text: "First"},
				{Type: "text", Text: "Second"},
				{Type: "text", Text: "Third"},
			},
			wantLen: 3,
		},
		{
			name:     "empty slice",
			contents: []types.TextContent{},
			wantLen:  0,
		},
		{
			name: "empty text",
			contents: []types.TextContent{
				{Type: "text", Text: ""},
			},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TextContentToMCP(tt.contents)
			if len(result) != tt.wantLen {
				t.Errorf("TextContentToMCP() len = %d, want %d", len(result), tt.wantLen)
				return
			}

			// Verify each content is properly converted
			for i, content := range tt.contents {
				mcpContent, ok := result[i].(*mcp.TextContent)
				if !ok {
					t.Errorf("TextContentToMCP() result[%d] is not *mcp.TextContent", i)
					return
				}
				if mcpContent.Text != content.Text {
					t.Errorf("TextContentToMCP() result[%d].Text = %q, want %q", i, mcpContent.Text, content.Text)
				}
			}
		})
	}
}

func TestToolSchemaToMCP(t *testing.T) {
	tests := []struct {
		name   string
		schema types.ToolSchema
		want   map[string]interface{}
	}{
		{
			name: "schema with required fields",
			schema: types.ToolSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"name": map[string]interface{}{
						"type": "string",
					},
				},
				Required: []string{"name"},
			},
			want: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type": "string",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			name: "schema without required fields",
			schema: types.ToolSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"optional": map[string]interface{}{
						"type": "string",
					},
				},
				Required: []string{},
			},
			want: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{
					"optional": map[string]interface{}{
						"type": "string",
					},
				},
			},
		},
		{
			name: "empty schema",
			schema: types.ToolSchema{
				Type:       "object",
				Properties: map[string]interface{}{},
				Required:   []string{},
			},
			want: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToolSchemaToMCP(tt.schema)

			// Check type
			if result["type"] != tt.want["type"] {
				t.Errorf("ToolSchemaToMCP() type = %v, want %v", result["type"], tt.want["type"])
			}

			// Check properties
			if result["properties"] == nil {
				t.Error("ToolSchemaToMCP() properties is nil")
			}

			// Check required (only if present in want)
			if required, ok := tt.want["required"]; ok {
				if result["required"] == nil {
					t.Error("ToolSchemaToMCP() required is nil, but expected")
				}
				resultRequired, ok := result["required"].([]string)
				if !ok {
					t.Error("ToolSchemaToMCP() required is not []string")
					return
				}
				wantRequired := required.([]string)
				if len(resultRequired) != len(wantRequired) {
					t.Errorf("ToolSchemaToMCP() required len = %d, want %d", len(resultRequired), len(wantRequired))
				}
			} else {
				// If required is not in want, it should not be in result (unless empty)
				if required, ok := result["required"]; ok && len(required.([]string)) > 0 {
					t.Error("ToolSchemaToMCP() required is present but should not be")
				}
			}
		})
	}
}
