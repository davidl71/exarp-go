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
				"action":   "save",
				"title":    "Test Memory",
				"content":  "This is a test memory",
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
				"action":  "recall",
				"task_id": "test-task",
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
			wantError: true, // Native returns error for unknown action
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
			name: "recall by task_id",
			params: map[string]interface{}{
				"task_id": "test-task",
			},
			wantError: false,
		},
		{
			name: "recall by task_id alternate",
			params: map[string]interface{}{
				"task_id": "T-123",
			},
			wantError: false,
		},
		{
			name:      "missing task_id",
			params:    map[string]interface{}{},
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

// memoryResponseAllowedKeys is the set of top-level keys allowed in memory tool
// success responses (from proto.MemoryResponse / MemoryResponseToMap). Used to
// enforce that all memory tool responses stay aligned to MemoryResponse.
var memoryResponseAllowedKeys = map[string]bool{
	"success": true, "method": true, "count": true, "memory_id": true,
	"message": true, "memories": true, "categories": true, "available_categories": true,
	"total_found": true, "task_id": true, "include_related": true, "query": true,
	"total": true, "returned": true,
}

// TestMemoryToolResponsesUseMemoryResponseShape ensures every success response
// from the memory tool has only MemoryResponse proto fields (no ad-hoc keys).
func TestMemoryToolResponsesUseMemoryResponseShape(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PROJECT_ROOT", tmpDir)

	defer os.Unsetenv("PROJECT_ROOT")

	exarpDir := filepath.Join(tmpDir, ".exarp")
	if err := os.MkdirAll(exarpDir, 0755); err != nil {
		t.Fatalf("failed to create .exarp directory: %v", err)
	}

	ctx := context.Background()
	actions := []struct {
		name   string
		params map[string]interface{}
	}{
		{"save", map[string]interface{}{"action": "save", "title": "T", "content": "C", "category": "insight"}},
		{"recall", map[string]interface{}{"action": "recall", "task_id": "x"}},
		{"search", map[string]interface{}{"action": "search", "query": "q"}},
		{"list", map[string]interface{}{"action": "list"}},
	}

	for _, a := range actions {
		t.Run(a.name, func(t *testing.T) {
			argsJSON, _ := json.Marshal(a.params)

			result, err := handleMemoryNative(ctx, argsJSON)
			if err != nil {
				t.Fatalf("handleMemoryNative() error = %v", err)
			}

			if len(result) == 0 {
				t.Fatal("expected non-empty result")
			}

			var data map[string]interface{}
			if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
				t.Fatalf("invalid JSON: %v", err)
			}

			if success, _ := data["success"].(bool); !success {
				t.Error("expected success=true")
			}

			if _, ok := data["method"]; !ok {
				t.Error("expected method key (MemoryResponse shape)")
			}

			for k := range data {
				if !memoryResponseAllowedKeys[k] {
					t.Errorf("response key %q is not a MemoryResponse field; keep memory tool aligned to proto", k)
				}
			}
		})
	}
}
