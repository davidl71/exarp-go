package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/davidl71/devwisdom-go/pkg/wisdom"
	"github.com/davidl71/exarp-go/internal/database"
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

	includeRationale := true
	if rationaleRaw, ok := params["include_rationale"].(bool); ok {
		includeRationale = rationaleRaw
	}

	// If task_id provided, load task from database
	if taskID != "" && taskDescription == "" {
		task, err := database.GetTask(ctx, taskID)
		if err == nil {
			taskDescription = task.Content + " " + task.LongDescription
		}
	}

	// Analyze task and recommend workflow mode
	recommendation := analyzeWorkflowMode(taskDescription, includeRationale)

	// Build result
	result := map[string]interface{}{
		"recommended_mode": recommendation.Mode,
		"confidence":       recommendation.Confidence,
		"description":      recommendation.Description,
		"agent_score":      recommendation.AgentScore,
		"ask_score":        recommendation.AskScore,
	}

	if includeRationale {
		result["rationale"] = recommendation.Rationale
		result["guidelines"] = map[string]string{
			"AGENT": "Best for: Multi-file changes, feature implementation, refactoring, scaffolding",
			"ASK":   "Best for: Questions, code review, single-file edits, debugging help",
		}
		result["suggestion"] = recommendation.Suggestion
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

// WorkflowRecommendation represents a workflow mode recommendation
type WorkflowRecommendation struct {
	Mode        string
	Confidence  float64
	Description string
	AgentScore  int
	AskScore    int
	Rationale   []string
	Suggestion  map[string]interface{}
}

// analyzeWorkflowMode analyzes a task and recommends AGENT or ASK mode
func analyzeWorkflowMode(taskDescription string, includeRationale bool) WorkflowRecommendation {
	taskLower := strings.ToLower(taskDescription)

	// AGENT indicators
	agentKeywords := []string{
		"implement", "create", "add", "build", "refactor", "migrate",
		"multi-file", "multiple files", "architecture", "scaffold",
		"feature", "system", "framework", "restructure",
	}
	agentPatterns := []string{
		`\b(implement|create|add|build|refactor|migrate)\s+\w+`,
		`\b(multi|multiple)\s+\w*\s*file`,
		`\b(architecture|system|framework)\s+\w+`,
	}

	// ASK indicators
	askKeywords := []string{
		"question", "how", "what", "why", "explain", "review",
		"debug", "fix", "error", "bug", "single file", "one file",
		"help", "clarify", "understand",
	}
	askPatterns := []string{
		`\b(how|what|why)\s+\w+`,
		`\b(explain|review|debug|fix)\s+\w+`,
		`\b(single|one)\s+\w*\s*file`,
	}

	// Score AGENT indicators
	agentScore := 0
	agentReasons := []string{}

	for _, kw := range agentKeywords {
		if strings.Contains(taskLower, kw) {
			agentScore += 2
			agentReasons = append(agentReasons, fmt.Sprintf("Keyword: '%s'", kw))
		}
	}

	// Simple pattern matching (basic regex)
	for _, pattern := range agentPatterns {
		if strings.Contains(taskLower, strings.TrimPrefix(pattern, `\b(`)) {
			agentScore += 3
			agentReasons = append(agentReasons, fmt.Sprintf("Pattern: '%s'", pattern))
		}
	}

	// Score ASK indicators
	askScore := 0
	askReasons := []string{}

	for _, kw := range askKeywords {
		if strings.Contains(taskLower, kw) {
			askScore += 2
			askReasons = append(askReasons, fmt.Sprintf("Keyword: '%s'", kw))
		}
	}

	// Simple pattern matching
	for _, pattern := range askPatterns {
		if strings.Contains(taskLower, strings.TrimPrefix(pattern, `\b(`)) {
			askScore += 3
			askReasons = append(askReasons, fmt.Sprintf("Pattern: '%s'", pattern))
		}
	}

	// Determine recommendation
	var mode string
	var confidence float64
	var description string
	var reasons []string

	if agentScore > askScore {
		mode = "AGENT"
		totalScore := agentScore + askScore
		if totalScore > 0 {
			confidence = float64(agentScore) / float64(totalScore) * 100
		} else {
			confidence = 50
		}
		if confidence > 95 {
			confidence = 95
		}
		description = "Use AGENT mode for autonomous multi-step implementation"
		reasons = agentReasons
	} else if askScore > agentScore {
		mode = "ASK"
		totalScore := agentScore + askScore
		if totalScore > 0 {
			confidence = float64(askScore) / float64(totalScore) * 100
		} else {
			confidence = 50
		}
		if confidence > 95 {
			confidence = 95
		}
		description = "Use ASK mode for focused questions and single edits"
		reasons = askReasons
	} else {
		mode = "ASK" // Default to ASK when uncertain
		confidence = 50
		description = "Unclear complexity - start with ASK, escalate to AGENT if needed"
		reasons = []string{"No strong indicators - defaulting to ASK for safety"}
	}

	// Build suggestion
	suggestion := map[string]interface{}{
		"message": fmt.Sprintf("Recommended: %s mode (%.1f%% confidence)", mode, confidence),
		"action": map[string]string{
			"AGENT": "Switch to AGENT mode for autonomous implementation",
			"ASK":   "Stay in ASK mode for guided assistance",
		}[mode],
		"instruction": map[string]string{
			"AGENT": "Enable AGENT mode in Cursor settings",
			"ASK":   "Continue with ASK mode for this task",
		}[mode],
		"benefits": map[string][]string{
			"AGENT": {"Autonomous multi-step execution", "Handles complex refactoring", "Manages multiple files"},
			"ASK":   {"Focused assistance", "Quick answers", "Single-file edits"},
		}[mode],
	}

	return WorkflowRecommendation{
		Mode:        mode,
		Confidence:  confidence,
		Description: description,
		AgentScore:  agentScore,
		AskScore:    askScore,
		Rationale:   reasons[:minInt(5, len(reasons))], // Top 5 reasons
		Suggestion:  suggestion,
	}
}

// handleRecommendAdvisorNative handles the "advisor" action for recommend tool
// Uses devwisdom-go wisdom engine directly (no MCP client needed)
func handleRecommendAdvisorNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	// Extract parameters
	var metric, tool, stage, context string
	var score float64

	if m, ok := params["metric"].(string); ok {
		metric = m
	}
	if t, ok := params["tool"].(string); ok {
		tool = t
	}
	if st, ok := params["stage"].(string); ok {
		stage = st
	}
	if c, ok := params["context"].(string); ok {
		context = c
	}
	if sc, ok := params["score"].(float64); ok {
		score = sc
	} else if sc, ok := params["score"].(int); ok {
		score = float64(sc)
	}
	// Validate and clamp score to 0-100 range
	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}

	// Get wisdom engine
	engine, err := getWisdomEngine()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize wisdom engine: %w", err)
	}

	// Determine advisor based on metric, tool, or stage
	var advisorInfo *wisdom.AdvisorInfo
	if metric != "" {
		advisorInfo, err = engine.GetAdvisors().GetAdvisorForMetric(metric)
	} else if tool != "" {
		advisorInfo, err = engine.GetAdvisors().GetAdvisorForTool(tool)
	} else if stage != "" {
		advisorInfo, err = engine.GetAdvisors().GetAdvisorForStage(stage)
	} else {
		// Default advisor
		advisorInfo = &wisdom.AdvisorInfo{
			Advisor:   "pistis_sophia",
			Icon:      "📜",
			Rationale: "Default wisdom advisor",
		}
	}

	if err != nil || advisorInfo == nil {
		// Fallback to default
		advisorInfo = &wisdom.AdvisorInfo{
			Advisor:   "pistis_sophia",
			Icon:      "📜",
			Rationale: "Default wisdom advisor",
		}
	}

	// Get wisdom quote
	quote, err := engine.GetWisdom(score, advisorInfo.Advisor)
	if err != nil {
		// Fallback quote
		quote = &wisdom.Quote{
			Quote:         "Wisdom comes from experience.",
			Source:        "Unknown",
			Encouragement: "Keep learning and growing.",
		}
	}

	// Get consultation mode based on score
	modeConfig := wisdom.GetConsultationMode(score)

	// Create consultation (use map since Consultation type not exported in public API)
	consultation := map[string]interface{}{
		"timestamp":         time.Now().Format(time.RFC3339),
		"consultation_type": "advisor",
		"advisor":           advisorInfo.Advisor,
		"advisor_icon":      advisorInfo.Icon,
		"advisor_name":      advisorInfo.Advisor,
		"rationale":         advisorInfo.Rationale,
		"score_at_time":     score,
		"consultation_mode": modeConfig.Name,
		"mode_icon":         modeConfig.Icon,
		"mode_frequency":    modeConfig.Frequency,
		"mode_guidance":     modeConfig.Description,
		"quote":             quote.Quote,
		"quote_source":      quote.Source,
		"encouragement":     quote.Encouragement,
		"context":           context,
	}

	// Convert to JSON
	resultJSON, err := json.MarshalIndent(consultation, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal consultation: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: string(resultJSON)},
	}, nil
}

// minInt returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
