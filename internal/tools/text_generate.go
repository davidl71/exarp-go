package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/davidl71/exarp-go/internal/framework"
)

// handleTextGenerate implements the unified text_generate tool.
// provider=fm uses DefaultFMProvider(); provider=insight uses DefaultReportInsight(); provider=mlx uses DefaultMLXProvider().
// provider=auto (or task_type/task_description provided) uses ResolveModelForTask + ModelRouter for model selection (T-207).
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

	taskType, _ := params["task_type"].(string)
	taskDesc, _ := params["task_description"].(string)

	optimizeFor := "quality"
	if o, ok := params["optimize_for"].(string); ok && o != "" {
		optimizeFor = o
	}

	maxTokens := getMaxTokens(params)
	temperature := getTemperature(params)

	// Model selection (T-207): when provider=auto or task hints provided, use recommend + router
	useModelSelection := provider == "auto" || taskType != "" || taskDesc != ""

	if useModelSelection {
		modelType, _ := ResolveModelForTask(taskDesc, taskType, optimizeFor)

		text, err := DefaultModelRouter.Generate(ctx, modelType, prompt, maxTokens, temperature)
		if err != nil {
			return nil, fmt.Errorf("text_generate (model selection) failed: %w", err)
		}

		return []framework.TextContent{{Type: "text", Text: text}}, nil
	}

	var gen TextGenerator

	switch provider {
	case "fm":
		gen = DefaultFMProvider()
	case "insight":
		gen = DefaultReportInsight()
	case "mlx":
		gen = DefaultMLXProvider()
	default:
		return nil, fmt.Errorf("unknown provider: %q (use \"fm\", \"insight\", \"mlx\", or \"auto\")", provider)
	}

	if gen == nil || !gen.Supported() {
		return nil, fmt.Errorf("provider %q is not available", provider)
	}

	text, err := gen.Generate(ctx, prompt, maxTokens, temperature)
	if err != nil {
		return nil, fmt.Errorf("text_generate failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: text},
	}, nil
}
