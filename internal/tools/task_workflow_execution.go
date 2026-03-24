package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/spf13/cast"
)

func handleTaskWorkflowStartRun(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	taskID := strings.TrimSpace(ParamString(params, "task_id"))
	if taskID == "" {
		return nil, fmt.Errorf("start_run: task_id is required")
	}
	if _, err := database.GetTask(ctx, taskID); err != nil {
		return nil, fmt.Errorf("start_run: %w", err)
	}

	agentID := strings.TrimSpace(ParamString(params, "agent_id"))
	if agentID == "" {
		if id, err := database.GetAgentID(); err == nil {
			agentID = id
		}
	}
	host, _ := os.Hostname()
	run := &database.TaskExecutionRun{
		TaskID:  taskID,
		AgentID: agentID,
		Host:    host,
		Status:  "running",
		Summary: strings.TrimSpace(cast.ToString(params["summary"])),
		Notes:   strings.TrimSpace(cast.ToString(params["notes"])),
	}
	if err := database.StartTaskExecutionRun(ctx, run); err != nil {
		return nil, fmt.Errorf("start_run: %w", err)
	}
	return framework.FormatResult(map[string]interface{}{
		"success": true,
		"method":  "start_run",
		"run":     runToMap(run),
	}, "")
}

func handleTaskWorkflowEndRun(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	runID := strings.TrimSpace(ParamString(params, "run_id"))
	if runID == "" {
		return nil, fmt.Errorf("end_run: run_id is required")
	}
	status := strings.TrimSpace(cast.ToString(params["result"]))
	if status == "" {
		status = strings.TrimSpace(cast.ToString(params["status"]))
	}
	if status == "" {
		status = "completed"
	}
	if err := database.EndTaskExecutionRun(
		ctx,
		runID,
		status,
		strings.TrimSpace(cast.ToString(params["summary"])),
		parseStringListParam(params, "files_touched"),
		parseStringListParam(params, "commands_run"),
		strings.TrimSpace(cast.ToString(params["notes"])),
	); err != nil {
		return nil, fmt.Errorf("end_run: %w", err)
	}
	run, err := database.GetTaskExecutionRun(ctx, runID)
	if err != nil {
		return nil, fmt.Errorf("end_run: %w", err)
	}
	return framework.FormatResult(map[string]interface{}{
		"success": true,
		"method":  "end_run",
		"run":     runToMap(run),
	}, "")
}

func handleTaskWorkflowListRuns(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	limit := cast.ToInt(params["limit"])
	runs, err := database.ListTaskExecutionRuns(ctx, strings.TrimSpace(ParamString(params, "task_id")), strings.TrimSpace(cast.ToString(params["status"])), limit)
	if err != nil {
		return nil, fmt.Errorf("list_runs: %w", err)
	}
	items := make([]map[string]interface{}, 0, len(runs))
	for i := range runs {
		items = append(items, runToMap(&runs[i]))
	}
	return framework.FormatResult(map[string]interface{}{
		"success": true,
		"method":  "list_runs",
		"runs":    items,
	}, "")
}

func handleTaskWorkflowShowRun(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	runID := strings.TrimSpace(ParamString(params, "run_id"))
	if runID == "" {
		return nil, fmt.Errorf("show_run: run_id is required")
	}
	run, err := database.GetTaskExecutionRun(ctx, runID)
	if err != nil {
		return nil, fmt.Errorf("show_run: %w", err)
	}
	verifications, _ := database.ListTaskVerifications(ctx, run.TaskID, runID, 10)
	progressEntries, _ := database.ListTaskProgressEntries(ctx, run.TaskID, runID, 10)
	return framework.FormatResult(map[string]interface{}{
		"success":       true,
		"method":        "show_run",
		"run":           runToMap(run),
		"verifications": verificationListToMaps(verifications),
		"progress":      progressListToMaps(progressEntries),
	}, "")
}

