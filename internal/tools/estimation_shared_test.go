package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleEstimationAnalyze(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a test .todo2 directory with state file
	todo2Dir := filepath.Join(tmpDir, ".todo2")
	if err := os.MkdirAll(todo2Dir, 0755); err != nil {
		t.Fatalf("failed to create .todo2 directory: %v", err)
	}

	// Create a test state file with todos that have estimated and actual hours
	// Format matches .todo2/state.todo2.json (todos array, camelCase for hours)
	stateFile := filepath.Join(todo2Dir, "state.todo2.json")
	stateContent := `{
  "todos": [
    {
      "id": "T-1",
      "content": "Test task 1",
      "status": "Done",
      "priority": "high",
      "tags": ["testing", "backend"],
      "estimatedHours": 5.0,
      "actualHours": 6.0
    },
    {
      "id": "T-2",
      "content": "Test task 2",
      "status": "Done",
      "priority": "medium",
      "tags": ["testing", "frontend"],
      "estimatedHours": 3.0,
      "actualHours": 2.5
    },
    {
      "id": "T-3",
      "content": "Test task 3",
      "status": "Done",
      "priority": "low",
      "tags": ["documentation"],
      "estimatedHours": 2.0,
      "actualHours": 2.0
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

func TestGetPreferredBackend(t *testing.T) {
	tests := []struct {
		name     string
		metadata map[string]interface{}
		want     string
	}{
		{"nil", nil, ""},
		{"empty", map[string]interface{}{}, ""},
		{"ollama", map[string]interface{}{MetadataKeyPreferredBackend: "ollama"}, "ollama"},
		{"fm", map[string]interface{}{MetadataKeyPreferredBackend: "fm"}, "fm"},
		{"mlx", map[string]interface{}{MetadataKeyPreferredBackend: "mlx"}, "mlx"},
		{"uppercase", map[string]interface{}{MetadataKeyPreferredBackend: "OLLAMA"}, "ollama"},
		{"invalid", map[string]interface{}{MetadataKeyPreferredBackend: "other"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetPreferredBackend(tt.metadata); got != tt.want {
				t.Errorf("GetPreferredBackend() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildEstimationPrompt(t *testing.T) {
	p := BuildEstimationPrompt("Add login", "Implement OAuth", []string{"auth"}, "high")
	if p == "" {
		t.Error("expected non-empty prompt")
	}
	if !strings.Contains(p, "Add login") || !strings.Contains(p, "OAuth") || !strings.Contains(p, "auth") || !strings.Contains(p, "high") {
		t.Error("prompt should contain task name, details, tags, priority")
	}
	if !strings.Contains(p, "estimate_hours") {
		t.Error("prompt should request JSON with estimate_hours")
	}
}

func TestParseLLMEstimationResponse(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantErr bool
		check   func(*testing.T, *EstimationResult)
	}{
		{
			name:    "valid json",
			text:    `{"estimate_hours": 3.5, "confidence": 0.8, "complexity": 5, "reasoning": "Moderate task"}`,
			wantErr: false,
			check: func(t *testing.T, r *EstimationResult) {
				if r.EstimateHours != 3.5 || r.Confidence != 0.8 || r.Method != "ollama" {
					t.Errorf("got EstimateHours=%.1f Confidence=%.1f Method=%s", r.EstimateHours, r.Confidence, r.Method)
				}
			},
		},
		{
			name:    "with markdown",
			text:    "Here is the estimate:\n```json\n{\"estimate_hours\": 2, \"confidence\": 0.7, \"complexity\": 3, \"reasoning\": \"Simple\"}\n```",
			wantErr: false,
			check: func(t *testing.T, r *EstimationResult) {
				if r.EstimateHours != 2 {
					t.Errorf("got EstimateHours=%.1f", r.EstimateHours)
				}
			},
		},
		{
			name:    "no json",
			text:    "no json here",
			wantErr: true,
			check:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLLMEstimationResponse(tt.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLLMEstimationResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil && got != nil {
				tt.check(t, got)
			}
		})
	}
}
