# gotoHuman API / Tools Reference

**Canonical reference for exarp-go integration.**  
**Server:** `gotohuman` MCP Server (e.g. `@gotohuman/mcp-server`)  
**Auth:** API key required (`GOTOHUMAN_API_KEY` or params).

---

## Quick Reference

| exarp-go action | gotoHuman tool | Purpose |
|-----------------|----------------|---------|
| `task_workflow` `request_approval` | `request-human-review-with-form` | Build payload for one task, send to gotoHuman |
| `task_workflow` `sync_approvals` | `request-human-review-with-form` | Get all Review tasks as approval requests; send each to gotoHuman |
| `task_workflow` `apply_approval_result` | (human decides in gotoHuman) | Update Todo2 status from approval/rejection |
| — | `list-forms` | Get form IDs for approval requests |
| — | `get-form-schema` | Inspect form fields before sending |

---

## Tools

### `list-forms`

List all available review forms.

| Aspect | Value |
|--------|--------|
| **Purpose** | List form IDs and names for use in approval requests |
| **Parameters** | None (API key via env or params) |
| **Returns** | List of form IDs and metadata |

**Usage (Cursor):** `@gotoHuman list-forms`

---

### `get-form-schema`

Get the schema for a specific form (field definitions and types).

| Aspect | Value |
|--------|--------|
| **Purpose** | Inspect form fields before calling `request-human-review-with-form` |
| **Parameters** | `formId` (string) — form ID from `list-forms` |
| **Returns** | Form schema with field definitions |

**Usage (Cursor):** `@gotoHuman get-form-schema <formId>`

---

### `request-human-review-with-form`

Send a human approval request using a form.

| Aspect | Value |
|--------|--------|
| **Purpose** | Request human review/approval; human receives in gotoHuman inbox |
| **Parameters** | |
| | `formId` (string) — form to use |
| | `fieldData` (object) — form field values |
| | `metadata` (object, optional) — extra context |
| | `assignToUsers` (array, optional) — user emails to assign |
| **Returns** | Review request ID (for status/response handling) |

**Usage (Cursor):** `@gotoHuman request-human-review-with-form` with formId and fieldData.

---

## Setup

- **API key:** Set `GOTOHUMAN_API_KEY` in env or in MCP server `env` in `~/.cursor/mcp.json`.
- **Setup guide:** [docs/GOTOHUMAN_SETUP.md](GOTOHUMAN_SETUP.md)
- **Integration plan:** [docs/GOTOHUMAN_INTEGRATION_PLAN.md](GOTOHUMAN_INTEGRATION_PLAN.md)
- **Discovery summary:** [docs/GOTOHUMAN_DISCOVERY_SUMMARY.md](GOTOHUMAN_DISCOVERY_SUMMARY.md)

---

## Integration with task_workflow (exarp-go)

Planned enhancements (see plan T-106–T-112):

- When a task moves to **Review**, send an approval request via gotoHuman (`request-human-review-with-form`).
- Map Todo2 task fields to gotoHuman form `fieldData` (e.g. task_id, title, description).
- Poll or receive approval/rejection and update Todo2 (e.g. Done / back to In Progress).

**task_workflow actions:**  
- **`request_approval`** — `task_id` (required), optional `form_id`. Returns one `approval_request` and instructions to call @gotoHuman.  
- **`sync_approvals`** (T-111) — Returns `approval_requests` for all tasks in Review; optional `form_id`. Send each to @gotoHuman.  
- **`apply_approval_result`** (T-112) — `task_id`, `result=approved|rejected`; optional `feedback` for rejection. Sets status to Done (approved) or In Progress (rejected); rejection feedback appended to task.  
Helper: `internal/tools/gotohuman.BuildApprovalRequestFromTask`.

---

## Testing the approval flow (T-106)

1. **Set API key:** `export GOTOHUMAN_API_KEY="your-key"` (or add to MCP server env in `~/.cursor/mcp.json`).
2. **Get a form ID:** In Cursor, call `@gotoHuman list-forms` and note a form ID.
3. **Build approval payload:** Call `task_workflow` with `action=request_approval`, `task_id=<a Todo2 task ID>`, and optional `form_id=<from list-forms>`.
4. **Send to gotoHuman:** Call `@gotoHuman request-human-review-with-form` with the returned `approval_request.form_id` (or your form_id) and `approval_request.field_data`.
5. **When moving to Review:** Use `task_workflow` with `action=update`, `task_ids=<id>`, `new_status=Review`. The response includes `approval_requests` and `goto_human_instructions`; use each payload with @gotoHuman as above.

---

## End-to-end approval workflow test (T-110)

1. **Move task to Review:** `task_workflow` with `action=update`, `task_ids=T-xxx`, `new_status=Review`. Response includes `approval_requests` and `goto_human_instructions`.
2. **Send to gotoHuman:** For each item in `approval_requests`, call `@gotoHuman request-human-review-with-form` with that item’s `form_id` (or a default from list-forms) and `field_data`. Human receives the request in gotoHuman.
3. **Human decides:** In gotoHuman, human approves or rejects.
4. **Apply result in Todo2:** Call `task_workflow` with `action=apply_approval_result`, `task_id=T-xxx`, `result=approved` (or `rejected`). Approved → task status set to Done; rejected → set to In Progress. Optional `feedback` for rejection.
5. **Batch sync (optional):** Use `task_workflow` with `action=sync_approvals` to get approval requests for all current Review tasks; send each to gotoHuman, then apply results as above.

**Success criteria:** Task moves to Review → approval sent to gotoHuman → human approves/rejects → Todo2 status updated accordingly.

---

---

## See Also

- [GOTOHUMAN_SETUP.md](GOTOHUMAN_SETUP.md) — API key and server setup
- [GOTOHUMAN_INTEGRATION_PLAN.md](GOTOHUMAN_INTEGRATION_PLAN.md) — Integration roadmap
- [GOTOHUMAN_TOOLS_DOCUMENTATION.md](GOTOHUMAN_TOOLS_DOCUMENTATION.md) — Discovery notes

**Last updated:** 2026-02-02
