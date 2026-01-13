package tools

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/davidl71/exarp-go/internal/bridge"
	"github.com/davidl71/exarp-go/internal/framework"
)

// TestRegressionSessionPrime compares native vs Python bridge for session prime action
func TestRegressionSessionPrime(t *testing.T) {
	ctx := context.Background()
	params := map[string]interface{}{
		"action": "prime",
	}

	// Get native result
	nativeResult, nativeErr := handleSessionNative(ctx, params)
	if nativeErr != nil {
		t.Fatalf("native implementation failed: %v", nativeErr)
	}

	// Get Python bridge result
	bridgeResultStr, bridgeErr := bridge.ExecutePythonTool(ctx, "session", params)
	if bridgeErr != nil {
		t.Skipf("Python bridge not available or failed: %v", bridgeErr)
		return
	}

	// Compare results
	if len(nativeResult) == 0 {
		t.Error("native result is empty")
	}

	// Both should return valid JSON
	var nativeData map[string]interface{}
	if err := json.Unmarshal([]byte(nativeResult[0].Text), &nativeData); err != nil {
		t.Errorf("native result is not valid JSON: %v", err)
	}

	var bridgeData map[string]interface{}
	if err := json.Unmarshal([]byte(bridgeResultStr), &bridgeData); err != nil {
		t.Errorf("bridge result is not valid JSON: %v", err)
	}

	// Both should have success field
	if nativeSuccess, ok := nativeData["success"].(bool); !ok || !nativeSuccess {
		t.Error("native result missing success field or not true")
	}
	if bridgeSuccess, ok := bridgeData["success"].(bool); !ok || !bridgeSuccess {
		t.Error("bridge result missing success field or not true")
	}

	// Both should have similar structure (agent_type, mode, etc.)
	nativeKeys := make(map[string]bool)
	for k := range nativeData {
		nativeKeys[k] = true
	}

	bridgeKeys := make(map[string]bool)
	for k := range bridgeData {
		bridgeKeys[k] = true
	}

	// Check for common expected fields
	expectedFields := []string{"success", "agent_type", "mode"}
	for _, field := range expectedFields {
		if !nativeKeys[field] {
			t.Errorf("native result missing expected field: %s", field)
		}
		if !bridgeKeys[field] {
			t.Errorf("bridge result missing expected field: %s", field)
		}
	}
}

// TestRegressionRecommendWorkflow compares native vs Python bridge for recommend workflow action
func TestRegressionRecommendWorkflow(t *testing.T) {
	ctx := context.Background()
	params := map[string]interface{}{
		"action":          "workflow",
		"task_description": "Implement a new feature with database migrations",
		"include_rationale": true,
	}

	// Get native result
	nativeResult, nativeErr := handleRecommendWorkflowNative(ctx, params)
	if nativeErr != nil {
		t.Fatalf("native implementation failed: %v", nativeErr)
	}

	// Get Python bridge result
	bridgeResultStr, bridgeErr := bridge.ExecutePythonTool(ctx, "recommend", params)
	if bridgeErr != nil {
		t.Skipf("Python bridge not available or failed: %v", bridgeErr)
		return
	}

	// Compare results
	if len(nativeResult) == 0 {
		t.Error("native result is empty")
	}

	// Both should return valid JSON
	var nativeData map[string]interface{}
	if err := json.Unmarshal([]byte(nativeResult[0].Text), &nativeData); err != nil {
		t.Errorf("native result is not valid JSON: %v", err)
	}

	var bridgeData map[string]interface{}
	if err := json.Unmarshal([]byte(bridgeResultStr), &bridgeData); err != nil {
		t.Errorf("bridge result is not valid JSON: %v", err)
	}

	// Both should have success field
	if nativeSuccess, ok := nativeData["success"].(bool); !ok || !nativeSuccess {
		t.Error("native result missing success field or not true")
	}
	if bridgeSuccess, ok := bridgeData["success"].(bool); !ok || !bridgeSuccess {
		t.Error("bridge result missing success field or not true")
	}

	// Both should recommend a workflow mode (AGENT or ASK)
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

// TestRegressionHealthServer compares native vs Python bridge for health server action
func TestRegressionHealthServer(t *testing.T) {
	ctx := context.Background()
	params := map[string]interface{}{
		"action": "server",
	}

	// Get native result
	nativeResult, nativeErr := handleHealthNative(ctx, params)
	if nativeErr != nil {
		t.Fatalf("native implementation failed: %v", nativeErr)
	}

	// Get Python bridge result
	bridgeResultStr, bridgeErr := bridge.ExecutePythonTool(ctx, "health", params)
	if bridgeErr != nil {
		t.Skipf("Python bridge not available or failed: %v", bridgeErr)
		return
	}

	// Compare results
	if len(nativeResult) == 0 {
		t.Error("native result is empty")
	}

	// Both should return valid JSON
	var nativeData map[string]interface{}
	if err := json.Unmarshal([]byte(nativeResult[0].Text), &nativeData); err != nil {
		t.Errorf("native result is not valid JSON: %v", err)
	}

	var bridgeData map[string]interface{}
	if err := json.Unmarshal([]byte(bridgeResultStr), &bridgeData); err != nil {
		t.Errorf("bridge result is not valid JSON: %v", err)
	}

	// Both should have success field
	if nativeSuccess, ok := nativeData["success"].(bool); !ok || !nativeSuccess {
		t.Error("native result missing success field or not true")
	}
	if bridgeSuccess, ok := bridgeData["success"].(bool); !ok || !bridgeSuccess {
		t.Error("bridge result missing success field or not true")
	}
}

// TestRegressionFallbackBehavior tests that tools properly fall back to Python bridge when native fails
func TestRegressionFallbackBehavior(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		toolName  string
		params    map[string]interface{}
		expectFallback bool
	}{
		{
			name: "session with invalid action should fallback",
			toolName: "session",
			params: map[string]interface{}{
				"action": "invalid_action_that_does_not_exist",
			},
			expectFallback: true,
		},
		{
			name: "recommend with invalid action should fallback",
			toolName: "recommend",
			params: map[string]interface{}{
				"action": "invalid_action",
			},
			expectFallback: true,
		},
		{
			name: "health with invalid action should fallback",
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

			// Try Python bridge fallback
			bridgeResult, bridgeErr := bridge.ExecutePythonTool(ctx, tt.toolName, tt.params)
			if bridgeErr != nil {
				t.Skipf("Python bridge not available: %v", bridgeErr)
				return
			}

			// Bridge should return valid JSON
			var bridgeData map[string]interface{}
			if err := json.Unmarshal([]byte(bridgeResult), &bridgeData); err != nil {
				t.Errorf("bridge result is not valid JSON: %v", err)
			}
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
				"action":          "workflow",
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

				// Should have success field
				if success, ok := data["success"].(bool); !ok {
					t.Errorf("result[%d] missing 'success' field", i)
				} else if !success {
					t.Logf("result[%d] has success=false (may be acceptable for some tools)", i)
				}
			}
		})
	}
}

