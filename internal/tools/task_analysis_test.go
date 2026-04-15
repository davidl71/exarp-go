package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
)

func TestHandleTaskAnalysisNative(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROJECT_ROOT", tmpDir)

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		validate  func(*testing.T, []framework.TextContent)
	}{
		{
			name: "duplicates action",
			params: map[string]interface{}{
				"action": "duplicates",
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
			},
		},
		{
			name: "tags action",
			params: map[string]interface{}{
				"action": "tags",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
			},
		},
		{
			name: "tags action with limit and prioritize_untagged",
			params: map[string]interface{}{
				"action":              "tags",
				"limit":               10,
				"prioritize_untagged": true,
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}

				if _, ok := data["tag_analysis"]; !ok {
					t.Error("expected tag_analysis in result")
				}
			},
		},
		{
			name: "conflicts action",
			params: map[string]interface{}{
				"action": "conflicts",
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

				if _, ok := data["conflict"]; !ok {
					t.Error("expected conflict key in result")
				}

				if _, ok := data["conflicts"]; !ok {
					t.Error("expected conflicts key in result")
				}
			},
		},
		{
			name: "dependencies action",
			params: map[string]interface{}{
				"action": "dependencies",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
				// Accept either JSON envelope or human-readable text (e.g. output_format=text)
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					if len(result[0].Text) == 0 {
						t.Error("expected non-empty result (JSON or text)")
					}

					return
				}
			},
		},
		{
			name: "parallelization action",
			params: map[string]interface{}{
				"action": "parallelization",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
				// Accept either JSON envelope or human-readable text (e.g. output_format=text)
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					if len(result[0].Text) == 0 {
						t.Error("expected non-empty result (JSON or text)")
					}

					return
				}
			},
		},
		{
			name: "validate action",
			params: map[string]interface{}{
				"action": "validate",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}

				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("validate action should return JSON: %v", err)
					return
				}

				if _, ok := data["missing_deps"]; !ok {
					t.Error("validate result should contain missing_deps")
				}

				if _, ok := data["missing_count"]; !ok {
					t.Error("validate result should contain missing_count")
				}
			},
		},
		{
			name: "noise action",
			params: map[string]interface{}{
				"action": "noise",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("noise action should return JSON: %v", err)
					return
				}
				if _, ok := data["noise_candidates"]; !ok {
					t.Error("noise result should contain noise_candidates")
				}
				if _, ok := data["filter_tag"]; !ok {
					t.Error("noise result should contain filter_tag")
				}
				if tag, _ := data["filter_tag"].(string); tag != "discovered" {
					t.Errorf("noise filter_tag = %q, want discovered", tag)
				}
			},
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

			result, err := handleTaskAnalysisNative(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleTaskAnalysisNative() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestHandleTaskAnalysis(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROJECT_ROOT", tmpDir)

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name: "duplicates action",
			params: map[string]interface{}{
				"action": "duplicates",
			},
			wantError: false,
		},
		{
			name: "tags action",
			params: map[string]interface{}{
				"action": "tags",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			argsJSON, _ := json.Marshal(tt.params)

			result, err := handleTaskAnalysis(ctx, argsJSON)
			if (err != nil) != tt.wantError {
				t.Errorf("handleTaskAnalysis() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && (result == nil || len(result) == 0) {
				t.Error("expected non-empty result")
			}
		})
	}
}

func TestBuildBatchTagPrompt(t *testing.T) {
	canonical := []string{"testing", "docs"}
	project := []string{"migration"}
	batch := []map[string]interface{}{
		{"file": "docs/A.md", "tags": []string{"a"}},
		{"file": "docs/B.md", "tags": []string{}},
	}

	prompt := buildBatchTagPrompt(batch, canonical, project)
	if prompt == "" {
		t.Fatal("expected non-empty prompt")
	}

	if !strings.Contains(prompt, "testing") || !strings.Contains(prompt, "docs") {
		t.Error("prompt should include canonical tags")
	}

	if !strings.Contains(prompt, "docs/A.md") || !strings.Contains(prompt, "docs/B.md") {
		t.Error("prompt should include file paths")
	}

	if !strings.Contains(prompt, "[\"a\"]") {
		t.Error("prompt should include existing tags as JSON")
	}

	if !strings.Contains(prompt, "{\"path\"") {
		t.Error("prompt should ask for JSON object")
	}
}

func TestSuggestNextLLMBatchSize(t *testing.T) {
	tests := []struct {
		name        string
		used        int
		timeouts    int
		discoveries int
		want        int
	}{
		{"no timeouts, try larger", 15, 0, 36, 25},
		{"no timeouts, cap at 50", 45, 0, 99, 50},
		{"timeouts, halve", 30, 1, 36, 25},            // 30/2=15, floor at llmTagBatchSize (25)
		{"timeouts, floor at default", 20, 1, 36, 25}, // 20/2=10, floor at 25
		{"no llm, default", 15, 0, 0, 25},             // default llmTagBatchSize is 25
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := suggestNextLLMBatchSize(tt.used, tt.timeouts, tt.discoveries)
			if got != tt.want {
				t.Errorf("suggestNextLLMBatchSize(%d, %d, %d) = %d, want %d", tt.used, tt.timeouts, tt.discoveries, got, tt.want)
			}
		})
	}
}

