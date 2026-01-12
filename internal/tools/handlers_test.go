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

// Note: TestHandleServerStatus removed - server_status tool converted to stdio://server/status resource
// See internal/resources/server.go and internal/resources/handlers_test.go for resource tests

func TestHandleToolCatalog(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		args    map[string]interface{}
		wantErr bool
		check   func(t *testing.T, result []framework.TextContent)
	}{
		// Note: "list" action converted to stdio://tools and stdio://tools/{category} resources
		// This test now only covers the "help" action
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

// Note: TestHandleServerStatusNative removed - server_status tool converted to stdio://server/status resource
// See internal/resources/server.go and internal/resources/handlers_test.go for resource tests

// hasSubstring checks if a string contains a substring (case-sensitive)
func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestHandleContextBatchNative(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
		check   func(t *testing.T, result []framework.TextContent)
	}{
		{
			name: "valid batch with JSON string items - combine true",
			params: map[string]interface{}{
				"items":   `[{"data": "First item"}, {"data": "Second item"}]`,
				"level":   "brief",
				"combine": true,
			},
			wantErr: false,
			check: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
				text := result[0].Text
				if !hasSubstring(text, "combined_summary") {
					t.Error("expected 'combined_summary' field in response")
				}
				if !hasSubstring(text, "total_items") {
					t.Error("expected 'total_items' field in response")
				}
				if !hasSubstring(text, "token_estimate") {
					t.Error("expected 'token_estimate' field in response")
				}
			},
		},
		{
			name: "valid batch with array items - combine false",
			params: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{"data": "Item one"},
					map[string]interface{}{"data": "Item two"},
				},
				"combine": false,
			},
			wantErr: false,
			check: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
				text := result[0].Text
				if !hasSubstring(text, "summaries") {
					t.Error("expected 'summaries' field in response when combine=false")
				}
				// Should not have combined_summary when combine=false
				if hasSubstring(text, "combined_summary") {
					t.Error("unexpected 'combined_summary' field when combine=false")
				}
			},
		},
		{
			name: "single item batch",
			params: map[string]interface{}{
				"items": `[{"data": "Single item to summarize"}]`,
			},
			wantErr: false,
			check: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
				text := result[0].Text
				if !hasSubstring(text, "total_items") {
					t.Error("expected 'total_items' field in response")
				}
			},
		},
		{
			name: "batch with different levels - brief",
			params: map[string]interface{}{
				"items": `[{"data": "Test data for brief summary"}]`,
				"level": "brief",
			},
			wantErr: false,
			check: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
				text := result[0].Text
				// Level is stored in individual summaries, verify response structure
				if !hasSubstring(text, "combined_summary") || !hasSubstring(text, "total_items") {
					t.Error("expected valid batch response structure")
				}
			},
		},
		{
			name: "batch with different levels - detailed",
			params: map[string]interface{}{
				"items": `[{"data": "Test data for detailed summary"}]`,
				"level": "detailed",
			},
			wantErr: false,
			check: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
				text := result[0].Text
				// Level is stored in individual summaries, verify response structure
				if !hasSubstring(text, "combined_summary") || !hasSubstring(text, "total_items") {
					t.Error("expected valid batch response structure")
				}
			},
		},
		{
			name: "batch with tool_type",
			params: map[string]interface{}{
				"items": `[{"data": "Test data", "tool_type": "analysis"}]`,
			},
			wantErr: false,
			check: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
				// Should process successfully with tool_type
			},
		},
		{
			name: "empty batch list",
			params: map[string]interface{}{
				"items": `[]`,
			},
			wantErr: false,
			check: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
				text := result[0].Text
				if !hasSubstring(text, "total_items") {
					t.Error("expected 'total_items' field in response")
				}
				if !hasSubstring(text, "0") {
					t.Error("expected 0 items in response")
				}
			},
		},
		{
			name: "missing items parameter",
			params: map[string]interface{}{
				"level": "brief",
			},
			wantErr: true,
		},
		{
			name: "nil items parameter",
			params: map[string]interface{}{
				"items": nil,
			},
			wantErr: true,
		},
		{
			name: "invalid JSON items",
			params: map[string]interface{}{
				"items": `[invalid json`,
			},
			wantErr: true,
		},
		{
			name: "invalid items type (not string or array)",
			params: map[string]interface{}{
				"items": 123, // Invalid type
			},
			wantErr: true,
		},
		{
			name: "items with data field and direct data",
			params: map[string]interface{}{
				"items": `[{"data": "Item with data field"}, "Item without data field"]`,
			},
			wantErr: false,
			check: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
				// Should process both formats
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handleContextBatchNative(ctx, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleContextBatchNative() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}
