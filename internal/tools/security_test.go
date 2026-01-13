package tools

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
)

func TestHandleSecurityScan(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PROJECT_ROOT", tmpDir)
	defer os.Unsetenv("PROJECT_ROOT")

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		validate  func(*testing.T, []framework.TextContent)
	}{
		{
			name: "scan action for Go project",
			params: map[string]interface{}{
				"action": "scan",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
				// Result may be scan results or error message
			},
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
	os.Setenv("PROJECT_ROOT", tmpDir)
	defer os.Unsetenv("PROJECT_ROOT")

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
	os.Setenv("PROJECT_ROOT", tmpDir)
	defer os.Unsetenv("PROJECT_ROOT")

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name: "scan action",
			params: map[string]interface{}{
				"action": "scan",
			},
			wantError: false,
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
