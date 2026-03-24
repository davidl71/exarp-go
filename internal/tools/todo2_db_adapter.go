// todo2_db_adapter.go — SQLite adapter for Todo2 task storage (DB-first with JSON fallback).
package tools

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/models"
)

// loadTodo2TasksFromDB loads tasks from database (scoped to project when projectRoot is set).
func loadTodo2TasksFromDB(projectRoot string) ([]Todo2Task, error) {
	if db, err := database.GetDB(); err != nil || db == nil {
		return nil, fmt.Errorf("database not available")
	}

	ctx := context.Background()
	var filters *database.TaskFilters
	if projectRoot != "" {
		projectID := filepath.Base(projectRoot)
		if projectID != "" && projectID != "." {
			filters = &database.TaskFilters{ProjectID: &projectID}
		}
	}
	tasks, err := database.ListTasks(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks from database: %w", err)
	}

	// Convert []*Todo2Task to []Todo2Task and backfill content hash
	result := make([]Todo2Task, len(tasks))
	for i, task := range tasks {
		models.EnsureContentHash(task)
		result[i] = *task
	}

	return result, nil
}

// saveTodo2TasksToDB saves tasks to database (projectRoot used to scope obsolete-task cleanup).
// Matches tasks by ID first, then by content/description if ID doesn't match.
// Sorts tasks by dependency order to ensure dependencies are created before dependents.
//
// Performance: loads existing DB tasks once up front and builds ID+content maps to avoid
// N GetTask/ListTasks calls inside the loop. Obsolete tasks are removed via BatchDeleteTasks.
func saveTodo2TasksToDB(projectRoot string, tasks []Todo2Task) error {
	if db, err := database.GetDB(); err != nil || db == nil {
		return fmt.Errorf("database not available")
	}

	// Sort tasks by dependency order to avoid foreign key constraint failures
	sortedTasks, err := sortTasksByDependencies(tasks)
	if err != nil {
		sortedTasks = tasks
	}

	ctx := context.Background()

	// ── Preload all existing DB tasks once ──────────────────────────────────
	// Build project-scoped filter (same scope used later for obsolete cleanup).
	var listFilters *database.TaskFilters
	if projectRoot != "" {
		pid := filepath.Base(projectRoot)
		if pid != "" && pid != "." {
			listFilters = &database.TaskFilters{ProjectID: &pid}
		}
	}

	existingPtrs, _ := database.ListTasks(ctx, listFilters)

	// Two lookup maps: by ID and by normalised content.
	existingByID := make(map[string]*database.Todo2Task, len(existingPtrs))
	existingByContent := make(map[string]*database.Todo2Task, len(existingPtrs))
	for _, t := range existingPtrs {
		models.EnsureContentHash(t)
		existingByID[t.ID] = t
		key := normalizeTaskContent(t.Content, t.LongDescription)
		if key != "" {
			existingByContent[key] = t
		}
	}
	// ────────────────────────────────────────────────────────────────────────

	var errs []string
	successCount := 0

	// Track tasks that need dependency updates (created without deps due to FK constraints)
	tasksNeedingDeps := make(map[string][]string)

	for _, task := range sortedTasks {
		models.SetContentHash(&task)

		var saveErr error

		// First, try to find by ID using preloaded map (no DB round-trip)
		existing := existingByID[task.ID]
		if existing == nil {
			existing, err = loadExistingTask(ctx, task.ID)
			if err != nil {
				saveErr = fmt.Errorf("failed to load existing task %s: %w", task.ID, err)
				if saveErr != nil {
					errs = append(errs, saveErr.Error())
				}
				continue
			}
			if existing != nil {
				existingByID[task.ID] = existing
			}
		}

		if existing == nil {
			// Not found by ID — try content match for non-AUTO tasks
			var matchedTask *database.Todo2Task
			if !strings.HasPrefix(task.ID, "AUTO-") {
				contentKey := normalizeTaskContent(task.Content, task.LongDescription)
				matchedTask = existingByContent[contentKey]
			}

			if matchedTask != nil {
				task.ID = matchedTask.ID
				if err := database.UpdateTask(ctx, &task); err != nil {
					saveErr = fmt.Errorf("failed to update matched task %s (content match): %w", matchedTask.ID, err)
				} else {
					existingByID[task.ID] = &task // keep map current
					successCount++
					continue
				}
			} else {
				// Task does not exist — create it
				validDeps := filterValidDependencies(task.Dependencies, sortedTasks, existingByID)
				if len(validDeps) < len(task.Dependencies) {
					errs = append(errs, fmt.Sprintf("warning: task %s has invalid dependencies, filtered %d/%d", task.ID, len(task.Dependencies)-len(validDeps), len(task.Dependencies)))
				}
				task.Dependencies = validDeps

				if err := database.CreateTask(ctx, &task); err != nil {
					if strings.Contains(err.Error(), "FOREIGN KEY constraint") || strings.Contains(err.Error(), "foreign key") {
						originalDeps := task.Dependencies
						task.Dependencies = []string{}
						if createErr := database.CreateTask(ctx, &task); createErr != nil {
							saveErr = fmt.Errorf("failed to create task %s (even without dependencies): %w", task.ID, createErr)
						} else {
							if len(originalDeps) > 0 {
								tasksNeedingDeps[task.ID] = originalDeps
							}
							existingByID[task.ID] = &task
							successCount++
							continue
						}
					} else if isUniqueConstraintError(err) {
						if updateErr := database.UpdateTask(ctx, &task); updateErr != nil {
							saveErr = fmt.Errorf("failed to update task %s after unique constraint: %w", task.ID, updateErr)
						} else {
							existingByID[task.ID] = &task
							successCount++
							continue
						}
					} else {
						saveErr = fmt.Errorf("failed to create task %s: %w", task.ID, err)
					}
				} else {
					existingByID[task.ID] = &task
					successCount++
					continue
				}
			}
		} else {
			// Task exists by ID — update it
			if !tasksMatchByContent(*existing, task) {
				errs = append(errs, fmt.Sprintf("warning: task %s content mismatch (ID match but content differs)", task.ID))
			}
			if err := database.UpdateTask(ctx, &task); err != nil {
				saveErr = fmt.Errorf("failed to update task %s: %w", task.ID, err)
			} else {
				existingByID[task.ID] = &task
				successCount++
				continue
			}
		}

		if saveErr != nil {
			errs = append(errs, saveErr.Error())
		}
	}

	// Second pass: add dependencies deferred due to FK constraints
	for taskID, originalDeps := range tasksNeedingDeps {
		task, err := database.GetTask(ctx, taskID)
		if err != nil || task == nil {
			errs = append(errs, fmt.Sprintf("warning: task %s not found for dependency update", taskID))
			continue
		}
		validDeps := filterValidDependencies(originalDeps, sortedTasks, existingByID)
		if len(validDeps) < len(originalDeps) {
			errs = append(errs, fmt.Sprintf("warning: task %s has %d invalid dependencies (filtered out)", taskID, len(originalDeps)-len(validDeps)))
		}
		task.Dependencies = validDeps
		if err := database.UpdateTask(ctx, task); err != nil {
			errs = append(errs, fmt.Sprintf("warning: task %s created but dependencies could not be added: %v", taskID, err))
		}
	}

	// Remove obsolete tasks (in DB but not in the input list) via BatchDeleteTasks
	inputIDs := make(map[string]bool, len(tasks))
	for _, t := range tasks {
		if !strings.HasPrefix(t.ID, "AUTO-") {
			inputIDs[t.ID] = true
		}
	}

	var obsoleteIDs []string
	for _, t := range existingPtrs {
		if t != nil && !inputIDs[t.ID] {
			obsoleteIDs = append(obsoleteIDs, t.ID)
		}
	}
	if len(obsoleteIDs) > 0 {
		if _, delErr := database.BatchDeleteTasks(ctx, obsoleteIDs); delErr != nil {
			errs = append(errs, fmt.Sprintf("warning: batch delete obsolete tasks failed: %v", delErr))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to save %d/%d tasks. Success: %d, Errors: %v", len(errs), len(tasks), successCount, errs)
	}
	return nil
}

// tasksMatchByContent checks if two tasks have matching content/description.
func tasksMatchByContent(task1, task2 Todo2Task) bool {
	return normalizeTaskContent(task1.Content, task1.LongDescription) ==
		normalizeTaskContent(task2.Content, task2.LongDescription)
}

// normalizeTaskContent normalizes task content for comparison.
func normalizeTaskContent(content, description string) string {
	return models.NormalizeForComparison(content, description)
}

func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}

