package tools

import (
	"context"
	"testing"
)

type benchmarkModelRouter struct {
	response string
}

func (m *benchmarkModelRouter) SelectModel(_ string, _ ModelRequirements) ModelType {
	return ModelFM
}

func (m *benchmarkModelRouter) Generate(_ context.Context, _ ModelType, _ string, _ int, _ float32) (string, error) {
	return m.response, nil
}

func BenchmarkDefaultModelRouterGenerate(b *testing.B) {
	ctx := context.Background()
	router := &benchmarkModelRouter{response: `{"result":"benchmark output"}`}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = router.Generate(ctx, ModelFM, "benchmark prompt", 256, 0.7)
	}
}
