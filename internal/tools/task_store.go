// task_store.go — Task store interface and context helpers (DB-first, JSON fallback).
package tools

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/davidl71/exarp-go/internal/database"
)

// dbOrFileStore implements database.TaskStore with DB-first, JSON-file fallback.
type dbOrFileStore struct {
	projectRoot string
}

// NewDefaultTaskStore returns a TaskStore that uses the database when available,
// otherwise falls back to .todo2/state.todo2.json. Pass the project root (e.g. from FindProjectRoot).
func NewDefaultTaskStore(projectRoot string) database.TaskStore {
	return &dbOrFileStore{projectRoot: projectRoot}
}

func (s *dbOrFileStore) GetTask(ctx context.Context, id string) (*database.Todo2Task, error) {
	if db, err := database.GetDB(); err == nil && db != nil {
		return database.GetTask(ctx, id)
	}

	tasks, err := LoadTodo2Tasks(s.projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	for i := range tasks {
		if tasks[i].ID == id {
			return &tasks[i], nil
		}
	}

	return nil, fmt.Errorf("task %s not found", id)
}

func (s *dbOrFileStore) UpdateTask(ctx context.Context, task *database.Todo2Task) error {
	if task == nil || task.ID == "" {
		return fmt.Errorf("task and task.ID are required")
	}

	if db, err := database.GetDB(); err == nil && db != nil {
		if err := database.UpdateTask(ctx, task); err != nil {
			return err
		}

		return SyncTodo2Tasks(s.projectRoot)
	}

	tasks, err := LoadTodo2Tasks(s.projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	for i := range tasks {
		if tasks[i].ID == task.ID {
			tasks[i] = *task
			return SaveTodo2Tasks(s.projectRoot, tasks)
		}
	}

	return fmt.Errorf("task %s not found", task.ID)
}

func (s *dbOrFileStore) ListTasks(ctx context.Context, filters *database.TaskFilters) ([]*database.Todo2Task, error) {
	projectID := filepath.Base(s.projectRoot)
	if projectID == "" || projectID == "." {
		projectID = "default"
	}
	// Default list to current project so multiple projects using the same DB don't clobber each other.
	if filters == nil {
		filters = &database.TaskFilters{ProjectID: &projectID}
	} else if filters.ProjectID == nil {
		f2 := *filters
		f2.ProjectID = &projectID
		filters = &f2
	}

	if db, err := database.GetDB(); err == nil && db != nil {
		return database.ListTasks(ctx, filters)
	}

	tasks, err := LoadTodo2Tasks(s.projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	return filterTasksToPtrs(tasks, filters), nil
}

func (s *dbOrFileStore) CreateTask(ctx context.Context, task *database.Todo2Task) error {
	if task == nil {
		return fmt.Errorf("task is required")
	}

	// Tag task with current project so tasks don't clobber across projects and can be aggregated.
	if task.ProjectID == "" {
		task.ProjectID = filepath.Base(s.projectRoot)
		if task.ProjectID == "" || task.ProjectID == "." {
			task.ProjectID = "default"
		}
	}

	if db, err := database.GetDB(); err == nil && db != nil {
		return database.CreateTask(ctx, task)
	}

	tasks, err := LoadTodo2Tasks(s.projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	tasks = append(tasks, *task)

	return SaveTodo2Tasks(s.projectRoot, tasks)
}

func (s *dbOrFileStore) DeleteTask(ctx context.Context, id string) error {
	if db, err := database.GetDB(); err == nil && db != nil {
		return database.DeleteTask(ctx, id)
	}

	tasks, err := LoadTodo2Tasks(s.projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	var kept []Todo2Task

	for _, t := range tasks {
		if t.ID != id {
			kept = append(kept, t)
		}
	}

	if len(kept) == len(tasks) {
		return fmt.Errorf("task %s not found", id)
	}

	return SaveTodo2Tasks(s.projectRoot, kept)
}

// tasksFromPtrs converts []*Todo2Task from TaskStore.ListTasks to []Todo2Task for helpers
// that expect a value slice (e.g. getTasksSummaryFromTasks, BacklogExecutionOrder).
func tasksFromPtrs(pts []*database.Todo2Task) []Todo2Task {
	if pts == nil {
		return nil
	}

	out := make([]Todo2Task, 0, len(pts))

	for _, p := range pts {
		if p != nil {
			out = append(out, *p)
		}
	}

	return out
}

// filterTasksToPtrs filters a slice of tasks by database.TaskFilters and returns pointers.
func filterTasksToPtrs(tasks []Todo2Task, filters *database.TaskFilters) []*database.Todo2Task {
	if filters == nil {
		out := make([]*database.Todo2Task, len(tasks))
		for i := range tasks {
			out[i] = &tasks[i]
		}

		return out
	}

	var out []*database.Todo2Task

	for i := range tasks {
		t := &tasks[i]
		if filters.Status != nil && t.Status != *filters.Status {
			continue
		}

		if filters.Priority != nil && t.Priority != *filters.Priority {
			continue
		}

		if filters.Tag != nil {
			found := false

			for _, tag := range t.Tags {
				if tag == *filters.Tag {
					found = true
					break
				}
			}

			if !found {
				continue
			}
		}

		if filters.ProjectID != nil && t.ProjectID != *filters.ProjectID {
			continue
		}

		out = append(out, t)
	}

	return out
}
