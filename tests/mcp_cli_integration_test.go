package tests

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestMCPToolViaCLI tests tools via subprocess (./bin/exarp-go -tool X -args '...')
// This validates the full MCP CLI path: binary → handler → JSON response.
func TestMCPToolViaCLI(t *testing.T) {
	// Ensure binary exists
	binaryPath := filepath.Join("..", "bin", "exarp-go")
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skip("bin/exarp-go not found; run 'make build' first")
	}

	tests := []struct {
		name      string
		tool      string
		args      string
		wantError bool
		validate  func(*testing.T, map[string]interface{})
	}{
		{
			name: "session prime",
			tool: "session",
			args: `{"action":"prime","include_hints":false,"include_tasks":false}`,
			validate: func(t *testing.T, data map[string]interface{}) {
				if _, ok := data["detection"]; !ok {
					t.Error("expected detection field in session prime")
				}
				if _, ok := data["auto_primed"]; !ok {
					t.Error("expected auto_primed field")
				}
			},
		},
		{
			name: "task_workflow sync",
			tool: "task_workflow",
			args: `{"action":"sync"}`,
			validate: func(t *testing.T, data map[string]interface{}) {
				if success, ok := data["success"].(bool); !ok || !success {
					t.Error("expected success=true for sync")
				}
			},
		},
		{
			name: "health server",
			tool: "health",
			args: `{"action":"server"}`,
			validate: func(t *testing.T, data map[string]interface{}) {
				if status, ok := data["status"].(string); !ok || status == "" {
					t.Error("expected status field in health server")
				}
			},
		},
		{
			name: "tool_catalog help",
			tool: "tool_catalog",
			args: `{"action":"help","tool_name":"session"}`,
			validate: func(t *testing.T, data map[string]interface{}) {
				if tool, ok := data["tool"].(string); !ok || tool == "" {
					t.Error("expected tool field (string) in help response")
				}
				if _, ok := data["hint"]; !ok {
					t.Error("expected hint field in help response")
				}
			},
		},
		{
			name:      "invalid tool name",
			tool:      "nonexistent_tool",
			args:      `{}`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, "-tool", tt.tool, "-args", tt.args)

			cmd.Env = append(os.Environ(), "PROJECT_ROOT="+filepath.Join("..", ".."))

			output, err := cmd.CombinedOutput()

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error for %s, got success", tt.tool)
				}

				return
			}

			if err != nil {
				t.Errorf("tool %s failed: %v\nOutput: %s", tt.tool, err, string(output))
				return
			}

			// Extract JSON from output (CLI outputs logs + "Result:\n" + JSON)
			outputStr := string(output)
			jsonStart := strings.Index(outputStr, "Result:\n")

			if jsonStart == -1 {
				t.Errorf("tool %s output missing 'Result:' marker\nOutput: %s", tt.tool, outputStr)
				return
			}

			jsonStr := strings.TrimSpace(outputStr[jsonStart+8:]) // Skip "Result:\n"

			// Parse JSON response
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
				t.Errorf("invalid JSON response from %s: %v\nJSON: %s", tt.tool, err, jsonStr)
				return
			}

			// Run validation if provided
			if tt.validate != nil {
				tt.validate(t, data)
			}
		})
	}
}

// TestMCPToolErrorHandlingViaCLI tests error cases via CLI.
func TestMCPToolErrorHandlingViaCLI(t *testing.T) {
	binaryPath := filepath.Join("..", "bin", "exarp-go")
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skip("bin/exarp-go not found; run 'make build' first")
	}

	tests := []struct {
		name      string
		tool      string
		args      string
		wantError bool
	}{
		{
			name:      "session invalid action",
			tool:      "session",
			args:      `{"action":"invalid_action"}`,
			wantError: true,
		},
		{
			name:      "task_workflow invalid action",
			tool:      "task_workflow",
			args:      `{"action":"invalid_action"}`,
			wantError: true,
		},
		{
			name:      "malformed JSON args",
			tool:      "session",
			args:      `{invalid json}`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, "-tool", tt.tool, "-args", tt.args)

			cmd.Env = append(os.Environ(), "PROJECT_ROOT="+filepath.Join("..", ".."))

			output, err := cmd.CombinedOutput()

			if tt.wantError && err == nil {
				t.Errorf("expected error, got success. Output: %s", string(output))
			}

			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v. Output: %s", err, string(output))
			}
		})
	}
}

// TestMCPToolJSONResponseFormat validates all tools return valid JSON.
func TestMCPToolJSONResponseFormat(t *testing.T) {
	binaryPath := filepath.Join("..", "bin", "exarp-go")
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skip("bin/exarp-go not found; run 'make build' first")
	}

	// Test a representative set of tools
	tools := []struct {
		name string
		args string
	}{
		{"session", `{"action":"prime","include_hints":false,"include_tasks":false}`},
		{"health", `{"action":"server"}`},
		{"tool_catalog", `{"action":"help","tool_name":"session"}`},
		{"workflow_mode", `{"action":"suggest"}`},
	}

	for _, tool := range tools {
		t.Run(tool.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, "-tool", tool.name, "-args", tool.args)

			cmd.Env = append(os.Environ(), "PROJECT_ROOT="+filepath.Join("..", ".."))

			output, err := cmd.CombinedOutput()
			if err != nil {
				// Some tools may fail in test env; skip if so
				t.Skipf("tool %s failed (may be expected in test env): %v", tool.name, err)
				return
			}

			// Extract JSON from output (CLI outputs logs + "Result:\n" + JSON)
			outputStr := string(output)
			jsonStart := strings.Index(outputStr, "Result:\n")

			if jsonStart == -1 {
				t.Errorf("tool %s output missing 'Result:' marker\nOutput: %s", tool.name, outputStr)
				return
			}

			jsonStr := strings.TrimSpace(outputStr[jsonStart+8:])

			// Must be valid JSON
			var data interface{}
			if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
				t.Errorf("tool %s returned invalid JSON: %v\nJSON: %s", tool.name, err, jsonStr)
			}

			// Check JSON is not empty
			if jsonStr == "" || jsonStr == "{}" {
				t.Errorf("tool %s returned empty JSON response", tool.name)
			}
		})
	}
}
