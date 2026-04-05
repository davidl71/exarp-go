// Package tools: model router component for selecting and calling LLM backends.
// ModelRouter selects the best available backend (FM, Ollama) for a task type
// and dispatches Generate to that backend.

package tools

import (
	"context"
	"fmt"

	"github.com/davidl71/exarp-go/internal/config"
)

// ModelType identifies which backend/model to use (FM chain, Ollama variant).
type ModelType string

const (
	// ModelFM uses the FM chain (Apple → Ollama → stub) via DefaultFMProvider().
	ModelFM ModelType = "fm"
	// ModelGateway uses the OpenAI-compatible gateway (OPENAI_GATEWAY_BASE_URL) when set; used by provider=auto.
	ModelGateway ModelType = "gateway"
	// ModelOllamaLlama uses Ollama with a general model (e.g. llama3.2).
	ModelOllamaLlama ModelType = "ollama-llama"
	// ModelOllamaCode uses Ollama with a code model (e.g. codellama).
	ModelOllamaCode ModelType = "ollama-codellama"
)

// ModelRequirements holds optional preferences for model selection.
type ModelRequirements struct {
	PreferSpeed bool
	PreferCost  bool
	AgentRole   string
}

// ModelRouter selects the best model for a task and runs generation.
type ModelRouter interface {
	SelectModel(taskType string, requirements ModelRequirements) ModelType
	Generate(ctx context.Context, model ModelType, prompt string, maxTokens int, temperature float32) (string, error)
}

// defaultModelRouter implements ModelRouter using FM, gateway, and Ollama.
type defaultModelRouter struct{}

// DefaultModelRouter is the shared router instance.
var DefaultModelRouter ModelRouter = &defaultModelRouter{}

// SelectModel picks the best available backend for task type and requirements.
func (r *defaultModelRouter) SelectModel(taskType string, requirements ModelRequirements) ModelType {
	if requirements.AgentRole != "" {
		switch requirements.AgentRole {
		case AgentRolePlanner, AgentRoleReviewer, AgentRoleResearcher:
			if FMAvailable() {
				return ModelFM
			}
		}
	}
	isCode := taskType == "code" || taskType == "code_analysis" || taskType == "code_generation"

	if FMAvailable() {
		return ModelFM
	}
	if GatewayAvailable() {
		return ModelGateway
	}

	if isCode {
		return ModelOllamaCode
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
	case ModelGateway:
		g := DefaultGatewayProvider()
		if g == nil || !g.Supported() {
			return "", fmt.Errorf("gateway not configured (set OPENAI_GATEWAY_BASE_URL)")
		}
		return g.Generate(ctx, prompt, maxTokens, temperature)
	case ModelOllamaLlama:
		return r.generateOllama(ctx, config.GetOllamaDefaultModel(), prompt, maxTokens, temperature)
	case ModelOllamaCode:
		return r.generateOllama(ctx, config.GetOllamaCodeModel(), prompt, maxTokens, temperature)
	default:
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
// task description and type, then maps to our local ModelType.
func ResolveModelForTask(taskDescription, taskType, optimizeFor, agentRole string) (ModelType, ModelRequirements) {
	recommended := findBestModel(taskDescription, taskType, optimizeFor)

	req := ModelRequirements{}
	req.AgentRole = agentRole

	switch optimizeFor {
	case "speed":
		req.PreferSpeed = true
	case "cost":
		req.PreferCost = true
	}
	switch recommended.ModelID {
	case "ollama-codellama":
		return ModelOllamaCode, req
	case "ollama-mistral", "ollama-phi3":
		return ModelOllamaLlama, req
	case "ollama-llama3.2":
		return ModelOllamaLlama, req
	default:
		if GatewayAvailable() {
			return ModelGateway, req
		}
		return ModelFM, req
	}
}
