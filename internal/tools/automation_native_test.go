package tools

import (
	"context"
	"encoding/json"
	"testing"
)

func TestHandleAutomationNative(t *testing.T) {
	tests := []struct {
		name      string
		action    string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name:      "daily action",
			action:    "daily",
			params:    map[string]interface{}{"action": "daily"},
			wantError: false,
		},
		{
			name:      "discover action",
			action:    "discover",
			params:    map[string]interface{}{"action": "discover"},
			wantError: false,
		},
		{
			name:      "nightly action",
			action:    "nightly",
			params:    map[string]interface{}{"action": "nightly"},
			wantError: false,
		},
		{
			name:      "sprint action",
			action:    "sprint",
			params:    map[string]interface{}{"action": "sprint"},
			wantError: false,
		},
		{
			name:      "unknown action",
			action:    "unknown",
			params:    map[string]interface{}{"action": "unknown"},
			wantError: true,
		},
		{
			name:      "empty action defaults to daily",
			action:    "",
			params:    map[string]interface{}{},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := handleAutomationNative(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleAutomationNative() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError {
				if result == nil || len(result) == 0 {
					t.Error("handleAutomationNative() returned empty result")
					return
				}
				// Verify result is valid JSON
				var resultData map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &resultData); err != nil {
					t.Errorf("handleAutomationNative() returned invalid JSON: %v", err)
				}
			}
		})
	}
}

func TestHandleAutomationDaily(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name:      "valid daily action",
			params:    map[string]interface{}{"action": "daily"},
			wantError: false,
		},
		{
			name:      "empty params",
			params:    map[string]interface{}{},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := handleAutomationDaily(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleAutomationDaily() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && (result == nil || len(result) == 0) {
				t.Error("handleAutomationDaily() returned empty result")
			}
		})
	}
}

func TestHandleAutomationNightly(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name: "valid nightly action",
			params: map[string]interface{}{
				"action":             "nightly",
				"max_tasks_per_host": 5,
				"max_parallel_tasks": 10,
				"dry_run":            true,
			},
			wantError: false,
		},
		{
			name:      "empty params",
			params:    map[string]interface{}{"action": "nightly"},
			wantError: false,
		},
		{
			name: "with priority filter",
			params: map[string]interface{}{
				"action":          "nightly",
				"priority_filter": "high",
				"dry_run":         true,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := handleAutomationNightly(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleAutomationNightly() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && (result == nil || len(result) == 0) {
				t.Error("handleAutomationNightly() returned empty result")
			}
		})
	}
}

func TestHandleAutomationSprint(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name: "valid sprint action",
			params: map[string]interface{}{
				"action":              "sprint",
				"max_iterations":      1,
				"auto_approve":        false,
				"extract_subtasks":    false,
				"run_analysis_tools":  false,
				"run_testing_tools":   false,
				"dry_run":             true,
			},
			wantError: false,
		},
		{
			name:      "empty params",
			params:    map[string]interface{}{"action": "sprint"},
			wantError: false,
		},
		{
			name: "sprint with all features enabled",
			params: map[string]interface{}{
				"action":              "sprint",
				"max_iterations":      2,
				"auto_approve":        true,
				"extract_subtasks":    true,
				"run_analysis_tools":  true,
				"run_testing_tools":   true,
				"dry_run":             true,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := handleAutomationSprint(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleAutomationSprint() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && (result == nil || len(result) == 0) {
				t.Error("handleAutomationSprint() returned empty result")
				return
			}
			if !tt.wantError {
				// Verify result contains expected fields
				var resultData map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &resultData); err == nil {
					if action, ok := resultData["action"].(string); !ok || action != "sprint" {
						t.Errorf("handleAutomationSprint() result action = %v, want 'sprint'", action)
					}
				}
			}
		})
	}
}

func TestHandleAutomationDiscover(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name:      "valid discover action",
			params:    map[string]interface{}{"action": "discover"},
			wantError: false,
		},
		{
			name:      "empty params",
			params:    map[string]interface{}{},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := handleAutomationDiscover(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleAutomationDiscover() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && (result == nil || len(result) == 0) {
				t.Error("handleAutomationDiscover() returned empty result")
			}
		})
	}
}
