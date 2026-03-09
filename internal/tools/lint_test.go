package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
)

func TestHandleLint(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROJECT_ROOT", tmpDir)

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		validate  func(*testing.T, []framework.TextContent)
	}{
		{
			name: "lint with golangci-lint",
			params: map[string]interface{}{
				"action": "run",
				"linter": "golangci-lint",
				"path":   ".",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
				// Result may be lint output or error message (if linter not available)
			},
		},
		{
			name: "lint with auto detection",
			params: map[string]interface{}{
				"action": "run",
				"linter": "auto",
				"path":   ".",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				// Result may be lint output or error message
			},
		},
		{
			name: "lint with fix",
			params: map[string]interface{}{
				"action": "run",
				"linter": "golangci-lint",
				"path":   ".",
				"fix":    true,
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				// Result may be lint output or error message
			},
		},
		{
			name: "analyze action",
			params: map[string]interface{}{
				"action": "analyze",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				// Result may be analysis output
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			argsJSON, _ := json.Marshal(tt.params)

			result, err := handleLint(ctx, argsJSON)
			if (err != nil) != tt.wantError {
				t.Errorf("handleLint() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestHandleLint_DefaultsToAutoDetection(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROJECT_ROOT", tmpDir)

	pyFile := filepath.Join(tmpDir, "test.py")
	if err := os.WriteFile(pyFile, []byte("print('hello')\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	argsJSON, _ := json.Marshal(map[string]interface{}{
		"action": "run",
		"path":   pyFile,
	})

	result, err := handleLint(context.Background(), argsJSON)
	if err != nil {
		t.Fatalf("handleLint() error = %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(result[0].Text), &payload); err != nil {
		t.Fatalf("invalid JSON result: %v", err)
	}

	got, _ := payload["linter"].(string)
	switch got {
	case "ruff", "flake8", "pylint":
		// Expected: a Python linter selected from auto-detection.
	default:
		t.Fatalf("default linter = %v, want Python linter from auto-detection", got)
	}
}

func TestHandleLint_AutoDirectoryAggregatesLinters(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROJECT_ROOT", tmpDir)

	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644); err != nil {
		t.Fatalf("WriteFile go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# test\n"), 0644); err != nil {
		t.Fatalf("WriteFile md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "script.sh"), []byte("#!/bin/sh\necho ok\n"), 0755); err != nil {
		t.Fatalf("WriteFile sh: %v", err)
	}

	argsJSON, _ := json.Marshal(map[string]interface{}{
		"action": "run",
		"linter": "auto",
		"path":   tmpDir,
	})

	result, err := handleLint(context.Background(), argsJSON)
	if err != nil {
		t.Fatalf("handleLint() error = %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(result[0].Text), &payload); err != nil {
		t.Fatalf("invalid JSON result: %v", err)
	}

	if got := payload["linter"]; got != "auto" {
		t.Fatalf("linter = %v, want auto", got)
	}

	raw, ok := payload["raw"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected raw object, got %#v", payload["raw"])
	}

	selected, ok := raw["selected_linters"].([]interface{})
	if !ok || len(selected) == 0 {
		t.Fatalf("expected selected_linters, got %#v", raw["selected_linters"])
	}

	selectedText := make([]string, 0, len(selected))
	for _, item := range selected {
		if s, ok := item.(string); ok {
			selectedText = append(selectedText, s)
		}
	}

	joined := strings.Join(selectedText, ",")
	if !strings.Contains(joined, "markdownlint") {
		t.Fatalf("selected_linters missing markdownlint: %v", selectedText)
	}
	if !strings.Contains(joined, "shellcheck") {
		t.Fatalf("selected_linters missing shellcheck: %v", selectedText)
	}
	if !strings.Contains(joined, "go-vet") && !strings.Contains(joined, "golangci-lint") {
		t.Fatalf("selected_linters missing Go linter: %v", selectedText)
	}
}

func TestRunLinter(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROJECT_ROOT", tmpDir)

	tests := []struct {
		name      string
		linter    string
		path      string
		fix       bool
		wantError bool
	}{
		{
			name:      "golangci-lint",
			linter:    "golangci-lint",
			path:      ".",
			fix:       false,
			wantError: false, // May fail if linter not available, but function should handle it
		},
		{
			name:      "unsupported linter",
			linter:    "unsupported-linter",
			path:      ".",
			fix:       false,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			result, err := runLinter(ctx, tt.linter, tt.path, tt.fix)
			if (err != nil) != tt.wantError {
				t.Errorf("runLinter() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && result == nil {
				t.Error("expected non-nil result")
			}
		})
	}
}
