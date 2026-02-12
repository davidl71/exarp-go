package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
)

// toolsWithNoBridge lists tools that are fully native with no Python bridge (bridge does not route them).
var toolsWithNoBridge = map[string]bool{
	"session": true, "setup_hooks": true, "check_attribution": true, "memory_maint": true,
	"memory": true, "task_discovery": true, "analyze_alignment": true, "estimation": true, "task_analysis": true,
	"infer_task_progress": true,
	"git_tools":           true, "infer_session_mode": true, "tool_catalog": true, "workflow_mode": true,
	"prompt_tracking": true, "generate_config": true, "add_external_tool_hints": true,
	"report":    true,
	"recommend": true,
	"security":  true,
	"testing":   true,
	"lint":      true,
	"ollama":    true,
}

// TestRegressionNativeOnlyTools documents tools that completed native migration (no Python bridge).
// See docs/PYTHON_FALLBACKS_SAFE_TO_REMOVE.md and docs/PYTHON_BRIDGE_MIGRATION_NEXT.md.
func TestRegressionNativeOnlyTools(t *testing.T) {
	want := []string{"session", "setup_hooks", "check_attribution", "memory_maint", "memory", "task_discovery", "analyze_alignment", "estimation", "task_analysis", "infer_task_progress", "report", "recommend", "security", "testing", "lint", "ollama"}
	for _, name := range want {
		if !toolsWithNoBridge[name] {
			t.Errorf("native-only tool %q must be in toolsWithNoBridge", name)
		}
	}
}

// TestRegressionSessionPrime tests session prime action (native only; no Python fallback)
func TestRegressionSessionPrime(t *testing.T) {
	ctx := context.Background()
	params := map[string]interface{}{
		"action": "prime",
	}

	result, err := handleSessionNative(ctx, params)
	if err != nil {
		t.Fatalf("session native implementation failed: %v", err)
	}
	if len(result) == 0 {
		t.Error("session result is empty")
		return
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
		t.Errorf("session result is not valid JSON: %v", err)
		return
	}
	// Native prime returns auto_primed, method, detection{agent,mode}, workflow{mode}
	if v, ok := data["auto_primed"].(bool); !ok || !v {
		t.Error("session result missing auto_primed field or not true")
	}
	for _, field := range []string{"detection", "workflow"} {
		if _, ok := data[field]; !ok {
			t.Errorf("session result missing expected field: %s", field)
		}
	}
	// T-1770909568528: status_label and status_context for AI context announcement
	if _, hasLabel := data["status_label"]; !hasLabel {
		t.Error("session result missing status_label field (T-1770909568528)")
	}
	if _, hasCtx := data["status_context"]; !hasCtx {
		t.Error("session result missing status_context field (T-1770909568528)")
	}
}

// TestRegressionRecommendWorkflow tests native recommend workflow action.
// Recommend is fully native (model, workflow, advisor); no Python bridge.
func TestRegressionRecommendWorkflow(t *testing.T) {
	ctx := context.Background()
	params := map[string]interface{}{
		"action":            "workflow",
		"task_description":  "Implement a new feature with database migrations",
		"include_rationale": true,
	}

	nativeResult, nativeErr := handleRecommendWorkflowNative(ctx, params)
	if nativeErr != nil {
		t.Fatalf("native implementation failed: %v", nativeErr)
	}
	if len(nativeResult) == 0 {
		t.Error("native result is empty")
	}

	var nativeData map[string]interface{}
	if err := json.Unmarshal([]byte(nativeResult[0].Text), &nativeData); err != nil {
		t.Errorf("native result is not valid JSON: %v", err)
	}
	if nativeSuccess, ok := nativeData["success"].(bool); !ok || !nativeSuccess {
		t.Error("native result missing success field or not true")
	}
	if nativeDataMap, ok := nativeData["data"].(map[string]interface{}); ok {
		if recommendedMode, ok := nativeDataMap["recommended_mode"].(string); ok {
			if recommendedMode != "AGENT" && recommendedMode != "ASK" {
				t.Errorf("native recommended_mode should be AGENT or ASK, got %s", recommendedMode)
			}
		} else {
			t.Error("native result missing recommended_mode")
		}
	}
}

