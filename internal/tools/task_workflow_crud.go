// task_workflow_crud.go — task_workflow crud handlers.
package tools

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/models"
	"github.com/davidl71/exarp-go/proto"
	"github.com/spf13/cast"
)

func handleTaskWorkflowApprove(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	// Extract parameters
	status := models.StatusReview
	if s, ok := params["status"].(string); ok && s != "" {
		status = normalizeStatus(s)
	}

	newStatus := models.StatusTodo
	if s, ok := params["new_status"].(string); ok && s != "" {
		newStatus = normalizeStatus(s)
	}

	// Default false: include all matching tasks (including short/empty descriptions)
	clarificationNone := false
	if cn, ok := params["clarification_none"].(bool); ok {
		clarificationNone = cn
	}

	var filterTag string
	if tag, ok := params["filter_tag"].(string); ok {
		filterTag = tag
	}

	taskIDs := ParseTaskIDsFromParams(params)

	dryRun := false
	if dr, ok := params["dry_run"].(bool); ok {
		dryRun = dr
	}

	// Optional MCP Elicitation: confirm batch approve when confirm_via_elicitation is true.
	// Use a timeout so elicitation never blocks indefinitely.
	const elicitationTimeout = 15 * time.Second

	if confirm, _ := params["confirm_via_elicitation"].(bool); confirm {
		if eliciter := framework.EliciterFromContext(ctx); eliciter != nil {
			schema := map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"proceed": map[string]interface{}{"type": "boolean", "description": "Proceed with batch approve?"},
					"dry_run": map[string]interface{}{"type": "boolean", "description": "Preview only (no updates)"},
				},
			}

			elicitCtx, cancel := context.WithTimeout(ctx, elicitationTimeout)
			defer cancel()

			action, content, err := eliciter.ElicitForm(elicitCtx, "Proceed with batch approve? You can choose dry run to preview only.", schema)
			if err != nil || action != "accept" {
				msg := "Batch approve cancelled by user or elicitation unavailable"
				if err != nil && (errors.Is(err, context.DeadlineExceeded) || (elicitCtx.Err() != nil && errors.Is(elicitCtx.Err(), context.DeadlineExceeded))) {
					msg = "Batch approve cancelled: elicitation timed out"
				}

				return framework.FormatResult(TaskWorkflowResponseToMap(&proto.TaskWorkflowResponse{Success: false, Cancelled: true, Message: msg}), "")
			}

			if content != nil {
				if proceed, ok := content["proceed"].(bool); ok && !proceed {
					return framework.FormatResult(TaskWorkflowResponseToMap(&proto.TaskWorkflowResponse{Success: false, Cancelled: true, Message: "Batch approve cancelled by user"}), "")
				}

				if dr, ok := content["dry_run"].(bool); ok && dr {
					dryRun = true
				}
			}
		}
	}

	// Use TaskStore (DB or file fallback) for filtering and updates
	store, err := getTaskStore(ctx)
	if err != nil {
		return handleTaskWorkflowApproveMCP(ctx, params)
	}

	filters := &database.TaskFilters{Status: &status}
	if filterTag != "" {
		filters.Tag = &filterTag
	}

	allTasks, err := store.ListTasks(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	// Filter candidates
	candidates := []*models.Todo2Task{}

	for _, task := range allTasks {
		if len(taskIDs) > 0 {
			found := false

			for _, id := range taskIDs {
				if task.ID == id {
					found = true
					break
				}
			}

			if !found {
				continue
			}
		}

		if clarificationNone {
			minDescLen := config.TaskMinDescriptionLength()

			needsClarification := task.LongDescription == "" || len(task.LongDescription) < minDescLen
			if needsClarification {
				continue
			}
		}

		candidates = append(candidates, task)
	}

	if dryRun {
		summaries := make([]*proto.TaskSummary, len(candidates))
		taskIDList := make([]string, len(candidates))

		for i, task := range candidates {
			summaries[i] = taskToTaskSummary(task)
			taskIDList[i] = task.ID
		}

		resp := &proto.TaskWorkflowResponse{
			Success:       true,
			Method:        "store",
			DryRun:        true,
			ApprovedCount: int32(len(candidates)),
			TaskIds:       taskIDList,
			Tasks:         summaries,
		}

		return framework.FormatResult(TaskWorkflowResponseToMap(resp), "")
	}

	// Update tasks via store (handles DB and file; sync is internal)
	approvedIDs := []string{}
	updatedCount := 0

	for _, task := range candidates {
		task.Status = newStatus
		if err := store.UpdateTask(ctx, task); err == nil {
			approvedIDs = append(approvedIDs, task.ID)
			updatedCount++
		}
	}

	resp := &proto.TaskWorkflowResponse{
		Success:       true,
		Method:        "store",
		ApprovedCount: int32(updatedCount),
		TaskIds:       approvedIDs,
	}

	return framework.FormatResult(TaskWorkflowResponseToMap(resp), "")
}

