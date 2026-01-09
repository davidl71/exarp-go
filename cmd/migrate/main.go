package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/models"
)

func main() {
	var (
		projectRoot = flag.String("project-root", "", "Project root directory (default: auto-detect)")
		dryRun      = flag.Bool("dry-run", false, "Dry run: show what would be migrated without actually migrating")
		backup      = flag.Bool("backup", true, "Create backup of JSON file before migration")
	)
	flag.Parse()

	// Find project root
	var root string
	var err error
	if *projectRoot != "" {
		root = *projectRoot
	} else {
		root, err = findProjectRoot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			fmt.Fprintf(os.Stderr, "Please specify --project-root or run from project directory\n")
			os.Exit(1)
		}
	}

	fmt.Printf("Project root: %s\n", root)

	// Load JSON tasks
	jsonPath := filepath.Join(root, ".todo2", "state.todo2.json")
	fmt.Printf("Loading tasks from: %s\n", jsonPath)

	tasks, comments, err := loadJSONState(jsonPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading JSON state: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d tasks and %d comments in JSON file\n", len(tasks), len(comments))

	if len(tasks) == 0 {
		fmt.Println("No tasks to migrate")
		return
	}

	// Initialize database
	fmt.Printf("Initializing database...\n")
	if err := database.Init(root); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	// Check if database already has tasks
	existingTasks, err := database.ListTasks(context.Background(), nil)
	if err == nil && len(existingTasks) > 0 {
		fmt.Printf("Warning: Database already contains %d tasks\n", len(existingTasks))
		fmt.Print("Continue anyway? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Migration cancelled")
			os.Exit(0)
		}
	}

	if *dryRun {
		fmt.Println("\n=== DRY RUN MODE ===")
		fmt.Println()
		
		// Check which tasks would be created vs updated
		createCount := 0
		updateCount := 0
		for _, task := range tasks {
			existing, err := database.GetTask(context.Background(), task.ID)
			if err == nil && existing != nil {
				updateCount++
			} else {
				createCount++
			}
		}
		
		fmt.Printf("Would migrate:\n")
		fmt.Printf("  - %d tasks (would create: %d, would update: %d)\n", len(tasks), createCount, updateCount)
		fmt.Printf("  - %d comments\n", len(comments))
		fmt.Println("\nUse without --dry-run to perform actual migration")
		return
	}

	// Create backup if requested
	if *backup {
		backupPath := jsonPath + ".backup"
		fmt.Printf("Creating backup: %s\n", backupPath)
		data, err := os.ReadFile(jsonPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not create backup: %v\n", err)
		} else {
			if err := os.WriteFile(backupPath, data, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Could not write backup: %v\n", err)
			} else {
				fmt.Printf("Backup created successfully\n")
			}
		}
	}

	// Migrate tasks
	fmt.Printf("\nMigrating tasks to database...\n")
	taskCount := 0
	updateCount := 0
	skipCount := 0
	commentCount := 0
	commentSkipCount := 0

	for _, task := range tasks {
		// Check if task already exists
		existing, err := database.GetTask(context.Background(), task.ID)
		if err == nil && existing != nil {
			// Task exists - update it instead of skipping
			if err := database.UpdateTask(context.Background(), &task); err != nil {
				fmt.Fprintf(os.Stderr, "  Error updating task %s: %v\n", task.ID, err)
				skipCount++
				continue
			}
			updateCount++
			fmt.Printf("  Updated existing task: %s (%s)\n", task.ID, task.Content)
		} else {
			// Task doesn't exist - create it
			if err := database.CreateTask(context.Background(), &task); err != nil {
				fmt.Fprintf(os.Stderr, "  Error creating task %s: %v\n", task.ID, err)
				skipCount++
				continue
			}
			taskCount++
			fmt.Printf("  Created new task: %s (%s)\n", task.ID, task.Content)
		}

		// Migrate comments for this task
		taskComments := []database.Comment{}
		for _, comment := range comments {
			if comment.TaskID == task.ID {
				taskComments = append(taskComments, comment)
			}
		}

		if len(taskComments) > 0 {
			// Check which comments already exist to avoid duplicates
			existingComments, err := database.GetComments(context.Background(), task.ID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  Warning: Could not check existing comments for task %s: %v\n", task.ID, err)
			}

			// Build a map of existing comments by content+type to detect duplicates
			existingCommentMap := make(map[string]bool)
			for _, ec := range existingComments {
				key := fmt.Sprintf("%s:%s", ec.Type, ec.Content)
				existingCommentMap[key] = true
			}

			// Filter out comments that already exist
			newComments := []database.Comment{}
			taskCommentSkipCount := 0
			for _, comment := range taskComments {
				key := fmt.Sprintf("%s:%s", comment.Type, comment.Content)
				if existingCommentMap[key] {
					taskCommentSkipCount++
					continue
				}
				newComments = append(newComments, comment)
			}

			if len(newComments) > 0 {
				if err := database.AddComments(context.Background(), task.ID, newComments); err != nil {
					fmt.Fprintf(os.Stderr, "  Error adding comments for task %s: %v\n", task.ID, err)
				} else {
					commentCount += len(newComments)
					commentSkipCount += taskCommentSkipCount
					fmt.Printf("    Added %d new comment(s)", len(newComments))
					if taskCommentSkipCount > 0 {
						fmt.Printf(" (skipped %d duplicate(s))", taskCommentSkipCount)
					}
					fmt.Println()
				}
			} else if len(taskComments) > 0 {
				commentSkipCount += taskCommentSkipCount
				fmt.Printf("    All %d comment(s) already exist, skipped\n", len(taskComments))
			}
		}
	}

	fmt.Printf("\n=== Migration Complete ===\n")
	fmt.Printf("Tasks created: %d\n", taskCount)
	fmt.Printf("Tasks updated: %d\n", updateCount)
	if skipCount > 0 {
		fmt.Printf("Tasks skipped (errors): %d\n", skipCount)
	}
	fmt.Printf("Comments added: %d\n", commentCount)
	if commentSkipCount > 0 {
		fmt.Printf("Comments skipped (duplicates): %d\n", commentSkipCount)
	}
	fmt.Printf("Database location: %s\n", filepath.Join(root, ".todo2", "todo2.db"))
}

