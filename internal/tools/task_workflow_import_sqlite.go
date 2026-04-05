// task_workflow_import_sqlite.go — Merge multiple Todo2 SQLite DBs into the current PROJECT_ROOT store.
// Imports are idempotent: same sources re-applied skip tasks already in the target when content+long_description match (models.ContentHash); no duplicate inserts; JSON sync runs only when new rows were inserted.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/models"
	"github.com/spf13/cast"
)

// importTaskContentKey is the canonical import/dedup key: always derived from content + long_description
// (see models.ContentHash). Stored metadata["content_hash"] is not used so imports stay idempotent even when
// metadata drifts after CreateTask or Round-trip serialization.
func importTaskContentKey(t *database.Todo2Task) string {
	if t == nil {
		return ""
	}

	return models.ContentHash(t)
}

func isSQLiteUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())

	return strings.Contains(s, "unique constraint")
}

// ImportSQLiteResolved is one discovered todo2.db with a label for default project_id.
type ImportSQLiteResolved struct {
	DBPath string `json:"db_path"`
	Label  string `json:"label"`
}

// todo2ProjectDepth is the number of path segments between sourceRoot and the project directory
// (the parent of .todo2). Zero when the DB is sourceRoot/.todo2/todo2.db.
func todo2ProjectDepth(sourceRoot, dbPath string) (int, error) {
	sr, err := filepath.Abs(filepath.Clean(sourceRoot))
	if err != nil {
		return 0, err
	}

	dbAbs, err := filepath.Abs(filepath.Clean(dbPath))
	if err != nil {
		return 0, err
	}

	projDir := filepath.Clean(filepath.Dir(filepath.Dir(dbAbs)))
	rel, err := filepath.Rel(sr, projDir)
	if err != nil {
		return 0, err
	}

	if rel == "." {
		return 0, nil
	}

	n := 0
	for _, seg := range strings.Split(rel, string(filepath.Separator)) {
		if seg != "" {
			n++
		}
	}

	return n, nil
}

func resolveImportSQLitePaths(projectRoot string, raw []string, scanMode string, maxDepth int) ([]ImportSQLiteResolved, error) {
	scanMode = strings.ToLower(strings.TrimSpace(scanMode))
	if scanMode == "" {
		scanMode = "none"
	}

	seen := make(map[string]bool)
	var out []ImportSQLiteResolved

	addDB := func(dbPath string) error {
		st, err := os.Stat(dbPath)
		if err != nil {
			return err
		}
		if st.IsDir() {
			return fmt.Errorf("not a sqlite file: %s", dbPath)
		}
		ap, err := filepath.Abs(dbPath)
		if err != nil {
			ap = dbPath
		}
		if seen[ap] {
			return nil
		}
		seen[ap] = true
		out = append(out, ImportSQLiteResolved{DBPath: ap, Label: labelForTodo2DBPath(ap)})

		return nil
	}

	// tryAdd respects import_max_depth relative to depthRoot (directory source). Empty depthRoot skips the check.
	tryAdd := func(dbPath, depthRoot string) error {
		if maxDepth > 0 && depthRoot != "" {
			dep, err := todo2ProjectDepth(depthRoot, dbPath)
			if err != nil {
				return fmt.Errorf("import depth %s vs %s: %w", depthRoot, dbPath, err)
			}
			if dep > maxDepth {
				return nil
			}
		}
		return addDB(dbPath)
	}

	for _, r := range raw {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}

		p := r
		if !filepath.IsAbs(p) {
			p = filepath.Join(projectRoot, p)
		}
		p = filepath.Clean(p)

		st, err := os.Stat(p)
		if err != nil {
			return nil, fmt.Errorf("import source %q: %w", r, err)
		}

		if !st.IsDir() {
			if err := addDB(p); err != nil {
				return nil, fmt.Errorf("import source %q: %w", r, err)
			}

			continue
		}

		mainDb := filepath.Join(p, ".todo2", "todo2.db")
		if err := tryAdd(mainDb, p); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("import dir %q (expected %s): %w", r, mainDb, err)
		}

		switch scanMode {
		case "none":
			continue
		case "immediate":
			entries, err := os.ReadDir(p)
			if err != nil {
				return nil, err
			}
			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				cand := filepath.Join(p, e.Name(), ".todo2", "todo2.db")
				_ = tryAdd(cand, p) // ignore missing child db
			}
		case "recursive":
			err := filepath.WalkDir(p, func(path string, d os.DirEntry, werr error) error {
				if werr != nil {
					return werr
				}
				if d.IsDir() {
					return nil
				}
				if filepath.Base(path) != "todo2.db" {
					return nil
				}
				parent := filepath.Base(filepath.Dir(path))
				if parent != ".todo2" {
					return nil
				}
				_ = tryAdd(path, p)

				return nil
			})
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("import_scan_mode must be none, immediate, or recursive, got %q", scanMode)
		}
	}

	return out, nil
}

