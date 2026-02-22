// Package tools: LLM backend status for discovery (stdio://models, tool hints).
// Exposes which LLM backends are available so clients can choose the right tool.

package tools

import (
	"os"
	"runtime"
	"strings"
)

// LLMBackendStatus returns a map describing available LLM backends for discovery.
// Used by stdio://models and by clients that need to know what is available
// (FM, Ollama, MLX, LocalAI) without calling each tool.
func LLMBackendStatus() map[string]interface{} {
	return map[string]interface{}{
		"fm_available":       FMAvailable(),
		"mlx_available":      MLAvailable(),
		"localai_available":  LocalAIAvailable(),
		"llamacpp_available": LlamaCppAvailable(),
		"ollama_tool":        "ollama",
		"mlx_tool":           "mlx",
		"apple_fm_tool":      "apple_foundation_models",
		"localai_tool":       "text_generate",
		"llamacpp_tool":      "llamacpp",
		"hint":               "text_generate is the unified generate-text dispatcher (provider=fm|ollama|mlx|localai|llamacpp|insight|auto). Use provider=auto for model selection. Separate tools (apple_foundation_models, ollama, mlx, llamacpp) offer rich actions (status, models, pull, hardware, docs, quality) beyond generation.",
	}
}

// LocalAIAvailable reports whether LocalAI is configured (LOCALAI_BASE_URL set).
// Does not verify the server is reachable—Generate may still fail.
func LocalAIAvailable() bool {
	return strings.TrimSpace(os.Getenv("LOCALAI_BASE_URL")) != ""
}

// MLAvailable reports whether MLX is potentially available.
// MLX (bridge) runs on Apple Silicon; returns true only for darwin/arm64.
// Does not verify MLX is actually installed—Generate may still fail.
func MLAvailable() bool {
	return runtime.GOOS == "darwin" && runtime.GOARCH == "arm64"
}

// LlamaCppAvailable reports whether llama.cpp is compiled in.
// True only when built with the llamacpp build tag and CGO enabled.
func LlamaCppAvailable() bool {
	return DefaultLlamaCppProvider() != nil
}
