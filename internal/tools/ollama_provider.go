// Package tools: Ollama provider abstraction (native only; no bridge fallback).
// Handlers use DefaultOllama; native failure returns error.

package tools

import (
	"context"

	"github.com/davidl71/exarp-go/internal/framework"
)

// OllamaProvider invokes the ollama tool (native only).
type OllamaProvider interface {
	Invoke(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error)
}

// defaultOllama is the shared default; set by init to native-only.
var defaultOllama OllamaProvider

func init() {
	defaultOllama = &nativeOnlyOllama{}
}

// DefaultOllama returns the default ollama provider (native only; no bridge).
func DefaultOllama() OllamaProvider {
	if defaultOllama == nil {
		return &nativeOnlyOllama{}
	}

	return defaultOllama
}

// nativeOnlyOllama uses only the native Go implementation; returns error on failure.
type nativeOnlyOllama struct{}

func (n *nativeOnlyOllama) Invoke(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	return handleOllamaNative(ctx, params)
}
