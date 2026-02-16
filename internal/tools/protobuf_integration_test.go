package tools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/davidl71/exarp-go/internal/models"
)

// TestProtobufRoundTripWithRealTasks verifies protobuf serialization with real Todo2 tasks (T-1768317405631).
// Loads tasks from .todo2 (DB or JSON), serializes first N to protobuf and back, verifies round-trip.
func TestProtobufRoundTripWithRealTasks(t *testing.T) {
	projectRoot := findProjectRootForTest(t)
	if projectRoot == "" {
		t.Skip("PROJECT_ROOT not set; skipping protobuf integration test with real tasks")
	}

	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		t.Skipf("LoadTodo2Tasks failed (no .todo2?): %v", err)
	}

	if len(tasks) == 0 {
		t.Skip("No tasks in Todo2; skipping protobuf round-trip test")
	}

	limit := 10
	if len(tasks) < limit {
		limit = len(tasks)
	}

	for i := 0; i < limit; i++ {
		task := &tasks[i]
		t.Run(task.ID, func(t *testing.T) {
			data, err := models.SerializeTaskToProtobuf(task)
			if err != nil {
				t.Fatalf("SerializeTaskToProtobuf() error = %v", err)
			}

			if len(data) == 0 {
				t.Error("SerializeTaskToProtobuf() returned empty data")
				return
			}

			deserialized, err := models.DeserializeTaskFromProtobuf(data)
			if err != nil {
				t.Fatalf("DeserializeTaskFromProtobuf() error = %v", err)
			}

			if deserialized.ID != task.ID {
				t.Errorf("ID = %v, want %v", deserialized.ID, task.ID)
			}

			if deserialized.Content != task.Content {
				t.Errorf("Content = %v, want %v", deserialized.Content, task.Content)
			}

			if deserialized.Status != task.Status {
				t.Errorf("Status = %v, want %v", deserialized.Status, task.Status)
			}
		})
	}
}

func findProjectRootForTest(t *testing.T) string {
	t.Helper()

	if root := os.Getenv("PROJECT_ROOT"); root != "" {
		return root
	}
	// Try common locations relative to test
	cwd, _ := os.Getwd()
	for _, rel := range []string{".", "..", "../..", "../../.."} {
		p := filepath.Clean(filepath.Join(cwd, rel))
		if _, err := os.Stat(filepath.Join(p, ".todo2")); err == nil {
			return p
		}
	}

	return ""
}
