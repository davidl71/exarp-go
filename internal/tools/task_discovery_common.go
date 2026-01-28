package tools

import (
	"fmt"
	"strings"
)

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
func createTasksFromDiscoveries(projectRoot string, discoveries []map[string]interface{}) []map[string]interface{} {
	createdTasks := []map[string]interface{}{}

	existingTasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return createdTasks
	}

	existingContent := make(map[string]bool)
	for _, task := range existingTasks {
		existingContent[strings.ToLower(strings.TrimSpace(task.Content))] = true
	}

	for _, discovery := range discoveries {
		text, ok := discovery["text"].(string)
		if !ok || text == "" {
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
		newTask := Todo2Task{
			ID:       taskID,
			Content:  text,
			Status:   "Todo",
			Priority: "medium",
			Tags:     []string{"discovered", sourceTag},
			Metadata: SanitizeMetadataForWrite(metadata),
		}

		existingTasks = append(existingTasks, newTask)
		existingContent[textLower] = true

		// Coerce source to string so tool response JSON is always valid
		createdTasks = append(createdTasks, map[string]interface{}{
			"id":      taskID,
			"content": text,
			"source":  toJSONSafeString(discovery["source"]),
		})
	}

	if len(createdTasks) > 0 {
		_ = SaveTodo2Tasks(projectRoot, existingTasks)
	}

	return createdTasks
}
