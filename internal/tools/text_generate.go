package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/davidl71/exarp-go/internal/framework"
)

// handleTextGenerate implements the unified text_generate tool.
// provider=fm uses DefaultFMProvider(); provider=insight uses DefaultReportInsight(); provider=mlx uses DefaultMLXProvider().
// All implement TextGenerator; this tool is a thin wrapper for generate-text use cases.
func handleTextGenerate(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	provider := "fm"
	if p, ok := params["provider"].(string); ok && p != "" {
		provider = p
	}

	prompt, _ := params["prompt"].(string)
	if prompt == "" {
		return nil, fmt.Errorf("prompt is required for text_generate")
	}

	maxTokens := getMaxTokens(params)
	temperature := getTemperature(params)

	var gen TextGenerator
	switch provider {
	case "fm":
		gen = DefaultFMProvider()
	case "insight":
		gen = DefaultReportInsight()
	case "mlx":
		gen = DefaultMLXProvider()
	default:
		return nil, fmt.Errorf("unknown provider: %q (use \"fm\", \"insight\", or \"mlx\")", provider)
	}

	if gen == nil || !gen.Supported() {
		return nil, fmt.Errorf("provider %q is not available", provider)
	}

	text, err := gen.Generate(ctx, prompt, maxTokens, temperature)
	if err != nil {
		return nil, fmt.Errorf("text_generate failed: %w", err)
	}

	// Return generated text; optionally wrap in JSON if we add structured output later
	return []framework.TextContent{
		{Type: "text", Text: text},
	}, nil
}
