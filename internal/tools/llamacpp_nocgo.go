//go:build !llamacpp || !cgo
// +build !llamacpp !cgo

// llamacpp_nocgo.go — stub provider when llamacpp build tag is not set or CGO is disabled.
// Returns nil/unsupported so the rest of the codebase compiles cleanly.
package tools

// DefaultLlamaCppProvider returns nil when llamacpp is not built.
func DefaultLlamaCppProvider() TextGenerator {
	return nil
}

// llamacppIfAvailable returns nil when llamacpp is not built (used by FM chain).
func llamacppIfAvailable() TextGenerator {
	return nil
}
