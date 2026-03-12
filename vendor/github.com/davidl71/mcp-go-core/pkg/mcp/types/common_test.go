package types

import (
	"encoding/json"
	"testing"
)

func TestTextContent_JSON(t *testing.T) {
	tests := []struct {
		name    string
		content TextContent
		want    string
	}{
		{
			name: "basic text content",
			content: TextContent{
				Type: "text",
				Text: "Hello, world!",
			},
			want: `{"type":"text","text":"Hello, world!"}`,
		},
		{
			name: "empty text",
			content: TextContent{
				Type: "text",
				Text: "",
			},
			want: `{"type":"text","text":""}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.content)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}

			if string(got) != tt.want {
				t.Errorf("json.Marshal() = %v, want %v", string(got), tt.want)
			}

			// Test unmarshaling
			var unmarshaled TextContent
			if err := json.Unmarshal(got, &unmarshaled); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			if unmarshaled.Type != tt.content.Type {
				t.Errorf("unmarshaled.Type = %v, want %v", unmarshaled.Type, tt.content.Type)
			}
			if unmarshaled.Text != tt.content.Text {
				t.Errorf("unmarshaled.Text = %v, want %v", unmarshaled.Text, tt.content.Text)
			}
		})
	}
}

func TestToolSchema_JSON(t *testing.T) {
	schema := ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "The name",
			},
			"count": map[string]interface{}{
				"type":    "number",
				"minimum": 0,
			},
		},
		Required: []string{"name"},
	}

	got, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Verify it's valid JSON
	var unmarshaled ToolSchema
	if err := json.Unmarshal(got, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Type != schema.Type {
		t.Errorf("unmarshaled.Type = %v, want %v", unmarshaled.Type, schema.Type)
	}

	if len(unmarshaled.Required) != 1 || unmarshaled.Required[0] != "name" {
		t.Errorf("unmarshaled.Required = %v, want [name]", unmarshaled.Required)
	}
}

func TestToolInfo(t *testing.T) {
	info := ToolInfo{
		Name:        "test_tool",
		Description: "A test tool",
		Schema: ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"param": map[string]interface{}{
					"type": "string",
				},
			},
		},
	}

	if info.Name != "test_tool" {
		t.Errorf("info.Name = %v, want test_tool", info.Name)
	}

	if info.Description != "A test tool" {
		t.Errorf("info.Description = %v, want A test tool", info.Description)
	}

	if info.Schema.Type != "object" {
		t.Errorf("info.Schema.Type = %v, want object", info.Schema.Type)
	}
}
