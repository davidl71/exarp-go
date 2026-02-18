package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/mcp-go-core/pkg/mcp/response"
)

// ollamaHTTPClient is a shared HTTP client for Ollama API calls. Reused across
// requests to enable connection pooling (Keep-Alive) instead of creating a new
// client per call. No Client.Timeout; callers use context.WithTimeout per request.
var ollamaHTTPClient = &http.Client{}

// OllamaModel represents an Ollama model.
type OllamaModel struct {
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	ModifiedAt string `json:"modified_at"`
	Digest     string `json:"digest"`
}

// OllamaGenerateRequest represents the request for text generation.
type OllamaGenerateRequest struct {
	Model    string                 `json:"model"`
	Prompt   string                 `json:"prompt"`
	Stream   bool                   `json:"stream,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
	System   string                 `json:"system,omitempty"`
	Template string                 `json:"template,omitempty"`
	Context  []int                  `json:"context,omitempty"`
}

// OllamaGenerateResponse represents the response from generation.
type OllamaGenerateResponse struct {
	Model              string `json:"model"`
	CreatedAt          string `json:"created_at"`
	Response           string `json:"response"`
	Done               bool   `json:"done"`
	Context            []int  `json:"context,omitempty"`
	TotalDuration      int64  `json:"total_duration,omitempty"`
	LoadDuration       int64  `json:"load_duration,omitempty"`
	PromptEvalCount    int    `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64  `json:"prompt_eval_duration,omitempty"`
	EvalCount          int    `json:"eval_count,omitempty"`
	EvalDuration       int64  `json:"eval_duration,omitempty"`
}

// getOllamaModelParam returns the model name from params. Ollama API uses "model"; accept "name" as alias for callers that use it (T-53).
func getOllamaModelParam(params map[string]interface{}, defaultVal string) string {
	if m, ok := params["model"].(string); ok && m != "" {
		return m
	}

	if n, ok := params["name"].(string); ok && n != "" {
		return n
	}

	return defaultVal
}

// handleOllamaNative handles the ollama tool with native Go HTTP client.
func handleOllamaNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	action, _ := params["action"].(string)
	if action == "" {
		action = "status"
	}

	host := "http://localhost:11434"
	if h, ok := params["host"].(string); ok && h != "" {
		host = h
	}

	switch action {
	case "status":
		return handleOllamaStatus(ctx, host)
	case "models":
		return handleOllamaModels(ctx, host)
	case "generate":
		return handleOllamaGenerate(ctx, params, host)
	case "pull":
		return handleOllamaPull(ctx, params, host)
	case "hardware":
		return handleOllamaHardware(ctx)
	case "docs":
		return handleOllamaDocs(ctx, params, host)
	case "quality":
		return handleOllamaQuality(ctx, params, host)
	case "summary":
		return handleOllamaSummary(ctx, params, host)
	default:
		return nil, fmt.Errorf("unknown action: %s (use 'status', 'models', 'generate', 'pull', 'hardware', 'docs', 'quality', or 'summary')", action)
	}
}

// ollamaAvailable returns true if the Ollama server at host is reachable (quick GET /api/tags).
func ollamaAvailable(ctx context.Context, host string) bool {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/api/tags", host)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false
	}

	resp, err := ollamaHTTPClient.Do(req)
	if err != nil {
		return false
	}

	_ = resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// ollamaGenerateText performs non-streaming generate and returns only the response text.
