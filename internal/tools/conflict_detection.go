// Package tools: multi-agent conflict detection (T-1770829104089, T-1770829100451).
// Task-overlap: two In Progress tasks where one depends on the other.
// File-level: In Progress tasks that list the same files in long_description.

package tools

import (
	"context"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/davidl71/exarp-go/internal/database"
)

// TaskOverlapConflict describes a dependency conflict: two tasks both In Progress where one blocks the other.
type TaskOverlapConflict struct {
	TaskID    string `json:"task_id"`    // Task that is In Progress
	DepTaskID string `json:"dep_task_id"` // Dependency of TaskID that is also In Progress
	Reason    string `json:"reason"`     // Human-readable reason
}

// FileConflict describes overlapping file access: multiple In Progress tasks touch the same file(s).
type FileConflict struct {
	TaskIDs []string `json:"task_ids"` // Task IDs that overlap
	Files   []string `json:"files"`    // File paths that overlap
}

// DetectTaskOverlapConflicts returns overlapping In Progress tasks (A blocks B, both In Progress).
// Pass tasks from store.ListTasks(ctx, nil); only In Progress tasks are considered.
func DetectTaskOverlapConflicts(tasks []*database.Todo2Task) []TaskOverlapConflict {
	inProgressSet := make(map[string]bool)
	for _, t := range tasks {
		if t != nil && NormalizeStatusToTitleCase(t.Status) == "In Progress" {
			inProgressSet[t.ID] = true
		}
	}
	var out []TaskOverlapConflict
	for _, t := range tasks {
		if t == nil || NormalizeStatusToTitleCase(t.Status) != "In Progress" {
			continue
		}
		for _, depID := range t.Dependencies {
			if inProgressSet[depID] {
				out = append(out, TaskOverlapConflict{
					TaskID:    t.ID,
					DepTaskID: depID,
					Reason:    depID + " blocks " + t.ID + "; both In Progress",
				})
			}
		}
	}
	return out
}

// filesFromLongDescription extracts file paths from a task long_description (Files/Components section).
// Looks for lines like "- Update: path" or "- Create: path" or "Update: path" (optional leading dash).
var filePathInDescriptionRE = regexp.MustCompile(`(?m)^\s*-\s*(?:Update|Create|Modify|Delete):\s*([^\s#\n]+)`)

func filesFromLongDescription(longDesc string) []string {
	// Normalize path separators to forward slash for dedup
	norm := func(p string) string {
		p = strings.TrimSpace(p)
		p = filepath.Clean(p)
		return filepath.ToSlash(p)
	}
	seen := make(map[string]bool)
	var out []string
	for _, m := range filePathInDescriptionRE.FindAllStringSubmatch(longDesc, -1) {
		if len(m) < 2 || m[1] == "" {
			continue
		}
		p := norm(m[1])
		if p != "" && !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	return out
}

// DetectFileConflicts returns file-level conflicts: In Progress tasks that list the same file(s).
// File paths are extracted from long_description (Files/Components patterns).
func DetectFileConflicts(tasks []*database.Todo2Task) []FileConflict {
	taskFiles := make(map[string][]string)
	for _, t := range tasks {
		if t == nil || NormalizeStatusToTitleCase(t.Status) != "In Progress" {
			continue
		}
		files := filesFromLongDescription(t.LongDescription)
		if len(files) > 0 {
			taskFiles[t.ID] = files
		}
	}
	fileToTasks := make(map[string][]string)
	for taskID, files := range taskFiles {
		for _, f := range files {
			fileToTasks[f] = append(fileToTasks[f], taskID)
		}
	}
	// Group by task set: key = sorted task IDs, value = all shared files
	setToFiles := make(map[string][]string)
	for file, ids := range fileToTasks {
		if len(ids) < 2 {
			continue
		}
		sortStrings(ids)
		key := strings.Join(ids, "|")
		setToFiles[key] = append(setToFiles[key], file)
	}
	var out []FileConflict
	for key, files := range setToFiles {
		ids := strings.Split(key, "|")
		out = append(out, FileConflict{TaskIDs: ids, Files: files})
	}
	return out
}

func sortStrings(s []string) {
	for i := 0; i < len(s); i++ {
		for j := i + 1; j < len(s); j++ {
			if s[j] < s[i] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

// DetectConflicts loads tasks from the store and returns task-overlap and file-level conflicts.
func DetectConflicts(ctx context.Context, projectRoot string) (taskOverlaps []TaskOverlapConflict, fileConflicts []FileConflict, err error) {
	store := NewDefaultTaskStore(projectRoot)
	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	taskOverlaps = DetectTaskOverlapConflicts(list)
	fileConflicts = DetectFileConflicts(list)
	return taskOverlaps, fileConflicts, nil
}