// parseTagsFromParams extracts tags from params (comma-separated string or array). Used by create and update.
func parseTagsFromParams(params map[string]interface{}) []string {
	var tags []string

	if t, ok := params["tags"].([]interface{}); ok {
		for _, tag := range t {
			if tagStr, ok := tag.(string); ok && tagStr != "" {
				tags = append(tags, strings.TrimSpace(tagStr))
			}
		}
	} else if tStr, ok := params["tags"].(string); ok && tStr != "" {
		for _, tag := range strings.Split(tStr, ",") {
			if trimmed := strings.TrimSpace(tag); trimmed != "" {
				tags = append(tags, trimmed)
			}
		}
	}

	return tags
}

// parseRemoveTagsFromParams extracts remove_tags from params (comma-separated string or array). Used by update.
func parseRemoveTagsFromParams(params map[string]interface{}) []string {
	var tags []string

	if t, ok := params["remove_tags"].([]interface{}); ok {
		for _, tag := range t {
			if tagStr, ok := tag.(string); ok && tagStr != "" {
				tags = append(tags, strings.TrimSpace(tagStr))
			}
		}
	} else if tStr, ok := params["remove_tags"].(string); ok && tStr != "" {
		for _, tag := range strings.Split(tStr, ",") {
			if trimmed := strings.TrimSpace(tag); trimmed != "" {
				tags = append(tags, trimmed)
			}
		}
	}

	return tags
}

// parseRecommendedToolsFromParams extracts recommended_tools from params (comma-separated string or array of tool IDs). Returns nil if not provided.
func parseRecommendedToolsFromParams(params map[string]interface{}) []string {
	var tools []string
	if t, ok := params["recommended_tools"].([]interface{}); ok {
		for _, v := range t {
			if s, ok := v.(string); ok && s != "" {
				tools = append(tools, strings.TrimSpace(s))
			}
		}
	} else if tStr, ok := params["recommended_tools"].(string); ok && tStr != "" {
		for _, s := range strings.Split(tStr, ",") {
			if trimmed := strings.TrimSpace(s); trimmed != "" {
				tools = append(tools, trimmed)
			}
		}
	}
	return tools
}

// parseDependenciesFromParams extracts dependencies from params (comma-separated string or array). Returns nil if not provided.
func parseDependenciesFromParams(params map[string]interface{}) []string {
	if d, ok := params["dependencies"].([]interface{}); ok {
		var deps []string

		for _, dep := range d {
			if depStr, ok := dep.(string); ok && depStr != "" {
				deps = append(deps, strings.TrimSpace(depStr))
			}
		}

		return deps
	}

	if dStr, ok := params["dependencies"].(string); ok && dStr != "" {
		var deps []string

		for _, dep := range strings.Split(dStr, ",") {
			if trimmed := strings.TrimSpace(dep); trimmed != "" {
				deps = append(deps, trimmed)
			}
		}

		return deps
	}

	return nil
}

