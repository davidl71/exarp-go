// todo2_sync.go — Strategy-aware sync between SQLite and JSON task stores.
//
// The original SyncTodo2Tasks used a hard-coded "DB wins" strategy.
// This file adds:
//   - SyncConflictStrategy with four modes
//   - SyncConflict / SyncReport for structured conflict reporting
//   - SyncTodo2TasksWithStrategy — strategy-aware sync (wraps the existing flow)
//   - ImportFromJSON — import tasks from an arbitrary JSON file with conflict control
package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/davidl71/exarp-go/internal/models"
)

// SyncConflictStrategy determines how to resolve conflicting task versions
// (same ID present in both SQLite and JSON with different content).
type SyncConflictStrategy int

const (
	// StrategyDBWins discards JSON changes in favour of the SQLite version.
	// This is the original / default behaviour.
	StrategyDBWins SyncConflictStrategy = iota

	// StrategyJSONWins replaces the SQLite version entirely with the JSON version.
	// Useful for restoring tasks from a backup.
	StrategyJSONWins

	// StrategyMergePreferDB starts from the DB record and fills any empty fields
	// from the JSON record (no JSON field can overwrite a non-empty DB field).
	StrategyMergePreferDB

	// StrategyMergePreferJSON starts from the JSON record but preserves the DB's
	// runtime-state fields (Status, CompletedAt, AssignedTo).
	// Useful for re-importing original task descriptions while keeping current progress.
	StrategyMergePreferJSON
)

// ParseSyncConflictStrategy converts a string to a SyncConflictStrategy.
// Accepts: "db_wins", "json_wins", "merge_prefer_db", "merge_prefer_json".
// Returns StrategyDBWins and an error for unknown strings.
func ParseSyncConflictStrategy(s string) (SyncConflictStrategy, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "db_wins", "db":
		return StrategyDBWins, nil
	case "json_wins", "json":
		return StrategyJSONWins, nil
	case "merge_prefer_db", "merge_db":
		return StrategyMergePreferDB, nil
	case "merge_prefer_json", "merge_json":
		return StrategyMergePreferJSON, nil
	default:
		return StrategyDBWins, fmt.Errorf("unknown conflict strategy %q: use db_wins|json_wins|merge_prefer_db|merge_prefer_json", s)
	}
}

func (s SyncConflictStrategy) String() string {
	switch s {
	case StrategyDBWins:
		return "db_wins"
	case StrategyJSONWins:
		return "json_wins"
	case StrategyMergePreferDB:
		return "merge_prefer_db"
	case StrategyMergePreferJSON:
		return "merge_prefer_json"
	default:
		return "unknown"
	}
}

// SyncConflict describes a single task conflict detected during sync.
type SyncConflict struct {
	TaskID     string `json:"task_id"`
	DBHash     string `json:"db_hash,omitempty"`
	JSONHash   string `json:"json_hash,omitempty"`
	Resolution string `json:"resolution"` // "db_wins", "json_wins", "merged_prefer_db", "merged_prefer_json"
}

// SyncReport is returned by SyncTodo2TasksWithStrategy and ImportFromJSON.
// It replaces the previous stderr-only conflict warnings with a structured result.
type SyncReport struct {
	Strategy  string         `json:"strategy"`
	Conflicts []SyncConflict `json:"conflicts,omitempty"`
	Created   int            `json:"created"`   // tasks created in DB from JSON-only entries
	Updated   int            `json:"updated"`   // existing tasks updated due to strategy
	Unchanged int            `json:"unchanged"` // tasks identical in both sources
	DBOnly    int            `json:"db_only"`   // tasks in DB not in JSON (left untouched)
	JSONOnly  int            `json:"json_only"` // tasks in JSON not in DB (imported as new)
	Errors    []string       `json:"errors,omitempty"`
}

