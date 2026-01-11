package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

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

// handleRecommendWorkflowNative handles the "workflow" action for recommend tool
// Recommends AGENT vs ASK mode based on task complexity
func handleRecommendWorkflowNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	// Get task description
	taskDescription := ""
	if descRaw, ok := params["task_description"].(string); ok {
		taskDescription = descRaw
	}

	// Get task ID if provided
	taskID := ""
	if idRaw, ok := params["task_id"].(string); ok {
		taskID = idRaw
	}

	// If task ID provided, try to load task description
	if taskID != "" && taskDescription == "" {
		// Try to load task from Todo2 database
		projectRoot, err := FindProjectRoot()
		if err == nil {
				tasks, err := LoadTodo2Tasks(projectRoot)
				if err == nil {
					for _, task := range tasks {
						if task.ID == taskID {
							taskDescription = task.Content + " " + task.LongDescription
							break
						}
					}
				}
		}
	}

	// Get tags if provided (JSON string)
	tagList := []string{}
	if tagsRaw, ok := params["tags"].(string); ok && tagsRaw != "" {
		var tags []interface{}
		if err := json.Unmarshal([]byte(tagsRaw), &tags); err == nil {
			for _, tag := range tags {
				if tagStr, ok := tag.(string); ok {
					tagList = append(tagList, tagStr)
				}
			}
		}
	}

	// Get include_rationale flag
	includeRationale := true
	if rationaleRaw, ok := params["include_rationale"].(bool); ok {
		includeRationale = rationaleRaw
	}

	// Score AGENT vs ASK mode
	agentScore, agentReasons := scoreAgentIndicators(taskDescription, tagList)
	askScore, askReasons := scoreAskIndicators(taskDescription, tagList)

	// Determine recommendation
	var mode string
	var confidence float64
	var reasons []string
	var description string

	if agentScore > askScore {
		mode = "AGENT"
		confidence = minFloat(float64(agentScore)/(float64(agentScore)+float64(askScore)+1)*100, 95)
		reasons = agentReasons
		description = "Use AGENT mode for autonomous multi-step implementation"
	} else if askScore > agentScore {
		mode = "ASK"
		confidence = minFloat(float64(askScore)/(float64(agentScore)+float64(askScore)+1)*100, 95)
		reasons = askReasons
		description = "Use ASK mode for focused questions and single edits"
	} else {
		mode = "ASK" // Default to ASK when uncertain
		confidence = 50
		reasons = []string{"No strong indicators - defaulting to ASK for safety"}
		description = "Unclear complexity - start with ASK, escalate to AGENT if needed"
	}

	// Build result
	result := map[string]interface{}{
		"recommended_mode": mode,
		"confidence":       roundFloat(confidence, 1),
		"description":      description,
		"agent_score":      agentScore,
		"ask_score":        askScore,
	}

	if includeRationale {
		// Top 5 reasons
		if len(reasons) > 5 {
			reasons = reasons[:5]
		}
		result["rationale"] = reasons
		result["guidelines"] = map[string]string{
			"AGENT": "Best for: Multi-file changes, feature implementation, refactoring, scaffolding",
			"ASK":   "Best for: Questions, code review, single-file edits, debugging help",
		}
	}

	// Add user-facing suggestion
	suggestion := getModeSuggestion(mode, confidence, taskDescription)
	result["suggestion"] = suggestion

	// Wrap in success response format
	response := map[string]interface{}{
		"success":   true,
		"data":      result,
		"timestamp": time.Now().Unix(),
	}

	resultJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: string(resultJSON)},
	}, nil
}

// AGENT indicators
var agentKeywords = []string{
	"implement", "create", "build", "develop", "refactor",
	"migrate", "upgrade", "integrate", "deploy", "configure",
	"setup", "install", "automate", "generate", "scaffold",
}

var agentPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)multi.?file`),
	regexp.MustCompile(`(?i)cross.?module`),
	regexp.MustCompile(`(?i)end.?to.?end`),
	regexp.MustCompile(`(?i)full.?stack`),
	regexp.MustCompile(`(?i)complete\s+\w+`),
}

var agentTags = []string{
	"feature", "implementation", "infrastructure", "integration",
	"refactoring", "migration", "automation",
}

// ASK indicators
var askKeywords = []string{
	"explain", "what", "why", "how", "understand",
	"clarify", "review", "check", "validate", "analyze",
	"debug", "find", "locate", "show", "list",
}

var askPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)single\s+file`),
	regexp.MustCompile(`(?i)quick\s+\w+`),
	regexp.MustCompile(`(?i)simple\s+\w+`),
	regexp.MustCompile(`(?i)just\s+\w+`),
}

var askTags = []string{
	"question", "documentation", "review", "analysis",
	"debugging", "research",
}

