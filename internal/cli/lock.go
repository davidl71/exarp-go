package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/tools"
	"github.com/davidl71/exarp-go/internal/utils"
	mcpcli "github.com/davidl71/mcp-go-core/pkg/mcp/cli"
)

// handleLockCommand handles the lock subcommand (e.g. lock status).
// T-316: CLI command for lock status monitoring.
func handleLockCommand(parsed *mcpcli.Args) error {
	subcommand := parsed.Subcommand
	if subcommand == "" && len(parsed.Positional) > 0 {
		subcommand = parsed.Positional[0]
	}

	if subcommand == "" {
		return showLockUsage()
	}

	switch subcommand {
	case "status":
		return handleLockStatus(parsed)
	case "help":
		return showLockUsage()
	default:
		return fmt.Errorf("unknown lock command: %s (use: status, help)", subcommand)
	}
}

func handleLockStatus(parsed *mcpcli.Args) error {
	projectRoot, err := tools.FindProjectRoot()
	if err != nil {
		return fmt.Errorf("project root: %w", err)
	}

	EnsureConfigAndDatabase(projectRoot)

	ctx := context.Background()
	nearExpiry := 5 * time.Minute

	// Task locks (from database)
	info, err := database.DetectStaleLocks(ctx, nearExpiry)
	if err != nil {
		return fmt.Errorf("task locks: %w", err)
	}

	// Git sync lock (file-based)
	gitLockPath := utils.GitSyncLockPath(projectRoot)
	gitHeld := isGitLockHeld(gitLockPath)

	useJSON := parsed.GetBoolFlag("json", false) || parsed.GetBoolFlag("j", false)
	if useJSON {
		return printLockStatusJSON(info, gitLockPath, gitHeld)
	}

	printLockStatusText(info, gitLockPath, gitHeld, nearExpiry)

	return nil
}

func isGitLockHeld(path string) bool {
	fl, err := utils.NewFileLock(path, 0)
	if err != nil {
		return false // e.g. dir missing
	}

	defer func() {
		if closeErr := fl.Close(); closeErr != nil {
			logDebug(context.Background(), "Failed to close git lock file", "error", closeErr, "operation", "isGitLockHeld")
		}
	}()

	err = fl.TryLock()
	if err != nil {
		return true // lock is held by another process
	}

	_ = fl.Unlock()

	return false
}

func printLockStatusText(info *database.StaleLockInfo, gitLockPath string, gitHeld bool, nearExpiry time.Duration) {
	fmt.Println("## Task locks (database)")
	fmt.Printf("  Active (locked): %d\n", len(info.Locks))
	fmt.Printf("  Expired:         %d\n", info.ExpiredCount)
	fmt.Printf("  Near expiry:     %d\n", info.NearExpiryCount)
	fmt.Printf("  Stale (>5m):     %d\n", info.StaleCount)

	if len(info.Locks) > 0 {
		fmt.Println()

		for _, l := range info.Locks {
			state := "active"
			if l.IsStale {
				state = "stale"
			} else if l.IsExpired {
				state = "expired"
			} else if l.TimeRemaining < nearExpiry {
				state = "near expiry"
			}

			fmt.Printf("  %s  %s  until %s  (%s)\n", l.TaskID, l.Assignee, l.LockUntil.Format(time.RFC3339), state)
		}
	}

	fmt.Println()
	fmt.Println("## Git sync lock (file)")
	fmt.Printf("  Path: %s\n", gitLockPath)

	if _, err := os.Stat(gitLockPath); err != nil {
		fmt.Println("  Status: no lock file")
	} else if gitHeld {
		fmt.Println("  Status: held")
	} else {
		fmt.Println("  Status: free")
	}
}

func printLockStatusJSON(info *database.StaleLockInfo, gitLockPath string, gitHeld bool) error {
	type taskLock struct {
		TaskID        string `json:"task_id"`
		Assignee      string `json:"assignee"`
		LockUntil     string `json:"lock_until"`
		IsExpired     bool   `json:"is_expired"`
		IsStale       bool   `json:"is_stale"`
		TimeRemaining string `json:"time_remaining,omitempty"`
		TimeExpired   string `json:"time_expired,omitempty"`
	}

	locks := make([]taskLock, 0, len(info.Locks))

	for _, l := range info.Locks {
		ent := taskLock{
			TaskID:    l.TaskID,
			Assignee:  l.Assignee,
			LockUntil: l.LockUntil.Format(time.RFC3339),
			IsExpired: l.IsExpired,
			IsStale:   l.IsStale,
		}
		if l.TimeRemaining > 0 {
			ent.TimeRemaining = l.TimeRemaining.Round(time.Second).String()
		}

		if l.TimeExpired > 0 {
			ent.TimeExpired = l.TimeExpired.Round(time.Second).String()
		}

		locks = append(locks, ent)
	}

	_, err := os.Stat(gitLockPath)
	gitExists := err == nil

	out := struct {
		TaskLocks struct {
			Total      int        `json:"total"`
			Expired    int        `json:"expired"`
			NearExpiry int        `json:"near_expiry"`
			Stale      int        `json:"stale"`
			Locks      []taskLock `json:"locks"`
		} `json:"task_locks"`
		GitLock struct {
			Path   string `json:"path"`
			Exists bool   `json:"exists"`
			Held   bool   `json:"held"`
		} `json:"git_lock"`
	}{}
	out.TaskLocks.Total = len(info.Locks)
	out.TaskLocks.Expired = info.ExpiredCount
	out.TaskLocks.NearExpiry = info.NearExpiryCount
	out.TaskLocks.Stale = info.StaleCount
	out.TaskLocks.Locks = locks
	out.GitLock.Path = gitLockPath
	out.GitLock.Exists = gitExists
	out.GitLock.Held = gitHeld

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(data))

	return nil
}

func showLockUsage() error {
	fmt.Println(`Usage: exarp-go lock <subcommand> [options]

Subcommands:
  status    Show task lock and Git sync lock status (default)
  help      Show this message

Options for status:
  -json, -j   Output as JSON

Examples:
  exarp-go lock status
  exarp-go lock status -json`)

	return nil
}
