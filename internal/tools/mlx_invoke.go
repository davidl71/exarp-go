// Package tools: shared MLX invoke path (native then bridge) for generate and other callers.
// Used by the mlx tool handler, insight_provider, and the MLX text_generate provider.

package tools

import (
	"context"

	"github.com/davidl71/exarp-go/internal/bridge"
)

// InvokeMLXTool runs the MLX tool with the given params using the shared path:
// native (handleMlxNative) when available, then Python bridge for generate or when native fails.
// Returns the raw result string (JSON) so callers can parse generated_text or other fields.
func InvokeMLXTool(ctx context.Context, params map[string]interface{}) (string, error) {
	if params == nil {
		params = make(map[string]interface{})
	}
	if _, ok := params["model"]; !ok {
		params["model"] = "mlx-community/Phi-3.5-mini-instruct-4bit"
	}

	if MLXNativeAvailable() {
		tc, err := handleMlxNative(ctx, params)
		if err == nil {
			if len(tc) > 0 && tc[0].Text != "" {
				return tc[0].Text, nil
			}
		}
		action, _ := params["action"].(string)
		if action != "generate" {
			return "", err
		}
	}

	return bridge.ExecutePythonTool(ctx, "mlx", params)
}