func handleTaskWorkflowVerify(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	taskID := strings.TrimSpace(ParamString(params, "task_id"))
	if taskID == "" {
		return nil, fmt.Errorf("verify: task_id is required")
	}
	verification := &database.TaskVerification{
		TaskID:  taskID,
		RunID:   strings.TrimSpace(ParamString(params, "run_id")),
		Kind:    strings.TrimSpace(cast.ToString(params["kind"])),
		Command: strings.TrimSpace(cast.ToString(params["command"])),
		Result:  strings.TrimSpace(cast.ToString(params["result"])),
		Details: strings.TrimSpace(cast.ToString(params["details"])),
	}
	if err := database.AddTaskVerification(ctx, verification); err != nil {
		return nil, fmt.Errorf("verify: %w", err)
	}
	return framework.FormatResult(map[string]interface{}{
		"success":      true,
		"method":       "verify",
		"verification": verificationToMap(verification),
	}, "")
}

func handleTaskWorkflowAddProgress(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	taskID := strings.TrimSpace(ParamString(params, "task_id"))
	if taskID == "" {
		return nil, fmt.Errorf("add_progress: task_id is required")
	}
	entry := &database.TaskProgressEntry{
		TaskID:        taskID,
		RunID:         strings.TrimSpace(ParamString(params, "run_id")),
		Summary:       strings.TrimSpace(cast.ToString(params["summary"])),
		Files:         parseStringListParam(params, "files"),
		RemainingWork: strings.TrimSpace(cast.ToString(params["remaining_work"])),
	}
	if err := database.AddTaskProgressEntry(ctx, entry); err != nil {
		return nil, fmt.Errorf("add_progress: %w", err)
	}
	return framework.FormatResult(map[string]interface{}{
		"success":  true,
		"method":   "add_progress",
		"progress": progressToMap(entry),
	}, "")
}

func handleTaskWorkflowSplit(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	taskID := strings.TrimSpace(ParamString(params, "task_id"))
	if taskID == "" {
		return nil, fmt.Errorf("split: task_id is required")
	}
	parent, err := database.GetTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("split: %w", err)
	}
	childrenRaw := strings.TrimSpace(cast.ToString(params["children"]))
	if childrenRaw == "" {
		return nil, fmt.Errorf("split: children is required")
	}
	var children []map[string]interface{}
	if err := json.Unmarshal([]byte(childrenRaw), &children); err != nil {
		return nil, fmt.Errorf("split: children must be valid JSON: %w", err)
	}
	if len(children) == 0 {
		return nil, fmt.Errorf("split: children is empty")
	}

	dependencyMode := strings.ToLower(strings.TrimSpace(cast.ToString(params["dependency_mode"])))
	if dependencyMode == "" {
		dependencyMode = "parallel"
	}

	sharedPlanningDoc := ""
	if parent.Metadata != nil {
		if pd, ok := parent.Metadata["planning_doc"].(string); ok {
			sharedPlanningDoc = pd
		}
	}

	serialDependsOn := ""
	if dependencyMode == "serial" {
		serialDependsOn = taskID
	}
	for i := range children {
		children[i]["parent_id"] = taskID
		if _, ok := children[i]["tags"]; !ok && len(parent.Tags) > 0 {
			children[i]["tags"] = append([]string(nil), parent.Tags...)
		}
		if sharedPlanningDoc != "" {
			if _, ok := children[i]["planning_doc"]; !ok {
				children[i]["planning_doc"] = sharedPlanningDoc
			}
		}
		switch dependencyMode {
		case "parallel":
			if _, ok := children[i]["dependencies"]; !ok {
				children[i]["dependencies"] = []string{}
			}
		case "serial":
			children[i]["dependencies"] = mergeStringLists(asStringSlice(children[i]["dependencies"]), []string{serialDependsOn})
		default:
			return nil, fmt.Errorf("split: dependency_mode must be parallel or serial")
		}
	}

	createParams := map[string]interface{}{
		"tasks":         children,
		"parent_id":     taskID,
		"planning_doc":  sharedPlanningDoc,
		"auto_estimate": false,
	}
	result, err := handleTaskWorkflowCreate(ctx, createParams)
	if err != nil {
		return nil, fmt.Errorf("split: %w", err)
	}
	if dependencyMode == "serial" && len(result) > 0 {
		var payload map[string]interface{}
		if json.Unmarshal([]byte(result[0].Text), &payload) == nil {
			if ids := interfaceToStringSlice(payload["task_ids"]); len(ids) > 0 {
				for i := 1; i < len(ids); i++ {
					child, err := database.GetTask(ctx, ids[i])
					if err != nil {
						continue
					}
					child.Dependencies = mergeStringLists(child.Dependencies, []string{ids[i-1]})
					_ = database.UpdateTask(ctx, child)
				}
			}
		}
	}
	return result, nil
}

