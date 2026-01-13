package tools

import (
	"context"
	"testing"
)

// BenchmarkHandleSessionPrime benchmarks the session prime action
func BenchmarkHandleSessionPrime(b *testing.B) {
	ctx := context.Background()
	params := map[string]interface{}{
		"action": "prime",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = handleSessionNative(ctx, params)
	}
}

// BenchmarkHandleRecommendWorkflow benchmarks the recommend workflow action
func BenchmarkHandleRecommendWorkflow(b *testing.B) {
	ctx := context.Background()
	params := map[string]interface{}{
		"action":          "workflow",
		"task_description": "Implement a new feature with multiple components, database migrations, API endpoints, and frontend integration",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = handleRecommendWorkflowNative(ctx, params)
	}
}

// BenchmarkHandleHealthServer benchmarks the health server action
func BenchmarkHandleHealthServer(b *testing.B) {
	ctx := context.Background()
	params := map[string]interface{}{
		"action": "server",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = handleHealthNative(ctx, params)
	}
}

// BenchmarkHandleSetupPatternTriggers benchmarks the setup_hooks patterns action
func BenchmarkHandleSetupPatternTriggers(b *testing.B) {
	ctx := context.Background()
	params := map[string]interface{}{
		"action":  "patterns",
		"dry_run": true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = handleSetupPatternTriggers(ctx, params)
	}
}

// BenchmarkHandleAlignmentPRD benchmarks the analyze_alignment prd action
func BenchmarkHandleAlignmentPRD(b *testing.B) {
	ctx := context.Background()
	params := map[string]interface{}{
		"action": "prd",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = handleAlignmentPRD(ctx, params)
	}
}

// BenchmarkHandleEstimationAnalyze benchmarks the estimation analyze action
func BenchmarkHandleEstimationAnalyze(b *testing.B) {
	params := map[string]interface{}{
		"detailed": true,
	}

	// Use a temporary directory for testing
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = handleEstimationAnalyze("/tmp", params)
	}
}

// BenchmarkToolInvocationChain benchmarks chaining multiple tool invocations
func BenchmarkToolInvocationChain(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate a common workflow: prime -> recommend -> health
		_, _ = handleSessionNative(ctx, map[string]interface{}{"action": "prime"})
		_, _ = handleRecommendWorkflowNative(ctx, map[string]interface{}{
			"action":          "workflow",
			"task_description": "Test task",
		})
		_, _ = handleHealthNative(ctx, map[string]interface{}{"action": "server"})
	}
}
