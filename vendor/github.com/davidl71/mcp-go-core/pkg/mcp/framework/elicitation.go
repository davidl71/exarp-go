// Package framework provides framework-agnostic abstractions for MCP servers.
// This file adds MCP Elicitation support for client-side user prompts (e.g. forms).

package framework

import "context"

// Eliciter represents a client that can elicit user input via MCP forms.
// When set in context (e.g. by a Cursor integration), tools can prompt the user
// for preferences before proceeding. EliciterFromContext returns nil if no
// eliciter is configured, allowing tools to degrade gracefully.
type Eliciter interface {
	// ElicitForm prompts the user with a form. Returns action ("accept"/"decline"),
	// content (form values when action is "accept"), and error.
	ElicitForm(ctx context.Context, prompt string, schema map[string]interface{}) (action string, content map[string]interface{}, err error)
}

type eliciterKey struct{}

// EliciterFromContext returns the Eliciter from context, or nil if not set.
func EliciterFromContext(ctx context.Context) Eliciter {
	if ctx == nil {
		return nil
	}
	if v := ctx.Value(eliciterKey{}); v != nil {
		if e, ok := v.(Eliciter); ok {
			return e
		}
	}
	return nil
}

// ContextWithEliciter returns a new context with the Eliciter set.
func ContextWithEliciter(ctx context.Context, e Eliciter) context.Context {
	return context.WithValue(ctx, eliciterKey{}, e)
}