// SyncTodo2TasksWithStrategy is like SyncTodo2Tasks but accepts a conflict resolution
// strategy and returns a structured SyncReport instead of writing warnings to stderr.
//
// SyncTodo2Tasks (backward-compat) calls this with StrategyDBWins.
func SyncTodo2TasksWithStrategy(projectRoot string, strategy SyncConflictStrategy) (*SyncReport, error) {
	report := &SyncReport{Strategy: strategy.String()}

	// Load from both sources
	dbTasks, dbErr := loadTodo2TasksFromDB(projectRoot)
	jsonTasks, _ := loadTodo2TasksFromJSON(projectRoot)

	// Index both sources by task ID
	dbByID := make(map[string]Todo2Task, len(dbTasks))
	for _, t := range dbTasks {
		dbByID[t.ID] = t
	}
	jsonByID := make(map[string]Todo2Task, len(jsonTasks))
	for _, t := range jsonTasks {
		jsonByID[t.ID] = t
	}

	// Build the merged task map using the chosen strategy
	taskMap := make(map[string]Todo2Task, len(dbByID)+len(jsonByID))

	// Pass 1: all JSON tasks
	for id, jt := range jsonByID {
		taskMap[id] = jt
	}

	// Pass 2: DB tasks — either override JSON or merge, depending on strategy
	for id, dt := range dbByID {
		jt, inJSON := jsonByID[id]
		if !inJSON {
			// Only in DB — keep as-is
			taskMap[id] = dt
			report.DBOnly++
			continue
		}

		// Present in both; check for conflict
		dbHash := models.GetContentHash(&dt)
		jHash := models.GetContentHash(&jt)

		if dbHash == jHash || dbHash == "" || jHash == "" {
			// No conflict (same content or hash not computed)
			taskMap[id] = dt
			report.Unchanged++
			continue
		}

		// Genuine conflict: apply strategy
		merged := mergeTask(dt, jt, strategy)
		models.SetContentHash(&merged)
		taskMap[id] = merged

		resolution := conflictResolutionLabel(strategy)
		report.Conflicts = append(report.Conflicts, SyncConflict{
			TaskID:     id,
			DBHash:     truncateHash(dbHash),
			JSONHash:   truncateHash(jHash),
			Resolution: resolution,
		})
		report.Updated++
	}

	// Pass 3: JSON-only tasks (not in DB)
	for id, jt := range jsonByID {
		if _, inDB := dbByID[id]; !inDB {
			taskMap[id] = jt
			report.JSONOnly++
			report.Created++
		}
	}

	// Partition into DB tasks (no AUTO-*) and JSON tasks (all)
	projectID := filepath.Base(projectRoot)
	if projectID == "" || projectID == "." {
		projectID = "default"
	}

	dbTasksToSave := make([]Todo2Task, 0, len(taskMap))
	jsonTasksForSave := make([]Todo2Task, 0, len(taskMap))

	for _, task := range taskMap {
		if strings.HasPrefix(task.ID, "AUTO-") {
			jsonTasksForSave = append(jsonTasksForSave, task)
		} else {
			if task.ProjectID == "" || task.ProjectID == projectID {
				dbTasksToSave = append(dbTasksToSave, task)
			}
			jsonTasksForSave = append(jsonTasksForSave, task)
		}
	}

	// Save to DB when available
	if dbErr == nil {
		if err := cleanupAutoTasksFromDB(); err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("auto-task cleanup: %v", err))
		}
		if err := saveTodo2TasksToDB(projectRoot, dbTasksToSave); err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("db save: %v", err))
		}
	}

	// Always save JSON
	if err := saveTodo2TasksToJSON(projectRoot, jsonTasksForSave); err != nil {
		report.Errors = append(report.Errors, fmt.Sprintf("json save: %v", err))
		if dbErr != nil {
			return report, fmt.Errorf("failed to save to both sources: %w", err)
		}
		return report, fmt.Errorf("failed to save to JSON: %w", err)
	}

	return report, nil
}