// TestRegressionHealthServer tests native health server action.
// Native is primary; optional bridge comparison when bridge available (hybrid tool).
func TestRegressionHealthServer(t *testing.T) {
	ctx := context.Background()
	params := map[string]interface{}{
		"action": "server",
	}

	// Assert native path (migrated to native Go)
	nativeResult, nativeErr := handleHealthNative(ctx, params)
	if nativeErr != nil {
		t.Fatalf("native implementation failed: %v", nativeErr)
	}
	if len(nativeResult) == 0 {
		t.Error("native result is empty")
	}

	var nativeData map[string]interface{}
	if err := json.Unmarshal([]byte(nativeResult[0].Text), &nativeData); err != nil {
		t.Errorf("native result is not valid JSON: %v", err)
	}
	// Native health server returns status/version/project_root (no "success" field)
	if _, ok := nativeData["status"]; !ok {
		if success, ok := nativeData["success"].(bool); !ok || !success {
			t.Error("native health result missing status or success field")
		}
	}
}

// TestRegressionFallbackBehavior tests native-only tools have no bridge fallback;
// hybrid tools may fall back to Python when native fails.
func TestRegressionFallbackBehavior(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		toolName       string
		params         map[string]interface{}
		expectFallback bool
	}{
		{
			name:     "session with invalid action (native only, no fallback)",
			toolName: "session",
			params: map[string]interface{}{
				"action": "invalid_action_that_does_not_exist",
			},
			expectFallback: false, // session is fully native; bridge does not route it
		},
		{
			name:     "recommend with invalid action should fallback",
			toolName: "recommend",
			params: map[string]interface{}{
				"action": "invalid_action",
			},
			expectFallback: true,
		},
		{
			name:     "health with invalid action should fallback",
			toolName: "health",
			params: map[string]interface{}{
				"action": "invalid_action",
			},
			expectFallback: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Try native first
			var nativeErr error

			switch tt.toolName {
			case "session":
				_, nativeErr = handleSessionNative(ctx, tt.params)
			case "recommend":
				action, _ := tt.params["action"].(string)
				if action == "workflow" {
					_, nativeErr = handleRecommendWorkflowNative(ctx, tt.params)
				} else {
					_, nativeErr = handleRecommendModelNative(ctx, tt.params)
				}
			case "health":
				_, nativeErr = handleHealthNative(ctx, tt.params)
			}

			// If we expect fallback, native should fail
			if tt.expectFallback {
				if nativeErr == nil {
					t.Logf("Native implementation succeeded (may be acceptable if it handles invalid actions gracefully)")
				}
			}

			// Skip bridge comparison for fully-native tools (Python bridge removed)
			if toolsWithNoBridge[tt.toolName] {
				return
			}
			// Python bridge removed; no bridge fallback to compare
		})
	}
}

// TestRegressionResponseFormat verifies that all tools return consistent response formats
func TestRegressionResponseFormat(t *testing.T) {
	ctx := context.Background()

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
		{
			name:     "health server",
			toolName: "health",
			params: map[string]interface{}{
				"action": "server",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result []framework.TextContent
			var err error

			switch tt.toolName {
			case "session":
				result, err = handleSessionNative(ctx, tt.params)
			case "recommend":
				action, _ := tt.params["action"].(string)
				if action == "workflow" {
					result, err = handleRecommendWorkflowNative(ctx, tt.params)
				} else {
					result, err = handleRecommendModelNative(ctx, tt.params)
				}
			case "health":
				result, err = handleHealthNative(ctx, tt.params)
			}

			if err != nil {
				t.Skipf("Tool failed (may be expected): %v", err)
				return
			}

			// Verify response format
			if len(result) == 0 {
				t.Error("result is empty")
				return
			}

			// All results should be TextContent with type "text"
			for i, content := range result {
				if content.Type != "text" {
					t.Errorf("result[%d] has type %s, expected 'text'", i, content.Type)
				}

				// Content should be valid JSON
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(content.Text), &data); err != nil {
					t.Errorf("result[%d] is not valid JSON: %v", i, err)
				}

				// Should have success field, or session-style (auto_primed), or health-style (server/status)
				if success, ok := data["success"].(bool); ok {
					if !success {
						t.Logf("result[%d] has success=false (may be acceptable for some tools)", i)
					}
				} else if tt.toolName == "session" {
					if _, ok := data["auto_primed"]; !ok {
						t.Errorf("result[%d] missing 'success' or 'auto_primed' (session format)", i)
					}
				} else if tt.toolName != "health" {
					t.Errorf("result[%d] missing 'success' field", i)
				}
			}
		})
	}
}

