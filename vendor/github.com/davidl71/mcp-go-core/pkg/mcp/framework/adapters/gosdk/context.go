package gosdk

import (
	"context"
	"fmt"
)

// ValidateContext checks if context is valid and not cancelled
func ValidateContext(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}
	return nil
}
