// Package framework provides framework-agnostic abstractions for MCP servers.
// This file adds MCP Roots support: client workspace boundaries.

package framework

import "context"

// Root represents a client workspace root (e.g., a folder in the IDE).
type Root struct {
	URI  string `json:"uri"`
	Name string `json:"name,omitempty"`
}

// RootsHandler represents a handler that can provide client roots.
// Similar to Eliciter, this allows tools to know the client's workspace boundaries.
type RootsHandler interface {
	// GetRoots returns the current list of client roots.
	GetRoots() []Root
}

// rootsKey is the context key for roots.
type rootsKey struct{}

// RootsFromContext returns the client roots from context, or nil if not set.
func RootsFromContext(ctx context.Context) []Root {
	if ctx == nil {
		return nil
	}
	if v := ctx.Value(rootsKey{}); v != nil {
		if roots, ok := v.([]Root); ok {
			return roots
		}
	}
	return nil
}

// ContextWithRoots returns a new context with the roots set.
func ContextWithRoots(ctx context.Context, roots []Root) context.Context {
	return context.WithValue(ctx, rootsKey{}, roots)
}
