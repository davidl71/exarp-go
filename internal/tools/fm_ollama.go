package tools

import (
	"context"

	"github.com/davidl71/exarp-go/internal/config"
)

// ollamaTextGenerator implements TextGenerator using native Ollama generate (HTTP).
// Used in the FM chain (Apple → Ollama → stub) so DefaultFMProvider() can use Ollama
// when Apple FM is unavailable (e.g. Linux).
type ollamaTextGenerator struct{}

func (*ollamaTextGenerator) Supported() bool {
	// Always "try"; Generate will fail if Ollama is down. Chain falls through to stub.
	return true
}

func (o *ollamaTextGenerator) Generate(ctx context.Context, prompt string, maxTokens int, temperature float32) (string, error) {
	host := "http://localhost:11434"
	model := "llama3.2"

	if cfg := config.GetGlobalConfig(); cfg != nil && cfg.Tools.Ollama.DefaultHost != "" {
		host = cfg.Tools.Ollama.DefaultHost
	}

	if cfg := config.GetGlobalConfig(); cfg != nil && cfg.Tools.Ollama.DefaultModel != "" {
		model = cfg.Tools.Ollama.DefaultModel
	}

	return ollamaGenerateText(ctx, prompt, maxTokens, temperature, host, model)
}
