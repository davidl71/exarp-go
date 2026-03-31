package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/models"
	exarpb "github.com/davidl71/exarp-go/proto"
	gproto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func GetLatestLedgerSummary(projectRoot string) map[string]interface{} {
	return readLatestLedgerSummary(projectRoot)
}

func BuildAgentBriefingData(ctx context.Context, projectRoot string) (map[string]interface{}, error) {
	result, err := handleSessionPrime(ctx, map[string]interface{}{
		"include_hints": true,
		"include_tasks": true,
		"compact":       true,
	})
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("empty session prime result")
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(result[0].Text), &payload); err != nil {
		return nil, fmt.Errorf("parse session prime: %w", err)
	}

	activeLocks, _ := database.GetActiveLocks(ctx)
	activeRuns, _ := database.ListTaskExecutionRuns(ctx, "", "running", 10)
	if orchestration, err := BuildExecutionOrchestrationSummary(ctx, projectRoot, activeLocks, activeRuns); err == nil {
		for key, value := range orchestration {
			payload[key] = value
		}
	}

	if latestLedger := GetLatestLedgerSummary(projectRoot); latestLedger != nil {
		payload["latest_ledger_summary"] = latestLedger
	}

	if suggested, ok := payload["suggested_next"].([]interface{}); ok && len(suggested) > 0 {
		if first, ok := suggested[0].(map[string]interface{}); ok {
			if tools, ok := first["recommended_tools"]; ok {
				payload["top_recommended_tools"] = tools
			}
			if lazyContext, ok := first["lazy_context"]; ok {
				payload["top_lazy_context"] = lazyContext
			}
		}
	}

	payload["resource_kind"] = "agent_briefing"
	payload["generated_at"] = time.Now().Format(time.RFC3339)
	return payload, nil
}

func BuildTaskExecutionPackData(ctx context.Context, projectRoot, taskID string) (map[string]interface{}, error) {
	store := NewDefaultTaskStore(projectRoot)
	task, err := store.GetTask(ctx, taskID)
	if err != nil || task == nil {
		return nil, fmt.Errorf("task %s not found", taskID)
	}

	lockStatus := getActiveClaimForTask(ctx, taskID)
	runs, _ := database.ListTaskExecutionRuns(ctx, taskID, "", 5)
	verifications, _ := database.ListTaskVerifications(ctx, taskID, "", 5)
	progressEntries, _ := database.ListTaskProgressEntries(ctx, taskID, "", 5)

	taskMap := taskExecutionTaskMap(task)
	if allPtrs, listErr := store.ListTasks(ctx, nil); listErr == nil {
		allTasks := tasksFromPtrs(allPtrs)
		if wave := CalculateDependencyWave(allTasks, taskID); wave >= 0 {
			taskMap["execution_wave"] = wave
		}
	}

	workflowContract := BuildWorkflowContract(task, lockStatus, runs)
	pack := map[string]interface{}{
		"resource_kind": "task_execution_pack",
		"task":          taskMap,
		"generated_at":  time.Now().Format(time.RFC3339),
	}
	for key, value := range workflowContract {
		pack[key] = value
	}

	if lockStatus != nil {
		pack["active_claim"] = lockToMap(*lockStatus)
	}
	if len(runs) > 0 {
		items := make([]map[string]interface{}, 0, len(runs))
		for i := range runs {
			items = append(items, runToMap(&runs[i]))
		}
		pack["recent_runs"] = items
		pack["latest_run"] = runToMap(&runs[0])
	}
	if len(verifications) > 0 {
		pack["recent_verifications"] = verificationListToMaps(verifications)
	}
	if len(progressEntries) > 0 {
		pack["recent_progress"] = progressListToMaps(progressEntries)
	}
	if lazyContext := buildLazyTaskContext(*task); lazyContext != nil {
		pack["lazy_context"] = lazyContext
	}
	if role := AgentRoleFromTask(task); role != "" {
		pack["agent_role"] = role
		pack["agent_role_guidance"] = roleGuidance(role)
	}

	return pack, nil
}