// Used by ollamaTextGenerator (TextGenerator) for FM-style generate.
func ollamaGenerateText(ctx context.Context, prompt string, maxTokens int, temperature float32, host, model string) (string, error) {
	timeout := config.OllamaGenerateTimeout()
	if timeout <= 0 {
		timeout = 120 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	options := map[string]interface{}{
		"num_predict": maxTokens,
		"temperature": float64(temperature),
	}
	reqBody := OllamaGenerateRequest{
		Model:   model,
		Prompt:  prompt,
		Stream:  false,
		Options: options,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/generate", host)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := ollamaHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call Ollama API: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// best effort close
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama API status %d: %s", resp.StatusCode, string(body))
	}

	var genResp OllamaGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return genResp.Response, nil
}

// handleOllamaStatus checks if Ollama server is running.
func handleOllamaStatus(ctx context.Context, host string) ([]framework.TextContent, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/api/tags", host)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := ollamaHTTPClient.Do(req)
	if err != nil {
		result := map[string]interface{}{
			"status": "error",
			"host":   host,
			"error":  "Ollama server not running. Start it with: ollama serve",
		}

		return response.FormatResult(result, "")
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// best effort close
		}
	}()

	if resp.StatusCode != http.StatusOK {
		result := map[string]interface{}{
			"status": "error",
			"host":   host,
			"error":  fmt.Sprintf("Ollama server returned status %d", resp.StatusCode),
		}

		return response.FormatResult(result, "")
	}

	var tagsResp struct {
		Models []struct {
			Name       string `json:"name"`
			ModifiedAt string `json:"modified_at"`
			Size       int64  `json:"size"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	modelNames := []string{}

	for i, model := range tagsResp.Models {
		if i < 10 { // First 10
			modelNames = append(modelNames, model.Name)
		}
	}

	result := map[string]interface{}{
		"success":     true,
		"method":      "native_go",
		"status":      "running",
		"host":        host,
		"model_count": len(tagsResp.Models),
		"models":      modelNames,
	}

	return response.FormatResult(result, "")
}

// handleOllamaModels lists available models.
func handleOllamaModels(ctx context.Context, host string) ([]framework.TextContent, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/api/tags", host)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := ollamaHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Ollama server not running. Start it with: ollama serve: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// best effort close
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama server returned status %d", resp.StatusCode)
	}

	var tagsResp struct {
		Models []struct {
			Name       string    `json:"name"`
			ModifiedAt time.Time `json:"modified_at"`
			Size       int64     `json:"size"`
			Digest     string    `json:"digest"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	formattedModels := []OllamaModel{}

	for _, model := range tagsResp.Models {
		digest := model.Digest
		if len(digest) > 12 {
			digest = digest[:12]
		}

		formattedModels = append(formattedModels, OllamaModel{
			Name:       model.Name,
			Size:       model.Size,
			ModifiedAt: model.ModifiedAt.Format(time.RFC3339),
			Digest:     digest,
		})
	}

	result := map[string]interface{}{
		"success": true,
		"method":  "native_go",
		"models":  formattedModels,
		"count":   len(formattedModels),
		"tip":     "Use generate action to generate text with a model",
	}

	return response.FormatResult(result, "")
}

// handleOllamaGenerate generates text using Ollama.
func handleOllamaGenerate(ctx context.Context, params map[string]interface{}, host string) ([]framework.TextContent, error) {
	prompt, _ := params["prompt"].(string)
	if prompt == "" {
		return nil, fmt.Errorf("prompt parameter is required for generate action")
	}

	model := getOllamaModelParam(params, "llama3.2")

	stream := false
	if s, ok := params["stream"].(bool); ok {
		stream = s
	}

	// Build options
	options := make(map[string]interface{})

	// Parse num_gpu
	if numGPU, ok := params["num_gpu"].(float64); ok {
		options["num_gpu"] = int(numGPU)
	} else if envGPU := os.Getenv("OLLAMA_NUM_GPU"); envGPU != "" {
		if gpu, err := strconv.Atoi(envGPU); err == nil {
			options["num_gpu"] = gpu
		}
	}

	// Parse num_threads
	if numThreads, ok := params["num_threads"].(float64); ok {
		options["num_threads"] = int(numThreads)
	} else if envThreads := os.Getenv("OLLAMA_NUM_THREADS"); envThreads != "" {
		if threads, err := strconv.Atoi(envThreads); err == nil {
			options["num_threads"] = threads
		}
	}

	// Parse context_size
	if contextSize, ok := params["context_size"].(float64); ok {
		options["num_ctx"] = int(contextSize)
	} else if envCtx := os.Getenv("OLLAMA_NUM_CTX"); envCtx != "" {
		if ctx, err := strconv.Atoi(envCtx); err == nil {
			options["num_ctx"] = ctx
		}
	}

	// Parse options JSON string if provided
	if optionsStr, ok := params["options"].(string); ok && optionsStr != "" {
		var customOptions map[string]interface{}
		if err := json.Unmarshal([]byte(optionsStr), &customOptions); err == nil {
			// Merge custom options
			for k, v := range customOptions {
				options[k] = v
			}
		}
	}

	// Create request
	reqBody := OllamaGenerateRequest{
		Model:  model,
		Prompt: prompt,
		Stream: stream,
	}

	if len(options) > 0 {
		reqBody.Options = options
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	timeout := config.OllamaGenerateTimeout()
	if timeout <= 0 {
		timeout = 120 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	url := fmt.Sprintf("%s/api/generate", host)

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

	// Handle streaming vs non-streaming
	if stream {
		// For streaming, read line by line
		decoder := json.NewDecoder(resp.Body)

		var fullResponse string

		for decoder.More() {
			var chunk OllamaGenerateResponse
			if err := decoder.Decode(&chunk); err != nil {
				break
			}

			fullResponse += chunk.Response

			if chunk.Done {
				break
			}
		}

		result := map[string]interface{}{
			"success":  true,
			"method":   "native_go",
			"response": fullResponse,
			"model":    model,
			"streamed": true,
		}

		return response.FormatResult(result, "")
	} else {
		// Non-streaming: read single response
		var generateResp OllamaGenerateResponse
		if err := json.NewDecoder(resp.Body).Decode(&generateResp); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		result := map[string]interface{}{
			"success":  true,
			"method":   "native_go",
			"response": generateResp.Response,
			"model":    generateResp.Model,
			"streamed": false,
		}

		if generateResp.TotalDuration > 0 {
			result["total_duration_ms"] = generateResp.TotalDuration / 1000000 // Convert nanoseconds to milliseconds
		}

		return response.FormatResult(result, "")
	}
}

// handleOllamaPull pulls/downloads a model.
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
