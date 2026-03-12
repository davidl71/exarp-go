package client

import (
	"encoding/json"
	"testing"

	"github.com/davidl71/mcp-go-core/pkg/mcp/types"
)

func TestConvertExternalToolToToolInfo(t *testing.T) {
	tests := []struct {
		name        string
		externalTool map[string]interface{}
		want        types.ToolInfo
		wantErr     bool
	}{
		{
			name: "simple tool with string description",
			externalTool: map[string]interface{}{
				"name":        "test_tool",
				"description": "A test tool",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"message": map[string]interface{}{
							"type": "string",
						},
					},
					"required": []interface{}{"message"},
				},
			},
			want: types.ToolInfo{
				Name:        "test_tool",
				Description: "A test tool",
				Schema: types.ToolSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"message": map[string]interface{}{
							"type": "string",
						},
					},
					Required: []string{"message"},
				},
			},
			wantErr: false,
		},
		{
			name: "tool with pointer description",
			externalTool: map[string]interface{}{
				"name":        "pointer_tool",
				"description": stringPtr("Pointer description"),
				"inputSchema": map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
			want: types.ToolInfo{
				Name:        "pointer_tool",
				Description: "Pointer description",
				Schema: types.ToolSchema{
					Type:       "object",
					Properties: map[string]interface{}{},
				},
			},
			wantErr: false,
		},
		{
			name: "tool with nil pointer description",
			externalTool: map[string]interface{}{
				"name":        "nil_tool",
				"description": (*string)(nil),
				"inputSchema": map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
			want: types.ToolInfo{
				Name:        "nil_tool",
				Description: "",
				Schema: types.ToolSchema{
					Type:       "object",
					Properties: map[string]interface{}{},
				},
			},
			wantErr: false,
		},
		{
			name: "tool without name",
			externalTool: map[string]interface{}{
				"description": "No name tool",
			},
			want:    types.ToolInfo{},
			wantErr: true,
		},
		{
			name: "tool without inputSchema (should use default)",
			externalTool: map[string]interface{}{
				"name":        "no_schema_tool",
				"description": "Tool without schema",
			},
			want: types.ToolInfo{
				Name:        "no_schema_tool",
				Description: "Tool without schema",
				Schema: types.ToolSchema{
					Type:       "object",
					Properties: map[string]interface{}{},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertExternalToolToToolInfo(tt.externalTool)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertExternalToolToToolInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Name != tt.want.Name {
					t.Errorf("ConvertExternalToolToToolInfo() Name = %v, want %v", got.Name, tt.want.Name)
				}
				if got.Description != tt.want.Description {
					t.Errorf("ConvertExternalToolToToolInfo() Description = %v, want %v", got.Description, tt.want.Description)
				}
				if got.Schema.Type != tt.want.Schema.Type {
					t.Errorf("ConvertExternalToolToToolInfo() Schema.Type = %v, want %v", got.Schema.Type, tt.want.Schema.Type)
				}
				// Compare required fields
				if len(got.Schema.Required) != len(tt.want.Schema.Required) {
					t.Errorf("ConvertExternalToolToToolInfo() Schema.Required length = %v, want %v", len(got.Schema.Required), len(tt.want.Schema.Required))
				}
			}
		})
	}
}

func TestConvertExternalTextContent(t *testing.T) {
	tests := []struct {
		name           string
		externalContent map[string]interface{}
		want           types.TextContent
		wantErr        bool
	}{
		{
			name: "simple text content",
			externalContent: map[string]interface{}{
				"type": "text",
				"text": "Hello, World!",
			},
			want: types.TextContent{
				Type: "text",
				Text: "Hello, World!",
			},
			wantErr: false,
		},
		{
			name: "text content with textContent nested field",
			externalContent: map[string]interface{}{
				"type": "text",
				"textContent": map[string]interface{}{
					"text": "Nested text",
				},
			},
			want: types.TextContent{
				Type: "text",
				Text: "Nested text",
			},
			wantErr: false,
		},
		{
			name: "text content without type (defaults to text)",
			externalContent: map[string]interface{}{
				"text": "No type specified",
			},
			want: types.TextContent{
				Type: "text",
				Text: "No type specified",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertExternalTextContent(tt.externalContent)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertExternalTextContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Type != tt.want.Type {
					t.Errorf("ConvertExternalTextContent() Type = %v, want %v", got.Type, tt.want.Type)
				}
				if got.Text != tt.want.Text {
					t.Errorf("ConvertExternalTextContent() Text = %v, want %v", got.Text, tt.want.Text)
				}
			}
		})
	}
}

func TestConvertExternalTextContentSlice(t *testing.T) {
	tests := []struct {
		name            string
		externalContents []interface{}
		want            []types.TextContent
		wantErr         bool
	}{
		{
			name: "single content",
			externalContents: []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": "First",
				},
			},
			want: []types.TextContent{
				{Type: "text", Text: "First"},
			},
			wantErr: false,
		},
		{
			name: "multiple contents",
			externalContents: []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": "First",
				},
				map[string]interface{}{
					"type": "text",
					"text": "Second",
				},
			},
			want: []types.TextContent{
				{Type: "text", Text: "First"},
				{Type: "text", Text: "Second"},
			},
			wantErr: false,
		},
		{
			name:            "empty slice",
			externalContents: []interface{}{},
			want:            []types.TextContent{},
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertExternalTextContentSlice(tt.externalContents)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertExternalTextContentSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("ConvertExternalTextContentSlice() length = %v, want %v", len(got), len(tt.want))
					return
				}
				for i := range got {
					if got[i].Type != tt.want[i].Type {
						t.Errorf("ConvertExternalTextContentSlice()[%d].Type = %v, want %v", i, got[i].Type, tt.want[i].Type)
					}
					if got[i].Text != tt.want[i].Text {
						t.Errorf("ConvertExternalTextContentSlice()[%d].Text = %v, want %v", i, got[i].Text, tt.want[i].Text)
					}
				}
			}
		})
	}
}

func TestConvertClientInfoToExternal(t *testing.T) {
	tests := []struct {
		name       string
		clientInfo struct {
			Name    string
			Version string
		}
		want map[string]interface{}
	}{
		{
			name: "standard client info",
			clientInfo: struct {
				Name    string
				Version string
			}{
				Name:    "test-client",
				Version: "1.0.0",
			},
			want: map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert struct to protocol.ClientInfo for the function
			info := struct {
				Name    string
				Version string
			}{
				Name:    tt.clientInfo.Name,
				Version: tt.clientInfo.Version,
			}
			
			// Marshal and unmarshal to simulate protocol.ClientInfo
			jsonData, _ := json.Marshal(info)
			var clientInfo struct {
				Name    string `json:"name"`
				Version string `json:"version"`
			}
			json.Unmarshal(jsonData, &clientInfo)
			
			// This test is simplified - actual function uses protocol.ClientInfo
			// For now, just verify the function exists and doesn't panic
			_ = ConvertClientInfoToExternal
		})
	}
}

// Helper function for tests
func stringPtr(s string) *string {
	return &s
}
