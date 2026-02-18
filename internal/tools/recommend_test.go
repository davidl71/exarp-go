package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
)

func TestHandleRecommendWorkflowNative(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		validate  func(*testing.T, []framework.TextContent)
	}{
		{
			name: "simple task description",
			params: map[string]interface{}{
				"action":           "workflow",
				"task_description": "Fix a bug in the code",
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

				if recommendedMode, ok := data["data"].(map[string]interface{})["recommended_mode"].(string); !ok {
					t.Error("expected recommended_mode in result")
				} else if recommendedMode != "ASK" && recommendedMode != "AGENT" {
					t.Errorf("expected recommended_mode to be ASK or AGENT, got %s", recommendedMode)
				}
			},
		},
		{
			name: "complex task description",
			params: map[string]interface{}{
				"action":           "workflow",
				"task_description": "Implement a new feature with multiple components, database migrations, API endpoints, and frontend integration",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}

				if dataMap, ok := data["data"].(map[string]interface{}); ok {
					if recommendedMode, ok := dataMap["recommended_mode"].(string); ok {
						// Complex tasks should likely recommend AGENT mode
						if recommendedMode != "AGENT" && recommendedMode != "ASK" {
							t.Errorf("unexpected recommended_mode: %s", recommendedMode)
						}
					}
				}
			},
		},
		{
			name: "with tags",
			params: map[string]interface{}{
				"action":           "workflow",
				"task_description": "Refactor code",
				"tags":             []string{"refactoring", "backend"},
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				// Should have considered tags in recommendation
			},
		},
		{
			name: "with include_rationale",
			params: map[string]interface{}{
				"action":            "workflow",
				"task_description":  "Write unit tests",
				"include_rationale": true,
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}

				if dataMap, ok := data["data"].(map[string]interface{}); ok {
					if _, ok := dataMap["rationale"]; !ok {
						t.Error("expected rationale in result when include_rationale=true")
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			result, err := handleRecommendWorkflowNative(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleRecommendWorkflowNative() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

// Note: findBestWorkflowMode and handleRecommendNative are internal functions
// They are tested indirectly through handleRecommendWorkflowNative tests above
