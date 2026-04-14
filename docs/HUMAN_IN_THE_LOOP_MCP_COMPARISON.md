# Human-in-the-loop (historical note)

exarp-go previously documented a third-party human-review MCP integration. That integration has been **removed** from the codebase.

**Current behavior:** `task_workflow` still supports generic review payloads and status updates:

- `request_approval`, `sync_approvals` — return `approval_request` / `approval_requests` (see `internal/tools/approval_request.go`).
- `apply_approval_result` — set task to Done or In Progress from a human decision.
- Moving a task to **Review** via `update` may include `approval_requests` and `review_instructions` in the response.

Wire these payloads to whatever review process you use (email, ticketing, custom MCP, etc.); exarp-go does not call an external review service.

For design discussion of in-app vs external approval, see `docs/research/APPROVAL_PATTERN_DESIGN.md`.
