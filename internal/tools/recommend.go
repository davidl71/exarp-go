package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/davidl71/exarp-go/internal/framework"
)

// ModelInfo represents information about an AI model
type ModelInfo struct {
	ModelID   string   `json:"model_id"`
	Name      string   `json:"name"`
	BestFor   []string `json:"best_for"`
	TaskTypes []string `json:"task_types"`
	Cost      string   `json:"cost"`
	Speed     string   `json:"speed"`
}

// MODEL_CATALOG contains the static catalog of available AI models
var MODEL_CATALOG = []ModelInfo{
	{
		ModelID:   "claude-sonnet",
		Name:      "Claude Sonnet 4",
		BestFor:   []string{"Complex multi-file implementations", "Architecture decisions", "Code review with nuanced feedback", "Long context comprehension", "Reasoning-heavy tasks"},
		TaskTypes: []string{"architecture", "review", "analysis", "complex_implementation"},
		Cost:      "higher",
		Speed:     "moderate",
	},
	{
		ModelID:   "claude-haiku",
		Name:      "Claude Haiku",
		BestFor:   []string{"Quick code completions", "Simple bug fixes", "Syntax corrections", "Fast iterations", "Cost-sensitive workflows"},
		TaskTypes: []string{"quick_fix", "completion", "formatting", "simple_edit"},
		Cost:      "low",
		Speed:     "fast",
	},
	{
		ModelID:   "gpt-4o",
		Name:      "GPT-4o",
		BestFor:   []string{"General coding tasks", "API integrations", "Multi-modal tasks with images", "Quick prototyping"},
		TaskTypes: []string{"general", "integration", "prototyping", "multimodal"},
		Cost:      "moderate",
		Speed:     "fast",
	},
	{
		ModelID:   "o1-preview",
		Name:      "o1-preview",
		BestFor:   []string{"Mathematical reasoning", "Algorithm design", "Complex problem solving", "Scientific computing"},
		TaskTypes: []string{"math", "algorithm", "reasoning", "scientific"},
		Cost:      "highest",
		Speed:     "slow",
	},
	{
		ModelID:   "gemini-pro",
		Name:      "Gemini Pro",
		BestFor:   []string{"Large codebase analysis", "Very long context windows", "Cross-file understanding"},
		TaskTypes: []string{"large_context", "codebase_analysis"},
		Cost:      "moderate",
		Speed:     "moderate",
	},
	{
		ModelID:   "ollama-llama3.2",
		Name:      "Ollama Llama 3.2",
		BestFor:   []string{"Local development without API costs", "Privacy-sensitive tasks", "Offline development", "General coding tasks", "Quick prototyping"},
		TaskTypes: []string{"general", "local", "privacy", "offline"},
		Cost:      "free",
		Speed:     "moderate",
	},
	{
		ModelID:   "ollama-mistral",
		Name:      "Ollama Mistral",
		BestFor:   []string{"Fast local inference", "Code generation", "Quick iterations", "Cost-free development"},
		TaskTypes: []string{"code_generation", "quick_fix", "local"},
		Cost:      "free",
		Speed:     "fast",
	},
	{
		ModelID:   "ollama-codellama",
		Name:      "Ollama CodeLlama",
		BestFor:   []string{"Code-specific tasks", "Code completion", "Code explanation", "Local code analysis"},
		TaskTypes: []string{"code_analysis", "code_generation", "local"},
		Cost:      "free",
		Speed:     "moderate",
	},
	{
		ModelID:   "ollama-phi3",
		Name:      "Ollama Phi-3",
		BestFor:   []string{"Small model for quick tasks", "Resource-constrained environments", "Fast local inference", "Simple code edits"},
		TaskTypes: []string{"quick_fix", "simple_edit", "local"},
		Cost:      "free",
		Speed:     "fast",
	},
}

// handleListModels handles the list_models tool
// Lists all available AI models with capabilities
func handleListModels(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Build result
	result := map[string]interface{}{
		"models": MODEL_CATALOG,
		"count":  len(MODEL_CATALOG),
		"tip":    "Use recommend_model for task-specific recommendations",
	}

	// Wrap in success response format
	response := map[string]interface{}{
		"success":   true,
		"data":      result,
		"timestamp": 0, // Will be set by Python bridge if needed, but this is native Go
	}

	resultJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: string(resultJSON)},
	}, nil
}

