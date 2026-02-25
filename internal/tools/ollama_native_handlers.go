// ollama_native_handlers.go — Ollama native: pull, hardware info, docs, quality check, and summary handlers.
// See also: ollama_native.go
package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/mcp-go-core/pkg/mcp/response"
)

// ─── Contents ───────────────────────────────────────────────────────────────
//   handleOllamaPull
//   handleOllamaHardware — handleOllamaHardware returns hardware info and recommendations.
//   handleOllamaDocs — handleOllamaDocs generates code documentation using Ollama.
//   handleOllamaQuality — handleOllamaQuality analyzes code quality using Ollama.
//   handleOllamaSummary — handleOllamaSummary enhances context summaries using Ollama.
// ────────────────────────────────────────────────────────────────────────────

// ─── handleOllamaPull ───────────────────────────────────────────────────────
func handleOllamaPull(ctx context.Context, params map[string]interface{}, host string) ([]framework.TextContent, error) {
	model := getOllamaModelParam(params, "")
	if model == "" {
		return nil, fmt.Errorf("model parameter is required for pull action (Ollama API uses 'model'; 'name' is accepted as alias)")
	}

	// Create pull request (Ollama API expects "model", not "name")
	reqBody := map[string]interface{}{
		"model": model,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	timeout := config.OllamaDownloadTimeout()
	if timeout <= 0 {
		timeout = 300 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	url := fmt.Sprintf("%s/api/pull", host)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := ollamaHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Ollama API: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// best effort close
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Read streaming progress (pull returns progress messages)
	decoder := json.NewDecoder(resp.Body)

	var lastStatus string

	for decoder.More() {
		var progress map[string]interface{}
		if err := decoder.Decode(&progress); err != nil {
			break
		}

		if status, ok := progress["status"].(string); ok {
			lastStatus = status
		}

		if completed, ok := progress["completed_at"].(string); ok && completed != "" {
			break
		}
	}

	result := map[string]interface{}{
		"success": true,
		"method":  "native_go",
		"model":   model,
		"status":  lastStatus,
		"message": fmt.Sprintf("Model %s pull initiated. Check Ollama logs for progress.", model),
	}

	return response.FormatResult(result, "")
}

// ─── handleOllamaHardware ───────────────────────────────────────────────────
// handleOllamaHardware returns hardware info and recommendations.
func handleOllamaHardware(ctx context.Context) ([]framework.TextContent, error) {
	// Simple hardware detection (can be enhanced)
	result := map[string]interface{}{
		"success": true,
		"method":  "native_go",
		"message": "Hardware detection not yet implemented in native Go. Use Python bridge for detailed hardware info.",
		"tip":     "Set OLLAMA_NUM_GPU and OLLAMA_NUM_THREADS environment variables for optimization",
	}

	return response.FormatResult(result, "")
}

// ─── handleOllamaDocs ───────────────────────────────────────────────────────
// handleOllamaDocs generates code documentation using Ollama.
func handleOllamaDocs(ctx context.Context, params map[string]interface{}, host string) ([]framework.TextContent, error) {
	filePath, _ := params["file_path"].(string)
	if filePath == "" {
		return nil, fmt.Errorf("file_path parameter required for docs action")
	}

	outputPath, _ := params["output_path"].(string)

	style, _ := params["style"].(string)
	if style == "" {
		style = "google"
	}

	model := getOllamaModelParam(params, "codellama")

	// Read file
	code, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Build documentation prompt
	prompt := fmt.Sprintf(`Generate comprehensive documentation for this code.
Use %s docstring style.

Requirements:
1. Module-level docstring explaining the file's purpose
2. Function/class docstrings with:
   - Clear description
   - Parameters (Args section)
   - Returns section
   - Raises section (if applicable)
   - Examples section (if helpful)
3. Inline comments for complex logic
4. Type hints where appropriate

Code:
%s

Generate the documented version of this code.`, style, string(code))

	// Use generate action
	generateParams := map[string]interface{}{
		"prompt": prompt,
		"model":  model,
		"stream": false,
	}

	result, err := handleOllamaGenerate(ctx, generateParams, host)
	if err != nil {
		return nil, fmt.Errorf("failed to generate documentation: %w", err)
	}

	// Extract response text
	var responseText string

	if len(result) > 0 && result[0].Type == "text" {
		var generateResult map[string]interface{}
		if err := json.Unmarshal([]byte(result[0].Text), &generateResult); err == nil {
			if resp, ok := generateResult["response"].(string); ok {
				responseText = resp
			}
		}
	}

	// Save to output file if specified
	if outputPath != "" && responseText != "" {
		if err := os.WriteFile(outputPath, []byte(responseText), 0644); err != nil {
			return nil, fmt.Errorf("failed to write output file: %w", err)
		}
	}

	// Format result
	docResult := map[string]interface{}{
		"success":           true,
		"method":            "native_go",
		"file_path":         filePath,
		"output_path":       outputPath,
		"style":             style,
		"documentation":     responseText,
		"original_length":   len(code),
		"documented_length": len(responseText),
	}

	return response.FormatResult(docResult, "")
}

// ─── handleOllamaQuality ────────────────────────────────────────────────────
// handleOllamaQuality analyzes code quality using Ollama.
func handleOllamaQuality(ctx context.Context, params map[string]interface{}, host string) ([]framework.TextContent, error) {
	filePath, _ := params["file_path"].(string)
	if filePath == "" {
		return nil, fmt.Errorf("file_path parameter required for quality action")
	}

	includeSuggestions := true
	if suggestions, ok := params["include_suggestions"].(bool); ok {
		includeSuggestions = suggestions
	}

	model := getOllamaModelParam(params, "codellama")

	// Read file
	code, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Build quality analysis prompt
	suggestionsText := ""
	if includeSuggestions {
		suggestionsText = "7. Specific refactoring suggestions"
	}

	prompt := fmt.Sprintf(`Analyze this code for quality and provide a structured assessment.

Provide:
1. Overall quality score (0-100) with brief justification
2. Code smells detected (list specific issues)
3. Performance issues (if any)
4. Security concerns (if any)
5. Best practice violations
6. Code maintainability assessment
%s

Format your response as JSON with these keys:
- quality_score (number)
- code_smells (array of strings)
- performance_issues (array of strings)
- security_concerns (array of strings)
- best_practice_violations (array of strings)
- maintainability (string: "excellent" | "good" | "fair" | "poor")
- suggestions (array of strings, if include_suggestions is true)

Code:
%s`, suggestionsText, string(code))

	// Use generate action
	generateParams := map[string]interface{}{
		"prompt": prompt,
		"model":  model,
		"stream": false,
	}

	result, err := handleOllamaGenerate(ctx, generateParams, host)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze code quality: %w", err)
	}

	// Extract response text
	var responseText string

	if len(result) > 0 && result[0].Type == "text" {
		var generateResult map[string]interface{}
		if err := json.Unmarshal([]byte(result[0].Text), &generateResult); err == nil {
			if resp, ok := generateResult["response"].(string); ok {
				responseText = resp
			}
		}
	}

	// Try to parse JSON from response
	var qualityData map[string]interface{}

	if responseText != "" {
		jsonText := ExtractJSONObjectFromLLMResponse(responseText)
		if err := json.Unmarshal([]byte(jsonText), &qualityData); err != nil {
			// If parsing fails, use raw response
			qualityData = map[string]interface{}{
				"raw_analysis": responseText,
			}
		}
	}

	// Format result
	qualityResult := map[string]interface{}{
		"success":   true,
		"method":    "native_go",
		"file_path": filePath,
		"analysis":  qualityData,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	return response.FormatResult(qualityResult, "")
}

// ─── handleOllamaSummary ────────────────────────────────────────────────────
// handleOllamaSummary enhances context summaries using Ollama.
func handleOllamaSummary(ctx context.Context, params map[string]interface{}, host string) ([]framework.TextContent, error) {
	dataRaw := params["data"]
	if dataRaw == nil {
		return nil, fmt.Errorf("data parameter required for summary action")
	}

	level, _ := params["level"].(string)
	if level == "" {
		level = "brief"
	}

	model := getOllamaModelParam(params, "codellama")

	// Convert data to string if needed
	var dataStr string

	switch v := dataRaw.(type) {
	case string:
		dataStr = v
	default:
		dataJSON, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal data: %w", err)
		}

		dataStr = string(dataJSON)
	}

	// Try to parse as JSON to get structured summary
	var dataObj interface{}
	if err := json.Unmarshal([]byte(dataStr), &dataObj); err != nil {
		// If not JSON, use as-is
		dataObj = dataStr
	}

	// Build summary prompt based on level
	var prompt string

	switch level {
	case "brief":
		prompt = fmt.Sprintf(`Summarize this data in one brief sentence highlighting key points.

Data:
%s

Provide a concise summary (max 100 words).`, dataStr)
	case "detailed":
		prompt = fmt.Sprintf(`Provide a detailed summary of this data with key insights and metrics.

Data:
%s

Provide a comprehensive summary with:
- Key metrics
- Important findings
- Notable patterns or trends
- Actionable insights`, dataStr)
	case "key_metrics":
		prompt = fmt.Sprintf(`Extract and summarize only the key metrics from this data as a bulleted list.

Data:
%s

List only the numerical metrics and their values.`, dataStr)
	case "actionable":
		prompt = fmt.Sprintf(`Summarize this data focusing on actionable recommendations and next steps.

Data:
%s

Provide:
- Key actions to take
- Priorities
- Recommendations`, dataStr)
	default:
		prompt = fmt.Sprintf(`Summarize this data at a %s level.

Data:
%s

Provide a structured summary.`, level, dataStr)
	}

	// Use generate action
	generateParams := map[string]interface{}{
		"prompt": prompt,
		"model":  model,
		"stream": false,
	}

	result, err := handleOllamaGenerate(ctx, generateParams, host)
	if err != nil {
		return nil, fmt.Errorf("failed to generate summary: %w", err)
	}

	// Extract response text
	var responseText string

	if len(result) > 0 && result[0].Type == "text" {
		var generateResult map[string]interface{}
		if err := json.Unmarshal([]byte(result[0].Text), &generateResult); err == nil {
			if resp, ok := generateResult["response"].(string); ok {
				responseText = resp
			}
		}
	}

	// Format result
	summaryResult := map[string]interface{}{
		"success":         true,
		"method":          "native_go",
		"level":           level,
		"summary":         responseText,
		"original_length": len(dataStr),
		"summary_length":  len(responseText),
		"timestamp":       time.Now().Format(time.RFC3339),
	}

	return response.FormatResult(summaryResult, "")
}
