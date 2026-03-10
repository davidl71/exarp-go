---
name: Vendor Improvements
overview: Restored plan referenced by completed vendor improvement tasks. This plan points to the source analysis document and preserves task metadata links.
todos:
  - id: T-1771431392695147000
    content: Vendor improvements: Adopt spf13/cast in internal/tools
    status: done
  - id: T-1771431879332536000
    content: Vendor: Extend spf13/cast to session, task_workflow_common, handlers
    status: done
  - id: T-1771431882371389000
    content: Vendor: Run make vendor-licenses in pre-release or CI
    status: done
  - id: T-1771431400302079000
    content: Vendor improvements: Document rate limiter vs rate package
    status: done
  - id: T-1771431401371255000
    content: Vendor improvements: Optional use of go-humanize in report/CLI
    status: done
isProject: false
---

# Vendor Improvements

**Generated:** 2026-03-10

**Status:** restored

**Last updated:** 2026-03-10

**Source analysis:** [docs/VENDOR_USAGE_AND_IMPROVEMENTS.md](../../docs/VENDOR_USAGE_AND_IMPROVEMENTS.md)

## Scope

This plan restores the planning document path referenced by completed vendor-related tasks.
The substantive analysis and recommendations live in `docs/VENDOR_USAGE_AND_IMPROVEMENTS.md`.

## Linked Tasks

| Task | Status | Notes |
| --- | --- | --- |
| **T-1771431392695147000** | done | Adopt `spf13/cast` in internal tools. |
| **T-1771431879332536000** | done | Extend `cast` usage into more tool entry points. |
| **T-1771431882371389000** | done | Automate `make vendor-licenses` in release or CI flow. |
| **T-1771431400302079000** | done | Document when to use the sliding-window limiter vs `x/time/rate`. |
| **T-1771431401371255000** | done | Track optional `go-humanize` usage for CLI/report formatting. |
