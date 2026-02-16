package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
	mcpframework "github.com/davidl71/mcp-go-core/pkg/mcp/framework"
)

func TestHandleSessionPrompts(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		validate  func(*testing.T, []framework.TextContent)
	}{
		{
			name: "all prompts",
			params: map[string]interface{}{
				"action": "prompts",
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
				if method, ok := data["method"].(string); !ok || method != "native_go" {
					t.Error("expected method=native_go")
				}
			},
		},
		{
			name: "filter by mode",
			params: map[string]interface{}{
				"action": "prompts",
				"mode":   "daily_checkin",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				if filters, ok := data["filters_applied"].(map[string]interface{}); ok {
					if mode, ok := filters["mode"].(string); !ok || mode != "daily_checkin" {
						t.Errorf("expected mode filter, got %v", filters)
					}
				}
			},
		},
		{
			name: "filter by category",
			params: map[string]interface{}{
				"action":   "prompts",
				"category": "workflow",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				if filters, ok := data["filters_applied"].(map[string]interface{}); ok {
					if category, ok := filters["category"].(string); !ok || category != "workflow" {
						t.Errorf("expected category filter, got %v", filters)
					}
				}
			},
		},
		{
			name: "filter by keywords",
			params: map[string]interface{}{
				"action":   "prompts",
				"keywords": []string{"align", "discover"},
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				if filters, ok := data["filters_applied"].(map[string]interface{}); ok {
					if keywords, ok := filters["keywords"].([]interface{}); !ok || len(keywords) == 0 {
						t.Errorf("expected keywords filter, got %v", filters)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			result, err := handleSessionPrompts(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleSessionPrompts() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestHandleSessionAssignee(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		validate  func(*testing.T, []framework.TextContent)
	}{
		{
			name: "basic assignee request",
			params: map[string]interface{}{
				"action": "assignee",
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
				if method, ok := data["method"].(string); !ok || method != "native_go" {
					t.Error("expected method=native_go")
				}
			},
		},
		{
			name: "filter by status",
			params: map[string]interface{}{
				"action":        "assignee",
				"status_filter": "Todo",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				if filters, ok := data["filters_applied"].(map[string]interface{}); ok {
					if status, ok := filters["status_filter"].(string); !ok || status != "Todo" {
						t.Errorf("expected status filter, got %v", filters)
					}
				}
			},
		},
		{
			name: "filter by priority",
			params: map[string]interface{}{
				"action":          "assignee",
				"priority_filter": "high",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				if filters, ok := data["filters_applied"].(map[string]interface{}); ok {
					if priority, ok := filters["priority_filter"].(string); !ok || priority != "high" {
						t.Errorf("expected priority filter, got %v", filters)
					}
				}
			},
		},
		{
			name: "with assignee name",
			params: map[string]interface{}{
				"action":        "assignee",
				"assignee_name": "test-agent",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				if name, ok := data["assignee_name"].(string); !ok || name != "test-agent" {
					t.Errorf("expected assignee_name=test-agent, got %v", data["assignee_name"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			result, err := handleSessionAssignee(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleSessionAssignee() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

// mockEliciter implements mcpframework.Eliciter for tests (MCP Elicitation).
type mockEliciter struct {
	Action  string
	Content map[string]interface{}
	Err     error
}

func (m *mockEliciter) ElicitForm(_ context.Context, _ string, _ map[string]interface{}) (string, map[string]interface{}, error) {
	return m.Action, m.Content, m.Err
}

func TestHandleSessionPrimeElicitation(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		params    map[string]interface{}
		wantError bool
		wantTasks bool // result should include tasks summary
		wantHints bool // result should include hints
	}{
		{
			name:      "no eliciter ask_preferences true uses defaults",
			ctx:       context.Background(),
			params:    map[string]interface{}{"action": "prime", "ask_preferences": true},
			wantError: false,
			wantTasks: true,
			wantHints: true,
		},
		{
			name: "mock accept include_tasks false include_hints false",
			ctx: mcpframework.ContextWithEliciter(context.Background(), &mockEliciter{
				Action:  "accept",
				Content: map[string]interface{}{"include_tasks": false, "include_hints": false},
			}),
			params:    map[string]interface{}{"action": "prime", "ask_preferences": true},
			wantError: false,
			wantTasks: false,
			wantHints: false,
		},
		{
			name: "mock decline keeps defaults",
			ctx: mcpframework.ContextWithEliciter(context.Background(), &mockEliciter{
				Action: "decline",
			}),
			params:    map[string]interface{}{"action": "prime", "ask_preferences": true},
			wantError: false,
			wantTasks: true,
			wantHints: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handleSessionPrime(tt.ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleSessionPrime() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if tt.wantError {
				return
			}

			if result == nil || len(result) == 0 {
				t.Fatal("expected non-empty result")
			}

			var data map[string]interface{}
			if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
				t.Fatalf("invalid JSON: %v", err)
			}

			_, hasTasks := data["tasks"]
			if hasTasks != tt.wantTasks {
				t.Errorf("result has tasks = %v, want %v", hasTasks, tt.wantTasks)
			}

			_, hasHints := data["hints"]
			if hasHints != tt.wantHints {
				t.Errorf("result has hints = %v, want %v", hasHints, tt.wantHints)
			}
		})
	}
}

func TestHandleSessionNative(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name: "prompts action",
			params: map[string]interface{}{
				"action": "prompts",
			},
			wantError: false,
		},
		{
			name: "assignee action",
			params: map[string]interface{}{
				"action": "assignee",
			},
			wantError: false,
		},
		{
			name: "prime action",
			params: map[string]interface{}{
				"action": "prime",
			},
			wantError: false,
		},
		{
			name: "handoff action",
			params: map[string]interface{}{
				"action": "handoff",
			},
			wantError: false,
		},
		{
			name: "unknown action",
			params: map[string]interface{}{
				"action": "unknown",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			result, err := handleSessionNative(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleSessionNative() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && (result == nil || len(result) == 0) {
				t.Error("expected non-empty result")
			}
		})
	}
}

func TestBuildCursorCLISuggestion(t *testing.T) {
	tests := []struct {
		name string
		task map[string]interface{}
		want string
	}{
		{
			name: "full task",
			task: map[string]interface{}{"id": "T-123", "content": "Proto Task workflow response types"},
			want: `agent -p "Work on T-123: Proto Task workflow response types" --mode=plan`,
		},
		{
			name: "id only",
			task: map[string]interface{}{"id": "T-456"},
			want: `agent -p "Work on T-456" --mode=plan`,
		},
		{
			name: "empty id",
			task: map[string]interface{}{"content": "something"},
			want: "",
		},
		{
			name: "long content truncated",
			task: map[string]interface{}{"id": "T-789", "content": "This is a very long task name that definitely exceeds sixty characters and should be truncated"},
			want: `agent -p "Work on T-789: This is a very long task name that definitely exceeds six..." --mode=plan`,
		},
		{
			name: "content with quotes",
			task: map[string]interface{}{"id": "T-42", "content": `Fix "broken" thing`},
			want: `agent -p "Work on T-42: Fix \"broken\" thing" --mode=plan`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildCursorCLISuggestion(tt.task)
			if got != tt.want {
				t.Errorf("buildCursorCLISuggestion() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestShouldSuggestPlanMode(t *testing.T) {
	tests := []struct {
		name  string
		tasks []Todo2Task
		want  bool
	}{
		{
			name:  "empty tasks",
			tasks: nil,
			want:  false,
		},
		{
			name: "no backlog",
			tasks: []Todo2Task{
				{ID: "T-1", Status: "Done", Priority: "high"},
			},
			want: false,
		},
		{
			name:  "small backlog",
			tasks: makeBacklogTasks(5, 0, 0),
			want:  false,
		},
		{
			name:  "large backlog hits threshold",
			tasks: makeBacklogTasks(20, 0, 0),
			want:  true,
		},
		{
			name:  "many high priority hits threshold",
			tasks: makeBacklogTasksWithPriority(10, 6),
			want:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldSuggestPlanMode(tt.tasks); got != tt.want {
				t.Errorf("shouldSuggestPlanMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func makeBacklogTasks(n, highPriority, withDeps int) []Todo2Task {
	out := make([]Todo2Task, n)

	for i := 0; i < n; i++ {
		priority := "medium"
		if i < highPriority {
			priority = "high"
		}

		deps := []string{}
		if i < withDeps {
			deps = []string{"T-0"}
		}

		out[i] = Todo2Task{ID: "T-" + strconv.Itoa(i), Status: "Todo", Priority: priority, Dependencies: deps}
	}

	return out
}

func makeBacklogTasksWithPriority(total, highCount int) []Todo2Task {
	out := make([]Todo2Task, total)

	for i := 0; i < total; i++ {
		priority := "medium"
		if i < highCount {
			priority = "high"
		}

		out[i] = Todo2Task{ID: "T-" + strconv.Itoa(i), Status: "Todo", Priority: priority}
	}

	return out
}

func TestGetCurrentPlanPath(t *testing.T) {
	tmpDir := t.TempDir()

	projDir := filepath.Join(tmpDir, "myproj")
	if err := os.MkdirAll(projDir, 0755); err != nil {
		t.Fatalf("mkdir proj: %v", err)
	}

	plansDir := filepath.Join(projDir, ".cursor", "plans")
	if err := os.MkdirAll(plansDir, 0755); err != nil {
		t.Fatalf("mkdir plans: %v", err)
	}

	planFile := filepath.Join(plansDir, "myproj.plan.md")
	if err := os.WriteFile(planFile, []byte("# Plan\n"), 0644); err != nil {
		t.Fatalf("write plan: %v", err)
	}

	got := getCurrentPlanPath(projDir)
	want := filepath.Join(".cursor", "plans", "myproj.plan.md")

	if got != want {
		t.Errorf("getCurrentPlanPath() = %q, want %q", got, want)
	}
}
