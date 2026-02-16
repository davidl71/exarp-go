package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
)

func TestHandleToolCatalogNative(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		params         map[string]interface{}
		wantErr        bool
		validateResult func(t *testing.T, result []framework.TextContent)
	}{
		{
			name: "help action with valid tool",
			params: map[string]interface{}{
				"action":    "help",
				"tool_name": "task_workflow",
			},
			wantErr: false,
			validateResult: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("handleToolCatalogNative() returned empty result")
					return
				}

				var help map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &help); err != nil {
					t.Errorf("failed to unmarshal result: %v", err)
					return
				}

				if help["tool"] != "task_workflow" {
					t.Errorf("expected tool 'task_workflow', got %v", help["tool"])
				}

				if help["hint"] == nil {
					t.Error("expected hint field in result")
				}

				if help["category"] == nil {
					t.Error("expected category field in result")
				}
			},
		},
		{
			name: "help action without tool_name",
			params: map[string]interface{}{
				"action": "help",
			},
			wantErr: false,
			validateResult: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("handleToolCatalogNative() returned empty result")
					return
				}

				var response map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &response); err != nil {
					t.Errorf("failed to unmarshal result: %v", err)
					return
				}

				if response["status"] != "error" {
					t.Errorf("expected status 'error', got %v", response["status"])
				}

				if response["error"] == nil {
					t.Error("expected error field in result")
				}
			},
		},
		{
			name: "help action with non-existent tool",
			params: map[string]interface{}{
				"action":    "help",
				"tool_name": "non_existent_tool",
			},
			wantErr: false,
			validateResult: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("handleToolCatalogNative() returned empty result")
					return
				}

				var response map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &response); err != nil {
					t.Errorf("failed to unmarshal result: %v", err)
					return
				}

				if response["status"] != "error" {
					t.Errorf("expected status 'error', got %v", response["status"])
				}
			},
		},
		{
			name: "default action (help)",
			params: map[string]interface{}{
				"tool_name": "session",
			},
			wantErr: false,
			validateResult: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("handleToolCatalogNative() returned empty result")
					return
				}

				var help map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &help); err != nil {
					t.Errorf("failed to unmarshal result: %v", err)
					return
				}

				if help["tool"] != "session" {
					t.Errorf("expected tool 'session', got %v", help["tool"])
				}
			},
		},
		{
			name: "unknown action",
			params: map[string]interface{}{
				"action": "unknown",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handleToolCatalogNative(ctx, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleToolCatalogNative() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

func TestGetToolCatalog(t *testing.T) {
	catalog := GetToolCatalog()

	if len(catalog) == 0 {
		t.Error("GetToolCatalog() returned empty catalog")
	}

	// Check that some expected tools exist
	expectedTools := []string{
		"task_workflow",
		"session",
		"health",
		"project_scorecard",
	}

	for _, toolName := range expectedTools {
		if _, exists := catalog[toolName]; !exists {
			t.Errorf("expected tool '%s' in catalog, but not found", toolName)
		}
	}

	// Check that catalog entries have required fields
	for toolID, entry := range catalog {
		if entry.Tool == "" {
			t.Errorf("catalog entry '%s' has empty Tool field", toolID)
		}

		if entry.Hint == "" {
			t.Errorf("catalog entry '%s' has empty Hint field", toolID)
		}

		if entry.Category == "" {
			t.Errorf("catalog entry '%s' has empty Category field", toolID)
		}

		if entry.Description == "" {
			t.Errorf("catalog entry '%s' has empty Description field", toolID)
		}
	}
}