func TestParseBatchTagResponse(t *testing.T) {
	tests := []struct {
		name     string
		response string
		want     map[string][]string
	}{
		{
			name:     "valid JSON object",
			response: `{"docs/A.md": ["testing", "docs"], "docs/B.md": []}`,
			want:     map[string][]string{"docs/A.md": {"testing", "docs"}, "docs/B.md": {}},
		},
		{
			name:     "with surrounding text",
			response: `Here is the result: {"docs/X.md": ["migration"]} done.`,
			want:     map[string][]string{"docs/X.md": {"migration"}},
		},
		{
			name:     "empty object",
			response: `{}`,
			want:     map[string][]string{},
		},
		{
			name:     "no JSON object",
			response: `nothing here`,
			want:     nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseBatchTagResponse(tt.response)
			if tt.want == nil {
				if got != nil {
					t.Errorf("parseBatchTagResponse() = %v, want nil", got)
				}

				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("parseBatchTagResponse() keys count = %d, want %d", len(got), len(tt.want))
			}

			for path, wantTags := range tt.want {
				gotTags := got[path]
				if len(gotTags) != len(wantTags) {
					t.Errorf("path %q: got %v, want %v", path, gotTags, wantTags)
				}

				for i, w := range wantTags {
					if i >= len(gotTags) || gotTags[i] != w {
						t.Errorf("path %q: got %v, want %v", path, gotTags, wantTags)
						break
					}
				}
			}
		})
	}
}

func TestBuildTaskBatchTagPrompt(t *testing.T) {
	canonical := []string{"testing", "docs"}
	project := []string{"migration"}
	batch := []taskTagLLMBatchItem{
		{TaskID: "T-1", Title: "Add tests", Snippet: "Implement unit tests for handler.", ExistingTags: []string{"testing"}},
		{TaskID: "T-2", Title: "Fix bug", Snippet: "Error when parsing JSON.", ExistingTags: []string{}},
	}

	prompt := buildTaskBatchTagPrompt(batch, canonical, project)
	if prompt == "" {
		t.Fatal("expected non-empty prompt")
	}

	if !strings.Contains(prompt, "testing") || !strings.Contains(prompt, "docs") {
		t.Error("prompt should include canonical tags")
	}

	if !strings.Contains(prompt, "T-1") || !strings.Contains(prompt, "T-2") {
		t.Error("prompt should include task IDs")
	}

	if !strings.Contains(prompt, "Add tests") || !strings.Contains(prompt, "Fix bug") {
		t.Error("prompt should include task titles")
	}

	if !strings.Contains(prompt, "[\"testing\"]") {
		t.Error("prompt should include existing tags as JSON")
	}

	if !strings.Contains(prompt, "{\"T-123\"") {
		t.Error("prompt should ask for JSON object with task_id keys")
	}
}