func labelForTodo2DBPath(dbPath string) string {
	dir := filepath.Dir(dbPath)
	parent := filepath.Base(filepath.Dir(dir))
	if parent == "" || parent == "." {
		return "default"
	}

	return parent
}

func topologicalImportOrder(tasks []*database.Todo2Task) ([]*database.Todo2Task, error) {
	idToTask := make(map[string]*database.Todo2Task, len(tasks))
	for _, t := range tasks {
		idToTask[t.ID] = t
	}

	graph := make(map[string][]string)
	indeg := make(map[string]int, len(idToTask))
	for id := range idToTask {
		indeg[id] = 0
	}

	for _, t := range tasks {
		for _, d := range t.Dependencies {
			if _, ok := idToTask[d]; !ok {
				continue
			}
			// edge d -> t (d before t)
			graph[d] = append(graph[d], t.ID)
			indeg[t.ID]++
		}
	}

	var q []string
	for id, v := range indeg {
		if v == 0 {
			q = append(q, id)
		}
	}
	sort.Strings(q)

	var order []string
	for len(q) > 0 {
		sort.Strings(q)
		n := q[0]
		q = q[1:]
		order = append(order, n)
		succs := graph[n]
		sort.Strings(succs)
		for _, s := range succs {
			indeg[s]--
			if indeg[s] == 0 {
				q = append(q, s)
			}
		}
	}

	if len(order) != len(idToTask) {
		return nil, fmt.Errorf("cycle or unsortable dependencies among %d imported tasks", len(idToTask))
	}

	out := make([]*database.Todo2Task, 0, len(order))
	for _, id := range order {
		out = append(out, idToTask[id])
	}

	return out, nil
}

