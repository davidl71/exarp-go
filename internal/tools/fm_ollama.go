package tools

import (
	"context"

	"github.com/davidl71/exarp-go/internal/config"
)

// ollamaTextGenerator implements TextGenerator using native Ollama generate (HTTP).
// Used in the FM chain (Apple → Ollama → stub) so DefaultFMProvider() can use Ollama
// when Apple FM is unavailable (e.g. Linux). Also used by text_generate (provider=ollama).
type ollamaTextGenerator struct{}

// DefaultOllamaGen is the shared Ollama TextGenerator for text_generate (provider=ollama).
var DefaultOllamaGen TextGenerator = &ollamaTextGenerator{}

// DefaultOllamaTextGenerator returns the Ollama TextGenerator (native HTTP, config-aware host/model).
func DefaultOllamaTextGenerator() TextGenerator {
	return DefaultOllamaGen
}

func (*ollamaTextGenerator) Supported() bool {
	// Always "try"; Generate will fail if Ollama is down. Chain falls through to stub.
	return true
}

func (o *ollamaTextGenerator) Generate(ctx context.Context, prompt string, maxTokens int, temperature float32) (string, error) {
	host := "http://localhost:11434"

	if cfg := config.GetGlobalConfig(); cfg != nil && cfg.Tools.Ollama.DefaultHost != "" {
		host = cfg.Tools.Ollama.DefaultHost
	}

	model := config.GetOllamaDefaultModel()
	if model == "" {
		model = "llama3.2"
	}

	return ollamaGenerateText(ctx, prompt, maxTokens, temperature, host, model)
}
