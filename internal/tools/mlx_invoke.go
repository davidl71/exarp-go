// Package tools: shared MLX invoke path (native only). Python bridge removed.
// Used by the mlx tool handler, insight_provider, and the MLX text_generate provider.

package tools

import (
	"context"
)

// InvokeMLXTool runs the MLX tool with the given params using native handleMlxNative only.
// For generate, returns an error (use ollama or apple_foundation_models). Returns the raw result string (JSON) when action is models/status/hardware.
func InvokeMLXTool(ctx context.Context, params map[string]interface{}) (string, error) {
	if params == nil {
		params = make(map[string]interface{})
	}

	if _, ok := params["model"]; !ok {
		params["model"] = "mlx-community/Phi-3.5-mini-instruct-4bit"
	}

	tc, err := handleMlxNative(ctx, params)
	if err != nil {
		return "", err
	}

	if len(tc) > 0 && tc[0].Text != "" {
		return tc[0].Text, nil
	}

	return "", nil
}
