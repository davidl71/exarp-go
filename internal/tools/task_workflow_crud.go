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
	if v, ok := params["status"]; ok && cast.ToString(v) != "" {
		status = normalizeStatus(cast.ToString(v))
	}

	newStatus := models.StatusTodo
	if v, ok := params["new_status"]; ok && cast.ToString(v) != "" {
		newStatus = normalizeStatus(cast.ToString(v))
	}

	// Default false: include all matching tasks (including short/empty descriptions)
	clarificationNone := false
	if _, ok := params["clarification_none"]; ok {
		clarificationNone = cast.ToBool(params["clarification_none"])
	}

	var filterTag string
	if v, ok := params["filter_tag"]; ok {
		filterTag = cast.ToString(v)
	}
	if strings.TrimSpace(filterTag) != "" {
		if norm, ok := models.NormalizeTag(filterTag); ok {
			filterTag = norm
		} else {
			// Force an empty result set for invalid tag inputs rather than broadening the query.
			filterTag = "#__invalid_tag__"
		}
	}

	taskIDs := ParseTaskIDsFromParams(params)

	dryRun := false
	if _, ok := params["dry_run"]; ok {
		dryRun = cast.ToBool(params["dry_run"])
	}

	// Optional MCP Elicitation: confirm batch approve when confirm_via_elicitation is true.
	const elicitationTimeout = 15 * time.Second

	if cast.ToBool(params["confirm_via_elicitation"]) {
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
	return models.NormalizeTags(ParamStringSliceTrimmedCommaSeparated(params, "tags"))
}

// parseRemoveTagsFromParams extracts remove_tags from params (comma-separated string or array). Used by update.
func parseRemoveTagsFromParams(params map[string]interface{}) []string {
	return models.NormalizeTags(ParamStringSliceTrimmedCommaSeparated(params, "remove_tags"))
}

// parseRecommendedToolsFromParams extracts recommended_tools from params (comma-separated string or array of tool IDs). Returns nil if not provided.
func parseRecommendedToolsFromParams(params map[string]interface{}) []string {
	return ParamStringSliceTrimmedCommaSeparated(params, "recommended_tools")
}

// parseDependenciesFromParams extracts dependencies from params (comma-separated string or array). Returns nil if not provided.
func parseDependenciesFromParams(params map[string]interface{}) []string {
	return ParamTaskDependencyIDs(params, "dependencies")
}

// parseOwnershipFromParams extracts ownership metadata from params.
// Returns nil if no ownership fields are provided.
// Params: owned_files (string array or comma-separated), owned_globs, forbidden_files, ownership_confidence, lane.
func parseOwnershipFromParams(params map[string]interface{}) *models.TaskOwnership {
	ownedFiles := parseStringSliceFromParams(params, "owned_files")
	ownedGlobs := parseStringSliceFromParams(params, "owned_globs")
	forbiddenFiles := parseStringSliceFromParams(params, "forbidden_files")
	confidence := cast.ToString(params["ownership_confidence"])
	lane := cast.ToString(params["lane"])

	if len(ownedFiles) == 0 && len(ownedGlobs) == 0 && len(forbiddenFiles) == 0 && confidence == "" && lane == "" {
		return nil
	}

	return &models.TaskOwnership{
		OwnedFiles:          ownedFiles,
		OwnedGlobs:          ownedGlobs,
		ForbiddenFiles:      forbiddenFiles,
		OwnershipConfidence: confidence,
		Lane:                lane,
	}
}

// parseStringSliceFromParams extracts a string slice from params (array or comma-separated string).
func parseStringSliceFromParams(params map[string]interface{}, key string) []string {
	return ParamStringSliceTrimmedCommaSeparated(params, key)
}

