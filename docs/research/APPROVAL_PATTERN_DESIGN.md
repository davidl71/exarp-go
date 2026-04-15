# Promise-Based Approval Pattern for Review Workflow

**Date:** 2026-02-26
**Status:** Design / Research
**Source:** OpenWork (different-ai/openwork) pattern analysis

---

## Problem

The current Review → Done transition relies on polling or manual intervention:
1. AI moves task to "Review" status
2. Human must check task status and explicitly approve
3. No structured feedback mechanism beyond task comments
4. No timeout handling for abandoned reviews

## Current Implementation

```
task status: Review
  → Human reads task, checks result comment
  → Human runs: exarp-go task update T-xxx --new-status Done
  → Or: AI calls handleTaskWorkflowUpdate with new_status=Done
```

The `handleTaskWorkflowApplyApprovalResult` action (T-112) adds gotoHuman integration but requires an external service.

## Proposed Enhancement: In-Process Approval Queue

Inspired by OpenWork's Promise-based approval service, create an in-memory approval queue that:

1. **Registers pending approvals** when a task enters Review
2. **Waits with timeout** for human response (via MCP elicitation, CLI, or TUI)
3. **Resolves** approval with accept/reject + optional feedback
4. **Auto-expires** approvals that exceed the configured timeout

### Architecture

```
handleTaskWorkflowUpdate(status=Review)
  → ApprovalService.RequestApproval(taskID, timeout)
  → Returns immediately with approval_id

TUI/CLI/MCP client polls or subscribes:
  → ApprovalService.GetPending() → list of pending approvals
  → ApprovalService.Resolve(approval_id, decision, feedback)

Timeout handler:
  → After N minutes, auto-resolve as "needs_attention"
  → Optionally notify via toast/desktop notification
```

### Go Implementation Sketch

```go
type ApprovalRequest struct {
    ID        string
    TaskID    string
    Content   string
    CreatedAt time.Time
    ExpiresAt time.Time
    Result    chan ApprovalResult
}

type ApprovalResult struct {
    Decision string // "approved", "rejected", "expired"
    Feedback string
}

type ApprovalService struct {
    mu      sync.Mutex
    pending map[string]*ApprovalRequest
}

func (s *ApprovalService) RequestApproval(taskID, content string, timeout time.Duration) *ApprovalRequest {
    req := &ApprovalRequest{
        ID:        generateID(),
        TaskID:    taskID,
        Content:   content,
        CreatedAt: time.Now(),
        ExpiresAt: time.Now().Add(timeout),
        Result:    make(chan ApprovalResult, 1),
    }
    s.mu.Lock()
    s.pending[req.ID] = req
    s.mu.Unlock()

    go func() {
        select {
        case <-time.After(timeout):
            req.Result <- ApprovalResult{Decision: "expired"}
            s.remove(req.ID)
        }
    }()

    return req
}
```

## Integration Points

| Component | How it connects |
|-----------|----------------|
| **task_workflow (Review action)** | Calls `ApprovalService.RequestApproval()` |
| **TUI** | Shows pending approvals tab; resolve with keyboard |
| **MCP elicitation** | Uses `ElicitForm` for inline approval |
| **CLI** | `exarp-go task approve T-xxx` resolves the pending approval |
| **OpenCode plugin** | Could use `client.session.prompt()` to surface approvals |

## Trade-offs

| Pro | Con |
|-----|-----|
| Structured approval flow | Adds in-memory state (lost on restart) |
| Timeout prevents stale reviews | Complexity vs. current manual approach |
| Works with TUI, CLI, and MCP | Needs persistence for multi-session approvals |
| Feedback captured automatically | gotoHuman already provides external approval |

## Recommendation

**Phase 1 (low effort):** Add an `ApprovalService` singleton with in-memory pending queue. Wire into `handleTaskWorkflowUpdate` when status changes to Review. Expose via `task_workflow action=sync_approvals` (already exists). Add TUI tab for approvals.

**Phase 2 (if needed):** Persist pending approvals to SQLite for cross-session survival. Add webhook/notification integration.

**Decision:** Worth implementing as Phase 1 — it enhances the existing Review workflow without requiring external services. The gotoHuman integration (T-111, T-112) remains for external human review.

## References

- OpenWork approval service: `docs/research/OPENCODE_PLUGIN_PATTERNS.md` §3.1
- Current approval handlers: `internal/tools/task_workflow_actions.go`
- gotoHuman integration: `docs/GOTOHUMAN_API_REFERENCE.md`
- MCP elicitation: `internal/framework/elicitation.go`
