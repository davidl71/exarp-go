package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/davidl71/exarp-go/internal/tools"
)

// handleSessionStatus handles the stdio://session/status resource
// Returns current context (handoff, task, dashboard) for dynamic UI labels when Cursor adds support.
func handleSessionStatus(ctx context.Context, uri string) ([]byte, string, error) {
	projectRoot, err := tools.FindProjectRoot()
	if err != nil {
		return nil, "", fmt.Errorf("failed to find project root: %w", err)
	}

	label, statusCtx, details := tools.GetSessionStatus(projectRoot)

	result := map[string]interface{}{
		"label":     label,
		"context":   statusCtx,
		"details":   details,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal session status: %w", err)
	}

	return jsonData, "application/json", nil
}

// handleSessionMode handles the stdio://session/mode resource
// Returns current inferred session mode (AGENT/ASK/MANUAL) with confidence
// Uses native Go infer_session_mode implementation
func handleSessionMode(ctx context.Context, uri string) ([]byte, string, error) {
	// Use the native Go infer_session_mode implementation
	// We call the native function directly (same logic as the tool handler)
	mode, confidence, err := inferSessionMode(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to infer session mode: %w", err)
	}

	result := map[string]interface{}{
		"mode":       mode,
		"confidence": confidence,
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal session mode: %w", err)
	}

	return jsonData, "application/json", nil
}

// inferSessionMode infers the current session mode (AGENT/ASK/MANUAL) with confidence
// This uses the same logic as the infer_session_mode tool
func inferSessionMode(ctx context.Context) (string, float64, error) {
	// Use the native Go implementation from tools package
	// Call HandleInferSessionModeNative directly with empty params (no force_recompute)
	params := map[string]interface{}{}
	result, err := tools.HandleInferSessionModeNative(ctx, params)
	if err != nil {
		return "", 0.0, fmt.Errorf("session mode inference failed: %w", err)
	}

	// Parse result from TextContent (JSON string)
	if len(result) == 0 {
		return "", 0.0, fmt.Errorf("no result from session mode inference")
	}

	// Parse JSON result - result format matches ModeInferenceResult
	var parsedResult map[string]interface{}
	if err := json.Unmarshal([]byte(result[0].Text), &parsedResult); err != nil {
		return "", 0.0, fmt.Errorf("failed to parse session mode result: %w", err)
	}

	// Extract mode and confidence
	mode, ok := parsedResult["mode"].(string)
	if !ok {
		return "", 0.0, fmt.Errorf("invalid mode in result")
	}

	confidence, ok := parsedResult["confidence"].(float64)
	if !ok {
		confidence = 0.5 // Default confidence if not found
	}

	return mode, confidence, nil
}
