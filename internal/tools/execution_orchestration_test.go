package tools

import (
	"testing"
)

func TestCalculateDependencyWave(t *testing.T) {
	tests := []struct {
		name     string
		tasks    []Todo2Task
		taskID   string
		wantWave int
	}{
		{
			name: "single task no deps",
			tasks: []Todo2Task{
				{ID: "T-1", Content: "Task 1", Status: "Todo", Priority: "medium"},
			},
			taskID:   "T-1",
			wantWave: 0,
		},
		{
			name: "linear chain: T-2 depends on T-1",
			tasks: []Todo2Task{
				{ID: "T-1", Content: "Task 1", Status: "Todo", Priority: "medium"},
				{ID: "T-2", Content: "Task 2", Status: "Todo", Priority: "medium", Dependencies: []string{"T-1"}},
			},
			taskID:   "T-2",
			wantWave: 1,
		},
		{
			name: "linear chain: first task is wave 0",
			tasks: []Todo2Task{
				{ID: "T-1", Content: "Task 1", Status: "Todo", Priority: "medium"},
				{ID: "T-2", Content: "Task 2", Status: "Todo", Priority: "medium", Dependencies: []string{"T-1"}},
			},
			taskID:   "T-1",
			wantWave: 0,
		},
		{
			name: "parallel tasks all wave 0",
			tasks: []Todo2Task{
				{ID: "T-A", Content: "Task A", Status: "Todo", Priority: "medium"},
				{ID: "T-B", Content: "Task B", Status: "Todo", Priority: "medium"},
				{ID: "T-C", Content: "Task C", Status: "Todo", Priority: "medium"},
			},
			taskID:   "T-B",
			wantWave: 0,
		},
		{
			// Wave reflects structural position in the dependency graph.
			// T-1 being Done doesn't change T-2's graph depth (1); it only
			// means T-2 is currently unblocked, not that its wave changes.
			name: "done dependency: wave reflects structural depth",
			tasks: []Todo2Task{
				{ID: "T-1", Content: "Task 1", Status: "Done", Priority: "medium"},
				{ID: "T-2", Content: "Task 2", Status: "Todo", Priority: "medium", Dependencies: []string{"T-1"}},
			},
			taskID:   "T-2",
			wantWave: 1,
		},
		{
			name: "unknown task ID",
			tasks: []Todo2Task{
				{ID: "T-1", Content: "Task 1", Status: "Todo", Priority: "medium"},
			},
			taskID:   "T-999",
			wantWave: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateDependencyWave(tt.tasks, tt.taskID)
			if got != tt.wantWave {
				t.Errorf("CalculateDependencyWave(%q) = %d, want %d", tt.taskID, got, tt.wantWave)
			}
		})
	}
}
