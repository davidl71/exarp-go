// Package tools: multi-agent conflict detection (T-1770829104089, T-1770829100451).
// Task-overlap: two In Progress tasks where one depends on the other.
// File-level: In Progress tasks that list the same files in long_description.

package tools

import (
	"context"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/models"
)

// TaskOverlapConflict describes a dependency conflict: two tasks both In Progress where one blocks the other.
type TaskOverlapConflict struct {
	TaskID    string `json:"task_id"`     // Task that is In Progress
	DepTaskID string `json:"dep_task_id"` // Dependency of TaskID that is also In Progress
	Reason    string `json:"reason"`      // Human-readable reason
}

// FileConflict describes overlapping file access: multiple In Progress tasks touch the same file(s).
type FileConflict struct {
	TaskIDs    []string          `json:"task_ids"`              // Task IDs that overlap
	Files      []string          `json:"files"`                 // File paths that overlap
	TaskStatus map[string]string `json:"task_status,omitempty"` // normalized status per task (preflight / mixed-status)
}

// ForbiddenOwnershipConflict is when an In Progress task works on a path another In Progress
// task declared forbidden via ownership metadata (forbidden_files / glob).
type ForbiddenOwnershipConflict struct {
	TaskID      string `json:"task_id"`       // task whose concrete path triggers the violation
	OtherTaskID string `json:"other_task_id"` // peer task that forbids that path
	Path        string `json:"path"`
	Reason      string `json:"reason"` // e.g. forbidden_file, forbidden_glob
}

