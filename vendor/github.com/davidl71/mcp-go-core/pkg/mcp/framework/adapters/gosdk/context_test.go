package gosdk

import (
	"context"
	"testing"
	"time"

	"github.com/davidl71/mcp-go-core/pkg/mcp/framework"
)

func TestValidateContext(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid context",
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:    "nil context",
			ctx:     nil,
			wantErr: true,
			errMsg:  "context cannot be nil",
		},
		{
			name: "cancelled context",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			wantErr: true,
			errMsg:  "context cancelled",
		},
		{
			name: "context with timeout",
			ctx: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
				time.Sleep(time.Millisecond) // Wait for timeout
				cancel()
				return ctx
			}(),
			wantErr: true,
			errMsg:  "context cancelled",
		},
		{
			name:    "context with value",
			ctx:     context.WithValue(context.Background(), "key", "value"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateContext(tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err == nil {
					t.Error("ValidateContext() expected error but got nil")
					return
				}
				if tt.errMsg != "" {
					errorMsg := err.Error()
					// Check if error message contains expected string
					found := false
					for i := 0; i <= len(errorMsg)-len(tt.errMsg); i++ {
						if errorMsg[i:i+len(tt.errMsg)] == tt.errMsg {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("ValidateContext() error = %q, want error containing %q", errorMsg, tt.errMsg)
					}
				}
				// Verify typed error is ErrContextCancelled when context was cancelled/nil
				if tt.wantErr && err != nil && !framework.IsContextCancelled(err) {
					t.Error("ValidateContext() should return ErrContextCancelled for nil or cancelled context")
				}
			}
		})
	}
}
