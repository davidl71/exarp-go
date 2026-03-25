// session_helpers_handoff.go — Session helpers: handoff CRUD, git status, and suggested next actions.
// See also: session_helpers.go
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/davidl71/exarp-go/internal/cache"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/models"
	"github.com/spf13/cast"
)

// ─── Contents ───────────────────────────────────────────────────────────────
//   checkHandoffAlert
//   saveHandoff — saveHandoff saves a handoff note to the handoffs.json file.
//   updateHandoffStatus — updateHandoffStatus sets status on handoffs by id and writes handoffs.json back.
//   handleSessionHandoffStatus — handleSessionHandoffStatus closes or approves handoffs by id.
//   handleSessionHandoffDelete — handleSessionHandoffDelete removes handoffs by id from handoffs.json.
//   deleteHandoffs — deleteHandoffs removes handoffs by id from handoffs.json. Returns count deleted.
//   getGitStatus — getGitStatus gets current Git status.
//   buildSuggestedNextAction — buildSuggestedNextAction builds a client-agnostic next-action hint from a suggested task map.
//   buildCursorCliSuggestion — buildCursorCliSuggestion builds a ready-to-run Cursor CLI command from the first suggested task.
//   truncateString — truncateString truncates a string to max length.
// ────────────────────────────────────────────────────────────────────────────

// ─── checkHandoffAlert ──────────────────────────────────────────────────────
func checkHandoffAlert(projectRoot string) map[string]interface{} {
	handoffFile := filepath.Join(projectRoot, ".todo2", "handoffs.json")
	if _, err := os.Stat(handoffFile); os.IsNotExist(err) {
		return nil
	}

	fileCache := cache.GetGlobalFileCache()

	data, _, err := fileCache.ReadFile(handoffFile)
	if err != nil {
		return nil
	}

	var handoffData map[string]interface{}
	if err := json.Unmarshal(data, &handoffData); err != nil {
		return nil
	}

	handoffs, _ := handoffData["handoffs"].([]interface{})
	if len(handoffs) == 0 {
		return nil
	}

	// Get latest handoff
	latestHandoff := handoffs[len(handoffs)-1]

	handoffMap, ok := latestHandoff.(map[string]interface{})
	if !ok {
		return nil
	}

	hostname, _ := os.Hostname()
	handoffHost, _ := handoffMap["host"].(string)

	// Only show if from different host
	if handoffHost != hostname {
		alert := map[string]interface{}{
			"from_host":  handoffMap["host"],
			"timestamp":  handoffMap["timestamp"],
			"summary":    truncateString(fmt.Sprintf("%v", handoffMap["summary"]), 100),
			"blockers":   handoffMap["blockers"],
			"next_steps": handoffMap["next_steps"],
		}
		if latestLedger := readLatestLedgerSummary(projectRoot); latestLedger != nil {
			alert["latest_ledger"] = latestLedger
		}
		return alert
	}

	return nil
}