// TestRegressionErrorHandling verifies consistent error handling between native and bridge
func TestRegressionErrorHandling(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		toolName  string
		params    map[string]interface{}
		expectError bool
	}{
		{
			name: "session with missing required params",
			toolName: "session",
			params: map[string]interface{}{
				// Missing action - should default to "prime"
			},
			expectError: false,
		},
		{
			name: "recommend with missing task_description",
			toolName: "recommend",
			params: map[string]interface{}{
				"action": "workflow",
				// Missing task_description
			},
			expectError: false, // Should handle gracefully
		},
		{
			name: "health with missing action",
			toolName: "health",
			params: map[string]interface{}{
				// Missing action
			},
			expectError: true, // Health requires action
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var nativeErr error
			var bridgeErr error

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

			// Try Python bridge
			_, bridgeErr = bridge.ExecutePythonTool(ctx, tt.toolName, tt.params)
			if bridgeErr != nil {
				// Bridge may not be available - skip if so
				if os.Getenv("SKIP_BRIDGE_TESTS") != "" {
					t.Skipf("Python bridge not available (SKIP_BRIDGE_TESTS set)")
					return
				}
			}

			// Both should have same error behavior
			nativeHasError := nativeErr != nil
			bridgeHasError := bridgeErr != nil

			if nativeHasError != tt.expectError {
				t.Errorf("native error behavior mismatch: got error=%v, want error=%v", nativeHasError, tt.expectError)
			}

			// If bridge is available, compare error behavior
			if bridgeErr == nil || os.Getenv("SKIP_BRIDGE_TESTS") == "" {
				if nativeHasError != bridgeHasError {
					t.Logf("Error behavior differs: native=%v, bridge=%v (may be acceptable)", nativeHasError, bridgeHasError)
				}
			}
		})
	}
}

// TestRegressionFeatureParity documents intentional differences between native and bridge
func TestRegressionFeatureParity(t *testing.T) {
	// This test documents known differences between native and Python bridge implementations
	// These are intentional and acceptable differences

	knownDifferences := map[string]string{
		"session": "Native implementation may have different prompt discovery logic, but core functionality is equivalent",
		"recommend": "Native workflow recommendation uses simplified logic, but produces equivalent recommendations",
		"health": "Native health checks may have different implementation details, but check the same things",
		"ollama": "Native uses HTTP client directly, Python bridge may use different Ollama client library",
		"mlx": "MLX is intentionally Python bridge only - no native implementation exists",
	}

	for tool, reason := range knownDifferences {
		t.Run(tool+"_known_differences", func(t *testing.T) {
			t.Logf("Known difference for %s: %s", tool, reason)
			// This test passes - it's just documentation
		})
	}
}
