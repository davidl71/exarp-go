//go:build ignore
// +build ignore

// delete_expired_archive deletes expired files in docs/archive per ARCHIVE_RETENTION_POLICY.md.
// Usage: go run scripts/delete_expired_archive.go [-dry-run] [-project-root PATH]
// Default: -dry-run (list what would be deleted). Omit -dry-run to actually delete.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/davidl71/exarp-go/internal/archive"
)

func main() {
	dryRun := flag.Bool("dry-run", true, "list files that would be deleted; do not delete")
	projectRoot := flag.String("project-root", "", "project root (default: detect from cwd)")
	flag.Parse()

	root := *projectRoot
	if root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "getwd: %v\n", err)
			os.Exit(1)
		}
		root = cwd
		// Prefer repo root (containing .todo2 or go.mod)
		for {
			if _, err := os.Stat(filepath.Join(root, ".todo2")); err == nil {
				break
			}
			if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
				break
			}
			parent := filepath.Dir(root)
			if parent == root {
				break
			}
			root = parent
		}
	}

	asOf := time.Now().UTC()
	deleted, err := archive.DeleteExpired(root, asOf, *dryRun)
	if err != nil {
		fmt.Fprintf(os.Stderr, "delete expired: %v\n", err)
		os.Exit(1)
	}

	if *dryRun {
		fmt.Println("Dry run — would delete:")
		for _, p := range deleted {
			fmt.Println(" ", p)
		}
		if len(deleted) == 0 {
			fmt.Println(" (none — no categories past their Delete After date)")
		}
	} else {
		for _, p := range deleted {
			fmt.Println("Deleted:", p)
		}
	}
}
