package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	exarpconfig "github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/framework"
)

func TestHandleTaskDiscoveryNative(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROJECT_ROOT", tmpDir)

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		validate  func(*testing.T, []framework.TextContent)
	}{
		{
			name: "comments action",
			params: map[string]interface{}{
				"action": "comments",
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
			},
		},
		{
			name: "markdown action",
			params: map[string]interface{}{
				"action": "markdown",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
			},
		},
		{
			name: "orphans action",
			params: map[string]interface{}{
				"action": "orphans",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
			},
		},
		{
			name: "all action",
			params: map[string]interface{}{
				"action": "all",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			result, err := handleTaskDiscoveryNative(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleTaskDiscoveryNative() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestHandleTaskDiscovery(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROJECT_ROOT", tmpDir)

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name: "comments action",
			params: map[string]interface{}{
				"action": "comments",
			},
			wantError: false,
		},
		{
			name: "all action",
			params: map[string]interface{}{
				"action": "all",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			argsJSON, _ := json.Marshal(tt.params)

			result, err := handleTaskDiscovery(ctx, argsJSON)
			if (err != nil) != tt.wantError {
				t.Errorf("handleTaskDiscovery() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && (result == nil || len(result) == 0) {
				t.Error("expected non-empty result")
			}
		})
	}
}

func TestIsDeprecatedDiscoveryText(t *testing.T) {
	tests := []struct {
		text string
		want bool
	}{
		{"", true},
		{"   ", true},
		{"Add middleware", false},
		{"~~Add middleware (T-274 removed)~~", true},
		{"Add middleware (removed)", true},
		{"*(T-274 removed)*", true},
		{"Future improvement (T-123 removed)", true},
		{"Future improvement only", false},
		{"Normal task text", false},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			if got := IsDeprecatedDiscoveryText(tt.text); got != tt.want {
				t.Errorf("IsDeprecatedDiscoveryText(%q) = %v, want %v", tt.text, got, tt.want)
			}
		})
	}
}

func TestHandleTaskDiscoveryNativeHonorsIgnorePaths(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROJECT_ROOT", tmpDir)

	mustWriteFile(t, filepath.Join(tmpDir, "ib-gateway", "root", "webapps", "demo", "gateway.demo.js"), "// TODO: ignore me\n")
	mustWriteFile(t, filepath.Join(tmpDir, "app", "main.go"), "// TODO: keep me\n")

	result, err := handleTaskDiscoveryNative(context.Background(), map[string]interface{}{
		"action":       "comments",
		"ignore_paths": "ib-gateway",
		"use_llm":      false,
	})
	if err != nil {
		t.Fatalf("handleTaskDiscoveryNative() error = %v", err)
	}

	discoveries := decodeDiscoveries(t, result)
	if len(discoveries) != 1 {
		t.Fatalf("expected 1 discovery after ignore_paths filter, got %d: %#v", len(discoveries), discoveries)
	}

	if got := discoveries[0]["file"]; got != "app/main.go" {
		t.Fatalf("expected remaining discovery from app/main.go, got %v", got)
	}
}

func TestHandleTaskDiscoveryNativeScansNativeDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROJECT_ROOT", tmpDir)

	mustWriteFile(t, filepath.Join(tmpDir, "native", "src", "todo.go"), "// TODO: native path should be scanned\n")

	result, err := handleTaskDiscoveryNative(context.Background(), map[string]interface{}{
		"action":  "comments",
		"use_llm": false,
	})
	if err != nil {
		t.Fatalf("handleTaskDiscoveryNative() error = %v", err)
	}

	discoveries := decodeDiscoveries(t, result)
	if len(discoveries) != 1 {
		t.Fatalf("expected 1 discovery from native path, got %d: %#v", len(discoveries), discoveries)
	}

	if got := discoveries[0]["file"]; got != "native/src/todo.go" {
		t.Fatalf("expected discovery from native/src/todo.go, got %v", got)
	}
}

func TestHandleTaskDiscoveryNativeScansTaggedTodoSyntaxAcrossRepoLanguages(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROJECT_ROOT", tmpDir)

	mustWriteFile(t, filepath.Join(tmpDir, "native", "src", "margin_calculator.cpp"), strings.Join([]string{
		"// TODO: Source IV from live market data snapshot instead of 0.20 default.",
		"// TODO(exarp): Split by domain before financial math sprint",
	}, "\n")+"\n")
	mustWriteFile(t, filepath.Join(tmpDir, "agents", "backend", "services", "tui_service", "src", "ui.rs"), "// TODO(sparklines): add a Trend column using ratatui sparkline widget.\n")
	mustWriteFile(t, filepath.Join(tmpDir, "agents", "backend", "services", "tui_service", "Cargo.toml"), "# TODO(ratatui-textarea): add ratatui-textarea for interactive filters\n")
	mustWriteFile(t, filepath.Join(tmpDir, "web", "src", "api", "snapshot.ts"), "// TODO(exarp): WebSocket delta compression\n")

	result, err := handleTaskDiscoveryNative(context.Background(), map[string]interface{}{
		"action":  "comments",
		"use_llm": false,
	})
	if err != nil {
		t.Fatalf("handleTaskDiscoveryNative() error = %v", err)
	}

	discoveries := decodeDiscoveries(t, result)
	if len(discoveries) != 5 {
		t.Fatalf("expected 5 discoveries, got %d: %#v", len(discoveries), discoveries)
	}

	gotFiles := map[string]bool{}
	gotTexts := map[string]bool{}
	for _, discovery := range discoveries {
		if file, ok := discovery["file"].(string); ok {
			gotFiles[file] = true
		}
		if text, ok := discovery["text"].(string); ok {
			gotTexts[text] = true
		}
	}

	expectedFiles := []string{
		"native/src/margin_calculator.cpp",
		"agents/backend/services/tui_service/src/ui.rs",
		"agents/backend/services/tui_service/Cargo.toml",
		"web/src/api/snapshot.ts",
	}
	for _, file := range expectedFiles {
		if !gotFiles[file] {
			t.Fatalf("expected discovery from %s, got files %#v", file, gotFiles)
		}
	}

	expectedTexts := []string{
		"Source IV from live market data snapshot instead of 0.20 default.",
		"Split by domain before financial math sprint",
		"add a Trend column using ratatui sparkline widget.",
		"add ratatui-textarea for interactive filters",
		"WebSocket delta compression",
	}
	for _, text := range expectedTexts {
		if !gotTexts[text] {
			t.Fatalf("expected discovery text %q, got texts %#v", text, gotTexts)
		}
	}
}

