// task_discovery_common.go — Shared logic for discovering tasks from code comments and docs.
// Includes scanGitJSON (git-tracked JSON task state); used by CGO and nocgo task_discovery builds.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/models"
)

// tagPattern matches hashtags in TODO comments (e.g., #refactor, #bug, #performance).
var tagPattern = regexp.MustCompile(`#([a-zA-Z][a-zA-Z0-9_-]*)`)

// IsDeprecatedDiscoveryText returns true if the discovery text looks like a deprecated/removed
// item (e.g. strikethrough, "(removed)", "*(T-xxx removed)*") and should not be turned into
// a new Todo2 task. Used by scanMarkdown, scanMarkdownBasic, and createTasksFromDiscoveries.
func IsDeprecatedDiscoveryText(text string) bool {
	t := strings.TrimSpace(text)
	if t == "" {
		return true
	}

	lower := strings.ToLower(t)
	// Strikethrough in markdown (~~...~~)
	if strings.Contains(t, "~~") {
		return true
	}
	// Explicit "(removed)" or "*(T-xxx removed)*" style
	if strings.Contains(lower, "(removed)") || strings.Contains(lower, "removed)") {
		return true
	}
	// "Future improvement" in removed-context (often with T-xxx removed)
	if strings.Contains(lower, "future improvement") && strings.Contains(lower, "t-") {
		return true
	}

	return false
}

var defaultDiscoverySkipDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"__pycache__":  true,
	".venv":        true,
	"vendor":       true,
	".idea":        true,
	".vscode":      true,
	"dist":         true,
	"build":        true,
	"target":       true,
	"archive":      true,
	".cache":       true,
	"third_party":  true,
	"mcp-servers":  true,
}

func parseIgnorePaths(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	var parsed []string
	if strings.HasPrefix(raw, "[") {
		if err := json.Unmarshal([]byte(raw), &parsed); err == nil {
			return normalizeIgnorePaths(parsed)
		}
	}

	parts := strings.Split(raw, ",")
	return normalizeIgnorePaths(parts)
}

func normalizeIgnorePaths(paths []string) []string {
	normalized := make([]string, 0, len(paths))
	seen := make(map[string]bool)

	for _, path := range paths {
		path = strings.TrimSpace(path)
		path = normalizeDiscoveryPath(path)
		if path == "" || path == "." || seen[path] {
			continue
		}

		seen[path] = true
		normalized = append(normalized, path)
	}

	return normalized
}

func normalizeDiscoveryPath(path string) string {
	path = filepath.ToSlash(strings.TrimSpace(path))
	path = strings.TrimPrefix(path, "./")
	path = strings.Trim(path, "/")
	return path
}

func discoveryRelativePath(projectRoot, path string) string {
	rel, err := filepath.Rel(projectRoot, path)
	if err != nil {
		return normalizeDiscoveryPath(path)
	}

	return normalizeDiscoveryPath(rel)
}

func shouldSkipDiscoveryDir(projectRoot, path string, ignorePaths []string) bool {
	relPath := discoveryRelativePath(projectRoot, path)
	if relPath == "" {
		return false
	}

	parts := strings.Split(relPath, "/")
	for _, part := range parts {
		if defaultDiscoverySkipDirs[part] {
			return true
		}
	}

	if len(parts) > 0 && parts[0] == "bin" {
		return true
	}

	for _, ignorePath := range ignorePaths {
		if relPath == ignorePath || strings.HasPrefix(relPath, ignorePath+"/") {
			return true
		}
	}

	return false
}

func discoveryIgnorePathsForProject(projectRoot string, params map[string]interface{}) []string {
	ignorePaths := []string{}

	cfg, err := config.LoadConfig(projectRoot)
	if err == nil && cfg != nil && len(cfg.Project.TaskDiscoveryIgnorePaths) > 0 {
		ignorePaths = append(ignorePaths, cfg.Project.TaskDiscoveryIgnorePaths...)
	}

	if paramPaths := parseIgnorePaths(fmt.Sprint(params["ignore_paths"])); len(paramPaths) > 0 {
		ignorePaths = append(ignorePaths, paramPaths...)
	}

	return normalizeIgnorePaths(ignorePaths)
}