// ImportFromJSON loads tasks from an arbitrary JSON file and reconciles them with the
// current SQLite store using the given conflict strategy.
//
// Use cases:
//   - Restoring from a backup JSON file
//   - Re-importing tasks exported from another project
//   - Migrating from a JSON-only setup to SQLite
//
// The project's .todo2/state.todo2.json is updated so subsequent syncs stay consistent.
func ImportFromJSON(projectRoot, jsonPath string, strategy SyncConflictStrategy) (*SyncReport, error) {
	report := &SyncReport{Strategy: strategy.String()}

	// Read and parse the source JSON
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read import file %s: %w", jsonPath, err)
	}
	importedTasks, err := ParseTasksFromJSON(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse import file %s: %w", jsonPath, err)
	}

	// Backfill content hashes on imported tasks
	for i := range importedTasks {
		models.EnsureContentHash(&importedTasks[i])
	}

	// Load current DB
	dbTasks, dbErr := loadTodo2TasksFromDB(projectRoot)
	dbByID := make(map[string]Todo2Task, len(dbTasks))
	for _, t := range dbTasks {
		dbByID[t.ID] = t
	}

	// Build merged task map
	taskMap := make(map[string]Todo2Task, len(dbByID)+len(importedTasks))

	// Seed with all DB tasks first
	for id, dt := range dbByID {
		taskMap[id] = dt
	}

	// Apply imported tasks according to strategy
	for _, jt := range importedTasks {
		dt, inDB := dbByID[jt.ID]
		if !inDB {
			// New task — import it
			taskMap[jt.ID] = jt
			report.Created++
			report.JSONOnly++
			continue
		}

		dbHash := models.GetContentHash(&dt)
		jHash := models.GetContentHash(&jt)

		if dbHash == jHash || dbHash == "" || jHash == "" {
			report.Unchanged++
			continue
		}

		// Conflict
		merged := mergeTask(dt, jt, strategy)
		models.SetContentHash(&merged)
		taskMap[jt.ID] = merged

		report.Conflicts = append(report.Conflicts, SyncConflict{
			TaskID:     jt.ID,
			DBHash:     truncateHash(dbHash),
			JSONHash:   truncateHash(jHash),
			Resolution: conflictResolutionLabel(strategy),
		})
		report.Updated++
	}

	// Tasks in DB not in the imported file
	importedIDs := make(map[string]bool, len(importedTasks))
	for _, t := range importedTasks {
		importedIDs[t.ID] = true
	}
	for id := range dbByID {
		if !importedIDs[id] {
			report.DBOnly++
		}
	}

	// Partition and save
	projectID := filepath.Base(projectRoot)
	if projectID == "" || projectID == "." {
		projectID = "default"
	}

	dbTasksToSave := make([]Todo2Task, 0, len(taskMap))
	jsonTasksForSave := make([]Todo2Task, 0, len(taskMap))

	for _, task := range taskMap {
		if strings.HasPrefix(task.ID, "AUTO-") {
			jsonTasksForSave = append(jsonTasksForSave, task)
		} else {
			if task.ProjectID == "" || task.ProjectID == projectID {
				dbTasksToSave = append(dbTasksToSave, task)
			}
			jsonTasksForSave = append(jsonTasksForSave, task)
		}
	}

	if dbErr == nil {
		if err := cleanupAutoTasksFromDB(); err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("auto-task cleanup: %v", err))
		}
		if err := saveTodo2TasksToDB(projectRoot, dbTasksToSave); err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("db save: %v", err))
		}
	}

	if err := saveTodo2TasksToJSON(projectRoot, jsonTasksForSave); err != nil {
		report.Errors = append(report.Errors, fmt.Sprintf("json save: %v", err))
		return report, fmt.Errorf("failed to save merged state to JSON: %w", err)
	}

	return report, nil
}

// ─── merge helpers ────────────────────────��────────────────────���─────────────

// mergeTask produces the resolved task from a DB version and a JSON version
// according to the given strategy.
func mergeTask(dbTask, jsonTask Todo2Task, strategy SyncConflictStrategy) Todo2Task {
	switch strategy {
	case StrategyDBWins:
		return dbTask

	case StrategyJSONWins:
		return jsonTask

	case StrategyMergePreferDB:
		// Start from DB; JSON fills empty fields only.
		result := dbTask
		if result.Content == "" {
			result.Content = jsonTask.Content
		}
		if result.LongDescription == "" {
			result.LongDescription = jsonTask.LongDescription
		}
		if len(result.Tags) == 0 {
			result.Tags = jsonTask.Tags
		}
		if len(result.Dependencies) == 0 {
			result.Dependencies = jsonTask.Dependencies
		}
		if result.Priority == "" {
			result.Priority = jsonTask.Priority
		}
		// Merge metadata: JSON as base, DB overrides
		result.Metadata = mergeMetadata(jsonTask.Metadata, dbTask.Metadata)
		return result

	case StrategyMergePreferJSON:
		// Start from JSON; preserve DB's runtime-state fields.
		result := jsonTask
		// DB runtime state takes precedence (tracks actual progress)
		if dbTask.Status != "" {
			result.Status = dbTask.Status
		}
		if dbTask.CompletedAt != "" {
			result.CompletedAt = dbTask.CompletedAt
		}
		if dbTask.AssignedTo != "" {
			result.AssignedTo = dbTask.AssignedTo
		}
		// DB's last-modified is more recent (runtime activity)
		if dbTask.LastModified != "" {
			result.LastModified = dbTask.LastModified
		}
		// Merge metadata: DB as base, JSON overrides (JSON has the richer description context)
		result.Metadata = mergeMetadata(dbTask.Metadata, jsonTask.Metadata)
		return result
	}

	return dbTask
}

// mergeMetadata merges two metadata maps. base is the starting point; override keys win.
func mergeMetadata(base, override map[string]interface{}) map[string]interface{} {
	if len(base) == 0 && len(override) == 0 {
		return nil
	}
	out := make(map[string]interface{}, len(base)+len(override))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range override {
		out[k] = v
	}
	return out
}

func conflictResolutionLabel(strategy SyncConflictStrategy) string {
	switch strategy {
	case StrategyDBWins:
		return "db_wins"
	case StrategyJSONWins:
		return "json_wins"
	case StrategyMergePreferDB:
		return "merged_prefer_db"
	case StrategyMergePreferJSON:
		return "merged_prefer_json"
	default:
		return "db_wins"
	}
}

func truncateHash(h string) string {
	if len(h) > 12 {
		return h[:12]
	}
	return h
}
