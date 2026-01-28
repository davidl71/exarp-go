// Package tools: static recommended MLX models list for action=models (no CGO/luxfi required).
package tools

// mlxRecommendedModels is the curated list of recommended MLX models from Hugging Face mlx-community.
// Matches the Python bridge list_mlx_models() structure for compatibility.
var mlxRecommendedModels = map[string][]map[string]string{
	"code": {
		{"id": "mlx-community/CodeLlama-7b-mlx", "size": "7B", "best_for": "Code analysis, documentation, code generation", "size_gb": "~4GB"},
		{"id": "mlx-community/CodeLlama-7b-Python-mlx", "size": "7B", "best_for": "Python-focused code generation", "size_gb": "~4GB"},
	},
	"general": {
		{"id": "mlx-community/Meta-Llama-3.1-8B-Instruct-bf16", "size": "8B", "best_for": "General purpose, high quality responses (full precision)", "size_gb": "~16GB"},
		{"id": "mlx-community/Meta-Llama-3.1-8B-Instruct-4bit", "size": "8B", "best_for": "General purpose, high quality responses (quantized)", "size_gb": "~4.5GB"},
		{"id": "mlx-community/Phi-3.5-mini-instruct-4bit", "size": "3.8B", "best_for": "Fast, efficient general tasks", "size_gb": "~2.3GB"},
		{"id": "mlx-community/Mistral-7B-Instruct-v0.2", "size": "7B", "best_for": "Fast inference, good balance", "size_gb": "~4GB"},
	},
	"small": {
		{"id": "mlx-community/TinyLlama-1.1B-Chat-v1.0-mlx", "size": "1.1B", "best_for": "Very fast tasks, low memory", "size_gb": "~0.7GB"},
		{"id": "mlx-community/Phi-3-mini-128k-instruct-4bit", "size": "3.8B", "best_for": "Fast with long context (128k)", "size_gb": "~2.3GB"},
	},
}

// mlxModelsResponse returns the result map for mlx action=models (native Go, no bridge).
// Used by both mlx_native.go (darwin+cgo) and mlx_native_nocgo.go.
func mlxModelsResponse() map[string]interface{} {
	total := 0
	for _, list := range mlxRecommendedModels {
		total += len(list)
	}
	return map[string]interface{}{
		"success":     true,
		"method":      "native_go",
		"models":      mlxRecommendedModels,
		"total_count": total,
		"note":        "Models are downloaded from Hugging Face on first use",
		"tip":         "Use generate action to generate text with a model",
	}
}
