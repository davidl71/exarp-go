// task_workflow_create_ai.go — Task workflow: create, batch-create, single-create, enrich, fix-IDs, and estimate helpers.
// See also: task_workflow_ai_run.go
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/models"
	"github.com/spf13/cast"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ─── Contents ───────────────────────────────────────────────────────────────
//   handleTaskWorkflowCreate
//   handleTaskWorkflowCreateBatch — handleTaskWorkflowCreateBatch creates multiple tasks from a tasks JSON array.
//   handleTaskWorkflowCreateSingle — handleTaskWorkflowCreateSingle creates a single task from flat params.
//   handleTaskWorkflowEnrichToolHints — handleTaskWorkflowEnrichToolHints applies tag→tool rules to Todo and In Progress tasks:
//   handleTaskWorkflowFixInvalidIDs — handleTaskWorkflowFixInvalidIDs finds tasks with invalid IDs (e.g. T-NaN from JS), assigns new epoch IDs,
//   generateEpochTaskID — generateEpochTaskID generates a task ID using epoch nanoseconds (T-{epoch_nanoseconds}).
//   normalizePriority — normalizePriority normalizes priority values to canonical lowercase form.
//   addEstimateComment — addEstimateComment estimates task duration and adds it as a comment
//   formatEstimateComment — formatEstimateComment formats estimation result as a markdown comment.
// ────────────────────────────────────────────────────────────────────────────

// ─── handleTaskWorkflowCreate ───────────────────────────────────────────────
func handleTaskWorkflowCreate(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	// Check for batch mode: tasks param is a JSON array string or []interface{}
	if tasksParam, hasTasks := params["tasks"]; hasTasks {
		return handleTaskWorkflowCreateBatch(ctx, params, tasksParam)
	}

	return handleTaskWorkflowCreateSingle(ctx, params)
}

// ─── handleTaskWorkflowCreateBatch ──────────────────────────────────────────
// handleTaskWorkflowCreateBatch creates multiple tasks from a tasks JSON array.
func handleTaskWorkflowCreateBatch(ctx context.Context, params map[string]interface{}, tasksParam interface{}) ([]framework.TextContent, error) {
	var taskDefs []map[string]interface{}

	switch v := tasksParam.(type) {
	case string:
		if err := json.Unmarshal([]byte(v), &taskDefs); err != nil {
			return nil, fmt.Errorf("tasks param must be a valid JSON array: %w", err)
		}
	case []interface{}:
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				taskDefs = append(taskDefs, m)
			}
		}
	default:
		return nil, fmt.Errorf("tasks param must be a JSON array string or array of objects")
	}

	if len(taskDefs) == 0 {
		return nil, fmt.Errorf("tasks array is empty")
	}

	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task store: %w", err)
	}

	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load existing tasks for dependency validation: %w", err)
	}

	existingTasks := make([]Todo2Task, len(list))
	existingTaskMap := make(map[string]bool)
	for i, t := range list {
		existingTasks[i] = *t
		existingTaskMap[t.ID] = true
	}

	autoEstimate := cast.ToBool(params["auto_estimate"])
	if _, hasAutoEstimate := params["auto_estimate"]; !hasAutoEstimate {
		autoEstimate = true
	}

	var createdTasks []map[string]interface{}
	var createdIDs []string
	var errors []string

	for i, def := range taskDefs {
		merged := make(map[string]interface{})
		merged["auto_estimate"] = autoEstimate
		if sharedVal := cast.ToString(params["planning_doc"]); sharedVal != "" {
			merged["planning_doc"] = sharedVal
		}
		if sharedVal := cast.ToString(params["epic_id"]); sharedVal != "" {
			merged["epic_id"] = sharedVal
		}
		if sharedVal := cast.ToString(params["parent_id"]); sharedVal != "" {
			merged["parent_id"] = sharedVal
		}
		for k, v := range def {
			merged[k] = v
		}

		result, err := handleTaskWorkflowCreateSingleWithDeps(ctx, store, projectRoot, merged, existingTasks, existingTaskMap)
		if err != nil {
			errors = append(errors, fmt.Sprintf("task[%d] %q: %v", i, cast.ToString(def["name"]), err))
			continue
		}

		if len(result) > 0 {
			var data map[string]interface{}
			if json.Unmarshal([]byte(result[0].Text), &data) == nil {
				if taskData, ok := data["task"].(map[string]interface{}); ok {
					createdTasks = append(createdTasks, taskData)
					if id, ok := taskData["id"].(string); ok {
						createdIDs = append(createdIDs, id)
						existingTaskMap[id] = true
					}
				}
			}
		}
	}

	batchResult := map[string]interface{}{
		"success":       len(errors) == 0,
		"method":        "store",
		"created_count": len(createdTasks),
		"task_ids":      createdIDs,
		"tasks":         createdTasks,
	}
	if len(errors) > 0 {
		batchResult["errors"] = errors
		batchResult["success"] = len(createdTasks) > 0
	}

	outputPath := ParamOutputPath(params)
	return framework.FormatResult(batchResult, outputPath)
}