// handleTaskWorkflowUpdate updates task(s) by ID with optional new_status, priority, tags (merge), remove_tags, name, long_description, dependencies, local_ai_backend, project_id, or recommended_tools.
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

	newStatus, err := ParamEnum(params, "new_status",
		[]string{"Todo", "In Progress", "Review", "Done", "Blocked", "Cancelled"},
		"")
	if err != nil {
		return nil, fmt.Errorf("update: %w", err)
	}
	if newStatus != "" {
		newStatus = normalizeStatus(newStatus)
	}

	priority, err := ParamEnum(params, "priority",
		[]string{"low", "medium", "high", "critical"},
		"")
	if err != nil {
		return nil, fmt.Errorf("update: %w", err)
	}
	if priority != "" {
		priority = normalizePriority(priority)
	}

	priorityRankVal, hasPriorityRank := ParamIntOK(params, "priority_rank")

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
	ownership := parseOwnershipFromParams(params)
	hasOwnership := ownership != nil

	projectIDParam := strings.TrimSpace(cast.ToString(params["project_id"]))
	hasProjectID := projectIDParam != ""

	if newStatus == "" && priority == "" && !hasPriorityRank && len(addTags) == 0 && len(removeTags) == 0 && name == "" && longDescription == "" && parentID == "" && dependencies == nil && !hasLocalAIBackend && !hasRecommendedTools && !hasOwnership && !hasProjectID {
		return nil, fmt.Errorf("update action requires at least one of new_status, priority, priority_rank, tags, remove_tags, name, long_description, parent_id, project_id, dependencies, local_ai_backend, recommended_tools, or ownership fields (owned_files, lane, etc.)")
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

	canUseBatchUpdate := !useClaim &&
		!hasPriorityRank &&
		len(addTags) == 0 &&
		len(removeTags) == 0 &&
		name == "" &&
		longDescription == "" &&
		parentID == "" &&
		dependencies == nil &&
		!hasLocalAIBackend &&
		!hasRecommendedTools &&
		!hasOwnership &&
		!hasProjectID

	updatedIDs := []string{}
	updatedCount := 0

	var skippedLocked []string

	if canUseBatchUpdate && (newStatus != "" || priority != "") {
		batchUpdates := make([]database.TaskStatusUpdate, 0, len(taskIDs))
		for _, id := range taskIDs {
			batchUpdates = append(batchUpdates, database.TaskStatusUpdate{
				TaskID:   id,
				Status:   newStatus,
				Priority: priority,
			})
		}

		count, err := database.BatchUpdateTaskStatus(ctx, batchUpdates)
		if err == nil {
			updatedCount = count
			for _, u := range batchUpdates {
				updatedIDs = append(updatedIDs, u.TaskID)
			}
		}
	} else {
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

			if hasPriorityRank {
				task.PriorityRank = priorityRankVal
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
				task.Name = name
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

			if hasLocalAIBackend {
				backend := strings.TrimSpace(strings.ToLower(localAIBackend))
				if backend == "mlx" {
					backend = ""
				}
				if backend == "fm" || backend == "ollama" {
					if task.Metadata == nil {
						task.Metadata = make(map[string]interface{})
					}

					task.Metadata[MetadataKeyPreferredBackend] = backend
				}
			}

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

			if hasOwnership {
				models.SetTaskOwnership(task, ownership)
			}

			if hasProjectID {
				task.ProjectID = projectIDParam
			}

			if err := store.UpdateTask(ctx, task); err != nil {
				continue
			}

			updatedIDs = append(updatedIDs, id)
			updatedCount++
		}
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

	// Suggest follow-up tasks when task is completed (Done)
	if newStatus == models.StatusDone && updatedCount > 0 {
		for _, id := range updatedIDs {
			task, err := store.GetTask(ctx, id)
			if err != nil || task == nil {
				continue
			}

			// Try to suggest follow-ups using LLM
			suggestions, sugErr := SuggestFollowUps(ctx, task)
			if sugErr == nil && len(suggestions) > 0 {
				result["followup_suggestions"] = suggestions
				result["followup_instructions"] = "Use task_workflow create action to create follow-up tasks from suggestions"
				break // Only suggest for one task
			}
			break // Only check first task
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

	// Apply filters
	var status, priority, filterTag, taskID string

	var limit int

	if v, ok := params["status"]; ok {
		status = cast.ToString(v)
	}

	if v, ok := params["priority"]; ok {
		priority = cast.ToString(v)
	}

	if v, ok := params["filter_tag"]; ok {
		filterTag = cast.ToString(v)
	}
	if strings.TrimSpace(filterTag) != "" {
		if norm, ok := models.NormalizeTag(filterTag); ok {
			filterTag = norm
		} else {
			filterTag = "#__invalid_tag__"
		}
	}

	filterName := strings.TrimSpace(cast.ToString(params["name"]))
	if filterName == "" {
		filterName = strings.TrimSpace(cast.ToString(params["name_contains"]))
	}
	if filterName == "" {
		filterName = strings.TrimSpace(cast.ToString(params["filter_name"]))
	}

	if v, ok := params["task_id"]; ok {
		taskID = cast.ToString(v)
	}

	if l, ok := params["limit"]; ok {
		limit = cast.ToInt(l)
	}

	// Default to open tasks only (Todo + In Progress) when no status filter is given.
	// When querying a specific task_id (used by `task show`), include closed tasks too.
	openOnly := status == "" && taskID == ""
	if openOnly {
		status = "" // keep empty so we filter by open set below
	}

	showAll := strings.EqualFold(status, "all")

	order := cast.ToString(params["order"])
	wantsExecutionOrder := order == "execution" || order == "dependency"

	// Prefer pushing common filters down into the TaskStore/DB layer to reduce allocations
	// and row decoding work. Execution-order sorting currently requires the full task set.
	var list []*Todo2Task
	if taskID != "" {
		t, err := store.GetTask(ctx, taskID)
		if err != nil {
			return nil, err
		}
		list = []*Todo2Task{t}
	} else if wantsExecutionOrder {
		// Full task set is needed to compute dependency order.
		list, err = store.ListTasks(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to load tasks: %w", err)
		}
	} else {
		var filters database.TaskFilters
		includeMetadata := ParamBool(params, "include_metadata", false)
		filters.IncludeMetadata = &includeMetadata
		if openOnly {
			filters.Statuses = models.OpenStatuses()
		} else if status != "" && !showAll {
			filters.Status = &status
		}
		if priority != "" {
			filters.Priority = &priority
		}
		if filterTag != "" {
			filters.Tag = &filterTag
		}
		if filterName != "" {
			filters.NameContains = &filterName
		}

		list, err = store.ListTasks(ctx, &filters)
		if err != nil {
			return nil, fmt.Errorf("failed to load tasks: %w", err)
		}
	}

	// Filter tasks
	filtered := make([]*Todo2Task, 0, len(list))

	for _, task := range list {
		if task == nil {
			continue
		}
		if taskID != "" && task.ID != taskID {
			continue
		}

		if status != "" && !showAll && task.Status != status {
			continue
		}

		// When execution-order sorting is requested we load full task set above; apply filters here.
		if wantsExecutionOrder {
			if status != "" && !showAll && task.Status != status {
				continue
			}
			if openOnly && !models.IsOpenStatus(task.Status) {
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
			if filterName != "" && !taskMatchesNameFilter(task, filterName) {
				continue
			}
		}

		// Filter by owned file (check ownership metadata)
		if ownedFile := cast.ToString(params["owned_file"]); ownedFile != "" {
			ownedFiles := models.GetOwnedFiles(task)
			found := false
			for _, f := range ownedFiles {
				if f == ownedFile || strings.HasSuffix(f, ownedFile) {
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
	if wantsExecutionOrder {
		// BacklogExecutionOrder currently operates on []Todo2Task (values), so only
		// materialize the slice when ordering is requested.
		tasks := make([]Todo2Task, 0, len(list))
		for _, t := range list {
			if t != nil {
				tasks = append(tasks, *t)
			}
		}

		orderedIDs, _, _, err := BacklogExecutionOrder(tasks, nil)
		if err == nil {
			filteredMap := make(map[string]*Todo2Task, len(filtered))
			for _, t := range filtered {
				filteredMap[t.ID] = t
			}

			orderedSet := make(map[string]bool)
			for _, id := range orderedIDs {
				orderedSet[id] = true
			}

			orderedFiltered := make([]*Todo2Task, 0, len(filtered))

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
		outputFormat = "json"
	}

	if outputFormat == "json" {
		includeMetadata := ParamBool(params, "include_metadata", false)
		includeFullLongDescription := ParamBool(params, "include_full_long_description", false)
		includeLocks := ParamBool(params, "include_locks", false)

		taskIDs := make([]string, 0, len(filtered))
		for i := range filtered {
			taskIDs = append(taskIDs, filtered[i].ID)
		}
		var activeLocks map[string]database.LockStatus
		if includeLocks && len(taskIDs) > 0 {
			activeLocks, _ = database.GetActiveLockMapForTasks(ctx, taskIDs)
		}
		taskMaps := make([]map[string]interface{}, len(filtered))
		for i := range filtered {
			t := filtered[i]
			m := map[string]interface{}{"id": t.ID, "content": t.Content, "status": t.Status}
			if t.Priority != "" {
				m["priority"] = t.Priority
			}
			if len(t.Tags) > 0 {
				// Clone to avoid leaking references to internal slices.
				m["tags"] = append([]string(nil), t.Tags...)
			}
			if t.LongDescription != "" {
				ld := t.LongDescription
				// When listing many tasks, truncate long_description to keep responses compact.
				// When querying a specific task_id (e.g. CLI `task show`), return the full long_description.
				if !includeFullLongDescription && taskID == "" && len(ld) > 120 {
					ld = ld[:117] + "..."
				}
				m["long_description"] = ld
			}
			if len(t.Dependencies) > 0 {
				// Clone to avoid leaking references to internal slices.
				m["dependencies"] = append([]string(nil), t.Dependencies...)
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
			if includeMetadata && len(t.Metadata) > 0 {
				m["metadata"] = t.Metadata
			}
			if rt := GetRecommendedTools(t.Metadata); len(rt) > 0 {
				m["recommended_tools"] = rt
			}
			// Include ownership metadata if present
			if own := models.GetTaskOwnership(t); own != nil {
				ownershipMap := make(map[string]interface{})
				if len(own.OwnedFiles) > 0 {
					ownershipMap["owned_files"] = append([]string(nil), own.OwnedFiles...)
				}
				if len(own.OwnedGlobs) > 0 {
					ownershipMap["owned_globs"] = append([]string(nil), own.OwnedGlobs...)
				}
				if len(own.ForbiddenFiles) > 0 {
					ownershipMap["forbidden_files"] = append([]string(nil), own.ForbiddenFiles...)
				}
				if own.OwnershipConfidence != "" {
					ownershipMap["ownership_confidence"] = own.OwnershipConfidence
				}
				if own.Lane != "" {
					ownershipMap["lane"] = own.Lane
				}
				if len(ownershipMap) > 0 {
					m["ownership"] = ownershipMap
				}
			}
			if includeLocks {
				if lock, ok := activeLocks[t.ID]; ok {
					m["active_claim"] = lockToMap(lock)
				}
			}
			if taskID != "" {
				if runs, err := database.ListTaskExecutionRuns(ctx, t.ID, "", 5); err == nil && len(runs) > 0 {
					items := make([]map[string]interface{}, 0, len(runs))
					for j := range runs {
						items = append(items, runToMap(&runs[j]))
					}
					m["recent_runs"] = items
				}
				if verifications, err := database.ListTaskVerifications(ctx, t.ID, "", 5); err == nil && len(verifications) > 0 {
					m["recent_verifications"] = verificationListToMaps(verifications)
				}
				if progressEntries, err := database.ListTaskProgressEntries(ctx, t.ID, "", 5); err == nil && len(progressEntries) > 0 {
					m["recent_progress"] = progressListToMaps(progressEntries)
				}
			}

			taskMaps[i] = m
		}
		out := map[string]interface{}{"success": true, "method": "list", "tasks": taskMaps}
		AddTokenEstimateToResult(out)
		// Default compact=true for MCP callers to reduce token overhead; pass compact=false to opt out
		compact := ParamBool(params, "compact", true)
		return FormatResultOptionalCompact(out, "", compact)
	}
	// Text format: column widths aligned with TUI (internal/cli/tui.go colIDMedium, colStatus, colPriority)
	const colID = 22

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

	sb.WriteString(fmt.Sprintf("Tasks (%d total, %d shown)\n", len(list), len(filtered)))

	sepLen := colID + colStatus + colPriority + colContent + 3*3 // 3 " | " separators
	if sepLen < 80 {
		sepLen = 80
	}

	sb.WriteString(strings.Repeat("=", sepLen) + "\n")
	sb.WriteString(fmt.Sprintf("%-*s | %-*s | %-*s | %s\n", colID, "ID", colStatus, "Status", colPriority, "Priority", "Content"))
	sb.WriteString(strings.Repeat("-", sepLen) + "\n")

	for _, task := range filtered {
		id := pad(task.ID, colID)
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

	// When showing a single task (e.g. task show), append recommended_tools and ownership lines
	if len(filtered) == 1 {
		if rt := GetRecommendedTools(filtered[0].Metadata); len(rt) > 0 {
			sb.WriteString("\nRecommended tools: " + strings.Join(rt, ", ") + "\n")
		}
		if own := models.GetTaskOwnership(filtered[0]); own != nil {
			sb.WriteString("\nOwnership:\n")
			if own.Lane != "" {
				sb.WriteString(fmt.Sprintf("  Lane: %s\n", own.Lane))
			}
			if own.OwnershipConfidence != "" {
				sb.WriteString(fmt.Sprintf("  Confidence: %s\n", own.OwnershipConfidence))
			}
			if len(own.OwnedFiles) > 0 {
				sb.WriteString("  Owned files:\n")
				for _, f := range own.OwnedFiles {
					sb.WriteString(fmt.Sprintf("    - %s\n", f))
				}
			}
			if len(own.OwnedGlobs) > 0 {
				sb.WriteString("  Owned globs:\n")
				for _, g := range own.OwnedGlobs {
					sb.WriteString(fmt.Sprintf("    - %s\n", g))
				}
			}
			if len(own.ForbiddenFiles) > 0 {
				sb.WriteString("  Forbidden files:\n")
				for _, f := range own.ForbiddenFiles {
					sb.WriteString(fmt.Sprintf("    - %s\n", f))
				}
			}
		}
	}

	return []framework.TextContent{
		{Type: "text", Text: sb.String()},
	}, nil
}

// handleTaskWorkflowShow is a Cursor-friendly alias for fetching one task by ID.
// It delegates to list with task_id set, and forces full long_description output.
func handleTaskWorkflowShow(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	taskID := strings.TrimSpace(cast.ToString(params["task_id"]))
	if taskID == "" {
		taskID = strings.TrimSpace(cast.ToString(params["id"]))
	}
	if taskID == "" {
		return nil, fmt.Errorf("show action requires task_id (or id)")
	}

	// Clone params to avoid surprising the caller by mutating their map.
	next := make(map[string]interface{}, len(params)+2)
	for k, v := range params {
		next[k] = v
	}
	next["action"] = "list"
	next["task_id"] = taskID

	// For a single task, show full detail by default.
	if _, ok := next["output_format"]; !ok {
		next["output_format"] = "json"
	}
	if _, ok := next["include_full_long_description"]; !ok {
		next["include_full_long_description"] = true
	}
	if _, ok := next["include_metadata"]; !ok {
		next["include_metadata"] = true
	}

	return handleTaskWorkflowList(ctx, next)
}

// handleTaskWorkflowSync handles sync action for synchronizing tasks between SQLite and JSON.
// The "external" param (sync with external sources, e.g. infer_task_progress) is a future nice-to-have; if passed, it is ignored and SQLite↔JSON sync is performed.

// taskMatchesNameFilter returns true if needle is empty or task title/content contains needle (case-insensitive).
// Used when list loads the full task set (e.g. execution order); keep in sync with database.NameContains semantics.
func taskMatchesNameFilter(task *Todo2Task, needle string) bool {
	if needle == "" || task == nil {
		return true
	}

	n := strings.ToLower(needle)

	return strings.Contains(strings.ToLower(task.Name), n) || strings.Contains(strings.ToLower(task.Content), n)
}