// BuildTaskExecutionPackJSON returns the execution-pack resource as JSON bytes.
// It uses a small SQLite blob cache keyed by (task_id, task.version, kind) to avoid
// recomputing and re-marshaling for polling clients.
func BuildTaskExecutionPackJSON(ctx context.Context, projectRoot, taskID string) ([]byte, error) {
	store := NewDefaultTaskStore(projectRoot)
	task, err := store.GetTask(ctx, taskID)
	if err != nil || task == nil {
		return nil, fmt.Errorf("task %s not found", taskID)
	}

	const cacheKind = "task_execution_pack_v1"
	cachedBytes, _ := database.GetCacheBlob(ctx, taskID, task.Version, cacheKind)
	if len(cachedBytes) > 0 {
		var cached exarpb.CachedTaskExecutionPack
		if err := gproto.Unmarshal(cachedBytes, &cached); err == nil && len(cached.JsonPayload) > 0 {
			return cached.JsonPayload, nil
		}
	}

	pack, err := BuildTaskExecutionPackData(ctx, projectRoot, taskID)
	if err != nil {
		return nil, err
	}
	jsonBytes, err := json.Marshal(pack)
	if err != nil {
		return nil, fmt.Errorf("marshal execution pack: %w", err)
	}

	wrapped := &exarpb.CachedTaskExecutionPack{
		TaskId:      taskID,
		TaskVersion: task.Version,
		JsonPayload: jsonBytes,
		GeneratedAt: timestamppb.Now(),
	}
	if blob, err := gproto.Marshal(wrapped); err == nil {
		_ = database.PutCacheBlob(ctx, taskID, task.Version, cacheKind, blob)
	}

	return jsonBytes, nil
}

func BuildWorkflowContract(task *models.Todo2Task, activeClaim *database.LockStatus, runs []database.TaskExecutionRun) map[string]interface{} {
	return map[string]interface{}{
		"recommended_workflow":         buildRecommendedWorkflow(task, activeClaim, runs),
		"safe_next_actions":            buildSafeNextActions(task, activeClaim, runs),
		"required_preconditions":       buildRequiredPreconditions(task, activeClaim, runs),
		"completion_evidence_required": buildCompletionEvidenceRequired(task),
	}
}

func BuildAgentAlertsData(ctx context.Context, projectRoot string) (map[string]interface{}, error) {
	activeLocks, err := database.GetActiveLocks(ctx)
	if err != nil {
		return nil, err
	}
	activeRuns, err := database.ListTaskExecutionRuns(ctx, "", "running", 25)
	if err != nil {
		return nil, err
	}

	expiredClaims := make([]map[string]interface{}, 0)
	for _, lock := range activeLocks {
		if lock.IsExpired || lock.TimeRemaining < 0 {
			expiredClaims = append(expiredClaims, lockToMap(lock))
		}
	}

	const staleThreshold = 10 * time.Minute
	staleRuns := make([]map[string]interface{}, 0)
	for i := range activeRuns {
		if time.Since(activeRuns[i].StartedAt) > staleThreshold {
			staleRuns = append(staleRuns, runToMap(&activeRuns[i]))
		}
	}

	reviewReady, err := listTasksByStatus(ctx, projectRoot, models.StatusReview)
	if err != nil {
		return nil, err
	}
	suggested := GetSuggestedNextTasks(projectRoot, 1)
	var nextTask interface{}
	if len(suggested) > 0 {
		nextTask = map[string]interface{}{
			"id":       suggested[0].ID,
			"content":  suggested[0].Content,
			"priority": suggested[0].Priority,
			"status":   suggested[0].Status,
		}
	}

	return map[string]interface{}{
		"resource_kind":         "agent_alerts",
		"expired_claims":        expiredClaims,
		"stale_runs":            staleRuns,
		"review_ready_tasks":    reviewReady,
		"review_ready_count":    len(reviewReady),
		"approval_needed_count": len(reviewReady),
		"recommended_next_task": nextTask,
		"generated_at":          time.Now().Format(time.RFC3339),
	}, nil
}

func taskExecutionTaskMap(task *models.Todo2Task) map[string]interface{} {
	result := map[string]interface{}{
		"id":               task.ID,
		"content":          task.Content,
		"long_description": task.LongDescription,
		"status":           task.Status,
		"priority":         task.Priority,
		"tags":             task.Tags,
		"dependencies":     task.Dependencies,
		"parent_id":        task.ParentID,
		"metadata":         task.Metadata,
	}
	if rt := GetRecommendedTools(task.Metadata); len(rt) > 0 {
		result["recommended_tools"] = rt
	}
	return result
}