// handleTaskWorkflowCreateSingleWithDeps creates a single task with pre-loaded existing tasks.
// This avoids repeated ListTasks calls in batch mode.
func handleTaskWorkflowCreateSingleWithDeps(
	ctx context.Context,
	store database.TaskStore,
	projectRoot string,
	params map[string]interface{},
	existingTasks []Todo2Task,
	existingTaskMap map[string]bool,
) ([]framework.TextContent, error) {
	name := cast.ToString(params["name"])
	if name == "" {
		return nil, fmt.Errorf("name is required for task creation")
	}

	longDescription := cast.ToString(params["long_description"])
	if longDescription == "" {
		longDescription = name
	}

	status := config.DefaultTaskStatus()
	if rawStatus, err := ParamEnum(params, "status",
		[]string{"Todo", "In Progress", "Review", "Done", "Blocked", "Cancelled"},
		""); err != nil {
		return nil, fmt.Errorf("create: %w", err)
	} else if rawStatus != "" {
		status = normalizeStatus(rawStatus)
	}

	priority := config.DefaultTaskPriority()
	if rawPriority, err := ParamEnum(params, "priority",
		[]string{"low", "medium", "high", "critical"},
		""); err != nil {
		return nil, fmt.Errorf("create: %w", err)
	} else if rawPriority != "" {
		priority = normalizePriority(rawPriority)
	}

	tags := config.DefaultTaskTags()
	if len(tags) == 0 {
		tags = []string{}
	}
	if extra := ParamStringSliceTrimmedCommaSeparated(params, "tags"); len(extra) > 0 {
		tags = append(tags, extra...)
	}

	dependencies := ParamStringSliceTrimmedCommaSeparated(params, "dependencies")
	if dependencies == nil {
		dependencies = []string{}
	}

	nextID := generateEpochTaskID()

	for _, dep := range dependencies {
		if !existingTaskMap[dep] {
			return nil, fmt.Errorf("dependency %s does not exist", dep)
		}
	}

	var planningDoc string
	if pd := cast.ToString(params["planning_doc"]); pd != "" {
		planningDoc = pd
		if err := ValidatePlanningLink(projectRoot, planningDoc); err != nil {
			return nil, fmt.Errorf("invalid planning document link: %w", err)
		}
	}

	var epicID string
	if eid := cast.ToString(params["epic_id"]); eid != "" {
		epicID = eid
		if err := ValidateTaskReference(epicID, existingTasks); err != nil {
			return nil, fmt.Errorf("invalid epic ID: %w", err)
		}
	}

	parentID := cast.ToString(params["parent_id"])
	if parentID != "" {
		if err := ValidateTaskReference(parentID, existingTasks); err != nil {
			return nil, fmt.Errorf("invalid parent_id: %w", err)
		}
	}

	task := &models.Todo2Task{
		ID:              nextID,
		Name:            name,
		Content:         name,
		LongDescription: longDescription,
		Status:          status,
		Priority:        priority,
		Tags:            tags,
		Dependencies:    dependencies,
		Completed:       false,
		Metadata:        make(map[string]interface{}),
	}

	if task.ProjectID == "" && projectRoot != "" {
		task.ProjectID = filepath.Base(projectRoot)
	}

	if parentID != "" {
		task.ParentID = parentID
	} else if epicID != "" {
		task.ParentID = epicID
	}

	if planningDoc != "" || epicID != "" {
		linkMeta := &PlanningLinkMetadata{
			PlanningDoc: planningDoc,
			EpicID:      epicID,
		}
		SetPlanningLinkMetadata(task, linkMeta)
	}

	if backend := strings.TrimSpace(strings.ToLower(cast.ToString(params["local_ai_backend"]))); backend != "" {
		if backend == "fm" || backend == "mlx" || backend == "ollama" {
			task.Metadata[MetadataKeyPreferredBackend] = backend
		}
	}

	if recommendedTools := parseRecommendedToolsFromParams(params); len(recommendedTools) > 0 {
		slice := make([]interface{}, len(recommendedTools))
		for i, t := range recommendedTools {
			slice[i] = t
		}
		task.Metadata[MetadataKeyRecommendedTools] = slice
	}

	// Handle ownership metadata (owned_files, lane, etc.)
	if ownership := parseOwnershipFromParams(params); ownership != nil {
		models.SetTaskOwnership(task, ownership)
	}

	if err := store.CreateTask(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	result := map[string]interface{}{
		"success": true, "method": "store",
		"task": map[string]interface{}{
			"id": task.ID, "name": task.Name, "long_description": task.LongDescription,
			"status": task.Status, "priority": task.Priority, "tags": task.Tags, "dependencies": task.Dependencies,
		},
	}

	autoEstimate := cast.ToBool(params["auto_estimate"])
	if _, hasAutoEstimate := params["auto_estimate"]; !hasAutoEstimate {
		autoEstimate = true
	}

	if autoEstimate {
		if err := addEstimateComment(ctx, projectRoot, task, name, longDescription, tags, priority); err != nil {
			if metadata, ok := result["metadata"].(map[string]interface{}); ok {
				metadata["estimation_error"] = err.Error()
			} else {
				result["metadata"] = map[string]interface{}{
					"estimation_error": err.Error(),
				}
			}
		} else {
			if metadata, ok := result["metadata"].(map[string]interface{}); ok {
				metadata["estimation_added"] = true
			} else {
				result["metadata"] = map[string]interface{}{
					"estimation_added": true,
				}
			}
		}
	}

	outputPath := ParamOutputPath(params)
	return framework.FormatResult(result, outputPath)
}

// ─── handleTaskWorkflowCreateSingle ─────────────────────────────────────────
// handleTaskWorkflowCreateSingle creates a single task from flat params.
func handleTaskWorkflowCreateSingle(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task store: %w", err)
	}

	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	name := cast.ToString(params["name"])
	if name == "" {
		return nil, fmt.Errorf("name is required for task creation")
	}

	longDescription := cast.ToString(params["long_description"])
	if longDescription == "" {
		longDescription = name
	}

	status := config.DefaultTaskStatus()
	if rawStatus, err := ParamEnum(params, "status",
		[]string{"Todo", "In Progress", "Review", "Done", "Blocked", "Cancelled"},
		""); err != nil {
		return nil, fmt.Errorf("create: %w", err)
	} else if rawStatus != "" {
		status = normalizeStatus(rawStatus)
	}

	priority := config.DefaultTaskPriority()
	if rawPriority, err := ParamEnum(params, "priority",
		[]string{"low", "medium", "high", "critical"},
		""); err != nil {
		return nil, fmt.Errorf("create: %w", err)
	} else if rawPriority != "" {
		priority = normalizePriority(rawPriority)
	}

	tags := config.DefaultTaskTags()
	if len(tags) == 0 {
		tags = []string{}
	}
	if extra := ParamStringSliceTrimmedCommaSeparated(params, "tags"); len(extra) > 0 {
		tags = append(tags, extra...)
	}

	dependencies := ParamStringSliceTrimmedCommaSeparated(params, "dependencies")
	if dependencies == nil {
		dependencies = []string{}
	}

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	tasks := make([]Todo2Task, len(list))
	taskMap := make(map[string]bool)
	for i, t := range list {
		tasks[i] = *t
		taskMap[t.ID] = true
	}

	nextID := generateEpochTaskID()

	for _, dep := range dependencies {
		if !taskMap[dep] {
			return nil, fmt.Errorf("dependency %s does not exist", dep)
		}
	}

	var planningDoc string
	if pd := cast.ToString(params["planning_doc"]); pd != "" {
		planningDoc = pd
		if err := ValidatePlanningLink(projectRoot, planningDoc); err != nil {
			return nil, fmt.Errorf("invalid planning document link: %w", err)
		}
	}

	var epicID string
	if eid := cast.ToString(params["epic_id"]); eid != "" {
		epicID = eid
		if err := ValidateTaskReference(epicID, tasks); err != nil {
			return nil, fmt.Errorf("invalid epic ID: %w", err)
		}
	}

	parentID := cast.ToString(params["parent_id"])
	if parentID != "" {
		if err := ValidateTaskReference(parentID, tasks); err != nil {
			return nil, fmt.Errorf("invalid parent_id: %w", err)
		}
	}

	task := &models.Todo2Task{
		ID:              nextID,
		Content:         name,
		LongDescription: longDescription,
		Status:          status,
		Priority:        priority,
		Tags:            tags,
		Dependencies:    dependencies,
		Completed:       false,
		Metadata:        make(map[string]interface{}),
	}

	if task.ProjectID == "" && projectRoot != "" {
		task.ProjectID = filepath.Base(projectRoot)
	}

	if parentID != "" {
		task.ParentID = parentID
	} else if epicID != "" {
		task.ParentID = epicID
	}

	if planningDoc != "" || epicID != "" {
		linkMeta := &PlanningLinkMetadata{
			PlanningDoc: planningDoc,
			EpicID:      epicID,
		}
		SetPlanningLinkMetadata(task, linkMeta)
	}

	if backend := strings.TrimSpace(strings.ToLower(cast.ToString(params["local_ai_backend"]))); backend != "" {
		if backend == "fm" || backend == "mlx" || backend == "ollama" {
			task.Metadata[MetadataKeyPreferredBackend] = backend
		}
	}

	if recommendedTools := parseRecommendedToolsFromParams(params); len(recommendedTools) > 0 {
		slice := make([]interface{}, len(recommendedTools))
		for i, t := range recommendedTools {
			slice[i] = t
		}
		task.Metadata[MetadataKeyRecommendedTools] = slice
	}

	// Handle ownership metadata (owned_files, lane, etc.)
	if ownership := parseOwnershipFromParams(params); ownership != nil {
		models.SetTaskOwnership(task, ownership)
	}

	if err := store.CreateTask(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	result := map[string]interface{}{
		"success": true, "method": "store",
		"task": map[string]interface{}{
			"id": task.ID, "name": task.Content, "long_description": task.LongDescription,
			"status": task.Status, "priority": task.Priority, "tags": task.Tags, "dependencies": task.Dependencies,
		},
	}

	autoEstimate := true
	if _, ok := params["auto_estimate"]; ok {
		autoEstimate = cast.ToBool(params["auto_estimate"])
	}

	if autoEstimate {
		if err := addEstimateComment(ctx, projectRoot, task, name, longDescription, tags, priority); err != nil {
			if metadata, ok := result["metadata"].(map[string]interface{}); ok {
				metadata["estimation_error"] = err.Error()
			} else {
				result["metadata"] = map[string]interface{}{
					"estimation_error": err.Error(),
				}
			}
		} else {
			if metadata, ok := result["metadata"].(map[string]interface{}); ok {
				metadata["estimation_added"] = true
			} else {
				result["metadata"] = map[string]interface{}{
					"estimation_added": true,
				}
			}
		}
	}

	outputPath := ParamOutputPath(params)
	return framework.FormatResult(result, outputPath)
}

// ─── handleTaskWorkflowEnrichToolHints ──────────────────────────────────────
// handleTaskWorkflowEnrichToolHints applies tag→tool rules to Todo and In Progress tasks:
// merges recommended_tools from tags into task metadata. Uses default rules and optional
// .cursor/task_tool_rules.yaml. Idempotent; run on demand or after sync.
func handleTaskWorkflowEnrichToolHints(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("enrich_tool_hints: %w", err)
	}

	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("enrich_tool_hints: %w", err)
	}

	tagToTools := LoadTaskToolRules(projectRoot)

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("enrich_tool_hints: failed to load tasks: %w", err)
	}

	var updates []database.TaskMetadataUpdate
	for _, t := range list {
		if t == nil {
			continue
		}
		if !models.IsOpenStatus(t.Status) {
			continue
		}

		toolsFromTags := ToolsForTags(tagToTools, t.Tags)
		if len(toolsFromTags) == 0 {
			continue
		}

		existing := GetRecommendedTools(t.Metadata)
		merged := MergeRecommendedTools(existing, toolsFromTags)
		if t.Metadata == nil {
			t.Metadata = make(map[string]interface{})
		}
		slice := make([]interface{}, len(merged))
		for i, s := range merged {
			slice[i] = s
		}
		t.Metadata[MetadataKeyRecommendedTools] = slice

		updates = append(updates, database.TaskMetadataUpdate{
			TaskID:   t.ID,
			Metadata: t.Metadata,
		})
	}

	var updatedIDs []string
	if len(updates) > 0 {
		if _, err := database.BatchUpdateTaskMetadata(ctx, updates); err == nil {
			for _, u := range updates {
				updatedIDs = append(updatedIDs, u.TaskID)
			}
		}
	}

	if len(updatedIDs) > 0 {
		if syncErr := SyncTodo2Tasks(projectRoot); syncErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: sync after enrich_tool_hints failed: %v\n", syncErr)
		}
	}

	result := map[string]interface{}{
		"success":       true,
		"method":        "enrich_tool_hints",
		"updated_count": len(updatedIDs),
		"task_ids":      updatedIDs,
	}
	outputPath := ParamOutputPath(params)
	return framework.FormatResult(result, outputPath)
}