func loadExistingTask(ctx context.Context, id string) (*database.Todo2Task, error) {
	task, err := database.GetTask(ctx, id)
	if err != nil {
		if isTaskNotFoundError(err, id) {
			return nil, nil
		}
		return nil, err
	}
	return task, nil
}

func isTaskNotFoundError(err error, id string) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), fmt.Sprintf("task %s not found", id))
}

// filterValidDependencies filters out dependencies that do not exist in the sorted task list or
// the preloaded DB map. Accepts the preloaded existingByID map to avoid per-dep GetTask calls.
func filterValidDependencies(dependencies []string, allTasks []Todo2Task, existingByID map[string]*database.Todo2Task) []string {
	// Build set of valid IDs from the input batch
	validTaskIDs := make(map[string]bool, len(allTasks))
	for _, task := range allTasks {
		if !strings.HasPrefix(task.ID, "AUTO-") {
			validTaskIDs[task.ID] = true
		}
	}

	// Extend with preloaded DB tasks (no extra queries needed)
	for id := range existingByID {
		if !strings.HasPrefix(id, "AUTO-") {
			validTaskIDs[id] = true
		}
	}

	validDeps := make([]string, 0, len(dependencies))
	for _, depID := range dependencies {
		if strings.HasPrefix(depID, "AUTO-") {
			continue
		}
		if validTaskIDs[depID] {
			validDeps = append(validDeps, depID)
		}
	}
	return validDeps
}

// sortTasksByDependencies sorts tasks by dependency order using topological sort.
func sortTasksByDependencies(tasks []Todo2Task) ([]Todo2Task, error) {
	if len(tasks) == 0 {
		return tasks, nil
	}

	tg, err := BuildTaskGraph(tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to build task graph: %w", err)
	}

	hasCycles, err := HasCycles(tg)
	if err != nil {
		return nil, fmt.Errorf("failed to check for cycles: %w", err)
	}
	if hasCycles {
		return tasks, fmt.Errorf("task graph has cycles, cannot sort by dependencies")
	}

	sortedIDs, err := TopoSortTasks(tg)
	if err != nil {
		return nil, fmt.Errorf("failed to sort tasks: %w", err)
	}

	taskMap := make(map[string]Todo2Task, len(tasks))
	for _, task := range tasks {
		taskMap[task.ID] = task
	}

	sortedTasks := make([]Todo2Task, 0, len(tasks))
	for _, taskID := range sortedIDs {
		if task, exists := taskMap[taskID]; exists {
			sortedTasks = append(sortedTasks, task)
		}
	}
	for _, task := range tasks {
		if _, found := taskMap[task.ID]; !found {
			sortedTasks = append(sortedTasks, task)
		}
	}

	return sortedTasks, nil
}
