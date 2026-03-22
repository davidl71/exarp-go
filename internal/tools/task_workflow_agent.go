// task_workflow_agent.go — task_workflow claim/release actions for agent coordination.
// Used by Cursor Background Agent and other automated agents to atomically acquire tasks.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/models"
)

// handleTaskWorkflowClaim atomically claims a task for the calling agent.
//
// Params:
//   - task_id  (optional) — specific task to claim; if omitted, the next available
//     high-priority Todo task with no active lock is claimed.
//   - agent_id (optional) — defaults to database.GetAgentID() (EXARP_AGENT env + hostname + PID).
//   - lease_minutes (optional, default 30) — lock duration; must renew or release before expiry.
func handleTaskWorkflowClaim(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	taskID := ParamString(params, "task_id")

	agentID := ParamString(params, "agent_id")
	if agentID == "" {
		id, err := database.GetAgentID()
		if err != nil {
			return nil, fmt.Errorf("claim: could not determine agent_id: %w", err)
		}
		agentID = id
	}

	leaseMins := 30
	if v, ok := params["lease_minutes"].(float64); ok && v > 0 {
		leaseMins = int(v)
	}
	lease := time.Duration(leaseMins) * time.Minute

	if taskID == "" {
		next, err := findNextClaimableTask(ctx)
		if err != nil {
			return nil, fmt.Errorf("claim: finding next task: %w", err)
		}
		if next == nil {
			return text("no claimable tasks available (all Todo tasks are already locked)"), nil
		}
		taskID = next.ID
	}

	result, err := database.ClaimTaskForAgent(ctx, taskID, agentID, lease)
	if err != nil {
		return nil, fmt.Errorf("claim: %w", err)
	}
	if !result.Success {
		if result.WasLocked {
			return nil, fmt.Errorf("task %s is already claimed by %s", taskID, result.LockedBy)
		}
		return nil, fmt.Errorf("claim failed for task %s", taskID)
	}

	out := map[string]interface{}{
		"claimed":       true,
		"task_id":       taskID,
		"agent_id":      agentID,
		"lease_minutes": leaseMins,
		"lease_expires": time.Now().Add(lease).Format(time.RFC3339),
	}
	if result.Task != nil {
		out["task"] = map[string]interface{}{
			"id":       result.Task.ID,
			"content":  result.Task.Content,
			"status":   result.Task.Status,
			"priority": result.Task.Priority,
		}
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("claim: marshaling result: %w", err)
	}
	return []framework.TextContent{{Type: "text", Text: string(data)}}, nil
}

// handleTaskWorkflowRelease releases a task lock held by the calling agent.
//
// Params:
//   - task_id  (required)
//   - agent_id (optional) — defaults to database.GetAgentID().
func handleTaskWorkflowRelease(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	taskID := ParamString(params, "task_id")
	if taskID == "" {
		return nil, fmt.Errorf("release: task_id is required")
	}

	agentID := ParamString(params, "agent_id")
	if agentID == "" {
		id, err := database.GetAgentID()
		if err != nil {
			return nil, fmt.Errorf("release: could not determine agent_id: %w", err)
		}
		agentID = id
	}

	if err := database.ReleaseTask(ctx, taskID, agentID); err != nil {
		return nil, fmt.Errorf("release: %w", err)
	}

	out := map[string]interface{}{
		"released": true,
		"task_id":  taskID,
		"agent_id": agentID,
	}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("release: marshaling result: %w", err)
	}
	return []framework.TextContent{{Type: "text", Text: string(data)}}, nil
}

// findNextClaimableTask returns the first Todo task ordered by priority (high → critical → medium)
// with no persistent assignee. The database-level lock is checked atomically by ClaimTaskForAgent.
func findNextClaimableTask(ctx context.Context) (*models.Todo2Task, error) {
	for _, priority := range []string{models.PriorityHigh, models.PriorityCritical, models.PriorityMedium} {
		tasks, err := database.GetTasksByPriority(ctx, priority)
		if err != nil {
			continue
		}
		for _, t := range tasks {
			if t.Status != models.StatusTodo {
				continue
			}
			if t.AssignedTo == "" {
				return t, nil
			}
		}
	}
	return nil, nil
}

