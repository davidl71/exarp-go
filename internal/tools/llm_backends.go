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
		"fm_available":     FMAvailable(),
		"mlx_available":    MLAvailable(),
		"localai_available": LocalAIAvailable(),
		"ollama_tool":      "ollama",
		"mlx_tool":         "mlx",
		"apple_fm_tool":    "apple_foundation_models",
		"localai_tool":     "text_generate",
		"hint":             "Use apple_foundation_models for on-device FM; ollama for Ollama (native then bridge); mlx for MLX; text_generate provider=localai for LocalAI (OpenAI-compatible). text_generate for unified generate-text (provider=fm|insight|mlx|localai). FM chain (Apple → Ollama → stub). ReportInsight: FM then MLX (via shared path).",
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