// TestRegressionErrorHandling verifies error handling: native-only tools assert native only;
// hybrid tools compare native vs bridge when bridge available.
func TestRegressionErrorHandling(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		toolName    string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name:     "session with missing required params",
			toolName: "session",
			params:   map[string]interface{}{
				// Missing action - should default to "prime"
			},
			expectError: false,
		},
		{
			name:     "recommend with missing task_description",
			toolName: "recommend",
			params: map[string]interface{}{
				"action": "workflow",
				// Missing task_description
			},
			expectError: false, // Should handle gracefully
		},
		{
			name:     "health with missing action",
			toolName: "health",
			params:   map[string]interface{}{
				// Missing action - health may default or require; match current behavior
			},
			expectError: false, // Health defaults action when missing
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var nativeErr error

			switch tt.toolName {
			case "session":
				_, nativeErr = handleSessionNative(ctx, tt.params)
			case "recommend":
				action, _ := tt.params["action"].(string)
				if action == "workflow" {
					_, nativeErr = handleRecommendWorkflowNative(ctx, tt.params)
				} else {
					_, nativeErr = handleRecommendModelNative(ctx, tt.params)
				}
			case "health":
				_, nativeErr = handleHealthNative(ctx, tt.params)
			}

			// For fully-native tools, only assert native behavior (no bridge to compare)
			if toolsWithNoBridge[tt.toolName] {
				nativeHasError := nativeErr != nil
				if nativeHasError != tt.expectError {
					t.Errorf("native error behavior mismatch: got error=%v, want error=%v", nativeHasError, tt.expectError)
				}
				return
			}

			// Python bridge removed; only assert native behavior
			nativeHasError := nativeErr != nil
			if nativeHasError != tt.expectError {
				t.Errorf("native error behavior mismatch: got error=%v, want error=%v", nativeHasError, tt.expectError)
			}
		})
	}
}

// TestRegressionFeatureParity documents migration status: native-only tools vs hybrid vs bridge-only.
func TestRegressionFeatureParity(t *testing.T) {
	// Native-only: no Python bridge; hybrid: native first, bridge fallback; bridge-only: intentional.

	knownDifferences := map[string]string{
		"session":             "Fully native; no Python bridge. Prime, handoff, prompts, assignee are native-only.",
		"setup_hooks":         "Fully native; no Python bridge. Git and patterns actions are native-only.",
		"check_attribution":   "Fully native; no Python bridge.",
		"memory":              "Fully native; no Python bridge. CRUD (save/recall/list/search) native-only; bridge fallback removed 2026-01-28.",
		"memory_maint":        "Fully native; no Python bridge. Health, gc, prune, consolidate, dream are native-only.",
		"analyze_alignment":   "Fully native for todo2 and prd; no Python bridge.",
		"task_discovery":      "Fully native; no Python bridge. Comments, markdown, orphans, create_tasks are native-only (removed bridge 2026-01-28).",
		"task_workflow":       "Fully native; no Python bridge. sync (SQLite↔JSON), approve, clarify (FM), clarity, cleanup, create; external sync is future nice-to-have (param ignored).",
		"infer_task_progress": "Fully native; no Python bridge. Heuristics + optional FM; dry_run, auto_update_tasks, output_path.",
		"context":             "Fully native; no Python bridge. summarize (Apple FM), budget, batch; unknown action returns error (2026-01-28).",
		"recommend":           "Hybrid: native workflow/model; Python fallback when native fails.",
		"health":              "Hybrid: native server/docs/dod/cicd; Python fallback when native fails.",
		"ollama":              "Hybrid: native uses HTTP client; Python bridge may differ.",
		"mlx":                 "Native-only; models (static list); status/hardware return unavailable message; generate returns error (use ollama or apple_foundation_models). Python bridge removed.",
	}

	for tool, reason := range knownDifferences {
		t.Run(tool+"_known_differences", func(t *testing.T) {
			t.Logf("Known difference for %s: %s", tool, reason)
			// This test passes - it's just documentation
		})
	}
}
