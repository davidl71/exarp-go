package tools

import (
	"testing"
)

// BenchmarkGetTaskLevelsIterative benchmarks getTaskLevelsIterative with various graph sizes
func BenchmarkGetTaskLevelsIterative(b *testing.B) {
	sizes := []struct {
		name      string
		taskCount int
		depsPerTask int
		cyclic    bool
	}{
		{"Small_Acyclic_10", 10, 1, false},
		{"Small_Cyclic_10", 10, 1, true},
		{"Medium_Acyclic_100", 100, 2, false},
		{"Medium_Cyclic_100", 100, 2, true},
		{"Large_Acyclic_500", 500, 3, false},
		{"Large_Cyclic_500", 500, 3, true},
		{"VeryLarge_Acyclic_1000", 1000, 4, false},
		{"VeryLarge_Cyclic_1000", 1000, 4, true},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			var tasks []Todo2Task
			if size.cyclic {
				tasks = generateCyclicTasks(size.taskCount)
			} else {
				tasks = generateTestTasks(size.taskCount, size.depsPerTask)
			}

			tg, err := BuildTaskGraph(tasks)
			if err != nil {
				b.Fatalf("Failed to build graph: %v", err)
			}

			// Force use of iterative method by ensuring graph has cycles
			if size.cyclic {
				// Graph already has cycles
			} else {
				// For acyclic graphs, we'll still test iterative method
				// by calling it directly
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = getTaskLevelsIterative(tg)
			}
		})
	}
}

// BenchmarkGetTaskLevelsOptimized benchmarks the optimized GetTaskLevels (uses topo sort for acyclic)
func BenchmarkGetTaskLevelsOptimized(b *testing.B) {
	sizes := []struct {
		name       string
		taskCount  int
		depsPerTask int
		cyclic     bool
	}{
		{"Small_Acyclic_10", 10, 1, false},
		{"Small_Cyclic_10", 10, 1, true},
		{"Medium_Acyclic_100", 100, 2, false},
		{"Medium_Cyclic_100", 100, 2, true},
		{"Large_Acyclic_500", 500, 3, false},
		{"Large_Cyclic_500", 500, 3, true},
		{"VeryLarge_Acyclic_1000", 1000, 4, false},
		{"VeryLarge_Cyclic_1000", 1000, 4, true},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			var tasks []Todo2Task
			if size.cyclic {
				tasks = generateCyclicTasks(size.taskCount)
			} else {
				tasks = generateTestTasks(size.taskCount, size.depsPerTask)
			}

			tg, err := BuildTaskGraph(tasks)
			if err != nil {
				b.Fatalf("Failed to build graph: %v", err)
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = GetTaskLevels(tg)
			}
		})
	}
}

// BenchmarkBuildTaskGraph benchmarks graph construction
func BenchmarkBuildTaskGraph(b *testing.B) {
	sizes := []struct {
		name       string
		taskCount  int
		depsPerTask int
	}{
		{"Small_10", 10, 1},
		{"Medium_100", 100, 2},
		{"Large_500", 500, 3},
		{"VeryLarge_1000", 1000, 4},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			tasks := generateTestTasks(size.taskCount, size.depsPerTask)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, _ = BuildTaskGraph(tasks)
			}
		})
	}
}