// ─── handleTaskWorkflowFixInvalidIDs ────────────────────────────────────────
// handleTaskWorkflowFixInvalidIDs finds tasks with invalid IDs (e.g. T-NaN), assigns new epoch IDs,
// updates dependencies, removes old DB rows, and saves. Use after sanity_check reports invalid_task_id.
// SQLite is canonical; JSON-only tasks are not supported.
func handleTaskWorkflowFixInvalidIDs(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	// Load from DB.
	taskMap := make(map[string]Todo2Task)

	if dbList, err := database.ListTasks(ctx, nil); err == nil {
		for _, t := range dbList {
			if t != nil {
				taskMap[t.ID] = *t
			}
		}
	}

	tasks := make([]Todo2Task, 0, len(taskMap))
	for _, t := range taskMap {
		tasks = append(tasks, t)
	}

	type fix struct{ oldID, newID string }

	var fixes []fix

	baseEpoch := time.Now().UnixMilli()
	counter := int64(0)

	for i := range tasks {
		if !models.IsValidTaskID(tasks[i].ID) {
			oldID := tasks[i].ID
			counter++
			newID := fmt.Sprintf("T-%d", baseEpoch+counter)
			tasks[i].ID = newID
			fixes = append(fixes, fix{oldID, newID})
		}
	}

	if len(fixes) == 0 {
		result := map[string]interface{}{
			"success": true, "method": "native_go",
			"fixed_count": 0,
			"message":     "no invalid task IDs found",
		}

		return framework.FormatResult(result, "")
	}

	// Update dependencies: any task referencing an old ID should reference the new ID
	oldToNew := make(map[string]string)
	fixedNewIDs := make(map[string]bool)

	for _, f := range fixes {
		oldToNew[f.oldID] = f.newID
		fixedNewIDs[f.newID] = true
	}

	for i := range tasks {
		for j, dep := range tasks[i].Dependencies {
			if newID, ok := oldToNew[dep]; ok {
				tasks[i].Dependencies[j] = newID
			}
		}
	}

	// Persist corrected tasks to both DB and JSON via SaveTodo2Tasks
	if err := SaveTodo2Tasks(projectRoot, tasks); err != nil {
		return nil, fmt.Errorf("failed to save fixed tasks: %w", err)
	}

	// Regenerate overview so .cursor/rules/todo2-overview.mdc reflects new IDs
	if overviewErr := WriteTodo2Overview(projectRoot); overviewErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to write todo2-overview.mdc: %v\n", overviewErr)
	}

	mapping := make(map[string]string)
	for _, f := range fixes {
		mapping[f.oldID] = f.newID
	}

	result := map[string]interface{}{
		"success":     true,
		"method":      "native_go",
		"fixed_count": len(fixes),
		"id_mapping":  mapping,
	}

	return framework.FormatResult(result, "")
}

