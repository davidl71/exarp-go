# Coordinator Prompts & Resources Analysis

**Date:** 2026-01-07  
**Purpose:** Identify prompts and resources in coordinator that should be migrated to exarp-go

---

## Elicit API Support

### ❌ **exarp-go does NOT support Elicit API**

**Reasons:**
1. **FastMCP-specific** - Elicit API (`ctx.elicit()`) is a FastMCP feature, not part of standard MCP protocol
2. **STDIO limitation** - Requires FastMCP Context object, which isn't available in STDIO transport
3. **Go SDK limitation** - Standard MCP Go SDKs (go-sdk, mcp-go) don't implement FastMCP-specific features

**What you'll lose:**
- `demonstrate_elicit` - Demo tool (doesn't work in stdio anyway)
- `interactive_task_create` - Example tool (doesn't work in stdio anyway)

**Impact:** ⚠️ **MINIMAL** - These are demo/example tools that don't work in Cursor's STDIO mode

**Alternatives:**
- Standard MCP tools can return questions in responses
- Natural chat interaction (no special API needed)
- Can implement similar functionality using standard tool responses

---

## Prompts Comparison

### exarp-go: **8 prompts**
1. ✅ `align` - Task alignment analysis
2. ✅ `discover` - Task discovery
3. ✅ `config` - Config generation
4. ✅ `scan` - Security scanning
5. ✅ `scorecard` - Project scorecard
6. ✅ `overview` - Project overview
7. ✅ `dashboard` - Project dashboard (exarp-go only)
8. ✅ `remember` - Memory system

### Coordinator: **~30 prompts** (many unique)

**Prompts in coordinator NOT in exarp-go:**

**Documentation:**
- `doc_check` - Analyze documentation health and create tasks
- `doc_quick` - Quick documentation health check

**Task Management:**
- `dups` - Find and consolidate duplicate Todo2 tasks
- `sync` - Synchronize tasks between shared TODO table and Todo2

**Automation:**
- `auto` - Discover new automation opportunities
- `auto_high` - Find only high-value automation opportunities

**Workflow Prompts:**
- `pre_sprint` - Pre-sprint cleanup workflow
- `post_impl` - Post-implementation review workflow
- `weekly` - Weekly maintenance workflow
- `daily_checkin` - Daily check-in workflow
- `sprint_start` - Sprint start workflow
- `sprint_end` - Sprint end workflow
- `project_health` - Full project health assessment
- `automation_setup` - One-time automation setup

**Session Management:**
- `end_of_day` - End work session and create handoff
- `resume_session` - Resume work by reviewing handoff
- `view_handoffs` - View recent handoff notes

**Persona Workflows:**
- `dev` - Developer daily workflow
- `pm` - Project Manager workflow
- `reviewer` - Code Reviewer workflow
- `exec` - Executive/Stakeholder workflow

**Other:**
- `mode` - Suggest optimal Cursor IDE mode (Agent vs Ask)
- `context` - Manage LLM context with summarization

**Total Unique Prompts in Coordinator: ~22 prompts**

**✅ MIGRATED (7 high-value prompts):**
- `daily_checkin` - Daily check-in workflow ✅
- `sprint_start` - Sprint start workflow ✅
- `sprint_end` - Sprint end workflow ✅
- `pre_sprint` - Pre-sprint cleanup ✅
- `post_impl` - Post-implementation review ✅
- `sync` - Synchronize tasks ✅
- `dups` - Find duplicate tasks ✅

**Remaining Unique Prompts: ~15 prompts** (medium/low priority)

---

## Resources Comparison

### exarp-go: **6 resources**
1. ✅ `stdio://scorecard` - Project scorecard
2. ✅ `stdio://memories` - All memories
3. ✅ `stdio://memories/category/{category}` - Memories by category
4. ✅ `stdio://memories/task/{task_id}` - Memories for task
5. ✅ `stdio://memories/recent` - Recent memories
6. ✅ `stdio://memories/session/{date}` - Session memories

### Coordinator: **Need to verify**
- May have additional resources (status, tasks, etc.)
- Need to check coordinator's resource list

---

## Migration Recommendations

### High Priority Prompts to Migrate:

**Workflow Prompts (Most Useful):**
1. `daily_checkin` - Daily check-in workflow ⭐ **HIGH VALUE**
2. `sprint_start` - Sprint start workflow ⭐ **HIGH VALUE**
3. `sprint_end` - Sprint end workflow ⭐ **HIGH VALUE**
4. `pre_sprint` - Pre-sprint cleanup ⭐ **HIGH VALUE**
5. `post_impl` - Post-implementation review ⭐ **HIGH VALUE**

**Task Management:**
6. `dups` - Find duplicate tasks ⭐ **HIGH VALUE**
7. `sync` - Sync tasks between TODO and Todo2 ⭐ **HIGH VALUE**

**Documentation:**
8. `doc_check` - Documentation health check ⭐ **MEDIUM VALUE**

**Session Management:**
9. `end_of_day` - End of day handoff ⭐ **MEDIUM VALUE**
10. `resume_session` - Resume session ⭐ **MEDIUM VALUE**

### Medium Priority:
- `weekly` - Weekly maintenance
- `project_health` - Full health assessment
- `automation_setup` - One-time setup
- `auto` / `auto_high` - Automation discovery

### Low Priority (Persona-specific):
- `dev`, `pm`, `reviewer`, `exec` - Persona workflows (may be less critical)

---

## Summary

**Elicit API:** ❌ Not supported, but not needed (demo tools only)

**Prompts:** ✅ **7 high-value prompts migrated** to exarp-go

**Resources:** ✅ exarp-go has all essential resources

**Status:**
- ✅ **Coordinator disabled** - No longer needed
- ✅ **High-value prompts migrated** - 7 workflow prompts added to exarp-go
- ✅ **exarp-go is primary** - 24 tools, 15 prompts, 6 resources

**Remaining:** ~15 medium/low priority prompts in coordinator (can migrate later if needed)

