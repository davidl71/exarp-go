# Access Control Model

**Date:** 2026-02-02  
**Status:** Design (T-278)  
**Implementation:** `internal/security/access.go`, config `security.access_control`

---

## Purpose

This document describes the access control model for the exarp-go MCP server: who can do what to which objects, how policy is defined, and how it is enforced.

---

## Scope

- **In scope:** Tool invocation and resource read access.
- **Out of scope (current):** User/identity, roles, prompts, per-request attributes. The server runs over STDIO; typically a single client (e.g. Cursor) is assumed.

---

## Model

### Subjects

- **Current:** Implicit single “caller” (the MCP client). No user or role identity is passed; access control is process-wide.
- **Future:** If the transport gains identity (e.g. token, client ID), the model can be extended to per-subject policy.

### Objects

| Object type | Identifier | Actions |
|-------------|------------|--------|
| **Tool**   | Tool name (e.g. `task_workflow`, `report`) | Invoke |
| **Resource** | Resource URI (e.g. `stdio://sources`) | Read |

Prompts are not currently part of the access model; they could be added as a third object type.

### Actions

- **Tool:** Invoke (call with arguments).
- **Resource:** Read (fetch content).

There is no fine-grained “write” or “delete” at the object level; tools perform their own operations internally.

### Permissions

- **Allow** — The subject may perform the action on the object.
- **Deny** — The subject may not; the server returns an access-denied error.
- **Default** — No explicit rule; fall back to the default policy.

---

## Policy

### Default policy

- **Allow (default):** If no explicit rule matches, access is allowed. Used for local development and permissive setups.
- **Deny:** If no explicit rule matches, access is denied. Used for locked-down environments; then only explicitly allowed tools/resources are permitted.

### Explicit rules

- **Allow list:** Names/URIs that are explicitly allowed (used when default is Deny).
- **Deny list:** Names/URIs that are explicitly denied (used when default is Allow). Deny overrides allow for the same object.

### Evaluation order

1. **Explicit permission** for the object (Allow or Deny) → use it.
2. **Deny list** → if the object is in the deny list, deny.
3. **Allow list** (when default is Deny) → if the object is not in the allow list, deny.
4. **Default policy** → allow or deny according to default.

So: explicit rule > deny list > allow list (under default-deny) > default policy.

---

## Configuration

Config lives under `security.access_control` (YAML and centralized config).

| Setting | Type | Meaning |
|--------|------|--------|
| `enabled` | bool | If true, apply access control; if false, allow all (current behavior). |
| `default_policy` | string | `"allow"` or `"deny"`. |
| `restricted_tools` | list of strings | Tool names to deny (deny list). Used when `enabled` is true and default is allow. |

Example (deny list, default allow):

```yaml
security:
  access_control:
    enabled: true
    default_policy: allow
    restricted_tools:
      - delete_todos
      - git_tools
```

Example (allow list, default deny):

```yaml
security:
  access_control:
    enabled: true
    default_policy: deny
    # In code, allowed tools would be configured separately; config currently supports restricted_tools only.
```

The implementation builds an `AccessControl` from config: `NewAccessControlFromConfig()` sets default policy and populates the deny list from `restricted_tools`. Allow list for default-deny would require an extra config field (e.g. `allowed_tools`) if added.

---

## Enforcement

- **Tools:** Before dispatching a tool call, the server calls `security.CheckToolAccess(toolName)`. On error (`AccessDeniedError`), the request is rejected and the client receives an access-denied response.
- **Resources:** Before returning a resource, the server calls `security.CheckResourceAccess(uri)`. On error, the resource is not returned.

Integration points: tool dispatch and resource handlers (e.g. in `internal/tools/handlers.go` or the framework adapter). Default is permissive (allow all) when access control is disabled.

---

## Error handling

- **AccessDeniedError:** Resource type (`"tool"` or `"resource"`) and name/URI. Clients can show a clear “access denied” message.
- No distinction yet between “not found” and “denied”; both can result in error responses.

---

## Future extensions

- **Roles:** Map a subject (or client) to a role; map role to allow/deny lists.
- **Allowed list in config:** `allowed_tools` / `allowed_resources` for default-deny deployments.
- **Prompts:** Add prompt IDs as objects and check before returning a prompt.
- **Per-resource templates:** e.g. allow `stdio://sources` but deny `stdio://secrets`.
- **Audit logging:** Log denied (and optionally allowed) requests for compliance.

---

## References

- Implementation: `internal/security/access.go`, `internal/security/access_test.go`
- Config: `internal/config/schema.go` (`AccessControlConfig`), `internal/config/defaults.go`, `internal/config/loader.go`
- Status: `docs/SECURITY_IMPLEMENTATION_STATUS.md` (Phase 3: Access Control)
