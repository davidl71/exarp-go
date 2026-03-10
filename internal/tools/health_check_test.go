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
	readmePath := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Test Project\n"), 0644); err != nil {
		t.Fatalf("failed to create README: %v", err)
	}
	// Create a test project structure
	docsDir := filepath.Join(tmpDir, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatalf("failed to create docs directory: %v", err)
	}
	// Create a test markdown file
	testDoc := filepath.Join(docsDir, "test.md")
	if err := os.WriteFile(testDoc, []byte("# Test Document\n\nSee [Missing](missing.md).\nOld path: /Users/davidl/Projects/exarp-go"), 0644); err != nil {
		t.Fatalf("failed to create test doc: %v", err)
	}
	nestedDir := filepath.Join(docsDir, "guides")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("failed to create nested docs directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nestedDir, "nested.md"), []byte("# Nested\n"), 0644); err != nil {
		t.Fatalf("failed to create nested doc: %v", err)
	}
	archiveDir := filepath.Join(docsDir, "archive")
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		t.Fatalf("failed to create archive docs directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(archiveDir, "archived.md"), []byte("# Archived\n"), 0644); err != nil {
		t.Fatalf("failed to create archive doc: %v", err)
	}
	if err := os.WriteFile(filepath.Join(docsDir, "DOCUMENTATION_HEALTH_REPORT.md"), []byte("# Generated\n"), 0644); err != nil {
		t.Fatalf("failed to create generated report doc: %v", err)
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

				if status, ok := data["status"].(string); !ok || status != "completed" {
					t.Fatalf("expected status=completed, got %v", data["status"])
				}
				if scanMode, ok := data["scan_mode"].(string); !ok || scanMode != "recursive" {
					t.Fatalf("expected scan_mode=recursive, got %v", data["scan_mode"])
				}
				checks, ok := data["checks"].(map[string]interface{})
				if !ok {
					t.Fatalf("expected checks map, got %T", data["checks"])
				}
				if got := int(checks["docs_file_count"].(float64)); got != 4 {
					t.Fatalf("docs_file_count = %d, want 4", got)
				}
				if got := int(checks["live_docs_count"].(float64)); got != 2 {
					t.Fatalf("live_docs_count = %d, want 2", got)
				}
				if got := int(checks["archive_docs_count"].(float64)); got != 1 {
					t.Fatalf("archive_docs_count = %d, want 1", got)
				}
				if got := int(checks["generated_docs_count"].(float64)); got != 1 {
					t.Fatalf("generated_docs_count = %d, want 1", got)
				}
				if got := int(checks["stale_path_matches"].(float64)); got != 1 {
					t.Fatalf("stale_path_matches = %d, want 1", got)
				}
				if got := int(checks["missing_reference_count"].(float64)); got != 1 {
					t.Fatalf("missing_reference_count = %d, want 1", got)
				}
				if score := data["health_score"].(float64); score >= 100 {
					t.Fatalf("health_score = %v, want degraded score", score)
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
				if got, _ := data["changed_files"].(string); got != "docs/test.md" {
					t.Fatalf("changed_files = %q, want docs/test.md", got)
				}
			},
		},
		{
			name: "writes report file",
			params: map[string]interface{}{
				"action":      "docs",
				"output_path": "out/docs-health.md",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				reportPath, _ := data["report_path"].(string)
				if reportPath == "" {
					t.Fatal("expected report_path")
				}
				raw, err := os.ReadFile(reportPath)
				if err != nil {
					t.Fatalf("expected report file: %v", err)
				}
				if !filepath.IsAbs(reportPath) {
					t.Fatalf("report_path should be absolute, got %q", reportPath)
				}
				if len(raw) == 0 {
					t.Fatal("expected non-empty report file")
				}
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

				if success, ok := data["success"].(bool); ok && success {
					// ok
				} else if status, ok := data["status"].(string); ok && status == "completed" {
					// ok
				} else {
					t.Error("expected success=true or status=completed")
				}
				if categories, ok := data["categories"].(map[string]interface{}); ok && len(categories) > 0 {
					return
				}
				if checks, ok := data["checks"].(map[string]interface{}); ok && len(checks) > 0 {
					return
				}
				t.Error("expected categories or checks in result")
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

				if success, ok := data["success"].(bool); ok && success {
					return
				}
				if status, ok := data["status"].(string); ok && status == "completed" {
					return
				}
				t.Error("expected success=true or status=completed")
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
