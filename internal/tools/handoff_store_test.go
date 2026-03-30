package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cast"
)

func TestHandoffStoreBinaryRoundTrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	store := HandoffStore{
		Handoffs: []HandoffEntry{
			{
				ID:        "handoff-1",
				Timestamp: "2026-03-30T12:00:00Z",
				Host:      "test-host",
				Summary:   "done",
				Blockers:  []string{"b1"},
				NextSteps: []string{"n1"},
				GitStatus: &GitStatusHandoff{Branch: "main", UncommittedFiles: 2, ChangedFiles: []string{"a.go"}},
				TasksInProgress: []TaskInProgressHandoff{
					{ID: "T-1", Content: "c", Status: "In Progress"},
				},
				TaskJournal: []TaskJournalEntry{{ID: "T-1", Action: "modified"}},
			},
		},
	}
	if err := saveHandoffStore(dir, store); err != nil {
		t.Fatal(err)
	}
	legacy := handoffsLegacyJSONPath(dir)
	if _, err := os.Stat(legacy); err == nil {
		t.Fatal("expected legacy handoffs.json removed after binary save")
	}
	got, err := loadHandoffStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Handoffs) != 1 || got.Handoffs[0].ID != "handoff-1" || got.Handoffs[0].Summary != "done" {
		t.Fatalf("unexpected decode: %+v", got.Handoffs)
	}
	if got.Handoffs[0].GitStatus == nil || got.Handoffs[0].GitStatus.Branch != "main" {
		t.Fatalf("git status: %+v", got.Handoffs[0].GitStatus)
	}
	// Second load via file cache path (ReadFile + decode)
	data, err := os.ReadFile(handoffsStorePath(dir))
	if err != nil {
		t.Fatal(err)
	}
	got2, err := loadHandoffStoreFromBytes("", data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got2.Handoffs) != 1 {
		t.Fatalf("from bytes: %+v", got2)
	}
}

func TestHandoffStoreLegacyJSONSingleDecode(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".todo2"), 0o755); err != nil {
		t.Fatal(err)
	}
	legacyJSON := `{"handoffs":[{"id":"h-old","timestamp":"t","host":"h","summary":"legacy","blockers":["x"],"next_steps":["y"]}]}`
	if err := os.WriteFile(handoffsLegacyJSONPath(dir), []byte(legacyJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := loadHandoffStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Handoffs) != 1 || got.Handoffs[0].ID != "h-old" || got.Handoffs[0].Summary != "legacy" {
		t.Fatalf("%+v", got)
	}
	if err := saveHandoffStore(dir, got); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(handoffsLegacyJSONPath(dir)); err == nil {
		t.Fatal("legacy json should be removed after migration write")
	}
	got3, err := loadHandoffStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got3.Handoffs) != 1 || got3.Handoffs[0].ID != "h-old" {
		t.Fatalf("after migration: %+v", got3)
	}
}

func TestHandoffEntryToMapJSONNumbers(t *testing.T) {
	t.Parallel()
	e := HandoffEntry{
		ID:                           "h",
		PointInTimeSnapshotTaskCount: 42,
	}
	m, err := handoffEntryToMap(e)
	if err != nil {
		t.Fatal(err)
	}
	// Direct projection keeps int; JSON decode paths often yield float64.
	var n int
	switch v := m["point_in_time_snapshot_task_count"].(type) {
	case int:
		n = v
	case float64:
		n = int(v)
	default:
		t.Fatalf("got %T %v", m["point_in_time_snapshot_task_count"], m["point_in_time_snapshot_task_count"])
	}
	if n != 42 {
		t.Fatalf("count %d", n)
	}
	// Round-trip struct
	b, _ := json.Marshal(m)
	var back HandoffEntry
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if back.PointInTimeSnapshotTaskCount != 42 {
		t.Fatalf("back: %+v", back)
	}
}

