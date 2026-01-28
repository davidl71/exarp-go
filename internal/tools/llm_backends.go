// Package tools: LLM backend status for discovery (stdio://models, tool hints).
// Exposes which LLM backends are available so clients can choose the right tool.

package tools

// LLMBackendStatus returns a map describing available LLM backends for discovery.
// Used by stdio://models and by clients that need to know what is available
// (FM, Ollama, MLX) without calling each tool.
func LLMBackendStatus() map[string]interface{} {
	return map[string]interface{}{
		"fm_available":  FMAvailable(),
		"ollama_tool":   "ollama",
		"mlx_tool":      "mlx",
		"apple_fm_tool": "apple_foundation_models",
		"hint":          "Use apple_foundation_models for on-device FM; ollama for Ollama (native then bridge); mlx for MLX. Report/scorecard insights use DefaultReportInsight (MLX then FM).",
	}
}
