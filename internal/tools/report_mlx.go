package tools

import (
	"context"
	"fmt"
	"strings"
)

// enhanceReportWithMLX enhances report data with AI-generated insights.
// Uses DefaultReportInsight (tries MLX via bridge, then DefaultFMProvider() when MLX unavailable).
// Report code does not depend on the bridge or MLX by name.
func enhanceReportWithMLX(ctx context.Context, reportData map[string]interface{}, action string) (map[string]interface{}, error) {
	// Build prompt based on action
	var prompt string

	var maxTokens int

	switch action {
	case "scorecard":
		prompt = buildScorecardInsightPrompt(reportData)
		maxTokens = 1000 // Longer for detailed insights
	case "overview":
		prompt = buildOverviewInsightPrompt(reportData)
		maxTokens = 1500 // Even longer for comprehensive overview
	case "briefing":
		// Briefing uses devwisdom-go, skip insight enhancement
		return reportData, nil
	case "prd":
		prompt = buildPRDInsightPrompt(reportData)
		maxTokens = 2000 // Longest for PRD generation
	default:
		return reportData, nil
	}

	provider := DefaultReportInsight()
	if provider == nil || !provider.Supported() {
		return reportData, nil
	}

	generatedText, err := provider.Generate(ctx, prompt, maxTokens, 0.4)
	if err != nil || generatedText == "" {
		// If provider fails or returns empty, return original data without enhancement
		return reportData, nil
	}

	// Add AI insights to report data
	enhancedData := make(map[string]interface{})
	for k, v := range reportData {
		enhancedData[k] = v
	}

	enhancedData["ai_insights"] = map[string]interface{}{
		"generated_by": "insight_provider",
		"insights":     generatedText,
		"method":       "insight_enhanced",
	}

	return enhancedData, nil
}

// buildScorecardInsightPrompt builds prompt for scorecard insights.
func buildScorecardInsightPrompt(data map[string]interface{}) string {
	// Extract key metrics
	scores := map[string]interface{}{}
	if s, ok := data["scores"].(map[string]interface{}); ok {
		scores = s
	}

	overallScore := 0.0
	if score, ok := data["overall_score"].(float64); ok {
		overallScore = score
	}

	blockers := []string{}

	if b, ok := data["blockers"].([]interface{}); ok {
		for _, blocker := range b {
			if str, ok := blocker.(string); ok {
				blockers = append(blockers, str)
			}
		}
	}

	// Build prompt
	prompt := fmt.Sprintf(`Analyze this project health scorecard and generate intelligent insights:

Overall Score: %.1f/100

Component Scores:
%s

Blockers: %s

Generate a comprehensive analysis with:
1. Key Strengths - What's working well
2. Critical Issues - What needs immediate attention
3. Recommendations - Specific, actionable improvements
4. Trends - Patterns and progress indicators
5. Priority Actions - Top 3 things to focus on next

Format as clear, actionable insights suitable for stakeholders.`,
		overallScore,
		formatScoresForPrompt(scores),
		strings.Join(blockers, ", "))

	return prompt
}

// buildOverviewInsightPrompt builds prompt for overview insights.
func buildOverviewInsightPrompt(data map[string]interface{}) string {
	// Extract key sections
	projectInfo := map[string]interface{}{}
	if p, ok := data["project"].(map[string]interface{}); ok {
		projectInfo = p
	}

	health := map[string]interface{}{}
	if h, ok := data["health"].(map[string]interface{}); ok {
		health = h
	}

	tasks := map[string]interface{}{}
	if t, ok := data["tasks"].(map[string]interface{}); ok {
		tasks = t
	}

	risks := []string{}

	if r, ok := data["risks"].([]interface{}); ok {
		for _, risk := range r {
			if str, ok := risk.(string); ok {
				risks = append(risks, str)
			}
		}
	}

	prompt := fmt.Sprintf(`Generate an executive summary and insights for this project overview:

Project: %s
Status: %s

Health Score: %.1f/100

Task Status:
%s

Risks & Blockers:
%s

Generate:
1. Executive Summary - One paragraph overview for stakeholders
2. Key Achievements - What's been accomplished
3. Current Challenges - What's blocking progress
4. Strategic Recommendations - High-level guidance
5. Next Steps - Prioritized action items

Write in a professional, stakeholder-friendly tone.`,
		getString(projectInfo, "name"),
		getString(projectInfo, "status"),
		getFloat(health, "overall_score"),
		formatTasksForPrompt(tasks),
		strings.Join(risks, "\n- "))

	return prompt
}

// buildPRDInsightPrompt builds prompt for PRD generation.
func buildPRDInsightPrompt(data map[string]interface{}) string {
	// PRD generation would need project context
	// For now, return a basic prompt
	return `Generate a comprehensive Product Requirements Document (PRD) based on the project data provided. Include sections for: Overview, Goals, Features, Technical Requirements, Success Metrics, and Timeline.`
}

// Helper functions.
func formatScoresForPrompt(scores map[string]interface{}) string {
	var parts []string

	for key, value := range scores {
		if score, ok := value.(float64); ok {
			parts = append(parts, fmt.Sprintf("- %s: %.1f/100", key, score))
		}
	}

	return strings.Join(parts, "\n")
}

func formatTasksForPrompt(tasks map[string]interface{}) string {
	var parts []string
	if total, ok := tasks["total"].(float64); ok {
		parts = append(parts, fmt.Sprintf("Total: %.0f", total))
	}

	if pending, ok := tasks["pending"].(float64); ok {
		parts = append(parts, fmt.Sprintf("Pending: %.0f", pending))
	}

	if completed, ok := tasks["completed"].(float64); ok {
		parts = append(parts, fmt.Sprintf("Completed: %.0f", completed))
	}

	return strings.Join(parts, ", ")
}

func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}

	return "Unknown"
}

func getFloat(m map[string]interface{}, key string) float64 {
	if val, ok := m[key].(float64); ok {
		return val
	}

	return 0.0
}
