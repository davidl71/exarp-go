// fm_chain.go — Default DefaultFM provider: tries TextGenerator backends in order (stock: Ollama then stub).
// Supported() is truthful for discovery (Ollama via OllamaReachableForFM); Generate still walks backends using each generator’s Supported().
package tools

import "context"

// chainFMProvider implements FMProvider by trying a sequence of TextGenerators (e.g. Ollama → stub).
// Set as DefaultFM in init so DefaultFMProvider() uses the chain.
type chainFMProvider struct {
	backends []TextGenerator
}

func (c *chainFMProvider) Supported() bool {
	if c == nil || len(c.backends) == 0 {
		return false
	}
	for _, b := range c.backends {
		if b == nil {
			continue
		}
		// Stub never counts as an available backend for discovery / FMAvailable.
		if _, stub := b.(*chainStubFMProvider); stub {
			continue
		}
		// Ollama: truthful reachability (cached GET /api/tags). Generate still tries when invoked.
		if _, ollama := b.(*ollamaTextGenerator); ollama {
			if OllamaReachableForFM() {
				return true
			}
			continue
		}
		if b.Supported() {
			return true
		}
	}
	return false
}

func (c *chainFMProvider) Generate(ctx context.Context, prompt string, maxTokens int, temperature float32) (string, error) {
	var lastErr error

	for _, b := range c.backends {
		if b == nil {
			continue
		}

		if !b.Supported() {
			continue
		}

		out, err := b.Generate(ctx, prompt, maxTokens, temperature)
		if err == nil && out != "" {
			return out, nil
		}

		lastErr = err
	}

	if lastErr != nil {
		return "", lastErr
	}

	return "", ErrFMNotSupported
}

// chainStubFMProvider is the fallback when Apple FM and Ollama are unavailable.
type chainStubFMProvider struct{}

func (*chainStubFMProvider) Supported() bool { return false }

func (*chainStubFMProvider) Generate(_ context.Context, _ string, _ int, _ float32) (string, error) {
	return "", ErrFMNotSupported
}

func init() {
	ollamaTG := &ollamaTextGenerator{}
	stub := &chainStubFMProvider{}

	backends := []TextGenerator{ollamaTG, stub}
	DefaultFM = &chainFMProvider{backends: backends}
}
