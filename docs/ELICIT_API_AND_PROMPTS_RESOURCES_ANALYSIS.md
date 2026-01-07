# Elicit API & Prompts/Resources Analysis

**Date:** 2026-01-07  
**Status:** ✅ Analysis Complete

---

## Elicit API Support

### ❌ **exarp-go does NOT support Elicit API**

**Why:**
- **Elicit API is FastMCP-specific** - Not part of standard MCP protocol
- **STDIO transport limitation** - Elicit requires FastMCP Context, which isn't available in STDIO mode
- **Go SDK limitation** - Standard MCP Go SDKs don't implement FastMCP-specific features

**What is Elicit API?**
- FastMCP's `ctx.elicit()` method for inline chat questions
- Allows AI to ask questions directly in chat (not pop-ups)
- Requires FastMCP Context object (not available in STDIO)

**Current Status:**
- `demonstrate_elicit` - Demo tool (FastMCP only, doesn't work in stdio)
- `interactive_task_create` - Example tool (FastMCP only, doesn't work in stdio)

**Impact:** ⚠️ **MINIMAL** - These are demo/example tools that don't work in Cursor's STDIO mode anyway

**Alternatives:**
- Standard MCP tools can return questions in their responses
- User can respond via chat naturally
- No special API needed for basic interaction

---

## Prompts Comparison

### exarp-go: **8 prompts** ✅
1. ✅ `align` - Task alignment analysis
2. ✅ `discover` - Task discovery
3. ✅ `config` - Config generation
4. ✅ `scan` - Security scanning
5. ✅ `scorecard` - Project scorecard
6. ✅ `overview` - Project overview
7. ✅ `dashboard` - Project dashboard
8. ✅ `remember` - Memory system

### Coordinator: **~30 prompts** (22 unique)

**Unique Prompts in Coordinator (NOT in exarp-go):**

**High Priority:**
- `daily_checkin` - Daily check-in workflow ⭐
- `sprint_start` - Sprint start workflow ⭐
- `sprint_end` - Sprint end workflow ⭐
- `pre_sprint` - Pre-sprint cleanup ⭐
- `post_impl` - Post-implementation review ⭐
- `dups` - Find duplicate tasks ⭐
- `sync` - Sync tasks between TODO and Todo2 ⭐

**Medium Priority:**
- `doc_check` - Documentation health check
- `doc_quick` - Quick doc check
- `weekly` - Weekly maintenance
- `project_health` - Full health assessment
- `end_of_day` - End of day handoff
- `resume_session` - Resume session
- `view_handoffs` - View handoffs
- `automation_setup` - One-time setup
- `auto` / `auto_high` - Automation discovery

**Low Priority (Persona-specific):**
- `dev`, `pm`, `reviewer`, `exec` - Persona workflows
- `mode` - Suggest Cursor IDE mode
- `context` - LLM context management

**Total:** ~22 unique prompts that could be migrated

---

## Resources Comparison

### exarp-go: **6 resources** ✅
1. ✅ `stdio://scorecard` - Project scorecard
2. ✅ `stdio://memories` - All memories
3. ✅ `stdio://memories/category/{category}` - Memories by category
4. ✅ `stdio://memories/task/{task_id}` - Memories for task
5. ✅ `stdio://memories/recent` - Recent memories
6. ✅ `stdio://memories/session/{date}` - Session memories

### Coordinator: **Need to verify**
- May have additional resources (status, tasks, etc.)
- Likely similar to exarp-go's resources

---

## Recommendations

### Elicit API:
- ✅ **Safe to disable coordinator** - Elicit tools don't work in stdio anyway
- ✅ **No migration needed** - Demo tools only

### Prompts:
- ⚠️ **Consider migrating high-priority prompts** if you use them
- ✅ **exarp-go has core prompts** - 8 essential prompts already migrated
- 📋 **Optional:** Migrate workflow prompts (`daily_checkin`, `sprint_start`, etc.) if needed

### Resources:
- ✅ **exarp-go likely complete** - Has all essential resources
- 📋 **Verify coordinator resources** - Check if any unique resources exist

---

**Conclusion:** Safe to disable coordinator. exarp-go has all functional tools and core prompts/resources.

