package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
)

func TestHandleHealthDocs(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a test project structure
	docsDir := filepath.Join(tmpDir, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatalf("failed to create docs directory: %v", err)
	}
	// Create a test markdown file
	testDoc := filepath.Join(docsDir, "test.md")
	if err := os.WriteFile(testDoc, []byte("# Test Document\n\nContent here."), 0644); err != nil {
		t.Fatalf("failed to create test doc: %v", err)
	}

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		validate  func(*testing.T, []framework.TextContent)
	}{
		{
			name: "basic docs check",
			params: map[string]interface{}{
				"action": "docs",
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
			name: "with changed files",
			params: map[string]interface{}{
				"action":        "docs",
				"changed_files": "docs/test.md",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				// Should have checked the changed file
			},
		},
	}

	// Set PROJECT_ROOT for tests
	oldRoot := os.Getenv("PROJECT_ROOT")
	os.Setenv("PROJECT_ROOT", tmpDir)
	defer os.Setenv("PROJECT_ROOT", oldRoot)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := handleHealthDocs(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleHealthDocs() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestHandleHealthDOD(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		validate  func(*testing.T, []framework.TextContent)
	}{
		{
			name: "basic DoD check",
			params: map[string]interface{}{
				"action": "dod",
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
				// Should have categories
				if categories, ok := data["categories"].(map[string]interface{}); !ok || len(categories) == 0 {
					t.Error("expected categories in result")
				}
			},
		},
		{
			name: "with task_id",
			params: map[string]interface{}{
				"action":  "dod",
				"task_id": "T-123",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				if taskID, ok := data["task_id"].(string); !ok || taskID != "T-123" {
					t.Errorf("expected task_id=T-123, got %v", data["task_id"])
				}
			},
		},
		{
			name: "with auto_check",
			params: map[string]interface{}{
				"action":     "dod",
				"auto_check": true,
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				// Should have run checks
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := handleHealthDOD(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleHealthDOD() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestHandleHealthCICD(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		validate  func(*testing.T, []framework.TextContent)
	}{
		{
			name: "basic CI/CD check",
			params: map[string]interface{}{
				"action": "cicd",
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
			name: "with workflow_path",
			params: map[string]interface{}{
				"action":        "cicd",
				"workflow_path": ".github/workflows/test.yml",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				// Should have validated workflow
			},
		},
		{
			name: "with check_runners",
			params: map[string]interface{}{
				"action":        "cicd",
				"check_runners": true,
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				// Should have checked runners
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := handleHealthCICD(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleHealthCICD() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestHandleHealthNative(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name: "server action",
			params: map[string]interface{}{
				"action": "server",
			},
			wantError: false,
		},
		{
			name: "git action",
			params: map[string]interface{}{
				"action": "git",
			},
			wantError: false,
		},
		{
			name: "docs action",
			params: map[string]interface{}{
				"action": "docs",
			},
			wantError: false,
		},
		{
			name: "dod action",
			params: map[string]interface{}{
				"action": "dod",
			},
			wantError: false,
		},
		{
			name: "cicd action",
			params: map[string]interface{}{
				"action": "cicd",
			},
			wantError: false,
		},
		{
			name: "unknown action",
			params: map[string]interface{}{
				"action": "unknown",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := handleHealthNative(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleHealthNative() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && (result == nil || len(result) == 0) {
				t.Error("expected non-empty result")
			}
		})
	}
}