// isThirdPartyOrBinaryPath returns true if path indicates bin/ or third-party code (vendor, node_modules, etc.).
// Used to skip creating tasks from discoveries in those locations.
func isThirdPartyOrBinaryPath(path string) bool {
	pathLower := strings.ToLower(path)
	// Normalize separators for consistent matching
	pathNorm := strings.ReplaceAll(pathLower, "\\", "/")
	segments := []string{"/bin/", "/vendor/", "/node_modules/", "/__pycache__/", "/.venv/", "/dist/", "/build/", "/target/"}
	for _, seg := range segments {
		if strings.Contains(pathNorm, seg) {
			return true
		}
	}
	// Top-level bin or vendor
	if strings.HasPrefix(pathNorm, "bin/") || strings.HasPrefix(pathNorm, "vendor/") {
		return true
	}
	return false
}

// extractTagsFromText extracts hashtag-style tags from comment text.
// Returns a slice of tags (without the # prefix) and the text with tags removed.
func extractTagsFromText(text string) ([]string, string) {
	matches := tagPattern.FindAllStringSubmatch(text, -1)
	tags := []string{}
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			tag := strings.ToLower(match[1])
			if !seen[tag] {
				tags = append(tags, tag)
				seen[tag] = true
			}
		}
	}

	// Optionally remove tags from text for cleaner display
	cleanText := tagPattern.ReplaceAllString(text, "")
	cleanText = strings.TrimSpace(cleanText)
	// Clean up multiple spaces
	cleanText = regexp.MustCompile(`\s+`).ReplaceAllString(cleanText, " ")

	return tags, cleanText
}

// toJSONSafeString returns a string for use in JSON output; avoids non-scalar types in tool response.
func toJSONSafeString(v interface{}) string {
	if v == nil {
		return ""
	}

	if s, ok := v.(string); ok {
		return s
	}

	return fmt.Sprint(v)
}

// createTasksFromDiscoveries creates Todo2 tasks from discovered items.
// Shared by both CGO (task_discovery_native.go) and nocgo (task_discovery_native_nocgo.go) builds.
func createTasksFromDiscoveries(ctx context.Context, projectRoot string, discoveries []map[string]interface{}) []map[string]interface{} {
	createdTasks := []map[string]interface{}{}

	store := NewDefaultTaskStore(projectRoot)

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return createdTasks
	}

	existingTasks := tasksFromPtrs(list)

	existingContent := make(map[string]bool)
	for _, task := range existingTasks {
		existingContent[strings.ToLower(strings.TrimSpace(task.Content))] = true
	}

	for _, discovery := range discoveries {
		text, ok := discovery["text"].(string)
		if !ok || text == "" {
			continue
		}

		if IsDeprecatedDiscoveryText(text) {
			continue
		}

		// Skip discoveries from bin/ or third-party paths (vendor, node_modules, etc.)
		if file, _ := discovery["file"].(string); file != "" && isThirdPartyOrBinaryPath(file) {
			continue
		}

		textLower := strings.ToLower(strings.TrimSpace(text))
		if existingContent[textLower] {
			continue
		}

		taskID := generateEpochTaskID()

		sourceTag := "discovered"
		if src, ok := discovery["source"].(string); ok && src != "" {
			sourceTag = src
		}

		// Build tags list: start with discovered + source tag
		taskTags := []string{"discovered", sourceTag}

		// Add any tags extracted from the TODO comment
		if discoveredTags, ok := discovery["tags"].([]string); ok && len(discoveredTags) > 0 {
			for _, tag := range discoveredTags {
				// Avoid duplicates
				isDuplicate := false

				for _, existing := range taskTags {
					if existing == tag {
						isDuplicate = true
						break
					}
				}

				if !isDuplicate {
					taskTags = append(taskTags, tag)
				}
			}
		}

		metadata := map[string]interface{}{
			"discovery_type": discovery["type"],
		}
		if f, ok := discovery["file"]; ok {
			metadata["discovered_from"] = f
		}

		if line, ok := discovery["line"]; ok {
			metadata["discovered_line"] = line
		}
		// Sanitize so persisted state and DB never get non-JSON-serializable metadata
		newTask := &Todo2Task{
			ID:       taskID,
			Content:  text,
			Status:   models.StatusTodo,
			Priority: "medium",
			Tags:     taskTags,
			Metadata: database.SanitizeMetadataForWrite(metadata),
		}

		if err := store.CreateTask(ctx, newTask); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to create task %s after discovery: %v\n", taskID, err)
			continue
		}

		existingContent[textLower] = true

		// Coerce source to string so tool response JSON is always valid
		createdTasks = append(createdTasks, map[string]interface{}{
			"id":      taskID,
			"content": text,
			"source":  toJSONSafeString(discovery["source"]),
		})
	}

	return createdTasks
}