// ─── saveHandoff ────────────────────────────────────────────────────────────
// saveHandoff saves a handoff note to the handoffs.json file.
func saveHandoff(projectRoot string, handoff map[string]interface{}) error {
	handoffFile := filepath.Join(projectRoot, ".todo2", "handoffs.json")

	// Ensure .todo2 directory exists
	if err := os.MkdirAll(filepath.Dir(handoffFile), 0755); err != nil {
		return err
	}

	// Load existing handoffs
	fileCache := cache.GetGlobalFileCache()
	handoffs := []interface{}{}

	if data, _, err := fileCache.ReadFile(handoffFile); err == nil {
		var handoffData map[string]interface{}
		if err := json.Unmarshal(data, &handoffData); err == nil {
			if existing, ok := handoffData["handoffs"].([]interface{}); ok {
				handoffs = existing
			}
		}
	}

	// Add new handoff
	handoffs = append(handoffs, handoff)

	// Keep last 20 handoffs
	if len(handoffs) > 20 {
		handoffs = handoffs[len(handoffs)-20:]
	}

	// Save
	handoffData := map[string]interface{}{
		"handoffs": handoffs,
	}

	data, err := json.MarshalIndent(handoffData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(handoffFile, data, 0644)
}

// ─── updateHandoffStatus ────────────────────────────────────────────────────
// updateHandoffStatus sets status on handoffs by id and writes handoffs.json back.
func updateHandoffStatus(projectRoot string, handoffIDs []string, status string) error {
	if len(handoffIDs) == 0 {
		return nil
	}

	handoffFile := filepath.Join(projectRoot, ".todo2", "handoffs.json")
	if err := os.MkdirAll(filepath.Dir(handoffFile), 0755); err != nil {
		return err
	}

	idsSet := make(map[string]struct{})

	for _, id := range handoffIDs {
		if id != "" {
			idsSet[id] = struct{}{}
		}
	}

	fileCache := cache.GetGlobalFileCache()

	data, _, err := fileCache.ReadFile(handoffFile)
	if err != nil {
		return err
	}

	var handoffData map[string]interface{}
	if err := json.Unmarshal(data, &handoffData); err != nil {
		return err
	}

	handoffs, _ := handoffData["handoffs"].([]interface{})
	if len(handoffs) == 0 {
		return nil
	}

	updated := 0

	for _, v := range handoffs {
		h, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		id, _ := h["id"].(string)
		if _, want := idsSet[id]; want {
			h["status"] = status
			updated++
		}
	}

	if updated == 0 {
		return nil
	}

	handoffData["handoffs"] = handoffs

	out, err := json.MarshalIndent(handoffData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(handoffFile, out, 0644)
}

// ─── handleSessionHandoffStatus ─────────────────────────────────────────────
// handleSessionHandoffStatus closes or approves handoffs by id.
func handleSessionHandoffStatus(ctx context.Context, params map[string]interface{}, projectRoot, status string) ([]framework.TextContent, error) {
	var ids []string
	if id := strings.TrimSpace(cast.ToString(params["handoff_id"])); id != "" {
		ids = []string{id}
	} else if raw, ok := params["handoff_ids"]; ok {
		switch v := raw.(type) {
		case []interface{}:
			for _, i := range v {
				if s, ok := i.(string); ok && s != "" {
					ids = append(ids, s)
				}
			}
		case string:
			if v != "" {
				var list []string
				if json.Unmarshal([]byte(v), &list) == nil {
					ids = list
				} else {
					ids = []string{v}
				}
			}
		}
	}

	if len(ids) == 0 {
		return nil, fmt.Errorf("handoff_id or handoff_ids required for close/approve")
	}

	if err := updateHandoffStatus(projectRoot, ids, status); err != nil {
		return nil, fmt.Errorf("failed to update handoff status: %w", err)
	}

	label := "closed"
	if status == "approved" {
		label = "approved"
	}

	result := map[string]interface{}{
		"success": true,
		"method":  "native_go",
		"updated": len(ids),
		"status":  status,
		"message": fmt.Sprintf("%d handoff(s) %s", len(ids), label),
	}

	return framework.FormatResult(result, "")
}

// ─── handleSessionHandoffDelete ─────────────────────────────────────────────
// handleSessionHandoffDelete removes handoffs by id from handoffs.json.
func handleSessionHandoffDelete(ctx context.Context, params map[string]interface{}, projectRoot string) ([]framework.TextContent, error) {
	var ids []string
	if id := strings.TrimSpace(cast.ToString(params["handoff_id"])); id != "" {
		ids = []string{id}
	} else if raw, ok := params["handoff_ids"]; ok {
		switch v := raw.(type) {
		case []interface{}:
			for _, i := range v {
				if s, ok := i.(string); ok && s != "" {
					ids = append(ids, s)
				}
			}
		case string:
			if v != "" {
				var list []string
				if json.Unmarshal([]byte(v), &list) == nil {
					ids = list
				} else {
					ids = []string{v}
				}
			}
		}
	}

	if len(ids) == 0 {
		return nil, fmt.Errorf("handoff_id or handoff_ids required for delete")
	}

	deleted, err := deleteHandoffs(projectRoot, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to delete handoffs: %w", err)
	}

	result := map[string]interface{}{
		"success": true,
		"method":  "native_go",
		"deleted": deleted,
		"message": fmt.Sprintf("%d handoff(s) deleted", deleted),
	}

	return framework.FormatResult(result, "")
}

// ─── deleteHandoffs ─────────────────────────────────────────────────────────
// deleteHandoffs removes handoffs by id from handoffs.json. Returns count deleted.
func deleteHandoffs(projectRoot string, handoffIDs []string) (int, error) {
	if len(handoffIDs) == 0 {
		return 0, nil
	}

	handoffFile := filepath.Join(projectRoot, ".todo2", "handoffs.json")
	if err := os.MkdirAll(filepath.Dir(handoffFile), 0755); err != nil {
		return 0, err
	}

	idsSet := make(map[string]struct{})

	for _, id := range handoffIDs {
		if id != "" {
			idsSet[id] = struct{}{}
		}
	}

	if _, err := os.Stat(handoffFile); os.IsNotExist(err) {
		return 0, nil
	}

	data, err := os.ReadFile(handoffFile)
	if err != nil {
		return 0, err
	}

	var handoffData map[string]interface{}
	if err := json.Unmarshal(data, &handoffData); err != nil {
		return 0, err
	}

	handoffs, _ := handoffData["handoffs"].([]interface{})

	var kept []interface{}

	deleted := 0

	for _, v := range handoffs {
		h, ok := v.(map[string]interface{})
		if !ok {
			kept = append(kept, v)
			continue
		}

		id, _ := h["id"].(string)
		if _, want := idsSet[id]; want {
			deleted++
			continue
		}

		kept = append(kept, v)
	}

	if deleted == 0 {
		return 0, nil
	}

	handoffData["handoffs"] = kept

	out, err := json.MarshalIndent(handoffData, "", "  ")
	if err != nil {
		return 0, err
	}

	if err := os.WriteFile(handoffFile, out, 0644); err != nil {
		return 0, err
	}

	return deleted, nil
}

// ─── getGitStatus ───────────────────────────────────────────────────────────
// getGitStatus gets current Git status.
func getGitStatus(ctx context.Context, projectRoot string) map[string]interface{} {
	status := map[string]interface{}{}

	// Get branch
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = projectRoot

	if output, err := cmd.Output(); err == nil {
		status["branch"] = strings.TrimSpace(string(output))
	}

	// Get status
	cmd = exec.CommandContext(ctx, "git", "status", "--porcelain")
	cmd.Dir = projectRoot

	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")

		var changedFiles []string

		for _, line := range lines {
			if line != "" {
				changedFiles = append(changedFiles, strings.TrimSpace(line))
			}
		}

		status["uncommitted_files"] = len(changedFiles)
		status["changed_files"] = changedFiles

		if len(changedFiles) > 10 {
			status["changed_files"] = changedFiles[:10]
		}
	}

	return status
}

// ─── buildSuggestedNextAction ───────────────────────────────────────────────
// buildSuggestedNextAction builds a client-agnostic next-action hint from a suggested task map.
// Expects a map with "id" and "content" keys. Returns empty string if task info is missing.
// For Cursor: can be used as argument to `agent -p`. For Claude Code: descriptive action hint.
func buildSuggestedNextAction(task map[string]interface{}) string {
	id, _ := task["id"].(string)
	content, _ := task["content"].(string)

	if id == "" {
		return ""
	}

	if content != "" {
		return fmt.Sprintf("Work on %s: %s", id, truncateString(content, 80))
	}

	return fmt.Sprintf("Work on %s", id)
}

// ─── buildCursorCliSuggestion ───────────────────────────────────────────────
// buildCursorCliSuggestion builds a ready-to-run Cursor CLI command from the first suggested task.
// Returns e.g. `agent -p "Work on T-123: Task name" --mode=plan` for session prime/handoff JSON.
// See docs/CURSOR_API_AND_CLI_INTEGRATION.md §3.2.
func buildCursorCliSuggestion(task map[string]interface{}) string {
	action := buildSuggestedNextAction(task)
	if action == "" {
		return ""
	}

	return fmt.Sprintf("agent -p %q --mode=plan", action)
}

// ─── truncateString ─────────────────────────────────────────────────────────
// truncateString truncates a string to max length.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	return s[:maxLen-3] + "..."
}

