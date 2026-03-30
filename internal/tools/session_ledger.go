// session_ledger.go — Auto-compaction ledger: continuity notes written when context budget threshold is reached.
// Ledgers are stored as thoughts/ledgers/CONTINUITY_{unix_ts}.md and injected into session prime.
package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/models"
	"github.com/spf13/cast"
)

type continuityLedgerOptions struct {
	Reason          string
	Summary         string
	Blockers        []string
	NextSteps       []string
	TaskJournal     []map[string]interface{}
	TasksInProgress []Todo2Task
	Notes           string
}

// writeCompactionLedger creates a CONTINUITY_{ts}.md ledger in {projectRoot}/thoughts/ledgers/.
// Returns the path to the written file. Gathers in-progress tasks, latest handoff summary, and git branch.
func writeCompactionLedger(ctx context.Context, projectRoot string, params map[string]interface{}) (string, error) {
	var notes string
	if summary := cast.ToString(params["ledger_summary"]); summary != "" {
		notes = summary
	}

	return writeContinuityLedger(ctx, projectRoot, continuityLedgerOptions{
		Reason: "context_threshold",
		Notes:  notes,
	})
}

func writeContinuityLedger(ctx context.Context, projectRoot string, opts continuityLedgerOptions) (string, error) {
	dir := filepath.Join(projectRoot, "thoughts", "ledgers")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create ledger directory: %w", err)
	}

	ts := time.Now()
	filename := fmt.Sprintf("CONTINUITY_%d.md", ts.Unix())
	path := filepath.Join(dir, filename)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Continuity Ledger — %s\n\n", ts.Format("2006-01-02 15:04:05")))
	switch opts.Reason {
	case "handoff":
		sb.WriteString("<!-- Auto-written by session action=handoff to preserve execution continuity -->\n\n")
	default:
		sb.WriteString("<!-- Auto-written by session action=prime when context budget threshold exceeded -->\n\n")
	}

	if opts.Summary != "" {
		sb.WriteString("## Summary\n\n")
		sb.WriteString(opts.Summary + "\n\n")
	}

	if len(opts.Blockers) > 0 {
		sb.WriteString("## Blockers\n\n")
		for _, blocker := range opts.Blockers {
			sb.WriteString("- " + blocker + "\n")
		}
		sb.WriteString("\n")
	}

	if len(opts.NextSteps) > 0 {
		sb.WriteString("## Next Steps\n\n")
		for _, step := range opts.NextSteps {
			sb.WriteString("- " + step + "\n")
		}
		sb.WriteString("\n")
	}

	// In-progress tasks
	inProgress := opts.TasksInProgress
	if len(inProgress) == 0 {
		store := NewDefaultTaskStore(projectRoot)
		if list, err := store.ListTasks(ctx, nil); err == nil {
			for _, t := range list {
				if t.Status == models.StatusInProgress {
					inProgress = append(inProgress, *t)
				}
			}
		}
	}
	if len(inProgress) > 0 {
		sb.WriteString("## In-Progress Tasks\n\n")
		for _, t := range inProgress {
			line := fmt.Sprintf("- **%s** — %s", t.ID, t.Content)
			if t.LongDescription != "" {
				line += "\n  " + truncateString(t.LongDescription, 120)
			}
			sb.WriteString(line + "\n")
		}
		sb.WriteString("\n")
	}

	if len(opts.TaskJournal) > 0 {
		sb.WriteString("## Task Journal\n\n")
		for _, entry := range opts.TaskJournal {
			id, _ := entry["id"].(string)
			action, _ := entry["action"].(string)
			if id == "" && action == "" {
				continue
			}

			if action == "" {
				action = "modified"
			}

			line := fmt.Sprintf("- **%s** — %s", id, action)
			if summary, ok := entry["summary"].(string); ok && strings.TrimSpace(summary) != "" {
				line += ": " + summary
			}
			sb.WriteString(line + "\n")
		}
		sb.WriteString("\n")
	}

	if summary := latestHandoffSummaryFromProject(projectRoot); summary != "" {
		sb.WriteString("## Latest Handoff Summary\n\n")
		sb.WriteString(summary + "\n\n")
	}

	// Git branch and uncommitted files
	gitSt := getGitStatus(ctx, projectRoot)
	if branch, ok := gitSt["branch"].(string); ok && branch != "" {
		sb.WriteString(fmt.Sprintf("## Git Branch\n\n`%s`\n\n", branch))
	}
	if n, ok := gitSt["uncommitted_files"].(int); ok && n > 0 {
		sb.WriteString(fmt.Sprintf("Uncommitted files: %d\n\n", n))
		if files, ok := gitSt["changed_files"].([]string); ok && len(files) > 0 {
			for _, f := range files {
				sb.WriteString("- " + f + "\n")
			}
			sb.WriteString("\n")
		}
	}

	if opts.Notes != "" {
		sb.WriteString("## Notes\n\n")
		sb.WriteString(opts.Notes + "\n")
	}

	if err := os.WriteFile(path, []byte(sb.String()), 0o644); err != nil {
		return "", fmt.Errorf("failed to write ledger: %w", err)
	}

	return path, nil
}

// readLatestLedger returns the content and path of the most recent CONTINUITY_*.md in {projectRoot}/thoughts/ledgers/.
// Returns empty strings if none found.
func readLatestLedger(projectRoot string) (content, path string) {
	dir := filepath.Join(projectRoot, "thoughts", "ledgers")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", ""
	}

	var ledgers []string
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() && strings.HasPrefix(name, "CONTINUITY_") && strings.HasSuffix(name, ".md") {
			ledgers = append(ledgers, filepath.Join(dir, name))
		}
	}

	if len(ledgers) == 0 {
		return "", ""
	}

	sort.Strings(ledgers)
	latest := ledgers[len(ledgers)-1]

	data, err := os.ReadFile(latest)
	if err != nil {
		return "", ""
	}

	return string(data), latest
}

func readLatestLedgerSummary(projectRoot string) map[string]interface{} {
	content, path := readLatestLedger(projectRoot)
	if content == "" {
		return nil
	}

	return map[string]interface{}{
		"path":    path,
		"excerpt": truncateString(content, 400),
	}
}

// latestHandoffSummaryFromProject loads the handoff store and returns the latest entry summary.
func latestHandoffSummaryFromProject(projectRoot string) string {
	if !handoffsAnyFileExists(projectRoot) {
		return ""
	}
	store, err := loadHandoffStore(projectRoot)
	if err != nil {
		return ""
	}
	return latestHandoffSummary(store)
}
