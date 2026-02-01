package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/cache"
	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/framework"
	mcpresponse "github.com/davidl71/mcp-go-core/pkg/mcp/response"
)

// WorkflowModeState represents the persisted workflow mode state
type WorkflowModeState struct {
	CurrentMode    string   `json:"current_mode"`
	ExtraGroups    []string `json:"extra_groups"`
	DisabledGroups []string `json:"disabled_groups"`
	LastUpdated    string   `json:"last_updated"`
}

// WorkflowModeManager manages workflow mode state
type WorkflowModeManager struct {
	state     WorkflowModeState
	statePath string
}

var globalWorkflowManager *WorkflowModeManager

// getWorkflowManager returns the global workflow mode manager
func getWorkflowManager() *WorkflowModeManager {
	if globalWorkflowManager == nil {
		exarpPath := config.ProjectExarpPath()
		if exarpPath == "" {
			exarpPath = ".exarp"
		}
		defaultMode := config.WorkflowDefaultMode()
		if defaultMode == "" {
			defaultMode = "development"
		}

		statePath := filepath.Join(exarpPath, "workflow_mode.json")
		if projectRoot, err := FindProjectRoot(); err == nil && projectRoot != "" {
			statePath = filepath.Join(projectRoot, exarpPath, "workflow_mode.json")
		} else if projectRoot := os.Getenv("PROJECT_ROOT"); projectRoot != "" && projectRoot != "unknown" {
			statePath = filepath.Join(projectRoot, exarpPath, "workflow_mode.json")
		}

		globalWorkflowManager = &WorkflowModeManager{
			state: WorkflowModeState{
				CurrentMode:    defaultMode,
				ExtraGroups:    []string{},
				DisabledGroups: []string{},
				LastUpdated:    time.Now().Format(time.RFC3339),
			},
			statePath: statePath,
		}

		// Load existing state if available
		_ = globalWorkflowManager.loadState()
	}
	return globalWorkflowManager
}

// loadState loads workflow mode state from file
func (m *WorkflowModeManager) loadState() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(m.statePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Try to load existing state (using file cache)
	fileCache := cache.GetGlobalFileCache()
	data, _, err := fileCache.ReadFile(m.statePath)
	if err != nil {
		if os.IsNotExist(err) {
			// No existing state - use defaults
			return nil
		}
		return fmt.Errorf("failed to read state file: %w", err)
	}

	var state WorkflowModeState
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to unmarshal state: %w", err)
	}

	m.state = state
	return nil
}

