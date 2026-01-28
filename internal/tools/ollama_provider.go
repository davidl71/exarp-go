// Package tools: Ollama provider abstraction for native-first, bridge-fallback.
// Handlers use DefaultOllama so they do not reference the bridge directly for ollama.

package tools

import (
	"context"

	"github.com/davidl71/exarp-go/internal/bridge"
	"github.com/davidl71/exarp-go/internal/framework"
)

// OllamaProvider invokes the ollama tool (native first, then bridge fallback).
type OllamaProvider interface {
	Invoke(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error)
}

// defaultOllama is the shared default; set by init to composite (native then bridge).
var defaultOllama OllamaProvider

func init() {
	defaultOllama = &compositeOllama{}
}

// DefaultOllama returns the default ollama provider (tries native, then bridge).
func DefaultOllama() OllamaProvider {
	if defaultOllama == nil {
		return &compositeOllama{}
	}
	return defaultOllama
}

// compositeOllama tries native first, then bridge on error.
type compositeOllama struct{}

func (c *compositeOllama) Invoke(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	result, err := handleOllamaNative(ctx, params)
	if err == nil {
		return result, nil
	}
	return invokeOllamaViaBridge(ctx, params)
}

// invokeOllamaViaBridge calls the ollama tool via the Python bridge.
func invokeOllamaViaBridge(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	resultStr, err := bridge.ExecutePythonTool(ctx, "ollama", params)
	if err != nil {
		return nil, err
	}
	return []framework.TextContent{
		{Type: "text", Text: resultStr},
	}, nil
}
