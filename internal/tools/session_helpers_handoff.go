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
//   saveHandoff — appends a typed HandoffEntry to .todo2/handoffs.store (gzip+gob).
//   updateHandoffStatus — sets status on handoffs by id and rewrites the store.
//   handleSessionHandoffStatus — closes or approves handoffs by id.
//   handleSessionHandoffDelete — removes handoffs by id from the store.
//   deleteHandoffs — removes handoffs by id. Returns count deleted.
//   getGitStatus — getGitStatus gets current Git status.
//   buildSuggestedNextAction — buildSuggestedNextAction builds a client-agnostic next-action hint from a suggested task map.
//   buildCursorCliSuggestion — buildCursorCliSuggestion builds a ready-to-run Cursor CLI command from the first suggested task.
//   truncateString — truncateString truncates a string to max length.
// ────────────────────────────────────────────────────────────────────────────

// ─── checkHandoffAlert ──────────────────────────────────────────────────────
func checkHandoffAlert(projectRoot string) map[string]interface{} {
	if !handoffsAnyFileExists(projectRoot) {
		return nil
	}

	fileCache := cache.GetGlobalFileCache()
	path := handoffsPersistPath(projectRoot)
	data, _, err := fileCache.ReadFile(path)
	if err != nil {
		return nil
	}

	store, err := loadHandoffStoreFromBytes(path, data)
	if err != nil || len(store.Handoffs) == 0 {
		return nil
	}

	latest := store.Handoffs[len(store.Handoffs)-1]
	hostname, _ := os.Hostname()
	if latest.Host == hostname {
		return nil
	}

	alert := map[string]interface{}{
		"from_host":  latest.Host,
		"timestamp":  latest.Timestamp,
		"summary":    truncateString(latest.Summary, 100),
		"blockers":   latest.Blockers,
		"next_steps": latest.NextSteps,
	}
	if latestLedger := readLatestLedgerSummary(projectRoot); latestLedger != nil {
		alert["latest_ledger"] = latestLedger
	}
	return alert
}

// ─── saveHandoff ────────────────────────────────────────────────────────────
// saveHandoff appends a handoff entry and persists HandoffStore (gzip+gob).
func saveHandoff(projectRoot string, entry HandoffEntry) error {
	store, err := loadHandoffStore(projectRoot)
	if err != nil {
		return err
	}

	store.Handoffs = append(store.Handoffs, entry)
	if len(store.Handoffs) > 20 {
		store.Handoffs = store.Handoffs[len(store.Handoffs)-20:]
	}

	return saveHandoffStore(projectRoot, store)
}

// ─── updateHandoffStatus ────────────────────────────────────────────────────
// updateHandoffStatus sets status on handoffs by id and rewrites the store.
func updateHandoffStatus(projectRoot string, handoffIDs []string, status string) error {
	if len(handoffIDs) == 0 {
		return nil
	}

	if err := os.MkdirAll(filepath.Join(projectRoot, ".todo2"), 0o755); err != nil {
		return err
	}

	idsSet := make(map[string]struct{})
	for _, id := range handoffIDs {
		if id != "" {
			idsSet[id] = struct{}{}
		}
	}

	store, err := loadHandoffStore(projectRoot)
	if err != nil {
		return err
	}
	if len(store.Handoffs) == 0 {
		return nil
	}

	updated := 0
	for i := range store.Handoffs {
		if _, want := idsSet[store.Handoffs[i].ID]; want {
			store.Handoffs[i].Status = status
			updated++
		}
	}
	if updated == 0 {
		return nil
	}

	return saveHandoffStore(projectRoot, store)
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
// handleSessionHandoffDelete removes handoffs by id from the store.
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
// deleteHandoffs removes handoffs by id from the store. Returns count deleted.
func deleteHandoffs(projectRoot string, handoffIDs []string) (int, error) {
	if len(handoffIDs) == 0 {
		return 0, nil
	}

	if err := os.MkdirAll(filepath.Join(projectRoot, ".todo2"), 0o755); err != nil {
		return 0, err
	}

	idsSet := make(map[string]struct{})
	for _, id := range handoffIDs {
		if id != "" {
			idsSet[id] = struct{}{}
		}
	}

	if !handoffsAnyFileExists(projectRoot) {
		return 0, nil
	}

	store, err := loadHandoffStore(projectRoot)
	if err != nil {
		return 0, err
	}

	var kept []HandoffEntry
	deleted := 0
	for _, h := range store.Handoffs {
		if _, want := idsSet[h.ID]; want {
			deleted++
			continue
		}
		kept = append(kept, h)
	}
	if deleted == 0 {
		return 0, nil
	}

	store.Handoffs = kept
	if err := saveHandoffStore(projectRoot, store); err != nil {
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

// ─── buildHotspotSummary ────────────────────────────────────────────────────
// buildHotspotSummary analyzes all pending tasks and returns a summary of contested files.
// Returns a list of hotspot entries: "file_path: N tasks (task_ids)"
func buildHotspotSummary(tasks []Todo2Task) []string {
	// Count how many tasks touch each file
	fileToTasks := make(map[string][]string)

	for i := range tasks {
		task := &tasks[i]
		if !IsPendingStatus(task.Status) {
			continue
		}

		own := models.GetTaskOwnership(task)
		if own == nil {
			continue
		}

		for _, f := range own.OwnedFiles {
			fileToTasks[f] = append(fileToTasks[f], task.ID)
		}
	}

	// Build hotspot list (files touched by 2+ tasks)
	var hotspots []string
	for file, taskIDs := range fileToTasks {
		if len(taskIDs) >= 2 {
			sort.Strings(taskIDs)
			hotspots = append(hotspots, fmt.Sprintf("%s: %d tasks (%s)", file, len(taskIDs), strings.Join(taskIDs, ", ")))
		}
	}

	sort.Strings(hotspots)
	return hotspots
}

// handleSessionPrompts handles the prompts action - lists available prompts.
