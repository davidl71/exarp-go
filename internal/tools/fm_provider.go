// Package tools: FM provider abstraction for foundation model operations.
// DefaultFM is a chain (Apple → Ollama → stub) set in fm_chain.go init.

package tools

import (
	"context"
	"errors"
)

// ErrFMNotSupported is returned when a foundation model is requested but not available on this platform.
var ErrFMNotSupported = errors.New("foundation model not supported on this platform")

// TextGenerator is the shared contract for "generate text from prompt + options".
// Implemented by FMProvider and ReportInsightProvider; use when code only needs generate-text.
type TextGenerator interface {
	Supported() bool
	Generate(ctx context.Context, prompt string, maxTokens int, temperature float32) (string, error)
}

// TextGeneratorWithModel extends TextGenerator with optional model override for the request.
// Implemented by LocalAI and Gateway providers so text_generate can pass params["model"] (e.g. router-mf-0.11, ollama/phi4).
type TextGeneratorWithModel interface {
	TextGenerator
	GenerateWithModel(ctx context.Context, prompt string, maxTokens int, temperature float32, modelOverride string) (string, error)
}

// FMProvider abstracts foundation model access so tools (e.g. task_analysis hierarchy)
// can use Apple FM when available and fail cleanly otherwise, without Python fallback.
// FMProvider implements TextGenerator.
type FMProvider interface {
	TextGenerator
}

// DefaultFM is set by init() in fm_chain.go (stock: Ollama → stub; Apple may be prepended in other builds).
// Prefer DefaultFMProvider() for consistency with DefaultReportInsight() and DefaultOllama().
var DefaultFM FMProvider

// DefaultFMProvider returns the default FM provider (set in init; never nil).
// Use for consistency with DefaultReportInsight() and DefaultOllama().
func DefaultFMProvider() FMProvider {
	return DefaultFM
}

// FMAvailable reports whether the default FM provider likely has a working backend.
// For the stock chain (Ollama → stub), this follows a cached GET /api/tags probe to Ollama.
// Generate may still be called when a probe was stale; callers should handle Generate errors.
func FMAvailable() bool {
	p := DefaultFMProvider()
	return p != nil && p.Supported()
}
