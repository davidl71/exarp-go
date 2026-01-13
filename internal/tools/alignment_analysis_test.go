package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
)

func TestHandleAlignmentPRD(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a test PRD file
	prdDir := filepath.Join(tmpDir, "docs")
	if err := os.MkdirAll(prdDir, 0755); err != nil {
		t.Fatalf("failed to create docs directory: %v", err)
	}
	prdFile := filepath.Join(prdDir, "PRD.md")
	prdContent := `# Product Requirements Document

## User Stories

### US-1: User Authentication
As a user, I want to authenticate so that I can access my account.

### US-2: Data Export
As a user, I want to export my data so that I can backup my information.
`
	if err := os.WriteFile(prdFile, []byte(prdContent), 0644); err != nil {
		t.Fatalf("failed to create PRD file: %v", err)
	}

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		validate  func(*testing.T, []framework.TextContent)
	}{
		{
			name: "basic PRD alignment",
			params: map[string]interface{}{
				"action": "prd",
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
			name: "with create_followup_tasks",
			params: map[string]interface{}{
				"action":                "prd",
				"create_followup_tasks": true,
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				// Should have analyzed PRD alignment
			},
		},
		{
			name: "with output_path",
			params: map[string]interface{}{
				"action":      "prd",
				"output_path": "prd_alignment.json",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				// Should have saved report
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
			result, err := handleAlignmentPRD(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleAlignmentPRD() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestHandleAnalyzeAlignmentNative(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name: "todo2 action",
			params: map[string]interface{}{
				"action": "todo2",
			},
			wantError: false,
		},
		{
			name: "prd action",
			params: map[string]interface{}{
				"action": "prd",
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
			result, err := handleAnalyzeAlignmentNative(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleAnalyzeAlignmentNative() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && (result == nil || len(result) == 0) {
				t.Error("expected non-empty result")
			}
		})
	}
}
