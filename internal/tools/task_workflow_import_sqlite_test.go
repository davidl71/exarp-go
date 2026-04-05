package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/davidl71/exarp-go/internal/database"
)

func TestResolveImportSQLitePathsImmediate(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	parent := filepath.Join(base, "mono")
	svc1 := filepath.Join(parent, "svc1")
	svc2 := filepath.Join(parent, "svc2")
	for _, d := range []string{svc1, svc2} {
		if err := database.Init(d); err != nil {
			t.Fatal(err)
		}
		if err := database.Close(); err != nil {
			t.Fatal(err)
		}
	}

	got, err := resolveImportSQLitePaths(parent, []string{"."}, "immediate", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 child project dbs under mono/, got %d %+v", len(got), got)
	}
}

func TestTodo2ProjectDepth(t *testing.T) {
	t.Parallel()

	root := filepath.Join(t.TempDir(), "mono")
	db := filepath.Join(root, "svc", "a", ".todo2", "todo2.db")
	d, err := todo2ProjectDepth(root, db)
	if err != nil {
		t.Fatal(err)
	}
	if d != 2 {
		t.Fatalf("depth = %d, want 2 (svc/a)", d)
	}

	db2 := filepath.Join(root, ".todo2", "todo2.db")
	d2, err := todo2ProjectDepth(root, db2)
	if err != nil {
		t.Fatal(err)
	}
	if d2 != 0 {
		t.Fatalf("depth = %d, want 0", d2)
	}
}

func TestResolveImportRecursiveRespectsMaxDepth(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	mono := filepath.Join(base, "mono")
	shallow := filepath.Join(mono, "shallow")
	deep := filepath.Join(mono, "a", "b", "deep")
	for _, d := range []string{shallow, deep} {
		if err := database.Init(d); err != nil {
			t.Fatal(err)
		}
		if err := database.Close(); err != nil {
			t.Fatal(err)
		}
	}

	all, err := resolveImportSQLitePaths(base, []string{mono}, "recursive", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("unlimited depth: want 2 dbs, got %d", len(all))
	}

	lim, err := resolveImportSQLitePaths(base, []string{mono}, "recursive", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(lim) != 1 {
		t.Fatalf("max_depth=1: want 1 db (mono/shallow only), got %d %+v", len(lim), lim)
	}
	if !strings.Contains(lim[0].DBPath, "shallow") {
		t.Fatalf("expected shallow path, got %q", lim[0].DBPath)
	}
}

func TestHandleTaskWorkflowImportSQLiteDryRun(t *testing.T) {
	ctx := context.Background()
	base := t.TempDir()
	target := filepath.Join(base, "target")
	svc1 := filepath.Join(base, "alpha")
	svc2 := filepath.Join(base, "beta")

	ids := []string{"T-9000900000000000001", "T-9000900000000000002"}
	for i, d := range []string{svc1, svc2} {
		if err := database.Init(d); err != nil {
			t.Fatal(err)
		}
		if err := database.CreateTask(ctx, &database.Todo2Task{ID: ids[i], Content: "task " + filepath.Base(d), Status: "Todo"}); err != nil {
			t.Fatal(err)
		}
		if err := database.Close(); err != nil {
			t.Fatal(err)
		}
	}

	if err := database.Init(target); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })

	t.Setenv("PROJECT_ROOT", target)
	srcJSON, err := json.Marshal([]string{svc1, svc2})
	if err != nil {
		t.Fatal(err)
	}
	res, err := handleTaskWorkflowImportSQLite(ctx, map[string]interface{}{
		"import_sources":     string(srcJSON),
		"import_scan_mode":   "none",
		"import_on_conflict": "fail",
		"dry_run":            true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res) == 0 {
		t.Fatal("empty result")
	}
}

func TestHandleTaskWorkflowImportSQLiteIdempotent(t *testing.T) {
	ctx := context.Background()
	base := t.TempDir()
	target := filepath.Join(base, "target")
	src := filepath.Join(base, "src")

	if err := database.Init(src); err != nil {
		t.Fatal(err)
	}
	if err := database.CreateTask(ctx, &database.Todo2Task{ID: "T-9001100000000000001", Content: "idempotent slice", Status: "Todo"}); err != nil {
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}

	if err := database.Init(target); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })

	t.Setenv("PROJECT_ROOT", target)
	srcPaths, _ := json.Marshal([]string{src})

	run := func() map[string]interface{} {
		t.Helper()
		res, err := handleTaskWorkflowImportSQLite(ctx, map[string]interface{}{
			"import_sources":      string(srcPaths),
			"import_scan_mode":    "none",
			"import_on_conflict":  "fail",
			"import_sync_json":    false,
			"dry_run":             false,
		})
		if err != nil {
			t.Fatal(err)
		}
		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(res[0].Text), &payload); err != nil {
			t.Fatal(err)
		}

		return payload
	}

	first := run()
	if int(first["imported_count"].(float64)) != 1 {
		t.Fatalf("first import: imported_count = %v", first["imported_count"])
	}

	second := run()
	if int(second["imported_count"].(float64)) != 0 {
		t.Fatalf("second import: imported_count = %v want 0", second["imported_count"])
	}
	if int(second["skipped_same_content"].(float64)) != 1 {
		t.Fatalf("skipped_same_content = %v want 1", second["skipped_same_content"])
	}

	third := run()
	if int(third["imported_count"].(float64)) != 0 || int(third["skipped_same_content"].(float64)) != 1 {
		t.Fatalf("third import not idempotent: %+v", third)
	}
}

func TestIsSQLiteUniqueViolation(t *testing.T) {
	t.Parallel()
	if !isSQLiteUniqueViolation(fmt.Errorf("SQL logic error: UNIQUE constraint failed: tasks.id")) {
		t.Fatal("expected true")
	}
	if isSQLiteUniqueViolation(fmt.Errorf("foreign key failed")) {
		t.Fatal("expected false")
	}
}
