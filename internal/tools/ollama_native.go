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
)

// OllamaModel represents an Ollama model
type OllamaModel struct {
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	ModifiedAt string `json:"modified_at"`
	Digest     string `json:"digest"`
}

// OllamaGenerateRequest represents the request for text generation
type OllamaGenerateRequest struct {
	Model    string                 `json:"model"`
	Prompt   string                 `json:"prompt"`
	Stream   bool                   `json:"stream,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
	System   string                 `json:"system,omitempty"`
	Template string                 `json:"template,omitempty"`
	Context  []int                  `json:"context,omitempty"`
}

// OllamaGenerateResponse represents the response from generation
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

// handleOllamaNative handles the ollama tool with native Go HTTP client
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
	default:
		return nil, fmt.Errorf("unknown action: %s (use 'status', 'models', 'generate', 'pull', or 'hardware')", action)
	}
}

// handleOllamaStatus checks if Ollama server is running
func handleOllamaStatus(ctx context.Context, host string) ([]framework.TextContent, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf("%s/api/tags", host)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		result := map[string]interface{}{
			"status": "error",
			"host":   host,
			"error":  "Ollama server not running. Start it with: ollama serve",
		}
		output, _ := json.MarshalIndent(result, "", "  ")
		return []framework.TextContent{
			{Type: "text", Text: string(output)},
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		result := map[string]interface{}{
			"status": "error",
			"host":   host,
			"error":  fmt.Sprintf("Ollama server returned status %d", resp.StatusCode),
		}
		output, _ := json.MarshalIndent(result, "", "  ")
		return []framework.TextContent{
			{Type: "text", Text: string(output)},
		}, nil
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

	output, _ := json.MarshalIndent(result, "", "  ")
	return []framework.TextContent{
		{Type: "text", Text: string(output)},
	}, nil
}

// handleOllamaModels lists available models
func handleOllamaModels(ctx context.Context, host string) ([]framework.TextContent, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("%s/api/tags", host)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Ollama server not running. Start it with: ollama serve: %w", err)
	}
	defer resp.Body.Close()

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

	output, _ := json.MarshalIndent(result, "", "  ")
	return []framework.TextContent{
		{Type: "text", Text: string(output)},
	}, nil
}

// handleOllamaGenerate generates text using Ollama
func handleOllamaGenerate(ctx context.Context, params map[string]interface{}, host string) ([]framework.TextContent, error) {
	prompt, _ := params["prompt"].(string)
	if prompt == "" {
		return nil, fmt.Errorf("prompt parameter is required for generate action")
	}

	model := "llama3.2"
	if m, ok := params["model"].(string); ok && m != "" {
		model = m
	}

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

	// Make request
	client := &http.Client{Timeout: 120 * time.Second} // Long timeout for generation
	url := fmt.Sprintf("%s/api/generate", host)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Ollama API: %w", err)
	}
	defer resp.Body.Close()

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
		output, _ := json.MarshalIndent(result, "", "  ")
		return []framework.TextContent{
			{Type: "text", Text: string(output)},
		}, nil
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

		output, _ := json.MarshalIndent(result, "", "  ")
		return []framework.TextContent{
			{Type: "text", Text: string(output)},
		}, nil
	}
}

// handleOllamaPull pulls/downloads a model
func handleOllamaPull(ctx context.Context, params map[string]interface{}, host string) ([]framework.TextContent, error) {
	model, _ := params["model"].(string)
	if model == "" {
		return nil, fmt.Errorf("model parameter is required for pull action")
	}

	// Create pull request
	reqBody := map[string]interface{}{
		"name": model,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make request
	client := &http.Client{Timeout: config.OllamaDownloadTimeout()} // Long timeout for model downloads
	url := fmt.Sprintf("%s/api/pull", host)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Ollama API: %w", err)
	}
	defer resp.Body.Close()

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

	output, _ := json.MarshalIndent(result, "", "  ")
	return []framework.TextContent{
		{Type: "text", Text: string(output)},
	}, nil
}

// handleOllamaHardware returns hardware info and recommendations
func handleOllamaHardware(ctx context.Context) ([]framework.TextContent, error) {
	// Simple hardware detection (can be enhanced)
	result := map[string]interface{}{
		"success": true,
		"method":  "native_go",
		"message": "Hardware detection not yet implemented in native Go. Use Python bridge for detailed hardware info.",
		"tip":     "Set OLLAMA_NUM_GPU and OLLAMA_NUM_THREADS environment variables for optimization",
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return []framework.TextContent{
		{Type: "text", Text: string(output)},
	}, nil
}
