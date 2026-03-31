package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/davidl71/exarp-go/internal/models"
	"github.com/davidl71/exarp-go/proto"
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

// TestEstimationRequestToParamsLocalAIBackend verifies that EstimationRequest.local_ai_backend
// is included in the params map when set (A2: proto field 11).
func TestEstimationRequestToParamsLocalAIBackend(t *testing.T) {
	req := &proto.EstimationRequest{
		Action:         "estimate",
		Name:           "Test task",
		LocalAiBackend: "ollama",
	}
	params := EstimationRequestToParams(req)
	if got, ok := params["local_ai_backend"].(string); !ok || got != "ollama" {
		t.Errorf("EstimationRequestToParams() local_ai_backend = %v (ok=%v), want ollama", params["local_ai_backend"], ok)
	}
}

// TestDecodeArgsToProtoTaskWorkflowTagsJSONArray verifies CLI-style args use JSON arrays for
// repeated proto fields (canonical #tags must not be sent as a single comma-separated string).
func TestDecodeArgsToProtoTaskWorkflowTagsJSONArray(t *testing.T) {
	t.Parallel()
	raw := json.RawMessage(`{"action":"create","name":"x","tags":["#cli","#proto"],"dependencies":["T-1"]}`)
	req, err := decodeArgsToProto(raw, func() *proto.TaskWorkflowRequest { return &proto.TaskWorkflowRequest{} })
	if err != nil {
		t.Fatalf("decodeArgsToProto: %v", err)
	}
	if len(req.Tags) != 2 || req.Tags[0] != "#cli" || req.Tags[1] != "#proto" {
		t.Fatalf("Tags = %#v", req.Tags)
	}
	if len(req.Dependencies) != 1 || req.Dependencies[0] != "T-1" {
		t.Fatalf("Dependencies = %#v", req.Dependencies)
	}
}

func TestDecodeArgsToProtoTaskWorkflowTagsCSVStringRejected(t *testing.T) {
	t.Parallel()
	raw := json.RawMessage(`{"action":"create","name":"x","tags":"#cli,#proto"}`)
	_, err := decodeArgsToProto(raw, func() *proto.TaskWorkflowRequest { return &proto.TaskWorkflowRequest{} })
	if err == nil {
		t.Fatal("expected error when tags is a string for repeated field")
	}
}
