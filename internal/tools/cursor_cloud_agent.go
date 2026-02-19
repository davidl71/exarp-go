// Package tools: cursor_cloud_agent — Cursor Cloud Agents API (Beta).
// Requires CURSOR_API_KEY from Cursor Dashboard → Integrations.
// See docs/CURSOR_API_AND_CLI_INTEGRATION.md §3.4.

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

	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/spf13/cast"
)

const cursorCloudAgentBaseURL = "https://api.cursor.com"

// cursorCloudAgentHTTPClient is used for Cloud Agents API calls (short timeout).
var cursorCloudAgentHTTPClient = &http.Client{Timeout: 60 * time.Second}

// handleCursorCloudAgentNative handles the cursor_cloud_agent tool (launch, status, list, follow_up, delete).
func handleCursorCloudAgentNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	apiKey := os.Getenv("CURSOR_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("CURSOR_API_KEY is not set; create an API key at Cursor Dashboard → Integrations")
	}

	action, _ := params["action"].(string)
	if action == "" {
		action = "list"
	}

	var out map[string]interface{}
	var err error

	switch action {
	case "launch":
		out, err = cursorCloudAgentLaunch(ctx, apiKey, params)
	case "status":
		out, err = cursorCloudAgentStatus(ctx, apiKey, params)
	case "list":
		out, err = cursorCloudAgentList(ctx, apiKey, params)
	case "follow_up":
		out, err = cursorCloudAgentFollowUp(ctx, apiKey, params)
	case "delete":
		out, err = cursorCloudAgentDelete(ctx, apiKey, params)
	default:
		return nil, fmt.Errorf("cursor_cloud_agent action %q not supported; use launch, status, list, follow_up, or delete", action)
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		out = map[string]interface{}{"action": action, "result": "ok"}
	}

	compact := cast.ToBool(params["compact"])
	return FormatResultOptionalCompact(out, "", compact)
}

func cursorCloudAgentAuthHeader(apiKey string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(strings.TrimSpace(apiKey)+":"))
}

func cursorCloudAgentDo(ctx context.Context, method, path string, apiKey string, body []byte) (*http.Response, error) {
	url := cursorCloudAgentBaseURL + path
	var req *http.Request
	var err error
	if body != nil {
		req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	}
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", cursorCloudAgentAuthHeader(apiKey))
	req.Header.Set("Content-Type", "application/json")

	return cursorCloudAgentHTTPClient.Do(req)
}

func cursorCloudAgentLaunch(ctx context.Context, apiKey string, params map[string]interface{}) (map[string]interface{}, error) {
	// Cloud Agents API: POST to launch (exact path from Cursor docs when available).
	// Beta API: common pattern is POST /v1/agents or /api/agents with body { "prompt", "repo?", "model?" }.
	prompt := strings.TrimSpace(cast.ToString(params["prompt"]))
	if prompt == "" {
		return nil, fmt.Errorf("launch requires prompt")
	}

	payload := map[string]interface{}{"prompt": prompt}
	if repo := strings.TrimSpace(cast.ToString(params["repo"])); repo != "" {
		payload["repo"] = repo
	}
	if model := strings.TrimSpace(cast.ToString(params["model"])); model != "" {
		payload["model"] = model
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal launch body: %w", err)
	}

	// Beta: path may be /v1/agents or /api/cloud-agents/launch; adjust when Cursor docs are confirmed.
	resp, err := cursorCloudAgentDo(ctx, http.MethodPost, "/v1/agents", apiKey, body)
	if err != nil {
		return nil, fmt.Errorf("launch request: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode launch response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("launch failed (HTTP %d): %v", resp.StatusCode, result)
	}

	out := map[string]interface{}{"action": "launch", "status_code": resp.StatusCode, "response": result}
	return out, nil
}

func cursorCloudAgentStatus(ctx context.Context, apiKey string, params map[string]interface{}) (map[string]interface{}, error) {
	agentID := strings.TrimSpace(cast.ToString(params["agent_id"]))
	if agentID == "" {
		return nil, fmt.Errorf("status requires agent_id")
	}

	resp, err := cursorCloudAgentDo(ctx, http.MethodGet, "/v1/agents/"+agentID, apiKey, nil)
	if err != nil {
		return nil, fmt.Errorf("status request: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode status response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("status failed (HTTP %d): %v", resp.StatusCode, result)
	}

	out := map[string]interface{}{"action": "status", "status_code": resp.StatusCode, "response": result}
	return out, nil
}

func cursorCloudAgentList(ctx context.Context, apiKey string, params map[string]interface{}) (map[string]interface{}, error) {
	resp, err := cursorCloudAgentDo(ctx, http.MethodGet, "/v1/agents", apiKey, nil)
	if err != nil {
		return nil, fmt.Errorf("list request: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode list response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("list failed (HTTP %d): %v", resp.StatusCode, result)
	}

	out := map[string]interface{}{"action": "list", "status_code": resp.StatusCode, "response": result}
	return out, nil
}

func cursorCloudAgentFollowUp(ctx context.Context, apiKey string, params map[string]interface{}) (map[string]interface{}, error) {
	agentID := strings.TrimSpace(cast.ToString(params["agent_id"]))
	if agentID == "" {
		return nil, fmt.Errorf("follow_up requires agent_id")
	}

	prompt := strings.TrimSpace(cast.ToString(params["prompt"]))
	if prompt == "" {
		return nil, fmt.Errorf("follow_up requires prompt")
	}

	body, err := json.Marshal(map[string]interface{}{"prompt": prompt})
	if err != nil {
		return nil, fmt.Errorf("marshal follow_up body: %w", err)
	}

	resp, err := cursorCloudAgentDo(ctx, http.MethodPost, "/v1/agents/"+agentID+"/follow_up", apiKey, body)
	if err != nil {
		return nil, fmt.Errorf("follow_up request: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode follow_up response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("follow_up failed (HTTP %d): %v", resp.StatusCode, result)
	}

	out := map[string]interface{}{"action": "follow_up", "status_code": resp.StatusCode, "response": result}
	return out, nil
}

func cursorCloudAgentDelete(ctx context.Context, apiKey string, params map[string]interface{}) (map[string]interface{}, error) {
	agentID := strings.TrimSpace(cast.ToString(params["agent_id"]))
	if agentID == "" {
		return nil, fmt.Errorf("delete requires agent_id")
	}

	resp, err := cursorCloudAgentDo(ctx, http.MethodDelete, "/v1/agents/"+agentID, apiKey, nil)
	if err != nil {
		return nil, fmt.Errorf("delete request: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&result) // optional body

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("delete failed (HTTP %d): %v", resp.StatusCode, result)
	}

	out := map[string]interface{}{"action": "delete", "status_code": resp.StatusCode, "agent_id": agentID}
	if len(result) > 0 {
		out["response"] = result
	}
	return out, nil
}
