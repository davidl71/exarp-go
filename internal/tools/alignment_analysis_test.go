package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
)

func TestHandleAlignmentTodo2(t *testing.T) {
	tmpDir := t.TempDir()
	// Optional PROJECT_GOALS.md for goal-based alignment
	goalsPath := filepath.Join(tmpDir, "PROJECT_GOALS.md")
	_ = os.WriteFile(goalsPath, []byte("# Goals\nMigrate to Go.\n"), 0644)

	oldRoot := os.Getenv("PROJECT_ROOT")
	os.Setenv("PROJECT_ROOT", tmpDir)
	defer os.Setenv("PROJECT_ROOT", oldRoot)

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		validate  func(*testing.T, []framework.TextContent)
	}{
		{
			name: "todo2 default action",
			params: map[string]interface{}{
				"action": "todo2",
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
				if status, _ := data["status"].(string); status != "success" {
					t.Errorf("expected status=success, got %q", status)
				}
				if _, ok := data["total_tasks_analyzed"]; !ok {
					t.Error("expected total_tasks_analyzed")
				}
			},
		},
		{
			name: "todo2 with create_followup_tasks false",
			params: map[string]interface{}{
				"action":                "todo2",
				"create_followup_tasks": false,
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				if _, ok := data["report_path"]; !ok {
					t.Error("expected report_path in response")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			result, err := handleAlignmentTodo2(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleAlignmentTodo2() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestHandleAlignmentPRD(t *testing.T) {
	tmpDir := t.TempDir()

	prdDir := filepath.Join(tmpDir, "docs")
	if err := os.MkdirAll(prdDir, 0755); err != nil {
		t.Fatalf("failed to create docs directory: %v", err)
	}

	prdFile := filepath.Join(prdDir, "PRD.md")
	prdContent := `# Product Requirements Document

## 1. User Stories

### US-1: User Authentication
As a user, I want to authenticate so that I can access my account.

### US-2: Data Export
As a user, I want to export my data so that I can backup my information.
`

	if err := os.WriteFile(prdFile, []byte(prdContent), 0644); err != nil {
		t.Fatalf("failed to create PRD file: %v", err)
	}

	oldRoot := os.Getenv("PROJECT_ROOT")
	os.Setenv("PROJECT_ROOT", tmpDir)
	defer os.Setenv("PROJECT_ROOT", oldRoot)

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
				var envelope map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &envelope); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				if success, ok := envelope["success"].(bool); !ok || !success {
					t.Error("expected success=true in envelope")
				}
				data, _ := envelope["data"].(map[string]interface{})
				if data == nil {
					t.Fatal("expected data object")
				}
				if _, ok := data["tasks_analyzed"]; !ok {
					t.Error("expected tasks_analyzed in data")
				}
				if _, ok := data["overall_alignment_score"]; !ok {
					t.Error("expected overall_alignment_score in data")
				}
				if _, ok := data["recommendations"]; !ok {
					t.Error("expected recommendations in data")
				}
			},
		},
		{
			name: "PRD with output_path",
			params: map[string]interface{}{
				"action":      "prd",
				"output_path": "out/prd_alignment.json",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var envelope map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &envelope); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				data, _ := envelope["data"].(map[string]interface{})
				if data != nil {
					if reportPath, _ := data["report_path"].(string); reportPath != "" {
						if _, err := os.Stat(reportPath); err != nil {
							t.Errorf("report_path should exist: %v", err)
						}
					}
				}
			},
		},
	}

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
		validate  func(*testing.T, []framework.TextContent)
	}{
		{
			name: "todo2 action",
			params: map[string]interface{}{
				"action": "todo2",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
				var data map[string]interface{}
				if json.Unmarshal([]byte(result[0].Text), &data) != nil {
					return
				}
				if data["status"] != "success" {
					t.Error("expected status=success for todo2")
				}
			},
		},
		{
			name: "prd action",
			params: map[string]interface{}{
				"action": "prd",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
				var envelope map[string]interface{}
				if json.Unmarshal([]byte(result[0].Text), &envelope) != nil {
					return
				}
				if success, _ := envelope["success"].(bool); !success {
					t.Error("expected success=true for prd")
				}
			},
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

			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}
