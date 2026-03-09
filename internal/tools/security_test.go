package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
)

func withMockDependabotAlertsCommand(t *testing.T, fn func(context.Context, string, string) ([]byte, error)) {
	t.Helper()

	original := runDependabotAlertsCommand
	runDependabotAlertsCommand = fn
	t.Cleanup(func() {
		runDependabotAlertsCommand = original
	})
}

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
		mock      func(context.Context, string, string) ([]byte, error)
		validate  func(*testing.T, []framework.TextContent)
	}{
		{
			name: "alerts action",
			params: map[string]interface{}{
				"action": "alerts",
				"repo":   "davidl71/exarp-go",
			},
			wantError: false,
			mock: func(_ context.Context, repo, _ string) ([]byte, error) {
				if repo != "davidl71/exarp-go" {
					t.Fatalf("repo = %q, want davidl71/exarp-go", repo)
				}
				return []byte("{\"package\":\"cobra\",\"severity\":\"high\",\"cve\":\"CVE-2026-0001\",\"state\":\"open\",\"ecosystem\":\"go\",\"description\":\"mock alert\",\"fix_available\":true,\"fixed_version\":\"1.9.0\"}\n"), nil
			},
			validate: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
				text := result[0].Text
				if !strings.Contains(text, "Total Alerts: 1") {
					t.Errorf("expected single alert summary, got: %s", text)
				}
				if !strings.Contains(text, "cobra") {
					t.Errorf("expected package name in output, got: %s", text)
				}
			},
		},
		{
			name: "alerts action filters closed alerts when state open",
			params: map[string]interface{}{
				"action": "alerts",
				"repo":   "davidl71/exarp-go",
				"state":  "open",
			},
			wantError: false,
			mock: func(context.Context, string, string) ([]byte, error) {
				return []byte("{\"package\":\"cobra\",\"severity\":\"high\",\"cve\":\"CVE-2026-0001\",\"state\":\"closed\",\"ecosystem\":\"go\",\"description\":\"closed alert\",\"fix_available\":true,\"fixed_version\":\"1.9.0\"}\n"), nil
			},
			validate: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Fatal("expected non-empty result")
				}
				if !strings.Contains(result[0].Text, "Total Alerts: 0") {
					t.Errorf("expected closed alert to be filtered out, got: %s", result[0].Text)
				}
			},
		},
		{
			name: "alerts action command failure",
			params: map[string]interface{}{
				"action": "alerts",
				"repo":   "davidl71/exarp-go",
			},
			wantError: true,
			mock: func(context.Context, string, string) ([]byte, error) {
				return []byte("boom"), errors.New("mock gh failure")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			withMockDependabotAlertsCommand(t, tt.mock)

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
	withMockDependabotAlertsCommand(t, func(context.Context, string, string) ([]byte, error) {
		return []byte("{\"package\":\"cobra\",\"severity\":\"high\",\"cve\":\"CVE-2026-0001\",\"state\":\"open\",\"ecosystem\":\"go\",\"description\":\"mock alert\",\"fix_available\":true,\"fixed_version\":\"1.9.0\"}\n"), nil
	})

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

			if !tt.wantError && tt.params["action"] == "alerts" && !strings.Contains(result[0].Text, "Total Alerts: 1") {
				t.Errorf("expected mocked alerts output, got: %s", result[0].Text)
			}
		})
	}
}