// ─── generateEpochTaskID ────────────────────────────────────────────────────
// generateEpochTaskID generates a task ID using epoch nanoseconds (T-{epoch_nanoseconds}).
// Nanosecond precision avoids UNIQUE constraint failures when creating many tasks in a tight loop
// (e.g. task_discovery create_tasks). Format remains T-<digits> for IsValidTaskID compatibility.
func generateEpochTaskID() string {
	return database.GenerateTaskID()
}

// ─── normalizePriority ──────────────────────────────────────────────────────
// normalizePriority normalizes priority values to canonical lowercase form.
// This is a wrapper around NormalizePriority for backward compatibility.
func normalizePriority(priority string) string {
	return NormalizePriority(priority)
}

// ─── addEstimateComment ─────────────────────────────────────────────────────
// addEstimateComment estimates task duration and adds it as a comment
// This is called after task creation succeeds, and failures are handled gracefully.
// Uses task.Metadata["preferred_backend"] (fm|mlx|ollama) when set for local AI backend.
func addEstimateComment(ctx context.Context, projectRoot string, task *models.Todo2Task, name, details string, tags []string, priority string) error {
	estimationParams := map[string]interface{}{
		"action":         "estimate",
		"name":           name,
		"details":        details,
		"tag_list":       tags,
		"priority":       priority,
		"use_historical": true,
		"detailed":       false,
	}
	if backend := GetPreferredBackend(task.Metadata); backend != "" {
		estimationParams["local_ai_backend"] = backend
	}

	// Try native estimation first (platform-agnostic, will use statistical if Apple FM unavailable)
	estimationResult, err := handleEstimationNative(ctx, projectRoot, estimationParams)
	if err != nil {
		return fmt.Errorf("estimation failed: %w", err)
	}

	// Parse estimation result
	var estimate EstimationResult
	if err := json.Unmarshal([]byte(estimationResult), &estimate); err != nil {
		return fmt.Errorf("failed to parse estimation result: %w", err)
	}

	// Format estimate as markdown comment
	commentContent := formatEstimateComment(estimate)

	// Create comment
	comment := database.Comment{
		TaskID:  task.ID,
		Type:    "note",
		Content: commentContent,
	}

	// Add comment via database
	if err := database.AddComments(ctx, task.ID, []database.Comment{comment}); err != nil {
		// Database failed, try file-based fallback (future enhancement)
		// For now, just return error (task creation already succeeded)
		return fmt.Errorf("failed to add estimate comment: %w", err)
	}

	return nil
}