func TestHandleTaskDiscoveryNativeMarkdownHonorsIgnorePaths(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROJECT_ROOT", tmpDir)

	mustWriteFile(t, filepath.Join(tmpDir, "ib-gateway", "notes.md"), "- [ ] ignore markdown task\n")
	mustWriteFile(t, filepath.Join(tmpDir, "docs", "tasks.md"), "- [ ] keep markdown task\n")

	result, err := handleTaskDiscoveryNative(context.Background(), map[string]interface{}{
		"action":       "markdown",
		"ignore_paths": "ib-gateway",
	})
	if err != nil {
		t.Fatalf("handleTaskDiscoveryNative() error = %v", err)
	}

	discoveries := decodeDiscoveries(t, result)
	if len(discoveries) != 1 {
		t.Fatalf("expected 1 markdown discovery after ignore_paths filter, got %d: %#v", len(discoveries), discoveries)
	}

	if got := discoveries[0]["file"]; got != "docs/tasks.md" {
		t.Fatalf("expected remaining markdown discovery from docs/tasks.md, got %v", got)
	}
}

func TestHandleTaskDiscoveryNativePlanningHonorsIgnorePaths(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROJECT_ROOT", tmpDir)

	mustWriteFile(t, filepath.Join(tmpDir, "ib-gateway", "plan.md"), "Task ID: T-111\n")
	mustWriteFile(t, filepath.Join(tmpDir, "docs", "plan.md"), "Task ID: T-222\n")

	result, err := handleTaskDiscoveryNative(context.Background(), map[string]interface{}{
		"action":       "planning_links",
		"ignore_paths": "ib-gateway",
		"use_llm":      false,
	})
	if err != nil {
		t.Fatalf("handleTaskDiscoveryNative() error = %v", err)
	}

	discoveries := decodeDiscoveries(t, result)
	if len(discoveries) != 1 {
		t.Fatalf("expected 1 planning discovery after ignore_paths filter, got %d: %#v", len(discoveries), discoveries)
	}

	if got := discoveries[0]["file"]; got != "docs/plan.md" {
		t.Fatalf("expected remaining planning discovery from docs/plan.md, got %v", got)
	}
}

func TestHandleTaskDiscoveryNativeUsesProjectConfigIgnorePaths(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROJECT_ROOT", tmpDir)

	cfg := exarpconfig.GetDefaults()
	cfg.Project.TaskDiscoveryIgnorePaths = []string{"ib-gateway", "web/dev-dist"}
	if err := exarpconfig.WriteConfigToProtobufFile(tmpDir, cfg); err != nil {
		t.Fatalf("WriteConfigToProtobufFile() error = %v", err)
	}

	mustWriteFile(t, filepath.Join(tmpDir, "ib-gateway", "notes.md"), "- [ ] ignore markdown task\n")
	mustWriteFile(t, filepath.Join(tmpDir, "docs", "tasks.md"), "- [ ] keep markdown task\n")

	result, err := handleTaskDiscoveryNative(context.Background(), map[string]interface{}{
		"action": "markdown",
	})
	if err != nil {
		t.Fatalf("handleTaskDiscoveryNative() error = %v", err)
	}

	discoveries := decodeDiscoveries(t, result)
	if len(discoveries) != 1 {
		t.Fatalf("expected 1 markdown discovery from config ignore paths, got %d: %#v", len(discoveries), discoveries)
	}

	if got := discoveries[0]["file"]; got != "docs/tasks.md" {
		t.Fatalf("expected remaining markdown discovery from docs/tasks.md, got %v", got)
	}
}

func mustWriteFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func decodeDiscoveries(t *testing.T, result []framework.TextContent) []map[string]interface{} {
	t.Helper()

	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	raw, ok := data["discoveries"].([]interface{})
	if !ok {
		t.Fatalf("discoveries missing or invalid: %#v", data["discoveries"])
	}

	discoveries := make([]map[string]interface{}, 0, len(raw))
	for _, item := range raw {
		discovery, ok := item.(map[string]interface{})
		if !ok {
			t.Fatalf("invalid discovery item: %#v", item)
		}

		discoveries = append(discoveries, discovery)
	}

	return discoveries
}
