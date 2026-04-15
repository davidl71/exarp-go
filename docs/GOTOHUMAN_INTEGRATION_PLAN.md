# gotoHuman Integration Plan

## Overview

This document outlines the plan to integrate **gotoHuman** (human-in-the-loop MCP server) with the Todo2 workflow for automated approval requests.

## Current Status

✅ **gotoHuman MCP Server Added to Configuration**
- Location: `.cursor/mcp.json`
- Server: `gotohuman`
- Status: Configured and ready to load

## gotoHuman Capabilities (Research Needed)

Based on the description: "Human-in-the-loop platform - Allow AI agents and automations to send requests for approval to your gotoHuman inbox"

**Expected Capabilities:**
- Send approval requests to human inbox
- Check approval status
- Receive approval/rejection responses
- Manage approval workflow

**Note:** Specific tools and API need to be discovered after server is loaded in Cursor.

## Integration Points

### 1. Todo2 Workflow Integration

**Current Workflow:**
```
[IN PROGRESS]
    ↓ status: Review + result
[REVIEW - AWAITING HUMAN FEEDBACK] ⚠️ HUMAN APPROVAL REQUIRED ⚠️
    ↓ human approval → status: Done
```

**Enhanced Workflow with gotoHuman:**
```
[IN PROGRESS]
    ↓ status: Review + result
[REVIEW - AWAITING HUMAN FEEDBACK] ⚠️ HUMAN APPROVAL REQUIRED ⚠️
    ↓ AI sends approval request to gotoHuman
    ↓ gotoHuman sends notification to human inbox
    ↓ Human reviews in gotoHuman interface
    ↓ Human approves/rejects via gotoHuman
    ↓ gotoHuman notifies AI of decision
    ↓ If approved → status: Done
    ↓ If rejected → status: In Progress (with feedback)
```

### 2. Task Workflow Tool Enhancement

**Current Tool:** `task_workflow` (handles task lifecycle)

**Enhancement:** gotoHuman integration added to `task_workflow` tool:

```go
// Enhanced task_workflow handler
func handleTaskWorkflow(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
    // Parse action
    action := params["action"].(string)
    
    switch action {
    case "approve":
        // Send approval request to gotoHuman
        return requestApproval(ctx, taskID, taskDetails)
    case "check_approval":
        // Check approval status from gotoHuman
        return checkApprovalStatus(ctx, requestID)
    case "sync":
        // Sync Todo2 Review tasks with gotoHuman inbox
        return syncReviewTasks(ctx)
    }
}
```

### 3. New Tools to Add

**Option A: Direct gotoHuman Tools (if available)**
- Use gotoHuman MCP tools directly from Cursor
- No code changes needed
- Depends on gotoHuman tool availability

**Option B: Wrapper Tools in exarp-go**
- `request_task_approval` - Send Todo2 task for approval
- `check_approval_status` - Check status of approval request
- `sync_approvals` - Sync Todo2 Review tasks with gotoHuman

## Implementation Plan

### Phase 1: Discovery (Current)
- [x] Add gotoHuman to MCP configuration
- [ ] Load gotoHuman server in Cursor
- [ ] Discover available gotoHuman tools
- [x] Document gotoHuman API/tools — see [GOTOHUMAN_API_REFERENCE.md](GOTOHUMAN_API_REFERENCE.md)
- [x] Test basic approval request flow — see [GOTOHUMAN_API_REFERENCE.md](GOTOHUMAN_API_REFERENCE.md) "Testing the approval flow"

### Phase 2: Basic Integration
- [x] Create approval request helper function — `internal/tools/gotohuman.BuildApprovalRequestFromTask`
- [x] Enhance `task_workflow` tool with approval action — `action=request_approval` + approval payload in update when new_status=Review
- [x] Add approval request when task moves to Review — update response includes `approval_requests` and `goto_human_instructions`
- [x] Test approval workflow end-to-end — see GOTOHUMAN_API_REFERENCE.md "End-to-end approval workflow test"

### Phase 3: Advanced Integration
- [x] Auto-sync Review tasks with gotoHuman — task_workflow action=sync_approvals returns approval_requests for all Review tasks
- [x] Handle approval/rejection responses — task_workflow action=apply_approval_result (task_id, result=approved|rejected, optional feedback)
- [ ] Update Todo2 task status based on approval
- [ ] Add approval history tracking

