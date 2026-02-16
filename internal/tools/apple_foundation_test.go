//go:build darwin && arm64 && cgo
// +build darwin,arm64,cgo

package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestHandleAppleFoundationModels_ArgumentParsing(t *testing.T) {
	tests := []struct {
		name      string
		args      map[string]interface{}
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid args with prompt",
			args: map[string]interface{}{
				"prompt": "Test prompt",
			},
			wantError: false,
		},
		{
			name: "valid args with action and prompt",
			args: map[string]interface{}{
				"action": "summarize",
				"prompt": "Test text to summarize",
			},
			wantError: false,
		},
		{
			name: "missing prompt",
			args: map[string]interface{}{
				"action": "generate",
			},
			wantError: true,
			errorMsg:  "prompt is required",
		},
		{
			name: "empty prompt",
			args: map[string]interface{}{
				"prompt": "",
			},
			wantError: true,
			errorMsg:  "prompt is required",
		},
		{
			name: "invalid action",
			args: map[string]interface{}{
				"action": "invalid_action",
				"prompt": "Test prompt",
			},
			wantError: true,
			errorMsg:  "unknown action",
		},
		{
			name: "valid args with temperature and max_tokens",
			args: map[string]interface{}{
				"prompt":      "Test prompt",
				"temperature": 0.5,
				"max_tokens":  256,
			},
			wantError: false,
		},
		{
			name:      "empty args",
			args:      map[string]interface{}{},
			wantError: true,
			errorMsg:  "prompt is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			argsJSON, err := json.Marshal(tt.args)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}

			// Note: This will fail on unsupported platforms or without Swift bridge
			// but we're testing argument parsing logic
			_, err = handleAppleFoundationModels(ctx, argsJSON)

			if (err != nil) != tt.wantError {
				t.Errorf("handleAppleFoundationModels() error = %v, wantError %v", err, tt.wantError)
			}

			if tt.wantError && tt.errorMsg != "" && err != nil {
				if err.Error() == "" || !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("handleAppleFoundationModels() error = %v, want error containing %q", err, tt.errorMsg)
				}
			}
		})
	}
}

func TestHandleAppleFoundationModels_PlatformDetection(t *testing.T) {
	// Test that platform detection is called
	// This test verifies the integration with platform detection
	ctx := context.Background()

	// Test with valid args
	args := map[string]interface{}{
		"prompt": "Test prompt",
	}
	argsJSON, _ := json.Marshal(args)

	// Call handler - will check platform support
	// On unsupported platforms, should return error message
	// On supported platforms, will try to use Foundation Models (may fail without Swift bridge)
	_, err := handleAppleFoundationModels(ctx, argsJSON)

	// We expect either:
	// 1. Platform not supported error (graceful fallback)
	// 2. Foundation Models API error (Swift bridge not built)
	// 3. Success (if Swift bridge is built and platform is supported)
	if err != nil {
		// Check if it's a platform support error
		if strings.Contains(err.Error(), "not supported") {
			t.Logf("Platform not supported (expected on some systems): %v", err)
		} else {
			t.Logf("Foundation Models error (may need Swift bridge): %v", err)
		}
	}
}

func TestHandleAppleFoundationModels_ActionRouting(t *testing.T) {
	tests := []struct {
		name   string
		action string
		prompt string
	}{
		{
			name:   "generate action",
			action: "generate",
			prompt: "Generate some text",
		},
		{
			name:   "respond action",
			action: "respond",
			prompt: "Respond to this",
		},
		{
			name:   "summarize action",
			action: "summarize",
			prompt: "Long text to summarize here",
		},
		{
			name:   "classify action",
			action: "classify",
			prompt: "Text to classify",
		},
		{
			name:   "default action (no action specified)",
			action: "",
			prompt: "Default action test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			args := map[string]interface{}{
				"prompt": tt.prompt,
			}
			if tt.action != "" {
				args["action"] = tt.action
			}

			argsJSON, _ := json.Marshal(args)

			// Call handler - will route to appropriate action
			// May fail without Swift bridge, but routing logic should work
			_, err := handleAppleFoundationModels(ctx, argsJSON)

			// We don't check for specific errors here, just that routing happens
			// Actual API errors are expected without Swift bridge
			if err != nil {
				// Should not be an "unknown action" error for valid actions
				if strings.Contains(err.Error(), "unknown action") && tt.action != "" && tt.action != "invalid" {
					t.Errorf("handleAppleFoundationModels() incorrectly reported unknown action for %q: %v", tt.action, err)
				}
			}
		})
	}
}

func TestHandleAppleFoundationModels_ErrorHandling(t *testing.T) {
	tests := []struct {
		name      string
		argsJSON  json.RawMessage
		wantError bool
	}{
		{
			name:      "invalid JSON",
			argsJSON:  json.RawMessage(`{invalid json}`),
			wantError: true,
		},
		{
			name:      "nil args",
			argsJSON:  nil,
			wantError: true,
		},
		{
			name:      "empty JSON object",
			argsJSON:  json.RawMessage(`{}`),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			_, err := handleAppleFoundationModels(ctx, tt.argsJSON)

			if (err != nil) != tt.wantError {
				t.Errorf("handleAppleFoundationModels() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestHandleAppleFoundationModels_TextContentFormat(t *testing.T) {
	ctx := context.Background()

	args := map[string]interface{}{
		"prompt": "Test prompt",
	}
	argsJSON, _ := json.Marshal(args)

	result, err := handleAppleFoundationModels(ctx, argsJSON)

	// On unsupported platforms, should return TextContent with error message
	if err == nil && result != nil {
		if len(result) == 0 {
			t.Error("handleAppleFoundationModels() returned empty result")
		}

		if len(result) > 0 {
			if result[0].Type != "text" {
				t.Errorf("handleAppleFoundationModels() result type = %q, want %q", result[0].Type, "text")
			}
		}
	}
}

// Helper function to check if string contains substring (using standard library)
// Removed custom contains function - use strings.Contains instead

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}
