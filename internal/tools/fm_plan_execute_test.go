package tools

import (
	"testing"
)

func TestParseSubtasks(t *testing.T) {
	tests := []struct {
		name    string
		planOut string
		max     int
		wantMin int
	}{
		{
			name:    "json array",
			planOut: `["First step", "Second step", "Third step"]`,
			max:     10,
			wantMin: 3,
		},
		{
			name:    "json array with markdown fence",
			planOut: "```json\n[\"A\", \"B\"]\n```",
			max:     10,
			wantMin: 2,
		},
		{
			name: "numbered lines",
			planOut: `1. Research the topic
2. Write outline
3. Draft sections`,
			max:     10,
			wantMin: 3,
		},
		{
			name:    "empty",
			planOut: "",
			max:     5,
			wantMin: 0,
		},
		{
			name:    "json array capped by max",
			planOut: `["a","b","c","d","e","f"]`,
			max:     3,
			wantMin: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSubtasks(tt.planOut, tt.max)
			if len(got) < tt.wantMin {
				t.Errorf("parseSubtasks() len = %v, want at least %v", len(got), tt.wantMin)
			}
			if tt.max > 0 && len(got) > tt.max {
				t.Errorf("parseSubtasks() len = %v, want at most %v", len(got), tt.max)
			}
		})
	}
}
