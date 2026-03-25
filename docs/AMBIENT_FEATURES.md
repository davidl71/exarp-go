# Ambient Features Reference

**Last Updated:** 2026-03-25

Ambient features keep the AI agent contextually aware **without explicit prompts**. The agent "just knows" your tasks, progress, and project state.

---

## Current Ambient Features

### 1. System Prompt Injection

**Hook:** `experimental.chat.system.transform`  
**What it does:** Injects full task state into every LLM system prompt  
**Timing:** Every message, cached 30s TTL  
**Skipped for:** Sub-agents, title generation

```
Context: You have 6 Todo tasks, 1 In Progress.
Suggested next: Work on T-123: Fix auth middleware
⚠️ Ownership warnings: session.go shared by T-1, T-2
```

### 2. First-Message Injection

**Hook:** `chat.message` (first message only)  
**What it does:** Prepends compact task summary  
**Marked:** `synthetic: true` (agent knows it's injected)

```
📋 Tasks: 6 Todo, 1 In Progress
Next: T-123456 - Fix auth middleware (high priority)
```

### 3. Tool Output Reminders

**Hook:** `tool.execute.after`  
**What it does:** Appends in-progress task reminders to tool output  
**Effect:** Agent maintains task awareness during tool use

```
[After running tests]
Reminder: You are working on T-123: Fix auth middleware
```

### 4. Toast Notifications

**Hook:** `session.created`, `todo.updated`  
**What it does:** Visual popup in TUI  
**Content:** Task counts, change notifications

```
"6 tasks: 5 Todo, 1 In Progress" (3 second duration)
```

### 5. Compaction Context

**Hook:** `experimental.session.compacting`  
**What it does:** Preserves task state in compaction summaries  
**Effect:** Task context survives context window resets

### 6. Prompt Auto-Expand

**Hook:** `tui.prompt.append`  
**What it does:** When user types `T-123`, appends task details  
**Trigger:** Task ID pattern in prompt input

```
User types: "Let's do T-123"
Auto-appends: "[T-123] Fix auth middleware | Priority: high | Status: Todo"
```

### 7. Ownership Warnings (NEW)

**Location:** `ownership_warnings` in session prime  
**What it does:** Warns about file collisions in suggested tasks  
**Source:** `buildOwnershipHints()` analyzes task ownership metadata

```
"ownership_warnings": [
  "⚠️ File collision: session.go shared by T-1, T-2 (run serially)",
  "⚠️ Same lane (session): T-1, T-2 — may have related files"
]
```

### 8. Plugin Tools (Fast Path)

**What it does:** Native tools faster than MCP round-trip  
**Tools:** `exarp_tasks`, `exarp_update_task`, `exarp_prime`, `exarp_config`, `exarp_followup`

---

## Proposed Ambient Features

### File Watch → Task Correlation

**Concept:** Detect which files the agent edits and correlate to tasks  
**Implementation:**
- Plugin hooks into `tool.execute.after` for edit tools
- Tracks file paths modified
- Updates task's `owned_files` automatically
- Suggests marking task as "In Progress" when file touched

**Output:**
```
📁 You edited internal/tools/session.go
   Mapped to: T-123 (Update session prime)
   Suggested: Mark task as In Progress?
```

### Git Hook Integration

**Concept:** Auto-update task status from git events  
**Implementation:**
- Pre-commit hook: warn if committing to task without marking In Progress
- Post-commit hook: suggest marking Done, link commit to task
- Branch naming: detect `T-123-*` in branch name, auto-claim task

**Output:**
```
🌿 Branch T-123-fix-auth detected
   Claiming task: T-123 (Fix auth middleware)
   Owner: agent-xyz, expires: 2h
```

### Build/Test Result Awareness

**Concept:** Inject build/test results into context  
**Implementation:**
- Hook into `shell.execute.after` for make/test commands
- Parse exit code and output
- Inject pass/fail into next LLM context

**Output:**
```
🧪 Tests: 142 passed, 3 failed (auth_test.go:45, session_test.go:12)
```

### Deadline/Time Awareness

**Concept:** Time-based hints and warnings  
**Implementation:**
- Track task age (created vs now)
- Warn about stale tasks (>7 days untouched)
- Standup prep: auto-generate status from recent completions

**Output:**
```
⏰ Stale tasks (untouched >7 days):
   - T-45: Update docs (14 days)
   - T-67: Refactor legacy auth (21 days)

📋 Standup summary ready: 3 completed, 2 in progress, 1 blocked
```

### Cross-Session Memory

**Concept:** Persistent learnings across sessions  
**Implementation:**
- After task completion, store key learnings
- Inject relevant memories when similar task appears
- Pattern: "Last time you fixed auth, you found X issue"

**Output:**
```
💡 Memory: When fixing auth middleware, check JWT expiry logic
   (stored from T-99 completion, 2026-03-20)
```

### Dependency Chain Awareness

**Concept:** Warn when starting task with incomplete dependencies  
**Implementation:**
- Check task dependencies in prime
- Warn if blocking tasks not Done
- Suggest dependency order

**Output:**
```
🔗 Dependency warning: T-123 depends on T-100 (Todo)
   Recommended: Complete T-100 first, or confirm bypass
```

### Code Review Readiness

**Concept:** Detect when code is ready for review  
**Implementation:**
- After test pass + lint clean + no TODOs
- Suggest marking task as Review
- Auto-generate PR description from task

**Output:**
```
✅ T-123 appears ready for review:
   - Tests: pass (142/142)
   - Lint: clean
   - Files changed: 3
   Suggested: Mark as Review?
```

### Error Pattern Detection

**Concept:** Detect repeated errors and suggest fixes  
**Implementation:**
- Track error patterns in session
- When same error appears 2+ times, suggest root cause
- Link to past solutions if available

**Output:**
```
🔄 Same error 3 times: "cannot read property 'x' of undefined"
   Likely cause: Missing null check in auth middleware
   Suggested: Add optional chaining: user?.profile?.x
```

---

## Implementation Priority

| Feature | Effort | Value | Priority |
|---------|--------|-------|----------|
| File watch → task correlation | Medium | High | P0 |
| Git hook integration | Medium | High | P0 |
| Dependency chain awareness | Low | High | P1 |
| Deadline/time awareness | Low | Medium | P1 |
| Build/test awareness | Medium | Medium | P2 |
| Cross-session memory | High | High | P2 |
| Code review readiness | Medium | Medium | P3 |
| Error pattern detection | High | Low | P3 |

---

## Plugin Architecture for New Features

Add new hooks to `.opencode/plugins/exarp-go.ts`:

```typescript
// File watch correlation
tool: {
  execute: {
    after: async (ctx) => {
      if (isEditTool(ctx.tool)) {
        const files = getEditedFiles(ctx);
        const task = findTaskByFiles(files);
        if (task) showToast(`📁 Edited ${files[0]} → ${task.id}`);
      }
    }
  }
}

// Git branch detection
shell: {
  execute: {
    after: async (ctx) => {
      if (ctx.command.includes('checkout') || ctx.command.includes('switch')) {
        const branch = getCurrentBranch();
        const taskId = extractTaskId(branch);
        if (taskId) autoClaimTask(taskId);
      }
    }
  }
}
```
