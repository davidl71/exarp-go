package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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

// handleRecommendModelNative handles the "model" action for recommend tool
func handleRecommendModelNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	// Get task description
	taskDescription := ""
	if descRaw, ok := params["task_description"].(string); ok {
		taskDescription = descRaw
	}

	// Get task type if provided
	taskType := ""
	if typeRaw, ok := params["task_type"].(string); ok {
		taskType = typeRaw
	}

	// Get optimization target
	optimizeFor := "quality"
	if optimizeRaw, ok := params["optimize_for"].(string); ok {
		optimizeFor = optimizeRaw
	}

	includeAlternatives := true
	if altRaw, ok := params["include_alternatives"].(bool); ok {
		includeAlternatives = altRaw
	}

	// Find best matching model
	recommended := findBestModel(taskDescription, taskType, optimizeFor)

	// Build result
	result := map[string]interface{}{
		"recommended_model": recommended,
		"task_description":  taskDescription,
		"optimize_for":      optimizeFor,
		"rationale":         fmt.Sprintf("Selected %s based on task characteristics and optimization for %s", recommended.ModelID, optimizeFor),
	}

	if includeAlternatives {
		alternatives := findAlternativeModels(recommended, optimizeFor)
		result["alternatives"] = alternatives
	}

	// Wrap in success response format
	response := map[string]interface{}{
		"success":   true,
		"data":      result,
		"timestamp": 0,
	}

	resultJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: string(resultJSON)},
	}, nil
}

// findBestModel finds the best model for a given task
func findBestModel(taskDescription, taskType, optimizeFor string) ModelInfo {
	taskLower := strings.ToLower(taskDescription + " " + taskType)

	// Score each model
	bestModel := MODEL_CATALOG[0]
	bestScore := 0.0

	for _, model := range MODEL_CATALOG {
		score := 0.0

		// Check task type match
		for _, tt := range model.TaskTypes {
			if strings.Contains(taskLower, strings.ToLower(tt)) {
				score += 10.0
			}
		}

		// Check keywords in task description
		for _, keyword := range []string{"quick", "simple", "fast", "complex", "architecture", "review"} {
			if strings.Contains(taskLower, keyword) {
				for _, bestFor := range model.BestFor {
					if strings.Contains(strings.ToLower(bestFor), keyword) {
						score += 5.0
					}
				}
			}
		}

		// Apply optimization preference
		switch optimizeFor {
		case "speed":
			if model.Speed == "fast" {
				score += 20.0
			} else if model.Speed == "moderate" {
				score += 10.0
			}
		case "cost":
			if model.Cost == "free" {
				score += 20.0
			} else if model.Cost == "low" {
				score += 15.0
			} else if model.Cost == "moderate" {
				score += 10.0
			}
		case "quality":
			// Quality is default, no bonus needed
		}

		if score > bestScore {
			bestScore = score
			bestModel = model
		}
	}

	return bestModel
}

// findAlternativeModels finds alternative models
func findAlternativeModels(recommended ModelInfo, optimizeFor string) []ModelInfo {
	alternatives := []ModelInfo{}

	// Find 2-3 alternatives with different characteristics
	for _, model := range MODEL_CATALOG {
		if model.ModelID == recommended.ModelID {
			continue
		}

		// Prefer models with different cost/speed profiles
		if model.Cost != recommended.Cost || model.Speed != recommended.Speed {
			alternatives = append(alternatives, model)
			if len(alternatives) >= 3 {
				break
			}
		}
	}

	return alternatives
}
