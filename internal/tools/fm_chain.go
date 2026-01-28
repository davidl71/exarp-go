package tools

import "context"

// chainFMProvider implements FMProvider by trying a sequence of TextGenerators (Apple → Ollama → stub).
// Set as DefaultFM in init so DefaultFMProvider() uses the chain.
type chainFMProvider struct {
	backends []TextGenerator
}

func (c *chainFMProvider) Supported() bool {
	for _, b := range c.backends {
		if b != nil && b.Supported() {
			return true
		}
	}
	return true // we always "support" by trying; Generate may still fail
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
	backends := []TextGenerator{}
	if g := appleFMIfAvailable(); g != nil {
		backends = append(backends, g)
	}
	backends = append(backends, ollamaTG, stub)
	DefaultFM = &chainFMProvider{backends: backends}
}