func handleTaskWorkflowImportSQLite(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("import_sqlite: %w", err)
	}

	targetDB, _ := filepath.Abs(filepath.Join(projectRoot, ".todo2", "todo2.db"))

	rawJSON := strings.TrimSpace(ParamString(params, "import_sources"))
	if rawJSON == "" {
		rawJSON = strings.TrimSpace(cast.ToString(params["import_sources"]))
	}
	if rawJSON == "" {
		return nil, fmt.Errorf("import_sqlite: import_sources is required (JSON array of paths, e.g. [\"../svc/a\",\"../svc/b\"])")
	}

	var rawPaths []string
	if err := json.Unmarshal([]byte(rawJSON), &rawPaths); err != nil {
		return nil, fmt.Errorf("import_sqlite: import_sources must be a JSON array of strings: %w", err)
	}

	scanMode := strings.TrimSpace(cast.ToString(params["import_scan_mode"]))
	onConflict := strings.ToLower(strings.TrimSpace(cast.ToString(params["import_on_conflict"])))
	if onConflict == "" {
		onConflict = "fail"
	}
	if onConflict != "fail" && onConflict != "skip" {
		return nil, fmt.Errorf("import_sqlite: import_on_conflict must be fail or skip, got %q", onConflict)
	}

	defaultProj := strings.TrimSpace(cast.ToString(params["import_default_project_id"]))
	dryRun := ParamBool(params, "dry_run", false)
	syncJSON := true
	if params["import_sync_json"] != nil {
		syncJSON = cast.ToBool(params["import_sync_json"])
	}

	maxDepth := cast.ToInt(params["import_max_depth"])
	if maxDepth < 0 {
		return nil, fmt.Errorf("import_sqlite: import_max_depth must be >= 0 (0 means unlimited)")
	}

	resolved, err := resolveImportSQLitePaths(projectRoot, rawPaths, scanMode, maxDepth)
	if err != nil {
		return nil, fmt.Errorf("import_sqlite: %w", err)
	}
	if len(resolved) == 0 {
		return nil, fmt.Errorf("import_sqlite: no todo2.db files resolved from sources")
	}

	warnings := make([]string, 0, 8)
	skippedSelf := 0
	filtered := resolved[:0]
	for _, r := range resolved {
		ap := r.DBPath
		if ap == targetDB {
			skippedSelf++
			warnings = append(warnings, "skipped target database (cannot import from self): "+ap)

			continue
		}
		filtered = append(filtered, r)
	}
	resolved = filtered

	if len(resolved) == 0 {
		return nil, fmt.Errorf("import_sqlite: all sources were the target DB; nothing to import")
	}

	existing, err := database.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("import_sqlite: load target tasks: %w", err)
	}

	targetByID := make(map[string]*database.Todo2Task, len(existing))
	targetHash := make(map[string]string, len(existing))
	for _, t := range existing {
		if t == nil {
			continue
		}
		tCopy := *t
		targetByID[t.ID] = t
		targetHash[t.ID] = importTaskContentKey(&tCopy)
	}

	merged := make(map[string]*database.Todo2Task)
	sourceProvenance := make(map[string]string)

	for _, src := range resolved {
		tasks, err := database.ListTasksFromSQLiteFile(ctx, src.DBPath, nil)
		if err != nil {
			return nil, fmt.Errorf("import_sqlite: read %s: %w", src.DBPath, err)
		}

		for _, t := range tasks {
			if t == nil {
				continue
			}
			tCopy := *t
			if old, ok := merged[t.ID]; ok {
				if importTaskContentKey(old) != importTaskContentKey(&tCopy) {
					return nil, fmt.Errorf("import_sqlite: duplicate task id %s from multiple sources with differing content", t.ID)
				}

				continue
			}
			merged[t.ID] = &tCopy
			sourceProvenance[t.ID] = src.Label
		}
	}

	var missingProject []string
	conflicts := make([]map[string]string, 0)
	toImport := make([]*database.Todo2Task, 0, len(merged))
	skippedSameContent := 0

	for id, t := range merged {
		if th, ok := targetHash[id]; ok {
			if th == importTaskContentKey(t) {
				skippedSameContent++
				continue // already identical in target (idempotent re-run)
			}
			item := map[string]string{
				"task_id": id,
				"reason":  "id exists in target with different content hash",
			}
			conflicts = append(conflicts, item)
			if onConflict == "fail" {
				continue
			}
			// skip: omit from import
			continue
		}

		label := sourceProvenance[id]
		if strings.TrimSpace(t.ProjectID) == "" {
			missingProject = append(missingProject, id)
			if defaultProj != "" {
				t.ProjectID = defaultProj
			} else {
				t.ProjectID = label
			}
		}

		toImport = append(toImport, t)
	}

	sort.Slice(toImport, func(i, j int) bool { return toImport[i].ID < toImport[j].ID })

	allowedIDs := make(map[string]bool, len(targetByID)+len(toImport))
	for id := range targetByID {
		allowedIDs[id] = true
	}
	for _, t := range toImport {
		allowedIDs[t.ID] = true
	}

	dangling := make([]map[string]string, 0)
	for _, t := range toImport {
		for _, d := range t.Dependencies {
			if !allowedIDs[d] {
				dangling = append(dangling, map[string]string{
					"task_id":         t.ID,
					"missing_dependency": d,
				})
			}
		}
	}

	for _, t := range toImport {
		var keep []string
		for _, d := range t.Dependencies {
			if allowedIDs[d] {
				keep = append(keep, d)
			}
		}
		t.Dependencies = keep
	}

	ordered, err := topologicalImportOrder(toImport)
	if err != nil {
		return nil, fmt.Errorf("import_sqlite: %w", err)
	}

	if onConflict == "fail" && len(conflicts) > 0 {
		out := map[string]interface{}{
			"success":               false,
			"method":                "import_sqlite",
			"dry_run":               dryRun,
			"import_max_depth":      maxDepth,
			"resolved_sources":      resolved,
			"conflicts":             conflicts,
			"missing_project_fixed": len(missingProject),
			"would_import":          len(ordered),
			"skipped_same_content":  skippedSameContent,
			"dropped_dependencies":  dangling,
			"warnings":              warnings,
		}
		if skippedSelf > 0 {
			out["skipped_self_db"] = skippedSelf
		}

		return framework.FormatResult(out, "")
	}

	inserted := 0
	skippedAtInsert := 0
	if !dryRun {
		for _, t := range ordered {
			if ex, err := database.GetTask(ctx, t.ID); err == nil && ex != nil {
				if importTaskContentKey(ex) == importTaskContentKey(t) {
					skippedAtInsert++

					continue
				}

				return nil, fmt.Errorf("import_sqlite: task %s exists with different content than source (race or stale plan)", t.ID)
			}

			if err := database.CreateTask(ctx, t); err != nil {
				if isSQLiteUniqueViolation(err) {
					ex2, err2 := database.GetTask(ctx, t.ID)
					if err2 == nil && ex2 != nil && importTaskContentKey(ex2) == importTaskContentKey(t) {
						skippedAtInsert++

						continue
					}
				}

				return nil, fmt.Errorf("import_sqlite: create %s: %w", t.ID, err)
			}
			inserted++
		}
		// Avoid rewriting JSON when nothing was inserted (idempotent re-runs stay quiet).
		if syncJSON && inserted > 0 {
			if err := SyncTodo2Tasks(projectRoot); err != nil {
				return nil, fmt.Errorf("import_sqlite: post-import SyncTodo2Tasks: %w", err)
			}
		}
		if inserted > 0 {
			MarkTaskResourcesChanged(ctx)
		}
	} else {
		inserted = len(ordered)
	}

	out := map[string]interface{}{
		"success":               true,
		"method":                "import_sqlite",
		"dry_run":               dryRun,
		"import_max_depth":      maxDepth,
		"resolved_sources":      resolved,
		"imported_count":        inserted,
		"skipped_same_content": skippedSameContent,
		"skipped_at_insert":    skippedAtInsert,
		"conflicts_skipped":    conflicts,
		"missing_project_ids":   missingProject,
		"dropped_dependencies":  dangling,
		"warnings":              warnings,
		"import_sync_json": syncJSON && !dryRun && inserted > 0,
	}
	if skippedSelf > 0 {
		out["skipped_self_db"] = skippedSelf
	}

	return framework.FormatResult(out, "")
}