// ─── buildOwnershipHints ────────────────────────────────────────────────────
// buildOwnershipHints checks suggested tasks for file collisions and returns warning hints.
// Returns a list of warning strings about parallelization risks.
func buildOwnershipHints(suggestedTasks []Todo2Task) []string {
	if len(suggestedTasks) < 2 {
		return nil
	}

	// Build ownership map for suggested tasks
	ownershipMap := make(map[string]*models.TaskOwnership)
	for i := range suggestedTasks {
		own := models.GetTaskOwnership(&suggestedTasks[i])
		if own != nil && (len(own.OwnedFiles) > 0 || own.Lane != "") {
			ownershipMap[suggestedTasks[i].ID] = own
		}
	}

	if len(ownershipMap) < 2 {
		return nil
	}

	var hints []string

	// Check for file collisions
	fileToTasks := make(map[string][]string)
	for taskID, own := range ownershipMap {
		for _, f := range own.OwnedFiles {
			fileToTasks[f] = append(fileToTasks[f], taskID)
		}
	}

	for file, taskIDs := range fileToTasks {
		if len(taskIDs) >= 2 {
			sort.Strings(taskIDs)
			hints = append(hints, fmt.Sprintf("⚠️ File collision: %s shared by %s (run serially)", file, strings.Join(taskIDs, ", ")))
		}
	}

	// Check for same-lane tasks
	laneToTasks := make(map[string][]string)
	for taskID, own := range ownershipMap {
		if own.Lane != "" {
			laneToTasks[own.Lane] = append(laneToTasks[own.Lane], taskID)
		}
	}

	for lane, taskIDs := range laneToTasks {
		if len(taskIDs) >= 2 {
			sort.Strings(taskIDs)
			hints = append(hints, fmt.Sprintf("⚠️ Same lane (%s): %s — may have related files", lane, strings.Join(taskIDs, ", ")))
		}
	}

	return hints
}

// handleSessionPrompts handles the prompts action - lists available prompts.
