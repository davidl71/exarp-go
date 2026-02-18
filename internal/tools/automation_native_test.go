package tools

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

// automationTestTimeout is the max duration per automation subtest. Prevents
// long-running tests (e.g. discover with Apple FM, sprint with many subtasks) from blocking the suite.
// 25s keeps failures fast vs package default 90s.
const automationTestTimeout = 25 * time.Second

// runWithTimeout runs fn in a goroutine and fails the test if it exceeds d.
// Use for subtests that may block on I/O or Apple FM so they fail fast instead of waiting for package timeout.
func runWithTimeout(t *testing.T, d time.Duration, fn func()) {
	t.Helper()

	done := make(chan struct{}, 1)

	go func() {
		defer func() { done <- struct{}{} }()

		fn()
	}()

	select {
	case <-done:
		return
	case <-time.After(d):
		t.Fatalf("timeout: test exceeded %v (use -short to skip long-running tests)", d)
	}
}

func TestHandleAutomationNative(t *testing.T) {
	tests := []struct {
		name      string
		action    string
		params    map[string]interface{}
		wantError bool
		longRun   bool // longRun tests are skipped when -short is set
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
			params:    map[string]interface{}{"action": "discover", "use_llm": false},
			wantError: false,
			longRun:   true, // discover can block on Apple FM when CGO=1
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
			longRun:   true, // sprint runs many subtasks and can exceed quick timeout
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
			if tt.longRun && testing.Short() {
				t.Skip("skipping long-running discover in short mode")
			}

			runWithTimeout(t, automationTestTimeout, func() {
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
				"action":             "sprint",
				"max_iterations":     1,
				"auto_approve":       false,
				"extract_subtasks":   false,
				"run_analysis_tools": false,
				"run_testing_tools":  false,
				"dry_run":            true,
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
				"action":             "sprint",
				"max_iterations":     2,
				"auto_approve":       true,
				"extract_subtasks":   true,
				"run_analysis_tools": true,
				"run_testing_tools":  true,
				"dry_run":            true,
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
				// Verify result contains expected fields (action is under results)
				var resultData map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &resultData); err == nil {
					if results, ok := resultData["results"].(map[string]interface{}); ok {
						if action, ok := results["action"].(string); !ok || action != "sprint" {
							t.Errorf("handleAutomationSprint() result results.action = %v, want 'sprint'", action)
						}
					} else if action, ok := resultData["action"].(string); !ok || action != "sprint" {
						t.Errorf("handleAutomationSprint() result action = %v, want 'sprint'", action)
					}
				}
			}
		})
	}
}

func TestHandleAutomationDiscover(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping discover test in short mode (can block on Apple FM when CGO=1)")
	}

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name:      "valid discover action",
			params:    map[string]interface{}{"action": "discover", "use_llm": false},
			wantError: false,
		},
		{
			name:      "empty params",
			params:    map[string]interface{}{"use_llm": false},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runWithTimeout(t, automationTestTimeout, func() {
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
		})
	}
}