// TestHandoffEntryToMapParityWithJSONMarshaling ensures handoffEntryToMap matches
// json.Marshal→Unmarshal into map[string]any for a representative entry (int vs float64 normalized).
func TestHandoffEntryToMapParityWithJSONMarshaling(t *testing.T) {
	t.Parallel()
	e := HandoffEntry{
		ID:                           "handoff-parity",
		Timestamp:                    "2026-03-30T12:00:00Z",
		Host:                         "h1",
		Summary:                      "summary",
		Blockers:                     []string{"b1", "b2"},
		NextSteps:                    []string{"n1"},
		GitStatus:                    &GitStatusHandoff{Branch: "main", UncommittedFiles: 3, ChangedFiles: []string{"x.go", "y.go"}},
		TasksInProgress:              []TaskInProgressHandoff{{ID: "T-1", Content: "c", Status: "In Progress"}},
		TaskJournal:                  []TaskJournalEntry{{ID: "T-1", Action: "a", Summary: "s"}},
		PointInTimeSnapshot:          "snap",
		PointInTimeSnapshotFormat:    "fmt",
		PointInTimeSnapshotTaskCount: 7,
		LedgerWriteWarning:           "warn",
		ContinuityLedgerPath:         "path",
		Status:                       "open",
	}
	direct, err := handoffEntryToMap(e)
	if err != nil {
		t.Fatal(err)
	}
	jb, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	var viaJSON map[string]interface{}
	if err := json.Unmarshal(jb, &viaJSON); err != nil {
		t.Fatal(err)
	}
	if err := assertHandoffEntryMapsEqual(t, direct, viaJSON); err != nil {
		t.Fatal(err)
	}
}

func assertHandoffEntryMapsEqual(t *testing.T, direct, viaJSON map[string]interface{}) error {
	t.Helper()
	for k, dv := range direct {
		jv, ok := viaJSON[k]
		if !ok {
			return fmt.Errorf("key %q in direct missing from JSON map", k)
		}
		if err := handoffMapValuesEqual(t, k, dv, jv); err != nil {
			return err
		}
	}
	for k := range viaJSON {
		if _, ok := direct[k]; !ok {
			return fmt.Errorf("key %q in JSON map missing from direct", k)
		}
	}
	return nil
}

func handoffMapValuesEqual(t *testing.T, ctx string, a, b interface{}) error {
	t.Helper()
	switch av := a.(type) {
	case string:
		bs, ok := b.(string)
		if !ok || av != bs {
			return fmt.Errorf("%s: want string %q got %T %v", ctx, av, b, b)
		}
	case int:
		if int64(av) != cast.ToInt64(b) {
			return fmt.Errorf("%s: int mismatch %v vs %v", ctx, a, b)
		}
	case []interface{}:
		bv, ok := b.([]interface{})
		if !ok || len(av) != len(bv) {
			return fmt.Errorf("%s: slice mismatch %T len %d vs %T len %d", ctx, a, len(av), b, len(bv))
		}
		for i := range av {
			if err := handoffMapValuesEqual(t, fmt.Sprintf("%s[%d]", ctx, i), av[i], bv[i]); err != nil {
				return err
			}
		}
	case map[string]interface{}:
		bm, ok := b.(map[string]interface{})
		if !ok {
			return fmt.Errorf("%s: want map got %T", ctx, b)
		}
		if len(av) != len(bm) {
			return fmt.Errorf("%s: map len %d vs %d", ctx, len(av), len(bm))
		}
		for k, v := range av {
			w, ok := bm[k]
			if !ok {
				return fmt.Errorf("%s: missing key %q", ctx, k)
			}
			if err := handoffMapValuesEqual(t, ctx+"."+k, v, w); err != nil {
				return err
			}
		}
		for k := range bm {
			if _, ok := av[k]; !ok {
				return fmt.Errorf("%s: extra key %q in JSON side", ctx, k)
			}
		}
	default:
		return fmt.Errorf("%s: unhandled type %T", ctx, a)
	}
	return nil
}
