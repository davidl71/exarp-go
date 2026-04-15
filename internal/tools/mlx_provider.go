// Package tools: MLX as a TextGenerator provider for the LLM abstraction (text_generate, report insights).
// Uses InvokeMLXTool (native then bridge) so generate goes through the same path as the mlx tool.

package tools

import (
	"context"
)

// mlxProvider implements TextGenerator for MLX (Apple Silicon).
// Generate uses InvokeMLXTool so native generate is used when available, else bridge.
type mlxProvider struct{}

// DefaultMLX is the shared MLX provider for text_generate (provider=mlx) and the abstraction layer.
var DefaultMLX TextGenerator = &mlxProvider{}

// DefaultMLXProvider returns the default MLX provider (implements TextGenerator).
func DefaultMLXProvider() TextGenerator {
	return DefaultMLX
}

// Supported reports whether MLX generate is available (always true; invoke path uses native or bridge).
func (m *mlxProvider) Supported() bool {
	return true
}

// Generate runs MLX generate via the shared path (native then bridge) and returns the generated text.
func (m *mlxProvider) Generate(ctx context.Context, prompt string, maxTokens int, temperature float32) (string, error) {
	params := map[string]interface{}{
		"action":      "generate",
		"prompt":      prompt,
		"max_tokens":  maxTokens,
		"temperature": float64(temperature),
	}

	result, err := InvokeMLXTool(ctx, params)
	if err != nil {
		return "", err
	}

	return parseGeneratedTextFromMLXResponse(result)
}