// saveState saves workflow mode state to file
func (m *WorkflowModeManager) saveState() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(m.statePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	m.state.LastUpdated = time.Now().Format(time.RFC3339)
	data, err := json.MarshalIndent(m.state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(m.statePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// getStatus returns current workflow mode status
func (m *WorkflowModeManager) getStatus() map[string]interface{} {
	return map[string]interface{}{
		"mode":            m.state.CurrentMode,
		"extra_groups":    m.state.ExtraGroups,
		"disabled_groups": m.state.DisabledGroups,
		"last_updated":    m.state.LastUpdated,
		"available_modes": []string{
			"daily_checkin", "security_review", "task_management",
			"sprint_planning", "code_review", "development", "debugging", "all",
		},
		"available_groups": []string{
			"core", "tool_catalog", "health", "tasks", "security",
			"automation", "config", "testing", "advisors", "memory",
			"workflow", "prd",
		},
	}
}

// setMode sets the workflow mode and returns the old mode
func (m *WorkflowModeManager) setMode(mode string) (string, error) {
	validModes := map[string]bool{
		"daily_checkin":   true,
		"security_review": true,
		"task_management": true,
		"sprint_planning": true,
		"code_review":     true,
		"development":     true,
		"debugging":       true,
		"all":             true,
	}

	if !validModes[strings.ToLower(mode)] {
		return "", fmt.Errorf("unknown mode: %s", mode)
	}

	oldMode := m.state.CurrentMode
	m.state.CurrentMode = strings.ToLower(mode)
	m.state.ExtraGroups = []string{}
	m.state.DisabledGroups = []string{}

	_ = m.saveState()

	return oldMode, nil
}

// enableGroup enables a tool group
func (m *WorkflowModeManager) enableGroup(group string) error {
	group = strings.ToLower(group)

	// Remove from disabled if present
	for i, g := range m.state.DisabledGroups {
		if strings.ToLower(g) == group {
			m.state.DisabledGroups = append(m.state.DisabledGroups[:i], m.state.DisabledGroups[i+1:]...)
			break
		}
	}

	// Add to extra if not present
	found := false
	for _, g := range m.state.ExtraGroups {
		if strings.ToLower(g) == group {
			found = true
			break
		}
	}
	if !found {
		m.state.ExtraGroups = append(m.state.ExtraGroups, group)
	}

	_ = m.saveState()
	return nil
}

// disableGroup disables a tool group
func (m *WorkflowModeManager) disableGroup(group string) error {
	group = strings.ToLower(group)

	// Cannot disable core or tool_catalog
	if group == "core" || group == "tool_catalog" {
		return fmt.Errorf("cannot disable %s group - always required", group)
	}

	// Remove from extra if present
	for i, g := range m.state.ExtraGroups {
		if strings.ToLower(g) == group {
			m.state.ExtraGroups = append(m.state.ExtraGroups[:i], m.state.ExtraGroups[i+1:]...)
			break
		}
	}

	// Add to disabled if not present
	found := false
	for _, g := range m.state.DisabledGroups {
		if strings.ToLower(g) == group {
			found = true
			break
		}
	}
	if !found {
		m.state.DisabledGroups = append(m.state.DisabledGroups, group)
	}

	_ = m.saveState()
	return nil
}

// suggestMode suggests a workflow mode based on text (simplified version)
func (m *WorkflowModeManager) suggestMode(text string) map[string]interface{} {
	if text == "" {
		return map[string]interface{}{
			"suggested_mode": nil,
			"confidence":     0.0,
			"rationale":      "No text provided for mode suggestion",
			"current_mode":   m.state.CurrentMode,
		}
	}

	textLower := strings.ToLower(text)

	// Simple keyword-based suggestion
	modeScores := map[string]int{
		"security_review": strings.Count(textLower, "security") + strings.Count(textLower, "vulnerability"),
		"task_management": strings.Count(textLower, "task") + strings.Count(textLower, "todo"),
		"code_review":     strings.Count(textLower, "review") + strings.Count(textLower, "test"),
		"sprint_planning": strings.Count(textLower, "sprint") + strings.Count(textLower, "plan"),
		"daily_checkin":   strings.Count(textLower, "daily") + strings.Count(textLower, "status"),
		"development":     strings.Count(textLower, "implement") + strings.Count(textLower, "develop"),
		"debugging":       strings.Count(textLower, "debug") + strings.Count(textLower, "bug"),
	}

	bestMode := ""
	bestScore := 0
	for mode, score := range modeScores {
		if score > bestScore {
			bestScore = score
			bestMode = mode
		}
	}

	if bestScore == 0 {
		return map[string]interface{}{
			"suggested_mode": nil,
			"confidence":     0.0,
			"rationale":      "No clear mode suggestion based on text",
			"current_mode":   m.state.CurrentMode,
		}
	}

	confidence := float64(bestScore) / 10.0
	if confidence > 1.0 {
		confidence = 1.0
	}

	return map[string]interface{}{
		"suggested_mode": bestMode,
		"confidence":     confidence,
		"rationale":      fmt.Sprintf("Keywords in text suggest %s mode", bestMode),
		"current_mode":   m.state.CurrentMode,
		"would_change":   bestMode != m.state.CurrentMode,
	}
}

// handleWorkflowModeNative handles the workflow_mode tool with native Go implementation
func handleWorkflowModeNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	action, _ := params["action"].(string)
	if action == "" {
		action = "focus"
	}

	manager := getWorkflowManager()

	switch action {
	case "focus":
		return handleWorkflowModeFocus(ctx, params, manager)
	case "suggest":
		return handleWorkflowModeSuggest(ctx, params, manager)
	case "stats":
		return handleWorkflowModeStats(ctx, manager)
	default:
		errorResponse := map[string]interface{}{
			"status": "error",
			"error":  fmt.Sprintf("Unknown workflow_mode action: %s. Use 'focus', 'suggest', or 'stats'.", action),
		}
		return mcpresponse.FormatResult(errorResponse, "")
	}
}

// handleWorkflowModeFocus handles the focus action
func handleWorkflowModeFocus(ctx context.Context, params map[string]interface{}, manager *WorkflowModeManager) ([]framework.TextContent, error) {
	statusOnly, _ := params["status"].(bool)
	mode, _ := params["mode"].(string)
	enableGroup, _ := params["enable_group"].(string)
	disableGroup, _ := params["disable_group"].(string)

	// Status only
	if statusOnly || (mode == "" && enableGroup == "" && disableGroup == "") {
		statusMap, err := mcpresponse.ConvertToMap(manager.getStatus())
		if err != nil {
			return nil, fmt.Errorf("failed to convert status: %w", err)
		}
		return mcpresponse.FormatResult(statusMap, "")
	}

	var response map[string]interface{}

	if mode != "" {
		oldMode, err := manager.setMode(mode)
		if err != nil {
			errorResponse := map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			}
			return mcpresponse.FormatResult(errorResponse, "")
		}
		response = map[string]interface{}{
			"success":       true,
			"action":        "mode_changed",
			"mode":          manager.state.CurrentMode,
			"previous_mode": oldMode,
		}
		status := manager.getStatus()
		for k, v := range status {
			response[k] = v
		}
	} else if enableGroup != "" {
		if err := manager.enableGroup(enableGroup); err != nil {
			errorResponse := map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			}
			return mcpresponse.FormatResult(errorResponse, "")
		}
		response = map[string]interface{}{
			"success": true,
			"action":  "group_enabled",
			"group":   enableGroup,
		}
		status := manager.getStatus()
		for k, v := range status {
			response[k] = v
		}
	} else if disableGroup != "" {
		if err := manager.disableGroup(disableGroup); err != nil {
			errorResponse := map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			}
			return mcpresponse.FormatResult(errorResponse, "")
		}
		response = map[string]interface{}{
			"success": true,
			"action":  "group_disabled",
			"group":   disableGroup,
		}
		status := manager.getStatus()
		for k, v := range status {
			response[k] = v
		}
	}

	return mcpresponse.FormatResult(response, "")
}

