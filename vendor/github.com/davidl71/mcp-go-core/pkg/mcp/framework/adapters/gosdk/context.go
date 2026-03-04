package gosdk

import (
	"context"

	"github.com/davidl71/mcp-go-core/pkg/mcp/framework"
)

// ValidateContext checks if context is valid and not cancelled
func ValidateContext(ctx context.Context) error {
	if ctx == nil {
		return &framework.ErrContextCancelled{Err: nil}
	}
	if err := ctx.Err(); err != nil {
		return &framework.ErrContextCancelled{Err: err}
	}
	return nil
}
