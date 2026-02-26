package framework

import "context"

// eliciterKey is the context key for the Eliciter.
type eliciterKey struct{}

// Eliciter allows tools to request structured input from the client (MCP elicitation).
// When the client supports elicitation, the gosdk adapter injects an Eliciter into the context.
type Eliciter interface {
	// ElicitForm asks the user for structured input.
	// Returns action ("accept", "decline", "cancel"), content map, and error.
	ElicitForm(ctx context.Context, message string, schema map[string]interface{}) (action string, content map[string]interface{}, err error)
}

// ContextWithEliciter returns a context with the eliciter attached.
func ContextWithEliciter(ctx context.Context, eliciter Eliciter) context.Context {
	return context.WithValue(ctx, eliciterKey{}, eliciter)
}

// EliciterFromContext returns the Eliciter from the context, or nil if not set.
func EliciterFromContext(ctx context.Context) Eliciter {
	if ctx == nil {
		return nil
	}
	e, _ := ctx.Value(eliciterKey{}).(Eliciter)
	return e
}
