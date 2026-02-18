package tools

import (
	"context"
	"testing"
)

type benchmarkAnalysisGenerator struct {
	analysisResp   string
	suggestionResp string
	callCount      int
}

func (m *benchmarkAnalysisGenerator) Supported() bool { return true }

func (m *benchmarkAnalysisGenerator) Generate(_ context.Context, _ string, _ int, _ float32) (string, error) {
	m.callCount++
	if m.callCount%2 == 1 {
		return m.analysisResp, nil
	}
	return m.suggestionResp, nil
}

func BenchmarkRefinePromptLoop(b *testing.B) {
	ctx := context.Background()

	analysisJSON := `{"clarity":0.95,"specificity":0.9,"completeness":0.92,"structure":0.88,"actionability":0.91,"summary":"Strong prompt","notes":[]}`
	suggestionJSON := `{"suggestions":[{"dimension":"specificity","issue":"minor","recommendation":"add detail"}]}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen := &benchmarkAnalysisGenerator{
			analysisResp:   analysisJSON,
			suggestionResp: suggestionJSON,
		}
		_, _, err := RefinePromptLoop(ctx, "Implement a REST API for user management", "backend", "code", gen, &RefinePromptLoopOptions{
			MaxIterations:     1,
			MinScoreThreshold: 0.8,
		})
		if err != nil {
			b.Fatalf("RefinePromptLoop: %v", err)
		}
	}
}
