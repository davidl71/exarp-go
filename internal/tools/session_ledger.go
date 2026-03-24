// session_ledger.go — Auto-compaction ledger: continuity notes written when context budget threshold is reached.
// Ledgers are stored as thoughts/ledgers/CONTINUITY_{unix_ts}.md and injected into session prime.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/models"
	"github.com/spf13/cast"
)

// writeCompactionLedger creates a CONTINUITY_{ts}.md ledger in {projectRoot}/thoughts/ledgers/.
// Returns the path to the written file. Gathers in-progress tasks, latest handoff summary, and git branch.
func writeCompactionLedger(ctx context.Context, projectRoot string, params map[string]interface{}) (string, error) {
	dir := filepath.Join(projectRoot, "thoughts", "ledgers")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create ledger directory: %w", err)
	}

	ts := time.Now()
	filename := fmt.Sprintf("CONTINUITY_%d.md", ts.Unix())
	path := filepath.Join(dir, filename)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Continuity Ledger — %s\n\n", ts.Format("2006-01-02 15:04:05")))
	sb.WriteString("<!-- Auto-written by session action=prime when context budget threshold exceeded -->\n\n")

	// In-progress tasks
	store := NewDefaultTaskStore(projectRoot)
	if list, err := store.ListTasks(ctx, nil); err == nil {
		var inProgress []Todo2Task
		for _, t := range list {
			if t.Status == models.StatusInProgress {
				inProgress = append(inProgress, *t)
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
	}

	// Latest handoff summary
	handoffFile := filepath.Join(projectRoot, ".todo2", "handoffs.json")
	if data, err := os.ReadFile(handoffFile); err == nil {
		if summary := extractLatestHandoffSummary(data); summary != "" {
			sb.WriteString("## Latest Handoff Summary\n\n")
			sb.WriteString(summary + "\n\n")
		}
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

	// Optional caller-provided summary
	if summary := cast.ToString(params["ledger_summary"]); summary != "" {
		sb.WriteString("## Notes\n\n")
		sb.WriteString(summary + "\n")
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

// extractLatestHandoffSummary extracts the summary string from the last entry in handoffs.json bytes.
func extractLatestHandoffSummary(data []byte) string {
	var handoffData map[string]interface{}
	if err := json.Unmarshal(data, &handoffData); err != nil {
		return ""
	}
	handoffs, _ := handoffData["handoffs"].([]interface{})
	if len(handoffs) == 0 {
		return ""
	}
	last, _ := handoffs[len(handoffs)-1].(map[string]interface{})
	summary, _ := last["summary"].(string)
	return summary
}