// handleWorkflowModeSuggest handles the suggest action
func handleWorkflowModeSuggest(ctx context.Context, params map[string]interface{}, manager *WorkflowModeManager) ([]framework.TextContent, error) {
	text, _ := params["text"].(string)
	autoSwitch, _ := params["auto_switch"].(bool)

	suggestion := manager.suggestMode(text)

	// Auto-switch if requested and confident
	if autoSwitch && suggestion["suggested_mode"] != nil {
		if conf, ok := suggestion["confidence"].(float64); ok && conf >= 0.5 {
			if mode, ok := suggestion["suggested_mode"].(string); ok {
				_, err := manager.setMode(mode)
				if err == nil {
					suggestion["auto_switched"] = true
				} else {
					suggestion["auto_switched"] = false
					suggestion["error"] = fmt.Sprintf("Could not auto-switch to suggested mode: %v", err)
				}
			}
		} else {
			suggestion["auto_switched"] = false
		}
	} else {
		suggestion["auto_switched"] = false
	}

	return mcpresponse.FormatResult(suggestion, "")
}

// handleWorkflowModeStats handles the stats action
func handleWorkflowModeStats(ctx context.Context, manager *WorkflowModeManager) ([]framework.TextContent, error) {
	stats := map[string]interface{}{
		"success":         true,
		"current_mode":    manager.state.CurrentMode,
		"extra_groups":    manager.state.ExtraGroups,
		"disabled_groups": manager.state.DisabledGroups,
		"last_updated":    manager.state.LastUpdated,
		"note":            "Stats show current mode state; per-tool usage counts are not tracked.",
	}

	return mcpresponse.FormatResult(stats, "")
}
