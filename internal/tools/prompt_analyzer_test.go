package tools

import (
	"testing"
)

func TestParsePromptAnalysis(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantErr bool
		check   func(*testing.T, *PromptAnalysis)
	}{
		{
			name: "valid JSON",
			text: `{"clarity":0.85,"specificity":0.6,"completeness":0.7,"structure":0.8,"actionability":0.5,"summary":"Good","notes":[]}`,
			check: func(t *testing.T, r *PromptAnalysis) {
				if r.Clarity != 0.85 {
					t.Errorf("Clarity = %v, want 0.85", r.Clarity)
				}
				if r.Specificity != 0.6 {
					t.Errorf("Specificity = %v, want 0.6", r.Specificity)
				}
			},
		},
		{
			name: "fallback plain text",
			text: "Clarity: 0.9\nSpecificity: 0.7\nCompleteness: 0.8\nStructure: 0.85\nActionability: 0.6",
			check: func(t *testing.T, r *PromptAnalysis) {
				if r.Clarity != 0.9 {
					t.Errorf("Clarity = %v, want 0.9", r.Clarity)
				}
				if r.Specificity != 0.7 {
					t.Errorf("Specificity = %v, want 0.7", r.Specificity)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePromptAnalysis(tt.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePromptAnalysis() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil && tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}
