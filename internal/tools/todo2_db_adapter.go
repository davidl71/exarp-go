// todo2_db_adapter.go — SQLite adapter for Todo2 task storage (DB-first with JSON fallback).
package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/models"
)

// loadTodo2TasksFromDB loads tasks from database.
func loadTodo2TasksFromDB() ([]Todo2Task, error) {
	if db, err := database.GetDB(); err != nil || db == nil {
		return nil, fmt.Errorf("database not available")
	}

	tasks, err := database.ListTasks(context.Background(), nil)
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

// saveTodo2TasksToDB saves tasks to database
// Matches tasks by ID first, then by content/description if ID doesn't match
// This handles cases where same task was created with different epoch-based IDs
// Sorts tasks by dependency order to ensure dependencies are created before dependents.
func saveTodo2TasksToDB(tasks []Todo2Task) error {
	if db, err := database.GetDB(); err != nil || db == nil {
		return fmt.Errorf("database not available")
	}

	// Sort tasks by dependency order to avoid foreign key constraint failures
	// Dependencies will be created before dependents
	sortedTasks, err := sortTasksByDependencies(tasks)
	if err != nil {
		// If sorting fails (e.g., cycles), proceed with original order
		// Missing dependencies will be handled gracefully
		sortedTasks = tasks
	}

	ctx := context.Background()

	var errors []string

	successCount := 0

	// Track tasks that need dependency updates (created without dependencies due to FK constraints)
	tasksNeedingDeps := make(map[string][]string) // task ID -> original dependencies

	for _, task := range sortedTasks {
		models.SetContentHash(&task)

		var saveErr error

		// First, try to find by ID
		existing, err := database.GetTask(ctx, task.ID)
		if err != nil {
			// Task doesn't exist by ID
			// Only check content match for non-AUTO tasks (AUTO tasks are unique executions)
			// Content matching is for detecting same task created with different epoch-based IDs
			// AUTO tasks should always be created as new tasks even if content matches
			var matchedTask *Todo2Task
			if !strings.HasPrefix(task.ID, "AUTO-") {
				matchedTask = findTaskByContent(ctx, task)
			}

			if matchedTask != nil {
				// Found matching task by content - update it with new ID and data
				// Use the existing task's ID but update with new data
				task.ID = matchedTask.ID
				if err := database.UpdateTask(ctx, &task); err != nil {
					saveErr = fmt.Errorf("failed to update matched task %s (content match): %w", matchedTask.ID, err)
				} else {
					successCount++
					continue
				}
			} else {
				// Task doesn't exist at all, create it
				// Filter out invalid dependencies (dependencies that don't exist in the task list or database)
				// This prevents foreign key constraint failures
				// Also check if dependencies have been created in this batch (for tasks being created now)
				validDeps := filterValidDependencies(ctx, task.Dependencies, sortedTasks)
				if len(validDeps) < len(task.Dependencies) {
					// Some dependencies were filtered out - log warning
					errors = append(errors, fmt.Sprintf("warning: task %s has invalid dependencies, filtered %d/%d dependencies", task.ID, len(task.Dependencies)-len(validDeps), len(task.Dependencies)))
				}

				task.Dependencies = validDeps

				if err := database.CreateTask(ctx, &task); err != nil {
					// Check if error is due to foreign key constraint
					if strings.Contains(err.Error(), "FOREIGN KEY constraint") || strings.Contains(err.Error(), "foreign key") {
						// Foreign key constraint failed - dependencies don't exist yet
						// Try creating task without dependencies first, then add dependencies in second pass
						originalDeps := task.Dependencies
						task.Dependencies = []string{} // Create without dependencies first

						if createErr := database.CreateTask(ctx, &task); createErr != nil {
							saveErr = fmt.Errorf("failed to create task %s (even without dependencies): %w", task.ID, createErr)
						} else {
							// Task created successfully - store original dependencies for second pass
							if len(originalDeps) > 0 {
								tasksNeedingDeps[task.ID] = originalDeps
							}

							successCount++

							continue
						}
					} else {
						saveErr = fmt.Errorf("failed to create task %s: %w", task.ID, err)
					}
				} else {
					successCount++
					continue
				}
			}
		} else {
			// Task exists by ID, update it
			if existing != nil {
				// Verify content matches before updating (safety check)
				if !tasksMatchByContent(*existing, task) {
					// Content doesn't match - this might be a different task with same ID
					// Log warning but still update (ID takes precedence)
					errors = append(errors, fmt.Sprintf("warning: task %s content mismatch (ID match but content differs)", task.ID))
				}

				if err := database.UpdateTask(ctx, &task); err != nil {
					saveErr = fmt.Errorf("failed to update task %s: %w", task.ID, err)
				} else {
					successCount++
					continue
				}
			}
		}

		if saveErr != nil {
			errors = append(errors, saveErr.Error())
			// Continue processing remaining tasks instead of stopping on first error
		}
	}

	// Second pass: Add dependencies to tasks that were created without them
	// This ensures all tasks exist before we try to add dependencies
	for taskID, originalDeps := range tasksNeedingDeps {
		// Reload task to get current state
		task, err := database.GetTask(ctx, taskID)
		if err != nil || task == nil {
			errors = append(errors, fmt.Sprintf("warning: task %s not found for dependency update", taskID))
			continue
		}

		// Filter dependencies to only include valid ones (now that all tasks should be created)
		validDeps := filterValidDependencies(ctx, originalDeps, sortedTasks)
		if len(validDeps) < len(originalDeps) {
			errors = append(errors, fmt.Sprintf("warning: task %s has %d invalid dependencies (filtered out)", taskID, len(originalDeps)-len(validDeps)))
		}

		// Update task with dependencies
		task.Dependencies = validDeps
		if err := database.UpdateTask(ctx, task); err != nil {
			errors = append(errors, fmt.Sprintf("warning: task %s created but dependencies could not be added: %v", taskID, err))
		}
	}

	// Remove from DB any tasks not in the input list (replace semantics, e.g. after merge)
	inputIDs := make(map[string]bool)

	for _, t := range tasks {
		if !strings.HasPrefix(t.ID, "AUTO-") {
			inputIDs[t.ID] = true
		}
	}

	allDB, err := database.ListTasks(ctx, nil)
	if err == nil {
		for _, t := range allDB {
			if t != nil && !inputIDs[t.ID] {
				if delErr := database.DeleteTask(ctx, t.ID); delErr != nil {
					errors = append(errors, fmt.Sprintf("warning: failed to delete obsolete task %s: %v", t.ID, delErr))
				}
			}
		}
	}

	// Return error if any failures occurred, but include success count
	if len(errors) > 0 {
		return fmt.Errorf("failed to save %d/%d tasks. Success: %d, Errors: %v", len(errors), len(tasks), successCount, errors)
	}

	return nil
}

// findTaskByContent searches for a task with matching content/description
// Used to detect duplicate tasks with different IDs (e.g., epoch-based IDs).
func findTaskByContent(ctx context.Context, task Todo2Task) *Todo2Task {
	if db, err := database.GetDB(); err != nil || db == nil {
		return nil
	}

	// Load all tasks and check content match
	allTasks, err := database.ListTasks(ctx, nil)
	if err != nil {
		return nil
	}

	for _, existing := range allTasks {
		if tasksMatchByContent(*existing, task) {
			return existing
		}
	}

	return nil
}

// tasksMatchByContent checks if two tasks have matching content/description
// Used to detect duplicates with different IDs.
func tasksMatchByContent(task1, task2 Todo2Task) bool {
	// Compare normalized content
	content1 := normalizeTaskContent(task1.Content, task1.LongDescription)
	content2 := normalizeTaskContent(task2.Content, task2.LongDescription)

	return content1 == content2
}

// normalizeTaskContent normalizes task content for comparison.
// Delegates to models.NormalizeForComparison for shared normalization logic.
func normalizeTaskContent(content, description string) string {
	return models.NormalizeForComparison(content, description)
}

// filterValidDependencies filters out dependencies that don't exist in the task list or database
// Also filters out AUTO-* task dependencies (AUTO tasks are not saved to database)
// This prevents foreign key constraint failures when creating tasks.
func filterValidDependencies(ctx context.Context, dependencies []string, allTasks []Todo2Task) []string {
	// Build set of valid task IDs (from task list, excluding AUTO-* tasks)
	validTaskIDs := make(map[string]bool)

	for _, task := range allTasks {
		// Skip AUTO-* tasks - they're not saved to database
		if !strings.HasPrefix(task.ID, "AUTO-") {
			validTaskIDs[task.ID] = true
		}
	}

	// Also check database for existing tasks (excluding AUTO-* tasks)
	if db, err := database.GetDB(); err == nil && db != nil {
		for _, depID := range dependencies {
			// Skip AUTO-* dependencies - they're not in database
			if strings.HasPrefix(depID, "AUTO-") {
				continue
			}

			if !validTaskIDs[depID] {
				// Check if dependency exists in database
				if existing, err := database.GetTask(ctx, depID); err == nil && existing != nil {
					validTaskIDs[depID] = true
				}
			}
		}
	}

	// Filter dependencies to only include valid ones (excluding AUTO-* tasks)
	validDeps := make([]string, 0, len(dependencies))

	for _, depID := range dependencies {
		// Skip AUTO-* dependencies - they're not saved to database
		if strings.HasPrefix(depID, "AUTO-") {
			continue
		}

		if validTaskIDs[depID] {
			validDeps = append(validDeps, depID)
		}
	}

	return validDeps
}

// sortTasksByDependencies sorts tasks by dependency order using topological sort
// Tasks with no dependencies come first, then tasks that depend on them, etc.
// This ensures dependencies are created before dependents, avoiding foreign key constraint failures.
func sortTasksByDependencies(tasks []Todo2Task) ([]Todo2Task, error) {
	if len(tasks) == 0 {
		return tasks, nil
	}

	// Build task graph
	tg, err := BuildTaskGraph(tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to build task graph: %w", err)
	}

	// Check for cycles - if cycles exist, we can't sort properly
	hasCycles, err := HasCycles(tg)
	if err != nil {
		return nil, fmt.Errorf("failed to check for cycles: %w", err)
	}

	if hasCycles {
		// Graph has cycles - can't sort, return original order
		// Missing dependencies will be handled gracefully in save loop
		return tasks, fmt.Errorf("task graph has cycles, cannot sort by dependencies")
	}

	// Get topological sort (dependency order)
	sortedIDs, err := TopoSortTasks(tg)
	if err != nil {
		return nil, fmt.Errorf("failed to sort tasks: %w", err)
	}

	// Build map for quick lookup
	taskMap := make(map[string]Todo2Task)
	for _, task := range tasks {
		taskMap[task.ID] = task
	}

	// Reorder tasks according to topological sort
	sortedTasks := make([]Todo2Task, 0, len(tasks))

	for _, taskID := range sortedIDs {
		if task, exists := taskMap[taskID]; exists {
			sortedTasks = append(sortedTasks, task)
		}
	}

	// Add any tasks that weren't in the graph (shouldn't happen, but safety check)
	for _, task := range tasks {
		if _, found := taskMap[task.ID]; found {
			// Already added
			continue
		}
		// Task not in graph - add at end
		sortedTasks = append(sortedTasks, task)
	}

	return sortedTasks, nil
}
