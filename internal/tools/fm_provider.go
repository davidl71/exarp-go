// Package tools: FM provider abstraction for foundation model operations.
// Implementations are in fm_apple.go (darwin,arm64,cgo) and fm_stub.go (other platforms).

package tools

import (
	"context"
	"errors"
)

// ErrFMNotSupported is returned when a foundation model is requested but not available on this platform.
var ErrFMNotSupported = errors.New("foundation model not supported on this platform")

// FMProvider abstracts foundation model access so tools (e.g. task_analysis hierarchy)
// can use Apple FM when available and fail cleanly otherwise, without Python fallback.
type FMProvider interface {
	// Supported reports whether a foundation model is available on this platform.
	Supported() bool
	// Generate runs the model with the given prompt and options; returns ErrFMNotSupported if not supported.
	Generate(ctx context.Context, prompt string, maxTokens int, temperature float32) (string, error)
}

// DefaultFM is set by init() in fm_apple.go (when CGO) or fm_stub.go (otherwise).
// Tools use this for hierarchy, classification, etc., so they stay all-Go.
var DefaultFM FMProvider