// loadJSONState loads tasks and comments from JSON file
// Supports both formats:
// 1. Top-level "comments" array with "todoId" field
// 2. Comments nested inside each task object
func loadJSONState(jsonPath string) ([]models.Todo2Task, []database.Comment, error) {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.Todo2Task{}, []database.Comment{}, nil
		}
		return nil, nil, fmt.Errorf("failed to read JSON file: %w", err)
	}

	var state map[string]interface{}
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Load tasks
	var tasks []models.Todo2Task
	if todos, ok := state["todos"].([]interface{}); ok {
		for _, todo := range todos {
			todoBytes, err := json.Marshal(todo)
			if err != nil {
				continue
			}

			var task models.Todo2Task
			if err := json.Unmarshal(todoBytes, &task); err != nil {
				continue
			}

			tasks = append(tasks, task)
		}
	}

	// Load comments - check both formats
	var comments []database.Comment

	// Format 1: Top-level comments array
	if commentsArray, ok := state["comments"].([]interface{}); ok {
		for _, comment := range commentsArray {
			commentBytes, err := json.Marshal(comment)
			if err != nil {
				continue
			}

			var commentMap map[string]interface{}
			if err := json.Unmarshal(commentBytes, &commentMap); err != nil {
				continue
			}

			todoID, _ := commentMap["todoId"].(string)
			if todoID == "" {
				todoID, _ = commentMap["todo_id"].(string) // Try alternate field name
			}

			if todoID == "" {
				continue
			}

			commentType, _ := commentMap["type"].(string)
			content, _ := commentMap["content"].(string)

			comments = append(comments, database.Comment{
				TaskID:  todoID,
				Type:    commentType,
				Content: content,
			})
		}
	}

	// Format 2: Comments nested in tasks
	// (Note: tasks already loaded above, but we need to check for nested comments separately)
	if todos, ok := state["todos"].([]interface{}); ok {
		for _, todo := range todos {
			todoMap, ok := todo.(map[string]interface{})
			if !ok {
				continue
			}

			taskID, _ := todoMap["id"].(string)
			if taskID == "" {
				continue
			}

			if commentsArray, ok := todoMap["comments"].([]interface{}); ok {
				for _, comment := range commentsArray {
					commentMap, ok := comment.(map[string]interface{})
					if !ok {
						continue
					}

					commentType, _ := commentMap["type"].(string)
					content, _ := commentMap["content"].(string)

					comments = append(comments, database.Comment{
						TaskID:  taskID,
						Type:    commentType,
						Content: content,
					})
				}
			}
		}
	}

	return tasks, comments, nil
}

// findProjectRoot finds the project root by looking for .todo2 directory
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	for {
		todo2Path := filepath.Join(dir, ".todo2")
		if _, err := os.Stat(todo2Path); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("project root not found (no .todo2 directory)")
}
