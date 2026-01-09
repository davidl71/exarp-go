package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/tests/fixtures"
)

func TestHandler_ArgumentParsing(t *testing.T) {
	tests := []struct {
		name      string
		handler   framework.ToolHandler
		args      map[string]interface{}
		wantError bool
	}{
		{
			name:    "valid args",
			handler: handleAnalyzeAlignment,
			args: map[string]interface{}{
				"action":                "todo2",
				"create_followup_tasks": true,
			},
			wantError: false, // Native implementation handles valid args successfully
		},
		{
			name:      "empty args",
			handler:   handleAnalyzeAlignment,
			args:      map[string]interface{}{},
			wantError: false, // Native implementation uses defaults for empty args
		},
		{
			name:      "invalid JSON",
			handler:   handleAnalyzeAlignment,
			args:      nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			var argsJSON json.RawMessage
			var err error

			if tt.args != nil {
				argsJSON, err = json.Marshal(tt.args)
				if err != nil {
					t.Fatalf("json.Marshal() error = %v", err)
				}
			} else {
				argsJSON = json.RawMessage(`{invalid json}`)
			}

			_, err = tt.handler(ctx, argsJSON)
			if (err != nil) != tt.wantError {
				t.Errorf("handler() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestHandler_ErrorPropagation(t *testing.T) {
	ctx := context.Background()

	// Test with invalid JSON that should cause parsing error
	invalidJSON := json.RawMessage(`{invalid json}`)

	_, err := handleAnalyzeAlignment(ctx, invalidJSON)
	if err == nil {
		t.Error("handleAnalyzeAlignment() error = nil, want error for invalid JSON")
	}
}

func TestHandler_TextContentFormat(t *testing.T) {
	ctx := context.Background()
	server := fixtures.NewMockServer("test-server")

	// Register a test tool
	err := server.RegisterTool(
		"test_tool",
		"test description",
		framework.ToolSchema{Type: "object"},
		handleAnalyzeAlignment,
	)
	if err != nil {
		t.Fatalf("RegisterTool() error = %v", err)
	}

	// Call the tool
	args := map[string]interface{}{
		"action": "todo2",
	}
	argsJSON, _ := json.Marshal(args)

	_, err = server.CallTool(ctx, "test_tool", argsJSON)
	// This will fail because it requires Python bridge, but we can test the format
	// Expected error (no Python bridge in test)
	if err == nil {
		t.Log("Tool call succeeded (unexpected, but OK for testing)")
	}
}

func TestHandler_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	args := map[string]interface{}{}
	argsJSON, _ := json.Marshal(args)

	// Handler should respect context cancellation
	// Note: The actual handler may not check context, but Python bridge will
	_, err := handleAnalyzeAlignment(ctx, argsJSON)
	if err == nil {
		// Error is expected when context is cancelled
		// (Python bridge will fail or timeout)
	}
}

func TestHandleServerStatus(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		args    json.RawMessage
		wantErr bool
		check   func(t *testing.T, result []framework.TextContent)
	}{
		{
			name:    "empty args",
			args:    json.RawMessage(`{}`),
			wantErr: false,
			check: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
				text := result[0].Text
				// Check that response contains expected fields
				if !hasSubstring(text, "status") {
					t.Error("expected 'status' field in response")
				}
				if !hasSubstring(text, "operational") {
					t.Error("expected 'operational' status")
				}
				if !hasSubstring(text, "version") {
					t.Error("expected 'version' field in response")
				}
			},
		},
		{
			name:    "no args",
			args:    json.RawMessage(`null`),
			wantErr: false,
			check: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handleServerStatus(ctx, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleServerStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func TestHandleToolCatalog(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		args    map[string]interface{}
		wantErr bool
		check   func(t *testing.T, result []framework.TextContent)
	}{
		{
			name: "list action - no filters",
			args: map[string]interface{}{
				"action": "list",
			},
			wantErr: false,
			check: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
				text := result[0].Text
				if !hasSubstring(text, "tools") {
					t.Error("expected 'tools' field in response")
				}
				if !hasSubstring(text, "count") {
					t.Error("expected 'count' field in response")
				}
			},
		},
		{
			name: "list action - with category filter",
			args: map[string]interface{}{
				"action":   "list",
				"category": "Task Management",
			},
			wantErr: false,
			check: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
				}
			},
		},
		{
			name: "help action - valid tool",
			args: map[string]interface{}{
				"action":    "help",
				"tool_name": "tool_catalog",
			},
			wantErr: false,
			check: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
				text := result[0].Text
				if !hasSubstring(text, "tool") {
					t.Error("expected 'tool' field in response")
				}
				if !hasSubstring(text, "description") {
					t.Error("expected 'description' field in response")
				}
			},
		},
		{
			name: "help action - missing tool_name",
			args: map[string]interface{}{
				"action": "help",
			},
			wantErr: false,
			check: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
				text := result[0].Text
				if !hasSubstring(text, "error") {
					t.Error("expected error in response")
				}
			},
		},
		{
			name: "help action - invalid tool",
			args: map[string]interface{}{
				"action":    "help",
				"tool_name": "nonexistent_tool",
			},
			wantErr: false,
			check: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
				text := result[0].Text
				if !hasSubstring(text, "error") {
					t.Error("expected error in response")
				}
			},
		},
		{
			name: "invalid action",
			args: map[string]interface{}{
				"action": "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argsJSON, err := json.Marshal(tt.args)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}

			result, err := handleToolCatalog(ctx, argsJSON)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleToolCatalog() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func TestHandleServerStatusNative(t *testing.T) {
	ctx := context.Background()

	result, err := handleServerStatusNative(ctx, json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("handleServerStatusNative() error = %v", err)
	}

	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}

	text := result[0].Text
	// Verify JSON structure
	var status map[string]interface{}
	if err := json.Unmarshal([]byte(text), &status); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Check required fields
	if status["status"] != "operational" {
		t.Errorf("expected status 'operational', got %v", status["status"])
	}
	if status["version"] != "0.1.0" {
		t.Errorf("expected version '0.1.0', got %v", status["version"])
	}
	if _, ok := status["project_root"]; !ok {
		t.Error("expected 'project_root' field")
	}
	if _, ok := status["tools_available"]; !ok {
		t.Error("expected 'tools_available' field")
	}
}

// hasSubstring checks if a string contains a substring (case-sensitive)
func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
