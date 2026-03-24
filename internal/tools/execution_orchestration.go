package tools

import (
	"context"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/models"
)

type executionLaneSummary struct {
	Role             string                   `json:"role"`
	TotalCount       int                      `json:"total_count"`
	TodoCount        int                      `json:"todo_count"`
	InProgressCount  int                      `json:"in_progress_count"`
	ActiveClaimCount int                      `json:"active_claim_count"`
	ActiveRunCount   int                      `json:"active_run_count"`
	SuggestedTasks   []map[string]interface{} `json:"suggested_tasks,omitempty"`
}

func BuildExecutionOrchestrationSummary(ctx context.Context, projectRoot string, activeLocks []database.LockStatus, activeRuns []database.TaskExecutionRun) (map[string]interface{}, error) {
	store := NewDefaultTaskStore(projectRoot)
	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, err
	}

	tasks := tasksFromPtrs(list)
	taskByID := make(map[string]Todo2Task, len(tasks))
	roleCounts := make(map[string]int)
	claimCounts := make(map[string]int)
	runCounts := make(map[string]int)
	tasksWithRole := 0

	lanes := map[string]*executionLaneSummary{}
	for _, role := range orderedAgentRoles() {
		lanes[role] = &executionLaneSummary{Role: role}
	}

	for _, task := range tasks {
		role := AgentRoleFromTask(&task)
		if role == "" {
			continue
		}
		taskByID[task.ID] = task
		tasksWithRole++
		roleCounts[role]++

		lane := lanes[role]
		lane.TotalCount++
		switch task.Status {
		case models.StatusTodo:
			lane.TodoCount++
		case models.StatusInProgress:
			lane.InProgressCount++
		}
		if (task.Status == models.StatusTodo || task.Status == models.StatusInProgress) && len(lane.SuggestedTasks) < 3 {
			lane.SuggestedTasks = append(lane.SuggestedTasks, map[string]interface{}{
				"id":       task.ID,
				"content":  task.Content,
				"status":   task.Status,
				"priority": task.Priority,
			})
		}
	}

	for _, lock := range activeLocks {
		task, ok := taskByID[lock.TaskID]
		if !ok {
			continue
		}
		role := AgentRoleFromTask(&task)
		if role == "" {
			continue
		}
		claimCounts[role]++
		lanes[role].ActiveClaimCount++
	}

	for _, run := range activeRuns {
		task, ok := taskByID[run.TaskID]
		if !ok {
			continue
		}
		role := AgentRoleFromTask(&task)
		if role == "" {
			continue
		}
		runCounts[role]++
		lanes[role].ActiveRunCount++
	}

	laneList := make([]map[string]interface{}, 0, len(lanes))
	for _, role := range orderedAgentRoles() {
		lane := lanes[role]
		if lane.TotalCount == 0 && lane.ActiveClaimCount == 0 && lane.ActiveRunCount == 0 {
			continue
		}
		laneList = append(laneList, map[string]interface{}{
			"role":               lane.Role,
			"total_count":        lane.TotalCount,
			"todo_count":         lane.TodoCount,
			"in_progress_count":  lane.InProgressCount,
			"active_claim_count": lane.ActiveClaimCount,
			"active_run_count":   lane.ActiveRunCount,
			"suggested_tasks":    lane.SuggestedTasks,
		})
	}

	return map[string]interface{}{
		"agent_role_summary": map[string]interface{}{
			"distribution":          roleCounts,
			"dominant_role":         dominantAgentRole(tasks),
			"tasks_with_agent_role": tasksWithRole,
			"active_claims_by_role": claimCounts,
			"active_runs_by_role":   runCounts,
		},
		"orchestration_lanes":    laneList,
		"delegation_suggestions": buildDelegationSuggestions(lanes),
	}, nil
}

func buildDelegationSuggestions(lanes map[string]*executionLaneSummary) []map[string]interface{} {
	var out []map[string]interface{}
	for _, role := range orderedAgentRoles() {
		lane := lanes[role]
		if lane == nil || lane.TodoCount == 0 || lane.ActiveRunCount > 0 {
			continue
		}
		taskIDs := make([]string, 0, len(lane.SuggestedTasks))
		for _, task := range lane.SuggestedTasks {
			if id, ok := task["id"].(string); ok && id != "" {
				taskIDs = append(taskIDs, id)
			}
		}
		out = append(out, map[string]interface{}{
			"role":            role,
			"reason":          delegationReason(role, lane.TodoCount),
			"suggested_tasks": taskIDs,
		})
	}
	return out
}

func delegationReason(role string, todoCount int) string {
	switch role {
	case AgentRolePlanner:
		return "Planner lane has ready work and no active run; delegate planning or task-shaping next."
	case AgentRoleWorker:
		return "Worker lane has implementation slices ready with no active run."
	case AgentRoleReviewer:
		return "Reviewer lane has review-ready work and no active run."
	case AgentRoleResearcher:
		return "Research lane has exploration work queued with no active run."
	default:
		return "Agent role has ready work and no active run."
	}
}

func orderedAgentRoles() []string {
	return []string{AgentRolePlanner, AgentRoleWorker, AgentRoleReviewer, AgentRoleResearcher}
}

// CalculateDependencyWave returns the dependency wave (depth) for a task in the open backlog.
// Wave 0 means the task has no unmet open dependencies and can start immediately.
// Wave N means at least one dependency is in wave N-1, so this task cannot start until that wave completes.
// Returns -1 if the task ID is not found among open (non-Done/Cancelled) tasks.
func CalculateDependencyWave(tasks []Todo2Task, taskID string) int {
	_, waves, _, err := BacklogExecutionOrder(tasks, nil)
	if err != nil {
		return -1
	}
	for wave, ids := range waves {
		for _, id := range ids {
			if id == taskID {
				return wave
			}
		}
	}
	return -1
}