// ─── formatEstimateComment ──────────────────────────────────────────────────
// formatEstimateComment formats estimation result as a markdown comment.
func formatEstimateComment(estimate EstimationResult) string {
	var builder strings.Builder

	builder.WriteString("## Task Duration Estimate\n\n")
	builder.WriteString(fmt.Sprintf("**Estimated Duration:** %.1f hours\n\n", estimate.EstimateHours))
	builder.WriteString(fmt.Sprintf("**Confidence:** %.0f%%\n\n", estimate.Confidence*100))
	builder.WriteString(fmt.Sprintf("**Method:** %s\n", estimate.Method))

	if estimate.LowerBound > 0 && estimate.UpperBound > 0 {
		builder.WriteString(fmt.Sprintf("\n**Range:** %.1f - %.1f hours\n", estimate.LowerBound, estimate.UpperBound))
	}

	if len(estimate.Metadata) > 0 {
		if statisticalEst, ok := estimate.Metadata["statistical_estimate"].(float64); ok {
			builder.WriteString(fmt.Sprintf("\n**Statistical Estimate:** %.1f hours\n", statisticalEst))
		}

		if llmEst, ok := estimate.Metadata["llm_estimate"].(float64); ok {
			builder.WriteString(fmt.Sprintf("**AI Estimate:** %.1f hours\n", llmEst))
		} else if appleFMEst, ok := estimate.Metadata["apple_fm_estimate"].(float64); ok {
			builder.WriteString(fmt.Sprintf("**AI Estimate:** %.1f hours\n", appleFMEst))
		}
	}

	return builder.String()
}
