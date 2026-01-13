package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
)

func TestHandleMemoryNative(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PROJECT_ROOT", tmpDir)
	defer os.Unsetenv("PROJECT_ROOT")

	// Create .exarp directory
	exarpDir := filepath.Join(tmpDir, ".exarp")
	if err := os.MkdirAll(exarpDir, 0755); err != nil {
		t.Fatalf("failed to create .exarp directory: %v", err)
	}

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		validate  func(*testing.T, []framework.TextContent)
	}{
		{
			name: "save action",
			params: map[string]interface{}{
				"action":  "save",
				"title":  "Test Memory",
				"content": "This is a test memory",
				"category": "insight",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				if success, ok := data["success"].(bool); !ok || !success {
					t.Error("expected success=true")
				}
			},
		},
		{
			name: "recall action",
			params: map[string]interface{}{
				"action": "recall",
				"title":  "Test Memory",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				// Should return success even if memory not found
			},
		},
		{
			name: "list action",
			params: map[string]interface{}{
				"action": "list",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				if success, ok := data["success"].(bool); !ok || !success {
					t.Error("expected success=true")
				}
			},
		},
		{
			name: "search action",
			params: map[string]interface{}{
				"action": "search",
				"query":  "test",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				// Search may return empty results, but should succeed
			},
		},
		{
			name: "unknown action",
			params: map[string]interface{}{
				"action": "unknown",
			},
			wantError: false, // Falls back to Python bridge
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			argsJSON, _ := json.Marshal(tt.params)
			result, err := handleMemoryNative(ctx, argsJSON)
			if (err != nil) != tt.wantError {
				t.Errorf("handleMemoryNative() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestHandleMemorySave(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PROJECT_ROOT", tmpDir)
	defer os.Unsetenv("PROJECT_ROOT")

	// Create .exarp directory
	exarpDir := filepath.Join(tmpDir, ".exarp")
	if err := os.MkdirAll(exarpDir, 0755); err != nil {
		t.Fatalf("failed to create .exarp directory: %v", err)
	}

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name: "valid memory save",
			params: map[string]interface{}{
				"title":    "Test Memory",
				"content":  "This is test content",
				"category": "insight",
			},
			wantError: false,
		},
		{
			name: "missing title",
			params: map[string]interface{}{
				"content":  "Test content",
				"category": "insight",
			},
			wantError: true,
		},
		{
			name: "missing content",
			params: map[string]interface{}{
				"title":    "Test Memory",
				"category": "insight",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := handleMemorySave(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleMemorySave() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && len(result) == 0 {
				t.Error("expected non-empty result")
			}
		})
	}
}

func TestHandleMemoryRecall(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PROJECT_ROOT", tmpDir)
	defer os.Unsetenv("PROJECT_ROOT")

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name: "recall by title",
			params: map[string]interface{}{
				"title": "Test Memory",
			},
			wantError: false,
		},
		{
			name: "recall by id",
			params: map[string]interface{}{
				"id": "test-id",
			},
			wantError: false,
		},
		{
			name: "missing title and id",
			params: map[string]interface{}{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := handleMemoryRecall(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleMemoryRecall() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && len(result) == 0 {
				t.Error("expected non-empty result")
			}
		})
	}
}
