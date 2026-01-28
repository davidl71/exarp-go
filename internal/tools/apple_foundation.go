//go:build darwin && arm64 && cgo
// +build darwin,arm64,cgo

package tools

import (
	"context"
	"encoding/json"
	"fmt"

	fm "github.com/blacktop/go-foundationmodels"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/platform"
)

// handleAppleFoundationModels handles the apple_foundation_models tool
func handleAppleFoundationModels(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Check platform support first
	support := platform.CheckAppleFoundationModelsSupport()
	if !support.Supported {
		return []framework.TextContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Apple Foundation Models not supported: %s", support.Reason),
			},
		}, nil
	}

	// Parse arguments
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Get action (optional, default: "generate")
	action := "generate"
	if actionStr, ok := params["action"].(string); ok && actionStr != "" {
		action = actionStr
	}

	// Handle actions that don't require a prompt
	switch action {
	case "status":
		return handleStatusAction()
	case "hardware":
		return handleHardwareAction()
	case "models":
		return handleModelsAction()
	}

	// For other actions, prompt is required
	prompt, ok := params["prompt"].(string)
	if !ok || prompt == "" {
		return nil, fmt.Errorf("prompt is required for action: %s", action)
	}

	// Create session
	sess := fm.NewSession()
	defer sess.Release()

	var result string
	var err error

	switch action {
	case "generate", "respond":
		// Generate text response
		result, err = generateText(sess, prompt, params)
	case "summarize":
		// Summarize text
		result, err = summarizeText(sess, prompt, params)
	case "classify":
		// Classify text
		result, err = classifyText(sess, prompt, params)
	default:
		return nil, fmt.Errorf("unknown action: %s (supported: status, hardware, models, generate, respond, summarize, classify)", action)
	}

	if err != nil {
		return nil, fmt.Errorf("apple foundation models failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// GenerateWithOptions runs Apple FM with the given prompt and options.
// Used by apple_foundation_models tool and by FMProvider (task_analysis hierarchy, etc.).
func GenerateWithOptions(prompt string, maxTokens int, temperature float32) (string, error) {
	sess := fm.NewSession()
	defer sess.Release()
	return sess.RespondWithOptions(prompt, maxTokens, temperature), nil
}

// generateText generates text using Apple Foundation Models
func generateText(sess *fm.Session, prompt string, params map[string]interface{}) (string, error) {
	maxTokens := getMaxTokens(params)
	temperature := getTemperature(params)
	return sess.RespondWithOptions(prompt, maxTokens, temperature), nil
}

// summarizeText summarizes text using Apple Foundation Models
func summarizeText(sess *fm.Session, text string, params map[string]interface{}) (string, error) {
	// Create summarization prompt
	prompt := fmt.Sprintf("Summarize the following text concisely:\n\n%s", text)

	// Use lower temperature for summarization (more deterministic)
	temp := float32(0.3)
	if t := getTemperature(params); t > 0 {
		temp = t
	}

	// Generate summary
	response := sess.RespondWithOptions(prompt, getMaxTokens(params), temp)
	return response, nil
}

// classifyText classifies text using Apple Foundation Models
func classifyText(sess *fm.Session, text string, params map[string]interface{}) (string, error) {
	// Get categories (optional)
	categories := "positive, negative, neutral"
	if cats, ok := params["categories"].(string); ok && cats != "" {
		categories = cats
	}

	// Create classification prompt
	prompt := fmt.Sprintf("Classify the following text into one of these categories: %s\n\nText: %s\n\nRespond with only the category name.", categories, text)

	// Use lower temperature for classification (more deterministic)
	temp := float32(0.2)
	if t := getTemperature(params); t > 0 {
		temp = t
	}

	// Generate classification
	response := sess.RespondWithOptions(prompt, getMaxTokens(params), temp)
	return response, nil
}

// buildGenerationOptions builds generation options from parameters (for future use)
func buildGenerationOptions(params map[string]interface{}) *fm.GenerationOptions {
	options := &fm.GenerationOptions{}

	// Temperature
	if temp, ok := params["temperature"].(float64); ok {
		temp32 := float32(temp)
		options.Temperature = &temp32
	}

	// Max tokens
	if maxTokens, ok := params["max_tokens"].(float64); ok {
		maxTokensInt := int(maxTokens)
		options.MaxTokens = &maxTokensInt
	}

	return options
}
