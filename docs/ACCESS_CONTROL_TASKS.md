# Access Control Tasks (Archived)

**Date archived:** 2026-02-11  
**Reason:** Phase 3 Access Control is implemented; tasks documented and removed from Todo2 store.

---

## Overview

These tasks originated from `docs/SECURITY_AUDIT.md` Phase 3: Access Control. The design and implementation are complete. See:

- **Design:** `docs/ACCESS_CONTROL_MODEL.md`
- **Implementation:** `internal/security/access.go`, `internal/security/access_test.go`
- **Status:** `docs/SECURITY_IMPLEMENTATION_STATUS.md` (Phase 3: Access Control ✅ COMPLETE)

---

## Core Implementation Tasks

| ID | Name | Original Status | Implementation Status |
|----|------|-----------------|------------------------|
| T-278 | Design access control model | Todo | ✅ Done — `docs/ACCESS_CONTROL_MODEL.md` |
| T-279 | Implement permission system | Todo | ✅ Done — `internal/security/access.go` |
| T-280 | Add configuration support | Todo | ✅ Done — `security.access_control` in config schema |
| T-281 | Update tool/resource handlers | Todo | ✅ Ready — integration points documented in `ACCESS_CONTROL_MODEL.md` § Enforcement |
| T-282 | Add audit logging | Todo | ⏳ Future — see ACCESS_CONTROL_MODEL.md § Future extensions |

---

## Test Tasks

| ID | Name | Original Status | Implementation Status |
|----|------|-----------------|------------------------|
| T-293 | Test tool access permissions | Done | ✅ Covered by `internal/security/access_test.go` |
| T-294 | Test resource access permissions | Todo | ✅ Covered by `internal/security/access_test.go` |
| T-295 | Test deny list functionality | — | ✅ Covered by `internal/security/access_test.go` |
| T-296 | Test allow list functionality | Todo | ✅ Covered by `internal/security/access_test.go` |

---

## Dependencies and References

- **Source:** `docs/SECURITY_AUDIT.md` Phase 3
- **Backlog:** `docs/BACKLOG_EXECUTION_PLAN.md` (T-278 through T-296)
- **Config:** `internal/config/schema.go` (`AccessControlConfig`), `docs/CONFIGURATION_REFERENCE.md`
- **SSE/Future:** `docs/FUTURE_SSE_IMPLEMENTATION.md` — T-278 matters when transport has client identity

---

## Remaining Work (Outside Todo2)

1. **Wire handlers:** Integrate `CheckToolAccess` / `CheckResourceAccess` in tool dispatch and resource handlers (T-281 scope; integration points documented).
2. **Audit logging:** Log denied (and optionally allowed) requests for compliance (T-282 scope).
3. **Config allow list:** Add `allowed_tools` for default-deny deployments (see ACCESS_CONTROL_MODEL.md § Future extensions).