// scanGitJSON scans the git repository for tracked JSON files and extracts tasks (e.g. legacy state.todo2.json).
func scanGitJSON(ctx context.Context, projectRoot string, jsonPattern string) []map[string]interface{} {
	discoveries := []map[string]interface{}{}

	if jsonPattern == "" {
		jsonPattern = "**/.todo2/state.todo2.json"
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if err := ctx.Err(); err != nil {
		return discoveries
	}

	cmd := exec.CommandContext(ctx, "git", "ls-files", "*.json", "**/*.json")
	cmd.Dir = projectRoot
	output, err := cmd.Output()
	if err != nil {
		return discoveries
	}

	jsonFiles := []string{}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	defaultPattern := "**/.todo2/state.todo2.json"

	for _, line := range lines {
		if err := ctx.Err(); err != nil {
			return discoveries
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		patternToUse := jsonPattern
		if patternToUse == "" {
			patternToUse = defaultPattern
		}

		matched := false
		if patternToUse == defaultPattern {
			matched = strings.Contains(line, "state.todo2.json")
		} else {
			matched, _ = filepath.Match(patternToUse, line)
			if !matched {
				if strings.HasPrefix(patternToUse, "**/") {
					pattern := strings.TrimPrefix(patternToUse, "**/")
					matched, _ = filepath.Match(pattern, filepath.Base(line))
				}
				if !matched && strings.Contains(line, strings.TrimPrefix(patternToUse, "**/")) {
					matched = true
				}
			}
		}

		if matched {
			jsonFiles = append(jsonFiles, line)
		}
	}

	for _, jsonFile := range jsonFiles {
		if err := ctx.Err(); err != nil {
			return discoveries
		}

		fullPath := filepath.Join(projectRoot, jsonFile)

		cmd = exec.CommandContext(ctx, "git", "log", "--all", "--pretty=format:%H", "--", jsonFile)
		cmd.Dir = projectRoot
		commitOutput, err := cmd.Output()
		if err != nil {
			tasks, _, err := LoadJSONStateFromFile(fullPath)
			if err == nil {
				for _, task := range tasks {
					discoveries = append(discoveries, map[string]interface{}{
						"type":      "JSON_TASK",
						"text":      task.Content,
						"task_id":   task.ID,
						"status":    task.Status,
						"priority":  task.Priority,
						"file":      jsonFile,
						"source":    "git_json",
						"completed": task.Completed,
					})
				}
			}

			continue
		}

		commits := strings.Split(strings.TrimSpace(string(commitOutput)), "\n")
		processedTasks := make(map[string]bool)

		for _, commit := range commits {
			if err := ctx.Err(); err != nil {
				return discoveries
			}

			commit = strings.TrimSpace(commit)
			if commit == "" {
				continue
			}

			cmd = exec.CommandContext(ctx, "git", "show", commit+":"+jsonFile)
			cmd.Dir = projectRoot
			fileContent, err := cmd.Output()
			if err != nil {
				continue
			}

			tasks, _, err := LoadJSONStateFromContent(fileContent)
			if err != nil {
				continue
			}

			for _, task := range tasks {
				uniqueKey := fmt.Sprintf("%s:%s", task.ID, commit)
				if processedTasks[uniqueKey] {
					continue
				}
				processedTasks[uniqueKey] = true

				discoveries = append(discoveries, map[string]interface{}{
					"type":      "JSON_TASK",
					"text":      task.Content,
					"task_id":   task.ID,
					"status":    task.Status,
					"priority":  task.Priority,
					"file":      jsonFile,
					"commit":    commit[:8],
					"source":    "git_json",
					"completed": task.Completed,
				})
			}
		}
	}

	return discoveries
}
