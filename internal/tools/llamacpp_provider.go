//go:build llamacpp && cgo
// +build llamacpp,cgo

// llamacpp_provider.go — llamacpp TextGenerator provider using go-skynet/go-llama.cpp.
// Built only with CGO and the "llamacpp" build tag. Provides GGUF model loading
// and text generation via llama.cpp bindings.
// See docs/LLAMACPP_TOOL_SCHEMA.md for the tool schema and docs/LLAMACPP_EVALUATION.md
// for the library evaluation.
package tools

import (
	"context"
	"fmt"
	"os"
	"sync"
)

// llamacppProvider implements TextGenerator for GGUF models via go-skynet/go-llama.cpp.
type llamacppProvider struct {
	mu        sync.RWMutex
	modelPath string
	supported bool
}

func (p *llamacppProvider) Supported() bool {
	return p.supported
}

func (p *llamacppProvider) Generate(ctx context.Context, prompt string, maxTokens int, temperature float32) (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.supported {
		return "", fmt.Errorf("llamacpp provider not available: no model loaded")
	}

	if p.modelPath == "" {
		return "", fmt.Errorf("llamacpp: no model path configured")
	}

	// TODO: implement actual generation using go-skynet/go-llama.cpp bindings
	// This requires adding the dependency to go.mod and building libbinding.a.
	// See docs/LLAMACPP_TOOL_SCHEMA.md for the full API design.
	return "", fmt.Errorf("llamacpp: generation not yet implemented (model=%s)", p.modelPath)
}

// defaultLlamaCppProvider is the singleton provider instance.
var defaultLlamaCppProvider *llamacppProvider

func init() {
	modelPath := os.Getenv("LLAMACPP_MODEL_PATH")
	defaultLlamaCppProvider = &llamacppProvider{
		modelPath: modelPath,
		supported: modelPath != "",
	}
}

// DefaultLlamaCppProvider returns the llamacpp TextGenerator (non-nil when built with llamacpp tag).
func DefaultLlamaCppProvider() TextGenerator {
	if defaultLlamaCppProvider == nil {
		return nil
	}
	return defaultLlamaCppProvider
}

// llamacppIfAvailable returns the llamacpp TextGenerator for the FM chain, or nil if not available.
func llamacppIfAvailable() TextGenerator {
	p := DefaultLlamaCppProvider()
	if p != nil && p.Supported() {
		return p
	}
	return nil
}