// DetectTaskOverlapConflicts returns overlapping In Progress tasks (A blocks B, both In Progress).
// Pass tasks from store.ListTasks(ctx, nil); only In Progress tasks are considered.
func DetectTaskOverlapConflicts(tasks []*database.Todo2Task) []TaskOverlapConflict {
	inProgressSet := make(map[string]bool)

	for _, t := range tasks {
		if t != nil && NormalizeStatusToTitleCase(t.Status) == models.StatusInProgress {
			inProgressSet[t.ID] = true
		}
	}

	var out []TaskOverlapConflict

	for _, t := range tasks {
		if t == nil || NormalizeStatusToTitleCase(t.Status) != models.StatusInProgress {
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

// conflictStatusAllow returns true for tasks that participate in file/forbidden conflict detection.
// Default (includeTodo false): In Progress only. Preflight (includeTodo true): Todo + In Progress.
func conflictStatusAllow(norm string, includeTodo bool) bool {
	if norm == models.StatusInProgress {
		return true
	}

	return includeTodo && norm == models.StatusTodo
}

// DetectFileConflicts returns file-level conflicts: In Progress tasks that list the same file(s).
// File paths are extracted from long_description (Files/Components patterns).
func DetectFileConflicts(tasks []*database.Todo2Task) []FileConflict {
	return DetectFileConflictsWithPreflight(tasks, false)
}

// DetectFileConflictsWithPreflight is like DetectFileConflicts; when includeTodo is true, Todo tasks
// are included so planners can see overlapping file claims before work starts.
func DetectFileConflictsWithPreflight(tasks []*database.Todo2Task, includeTodo bool) []FileConflict {
	taskFiles := make(map[string][]string)
	taskStatus := make(map[string]string)

	for _, t := range tasks {
		if t == nil {
			continue
		}

		st := NormalizeStatusToTitleCase(t.Status)
		if !conflictStatusAllow(st, includeTodo) {
			continue
		}

		taskStatus[t.ID] = st

		files := filesFromLongDescription(t.LongDescription)
		if own := models.GetTaskOwnership(t); own != nil {
			if len(own.OwnedFiles) > 0 {
				files = append(files, own.OwnedFiles...)
			}
			if len(own.OwnedGlobs) > 0 {
				// We treat globs as matchers against other tasks' owned_files and globs.
				files = append(files, own.OwnedGlobs...)
			}
		}
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

	// Expand overlaps where a glob key matches another key.
	// Best-effort: uses filepath.Match; safe for small in-progress sets.
	if len(fileToTasks) > 1 {
		keys := make([]string, 0, len(fileToTasks))
		for k := range fileToTasks {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, g := range keys {
			if !strings.ContainsAny(g, "*?[") {
				continue
			}
			for _, k := range keys {
				if g == k {
					continue
				}
				if ok, _ := filepath.Match(g, k); ok {
					fileToTasks[g] = append(fileToTasks[g], fileToTasks[k]...)
				}
			}
		}
	}
	// Group by task set: key = sorted task IDs, value = all shared files
	setToFiles := make(map[string][]string)

	for file, ids := range fileToTasks {
		if len(ids) < 2 {
			continue
		}

		sort.Strings(ids)
		key := strings.Join(ids, "|")
		setToFiles[key] = append(setToFiles[key], file)
	}

	var out []FileConflict

	for key, files := range setToFiles {
		ids := strings.Split(key, "|")
		ts := make(map[string]string, len(ids))
		for _, id := range ids {
			if s, ok := taskStatus[id]; ok {
				ts[id] = s
			}
		}

		fc := FileConflict{TaskIDs: ids, Files: files}
		if len(ts) > 0 {
			fc.TaskStatus = ts
		}

		out = append(out, fc)
	}

	return out
}

func normConflictPath(p string) string {
	p = strings.TrimSpace(p)
	p = filepath.Clean(p)

	return filepath.ToSlash(p)
}

// inProgressConcretePaths returns normalized concrete paths from long_description Files/Components
// and ownership metadata owned_files (not globs).
func inProgressConcretePaths(t *database.Todo2Task) []string {
	return concretePathsForConflict(t, false)
}

func concretePathsForConflict(t *database.Todo2Task, includeTodo bool) []string {
	if t == nil {
		return nil
	}

	st := NormalizeStatusToTitleCase(t.Status)
	if !conflictStatusAllow(st, includeTodo) {
		return nil
	}

	seen := make(map[string]bool)

	var out []string

	add := func(p string) {
		p = normConflictPath(p)
		if p == "" || seen[p] {
			return
		}

		seen[p] = true
		out = append(out, p)
	}

	for _, f := range filesFromLongDescription(t.LongDescription) {
		add(f)
	}

	if own := models.GetTaskOwnership(t); own != nil {
		for _, f := range own.OwnedFiles {
			add(f)
		}
	}

	return out
}

// pathForbiddenByOwnership returns (true, reason) if path p is forbidden by task other's ownership rules.
func pathForbiddenByOwnership(other *database.Todo2Task, p string) (bool, string) {
	if other == nil {
		return false, ""
	}

	own := models.GetTaskOwnership(other)
	if own == nil || len(own.ForbiddenFiles) == 0 {
		return false, ""
	}

	p = normConflictPath(p)
	if p == "" {
		return false, ""
	}

	for _, fb := range own.ForbiddenFiles {
		fb = normConflictPath(fb)
		if fb == "" {
			continue
		}

		if strings.ContainsAny(fb, "*?[") {
			if ok, _ := filepath.Match(fb, p); ok {
				return true, "forbidden_glob"
			}
		} else if fb == p {
			return true, "forbidden_file"
		}
	}

	return false, ""
}

// DetectForbiddenOwnershipConflicts returns ordered pairs (task_id, other_task_id, path) where one
// In Progress task's concrete path conflicts with another's forbidden_files.
func DetectForbiddenOwnershipConflicts(tasks []*database.Todo2Task) []ForbiddenOwnershipConflict {
	return DetectForbiddenOwnershipConflictsWithPreflight(tasks, false)
}

// DetectForbiddenOwnershipConflictsWithPreflight includes Todo tasks when includeTodo is true.
func DetectForbiddenOwnershipConflictsWithPreflight(tasks []*database.Todo2Task, includeTodo bool) []ForbiddenOwnershipConflict {
	var relevant []*database.Todo2Task

	for _, t := range tasks {
		if t != nil && conflictStatusAllow(NormalizeStatusToTitleCase(t.Status), includeTodo) {
			relevant = append(relevant, t)
		}
	}

	if len(relevant) < 2 {
		return nil
	}

	pathsByID := make(map[string][]string, len(relevant))

	for _, t := range relevant {
		pathsByID[t.ID] = concretePathsForConflict(t, includeTodo)
	}

	var out []ForbiddenOwnershipConflict

	for _, a := range relevant {
		for _, b := range relevant {
			if a.ID == b.ID {
				continue
			}

			for _, p := range pathsByID[a.ID] {
				if ok, reason := pathForbiddenByOwnership(b, p); ok {
					out = append(out, ForbiddenOwnershipConflict{
						TaskID:      a.ID,
						OtherTaskID: b.ID,
						Path:        p,
						Reason:      reason,
					})
				}
			}
		}
	}

	return out
}

// DetectConflicts loads tasks from the store and returns task-overlap, file-level, and
// forbidden-ownership conflicts.
func DetectConflicts(ctx context.Context, projectRoot string) (taskOverlaps []TaskOverlapConflict, fileConflicts []FileConflict, forbidden []ForbiddenOwnershipConflict, err error) {
	store := NewDefaultTaskStore(projectRoot)

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, nil, nil, err
	}

	taskOverlaps = DetectTaskOverlapConflicts(list)
	fileConflicts = DetectFileConflicts(list)
	forbidden = DetectForbiddenOwnershipConflicts(list)

	return taskOverlaps, fileConflicts, forbidden, nil
}
