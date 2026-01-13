package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestHandleEstimationAnalyze(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a test .todo2 directory with state file
	todo2Dir := filepath.Join(tmpDir, ".todo2")
	if err := os.MkdirAll(todo2Dir, 0755); err != nil {
		t.Fatalf("failed to create .todo2 directory: %v", err)
	}

	// Create a test state file with tasks that have estimated and actual hours
	stateFile := filepath.Join(todo2Dir, "state.todo2.json")
	stateContent := `{
  "tasks": [
    {
      "id": "T-1",
      "content": "Test task 1",
      "status": "Done",
      "priority": "high",
      "tags": ["testing", "backend"],
      "estimated_hours": 5.0,
      "actual_hours": 6.0
    },
    {
      "id": "T-2",
      "content": "Test task 2",
      "status": "Done",
      "priority": "medium",
      "tags": ["testing", "frontend"],
      "estimated_hours": 3.0,
      "actual_hours": 2.5
    },
    {
      "id": "T-3",
      "content": "Test task 3",
      "status": "Done",
      "priority": "low",
      "tags": ["documentation"],
      "estimated_hours": 2.0,
      "actual_hours": 2.0
    }
  ]
}`
	if err := os.WriteFile(stateFile, []byte(stateContent), 0644); err != nil {
		t.Fatalf("failed to create state file: %v", err)
	}

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		validate  func(*testing.T, string)
	}{
		{
			name:      "basic analyze",
			params:    map[string]interface{}{},
			wantError: false,
			validate: func(t *testing.T, result string) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				if success, ok := data["success"].(bool); !ok || !success {
					t.Error("expected success=true")
				}
				if metrics, ok := data["accuracy_metrics"].(map[string]interface{}); ok {
					if totalTasks, ok := metrics["total_tasks"].(float64); !ok || totalTasks != 3 {
						t.Errorf("expected total_tasks=3, got %v", totalTasks)
					}
				}
			},
		},
		{
			name: "with detailed flag",
			params: map[string]interface{}{
				"detailed": true,
			},
			wantError: false,
			validate: func(t *testing.T, result string) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				// Should have more detailed metrics
				if _, ok := data["tag_accuracy"]; !ok {
					t.Error("expected tag_accuracy in detailed result")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handleEstimationAnalyze(tmpDir, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleEstimationAnalyze() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestAnalyzeByTag(t *testing.T) {
	completedTasks := []struct {
		name           string
		tags           []string
		priority       string
		estimatedHours float64
		actualHours    float64
		error          float64
		errorPct       float64
		absErrorPct    float64
	}{
		{
			name:           "Task 1",
			tags:           []string{"backend", "testing"},
			priority:       "high",
			estimatedHours: 5.0,
			actualHours:    6.0,
			error:          1.0,
			errorPct:       20.0,
			absErrorPct:    20.0,
		},
		{
			name:           "Task 2",
			tags:           []string{"frontend", "testing"},
			priority:       "medium",
			estimatedHours: 3.0,
			actualHours:    2.5,
			error:          -0.5,
			errorPct:       -16.67,
			absErrorPct:    16.67,
		},
	}

	result := analyzeByTag(completedTasks)
	if len(result) == 0 {
		t.Error("expected non-empty tag analysis")
	}
	// Check for expected tags
	expectedTags := []string{"backend", "frontend", "testing"}
	for _, tag := range expectedTags {
		if _, ok := result[tag]; !ok {
			t.Errorf("expected tag %s in analysis", tag)
		}
	}
}

func TestAnalyzeByPriority(t *testing.T) {
	completedTasks := []struct {
		name           string
		tags           []string
		priority       string
		estimatedHours float64
		actualHours    float64
		error          float64
		errorPct       float64
		absErrorPct    float64
	}{
		{
			name:           "High priority task",
			tags:           []string{"backend"},
			priority:       "high",
			estimatedHours: 5.0,
			actualHours:    6.0,
			error:          1.0,
			errorPct:       20.0,
			absErrorPct:    20.0,
		},
		{
			name:           "Medium priority task",
			tags:           []string{"frontend"},
			priority:       "medium",
			estimatedHours: 3.0,
			actualHours:    2.5,
			error:          -0.5,
			errorPct:       -16.67,
			absErrorPct:    16.67,
		},
	}

	result := analyzeByPriority(completedTasks)
	if len(result) == 0 {
		t.Error("expected non-empty priority analysis")
	}
	// Check for expected priorities
	expectedPriorities := []string{"high", "medium"}
	for _, priority := range expectedPriorities {
		if _, ok := result[priority]; !ok {
			t.Errorf("expected priority %s in analysis", priority)
		}
	}
}
