//go:build !llamacpp || !cgo
// +build !llamacpp !cgo

// llamacpp_nocgo.go — stub provider when llamacpp build tag is not set or CGO is disabled.
// Returns nil/unsupported so the rest of the codebase compiles cleanly.
package tools

import "fmt"

// DefaultLlamaCppProvider returns nil when llamacpp is not built.
func DefaultLlamaCppProvider() TextGenerator {
	return nil
}

// llamacppIfAvailable returns nil when llamacpp is not built (used by FM chain).
func llamacppIfAvailable() TextGenerator {
	return nil
}

// LlamaCppLoad returns an error when llamacpp is not built.
func LlamaCppLoad(_ string) error {
	return fmt.Errorf("llamacpp not available: build with -tags llamacpp,cgo and libbinding.a")
}

// LlamaCppUnload is a no-op when llamacpp is not built.
func LlamaCppUnload() {}

// LlamaCppModelPath returns empty string when llamacpp is not built.
func LlamaCppModelPath() string { return "" }

// ModelManagerStats returns empty stats when llamacpp is not built.
func ModelManagerStats() map[string]interface{} {
	return map[string]interface{}{"available": false}
}
