package tools

import (
	"context"
	"encoding/json"
	"strings"
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
			name: "scan action (no go/python/rust/node in tmpDir)",
			params: map[string]interface{}{
				"action": "scan",
			},
			wantError: false, // multilang: returns success with message when no ecosystem detected
			validate: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
				text := result[0].Text
				if !strings.Contains(text, "No Go") && !strings.Contains(text, "No supported") && !strings.Contains(text, "vulnerabilities") {
					t.Errorf("expected result to mention no project or vulnerabilities, got: %s", text)
				}
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
			wantError: false, // multilang scan returns success with message
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

// TestHandleScanDependencySecurity verifies the scan_dependency_security alias (multilang; no Go-only error).
func TestHandleScanDependencySecurity(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROJECT_ROOT", tmpDir)

	ctx := context.Background()
	result, err := handleScanDependencySecurity(ctx, []byte("{}"))
	if err != nil {
		t.Fatalf("handleScanDependencySecurity() error = %v (should succeed for any project)", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty result")
		return
	}
	text := result[0].Text
	if strings.Contains(text, "only supported for Go") {
		t.Errorf("scan_dependency_security must not return Go-only message; got: %s", text)
	}
}
