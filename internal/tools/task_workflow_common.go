// task_workflow_common.go — Task workflow shared infrastructure (store, response helpers, parse utils).
package tools

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/models"
	"github.com/davidl71/exarp-go/proto"
)

// ParseTaskIDsFromParams extracts task IDs from params. Supports task_id (single string) and
// task_ids (comma-separated string or JSON array, or []interface{}). Returns trimmed, deduplicated
// IDs in order of first occurrence. Empty slice when none specified.
func ParseTaskIDsFromParams(params map[string]interface{}) []string {
	seen := make(map[string]bool)

	var out []string

	add := func(id string) {
		id = strings.TrimSpace(id)
		if id != "" && !seen[id] {
			seen[id] = true

			out = append(out, id)
		}
	}
	if tid, ok := params["task_id"].(string); ok && tid != "" {
		add(tid)
	}

	if ids, ok := params["task_ids"].(string); ok && ids != "" {
		var parsed []string
		if json.Unmarshal([]byte(ids), &parsed) == nil {
			for _, id := range parsed {
				add(id)
			}
		} else {
			for _, id := range strings.Split(ids, ",") {
				add(id)
			}
		}
	} else if idsList, ok := params["task_ids"].([]interface{}); ok {
		for _, id := range idsList {
			if idStr, ok := id.(string); ok {
				add(idStr)
			}
		}
	}

	return out
}

// taskStoreKey is the context key for TaskStore (injected by handleTaskWorkflowNative).
type taskStoreKeyType struct{}

var taskStoreKey = &taskStoreKeyType{}

// getTaskStore returns the TaskStore from context, or builds one from FindProjectRoot() if not set.
// Use this in handlers to access task storage; allows tests to inject database.NewMockTaskStore().
func getTaskStore(ctx context.Context) (database.TaskStore, error) {
	if s, ok := ctx.Value(taskStoreKey).(database.TaskStore); ok && s != nil {
		return s, nil
	}

	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, err
	}

	return NewDefaultTaskStore(projectRoot), nil
}

// TaskWorkflowResponseToMap converts a TaskWorkflowResponse proto to map for response.FormatResult (keeps JSON shape stable).
func TaskWorkflowResponseToMap(resp *proto.TaskWorkflowResponse) map[string]interface{} {
	if resp == nil {
		return nil
	}

	out := map[string]interface{}{
		"success": resp.Success,
		"method":  resp.Method,
	}
	if resp.Message != "" {
		out["message"] = resp.Message
	}

	if len(resp.TaskIds) > 0 {
		out["task_ids"] = resp.TaskIds
	}

	if resp.ApprovedCount != 0 {
		out["approved_count"] = int(resp.ApprovedCount)
	}

	if resp.DryRun {
		out["dry_run"] = true
	}

	if resp.Cancelled {
		out["cancelled"] = true
	}

	if len(resp.Tasks) > 0 {
		tasks := make([]map[string]interface{}, len(resp.Tasks))

		for i, t := range resp.Tasks {
			m := map[string]interface{}{"id": t.Id, "content": t.Content, "status": t.Status}
			if t.Priority != "" {
				m["priority"] = t.Priority
			}

			if len(t.Tags) > 0 {
				m["tags"] = t.Tags
			}

			if t.LongDescription != "" {
				m["long_description"] = t.LongDescription
			}

			if len(t.Dependencies) > 0 {
				m["dependencies"] = t.Dependencies
			}

			if len(t.RecommendedTools) > 0 {
				m["recommended_tools"] = t.RecommendedTools
			}

			tasks[i] = m
		}

		out["tasks"] = tasks
	}

	if resp.SyncResults != nil {
		sr := resp.SyncResults
		out["sync_results"] = map[string]interface{}{
			"validated_tasks": int(sr.ValidatedTasks),
			"issues_found":    int(sr.IssuesFound),
			"issues":          sr.Issues,
			"synced":          sr.Synced,
		}
	}

	return out
}

// taskToTaskSummary converts a Todo2Task to proto TaskSummary.
func taskToTaskSummary(t *models.Todo2Task) *proto.TaskSummary {
	if t == nil {
		return nil
	}

	return &proto.TaskSummary{
		Id:               t.ID,
		Content:          t.Content,
		Status:           t.Status,
		Priority:         t.Priority,
		Tags:             t.Tags,
		LongDescription:  t.LongDescription,
		Dependencies:     t.Dependencies,
		RecommendedTools: GetRecommendedTools(t.Metadata),
	}
}

// handleTaskWorkflowApprove handles approve action for batch approving tasks
// Uses database for efficient updates, falls back to file-based approach if needed
// This is platform-agnostic (doesn't require Apple FM).
