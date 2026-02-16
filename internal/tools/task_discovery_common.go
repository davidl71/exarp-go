package tools

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/davidl71/exarp-go/internal/database"
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
			Status:   "Todo",
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

	if len(createdTasks) > 0 {
		if err := SyncTodo2Tasks(projectRoot); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to sync after discovery: %v\n", err)
		}
	}

	return createdTasks
}