// handleTaskWorkflowUpdate updates task(s) by ID with optional new_status, priority, tags (merge), remove_tags, name, long_description, dependencies, or local_ai_backend.
// Uses TaskStore (DB or file); when moving to In Progress uses database.ClaimTaskForAgent for locking.
// Params: task_ids (required), new_status (optional), priority (optional), tags (optional; merged), remove_tags (optional), name (optional), long_description (optional), dependencies (optional; replaces), local_ai_backend (optional), recommended_tools (optional; MCP tool IDs).
func handleTaskWorkflowUpdate(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	taskIDs := ParseTaskIDsFromParams(params)
	if len(taskIDs) == 0 {
		return nil, fmt.Errorf("update action requires task_ids")
	}

	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task store: %w", err)
	}

	newStatus, _ := params["new_status"].(string)
	if newStatus != "" {
		newStatus = normalizeStatus(newStatus)
	}

	priority, _ := params["priority"].(string)
	if priority != "" {
		priority = normalizePriority(priority)
	}

	addTags := parseTagsFromParams(params)
	removeTags := parseRemoveTagsFromParams(params)
	name := cast.ToString(params["name"])
	longDescription := cast.ToString(params["long_description"])
	parentID := cast.ToString(params["parent_id"])
	dependencies := parseDependenciesFromParams(params)
	localAIBackend := cast.ToString(params["local_ai_backend"])
	hasLocalAIBackend := strings.TrimSpace(localAIBackend) != ""
	recommendedTools := parseRecommendedToolsFromParams(params)
	hasRecommendedTools := len(recommendedTools) > 0

	if newStatus == "" && priority == "" && len(addTags) == 0 && len(removeTags) == 0 && name == "" && longDescription == "" && parentID == "" && dependencies == nil && !hasLocalAIBackend && !hasRecommendedTools {
		return nil, fmt.Errorf("update action requires at least one of new_status, priority, tags, remove_tags, name, long_description, parent_id, dependencies, local_ai_backend, or recommended_tools")
	}

	useClaim := newStatus == models.StatusInProgress

	var agentID string

	if useClaim {
		if id, err := database.GetAgentID(); err == nil {
			agentID = id
		} else {
			useClaim = false
		}
	}

	updatedIDs := []string{}
	updatedCount := 0

	var skippedLocked []string

	for _, id := range taskIDs {
		var task *models.Todo2Task

		if useClaim && agentID != "" {
			leaseDuration := config.TaskLockLease()

			claimResult, err := database.ClaimTaskForAgent(ctx, id, agentID, leaseDuration)
			if err != nil {
				continue
			}

			if !claimResult.Success {
				if claimResult.WasLocked {
					skippedLocked = append(skippedLocked, id)
				}

				continue
			}

			task = claimResult.Task
		} else {
			var err error

			task, err = store.GetTask(ctx, id)
			if err != nil {
				continue
			}

			if newStatus != "" {
				task.Status = newStatus
			}
		}

		if priority != "" {
			task.Priority = priority
		}

		if len(removeTags) > 0 {
			removeSet := make(map[string]bool)
			for _, t := range removeTags {
				removeSet[t] = true
			}

			filtered := task.Tags[:0]

			for _, t := range task.Tags {
				if !removeSet[t] {
					filtered = append(filtered, t)
				}
			}

			task.Tags = filtered
		}

		if len(addTags) > 0 {
			existing := make(map[string]bool)
			for _, t := range task.Tags {
				existing[t] = true
			}

			for _, t := range addTags {
				if !existing[t] {
					task.Tags = append(task.Tags, t)
					existing[t] = true
				}
			}
		}

		if name != "" {
			task.Content = name
		}

		if longDescription != "" {
			task.LongDescription = longDescription
		}

		if !useClaim && newStatus != "" {
			task.Status = newStatus
		}

		if parentID != "" {
			task.ParentID = parentID
		}

		if dependencies != nil {
			task.Dependencies = dependencies
		}

		// Update preferred local AI backend if provided
		if backend, ok := params["local_ai_backend"].(string); ok && backend != "" {
			backend = strings.TrimSpace(strings.ToLower(backend))
			if backend == "fm" || backend == "mlx" || backend == "ollama" {
				if task.Metadata == nil {
					task.Metadata = make(map[string]interface{})
				}

				task.Metadata[MetadataKeyPreferredBackend] = backend
			}
		}

		// Update recommended_tools if provided
		if hasRecommendedTools {
			if task.Metadata == nil {
				task.Metadata = make(map[string]interface{})
			}
			slice := make([]interface{}, len(recommendedTools))
			for i, t := range recommendedTools {
				slice[i] = t
			}
			task.Metadata[MetadataKeyRecommendedTools] = slice
		}

		if err := store.UpdateTask(ctx, task); err != nil {
			continue
		}

		updatedIDs = append(updatedIDs, id)
		updatedCount++
	}

	result := map[string]interface{}{
		"success":       true,
		"method":        "store",
		"updated_count": updatedCount,
		"task_ids":      updatedIDs,
	}
	if len(skippedLocked) > 0 {
		result["skipped_locked"] = skippedLocked
	}

	if newStatus == models.StatusReview && updatedCount > 0 {
		approvalRequests := make([]ApprovalRequest, 0, len(updatedIDs))

		for _, id := range updatedIDs {
			task, err := store.GetTask(ctx, id)
			if err != nil || task == nil {
				continue
			}

			approvalRequests = append(approvalRequests, BuildApprovalRequestFromTask(task, ""))
		}

		if len(approvalRequests) > 0 {
			result["approval_requests"] = approvalRequests
			result["goto_human_instructions"] = "Call @gotoHuman request-human-review-with-form with each approval_request (form_id, field_data). Set GOTOHUMAN_API_KEY if needed. See docs/GOTOHUMAN_API_REFERENCE.md."
		}
	}

	return framework.FormatResult(result, "")
}

