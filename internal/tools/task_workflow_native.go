// task_workflow_native.go — MCP "task_workflow" tool entry point: action dispatch switch.
// Handlers are in task_workflow_crud.go (approve/sync/create/update),
// task_workflow_actions.go (clarify/approval/delete/fixDates), and task_workflow_common.go (shared infra).
package tools

import (
	"context"
	"fmt"

	"github.com/davidl71/exarp-go/internal/framework"
)

// handleTaskWorkflowNative handles task_workflow with native Go and FM chain (Apple → Ollama → stub).
func handleTaskWorkflowNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("task_workflow: %w", err)
	}

	ctx = context.WithValue(ctx, taskStoreKey, NewDefaultTaskStore(projectRoot))

	action, _ := params["action"].(string)
	if action == "" {
		action = "sync"
	}

	switch action {
	case "clarify":
		return handleTaskWorkflowClarify(ctx, params)
	case "approve":
		return handleTaskWorkflowApprove(ctx, params)
	case "sync":
		return handleTaskWorkflowSync(ctx, params)
	case "fix_dates":
		return handleTaskWorkflowFixDates(ctx, params)
	case "fix_empty_descriptions":
		return handleTaskWorkflowFixEmptyDescriptions(ctx, params)
	case "clarity":
		return handleTaskWorkflowClarity(ctx, params)
	case "cleanup":
		return handleTaskWorkflowCleanup(ctx, params)
	case "create":
		return handleTaskWorkflowCreate(ctx, params)
	case "sanity_check":
		return handleTaskWorkflowSanityCheck(ctx, params)
	case "fix_invalid_ids":
		return handleTaskWorkflowFixInvalidIDs(ctx, params)
	case "link_planning":
		return handleTaskWorkflowLinkPlanning(ctx, params)
	case "delete":
		return handleTaskWorkflowDelete(ctx, params)
	case "add_comment":
		return handleTaskWorkflowAddComment(ctx, params)
	case "update":
		return handleTaskWorkflowUpdate(ctx, params)
	case "request_approval":
		return handleTaskWorkflowRequestApproval(ctx, params)
	case "sync_approvals":
		return handleTaskWorkflowSyncApprovals(ctx, params)
	case "apply_approval_result":
		return handleTaskWorkflowApplyApprovalResult(ctx, params)
	case "sync_from_plan", "sync_plan_status":
		return handleTaskWorkflowSyncFromPlan(ctx, params)
	case "summarize":
		return handleTaskWorkflowSummarize(ctx, params)
	case "run_with_ai":
		return handleTaskWorkflowRunWithAI(ctx, params)
	case "enrich_tool_hints":
		return handleTaskWorkflowEnrichToolHints(ctx, params)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}
