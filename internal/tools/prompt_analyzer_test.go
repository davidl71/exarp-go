package tools

import (
	"context"
	"testing"
)

func TestPromptAnalysis_MinScore(t *testing.T) {
	p := &PromptAnalysis{Clarity: 0.9, Specificity: 0.7, Completeness: 0.8, Structure: 0.85, Actionability: 0.6}
	if got := p.MinScore(); got != 0.6 {
		t.Errorf("MinScore() = %v, want 0.6", got)
	}
	p2 := &PromptAnalysis{Clarity: 0.8, Specificity: 0.8, Completeness: 0.8, Structure: 0.8, Actionability: 0.8}
	if got := p2.MinScore(); got != 0.8 {
		t.Errorf("MinScore() = %v, want 0.8", got)
	}
}

func TestParseSuggestionsResponse(t *testing.T) {
	body := `{"suggestions":[{"dimension":"specificity","issue":"vague","recommendation":"add file paths"}]}`
	resp, err := parseSuggestionsResponse(body)
	if err != nil {
		t.Fatalf("parseSuggestionsResponse: %v", err)
	}
	if len(resp.Suggestions) != 1 || resp.Suggestions[0].Dimension != "specificity" {
		t.Errorf("got %+v", resp)
	}
	// with markdown fence
	body2 := "```json\n" + body + "\n```"
	resp2, err := parseSuggestionsResponse(body2)
	if err != nil {
		t.Fatalf("parseSuggestionsResponse(markdown): %v", err)
	}
	if len(resp2.Suggestions) != 1 {
		t.Errorf("got %+v", resp2)
	}
}

func TestRefinePromptLoop_InvalidInput(t *testing.T) {
	ctx := context.Background()
	_, _, err := RefinePromptLoop(ctx, "", "ctx", "code", nil, nil)
	if err == nil {
		t.Error("expected error for empty prompt")
	}
	_, _, err = RefinePromptLoop(ctx, "hello", "ctx", "code", nil, nil)
	if err == nil {
		t.Error("expected error for nil generator")
	}
}

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