### Phase 4: Workflow Enhancement
- [ ] Add approval timeout handling
- [ ] Add approval reminders
- [ ] Add approval delegation
- [ ] Add approval analytics

## Code Changes Required

### 1. Update `task_workflow` Handler

**File:** `internal/tools/handlers.go`

```go
// handleTaskWorkflow handles the task_workflow tool
func handleTaskWorkflow(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
    var params map[string]interface{}
    if err := json.Unmarshal(args, &params); err != nil {
        return nil, fmt.Errorf("failed to parse arguments: %w", err)
    }

    action := params["action"].(string)
    
    switch action {
    case "approve":
        // New: Send approval request to gotoHuman
        return requestTaskApproval(ctx, params)
    case "check_approval":
        // New: Check approval status
        return checkApprovalStatus(ctx, params)
    default:
        // Existing: Delegate to Python bridge
        result, err := bridge.ExecutePythonTool(ctx, "task_workflow", params)
        if err != nil {
            return nil, fmt.Errorf("task_workflow failed: %w", err)
        }
        return []framework.TextContent{
            {Type: "text", Text: result},
        }, nil
    }
}
```

### 2. Add Approval Helper Functions

**New File:** `internal/tools/approval.go`

```go
package tools

import (
    "context"
    "encoding/json"
    "fmt"
    
    "github.com/davidl71/exarp-go/internal/framework"
)

// requestTaskApproval sends a Todo2 task approval request to gotoHuman
func requestTaskApproval(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
    taskID := params["task_id"].(string)
    taskDetails := params["task_details"].(map[string]interface{})
    
    // TODO: Call gotoHuman MCP tool to send approval request
    // This requires discovering gotoHuman's tool API first
    
    return []framework.TextContent{
        {Type: "text", Text: fmt.Sprintf("Approval request sent for task %s", taskID)},
    }, nil
}

// checkApprovalStatus checks the status of an approval request
func checkApprovalStatus(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
    requestID := params["request_id"].(string)
    
    // TODO: Call gotoHuman MCP tool to check approval status
    
    return []framework.TextContent{
        {Type: "text", Text: fmt.Sprintf("Approval status for request %s: pending", requestID)},
    }, nil
}
```

### 3. Update Todo2 Workflow Rules

**File:** `.cursor/rules/todo2.mdc`

Add gotoHuman integration to Review workflow:

```markdown
**✅ CORRECT REVIEW WORKFLOW (Enhanced):**
1. **AI completes work** → Adds result comment → Moves to "Review"
2. **AI sends approval request** → Calls gotoHuman to request approval
3. **gotoHuman notifies human** → Human receives notification in inbox
4. **Human reviews** → Reviews task in gotoHuman interface
5. **Human approves/rejects** → Decision sent via gotoHuman
6. **If approved** → gotoHuman notifies AI → AI marks task as "Done"
7. **If rejected** → gotoHuman sends feedback → AI moves to "In Progress" → AI fixes issues
```

## Testing Plan

### 1. Basic Functionality
- [ ] Test gotoHuman server loads in Cursor
- [ ] List available gotoHuman tools
- [ ] Test sending approval request
- [ ] Test checking approval status

### 2. Integration Testing
- [ ] Test task moves to Review → approval request sent
- [ ] Test human approves → task moves to Done
- [ ] Test human rejects → task moves back to In Progress
- [ ] Test approval timeout handling

### 3. End-to-End Testing
- [ ] Complete workflow: Task → Review → Approval → Done
- [ ] Multiple concurrent approval requests
- [ ] Approval cancellation
- [ ] Approval delegation

## Next Steps

1. **Restart Cursor** to load gotoHuman MCP server
2. **Discover gotoHuman tools** - List available tools in Cursor
3. **Test basic approval** - Send test approval request
4. **Document API** - Document gotoHuman tool signatures
5. **Implement integration** - Add approval helpers to exarp-go
6. **Test workflow** - End-to-end approval workflow testing

## References

- **gotoHuman MCP Server:** `@gotohuman/mcp-server` (npm)
- **MCP Proxy:** `mcpower-proxy==0.0.87` (UV)
- **Configuration:** `.cursor/mcp.json`
- **Todo2 Workflow:** `.cursor/rules/todo2.mdc`

---

**Last Updated:** 2026-01-07
**Status:** ✅ Configuration added, ready for discovery phase

