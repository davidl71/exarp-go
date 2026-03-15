// client.go — MCP client identity context helpers.
package framework

import "context"

type clientNameKey struct{}

// WithClientName stores the MCP client name in context.
// The name is typically injected from the X-Client-Name HTTP header
// or from the client param passed to session prime.
func WithClientName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, clientNameKey{}, name)
}

// ClientNameFromContext retrieves the MCP client name from context.
// Returns empty string if not set.
func ClientNameFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(clientNameKey{}).(string); ok {
		return v
	}
	return ""
}
