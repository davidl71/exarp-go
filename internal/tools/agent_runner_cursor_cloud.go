// agent_runner_cursor_cloud.go — AgentRunner implementation for Cursor Cloud Agents API.
package tools

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const cursorCloudAgentBaseURLRunner = "https://api.cursor.com"

type CursorCloudRunner struct {
	httpClient *http.Client
}

func NewCursorCloudRunner() *CursorCloudRunner {
	return &CursorCloudRunner{
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (r *CursorCloudRunner) Type() AgentType {
	return AgentTypeCursorCloud
}

func (r *CursorCloudRunner) Available() bool {
	return os.Getenv("CURSOR_API_KEY") != ""
}

func (r *CursorCloudRunner) Run(ctx context.Context, opts AgentRunnerOptions) AgentResult {
	apiKey := os.Getenv("CURSOR_API_KEY")
	if apiKey == "" {
		return AgentResult{Error: fmt.Errorf("CURSOR_API_KEY not set")}
	}

	if opts.Prompt == "" {
		return AgentResult{Error: fmt.Errorf("prompt required")}
	}

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	payload := map[string]interface{}{"prompt": opts.Prompt}
	if opts.ProjectRoot != "" {
		payload["repo"] = opts.ProjectRoot
	}
	if opts.Model != "" {
		payload["model"] = opts.Model
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return AgentResult{Error: fmt.Errorf("marshal payload: %w", err)}
	}

	url := cursorCloudAgentBaseURLRunner + "/v1/agents"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return AgentResult{Error: fmt.Errorf("create request: %w", err)}
	}

	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(strings.TrimSpace(apiKey)+":")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return AgentResult{Error: fmt.Errorf("request failed: %w", err)}
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return AgentResult{Error: fmt.Errorf("decode response: %w", err), ExitCode: resp.StatusCode}
	}

	if resp.StatusCode >= 400 {
		return AgentResult{
			Error:    fmt.Errorf("launch failed (HTTP %d): %v", resp.StatusCode, result),
			ExitCode: resp.StatusCode,
		}
	}

	outputJSON, _ := json.Marshal(result)
	return AgentResult{
		Output:   string(outputJSON),
		ExitCode: 0,
	}
}

func init() {
	RegisterAgentRunner(NewCursorCLIRunner())
	RegisterAgentRunner(NewCursorCloudRunner())
}
