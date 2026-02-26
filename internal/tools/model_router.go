// Package tools: model router component for selecting and calling LLM backends.
// Tag hints for Todo2: #feature #refactor
//
// ModelRouter selects the best available backend (FM, Ollama, MLX) for a task type
// and dispatches Generate to that backend. Aligns with docs/MODEL_ASSISTED_WORKFLOW.md
// and docs/DEVWISDOM_GO_LESSONS.md (Model Router pattern).

package tools

import (
	"context"

	"github.com/davidl71/exarp-go/internal/config"
)

// ModelType identifies which backend/model to use (FM chain, Ollama variant, or MLX).
type ModelType string

const (
	// ModelFM uses the FM chain (Apple → Ollama → stub) via DefaultFMProvider().
	ModelFM ModelType = "fm"
	// ModelOllamaLlama uses Ollama with a general model (e.g. llama3.2).
	ModelOllamaLlama ModelType = "ollama-llama"
	// ModelOllamaCode uses Ollama with a code model (e.g. codellama).
	ModelOllamaCode ModelType = "ollama-codellama"
	// ModelMLX uses MLX (bridge) when available.
	ModelMLX ModelType = "mlx"
)

// ModelRequirements holds optional preferences for model selection.
type ModelRequirements struct {
	// PreferSpeed prefers faster models when set.
	PreferSpeed bool
	// PreferCost prefers free/low-cost backends when set.
	PreferCost bool
}

// ModelRouter selects the best model for a task and runs generation.
// Implementations should use LLMBackendStatus()/FMAvailable() to respect availability.
type ModelRouter interface {
	// SelectModel returns the best ModelType for the given task type and requirements.
	// taskType hints "code" vs "general"; empty is treated as general.
	SelectModel(taskType string, requirements ModelRequirements) ModelType
	// Generate runs the given model type with the prompt and returns generated text.
	Generate(ctx context.Context, model ModelType, prompt string, maxTokens int, temperature float32) (string, error)
}

// defaultModelRouter implements ModelRouter using existing backends (FM, Ollama, MLX).
type defaultModelRouter struct{}

// DefaultModelRouter is the shared router instance.
var DefaultModelRouter ModelRouter = &defaultModelRouter{}

// SelectModel picks the best available backend for task type and requirements.
// Code tasks prefer Ollama CodeLlama or MLX; general tasks prefer FM chain or Ollama Llama.
// Availability: FMAvailable(), MLAvailable() (darwin/arm64); MLX preferred for code on Apple Silicon.
func (r *defaultModelRouter) SelectModel(taskType string, requirements ModelRequirements) ModelType {
	isCode := taskType == "code" || taskType == "code_analysis" || taskType == "code_generation"

	if FMAvailable() {
		// FM chain (Apple or Ollama) is available; use it for both code and general.
		return ModelFM
	}
	// Prefer MLX for code tasks on Apple Silicon (local, no ollama serve needed).
	if isCode && MLAvailable() {
		return ModelMLX
	}

	if isCode {
		return ModelOllamaCode
	}
	// General: prefer MLX on Apple Silicon when PreferCost (local/free).
	if MLAvailable() && requirements.PreferCost {
		return ModelMLX
	}

	return ModelOllamaLlama
}

// Generate dispatches to the backend for the given ModelType.
func (r *defaultModelRouter) Generate(ctx context.Context, model ModelType, prompt string, maxTokens int, temperature float32) (string, error) {
	switch model {
	case ModelFM:
		p := DefaultFMProvider()
		if p == nil || !p.Supported() {
			return "", ErrFMNotSupported
		}

		return p.Generate(ctx, prompt, maxTokens, temperature)
	case ModelOllamaLlama:
		return r.generateOllama(ctx, config.GetOllamaDefaultModel(), prompt, maxTokens, temperature)
	case ModelOllamaCode:
		return r.generateOllama(ctx, config.GetOllamaCodeModel(), prompt, maxTokens, temperature)
	case ModelMLX:
		mlx := DefaultMLXProvider()
		if mlx == nil || !mlx.Supported() {
			return "", ErrFMNotSupported
		}

		return mlx.Generate(ctx, prompt, maxTokens, temperature)
	default:
		// Unknown type: try FM chain as fallback.
		p := DefaultFMProvider()
		if p != nil && p.Supported() {
			return p.Generate(ctx, prompt, maxTokens, temperature)
		}

		return "", ErrFMNotSupported
	}
}

func (r *defaultModelRouter) generateOllama(ctx context.Context, modelName, prompt string, maxTokens int, temperature float32) (string, error) {
	host := "http://localhost:11434"
	return ollamaGenerateText(ctx, prompt, maxTokens, temperature, host, modelName)
}

// ResolveModelForTask uses the recommend catalog (findBestModel) to pick a model for the given
// task description and type, then maps to our local ModelType. Implements T-207 model selection logic.
func ResolveModelForTask(taskDescription, taskType, optimizeFor string) (ModelType, ModelRequirements) {
	recommended := findBestModel(taskDescription, taskType, optimizeFor)

	req := ModelRequirements{}

	switch optimizeFor {
	case "speed":
		req.PreferSpeed = true
	case "cost":
		req.PreferCost = true
	}
	// Map ModelInfo.ModelID to ModelType (local backends only)
	switch recommended.ModelID {
	case "ollama-codellama":
		return ModelOllamaCode, req
	case "ollama-mistral", "ollama-phi3":
		return ModelOllamaLlama, req // Fast local models
	case "ollama-llama3.2":
		return ModelOllamaLlama, req
	default:
		// Cloud models (claude-*, gpt-4o, o1-*, gemini-*) → FM chain
		return ModelFM, req
	}
}
