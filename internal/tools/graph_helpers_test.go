package tools

import (
	"fmt"
	"testing"
)

// generateTestTasks creates test tasks with dependencies.
func generateTestTasks(count int, depsPerTask int) []Todo2Task {
	tasks := make([]Todo2Task, count)

	for i := 0; i < count; i++ {
		taskID := fmt.Sprintf("T-%d", i+1)
		tasks[i] = Todo2Task{
			ID:              taskID,
			Content:         "Test Task " + taskID,
			LongDescription: "Description for " + taskID,
			Status:          "Todo",
			Priority:        "medium",
			Tags:            []string{"test"},
			Dependencies:    []string{},
		}

		// Add dependencies (simple linear or tree structure)
		if depsPerTask > 0 && i > 0 {
			deps := []string{}
			// Each task depends on previous tasks (linear chain)
			start := i - depsPerTask
			if start < 0 {
				start = 0
			}

			for j := start; j < i; j++ {
				deps = append(deps, tasks[j].ID)
			}

			tasks[i].Dependencies = deps
		}
	}

	return tasks
}

// generateCyclicTasks creates tasks with cycles.
func generateCyclicTasks(count int) []Todo2Task {
	tasks := make([]Todo2Task, count)

	for i := 0; i < count; i++ {
		taskID := fmt.Sprintf("T-%d", i+1)
		deps := []string{}

		// Create cycle: T-1 -> T-2 -> ... -> T-N -> T-1
		if i == 0 {
			// Last task depends on first
			deps = append(deps, fmt.Sprintf("T-%d", count))
		} else {
			deps = append(deps, fmt.Sprintf("T-%d", i))
		}

		tasks[i] = Todo2Task{
			ID:           taskID,
			Content:      "Cyclic Task " + taskID,
			Status:       "Todo",
			Priority:     "medium",
			Dependencies: deps,
		}
	}

	return tasks
}

