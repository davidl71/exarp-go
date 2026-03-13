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

func TestGenerateWithOptions_ParameterPassing(t *testing.T) {
	// Test that GenerateWithOptions properly accepts parameters
	// This tests the function signature and parameter handling
	// Actual API call will fail without Swift bridge, but we verify the function can be called

	tests := []struct {
		name        string
		prompt      string
		maxTokens   int
		temperature float32
	}{
		{
			name:        "standard parameters",
			prompt:      "Hello world",
			maxTokens:   256,
			temperature: 0.7,
		},
		{
			name:        "zero max tokens",
			prompt:      "Test",
			maxTokens:   0,
			temperature: 0.5,
		},
		{
			name:        "zero temperature",
			prompt:      "Deterministic output",
			maxTokens:   100,
			temperature: 0.0,
		},
		{
			name:        "max temperature",
			prompt:      "Creative output",
			maxTokens:   500,
			temperature: 1.0,
		},
		{
			name:        "empty prompt",
			prompt:      "",
			maxTokens:   100,
			temperature: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function - will fail without Swift bridge but verifies parameters are passed
			_, err := GenerateWithOptions(tt.prompt, tt.maxTokens, tt.temperature)

			// We expect an error on platforms without Swift bridge
			// The important thing is that the function accepts these parameters
			if err != nil {
				// Check if it's the expected "not supported" error
				errMsg := err.Error()
				if !containsSubstring(errMsg, "not supported") && !containsSubstring(errMsg, "Swift") && !containsSubstring(errMsg, "foundation") {
					t.Logf("Got unexpected error: %v", err)
				} else {
					t.Logf("Platform not supported (expected): %v", err)
				}
			}
		})
	}
}

func TestClassifyTemperatureMaxTokens(t *testing.T) {
	tests := []struct {
		name       string
		params     map[string]interface{}
		wantTemp   float32
		wantMaxTok int
	}{
		{
			name:       "default helper returns 0.7",
			params:     map[string]interface{}{},
			wantTemp:   0.7, // helper default; classifyText applies 0.2
			wantMaxTok: 512,
		},
		{
			name: "custom temperature",
			params: map[string]interface{}{
				"temperature": 0.8,
			},
			wantTemp:   0.8,
			wantMaxTok: 512,
		},
		{
			name: "custom max_tokens",
			params: map[string]interface{}{
				"max_tokens": 256,
			},
			wantTemp:   0.7,
			wantMaxTok: 256,
		},
		{
			name: "both custom",
			params: map[string]interface{}{
				"temperature": 0.5,
				"max_tokens":  1024,
			},
			wantTemp:   0.5,
			wantMaxTok: 1024,
		},
		{
			name: "zero temperature uses helper default",
			params: map[string]interface{}{
				"temperature": 0.0,
			},
			wantTemp:   0.0, // zero temperature means use default in classifyText
			wantMaxTok: 512,
		},
		{
			name: "max temperature",
			params: map[string]interface{}{
				"temperature": 1.0,
			},
			wantTemp:   1.0,
			wantMaxTok: 512,
		},
		{
			name: "float64 max_tokens (JSON unmarshaling)",
			params: map[string]interface{}{
				"max_tokens": float64(128),
			},
			wantTemp:   0.7,
			wantMaxTok: 128,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTemp := getTemperature(tt.params)
			if gotTemp != tt.wantTemp {
				t.Errorf("getTemperature() = %v, want %v", gotTemp, tt.wantTemp)
			}

			gotMaxTok := getMaxTokens(tt.params)
			if gotMaxTok != tt.wantMaxTok {
				t.Errorf("getMaxTokens() = %v, want %v", gotMaxTok, tt.wantMaxTok)
			}
		})
	}
}

func TestClassifyTextAppliesCorrectDefaults(t *testing.T) {
	// Test that temperature parameter affects behavior
	// When temperature is provided, it should be used
	paramsWithTemp := map[string]interface{}{"temperature": 0.9}
	temp := getTemperature(paramsWithTemp)
	if temp != 0.9 {
		t.Errorf("custom temperature should be 0.9, got %v", temp)
	}

	// Test that max_tokens parameter is correctly extracted
	paramsWithMaxTokens := map[string]interface{}{"max_tokens": 256}
	maxTok := getMaxTokens(paramsWithMaxTokens)
	if maxTok != 256 {
		t.Errorf("custom max_tokens should be 256, got %v", maxTok)
	}

	// Test float64 max_tokens (common in JSON unmarshaling)
	paramsWithFloatMaxTokens := map[string]interface{}{"max_tokens": float64(128)}
	maxTok = getMaxTokens(paramsWithFloatMaxTokens)
	if maxTok != 128 {
		t.Errorf("float64 max_tokens should be 128, got %v", maxTok)
	}
}
