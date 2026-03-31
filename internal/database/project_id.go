// project_id.go — Default logical project identifier for task rows.
//
// SQLite lives at <PROJECT_ROOT>/.todo2/todo2.db per consumer repo; distinct roots never
// share a file unless the user copies or symlinks. project_id still tags rows for filtering
// and cross-project aggregation when IncludeNullProjectID is used.
package database

import (
	"os"
	"path/filepath"
	"strings"
)

// DefaultProjectIDFromEnv returns basename(PROJECT_ROOT) when that env is set and non-trivial,
// otherwise "default". Mirrors tools.NewDefaultTaskStore CreateTask behavior for callers
// that insert via database.CreateTask without going through the store.
func DefaultProjectIDFromEnv() string {
	root := strings.TrimSpace(os.Getenv("PROJECT_ROOT"))
	if root == "" || root == "." {
		return "default"
	}
	base := filepath.Base(root)
	if base == "" || base == "." {
		return "default"
	}
	return base
}
