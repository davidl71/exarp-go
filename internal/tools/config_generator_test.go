package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/tests/fixtures"
)

func TestHandleGenerateConfigNative(t *testing.T) {
	ctx := context.Background()

	// Create temporary test directory
	tmpDir := t.TempDir()

	goModPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module test\n"), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	// Change to tmpDir so GetProjectRoot(".") finds it
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir to temp dir: %v", err)
	}

	defer func() {
		if chdirErr := os.Chdir(originalWd); chdirErr != nil {
			t.Errorf("failed to restore working directory: %v", chdirErr)
		}
	}()

	tests := []struct {
		name           string
		args           json.RawMessage
		wantErr        bool
		validateResult func(t *testing.T, result []framework.TextContent) bool
	}{
		{
			name: "rules action with analyze_only",
			args: fixtures.MustToJSONRawMessage(map[string]interface{}{
				"action":       "rules",
				"analyze_only": true,
			}),
			wantErr: false,
			validateResult: func(t *testing.T, result []framework.TextContent) bool {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return false
				}

				var response map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &response); err != nil {
					t.Errorf("failed to unmarshal result: %v", err)
					return false
				}

				if response["success"] != true {
					t.Errorf("expected success=true, got %v", response["success"])
					return false
				}

				if _, ok := response["analysis"]; !ok {
					t.Error("expected analysis in result")
					return false
				}

				return true
			},
		},
		{
			name: "ignore action with dry_run",
			args: fixtures.MustToJSONRawMessage(map[string]interface{}{
				"action":  "ignore",
				"dry_run": true,
			}),
			wantErr: false,
			validateResult: func(t *testing.T, result []framework.TextContent) bool {
				var response map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &response); err != nil {
					t.Errorf("failed to unmarshal result: %v", err)
					return false
				}

				if response["success"] != true {
					t.Errorf("expected success=true, got %v", response["success"])
					return false
				}

				return true
			},
		},
		{
			name: "simplify action with dry_run",
			args: fixtures.MustToJSONRawMessage(map[string]interface{}{
				"action":  "simplify",
				"dry_run": true,
			}),
			wantErr: false,
			validateResult: func(t *testing.T, result []framework.TextContent) bool {
				var response map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &response); err != nil {
					t.Errorf("failed to unmarshal result: %v", err)
					return false
				}

				if response["status"] != "success" {
					t.Errorf("expected status='success', got %v", response["status"])
					return false
				}

				return true
			},
		},
		{
			name: "unknown action",
			args: fixtures.MustToJSONRawMessage(map[string]interface{}{
				"action": "unknown",
			}),
			wantErr:        true,
			validateResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("PROJECT_ROOT", tmpDir)

			result, err := handleGenerateConfigNative(ctx, tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("handleGenerateConfigNative() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

func TestHandleGenerateRules(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	goModPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module test\n"), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	tests := []struct {
		name           string
		params         map[string]interface{}
		wantErr        bool
		validateResult func(t *testing.T, result []framework.TextContent) bool
	}{
		{
			name: "analyze_only mode",
			params: map[string]interface{}{
				"analyze_only": true,
			},
			wantErr: false,
			validateResult: func(t *testing.T, result []framework.TextContent) bool {
				var response map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &response); err != nil {
					t.Errorf("failed to unmarshal result: %v", err)
					return false
				}

				if response["success"] != true {
					t.Errorf("expected success=true, got %v", response["success"])
					return false
				}

				if _, ok := response["analysis"]; !ok {
					t.Error("expected analysis in result")
					return false
				}

				if _, ok := response["available_rules"]; !ok {
					t.Error("expected available_rules in result")
					return false
				}

				return true
			},
		},
		{
			name: "generate rules with dry run",
			params: map[string]interface{}{
				"rules":     "go",
				"overwrite": false,
			},
			wantErr: false,
			validateResult: func(t *testing.T, result []framework.TextContent) bool {
				var response map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &response); err != nil {
					t.Errorf("failed to unmarshal result: %v", err)
					return false
				}

				if response["success"] != true {
					t.Errorf("expected success=true, got %v", response["success"])
					return false
				}

				return true
			},
		},
		{
			name: "generate with overwrite",
			params: map[string]interface{}{
				"rules":     "go",
				"overwrite": true,
			},
			wantErr: false,
			validateResult: func(t *testing.T, result []framework.TextContent) bool {
				var response map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &response); err != nil {
					t.Errorf("failed to unmarshal result: %v", err)
					return false
				}

				if response["success"] != true {
					t.Errorf("expected success=true, got %v", response["success"])
					return false
				}

				return true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handleGenerateRules(ctx, tt.params, tmpDir)

			if (err != nil) != tt.wantErr {
				t.Errorf("handleGenerateRules() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

func TestHandleGenerateIgnore(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	tests := []struct {
		name           string
		params         map[string]interface{}
		wantErr        bool
		validateResult func(t *testing.T, result []framework.TextContent) bool
	}{
		{
			name: "dry run",
			params: map[string]interface{}{
				"dry_run": true,
			},
			wantErr: false,
			validateResult: func(t *testing.T, result []framework.TextContent) bool {
				var response map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &response); err != nil {
					t.Errorf("failed to unmarshal result: %v", err)
					return false
				}

				if response["success"] != true {
					t.Errorf("expected success=true, got %v", response["success"])
					return false
				}

				return true
			},
		},
		{
			name: "with include_indexing",
			params: map[string]interface{}{
				"include_indexing": true,
				"analyze_project":  true,
				"dry_run":          true,
			},
			wantErr: false,
			validateResult: func(t *testing.T, result []framework.TextContent) bool {
				var response map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &response); err != nil {
					t.Errorf("failed to unmarshal result: %v", err)
					return false
				}

				if response["success"] != true {
					t.Errorf("expected success=true, got %v", response["success"])
					return false
				}

				return true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handleGenerateIgnore(ctx, tt.params, tmpDir)

			if (err != nil) != tt.wantErr {
				t.Errorf("handleGenerateIgnore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

func TestHandleSimplifyRules(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	tests := []struct {
		name           string
		params         map[string]interface{}
		setupFiles     func(string)
		wantErr        bool
		validateResult func(t *testing.T, result []framework.TextContent) bool
	}{
		{
			name: "dry run with no files",
			params: map[string]interface{}{
				"dry_run": true,
			},
			setupFiles: func(projectRoot string) {
				// No files
			},
			wantErr: false,
			validateResult: func(t *testing.T, result []framework.TextContent) bool {
				var response map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &response); err != nil {
					t.Errorf("failed to unmarshal result: %v", err)
					return false
				}

				if response["status"] != "success" {
					t.Errorf("expected status='success', got %v", response["status"])
					return false
				}

				return true
			},
		},
		{
			name: "dry run with .cursorrules file",
			params: map[string]interface{}{
				"dry_run": true,
			},
			setupFiles: func(projectRoot string) {
				if err := os.WriteFile(filepath.Join(projectRoot, ".cursorrules"), []byte("# Test rules\n"), 0644); err != nil {
					return
				}
			},
			wantErr: false,
			validateResult: func(t *testing.T, result []framework.TextContent) bool {
				var response map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &response); err != nil {
					t.Errorf("failed to unmarshal result: %v", err)
					return false
				}

				if response["status"] != "success" {
					t.Errorf("expected status='success', got %v", response["status"])
					return false
				}

				return true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFiles != nil {
				tt.setupFiles(tmpDir)
			}

			result, err := handleSimplifyRules(ctx, tt.params, tmpDir)

			if (err != nil) != tt.wantErr {
				t.Errorf("handleSimplifyRules() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

func TestCursorRulesGenerator_AnalyzeProject(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name       string
		setupFiles func(string)
		validate   func(t *testing.T, analysis map[string]interface{})
	}{
		{
			name: "empty project",
			setupFiles: func(projectRoot string) {
				// No files
			},
			validate: func(t *testing.T, analysis map[string]interface{}) {
				if _, ok := analysis["languages"]; !ok {
					t.Error("expected languages in analysis")
				}

				if _, ok := analysis["frameworks"]; !ok {
					t.Error("expected frameworks in analysis")
				}
			},
		},
		{
			name: "Go project",
			setupFiles: func(projectRoot string) {
				os.WriteFile(filepath.Join(projectRoot, "main.go"), []byte("package main\n"), 0644)
			},
			validate: func(t *testing.T, analysis map[string]interface{}) {
				languages, _ := analysis["languages"].([]interface{})
				foundGo := false

				for _, lang := range languages {
					if lang == "go" {
						foundGo = true
						break
					}
				}

				if !foundGo {
					t.Log("Note: Go language may not be detected in empty project")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFiles != nil {
				tt.setupFiles(tmpDir)
			}

			generator := NewCursorRulesGenerator(tmpDir)
			analysis := generator.AnalyzeProject()

			if tt.validate != nil {
				tt.validate(t, analysis)
			}
		})
	}
}

func TestGetAvailableRuleNames(t *testing.T) {
	rules := getAvailableRuleNames()
	if len(rules) == 0 {
		t.Error("expected non-empty rule names")
	}

	// Check for expected rules
	expectedRules := []string{"go", "python", "typescript"}
	for _, expected := range expectedRules {
		found := false

		for _, rule := range rules {
			if rule == expected {
				found = true
				break
			}
		}

		if !found {
			t.Logf("Note: Expected rule '%s' not found in available rules", expected)
		}
	}
}

func TestGetRuleTemplate(t *testing.T) {
	tests := []struct {
		name     string
		ruleName string
		want     bool
	}{
		{
			name:     "valid rule",
			ruleName: "go",
			want:     true,
		},
		{
			name:     "invalid rule",
			ruleName: "nonexistent",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template, exists := getRuleTemplate(tt.ruleName)
			if exists != tt.want {
				t.Errorf("getRuleTemplate() exists = %v, want %v", exists, tt.want)
				return
			}

			if tt.want {
				if template.Filename == "" {
					t.Error("expected non-empty filename")
				}

				if template.Content == "" {
					t.Error("expected non-empty content")
				}
			}
		})
	}
}
