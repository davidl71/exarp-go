// text_generate.go — MCP "text_generate" tool: model-routed text generation (auto/fm/ollama/mlx).
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/davidl71/exarp-go/internal/framework"
)

// handleTextGenerate implements the unified text_generate tool.
// provider=fm uses DefaultFMProvider(); provider=ollama uses DefaultOllamaTextGenerator(); provider=insight uses DefaultReportInsight();
// provider=mlx uses DefaultMLXProvider(); provider=localai uses DefaultLocalAIProvider().
// provider=auto (or task_type/task_description provided) uses ResolveModelForTask + ModelRouter for model selection (T-207).
func handleTextGenerate(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	provider := "fm"
	if p := ParamString(params, "provider"); p != "" {
		provider = p
	}

	prompt := ParamString(params, "prompt")
	if prompt == "" {
		return nil, fmt.Errorf("prompt is required for text_generate")
	}

	taskType := ParamString(params, "task_type")
	taskDesc := ParamString(params, "task_description")

	optimizeFor := "quality"
	if o := ParamString(params, "optimize_for"); o != "" {
		optimizeFor = o
	}

	maxTokens := getMaxTokens(params)
	temperature := getTemperature(params)

	modelOverride := ParamString(params, "model")

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
	case "ollama":
		gen = DefaultOllamaTextGenerator()
	case "insight":
		gen = DefaultReportInsight()
	case "mlx":
		gen = DefaultMLXProvider()
	case "localai":
		gen = DefaultLocalAIProvider()
	case "gateway":
		gen = DefaultGatewayProvider()
	case "llamacpp":
		gen = DefaultLlamaCppProvider()
	default:
		return nil, fmt.Errorf("unknown provider: %q (use \"fm\", \"ollama\", \"insight\", \"mlx\", \"localai\", \"gateway\", \"llamacpp\", or \"auto\")", provider)
	}

	if gen == nil || !gen.Supported() {
		return nil, fmt.Errorf("provider %q is not available", provider)
	}

	var text string
	var err error
	if modelOverride != "" && (provider == "localai" || provider == "gateway") {
		if withModel, ok := gen.(TextGeneratorWithModel); ok {
			text, err = withModel.GenerateWithModel(ctx, prompt, maxTokens, temperature, modelOverride)
		} else {
			text, err = gen.Generate(ctx, prompt, maxTokens, temperature)
		}
	} else {
		text, err = gen.Generate(ctx, prompt, maxTokens, temperature)
	}
	if err != nil {
		return nil, fmt.Errorf("text_generate failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: text},
	}, nil
}