func parseStringListParam(params map[string]interface{}, key string) []string {
	raw, ok := params[key]
	if !ok || raw == nil {
		return nil
	}
	switch v := raw.(type) {
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
				out = append(out, strings.TrimSpace(s))
			}
		}
		return out
	case []string:
		return v
	case string:
		var out []string
		if json.Unmarshal([]byte(v), &out) == nil {
			return out
		}
		parts := strings.Split(v, ",")
		out = make([]string, 0, len(parts))
		for _, part := range parts {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				out = append(out, trimmed)
			}
		}
		return out
	default:
		return nil
	}
}

func runToMap(run *database.TaskExecutionRun) map[string]interface{} {
	if run == nil {
		return nil
	}
	m := map[string]interface{}{
		"run_id":     run.RunID,
		"task_id":    run.TaskID,
		"status":     run.Status,
		"started_at": run.StartedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if run.AgentID != "" {
		m["agent_id"] = run.AgentID
	}
	if run.Host != "" {
		m["host"] = run.Host
	}
	if run.Summary != "" {
		m["summary"] = run.Summary
	}
	if len(run.FilesTouched) > 0 {
		m["files_touched"] = run.FilesTouched
	}
	if len(run.CommandsRun) > 0 {
		m["commands_run"] = run.CommandsRun
	}
	if run.Notes != "" {
		m["notes"] = run.Notes
	}
	if !run.EndedAt.IsZero() {
		m["ended_at"] = run.EndedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	return m
}

func verificationToMap(v *database.TaskVerification) map[string]interface{} {
	if v == nil {
		return nil
	}
	m := map[string]interface{}{
		"verification_id": v.VerificationID,
		"task_id":         v.TaskID,
		"kind":            v.Kind,
		"result":          v.Result,
		"created_at":      v.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if v.RunID != "" {
		m["run_id"] = v.RunID
	}
	if v.Command != "" {
		m["command"] = v.Command
	}
	if v.Details != "" {
		m["details"] = v.Details
	}
	return m
}

func progressToMap(p *database.TaskProgressEntry) map[string]interface{} {
	if p == nil {
		return nil
	}
	m := map[string]interface{}{
		"progress_id": p.ProgressID,
		"task_id":     p.TaskID,
		"summary":     p.Summary,
		"created_at":  p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if p.RunID != "" {
		m["run_id"] = p.RunID
	}
	if len(p.Files) > 0 {
		m["files"] = p.Files
	}
	if p.RemainingWork != "" {
		m["remaining_work"] = p.RemainingWork
	}
	return m
}

func verificationListToMaps(items []database.TaskVerification) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(items))
	for i := range items {
		out = append(out, verificationToMap(&items[i]))
	}
	return out
}

func progressListToMaps(items []database.TaskProgressEntry) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(items))
	for i := range items {
		out = append(out, progressToMap(&items[i]))
	}
	return out
}

func asStringSlice(v interface{}) []string {
	switch x := v.(type) {
	case []string:
		return x
	case []interface{}:
		out := make([]string, 0, len(x))
		for _, item := range x {
			if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
				out = append(out, strings.TrimSpace(s))
			}
		}
		return out
	default:
		return nil
	}
}

func interfaceToStringSlice(v interface{}) []string {
	switch x := v.(type) {
	case []string:
		return x
	case []interface{}:
		return asStringSlice(x)
	default:
		return nil
	}
}

func mergeStringLists(existing, extra []string) []string {
	seen := make(map[string]bool, len(existing)+len(extra))
	out := make([]string, 0, len(existing)+len(extra))
	for _, item := range append(existing, extra...) {
		item = strings.TrimSpace(item)
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	return out
}

func lockToMap(lock database.LockStatus) map[string]interface{} {
	m := map[string]interface{}{
		"task_id":        lock.TaskID,
		"assignee":       lock.Assignee,
		"assigned_at":    lock.AssignedAt.Format("2006-01-02T15:04:05Z07:00"),
		"lock_until":     lock.LockUntil.Format("2006-01-02T15:04:05Z07:00"),
		"time_remaining": lock.TimeRemaining.Round(0).String(),
	}
	if lock.IsExpired {
		m["status"] = "expired"
	} else {
		m["status"] = "active"
	}
	return m
}
