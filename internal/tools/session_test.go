package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
)

func TestHandleSessionPrompts(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		validate  func(*testing.T, []framework.TextContent)
	}{
		{
			name: "all prompts",
			params: map[string]interface{}{
				"action": "prompts",
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
				if method, ok := data["method"].(string); !ok || method != "native_go" {
					t.Error("expected method=native_go")
				}
			},
		},
		{
			name: "filter by mode",
			params: map[string]interface{}{
				"action": "prompts",
				"mode":   "daily_checkin",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				if filters, ok := data["filters_applied"].(map[string]interface{}); ok {
					if mode, ok := filters["mode"].(string); !ok || mode != "daily_checkin" {
						t.Errorf("expected mode filter, got %v", filters)
					}
				}
			},
		},
		{
			name: "filter by category",
			params: map[string]interface{}{
				"action":   "prompts",
				"category": "workflow",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				if filters, ok := data["filters_applied"].(map[string]interface{}); ok {
					if category, ok := filters["category"].(string); !ok || category != "workflow" {
						t.Errorf("expected category filter, got %v", filters)
					}
				}
			},
		},
		{
			name: "filter by keywords",
			params: map[string]interface{}{
				"action":   "prompts",
				"keywords": []string{"align", "discover"},
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				if filters, ok := data["filters_applied"].(map[string]interface{}); ok {
					if keywords, ok := filters["keywords"].([]interface{}); !ok || len(keywords) == 0 {
						t.Errorf("expected keywords filter, got %v", filters)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := handleSessionPrompts(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleSessionPrompts() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestHandleSessionAssignee(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		validate  func(*testing.T, []framework.TextContent)
	}{
		{
			name: "basic assignee request",
			params: map[string]interface{}{
				"action": "assignee",
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
				if method, ok := data["method"].(string); !ok || method != "native_go" {
					t.Error("expected method=native_go")
				}
			},
		},
		{
			name: "filter by status",
			params: map[string]interface{}{
				"action":        "assignee",
				"status_filter": "Todo",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				if filters, ok := data["filters_applied"].(map[string]interface{}); ok {
					if status, ok := filters["status_filter"].(string); !ok || status != "Todo" {
						t.Errorf("expected status filter, got %v", filters)
					}
				}
			},
		},
		{
			name: "filter by priority",
			params: map[string]interface{}{
				"action":          "assignee",
				"priority_filter": "high",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				if filters, ok := data["filters_applied"].(map[string]interface{}); ok {
					if priority, ok := filters["priority_filter"].(string); !ok || priority != "high" {
						t.Errorf("expected priority filter, got %v", filters)
					}
				}
			},
		},
		{
			name: "with assignee name",
			params: map[string]interface{}{
				"action":        "assignee",
				"assignee_name": "test-agent",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				if name, ok := data["assignee_name"].(string); !ok || name != "test-agent" {
					t.Errorf("expected assignee_name=test-agent, got %v", data["assignee_name"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := handleSessionAssignee(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleSessionAssignee() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestHandleSessionNative(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name: "prompts action",
			params: map[string]interface{}{
				"action": "prompts",
			},
			wantError: false,
		},
		{
			name: "assignee action",
			params: map[string]interface{}{
				"action": "assignee",
			},
			wantError: false,
		},
		{
			name: "prime action",
			params: map[string]interface{}{
				"action": "prime",
			},
			wantError: false,
		},
		{
			name: "handoff action",
			params: map[string]interface{}{
				"action": "handoff",
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
			result, err := handleSessionNative(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleSessionNative() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && (result == nil || len(result) == 0) {
				t.Error("expected non-empty result")
			}
		})
	}
}
