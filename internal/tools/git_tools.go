// git_tools.go — Git tools: types, branch/commit helpers, and HandleGitToolsNative dispatcher.
// See also: git_tools_actions.go
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// ─── Contents ───────────────────────────────────────────────────────────────
//   GitToolsParams — GitToolsParams represents parameters for git_tools.
//   TaskCommit — TaskCommit represents a commit (change) to a task.
//   extractBranchFromTags — extractBranchFromTags extracts branch name from task tags.
//   getTaskBranch — getTaskBranch gets branch for a task from its tags.
//   setTaskBranch — setTaskBranch sets branch for a task by updating its tags.
//   getAllBranches — getAllBranches extracts all unique branches from tasks.
//   filterTasksByBranch — filterTasksByBranch filters tasks to only those in a specific branch.
//   BranchStatistics — BranchStatistics represents statistics for a branch.
//   getBranchStatistics — getBranchStatistics gets statistics for a specific branch.
//   getAllBranchStatistics — getAllBranchStatistics gets statistics for all branches.
//   loadCommits — loadCommits loads commits from .todo2/commits.json.
//   saveCommits — saveCommits saves commits to .todo2/commits.json.
//   HandleGitToolsNative — HandleGitToolsNative handles git_tools using native Go implementation.
//   GitLocalCommit — GitLocalCommit represents a Git repo commit (not a task commit).
//   categorizeCommit — categorizeCommit assigns category from conventional commit patterns.
//   handleLocalCommitsAction — handleLocalCommitsAction lists Git commits ahead of origin/main and categorizes them (T-1768312778714).
//   handleCommitsAction — handleCommitsAction handles the commits action.
// ────────────────────────────────────────────────────────────────────────────

// Branch tag prefix.
const (
	branchTagPrefix = "branch:"
	mainBranch      = "main"
)

// ─── GitToolsParams ─────────────────────────────────────────────────────────
// GitToolsParams represents parameters for git_tools.
type GitToolsParams struct {
	Action           string `json:"action"`
	TaskID           string `json:"task_id,omitempty"`
	Branch           string `json:"branch,omitempty"`
	Limit            int    `json:"limit,omitempty"`
	Commit1          string `json:"commit1,omitempty"`
	Commit2          string `json:"commit2,omitempty"`
	Time1            string `json:"time1,omitempty"`
	Time2            string `json:"time2,omitempty"`
	Format           string `json:"format,omitempty"`
	OutputPath       string `json:"output_path,omitempty"`
	MaxCommits       int    `json:"max_commits,omitempty"`
	SourceBranch     string `json:"source_branch,omitempty"`
	TargetBranch     string `json:"target_branch,omitempty"`
	ConflictStrategy string `json:"conflict_strategy,omitempty"`
	Author           string `json:"author,omitempty"`
	DryRun           bool   `json:"dry_run,omitempty"`
}

