package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
)

func TestHandleSecurityScan(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROJECT_ROOT", tmpDir)

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		validate  func(*testing.T, []framework.TextContent)
	}{
		{
			name: "scan action (requires go.mod; tmpDir has none)",
			params: map[string]interface{}{
				"action": "scan",
			},
			wantError: true, // security scan is only supported for Go projects
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			result, err := handleSecurityScan(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleSecurityScan() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestHandleSecurityAlerts(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROJECT_ROOT", tmpDir)

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		validate  func(*testing.T, []framework.TextContent)
	}{
		{
			name: "alerts action",
			params: map[string]interface{}{
				"action": "alerts",
				"repo":   "davidl71/exarp-go",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
				// Result may be alerts or error message (if gh CLI not available)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			result, err := handleSecurityAlerts(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleSecurityAlerts() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestHandleSecurity(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROJECT_ROOT", tmpDir)

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name: "scan action (no go.mod in tmpDir)",
			params: map[string]interface{}{
				"action": "scan",
			},
			wantError: true,
		},
		{
			name: "alerts action",
			params: map[string]interface{}{
				"action": "alerts",
			},
			wantError: false,
		},
		{
			name: "report action",
			params: map[string]interface{}{
				"action": "report",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			argsJSON, _ := json.Marshal(tt.params)

			result, err := handleSecurity(ctx, argsJSON)
			if (err != nil) != tt.wantError {
				t.Errorf("handleSecurity() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && len(result) == 0 {
				t.Error("expected non-empty result")
			}
		})
	}
}
