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
	"runtime"
	"sync"

	llama "github.com/go-skynet/go-llama.cpp"
)

// llamacppProvider implements TextGenerator for GGUF models via go-skynet/go-llama.cpp.
type llamacppProvider struct {
	mu        sync.RWMutex
	modelPath string
	supported bool
	model     *llama.LLama
}

func (p *llamacppProvider) Supported() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.supported && p.model != nil
}

func (p *llamacppProvider) Generate(ctx context.Context, prompt string, maxTokens int, temperature float32) (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.supported || p.model == nil {
		return "", fmt.Errorf("llamacpp provider not available: no model loaded")
	}

	output, err := p.model.Predict(
		prompt,
		llama.SetTokens(maxTokens),
		llama.SetTemperature(temperature),
		llama.SetThreads(runtime.NumCPU()),
	)
	if err != nil {
		return "", fmt.Errorf("llamacpp predict: %w", err)
	}
	return output, nil
}

// Load loads (or reloads) a GGUF model from the given path, replacing any existing one.
func (p *llamacppProvider) Load(modelPath string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.model != nil {
		p.model.Free()
		p.model = nil
	}

	gpu := DetectGPU()
	gpuLayers := 0
	if gpu.Available {
		gpuLayers = gpu.Layers
	}

	m, err := llama.New(
		modelPath,
		llama.SetContext(2048),
		llama.EnableF16Memory,
		llama.SetGPULayers(gpuLayers),
	)
	if err != nil {
		return fmt.Errorf("llamacpp load %s: %w", modelPath, err)
	}
	p.model = m
	p.modelPath = modelPath
	p.supported = true
	return nil
}

// Unload frees the loaded model.
func (p *llamacppProvider) Unload() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.model != nil {
		p.model.Free()
		p.model = nil
	}
	p.supported = false
}

// ModelPath returns the currently loaded model path.
func (p *llamacppProvider) ModelPath() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.modelPath
}

// defaultLlamaCppProvider is the singleton provider instance.
var defaultLlamaCppProvider *llamacppProvider

func init() {
	p := &llamacppProvider{}
	if modelPath := os.Getenv("LLAMACPP_MODEL_PATH"); modelPath != "" {
		if err := p.Load(modelPath); err != nil {
			// Non-fatal: status will report unavailable.
			_ = err
		}
	}
	defaultLlamaCppProvider = p
}

// DefaultLlamaCppProvider returns the llamacpp TextGenerator (non-nil when built with llamacpp tag).
func DefaultLlamaCppProvider() TextGenerator {
	return defaultLlamaCppProvider
}

// llamacppIfAvailable returns the llamacpp TextGenerator for the FM chain, or nil if not available.
func llamacppIfAvailable() TextGenerator {
	if defaultLlamaCppProvider != nil && defaultLlamaCppProvider.Supported() {
		return defaultLlamaCppProvider
	}
	return nil
}

// LlamaCppLoad loads a GGUF model into the singleton provider by path.
func LlamaCppLoad(modelPath string) error {
	if defaultLlamaCppProvider == nil {
		return fmt.Errorf("llamacpp provider not initialised")
	}
	return defaultLlamaCppProvider.Load(modelPath)
}

// LlamaCppUnload frees the loaded model.
func LlamaCppUnload() {
	if defaultLlamaCppProvider != nil {
		defaultLlamaCppProvider.Unload()
	}
}

// LlamaCppModelPath returns the currently loaded model path.
func LlamaCppModelPath() string {
	if defaultLlamaCppProvider == nil {
		return ""
	}
	return defaultLlamaCppProvider.ModelPath()
}
