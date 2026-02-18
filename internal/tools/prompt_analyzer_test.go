package tools

import (
	"context"
	"testing"
)

// mockAnalysisGenerator returns fixed JSON so AnalyzePrompt can be tested without a real LLM.
type mockAnalysisGenerator struct {
	response string
}

func (m *mockAnalysisGenerator) Supported() bool { return true }

func (m *mockAnalysisGenerator) Generate(_ context.Context, _ string, _ int, _ float32) (string, error) {
	return m.response, nil
}

func TestAnalyzePrompt_WithMockGenerator(t *testing.T) {
	ctx := context.Background()

	jsonResp := `{"clarity":0.9,"specificity":0.75,"completeness":0.8,"structure":0.85,"actionability":0.7,"summary":"Good prompt","notes":[]}`
	gen := &mockAnalysisGenerator{response: jsonResp}

	result, err := AnalyzePrompt(ctx, "Add tests for module X.", "", "code", gen)
	if err != nil {
		t.Fatalf("AnalyzePrompt with mock: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil PromptAnalysis")
	}
	if result.Clarity != 0.9 || result.Specificity != 0.75 {
		t.Errorf("Clarity=%f Specificity=%f", result.Clarity, result.Specificity)
	}
	if result.Summary != "Good prompt" {
		t.Errorf("Summary = %q", result.Summary)
	}
}

func TestAnalyzePrompt_EmptyPromptOrNilGenerator(t *testing.T) {
	ctx := context.Background()

	_, err := AnalyzePrompt(ctx, "", "", "", &mockAnalysisGenerator{response: "{}"})
	if err == nil {
		t.Error("expected error for empty prompt")
	}

	_, err = AnalyzePrompt(ctx, "ok", "", "", nil)
	if err == nil {
		t.Error("expected error for nil generator")
	}
}

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

func TestGenerateSuggestions_WithMockGenerator(t *testing.T) {
	ctx := context.Background()

	suggestionsJSON := `{"suggestions":[{"dimension":"clarity","issue":"ambiguous wording","recommendation":"use precise terms"},{"dimension":"completeness","issue":"missing context","recommendation":"add file paths"}]}`
	gen := &mockAnalysisGenerator{response: suggestionsJSON}

	analysis := &PromptAnalysis{
		Clarity:       0.5,
		Specificity:   0.6,
		Completeness:  0.4,
		Structure:     0.7,
		Actionability: 0.6,
		Summary:       "Needs improvement",
	}

	result, err := GenerateSuggestions(ctx, "Refactor the module.", analysis, "", "code", gen)
	if err != nil {
		t.Fatalf("GenerateSuggestions: %v", err)
	}

	if len(result.Suggestions) != 2 {
		t.Fatalf("got %d suggestions, want 2", len(result.Suggestions))
	}
	if result.Suggestions[0].Dimension != "clarity" {
		t.Errorf("Suggestions[0].Dimension = %q, want clarity", result.Suggestions[0].Dimension)
	}
	if result.Suggestions[1].Recommendation != "add file paths" {
		t.Errorf("Suggestions[1].Recommendation = %q", result.Suggestions[1].Recommendation)
	}
}

func TestRefinePrompt_WithMockGenerator(t *testing.T) {
	ctx := context.Background()

	gen := &mockAnalysisGenerator{response: "Refactor the auth module in internal/auth by extracting token validation into a helper function."}

	suggestions := &SuggestionsResponse{
		Suggestions: []SuggestionItem{
			{Dimension: "specificity", Issue: "vague", Recommendation: "name the module and file"},
		},
	}

	refined, err := RefinePrompt(ctx, "Refactor the module.", suggestions, "", "code", gen)
	if err != nil {
		t.Fatalf("RefinePrompt: %v", err)
	}
	if refined == "" {
		t.Fatal("expected non-empty refined prompt")
	}
	if refined != "Refactor the auth module in internal/auth by extracting token validation into a helper function." {
		t.Errorf("refined = %q", refined)
	}
}

// mockLoopGenerator cycles through responses to simulate multiple iterations of the refinement loop.
type mockLoopGenerator struct {
	responses []string
	callIndex int
}

func (m *mockLoopGenerator) Supported() bool { return true }

func (m *mockLoopGenerator) Generate(_ context.Context, _ string, _ int, _ float32) (string, error) {
	if m.callIndex >= len(m.responses) {
		return m.responses[len(m.responses)-1], nil
	}
	resp := m.responses[m.callIndex]
	m.callIndex++
	return resp, nil
}

func TestRefinePromptLoop_SuccessfulIteration(t *testing.T) {
	ctx := context.Background()

	// The loop calls: AnalyzePrompt (1 Generate) → check MinScore → stop if above threshold.
	// Return analysis with all scores >= 0.9 so it stops on the first iteration.
	highScoreAnalysis := `{"clarity":0.95,"specificity":0.92,"completeness":0.90,"structure":0.93,"actionability":0.91,"summary":"Excellent prompt","notes":[]}`

	gen := &mockLoopGenerator{
		responses: []string{highScoreAnalysis},
	}

	finalPrompt, lastAnalysis, err := RefinePromptLoop(ctx, "Add unit tests for auth module.", "", "code", gen, &RefinePromptLoopOptions{
		MaxIterations:     3,
		MinScoreThreshold: 0.9,
	})
	if err != nil {
		t.Fatalf("RefinePromptLoop: %v", err)
	}

	if finalPrompt != "Add unit tests for auth module." {
		t.Errorf("finalPrompt = %q, want original (already above threshold)", finalPrompt)
	}
	if lastAnalysis == nil {
		t.Fatal("expected non-nil lastAnalysis")
	}
	if lastAnalysis.MinScore() < 0.9 {
		t.Errorf("MinScore() = %v, want >= 0.9", lastAnalysis.MinScore())
	}
	if lastAnalysis.Summary != "Excellent prompt" {
		t.Errorf("Summary = %q, want 'Excellent prompt'", lastAnalysis.Summary)
	}
}