func TestExtractOllamaGenerateResponseText(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "ollama wrapper",
			text: `{"success":true,"method":"native_go","response":"{\"T-1\":[\"testing\",\"bug\"]}","model":"tinyllama"}`,
			want: `{"T-1":["testing","bug"]}`,
		},
		{
			name: "plain text",
			text: `{"T-1": ["testing"]}`,
			want: `{"T-1": ["testing"]}`,
		},
		{
			name: "invalid json",
			text: `not json`,
			want: `not json`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractOllamaGenerateResponseText(tt.text)
			if got != tt.want {
				t.Errorf("extractOllamaGenerateResponseText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseTaskTagResponse(t *testing.T) {
	tests := []struct {
		name     string
		response string
		want     map[string][]string
	}{
		{
			name:     "valid task IDs",
			response: `{"T-1": ["testing", "bug"], "T-2": []}`,
			want:     map[string][]string{"T-1": {"testing", "bug"}, "T-2": {}},
		},
		{
			name:     "with surrounding text",
			response: `Result: {"T-42": ["docs", "migration"]} end`,
			want:     map[string][]string{"T-42": {"docs", "migration"}},
		},
		{
			name:     "empty object",
			response: `{}`,
			want:     map[string][]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTaskTagResponse(tt.response)
			if len(got) != len(tt.want) {
				t.Errorf("parseTaskTagResponse() keys count = %d, want %d", len(got), len(tt.want))
			}

			for taskID, wantTags := range tt.want {
				gotTags := got[taskID]
				if len(gotTags) != len(wantTags) {
					t.Errorf("task %q: got %v, want %v", taskID, gotTags, wantTags)
				}

				for i, w := range wantTags {
					if i >= len(gotTags) || gotTags[i] != w {
						t.Errorf("task %q: got %v, want %v", taskID, gotTags, wantTags)
						break
					}
				}
			}
		})
	}
}

func TestBuildTaskBatchTagPromptMatchOnly(t *testing.T) {
	allowed := []string{"testing", "docs", "bug", "feature"}
	batch := []taskTagLLMBatchItem{
		{TaskID: "T-1", Title: "Add tests", Snippet: "Unit tests for handler.", ExistingTags: nil},
	}

	prompt := buildTaskBatchTagPromptMatchOnly(batch, allowed)
	if prompt == "" {
		t.Fatal("expected non-empty prompt")
	}

	if !strings.Contains(prompt, "testing") || !strings.Contains(prompt, "docs") {
		t.Error("prompt should list allowed tags")
	}

	if !strings.Contains(prompt, "T-1") || !strings.Contains(prompt, "Add tests") {
		t.Error("prompt should include task ID and title")
	}

	if !strings.Contains(prompt, "Only tags from the list") {
		t.Error("prompt should constrain to allowed list")
	}
}

func TestDetectTaskOverlapConflicts(t *testing.T) {
	ptr := func(id, content, status string, deps []string) *database.Todo2Task {
		return &database.Todo2Task{ID: id, Content: content, Status: status, Dependencies: deps}
	}

	tests := []struct {
		name      string
		tasks     []*database.Todo2Task
		wantCount int
		wantPairs [][2]string // [dep_task_id, task_id] per conflict (DepTaskID blocks TaskID)
	}{
		{
			name:      "no conflicts - empty",
			tasks:     nil,
			wantCount: 0,
			wantPairs: nil,
		},
		{
			name: "no conflicts - single task",
			tasks: []*database.Todo2Task{
				ptr("T-1", "One", "In Progress", nil),
			},
			wantCount: 0,
			wantPairs: nil,
		},
		{
			name: "no conflicts - dep is Todo",
			tasks: []*database.Todo2Task{
				ptr("T-1", "One", "Todo", nil),
				ptr("T-2", "Two", "In Progress", []string{"T-1"}),
			},
			wantCount: 0,
			wantPairs: nil,
		},
		{
			name: "conflict - both In Progress with dep",
			tasks: []*database.Todo2Task{
				ptr("T-1", "One", "In Progress", nil),
				ptr("T-2", "Two", "In Progress", []string{"T-1"}),
			},
			wantCount: 1,
			wantPairs: [][2]string{{"T-1", "T-2"}},
		},
		{
			name: "conflict - chain",
			tasks: []*database.Todo2Task{
				ptr("T-1", "One", "In Progress", nil),
				ptr("T-2", "Two", "In Progress", []string{"T-1"}),
				ptr("T-3", "Three", "In Progress", []string{"T-2"}),
			},
			wantCount: 2,
			wantPairs: [][2]string{{"T-1", "T-2"}, {"T-2", "T-3"}},
		},
		{
			name: "conflict - multiple deps same blocker",
			tasks: []*database.Todo2Task{
				ptr("T-1", "One", "In Progress", nil),
				ptr("T-2", "Two", "In Progress", []string{"T-1"}),
				ptr("T-3", "Three", "In Progress", []string{"T-1"}),
			},
			wantCount: 2,
			wantPairs: [][2]string{{"T-1", "T-2"}, {"T-1", "T-3"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectTaskOverlapConflicts(tt.tasks)
			if len(got) != tt.wantCount {
				t.Errorf("DetectTaskOverlapConflicts() conflict count = %d, want %d", len(got), tt.wantCount)
			}

			for i, c := range got {
				if tt.wantPairs != nil && i < len(tt.wantPairs) {
					if c.DepTaskID != tt.wantPairs[i][0] || c.TaskID != tt.wantPairs[i][1] {
						t.Errorf("conflict[%d]: got DepTaskID=%s TaskID=%s, want %s blocks %s",
							i, c.DepTaskID, c.TaskID, tt.wantPairs[i][0], tt.wantPairs[i][1])
					}
				}

				if c.Reason == "" {
					t.Error("conflict reason should be non-empty")
				}
			}
		})
	}
}

func TestFilterSuggestionsToAllowed(t *testing.T) {
	allowed := map[string]bool{"testing": true, "docs": true, "bug": true}

	got := filterSuggestionsToAllowed(
		map[string][]string{
			"T-1": {"testing", "unknown", "docs"},
			"T-2": {"bug"},
			"T-3": {"hallucinated"},
		},
		allowed,
	)
	if len(got["T-1"]) != 2 || (got["T-1"][0] != "testing" && got["T-1"][1] != "testing") {
		t.Errorf("T-1: want [testing, docs], got %v", got["T-1"])
	}

	if len(got["T-2"]) != 1 || got["T-2"][0] != "bug" {
		t.Errorf("T-2: want [bug], got %v", got["T-2"])
	}

	if _, ok := got["T-3"]; ok {
		t.Error("T-3 should be omitted when all tags filtered out")
	}
}
