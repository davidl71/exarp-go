package tools

import (
	"fmt"
	"strings"
	"testing"
)

// TestCriticalPathOnRealData tests critical path with actual Todo2 tasks.
func TestCriticalPathOnRealData(t *testing.T) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		t.Skipf("Skipping test: %v", err)
		return
	}

	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		t.Fatalf("Failed to load tasks: %v", err)
	}

	if len(tasks) == 0 {
		t.Skip("No tasks found, skipping test")
		return
	}

	tg, err := BuildTaskGraph(tasks)
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Check for cycles first
	hasCycles, err := HasCycles(tg)
	if err != nil {
		t.Fatalf("Failed to check cycles: %v", err)
	}

	if hasCycles {
		cycles := DetectCycles(tg)
		t.Logf("Graph has %d cycles, cannot compute critical path", len(cycles))
		t.Logf("Cycles: %v", cycles)

		return
	}

	// Find critical path
	criticalPath, err := FindCriticalPath(tg)
	if err != nil {
		t.Fatalf("Failed to find critical path: %v", err)
	}

	t.Logf("Critical Path Length: %d tasks", len(criticalPath))
	t.Logf("Critical Path: %v", criticalPath)

	// Verify path makes sense
	if len(criticalPath) == 0 {
		t.Error("Critical path is empty")
	}

	// Print detailed path
	fmt.Println("\n" + strings.Repeat("=", 62))
	fmt.Println("CRITICAL PATH ANALYSIS")
	fmt.Println(strings.Repeat("=", 62))
	fmt.Printf("\nTotal Tasks: %d\n", len(tasks))
	fmt.Printf("Critical Path Length: %d tasks\n", len(criticalPath))

	// Get task levels for context
	levels := GetTaskLevels(tg)
	maxLevel := 0

	for _, level := range levels {
		if level > maxLevel {
			maxLevel = level
		}
	}

	fmt.Printf("Max Dependency Level: %d\n", maxLevel)

	fmt.Println("\nCritical Path Chain (Longest Dependency Path):")
	fmt.Println(strings.Repeat("-", 62))

	for i, taskID := range criticalPath {
		// Find task details
		for _, task := range tasks {
			if task.ID == taskID {
				fmt.Printf("\n%d. %s", i+1, taskID)

				if task.Content != "" {
					fmt.Printf(": %s", task.Content)
				}

				fmt.Println()

				if task.Priority != "" || task.Status != "" {
					fmt.Printf("   ")

					if task.Priority != "" {
						fmt.Printf("Priority: %s", task.Priority)
					}

					if task.Priority != "" && task.Status != "" {
						fmt.Printf(" | ")
					}

					if task.Status != "" {
						fmt.Printf("Status: %s", task.Status)
					}

					fmt.Println()
				}

				if len(task.Dependencies) > 0 {
					fmt.Printf("   Depends on: %v\n", task.Dependencies)
				}

				if level, ok := levels[taskID]; ok {
					fmt.Printf("   Dependency Level: %d\n", level)
				}

				break
			}
		}

		if i < len(criticalPath)-1 {
			fmt.Println("   ↓")
		}
	}

	fmt.Println()
	fmt.Println(strings.Repeat("-", 62))
	fmt.Println("\n💡 The critical path shows the longest dependency chain.")
	fmt.Println("   Tasks on this path determine the minimum project duration.")
	fmt.Println()
}

// Test uses the exported AnalyzeCriticalPath from graph_helpers.go