func getActiveClaimForTask(ctx context.Context, taskID string) *database.LockStatus {
	locks, err := database.GetActiveLocks(ctx)
	if err != nil {
		return nil
	}
	for _, lock := range locks {
		if lock.TaskID == taskID {
			copied := lock
			return &copied
		}
	}
	return nil
}

func buildRecommendedWorkflow(task *models.Todo2Task, activeClaim *database.LockStatus, runs []database.TaskExecutionRun) []string {
	role := AgentRoleFromTask(task)
	if len(runs) > 0 && runs[0].Status == "running" {
		switch role {
		case AgentRoleReviewer:
			return []string{"inspect current run", "record verification evidence", "end run with review summary"}
		case AgentRolePlanner:
			return []string{"inspect current run", "capture progress on planning state", "end run with split or next-step summary"}
		default:
			return []string{"inspect current run", "add progress", "record verification", "end run"}
		}
	}
	if activeClaim != nil || task.Status == models.StatusInProgress {
		switch role {
		case AgentRoleReviewer:
			return []string{"review task context", "start run", "record verification evidence"}
		case AgentRolePlanner:
			return []string{"review task context", "start run", "clarify or split work"}
		default:
			return []string{"review task context", "start run", "record progress slices"}
		}
	}
	switch role {
	case AgentRoleReviewer:
		return []string{"claim task", "review execution context", "start run", "record review evidence"}
	case AgentRolePlanner:
		return []string{"claim task", "review dependencies", "start run", "clarify or split task"}
	default:
		return []string{"claim task", "inspect task context", "start run", "record progress and verification"}
	}
}

func buildSafeNextActions(task *models.Todo2Task, activeClaim *database.LockStatus, runs []database.TaskExecutionRun) []string {
	if len(runs) > 0 && runs[0].Status == "running" {
		return []string{"add_progress", "verify", "end_run"}
	}
	if activeClaim != nil || task.Status == models.StatusInProgress {
		return []string{"start_run", "add_progress"}
	}
	return []string{"claim", "start_run", "read_resource"}
}

func buildRequiredPreconditions(task *models.Todo2Task, activeClaim *database.LockStatus, runs []database.TaskExecutionRun) []string {
	var out []string
	if len(task.Dependencies) > 0 {
		out = append(out, "confirm dependencies are complete or intentionally bypassed")
	}
	if activeClaim == nil && len(runs) == 0 {
		out = append(out, "claim task before starting autonomous execution")
	}
	if role := AgentRoleFromTask(task); role == AgentRoleReviewer {
		out = append(out, "load recent run or progress evidence before review")
	}
	if len(out) == 0 {
		out = append(out, "inspect task context before mutating task state")
	}
	return out
}

func buildCompletionEvidenceRequired(task *models.Todo2Task) []string {
	role := AgentRoleFromTask(task)
	switch role {
	case AgentRoleReviewer:
		return []string{"verification result", "review summary"}
	case AgentRolePlanner:
		return []string{"planning summary", "updated dependencies or child tasks"}
	default:
		return []string{"progress summary", "verification result"}
	}
}

func roleGuidance(role string) string {
	switch role {
	case AgentRolePlanner:
		return "Bias toward clarification, decomposition, dependencies, and sequencing."
	case AgentRoleReviewer:
		return "Bias toward verification, review evidence, and completion criteria."
	case AgentRoleResearcher:
		return "Bias toward exploration, evidence gathering, and recommendation notes."
	default:
		return "Bias toward implementation progress and execution evidence."
	}
}

func listTasksByStatus(ctx context.Context, projectRoot, status string) ([]map[string]interface{}, error) {
	store := NewDefaultTaskStore(projectRoot)
	filters := &database.TaskFilters{Status: &status}
	tasks, err := store.ListTasks(ctx, filters)
	if err != nil {
		return nil, err
	}
	items := make([]map[string]interface{}, 0, len(tasks))
	for _, task := range tasks {
		items = append(items, map[string]interface{}{
			"id":       task.ID,
			"content":  task.Content,
			"priority": task.Priority,
			"status":   task.Status,
		})
	}
	return items, nil
}
