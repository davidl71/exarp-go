package tools

import (
	"reflect"
	"testing"
)

func TestExtractTagsFromText(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		wantTags  []string
		wantClean string
	}{
		{
			name:      "single tag",
			text:      "#refactor fix this function",
			wantTags:  []string{"refactor"},
			wantClean: "fix this function",
		},
		{
			name:      "multiple tags",
			text:      "#refactor #performance Optimize this function",
			wantTags:  []string{"refactor", "performance"},
			wantClean: "Optimize this function",
		},
		{
			name:      "tags at end",
			text:      "Fix the bug here #bug #critical",
			wantTags:  []string{"bug", "critical"},
			wantClean: "Fix the bug here",
		},
		{
			name:      "mixed position tags",
			text:      "#security validate #important user input",
			wantTags:  []string{"security", "important"},
			wantClean: "validate user input",
		},
		{
			name:      "no tags",
			text:      "Regular todo without tags",
			wantTags:  []string{},
			wantClean: "Regular todo without tags",
		},
		{
			name:      "duplicate tags",
			text:      "#bug fix #bug here",
			wantTags:  []string{"bug"},
			wantClean: "fix here",
		},
		{
			name:      "tag with hyphen",
			text:      "#high-priority fix this",
			wantTags:  []string{"high-priority"},
			wantClean: "fix this",
		},
		{
			name:      "tag with underscore",
			text:      "#code_review needed",
			wantTags:  []string{"code_review"},
			wantClean: "needed",
		},
		{
			name:      "tag with numbers",
			text:      "#v2 upgrade needed",
			wantTags:  []string{"v2"},
			wantClean: "upgrade needed",
		},
		{
			name:      "not a tag - number first",
			text:      "#123 is not a tag",
			wantTags:  []string{},
			wantClean: "#123 is not a tag",
		},
		{
			name:      "case normalization",
			text:      "#BUG #Bug same tag",
			wantTags:  []string{"bug"},
			wantClean: "same tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTags, gotClean := extractTagsFromText(tt.text)

			// Handle nil vs empty slice
			if len(gotTags) == 0 && len(tt.wantTags) == 0 {
				// Both empty, OK
			} else if !reflect.DeepEqual(gotTags, tt.wantTags) {
				t.Errorf("extractTagsFromText() tags = %v, want %v", gotTags, tt.wantTags)
			}

			if gotClean != tt.wantClean {
				t.Errorf("extractTagsFromText() clean = %q, want %q", gotClean, tt.wantClean)
			}
		})
	}
}