// handleTaskWorkflowApproveMCP returns an error when project root or task load fails (no bridge).
func handleTaskWorkflowApproveMCP(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	return nil, fmt.Errorf("approve action: project root or task load failed; cannot approve tasks")
}

// handleTaskWorkflowList handles list sub-action for displaying tasks.
func handleTaskWorkflowList(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task store: %w", err)
	}

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	tasks := make([]Todo2Task, len(list))
	for i, t := range list {
		tasks[i] = *t
	}

	// Apply filters
	var status, priority, filterTag, taskID string

	var limit int

	if s, ok := params["status"].(string); ok {
		status = s
	}

	if p, ok := params["priority"].(string); ok {
		priority = p
	}

	if tag, ok := params["filter_tag"].(string); ok {
		filterTag = tag
	}

	if tid, ok := params["task_id"].(string); ok {
		taskID = tid
	}

	if l, ok := params["limit"].(float64); ok {
		limit = int(l)
	} else if l, ok := params["limit"].(int); ok {
		limit = l
	}

	// Default to open tasks only (Todo + In Progress) when no status filter is given
	openOnly := status == ""
	if openOnly {
		status = "" // keep empty so we filter by open set below
	}

	showAll := strings.EqualFold(status, "all")

	// Filter tasks
	filtered := []Todo2Task{}

	for _, task := range tasks {
		if taskID != "" && task.ID != taskID {
			continue
		}

		if status != "" && !showAll && task.Status != status {
			continue
		}

		if openOnly && task.Status != models.StatusTodo && task.Status != models.StatusInProgress {
			continue
		}

		if priority != "" && task.Priority != priority {
			continue
		}

		if filterTag != "" {
			found := false

			for _, tag := range task.Tags {
				if tag == filterTag {
					found = true
					break
				}
			}

			if !found {
				continue
			}
		}

		filtered = append(filtered, task)
	}

	// Optional: sort by execution order (dependency order)
	if order := cast.ToString(params["order"]); order == "execution" || order == "dependency" {
		orderedIDs, _, _, err := BacklogExecutionOrder(tasks, nil)
		if err == nil {
			filteredMap := make(map[string]Todo2Task)
			for _, t := range filtered {
				filteredMap[t.ID] = t
			}

			orderedSet := make(map[string]bool)
			for _, id := range orderedIDs {
				orderedSet[id] = true
			}

			orderedFiltered := make([]Todo2Task, 0, len(filtered))

			for _, id := range orderedIDs {
				if t, ok := filteredMap[id]; ok {
					orderedFiltered = append(orderedFiltered, t)
				}
			}

			for _, t := range filtered {
				if !orderedSet[t.ID] {
					orderedFiltered = append(orderedFiltered, t)
				}
			}

			filtered = orderedFiltered
		}
	}

	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}

	// Format output
	outputFormat := cast.ToString(params["output_format"])
	if outputFormat == "" {
		outputFormat = "text"
	}

	if outputFormat == "json" {
		taskMaps := make([]map[string]interface{}, len(filtered))
		for i := range filtered {
			t := &filtered[i]
			m := map[string]interface{}{"id": t.ID, "content": t.Content, "status": t.Status}
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
			if t.ParentID != "" {
				m["parent_id"] = t.ParentID
			}
			if t.LastModified != "" {
				m["last_modified"] = t.LastModified
			}
			if t.CreatedAt != "" {
				m["created_at"] = t.CreatedAt
			}
			if t.CompletedAt != "" {
				m["completed_at"] = t.CompletedAt
			}
			if len(t.Metadata) > 0 {
				m["metadata"] = t.Metadata
			}
			if rt := GetRecommendedTools(t.Metadata); len(rt) > 0 {
				m["recommended_tools"] = rt
			}

			taskMaps[i] = m
		}
		out := map[string]interface{}{"success": true, "method": "list", "tasks": taskMaps}
		AddTokenEstimateToResult(out)
		compact := cast.ToBool(params["compact"])
		return FormatResultOptionalCompact(out, "", compact)
	}
	// Text format: column widths aligned with TUI (internal/cli/tui.go colIDMedium, colStatus, colPriority)
	const colID = 18

	const colStatus = 12

	const colPriority = 10

	const colContent = 50

	truncate := func(s string, w int) string {
		if len(s) <= w {
			return s
		}

		if w <= 3 {
			return s[:w]
		}

		return s[:w-3] + "..."
	}
	pad := func(s string, w int) string {
		if len(s) >= w {
			return truncate(s, w)
		}

		return s + strings.Repeat(" ", w-len(s))
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Tasks (%d total, %d shown)\n", len(tasks), len(filtered)))

	sepLen := colID + colStatus + colPriority + colContent + 3*3 // 3 " | " separators
	if sepLen < 80 {
		sepLen = 80
	}

	sb.WriteString(strings.Repeat("=", sepLen) + "\n")
	sb.WriteString(fmt.Sprintf("%-*s | %-*s | %-*s | %s\n", colID, "ID", colStatus, "Status", colPriority, "Priority", "Content"))
	sb.WriteString(strings.Repeat("-", sepLen) + "\n")

	for _, task := range filtered {
		id := pad(truncate(task.ID, colID), colID)
		status := pad(truncate(task.Status, colStatus), colStatus)
		priority := pad(truncate(task.Priority, colPriority), colPriority)

		content := truncate(task.Content, colContent)
		if content == "" {
			content = truncate(task.LongDescription, colContent)
		}

		if content == "" {
			content = "(no description)"
		}

		sb.WriteString(fmt.Sprintf("%-*s | %-*s | %-*s | %s\n", colID, id, colStatus, status, colPriority, priority, content))
	}

	// When showing a single task (e.g. task show), append recommended_tools line
	if len(filtered) == 1 {
		if rt := GetRecommendedTools(filtered[0].Metadata); len(rt) > 0 {
			sb.WriteString("\nRecommended tools: " + strings.Join(rt, ", ") + "\n")
		}
	}

	return []framework.TextContent{
		{Type: "text", Text: sb.String()},
	}, nil
}

// handleTaskWorkflowSync handles sync action for synchronizing tasks between SQLite and JSON.
// The "external" param (sync with external sources, e.g. infer_task_progress) is a future nice-to-have; if passed, it is ignored and SQLite↔JSON sync is performed.
