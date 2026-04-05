package tools

import (
	"testing"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/models"
)

func TestDetectForbiddenOwnershipConflicts(t *testing.T) {
	t.Parallel()

	a := &database.Todo2Task{
		ID: "T-A", Status: models.StatusInProgress, LongDescription: "",
		Metadata: map[string]interface{}{
			"ownership": map[string]interface{}{
				"owned_files": []interface{}{"internal/foo.go"},
			},
		},
	}
	b := &database.Todo2Task{
		ID: "T-B", Status: models.StatusInProgress, LongDescription: "",
		Metadata: map[string]interface{}{
			"ownership": map[string]interface{}{
				"forbidden_files": []interface{}{"internal/foo.go"},
			},
		},
	}
	c := &database.Todo2Task{
		ID: "T-C", Status: models.StatusTodo,
	}

	list := []*database.Todo2Task{a, b, c}
	hits := DetectForbiddenOwnershipConflicts(list)
	if len(hits) != 1 {
		t.Fatalf("len(hits) = %d, want 1: %#v", len(hits), hits)
	}

	if hits[0].TaskID != "T-A" || hits[0].OtherTaskID != "T-B" || hits[0].Path != "internal/foo.go" {
		t.Fatalf("hit = %#v", hits[0])
	}

	if hits[0].Reason != "forbidden_file" {
		t.Fatalf("reason = %s", hits[0].Reason)
	}
}

func TestDetectForbiddenOwnershipConflictsGlob(t *testing.T) {
	t.Parallel()

	a := &database.Todo2Task{
		ID: "T-A", Status: models.StatusInProgress, LongDescription: "",
		Metadata: map[string]interface{}{
			"ownership": map[string]interface{}{
				"owned_files": []interface{}{"pkg/secret.go"},
			},
		},
	}
	b := &database.Todo2Task{
		ID: "T-B", Status: models.StatusInProgress, LongDescription: "",
		Metadata: map[string]interface{}{
			"ownership": map[string]interface{}{
				"forbidden_files": []interface{}{"pkg/*"},
			},
		},
	}

	hits := DetectForbiddenOwnershipConflicts([]*database.Todo2Task{a, b})
	if len(hits) != 1 {
		t.Fatalf("len(hits) = %d, want 1: %#v", len(hits), hits)
	}

	if hits[0].Reason != "forbidden_glob" {
		t.Fatalf("reason = %s", hits[0].Reason)
	}
}

func TestDetectFileConflictsWithPreflightIncludesTodo(t *testing.T) {
	t.Parallel()

	a := &database.Todo2Task{
		ID:              "T-A",
		Status:          models.StatusTodo,
		LongDescription: "Files/Components:\n- Update: internal/shared.go",
	}
	b := &database.Todo2Task{
		ID:              "T-B",
		Status:          models.StatusTodo,
		LongDescription: "Files/Components:\n- Update: internal/shared.go",
	}

	hits := DetectFileConflictsWithPreflight([]*database.Todo2Task{a, b}, true)
	if len(hits) != 1 {
		t.Fatalf("len(hits) = %d, want 1: %#v", len(hits), hits)
	}

	if len(hits[0].TaskStatus) != 2 {
		t.Fatalf("TaskStatus = %#v", hits[0].TaskStatus)
	}
}

func TestDetectForbiddenOwnershipPreflightIncludesTodo(t *testing.T) {
	t.Parallel()

	a := &database.Todo2Task{
		ID:     "T-A",
		Status: models.StatusTodo,
		Metadata: map[string]interface{}{
			"ownership": map[string]interface{}{
				"owned_files": []interface{}{"internal/foo.go"},
			},
		},
	}
	b := &database.Todo2Task{
		ID:     "T-B",
		Status: models.StatusTodo,
		Metadata: map[string]interface{}{
			"ownership": map[string]interface{}{
				"forbidden_files": []interface{}{"internal/foo.go"},
			},
		},
	}

	hits := DetectForbiddenOwnershipConflictsWithPreflight([]*database.Todo2Task{a, b}, true)
	if len(hits) != 1 {
		t.Fatalf("len(hits) = %d, want 1: %#v", len(hits), hits)
	}
}
