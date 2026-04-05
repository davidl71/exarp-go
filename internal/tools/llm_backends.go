// llm_backends.go — LLM backend discovery map for stdio://models and tool hints.
// fm_available follows DefaultFMProvider().Supported() (cached Ollama /api/tags probe on the stock chain).
// ollama_reachable is the explicit probe; Generate can still fail if Ollama stops after a successful probe.

package tools

import (
	"os"
	"strings"
)

// LLMBackendStatus returns a map describing available LLM backends for discovery.
// Used by stdio://models and by clients that need to know what is available
// (FM, Ollama, LocalAI, Gateway) without calling each tool.
func LLMBackendStatus() map[string]interface{} {
	return map[string]interface{}{
		"fm_available":      FMAvailable(),
		"ollama_reachable":  OllamaReachableForFM(),
		"localai_available": LocalAIAvailable(),
		"gateway_available": GatewayAvailable(),
		"ollama_tool":       "ollama",
		"localai_tool":      "text_generate",
		"gateway_tool":      "text_generate",
		"hint":              "text_generate is the unified generate-text dispatcher (provider=fm|ollama|localai|gateway|insight|auto). Use provider=auto for model selection. Use provider=gateway with OPENAI_GATEWAY_BASE_URL for any OpenAI-compatible router.",
	}
}

// LocalAIAvailable reports whether LocalAI is configured (LOCALAI_BASE_URL set).
// Does not verify the server is reachable—Generate may still fail.
func LocalAIAvailable() bool {
	return strings.TrimSpace(os.Getenv("LOCALAI_BASE_URL")) != ""
}

// GatewayAvailable reports whether the OpenAI-compatible gateway is configured (OPENAI_GATEWAY_BASE_URL set).
// Does not verify the server is reachable—Generate may still fail.
func GatewayAvailable() bool {
	return strings.TrimSpace(os.Getenv("OPENAI_GATEWAY_BASE_URL")) != ""
}