// ─── TaskCommit ─────────────────────────────────────────────────────────────
// TaskCommit represents a commit (change) to a task.
type TaskCommit struct {
	ID        string                 `json:"id"`
	TaskID    string                 `json:"task_id,omitempty"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Author    string                 `json:"author"`
	Branch    string                 `json:"branch"`
	OldState  map[string]interface{} `json:"old_state,omitempty"`
	NewState  map[string]interface{} `json:"new_state,omitempty"`
}

// Branch utilities

// ─── extractBranchFromTags ──────────────────────────────────────────────────
// extractBranchFromTags extracts branch name from task tags.
func extractBranchFromTags(tags []string) string {
	if tags == nil {
		return mainBranch
	}

	for _, tag := range tags {
		if strings.HasPrefix(tag, branchTagPrefix) {
			branch := strings.TrimSpace(strings.TrimPrefix(tag, branchTagPrefix))
			if branch != "" {
				return branch
			}
		}
	}

	return mainBranch
}

// ─── getTaskBranch ──────────────────────────────────────────────────────────
// getTaskBranch gets branch for a task from its tags.
func getTaskBranch(task Todo2Task) string {
	branch := extractBranchFromTags(task.Tags)
	if branch == "" {
		return mainBranch
	}

	return branch
}

// ─── setTaskBranch ──────────────────────────────────────────────────────────
// setTaskBranch sets branch for a task by updating its tags.
func setTaskBranch(task *Todo2Task, branch string) {
	// Remove existing branch tags
	newTags := make([]string, 0, len(task.Tags))

	for _, tag := range task.Tags {
		if !strings.HasPrefix(tag, branchTagPrefix) {
			newTags = append(newTags, tag)
		}
	}

	// Add new branch tag (skip if main branch)
	if branch != mainBranch && branch != "" {
		newTags = append(newTags, branchTagPrefix+branch)
	}

	task.Tags = newTags
}

// ─── getAllBranches ─────────────────────────────────────────────────────────
// getAllBranches extracts all unique branches from tasks.
func getAllBranches(tasks []Todo2Task) []string {
	branches := make(map[string]bool)
	branches[mainBranch] = true

	for _, task := range tasks {
		branch := getTaskBranch(task)
		branches[branch] = true
	}

	result := make([]string, 0, len(branches))
	for branch := range branches {
		result = append(result, branch)
	}

	sort.Strings(result)

	return result
}

// ─── filterTasksByBranch ────────────────────────────────────────────────────
// filterTasksByBranch filters tasks to only those in a specific branch.
func filterTasksByBranch(tasks []Todo2Task, branch string) []Todo2Task {
	result := make([]Todo2Task, 0)

	for _, task := range tasks {
		if getTaskBranch(task) == branch {
			result = append(result, task)
		}
	}

	return result
}

// ─── BranchStatistics ───────────────────────────────────────────────────────
// BranchStatistics represents statistics for a branch.
type BranchStatistics struct {
	Branch         string         `json:"branch"`
	TaskCount      int            `json:"task_count"`
	ByStatus       map[string]int `json:"by_status"`
	CompletedCount int            `json:"completed_count"`
	CompletionRate float64        `json:"completion_rate"`
}

// ─── getBranchStatistics ────────────────────────────────────────────────────
// getBranchStatistics gets statistics for a specific branch.
func getBranchStatistics(tasks []Todo2Task, branch string) BranchStatistics {
	branchTasks := filterTasksByBranch(tasks, branch)

	statusCounts := make(map[string]int)
	completedCount := 0

	for _, task := range branchTasks {
		status := normalizeStatus(task.Status)
		statusCounts[status]++

		if IsCompletedStatus(status) {
			completedCount++
		}
	}

	total := len(branchTasks)
	completionRate := 0.0

	if total > 0 {
		completionRate = float64(completedCount) / float64(total) * 100
	}

	return BranchStatistics{
		Branch:         branch,
		TaskCount:      total,
		ByStatus:       statusCounts,
		CompletedCount: completedCount,
		CompletionRate: completionRate,
	}
}

// ─── getAllBranchStatistics ─────────────────────────────────────────────────
// getAllBranchStatistics gets statistics for all branches.
func getAllBranchStatistics(tasks []Todo2Task) map[string]BranchStatistics {
	branches := getAllBranches(tasks)
	result := make(map[string]BranchStatistics)

	for _, branch := range branches {
		result[branch] = getBranchStatistics(tasks, branch)
	}

	return result
}

// Commit tracking utilities

// ─── loadCommits ────────────────────────────────────────────────────────────
// loadCommits loads commits from .todo2/commits.json.
func loadCommits(projectRoot string) ([]TaskCommit, error) {
	commitsPath := filepath.Join(projectRoot, ".todo2", "commits.json")

	data, err := os.ReadFile(commitsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []TaskCommit{}, nil
		}

		return nil, fmt.Errorf("failed to read commits file: %w", err)
	}

	var commitsData struct {
		Commits []TaskCommit `json:"commits"`
		Version string       `json:"version"`
	}

	if err := json.Unmarshal(data, &commitsData); err != nil {
		return nil, fmt.Errorf("failed to parse commits JSON: %w", err)
	}

	return commitsData.Commits, nil
}

// ─── saveCommits ────────────────────────────────────────────────────────────
// saveCommits saves commits to .todo2/commits.json.
func saveCommits(projectRoot string, commits []TaskCommit) error {
	commitsPath := filepath.Join(projectRoot, ".todo2", "commits.json")

	// Ensure directory exists
	dir := filepath.Dir(commitsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create .todo2 directory: %w", err)
	}

	commitsData := struct {
		Commits []TaskCommit `json:"commits"`
		Version string       `json:"version"`
	}{
		Commits: commits,
		Version: "1.0",
	}

	data, err := json.MarshalIndent(commitsData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal commits: %w", err)
	}

	if err := os.WriteFile(commitsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write commits file: %w", err)
	}

	return nil
}

// ─── HandleGitToolsNative ───────────────────────────────────────────────────
// HandleGitToolsNative handles git_tools using native Go implementation.
func HandleGitToolsNative(ctx context.Context, params GitToolsParams) (string, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return "", fmt.Errorf("failed to find project root: %w", err)
	}

	action := params.Action
	if action == "" {
		action = "commits"
	}

	switch action {
	case "local_commits":
		return handleLocalCommitsAction(ctx, projectRoot, params)
	case "commits":
		return handleCommitsAction(ctx, projectRoot, params)
	case "branches":
		return handleBranchesAction(ctx, projectRoot, params)
	case "tasks":
		return handleTasksAction(ctx, projectRoot, params)
	case "diff":
		return handleDiffAction(ctx, projectRoot, params)
	case "graph":
		return handleGraphAction(ctx, projectRoot, params)
	case "merge":
		return handleMergeAction(ctx, projectRoot, params)
	case "set_branch":
		return handleSetBranchAction(ctx, projectRoot, params)
	default:
		return "", fmt.Errorf("unknown action: %s", action)
	}
}

// ─── GitLocalCommit ─────────────────────────────────────────────────────────
// GitLocalCommit represents a Git repo commit (not a task commit).
type GitLocalCommit struct {
	Hash     string `json:"hash"`
	Subject  string `json:"subject"`
	Category string `json:"category"`
	Relevant bool   `json:"relevant"`
}

// ─── categorizeCommit ───────────────────────────────────────────────────────
// categorizeCommit assigns category from conventional commit patterns.
func categorizeCommit(subject string) string {
	lower := strings.ToLower(subject)
	if strings.HasPrefix(lower, "feat") || strings.HasPrefix(lower, "feature") {
		return "feature"
	}

	if strings.HasPrefix(lower, "fix") || strings.HasPrefix(lower, "bugfix") {
		return "fix"
	}

	if strings.HasPrefix(lower, "docs") || strings.HasPrefix(lower, "doc") {
		return "docs"
	}

	if strings.HasPrefix(lower, "refactor") {
		return "refactor"
	}

	if strings.HasPrefix(lower, "test") {
		return "testing"
	}

	if strings.HasPrefix(lower, "chore") || strings.HasPrefix(lower, "ci") || strings.HasPrefix(lower, "build") {
		return "cleanup"
	}

	return "other"
}

// ─── handleLocalCommitsAction ───────────────────────────────────────────────
// handleLocalCommitsAction lists Git commits ahead of origin/main and categorizes them (T-1768312778714).
func handleLocalCommitsAction(ctx context.Context, projectRoot string, params GitToolsParams) (string, error) {
	upstream := "origin/main"
	if params.TargetBranch != "" {
		upstream = params.TargetBranch
	}

	cmd := exec.CommandContext(ctx, "git", "log", upstream+"..HEAD", "--format=%h %s")
	cmd.Dir = projectRoot

	cmd.Env = append(os.Environ(), "LANG=C")

	output, err := cmd.Output()
	if err != nil {
		if strings.Contains(err.Error(), "executable file not found") ||
			strings.Contains(err.Error(), "not a git repository") {
			result := map[string]interface{}{
				"action":      "local_commits",
				"upstream":    upstream,
				"total":       0,
				"commits":     []GitLocalCommit{},
				"by_category": map[string]int{},
				"summary":     "no local commits or not a git repository",
			}
			data, _ := json.MarshalIndent(result, "", "  ")

			return string(data), nil
		}

		return "", fmt.Errorf("git log failed: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	commits := make([]GitLocalCommit, 0, len(lines))
	byCategory := make(map[string]int)
	hashRe := regexp.MustCompile(`^([a-f0-9]+)\s+(.*)$`)

	limit := params.Limit
	if limit <= 0 {
		limit = 50
	}

	for i, line := range lines {
		if i >= limit {
			break
		}

		if line == "" {
			continue
		}

		matches := hashRe.FindStringSubmatch(line)
		hash, subject := "", line

		if len(matches) >= 3 {
			hash, subject = matches[1], matches[2]
		}

		category := categorizeCommit(subject)
		commits = append(commits, GitLocalCommit{
			Hash:     hash,
			Subject:  subject,
			Category: category,
			Relevant: category != "cleanup" && category != "other",
		})
		byCategory[category]++
	}

	relevant := byCategory["feature"] + byCategory["fix"] + byCategory["docs"] + byCategory["refactor"] + byCategory["testing"]
	result := map[string]interface{}{
		"action":      "local_commits",
		"upstream":    upstream,
		"total":       len(commits),
		"commits":     commits,
		"by_category": byCategory,
		"summary":     fmt.Sprintf("%d commit(s) ahead of %s; %d likely relevant", len(commits), upstream, relevant),
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(data), nil
}

// ─── handleCommitsAction ────────────────────────────────────────────────────
// handleCommitsAction handles the commits action.
func handleCommitsAction(ctx context.Context, projectRoot string, params GitToolsParams) (string, error) {
	commits, err := loadCommits(projectRoot)
	if err != nil {
		return "", err
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 50
	}

	var filtered []TaskCommit

	if params.TaskID != "" {
		// Filter by task ID
		for _, commit := range commits {
			if commit.TaskID == params.TaskID {
				if params.Branch == "" || commit.Branch == params.Branch {
					filtered = append(filtered, commit)
				}
			}
		}
	} else if params.Branch != "" {
		// Filter by branch
		for _, commit := range commits {
			if commit.Branch == params.Branch {
				filtered = append(filtered, commit)
			}
		}
	} else {
		return "", fmt.Errorf("task_id or branch required for commits action")
	}

	// Sort by timestamp (newest first)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp.After(filtered[j].Timestamp)
	})

	// Limit results
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	result := map[string]interface{}{
		"task_id":       params.TaskID,
		"branch":        params.Branch,
		"total_commits": len(filtered),
		"commits":       filtered,
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(data), nil
}

// handleBranchesAction handles the branches action.
