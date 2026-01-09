package tools

import (
	"context"
	"fmt"

	"github.com/davidl71/exarp-go/internal/database"
)

// loadTodo2TasksFromDB loads tasks from database
func loadTodo2TasksFromDB() ([]Todo2Task, error) {
	if db, err := database.GetDB(); err != nil || db == nil {
		return nil, fmt.Errorf("database not available")
	}

	tasks, err := database.ListTasks(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks from database: %w", err)
	}

	// Convert []*Todo2Task to []Todo2Task
	result := make([]Todo2Task, len(tasks))
	for i, task := range tasks {
		result[i] = *task
	}
	return result, nil
}

// saveTodo2TasksToDB saves tasks to database
func saveTodo2TasksToDB(tasks []Todo2Task) error {
	if db, err := database.GetDB(); err != nil || db == nil {
		return fmt.Errorf("database not available")
	}

	for _, task := range tasks {
		// Check if task exists
		existing, err := database.GetTask(context.Background(), task.ID)
		if err != nil {
			// Task doesn't exist, create it
			if err := database.CreateTask(context.Background(), &task); err != nil {
				return fmt.Errorf("failed to create task %s: %w", task.ID, err)
			}
		} else {
			// Task exists, update it
			if existing != nil {
				if err := database.UpdateTask(context.Background(), &task); err != nil {
					return fmt.Errorf("failed to update task %s: %w", task.ID, err)
				}
			}
		}
	}
	return nil
}

