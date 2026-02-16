package tools

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestExtractJSONArrayFromLLMResponse(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "plain json array",
			input:  `[{"task_id":"T-1","level":"task"}]`,
			expect: `[{"task_id":"T-1","level":"task"}]`,
		},
		{
			name:   "markdown code block json",
			input:  "Here is the result:\n```json\n[{\"task_id\":\"T-1\"}]\n```",
			expect: `[{"task_id":"T-1"}]`,
		},
		{
			name:   "leading plain text then array",
			input:  "Priority: high\n\n[{\"task_id\":\"T-1\",\"level\":\"task\"}]",
			expect: `[{"task_id":"T-1","level":"task"}]`,
		},
		{
			name:   "code block without json label",
			input:  "```\n[{\"task_id\":\"T-2\"}]\n```",
			expect: `[{"task_id":"T-2"}]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractJSONArrayFromLLMResponse(tt.input)
			if tt.expect != "" && !strings.Contains(got, tt.expect) {
				t.Errorf("ExtractJSONArrayFromLLMResponse() = %q, want to contain %q", got, tt.expect)
			}

			var arr []map[string]interface{}
			if err := json.Unmarshal([]byte(got), &arr); err != nil {
				t.Errorf("extracted string not valid JSON: %v; got %q", err, got)
			}
		})
	}
}

func TestExtractJSONObjectFromLLMResponse(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "plain json object",
			input:  `{"quality": 8, "reason": "good"}`,
			expect: `{"quality": 8, "reason": "good"}`,
		},
		{
			name:   "markdown code block",
			input:  "Analysis:\n```json\n{\"score\": 7}\n```",
			expect: `{"score": 7}`,
		},
		{
			name:   "leading text then object",
			input:  "Here is the analysis.\n\n{\"quality\": 9}",
			expect: `{"quality": 9}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractJSONObjectFromLLMResponse(tt.input)
			if tt.expect != "" && !strings.Contains(got, tt.expect) {
				t.Errorf("ExtractJSONObjectFromLLMResponse() = %q, want to contain %q", got, tt.expect)
			}

			var m map[string]interface{}
			if err := json.Unmarshal([]byte(got), &m); err != nil {
				t.Errorf("extracted string not valid JSON: %v; got %q", err, got)
			}
		})
	}
}
