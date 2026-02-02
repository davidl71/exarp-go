// Package tools: LLM backend status for discovery (stdio://models, tool hints).
// Exposes which LLM backends are available so clients can choose the right tool.

package tools

import "runtime"

// LLMBackendStatus returns a map describing available LLM backends for discovery.
// Used by stdio://models and by clients that need to know what is available
// (FM, Ollama, MLX) without calling each tool.
func LLMBackendStatus() map[string]interface{} {
	return map[string]interface{}{
		"fm_available":  FMAvailable(),
		"mlx_available": MLAvailable(),
		"ollama_tool":   "ollama",
		"mlx_tool":      "mlx",
		"apple_fm_tool": "apple_foundation_models",
		"hint":          "Use apple_foundation_models for on-device FM; ollama for Ollama (native then bridge); mlx for MLX; text_generate for unified generate-text (provider=fm|insight|mlx). FM chain (Apple → Ollama → stub). ReportInsight: FM then MLX (via shared path).",
	}
}

// MLAvailable reports whether MLX is potentially available.
// MLX (bridge) runs on Apple Silicon; returns true only for darwin/arm64.
// Does not verify MLX is actually installed—Generate may still fail.
func MLAvailable() bool {
	return runtime.GOOS == "darwin" && runtime.GOARCH == "arm64"
}
