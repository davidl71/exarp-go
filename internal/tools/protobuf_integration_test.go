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

func TestEstimationRequestToParamsEnumOverridesString(t *testing.T) {
	req := &proto.EstimationRequest{
		Action:             "estimate",
		Name:               "x",
		LocalAiBackend:     "ollama",
		LocalAiBackendEnum: proto.LocalLLMBackend_LOCAL_LLM_BACKEND_FM,
		SummaryLevelEnum:   proto.LocalLLMSummaryLevel_LOCAL_LLM_SUMMARY_LEVEL_DETAILED,
	}
	params := EstimationRequestToParams(req)
	if got, ok := params["local_ai_backend"].(string); !ok || got != "fm" {
		t.Errorf("local_ai_backend = %v, want fm (enum wins)", params["local_ai_backend"])
	}
	if got, ok := params["summary_level"].(string); !ok || got != "detailed" {
		t.Errorf("summary_level = %v, want detailed", params["summary_level"])
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

// TestDecodeEnumOnlyMCPArgsReportTaskWorkflowTaskAnalysis verifies enum-first MCP payloads that
// omit legacy "action" / "output_format" strings (protojson field names actionEnum / outputFormatEnum).
func TestDecodeEnumOnlyMCPArgsReportTaskWorkflowTaskAnalysis(t *testing.T) {
	t.Parallel()

	t.Run("report", func(t *testing.T) {
		raw := json.RawMessage(`{"actionEnum":"REPORT_ACTION_SCORECARD","outputFormatEnum":"OUTPUT_FORMAT_JSON"}`)
		req, err := decodeArgsToProto(raw, func() *proto.ReportRequest { return &proto.ReportRequest{} })
		if err != nil {
			t.Fatalf("decodeArgsToProto: %v", err)
		}
		if req.GetActionEnum() != proto.ReportAction_REPORT_ACTION_SCORECARD {
			t.Fatalf("ActionEnum = %v", req.GetActionEnum())
		}
		params := ReportRequestToParams(req)
		if params["action"] != "scorecard" {
			t.Fatalf("params[action] = %v, want scorecard", params["action"])
		}
		if params["output_format"] != "json" {
			t.Fatalf("params[output_format] = %v, want json", params["output_format"])
		}
	})

	t.Run("task_workflow", func(t *testing.T) {
		raw := json.RawMessage(`{"actionEnum":"TASK_WORKFLOW_ACTION_VERIFY"}`)
		req, err := decodeArgsToProto(raw, func() *proto.TaskWorkflowRequest { return &proto.TaskWorkflowRequest{} })
		if err != nil {
			t.Fatalf("decodeArgsToProto: %v", err)
		}
		params := TaskWorkflowRequestToParams(req)
		if params["action"] != "verify" {
			t.Fatalf("params[action] = %v, want verify", params["action"])
		}
	})

	t.Run("task_analysis", func(t *testing.T) {
		raw := json.RawMessage(`{"actionEnum":"TASK_ANALYSIS_ACTION_NEXT_BATCH","outputFormatEnum":"OUTPUT_FORMAT_JSON"}`)
		req, err := decodeArgsToProto(raw, func() *proto.TaskAnalysisRequest { return &proto.TaskAnalysisRequest{} })
		if err != nil {
			t.Fatalf("decodeArgsToProto: %v", err)
		}
		params := TaskAnalysisRequestToParams(req)
		if params["action"] != "next_batch" {
			t.Fatalf("params[action] = %v, want next_batch", params["action"])
		}
		if params["output_format"] != "json" {
			t.Fatalf("params[output_format] = %v, want json", params["output_format"])
		}
	})
}
