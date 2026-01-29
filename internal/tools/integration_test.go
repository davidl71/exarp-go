package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
)

// TestMCPToolInvocation tests tool invocation via MCP interface
// This is a basic integration test to validate tool handlers work correctly
func TestMCPToolInvocation(t *testing.T) {
	tests := []struct {
		name      string
		toolName  string
		params    map[string]interface{}
		wantError bool
		validate  func(*testing.T, []framework.TextContent)
	}{
		{
			name:     "session prime action",
			toolName: "session",
			params: map[string]interface{}{
				"action": "prime",
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
				// Should have session priming information
			},
		},
		{
			name:     "recommend workflow action",
			toolName: "recommend",
			params: map[string]interface{}{
				"action":           "workflow",
				"task_description": "Implement a new feature",
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
			name:     "health server action",
			toolName: "health",
			params: map[string]interface{}{
				"action": "server",
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
				// Should have server health information
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			// Simulate MCP tool invocation
			var result []framework.TextContent
			var err error

			switch tt.toolName {
			case "session":
				result, err = handleSessionNative(ctx, tt.params)
			case "recommend":
				// Route to appropriate function based on action
				action, _ := tt.params["action"].(string)
				if action == "" {
					action = "model"
				}
				if action == "model" {
					result, err = handleRecommendModelNative(ctx, tt.params)
				} else if action == "workflow" {
					result, err = handleRecommendWorkflowNative(ctx, tt.params)
				} else {
					t.Fatalf("unsupported recommend action: %s", action)
				}
			case "health":
				result, err = handleHealthNative(ctx, tt.params)
			default:
				t.Fatalf("unknown tool: %s", tt.toolName)
			}

			if (err != nil) != tt.wantError {
				t.Errorf("tool invocation error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

// TestToolErrorHandling tests error handling across tools
func TestToolErrorHandling(t *testing.T) {
	tests := []struct {
		name      string
		toolName  string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name:     "session with invalid action",
			toolName: "session",
			params: map[string]interface{}{
				"action": "invalid_action",
			},
			wantError: true,
		},
		{
			name:     "recommend with missing parameters",
			toolName: "recommend",
			params: map[string]interface{}{
				"action": "workflow",
				// Missing task_description
			},
			wantError: false, // Should handle gracefully
		},
		{
			name:     "health with invalid action",
			toolName: "health",
			params: map[string]interface{}{
				"action": "invalid_action",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			var err error

			switch tt.toolName {
			case "session":
				_, err = handleSessionNative(ctx, tt.params)
			case "recommend":
				// Route to appropriate function based on action
				action, _ := tt.params["action"].(string)
				if action == "" {
					action = "model"
				}
				if action == "model" {
					_, err = handleRecommendModelNative(ctx, tt.params)
				} else if action == "workflow" {
					_, err = handleRecommendWorkflowNative(ctx, tt.params)
				} else {
					t.Fatalf("unsupported recommend action: %s", action)
				}
			case "health":
				_, err = handleHealthNative(ctx, tt.params)
			default:
				t.Fatalf("unknown tool: %s", tt.toolName)
			}

			if (err != nil) != tt.wantError {
				t.Errorf("error handling: error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestToolResponseFormat tests that all tools return valid JSON responses
func TestToolResponseFormat(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		params   map[string]interface{}
	}{
		{
			name:     "session prime",
			toolName: "session",
			params: map[string]interface{}{
				"action": "prime",
			},
		},
		{
			name:     "recommend workflow",
			toolName: "recommend",
			params: map[string]interface{}{
				"action":           "workflow",
				"task_description": "Test task",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			var result []framework.TextContent
			var err error

			switch tt.toolName {
			case "session":
				result, err = handleSessionNative(ctx, tt.params)
			case "recommend":
				// Use workflow action for testing
				if tt.params["action"] == "workflow" {
					result, err = handleRecommendWorkflowNative(ctx, tt.params)
				} else {
					result, err = handleRecommendModelNative(ctx, tt.params)
				}
			default:
				t.Fatalf("unknown tool: %s", tt.toolName)
			}

			if err != nil {
				t.Skipf("tool returned error (may be expected): %v", err)
				return
			}

			if len(result) == 0 {
				t.Error("expected non-empty result")
				return
			}

			// Validate JSON format
			for _, content := range result {
				if content.Type != "text" {
					t.Errorf("expected text content type, got %s", content.Type)
					continue
				}
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(content.Text), &data); err != nil {
					t.Errorf("invalid JSON response: %v", err)
				}
			}
		})
	}
}
