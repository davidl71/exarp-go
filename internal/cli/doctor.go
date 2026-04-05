// doctor.go — lightweight environment checks for agents and operators.
package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/projectroot"
	"github.com/davidl71/exarp-go/internal/tools"
)

// RunDoctor prints project root, database path, migrations dir, and binary path.
// It does not mutate state; use before debugging "task update did nothing" reports.

func RunDoctor() error {
	var root string
	projectRoot, err := tools.FindProjectRoot()
	if err != nil {
		fmt.Printf("project_root: (not found) %v\n", err)
	} else {
		root = projectRoot
		fmt.Printf("project_root: %s\n", root)
		if env := strings.TrimSpace(os.Getenv("PROJECT_ROOT")); env != "" &&
			!strings.Contains(env, "{{PROJECT_ROOT}}") {
			if absEnv, e := filepath.Abs(env); e == nil {
				if cwd, e2 := os.Getwd(); e2 == nil {
					if cwdRoot, e3 := projectroot.FindFrom(cwd); e3 == nil {
						envNorm, _ := filepath.EvalSymlinks(absEnv)
						cwdNorm, _ := filepath.EvalSymlinks(cwdRoot)
						effectNorm, _ := filepath.EvalSymlinks(root)
						if envNorm != cwdNorm && effectNorm == cwdNorm &&
							projectroot.IsExarpGoSourceRoot(envNorm) {
							fmt.Printf("note: PROJECT_ROOT=%s points at exarp-go source; effective root is cwd-based %s (unset PROJECT_ROOT or set EXARP_STRICT_PROJECT_ROOT=1 to force env).\n", absEnv, root)
						}
					}
				}
			}
		}
		dbPath := filepath.Join(root, ".todo2", "todo2.db")
		if st, e := os.Stat(dbPath); e != nil {
			fmt.Printf("todo2_db: %s (missing: %v)\n", dbPath, e)
		} else {
			fmt.Printf("todo2_db: %s (ok, size=%d)\n", dbPath, st.Size())
		}
		wavesPath := filepath.Join(root, ".cursor", "plans", "parallel-execution-waves.json")
		if _, e := os.Stat(wavesPath); e != nil {
			fmt.Printf("parallel_waves_json: %s (optional; missing)\n", wavesPath)
		} else {
			fmt.Printf("parallel_waves_json: %s (ok)\n", wavesPath)
		}
	}

	mig := os.Getenv("EXARP_MIGRATIONS_DIR")
	if mig != "" {
		fmt.Printf("EXARP_MIGRATIONS_DIR: %s\n", mig)
	} else if root != "" {
		fmt.Printf("EXARP_MIGRATIONS_DIR: (unset; derived from binary when using run_exarp_go.sh)\n")
	}

	if exe, e := os.Executable(); e == nil {
		fmt.Printf("exarp_binary: %s\n", exe)
		fmt.Printf("hint: build CLI with: go build -o bin/exarp-go ./cmd/server\n")
	}
	fmt.Printf("hint: list tasks with: task list --status Todo --json (use this repo as cwd / resolved project_root; avoid ad-hoc sqlite3 unless debugging schema)\n")

	if root != "" {
		if _, err := config.LoadConfig(root); err != nil {
			fmt.Printf("centralized_config: (optional) %v\n", err)
		} else {
			fmt.Printf("centralized_config: loaded\n")
		}
	}

	if initializeDatabaseForDoctor() {
		dbx, err := database.GetDBx()
		if err != nil {
			fmt.Printf("task_rows_in_db: (db) %v\n", err)
		} else {
			var n int
			if err := dbx.GetContext(context.Background(), &n, `SELECT COUNT(*) FROM tasks`); err != nil {
				fmt.Printf("task_rows_in_db: (error) %v\n", err)
			} else {
				fmt.Printf("task_rows_in_db: %d\n", n)
			}
		}
	}

	return nil
}

func initializeDatabaseForDoctor() bool {
	projectRoot, err := tools.FindProjectRoot()
	if err != nil {
		return false
	}
	EnsureConfigAndDatabase(projectRoot)
	return true
}
