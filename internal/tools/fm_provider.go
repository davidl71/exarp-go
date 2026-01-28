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

// FMProvider abstracts foundation model access so tools (e.g. task_analysis hierarchy)
// can use Apple FM when available and fail cleanly otherwise, without Python fallback.
// FMProvider implements TextGenerator.
type FMProvider interface {
	TextGenerator
}

// DefaultFM is set by init() in fm_chain.go (chain: Apple → Ollama → stub).
// Prefer DefaultFMProvider() for consistency with DefaultReportInsight() and DefaultOllama().
var DefaultFM FMProvider

// DefaultFMProvider returns the default FM provider (set in init; never nil).
// Use for consistency with DefaultReportInsight() and DefaultOllama().
func DefaultFMProvider() FMProvider {
	return DefaultFM
}

// FMAvailable reports whether a foundation model is available (non-nil and supported).
func FMAvailable() bool {
	p := DefaultFMProvider()
	return p != nil && p.Supported()
}
