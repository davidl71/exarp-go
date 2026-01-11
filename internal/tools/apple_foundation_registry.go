//go:build darwin && arm64 && cgo
// +build darwin,arm64,cgo

package tools

import (
	"fmt"

	"github.com/davidl71/exarp-go/internal/framework"
)

// registerAppleFoundationModelsTool registers the Apple Foundation Models tool
// This function is only compiled on supported platforms (darwin && arm64 && cgo)
func registerAppleFoundationModelsTool(server framework.MCPServer) error {
	if err := server.RegisterTool(
		"apple_foundation_models",
		"[HINT: Apple Foundation Models. action=generate|respond|summarize|classify. On-device AI on Apple Silicon.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"generate", "respond", "summarize", "classify"},
					"default": "generate",
				},
				"prompt": map[string]interface{}{
					"type": "string",
				},
				"mode": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"deterministic", "creative"},
					"default": "deterministic",
				},
				"temperature": map[string]interface{}{
					"type":    "number",
					"default": 0.7,
				},
				"max_tokens": map[string]interface{}{
					"type":    "integer",
					"default": 512,
				},
				"categories": map[string]interface{}{
					"type": "string",
				},
			},
		},
		handleAppleFoundationModels,
	); err != nil {
		return fmt.Errorf("failed to register apple_foundation_models: %w", err)
	}

	return nil
}