func TestBuildTaskGraph(t *testing.T) {
	tests := []struct {
		name  string
		tasks []Todo2Task
		want  int // expected node count
	}{
		{
			name:  "empty tasks",
			tasks: []Todo2Task{},
			want:  0,
		},
		{
			name:  "single task",
			tasks: generateTestTasks(1, 0),
			want:  1,
		},
		{
			name:  "linear chain",
			tasks: generateTestTasks(10, 1),
			want:  10,
		},
		{
			name:  "medium graph",
			tasks: generateTestTasks(50, 2),
			want:  50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tg, err := BuildTaskGraph(tt.tasks)
			if err != nil {
				t.Fatalf("BuildTaskGraph() error = %v", err)
			}

			if got := tg.Graph.Nodes().Len(); got != tt.want {
				t.Errorf("BuildTaskGraph() node count = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasCycles(t *testing.T) {
	tests := []struct {
		name    string
		tasks   []Todo2Task
		want    bool
		wantErr bool
	}{
		{
			name:  "acyclic graph",
			tasks: generateTestTasks(10, 1),
			want:  false,
		},
		{
			name:  "cyclic graph",
			tasks: generateCyclicTasks(5),
			want:  true,
		},
		{
			name:  "single node",
			tasks: generateTestTasks(1, 0),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tg, err := BuildTaskGraph(tt.tasks)
			if err != nil {
				t.Fatalf("BuildTaskGraph() error = %v", err)
			}

			got, err := HasCycles(tg)
			if (err != nil) != tt.wantErr {
				t.Errorf("HasCycles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("HasCycles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDependencyAnalysisFromTasks(t *testing.T) {
	tests := []struct {
		name          string
		tasks         []Todo2Task
		wantCycles    bool
		wantMissing   bool
		wantMissingID string // task ID that has missing dep
	}{
		{
			name:        "acyclic graph",
			tasks:       generateTestTasks(5, 1),
			wantCycles:  false,
			wantMissing: false,
		},
		{
			name:        "cyclic graph",
			tasks:       generateCyclicTasks(3),
			wantCycles:  true,
			wantMissing: false,
		},
		{
			name: "missing dependency",
			tasks: []Todo2Task{
				{ID: "T-1", Content: "Task 1", Status: "Todo", Dependencies: []string{}},
				{ID: "T-2", Content: "Task 2", Status: "Todo", Dependencies: []string{"T-1", "T-NONEXISTENT"}},
			},
			wantCycles:    false,
			wantMissing:   true,
			wantMissingID: "T-2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cycles, missing, err := GetDependencyAnalysisFromTasks(tt.tasks)
			if err != nil {
				t.Fatalf("GetDependencyAnalysisFromTasks() error = %v", err)
			}

			hasCycles := len(cycles) > 0
			if hasCycles != tt.wantCycles {
				t.Errorf("GetDependencyAnalysisFromTasks() cycles = %v, want %v", hasCycles, tt.wantCycles)
			}

			hasMissing := len(missing) > 0
			if hasMissing != tt.wantMissing {
				t.Errorf("GetDependencyAnalysisFromTasks() missing = %v, want %v", hasMissing, tt.wantMissing)
			}

			if tt.wantMissingID != "" {
				found := false

				for _, m := range missing {
					if tid, _ := m["task_id"].(string); tid == tt.wantMissingID {
						found = true
						break
					}
				}

				if !found {
					t.Errorf("GetDependencyAnalysisFromTasks() missing should include task %s, got %v", tt.wantMissingID, missing)
				}
			}
		})
	}
}

func TestDetectCycles(t *testing.T) {
	tests := []struct {
		name        string
		tasks       []Todo2Task
		wantCycles  bool
		minCycleLen int
	}{
		{
			name:       "acyclic graph",
			tasks:      generateTestTasks(10, 1),
			wantCycles: false,
		},
		{
			name:        "cyclic graph",
			tasks:       generateCyclicTasks(5),
			wantCycles:  true,
			minCycleLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tg, err := BuildTaskGraph(tt.tasks)
			if err != nil {
				t.Fatalf("BuildTaskGraph() error = %v", err)
			}

			cycles := DetectCycles(tg)
			hasCycles := len(cycles) > 0

			if hasCycles != tt.wantCycles {
				t.Errorf("DetectCycles() hasCycles = %v, want %v", hasCycles, tt.wantCycles)
			}

			if tt.wantCycles && len(cycles) < tt.minCycleLen {
				t.Errorf("DetectCycles() cycle count = %v, want at least %v", len(cycles), tt.minCycleLen)
			}
		})
	}
}

func TestTopoSortTasks(t *testing.T) {
	tests := []struct {
		name    string
		tasks   []Todo2Task
		wantErr bool
	}{
		{
			name:  "acyclic graph",
			tasks: generateTestTasks(10, 1),
		},
		{
			name:    "cyclic graph should error",
			tasks:   generateCyclicTasks(5),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tg, err := BuildTaskGraph(tt.tasks)
			if err != nil {
				t.Fatalf("BuildTaskGraph() error = %v", err)
			}

			sorted, err := TopoSortTasks(tg)
			if (err != nil) != tt.wantErr {
				t.Errorf("TopoSortTasks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(sorted) != len(tt.tasks) {
				t.Errorf("TopoSortTasks() sorted length = %v, want %v", len(sorted), len(tt.tasks))
			}
		})
	}
}

func TestGetTaskLevels(t *testing.T) {
	tasks := generateTestTasks(10, 2)

	tg, err := BuildTaskGraph(tasks)
	if err != nil {
		t.Fatalf("BuildTaskGraph() error = %v", err)
	}

	levels := GetTaskLevels(tg)

	if len(levels) != len(tasks) {
		t.Errorf("GetTaskLevels() level count = %v, want %v", len(levels), len(tasks))
	}

	// First task should be at level 0 (no dependencies)
	if level := levels[tasks[0].ID]; level != 0 {
		t.Errorf("GetTaskLevels() first task level = %v, want 0", level)
	}
}

func TestBacklogExecutionOrder(t *testing.T) {
	// Tasks with tags: T-1 (migration), T-2 (migration, bug), T-3 (bug), T-4 (migration)
	tasks := []Todo2Task{
		{ID: "T-1", Content: "One", Status: "Todo", Priority: "high", Tags: []string{"migration"}, Dependencies: []string{}},
		{ID: "T-2", Content: "Two", Status: "Todo", Priority: "medium", Tags: []string{"migration", "bug"}, Dependencies: []string{"T-1"}},
		{ID: "T-3", Content: "Three", Status: "Todo", Priority: "low", Tags: []string{"bug"}, Dependencies: []string{"T-2"}},
		{ID: "T-4", Content: "Four", Status: "In Progress", Priority: "medium", Tags: []string{"migration"}, Dependencies: []string{}},
		{ID: "T-5", Content: "Five", Status: "Done", Priority: "low", Tags: []string{"migration"}, Dependencies: []string{}},
	}

	orderedIDs, waves, details, err := BacklogExecutionOrder(tasks, nil)
	if err != nil {
		t.Fatalf("BacklogExecutionOrder(nil filter) error = %v", err)
	}

	if len(orderedIDs) != 4 {
		t.Errorf("BacklogExecutionOrder(nil) len(orderedIDs) = %v, want 4 (Todo + In Progress only)", len(orderedIDs))
	}

	if len(details) != len(orderedIDs) {
		t.Errorf("BacklogExecutionOrder(nil) len(details) = %v, want %v", len(details), len(orderedIDs))
	}

	for i, d := range details {
		if d.Tags == nil {
			t.Errorf("BacklogExecutionOrder(nil) details[%d].Tags is nil, want non-nil", i)
		}
	}

	// Filter by tag "migration": only T-1, T-2, T-4 (backlog with migration tag)
	backlogFilter := map[string]bool{"T-1": true, "T-2": true, "T-4": true}

	orderedIDs2, _, details2, err := BacklogExecutionOrder(tasks, backlogFilter)
	if err != nil {
		t.Fatalf("BacklogExecutionOrder(filter) error = %v", err)
	}

	if len(orderedIDs2) != 3 {
		t.Errorf("BacklogExecutionOrder(filter migration) len = %v, want 3", len(orderedIDs2))
	}

	seen := make(map[string]bool)
	for _, id := range orderedIDs2 {
		if seen[id] {
			t.Errorf("BacklogExecutionOrder(filter) duplicate ID %s", id)
		}

		seen[id] = true

		if id != "T-1" && id != "T-2" && id != "T-4" {
			t.Errorf("BacklogExecutionOrder(filter) unexpected ID %s", id)
		}
	}
	// Order should respect dependencies: T-1 before T-2
	idx1, idx2 := -1, -1

	for i, id := range orderedIDs2 {
		if id == "T-1" {
			idx1 = i
		}

		if id == "T-2" {
			idx2 = i
		}
	}

	if idx1 >= 0 && idx2 >= 0 && idx1 > idx2 {
		t.Errorf("BacklogExecutionOrder(filter) T-1 should come before T-2 (dependency order)")
	}

	for _, d := range details2 {
		if d.Tags == nil {
			t.Errorf("BacklogExecutionOrder(filter) detail %s has nil Tags", d.ID)
		}
	}

	// Empty filter map = no backlog tasks returned
	emptyFilter := map[string]bool{}

	orderedIDs3, _, _, err := BacklogExecutionOrder(tasks, emptyFilter)
	if err != nil {
		t.Fatalf("BacklogExecutionOrder(empty filter) error = %v", err)
	}

	if len(orderedIDs3) != 0 {
		t.Errorf("BacklogExecutionOrder(empty filter) len = %v, want 0", len(orderedIDs3))
	}

	// Waves should be populated when there are results
	if len(waves) == 0 && len(orderedIDs) > 0 {
		t.Errorf("BacklogExecutionOrder(nil) waves empty but orderedIDs non-empty")
	}
}

func TestBuildTaskGraphParentID(t *testing.T) {
	// parent_id creates dependency edges for wave ordering (parent before child)
	tasks := []Todo2Task{
		{ID: "T-P", Content: "Parent", Status: "Todo", Dependencies: []string{}},
		{ID: "T-C1", Content: "Child1", Status: "Todo", Dependencies: []string{}, ParentID: "T-P"},
		{ID: "T-C2", Content: "Child2", Status: "Todo", Dependencies: []string{}, ParentID: "T-P"},
	}
	tg, err := BuildTaskGraph(tasks)
	if err != nil {
		t.Fatalf("BuildTaskGraph(parent_id) error = %v", err)
	}
	levels := GetTaskLevels(tg)
	if levels["T-P"] != 0 {
		t.Errorf("parent level = %v, want 0", levels["T-P"])
	}
	if levels["T-C1"] != 1 {
		t.Errorf("child1 level = %v, want 1", levels["T-C1"])
	}
	if levels["T-C2"] != 1 {
		t.Errorf("child2 level = %v, want 1", levels["T-C2"])
	}
	// BacklogExecutionOrder: Wave 0 = parent, Wave 1 = children
	ordered, waves, _, err := BacklogExecutionOrder(tasks, nil)
	if err != nil {
		t.Fatalf("BacklogExecutionOrder error = %v", err)
	}
	if len(waves) != 2 {
		t.Errorf("waves count = %v, want 2", len(waves))
	}
	if len(ordered) != 3 {
		t.Errorf("ordered count = %v, want 3", len(ordered))
	}
	if ordered[0] != "T-P" {
		t.Errorf("first = %v, want T-P", ordered[0])
	}
}

func TestLimitWavesByMaxTasks(t *testing.T) {
	// No limit: returns same map (unchanged)
	waves := map[int][]string{0: {"T-1", "T-2", "T-3"}, 1: {"T-4"}}
	got := LimitWavesByMaxTasks(waves, 0)
	if len(got) != 2 || len(got[0]) != 3 || len(got[1]) != 1 {
		t.Errorf("LimitWavesByMaxTasks(0) should return unchanged waves, got %v", got)
	}
	got = LimitWavesByMaxTasks(waves, -1)
	if len(got) != 2 || len(got[0]) != 3 || len(got[1]) != 1 {
		t.Errorf("LimitWavesByMaxTasks(-1) should return unchanged waves, got %v", got)
	}

	// Limit 2: wave 0 splits into 2 waves, wave 1 stays
	got = LimitWavesByMaxTasks(waves, 2)
	wantLevels := 3 // 0: [T-1,T-2], 1: [T-3], 2: [T-4]
	if len(got) != wantLevels {
		t.Errorf("LimitWavesByMaxTasks(2) len = %v, want %v", len(got), wantLevels)
	}
	if len(got[0]) != 2 || len(got[1]) != 1 || len(got[2]) != 1 {
		t.Errorf("LimitWavesByMaxTasks(2) chunk sizes: got %v", got)
	}

	// Limit 10: no split
	got = LimitWavesByMaxTasks(waves, 10)
	if len(got) != 2 {
		t.Errorf("LimitWavesByMaxTasks(10) len = %v, want 2", len(got))
	}
}

// Benchmark tests.
func BenchmarkBuildTaskGraph_Small(b *testing.B) {
	tasks := generateTestTasks(50, 2)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = BuildTaskGraph(tasks)
	}
}

func BenchmarkBuildTaskGraph_Medium(b *testing.B) {
	tasks := generateTestTasks(200, 3)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = BuildTaskGraph(tasks)
	}
}

func BenchmarkBuildTaskGraph_Large(b *testing.B) {
	tasks := generateTestTasks(1000, 5)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = BuildTaskGraph(tasks)
	}
}

func BenchmarkHasCycles_Acyclic(b *testing.B) {
	tasks := generateTestTasks(200, 2)
	tg, _ := BuildTaskGraph(tasks)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = HasCycles(tg)
	}
}

func BenchmarkHasCycles_Cyclic(b *testing.B) {
	tasks := generateCyclicTasks(100)
	tg, _ := BuildTaskGraph(tasks)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = HasCycles(tg)
	}
}

func BenchmarkDetectCycles_Acyclic(b *testing.B) {
	tasks := generateTestTasks(100, 2)
	tg, _ := BuildTaskGraph(tasks)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = DetectCycles(tg)
	}
}

func BenchmarkDetectCycles_Cyclic(b *testing.B) {
	tasks := generateCyclicTasks(50)
	tg, _ := BuildTaskGraph(tasks)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = DetectCycles(tg)
	}
}

func BenchmarkTopoSortTasks(b *testing.B) {
	tasks := generateTestTasks(200, 2)
	tg, _ := BuildTaskGraph(tasks)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = TopoSortTasks(tg)
	}
}

func BenchmarkGetTaskLevels(b *testing.B) {
	tasks := generateTestTasks(200, 3)
	tg, _ := BuildTaskGraph(tasks)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = GetTaskLevels(tg)
	}
}

func BenchmarkFindCriticalPath(b *testing.B) {
	tasks := generateTestTasks(200, 2)
	tg, _ := BuildTaskGraph(tasks)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = FindCriticalPath(tg)
	}
}