// text returns a single-item TextContent slice — convenience helper.
func text(s string) []framework.TextContent {
	return []framework.TextContent{{Type: "text", Text: s}}
}

// handleTaskWorkflowBatchClaim atomically claims multiple tasks for the calling agent.
// Useful for parallel agent workflows where multiple tasks can be worked on simultaneously.
//
// Params:
//   - task_ids (optional) — array of specific task IDs to claim
//   - count (optional, default 3) — max number of tasks to auto-claim if task_ids not specified
//   - agent_id (optional) — defaults to database.GetAgentID()
//   - lease_minutes (optional, default 30)
func handleTaskWorkflowBatchClaim(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	taskIDs, _ := params["task_ids"].([]interface{})

	agentID := ParamString(params, "agent_id")
	if agentID == "" {
		id, err := database.GetAgentID()
		if err != nil {
			return nil, fmt.Errorf("batch_claim: could not determine agent_id: %w", err)
		}
		agentID = id
	}

	leaseMins := 30
	if v, ok := params["lease_minutes"].(float64); ok && v > 0 {
		leaseMins = int(v)
	}
	lease := time.Duration(leaseMins) * time.Minute

	count := 3
	if c, ok := params["count"].(float64); ok && c > 0 {
		count = int(c)
	}

	results := []map[string]interface{}{}
	claimed := []string{}
	failed := []map[string]interface{}{}

	if len(taskIDs) > 0 {
		for _, id := range taskIDs {
			tid, ok := id.(string)
			if !ok {
				continue
			}
			result, err := database.ClaimTaskForAgent(ctx, tid, agentID, lease)
			if err != nil || !result.Success {
				errMsg := ""
				if result.Error != nil {
					errMsg = result.Error.Error()
				} else if err != nil {
					errMsg = err.Error()
				}
				failed = append(failed, map[string]interface{}{
					"task_id": tid,
					"error":   errMsg,
				})
				continue
			}
			claimed = append(claimed, tid)
			results = append(results, map[string]interface{}{
				"task_id": tid,
				"claimed": true,
			})
		}
	} else {
		for i := 0; i < count; i++ {
			next, err := findNextClaimableTask(ctx)
			if err != nil || next == nil {
				break
			}
			result, err := database.ClaimTaskForAgent(ctx, next.ID, agentID, lease)
			if err != nil || !result.Success {
				continue
			}
			claimed = append(claimed, next.ID)
			results = append(results, map[string]interface{}{
				"task_id":  next.ID,
				"claimed":  true,
				"content":  next.Content,
				"priority": next.Priority,
			})
		}
	}

	out := map[string]interface{}{
		"claimed_count": len(claimed),
		"claimed":       claimed,
		"failed":        failed,
		"agent_id":      agentID,
		"lease_minutes": leaseMins,
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("batch_claim: marshaling result: %w", err)
	}
	return []framework.TextContent{{Type: "text", Text: string(data)}}, nil
}

// handleTaskWorkflowAgentStatus shows all task locks held by an agent.
//
// Params:
//   - agent_id (optional) — defaults to database.GetAgentID()
func handleTaskWorkflowAgentStatus(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	agentID := ParamString(params, "agent_id")
	if agentID == "" {
		id, err := database.GetAgentID()
		if err != nil {
			return nil, fmt.Errorf("agent_status: could not determine agent_id: %w", err)
		}
		agentID = id
	}

	locks, err := database.GetLocksByAgent(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("agent_status: %w", err)
	}

	now := time.Now()
	lockInfos := []map[string]interface{}{}
	for _, lock := range locks {
		info := map[string]interface{}{
			"task_id":     lock.TaskID,
			"assigned_at": lock.AssignedAt.Format(time.RFC3339),
			"lock_until":  lock.LockUntil.Format(time.RFC3339),
		}
		if lock.LockUntil.Before(now) {
			info["status"] = "expired"
		} else {
			info["status"] = "active"
			info["remaining"] = lock.LockUntil.Sub(now).Round(time.Second).String()
		}
		lockInfos = append(lockInfos, info)
	}

	out := map[string]interface{}{
		"agent_id":   agentID,
		"lock_count": len(lockInfos),
		"locks":      lockInfos,
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("agent_status: marshaling result: %w", err)
	}
	return []framework.TextContent{{Type: "text", Text: string(data)}}, nil
}
