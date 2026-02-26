package tools

import (
	"context"
	"encoding/json"
	"sort"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
)

func TestHandleWorkflowModeNative(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROJECT_ROOT", tmpDir)

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		validate  func(*testing.T, []framework.TextContent)
	}{
		{
			name: "focus action",
			params: map[string]interface{}{
				"action": "focus",
				"mode":   "development",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}

				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}

				if success, ok := data["success"].(bool); !ok || !success {
					t.Error("expected success=true")
				}
			},
		},
		{
			name: "suggest action",
			params: map[string]interface{}{
				"action": "suggest",
				"text":   "I need to implement a new feature",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				// Should suggest a workflow mode
			},
		},
		{
			name: "stats action",
			params: map[string]interface{}{
				"action": "stats",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}

				if success, ok := data["success"].(bool); !ok || !success {
					t.Error("expected success=true")
				}
			},
		},
		{
			name: "unknown action",
			params: map[string]interface{}{
				"action": "unknown",
			},
			wantError: false, // Returns error response, not an error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			result, err := handleWorkflowModeNative(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleWorkflowModeNative() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestToolFilterForMode(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROJECT_ROOT", tmpDir)

	// Reset global manager so each subtest starts fresh
	resetManager := func(mode string) {
		globalWorkflowManager = nil
		m := getWorkflowManager()
		m.state.CurrentMode = mode
		m.state.ExtraGroups = nil
		m.state.DisabledGroups = nil
	}

	allTools := []framework.ToolInfo{
		{Name: "task_workflow"},
		{Name: "session"},
		{Name: "report"},
		{Name: "health"},
		{Name: "tool_catalog"},
		{Name: "workflow_mode"},
		{Name: "list_resources"},
		{Name: "read_resource"},
		{Name: "task_analysis"},
		{Name: "task_discovery"},
		{Name: "memory"},
		{Name: "security"},
		{Name: "automation"},
		{Name: "testing"},
		{Name: "git_tools"},
		{Name: "ollama"},
		{Name: "recommend"},
		{Name: "analyze_alignment"},
		{Name: "generate_config"},
		{Name: "unknown_tool"},
	}

	toolNames := func(tools []framework.ToolInfo) []string {
		names := make([]string, len(tools))
		for i, t := range tools {
			names[i] = t.Name
		}
		sort.Strings(names)
		return names
	}

	t.Run("core and tool_catalog always visible", func(t *testing.T) {
		for _, mode := range []string{"task_management", "security_review", "development", "all"} {
			resetManager(mode)
			filter := ToolFilterForMode()
			result := filter(context.Background(), allTools)
			names := toolNames(result)
			for _, required := range []string{"task_workflow", "session", "report", "health", "tool_catalog", "workflow_mode"} {
				found := false
				for _, n := range names {
					if n == required {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("mode %s: expected %s to be visible", mode, required)
				}
			}
		}
	})

	t.Run("unknown tools always visible", func(t *testing.T) {
		resetManager("security_review")
		filter := ToolFilterForMode()
		result := filter(context.Background(), allTools)
		names := toolNames(result)
		found := false
		for _, n := range names {
			if n == "unknown_tool" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected unknown_tool to be visible in any mode")
		}
	})

	t.Run("security_review hides tasks and memory", func(t *testing.T) {
		resetManager("security_review")
		filter := ToolFilterForMode()
		result := filter(context.Background(), allTools)
		names := toolNames(result)
		for _, hidden := range []string{"task_analysis", "task_discovery", "memory", "ollama", "recommend"} {
			for _, n := range names {
				if n == hidden {
					t.Errorf("security_review: expected %s to be hidden", hidden)
				}
			}
		}
	})

	t.Run("all mode shows everything", func(t *testing.T) {
		resetManager("all")
		filter := ToolFilterForMode()
		result := filter(context.Background(), allTools)
		if len(result) != len(allTools) {
			t.Errorf("all mode: expected %d tools, got %d", len(allTools), len(result))
		}
	})

	t.Run("extra_groups override", func(t *testing.T) {
		resetManager("security_review")
		m := getWorkflowManager()
		m.state.ExtraGroups = []string{"memory"}
		filter := ToolFilterForMode()
		result := filter(context.Background(), allTools)
		names := toolNames(result)
		found := false
		for _, n := range names {
			if n == "memory" {
				found = true
				break
			}
		}
		if !found {
			t.Error("extra_groups: expected memory to be visible after adding to extra")
		}
	})

	t.Run("disabled_groups override", func(t *testing.T) {
		resetManager("development")
		m := getWorkflowManager()
		m.state.DisabledGroups = []string{"memory"}
		filter := ToolFilterForMode()
		result := filter(context.Background(), allTools)
		names := toolNames(result)
		for _, n := range names {
			if n == "memory" {
				t.Error("disabled_groups: expected memory to be hidden after disabling")
			}
		}
	})

	t.Run("cannot disable core via disabled_groups", func(t *testing.T) {
		resetManager("development")
		m := getWorkflowManager()
		m.state.DisabledGroups = []string{"core"}
		filter := ToolFilterForMode()
		result := filter(context.Background(), allTools)
		names := toolNames(result)
		found := false
		for _, n := range names {
			if n == "task_workflow" {
				found = true
				break
			}
		}
		if !found {
			t.Error("core should remain visible even when listed in disabled_groups")
		}
	})
}

func TestHandleWorkflowMode(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROJECT_ROOT", tmpDir)

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name: "focus action",
			params: map[string]interface{}{
				"action": "focus",
				"mode":   "development",
			},
			wantError: false,
		},
		{
			name: "suggest action",
			params: map[string]interface{}{
				"action": "suggest",
			},
			wantError: false,
		},
		{
			name: "stats action",
			params: map[string]interface{}{
				"action": "stats",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			argsJSON, _ := json.Marshal(tt.params)

			result, err := handleWorkflowMode(ctx, json.RawMessage(argsJSON))
			if (err != nil) != tt.wantError {
				t.Errorf("handleWorkflowMode() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && len(result) == 0 {
				t.Error("expected non-empty result")
			}
		})
	}
}