// scoreAgentIndicators scores AGENT mode indicators
func scoreAgentIndicators(content string, tags []string) (int, []string) {
	score := 0
	reasons := []string{}

	contentLower := strings.ToLower(content)

	// Score keywords
	for _, kw := range agentKeywords {
		if strings.Contains(contentLower, kw) {
			score += 2
			reasons = append(reasons, fmt.Sprintf("Keyword: '%s'", kw))
		}
	}

	// Score patterns
	for _, pattern := range agentPatterns {
		if pattern.MatchString(contentLower) {
			score += 3
			reasons = append(reasons, fmt.Sprintf("Pattern: '%s'", pattern.String()))
		}
	}

	// Score tags
	for _, tag := range tags {
		tagLower := strings.ToLower(tag)
		for _, agentTag := range agentTags {
			if tagLower == strings.ToLower(agentTag) {
				score += 2
				reasons = append(reasons, fmt.Sprintf("Tag: '%s'", tag))
				break
			}
		}
	}

	return score, reasons
}

// scoreAskIndicators scores ASK mode indicators
func scoreAskIndicators(content string, tags []string) (int, []string) {
	score := 0
	reasons := []string{}

	contentLower := strings.ToLower(content)

	// Score keywords
	for _, kw := range askKeywords {
		if strings.Contains(contentLower, kw) {
			score += 2
			reasons = append(reasons, fmt.Sprintf("Keyword: '%s'", kw))
		}
	}

	// Score patterns
	for _, pattern := range askPatterns {
		if pattern.MatchString(contentLower) {
			score += 3
			reasons = append(reasons, fmt.Sprintf("Pattern: '%s'", pattern.String()))
		}
	}

	// Score tags
	for _, tag := range tags {
		tagLower := strings.ToLower(tag)
		for _, askTag := range askTags {
			if tagLower == strings.ToLower(askTag) {
				score += 2
				reasons = append(reasons, fmt.Sprintf("Tag: '%s'", tag))
				break
			}
		}
	}

	return score, reasons
}

// getModeSuggestion generates user-facing suggestion message
func getModeSuggestion(mode string, confidence float64, taskSummary string) map[string]interface{} {
	modeSuggestions := map[string]map[string]interface{}{
		"AGENT": {
			"emoji":      "🤖",
			"action":     "Switch to AGENT mode",
			"instruction": "Click the mode selector (top of chat) → Select 'Agent'",
			"why":        "AGENT mode enables autonomous multi-file editing with automatic tool execution",
			"benefits": []string{
				"Autonomous file creation and modification",
				"Multi-step task execution",
				"Automatic tool calls without confirmation",
				"Better for implementation tasks",
			},
		},
		"ASK": {
			"emoji":      "💬",
			"action":     "Switch to ASK mode",
			"instruction": "Click the mode selector (top of chat) → Select 'Ask'",
			"why":        "ASK mode provides focused assistance with user control over changes",
			"benefits": []string{
				"User confirms each change",
				"Better for learning and understanding",
				"Safer for critical code review",
				"Ideal for questions and explanations",
			},
		},
	}

	suggestion := modeSuggestions[mode]
	if suggestion == nil {
		suggestion = modeSuggestions["ASK"]
	}

	// Build confidence qualifier
	var qualifier string
	if confidence >= 80 {
		qualifier = "strongly recommend"
	} else if confidence >= 60 {
		qualifier = "recommend"
	} else {
		qualifier = "suggest"
	}

	// Build message
	message := fmt.Sprintf("%s **Mode Suggestion: %s**\n\n", suggestion["emoji"], mode)
	if taskSummary != "" {
		summary := taskSummary
		if len(summary) > 50 {
			summary = summary[:50] + "..."
		}
		message += fmt.Sprintf("For this task (%s), I %s using **%s** mode.\n\n", summary, qualifier, mode)
	} else {
		message += fmt.Sprintf("I %s using **%s** mode for this task.\n\n", qualifier, mode)
	}
	message += fmt.Sprintf("**Why?** %s\n\n", suggestion["why"])
	message += "**To switch:**\n"
	message += fmt.Sprintf("→ %s", suggestion["instruction"])

	return map[string]interface{}{
		"message":     message,
		"action":      suggestion["action"],
		"instruction": suggestion["instruction"],
		"benefits":    sliceFirstN(suggestion["benefits"].([]string), 3),
	}
}

// Helper functions
func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func roundFloat(val float64, precision int) float64 {
	multiplier := 1.0
	for i := 0; i < precision; i++ {
		multiplier *= 10
	}
	return float64(int(val*multiplier+0.5)) / multiplier
}

func sliceFirstN(slice []string, n int) []string {
	if len(slice) <= n {
		return slice
	}
	return slice[:n]
}
