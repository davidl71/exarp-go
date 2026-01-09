package tools

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// generateDuplicateTestTasks creates tasks with varying similarity for duplicate detection testing
func generateDuplicateTestTasks(count int, duplicateRatio float64) []Todo2Task {
	tasks := make([]Todo2Task, count)
	rand.Seed(time.Now().UnixNano())

	// Create base tasks
	baseTasks := []string{
		"Implement user authentication system",
		"Add database migration scripts",
		"Create API endpoints for user management",
		"Write unit tests for authentication",
		"Set up CI/CD pipeline",
		"Implement error handling",
		"Add logging functionality",
		"Create documentation",
		"Optimize database queries",
		"Implement caching layer",
	}

	duplicateCount := int(float64(count) * duplicateRatio)
	uniqueCount := count - duplicateCount

	// Generate unique tasks
	for i := 0; i < uniqueCount && i < len(baseTasks); i++ {
		tasks[i] = Todo2Task{
			ID:              fmt.Sprintf("T-%d", i+1),
			Content:         baseTasks[i%len(baseTasks)],
			LongDescription: fmt.Sprintf("Description for task %d", i+1),
			Status:          "Todo",
			Priority:        "medium",
			Tags:            []string{"test"},
		}
	}

	// Generate duplicate tasks (variations of base tasks)
	for i := uniqueCount; i < count; i++ {
		baseIdx := rand.Intn(len(baseTasks))
		variation := baseTasks[baseIdx]

		// Add slight variations to create similar but not identical tasks
		if rand.Float64() > 0.5 {
			variation = variation + " with enhancements"
		} else {
			variation = "Enhanced " + variation
		}

		tasks[i] = Todo2Task{
			ID:              fmt.Sprintf("T-%d", i+1),
			Content:         variation,
			LongDescription: fmt.Sprintf("Description for task %d", i+1),
			Status:          "Todo",
			Priority:        "medium",
			Tags:            []string{"test"},
		}
	}

	// Shuffle tasks
	rand.Shuffle(len(tasks), func(i, j int) {
		tasks[i], tasks[j] = tasks[j], tasks[i]
	})

	return tasks
}

// BenchmarkFindDuplicateTasks benchmarks findDuplicateTasks with various task counts
func BenchmarkFindDuplicateTasks(b *testing.B) {
	sizes := []struct {
		name           string
		taskCount      int
		duplicateRatio float64
		threshold      float64
	}{
		{"Small_100", 100, 0.1, 0.85},
		{"Small_100_HighDup", 100, 0.3, 0.85},
		{"Medium_500", 500, 0.1, 0.85},
		{"Medium_500_HighDup", 500, 0.3, 0.85},
		{"Large_1000", 1000, 0.1, 0.85},
		{"Large_1000_HighDup", 1000, 0.3, 0.85},
		{"VeryLarge_5000", 5000, 0.1, 0.85},
		{"VeryLarge_5000_HighDup", 5000, 0.3, 0.85},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			tasks := generateDuplicateTestTasks(size.taskCount, size.duplicateRatio)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = findDuplicateTasks(tasks, size.threshold)
			}
		})
	}
}

// BenchmarkCalculateSimilarity benchmarks the similarity calculation function
func BenchmarkCalculateSimilarity(b *testing.B) {
	task1 := Todo2Task{
		ID:              "T-1",
		Content:         "Implement user authentication system",
		LongDescription: "Create a secure authentication system with JWT tokens",
		Status:          "Todo",
		Priority:        "high",
	}

	task2 := Todo2Task{
		ID:              "T-2",
		Content:         "Implement user authentication system with enhancements",
		LongDescription: "Create a secure authentication system with JWT tokens and OAuth",
		Status:          "Todo",
		Priority:        "high",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = calculateSimilarity(task1, task2)
	}
}

// BenchmarkAnalyzeTags benchmarks tag analysis
func BenchmarkAnalyzeTags(b *testing.B) {
	sizes := []struct {
		name        string
		taskCount   int
		tagsPerTask int
	}{
		{"Small_100", 100, 2},
		{"Medium_500", 500, 3},
		{"Large_1000", 1000, 4},
		{"VeryLarge_5000", 5000, 5},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			tasks := generateTaggedTestTasks(size.taskCount, size.tagsPerTask)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = analyzeTags(tasks)
			}
		})
	}
}

// generateTaggedTestTasks creates tasks with tags for testing
func generateTaggedTestTasks(count int, tagsPerTask int) []Todo2Task {
	tasks := make([]Todo2Task, count)
	tags := []string{"bug", "feature", "refactor", "testing", "documentation", "migration", "performance", "security"}

	for i := 0; i < count; i++ {
		taskTags := make([]string, 0, tagsPerTask)
		for j := 0; j < tagsPerTask; j++ {
			tagIdx := (i + j) % len(tags)
			taskTags = append(taskTags, tags[tagIdx])
		}

		tasks[i] = Todo2Task{
			ID:              fmt.Sprintf("T-%d", i+1),
			Content:         fmt.Sprintf("Task %d", i+1),
			LongDescription: fmt.Sprintf("Description for task %d", i+1),
			Status:          "Todo",
			Priority:        "medium",
			Tags:            taskTags,
		}
	}

	return tasks
}
