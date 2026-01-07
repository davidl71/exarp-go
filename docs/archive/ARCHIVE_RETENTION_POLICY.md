# Documentation Archive Retention Policy

**Date:** 2026-01-07  
**Status:** Active

---

## Purpose

This document defines the retention policy for archived documentation files in `docs/archive/`.

---

## Archive Structure

```
docs/archive/
├── migration-planning/      # Pre-migration planning documents
├── implementation-summaries/ # Completed implementation summaries
├── research-phase/          # Research phase documentation
├── analysis/                # Pre-migration analysis documents
├── status-updates/          # Historical status updates
└── ARCHIVE_RETENTION_POLICY.md  # This file
```

---

## Retention Periods

### Category 1: Migration Planning (6 months)
**Files:** Migration planning, task summaries, migration workflows  
**Retention:** 6 months from archive date  
**Delete After:** 2026-07-07  
**Rationale:** Planning documents become irrelevant once migration is complete and stable

### Category 2: Implementation Summaries (3 months)
**Files:** Batch summaries, completion summaries, task summaries  
**Retention:** 3 months from archive date  
**Delete After:** 2026-04-07  
**Rationale:** Implementation summaries are historical records, not reference material

### Category 3: Research Phase (6 months)
**Files:** Research phase docs, research comments, parallel research  
**Retention:** 6 months from archive date  
**Delete After:** 2026-07-07  
**Rationale:** Research findings may be useful for future reference, but become stale over time

### Category 4: Analysis Documents (12 months)
**Files:** Framework analysis, duplication analysis, documentation analysis  
**Retention:** 12 months from archive date  
**Delete After:** 2027-01-07  
**Rationale:** Analysis documents may contain useful insights for future decisions

### Category 5: Status Updates (1 month)
**Files:** Status updates, agent execution status, task assignments  
**Retention:** 1 month from archive date  
**Delete After:** 2026-02-07  
**Rationale:** Status updates are ephemeral and quickly become outdated

---

## Deletion Schedule

### Automatic Review Dates

| Category | Next Review | Delete After |
|----------|-------------|--------------|
| Status Updates | 2026-02-01 | 2026-02-07 |
| Implementation Summaries | 2026-04-01 | 2026-04-07 |
| Migration Planning | 2026-07-01 | 2026-07-07 |
| Research Phase | 2026-07-01 | 2026-07-07 |
| Analysis Documents | 2027-01-01 | 2027-01-07 |

### Review Process

1. **Monthly Review:** Check status updates category (1 month retention)
2. **Quarterly Review:** Check implementation summaries (3 months)
3. **Semi-Annual Review:** Check migration planning and research phase (6 months)
4. **Annual Review:** Check analysis documents (12 months)

---

## Deletion Criteria

### Safe to Delete
- ✅ Files older than retention period
- ✅ No references in active documentation
- ✅ No historical value for future work
- ✅ Superseded by newer documentation

### Keep Longer
- ⚠️ Files referenced in active docs (extend retention)
- ⚠️ Files with unique insights not captured elsewhere
- ⚠️ Files that may be needed for audits or compliance

### Never Delete
- 🔒 Files with legal/compliance requirements
- 🔒 Files documenting critical architectural decisions
- 🔒 Files that are the only record of important information

---

## Exclusion from Automated Tools

### Linting
- ✅ Archive excluded from `gomarklint` (via `.gomarklint.json`)
- ✅ Archive excluded from other markdown linters

### Testing
- ✅ Archive excluded from documentation tests
- ✅ Archive excluded from link checking

### CI/CD
- ✅ Archive excluded from automated documentation checks
- ✅ Archive excluded from build processes

---

## Manual Review Process

Before deleting files:

1. **Check References:** Search codebase for references to archived files
2. **Check Git History:** Verify file history is preserved in git
3. **Check Dependencies:** Ensure no active docs depend on archived content
4. **Document Deletion:** Update this file with deletion date and reason

---

## Current Archive Status

**Total Files Archived:** 49 files  
**Archive Date:** 2026-01-07  
**Next Review:** 2026-02-01 (Status Updates)

### Files by Category

- **Migration Planning:** 8 files (Delete: 2026-07-07)
- **Implementation Summaries:** 11 files (Delete: 2026-04-07)
- **Research Phase:** 12 files (Delete: 2026-07-07)
- **Analysis:** 9 files (Delete: 2027-01-07)
- **Status Updates:** 9 files (Delete: 2026-02-07)

---

## Maintenance

### Monthly Tasks
- [ ] Review status updates category
- [ ] Check for broken references
- [ ] Update this document if needed

### Quarterly Tasks
- [ ] Review implementation summaries
- [ ] Check retention dates
- [ ] Delete expired files

### Annual Tasks
- [ ] Review all categories
- [ ] Update retention policy if needed
- [ ] Archive this policy document if superseded

---

## Notes

- **Git History:** All archived files remain in git history even after deletion
- **Recovery:** Files can be recovered from git if needed
- **Documentation:** This policy should be reviewed annually
- **Exceptions:** Retention periods can be extended if files have ongoing value

---

**Last Updated:** 2026-01-07  
**Next Review:** 2026-02-01

