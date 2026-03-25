# Ambient Features Research & Implementation Plan

**Date:** 2026-03-25  
**Status:** Ready for Implementation  
**Plugin Location:** `.opencode/plugins/exarp-go.ts` (557 lines)

---

## Executive Summary

The exarp-go OpenCode plugin provides **ambient task awareness** — the AI agent always knows your tasks, progress, and project state without explicit prompts. This document researches current capabilities and proposes new features to deepen ambient integration.

---

## Current Architecture Analysis

### Plugin Hook System

The OpenCode plugin uses a hook-based architecture:

```typescript
export const ExarpGoPlugin: Plugin = async ({ $, client, directory }) => {
  return {
    "shell.env": async (_input, output) => { ... },
    "chat.message": async (input, output) => { ... },
    "experimental.chat.system.transform": async (input, output) => { ... },
    "tool.execute.after": async (_input, output) => { ... },
    event: async ({ event }) => { ... },
    tool: { ... },  // Native tools
    async config(config) { ... },
  };
};
```

### Available Hooks (Current)

| Hook | Type | Purpose |
|------|------|---------|
| `shell.env` | Env | Inject `PROJECT_ROOT` into all shell commands |
| `chat.message` | Transform | Inject task summary into first message |
| `chat.system.transform` | Transform | Inject full task state into system prompt |
| `tool.execute.after` | Lifecycle | Append reminders after any tool use |
| `experimental.session.compacting` | Lifecycle | Preserve task context across compaction |
| `event` | Event | Handle session/todo/TUI events |
| `config` | Config | Register slash commands, primary tools |

### Event Types Currently Handled

| Event | Action |
|-------|--------|
| `session.created` | Show toast with task counts |
| `session.idle` | macOS desktop notification |
| `session.error` | Error toast |
| `todo.updated` | Invalidate cache |
| `tui.command.execute` | Invalidate on /tasks, /prime, etc. |
| `tui.prompt.append` | Auto-expand task IDs |
| `tui.toast.show` | Invalidate on task-related toasts |

### Data Flow

```
User → OpenCode → Plugin Hook → exarp-go CLI → .todo2/state.todo2.json
                                    ↓
                            Parse JSON response
                                    ↓
                            Update cache / inject context
                                    ↓
                            LLM receives ambient context
```

---

## Current Ambient Features (8)

### 1. System Prompt Injection
- **Hook:** `experimental.chat.system.transform`
- **Content:** Full task list with status, priority, tags
- **Frequency:** Every message (cached 30s)
- **Skipped:** Sub-agents, title generation

### 2. First-Message Injection
- **Hook:** `chat.message`
- **Content:** Compact task summary
- **Marked:** `synthetic: true`

### 3. Tool Output Reminders
- **Hook:** `tool.execute.after`
- **Content:** In-progress task reminders
- **Example:** `[In Progress (2): T-1: Fix auth; T-2: Add tests]`

### 4. Toast Notifications
- **Hook:** `session.created`, `todo.updated`
- **Content:** Task count summaries
- **API:** `client.tui.showToast()`

### 5. Compaction Context
- **Hook:** `experimental.session.compacting`
- **Content:** Full task state preserved
- **Purpose:** Survive context window resets

### 6. Prompt Auto-Expand
- **Hook:** `tui.prompt.append`
- **Trigger:** User types `T-123` pattern
- **Action:** Append task details to prompt

### 7. Cache Invalidation
- **Events:** `todo.updated`, `tui.command.execute`, `tui.toast.show`
- **Purpose:** Keep task data fresh

### 8. Plugin Tools (Fast Path)
- **Tools:** `exarp_tasks`, `exarp_update_task`, `exarp_prime`, `exarp_config`, `exarp_followup`
- **Benefit:** 5x faster than MCP round-trip

---

## Proposed Ambient Features

### P0 - High Priority (Implement First)

#### F1: File Watch → Task Correlation

**Concept:** Detect file edits and correlate to active tasks

**Implementation Strategy:**
1. Hook into `tool.execute.after` for edit-capable tools
2. Parse `input.tool` for: `edit`, `write`, `str_replace`
3. Extract file path from `input.args`
4. Match against task ownership metadata
5. Show toast + inject context

**Code Sketch:**
```typescript
"tool.execute.after": async (input, output) => {
  const editTools = ["edit", "write", "str_replace", "sed"];
  if (editTools.includes(input.tool)) {
    const filePath = extractFilePath(input.args);
    const task = await findTaskByFile(filePath);
    if (task) {
      showToast(client, `📁 ${filePath} → ${task.id}`, "info");
      // Inject correlation into output
      output.result += `\n\n[File mapped to: ${task.id} - ${task.content}]`;
    }
  }
}
```

**New exarp-go Action Needed:**
```bash
# Find tasks that own a specific file
exarp-go -tool task_workflow -args '{"action":"list","owned_file":"path/to/file.go"}'
```

**OpenCode Hook:** `tool.execute.after` (extend existing)

---

#### F2: Git Hook Integration

**Concept:** Auto-detect branch patterns and correlate to tasks

**Implementation Strategy:**
1. Hook into `shell.execute.after` for git commands
2. Detect branch switches with `T-\d+` pattern
3. Auto-claim matching task
4. Show confirmation toast

**Code Sketch:**
```typescript
"shell.execute.after": async (input, output) => {
  const cmd = input.command;
  if (cmd.includes("checkout") || cmd.includes("switch")) {
    const branch = await getCurrentBranch($);
    const taskId = branch.match(/T-(\d+)/)?.[0];
    if (taskId) {
      await autoClaimTask(taskId);
      showToast(client, `🌿 Branch ${branch} → claimed ${taskId}`, "success");
    }
  }
}
```

**New exarp-go Action Needed:**
```bash
# Auto-claim task from branch name
exarp-go -tool task_workflow -args '{"action":"claim","task_id":"T-123","reason":"branch:T-123-fix-auth"}'
```

**OpenCode Hook:** `shell.execute.after` (new)

---

### P1 - Medium Priority

#### F3: Dependency Chain Awareness

**Concept:** Warn when starting task with incomplete dependencies

**Implementation Strategy:**
1. In `chat.system.transform`, analyze suggested tasks
2. Check if dependencies are complete
3. Add warning to system context

**Code Sketch:**
```typescript
"experimental.chat.system.transform": async (input, output) => {
  const cache = await refreshCache($);
  const warnings = [];
  for (const task of cache.tasks.filter(t => t.status === "In Progress")) {
    const blocked = await checkDependencies(task.id);
    if (blocked.length > 0) {
      warnings.push(`⚠️ ${task.id} blocked by: ${blocked.join(", ")}`);
    }
  }
  if (warnings.length > 0) {
    output.system.push(`## Dependency Warnings\n${warnings.join("\n")}`);
  }
}
```

**New exarp-go Action Needed:**
```bash
# Check task dependencies
exarp-go -tool task_analysis -args '{"action":"dependencies","task_id":"T-123"}'
```

---

#### F4: Deadline/Time Awareness

**Concept:** Alert about stale tasks, prep standup summaries

**Implementation Strategy:**
1. Calculate task age (created vs now)
2. Add stale warnings to system context
3. Generate standup summaries on request

**Code Sketch:**
```typescript
function calculateTaskAge(task: TaskSummary): number {
  // Parse task ID (timestamp-based) or use created_at if available
  return Date.now() - parseTaskTimestamp(task.id);
}

// In system.transform:
const stale = tasks.filter(t => calculateTaskAge(t) > 7 * 24 * 60 * 60 * 1000);
if (stale.length > 0) {
  output.system.push(`## Stale Tasks (>7 days)\n${formatTasks(stale)}`);
}
```

**No new exarp-go action needed** (can compute from existing data)

---

### P2 - Lower Priority

#### F5: Build/Test Result Awareness

**Hook:** `shell.execute.after`  
**Parse:** Exit codes, test output  
**Inject:** Pass/fail counts into next context

#### F6: Cross-Session Memory

**Storage:** `.exarp/memories/`  
**Hook:** `session.created` (load), task completion (save)  
**Inject:** Relevant memories when similar task appears

#### F7: Ownership Warnings in Prime

**Status:** ✅ IMPLEMENTED  
**Location:** `ownership_warnings` in session prime output

---

## Implementation Roadmap

### Phase 1: File Correlation (P0)
1. Add `list` action filter for `owned_file`
2. Extend plugin `tool.execute.after` hook
3. Test with edit tools

### Phase 2: Git Integration (P0)
1. Add `claim` action to task_workflow
2. Extend plugin with `shell.execute.after`
3. Test branch → task correlation

### Phase 3: Dependency Awareness (P1)
1. Add dependency check to system context
2. Add warning formatting
3. Test with blocked tasks

### Phase 4: Time Awareness (P1)
1. Add stale task detection
2. Add standup summary generation
3. Test with aged tasks

---

## Plugin Extension Points

### Adding New Hooks

```typescript
// In the plugin return object:
return {
  // New hook: shell.execute.after
  "shell.execute.after": async (input, output) => {
    // input.command: the shell command
    // output.stdout: command output
    // output.exitCode: exit code
  },

  // New hook: file.edit.after (hypothetical)
  "file.edit.after": async (input, output) => {
    // input.path: file path
    // input.content: new content
  },
};
```

### Adding New Tools

```typescript
tool: {
  exarp_file_task: tool({
    description: "Find task associated with a file",
    args: {
      path: tool.schema.string().describe("File path to check"),
    },
    async execute(args) {
      const result = await runExarp($, "task_workflow", {
        action: "list",
        owned_file: args.path,
        output_format: "text",
      });
      return result || "No task found for this file.";
    },
  }),
}
```

---

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Too many toasts → annoyance | Rate limit, make configurable |
| False file correlations | Use exact path matching, confidence threshold |
| Git hook race conditions | Debounce, handle branch switches |
| Performance impact | Cache aggressively, async operations |
| Plugin complexity | Modular functions, clear separation |

---

## Next Steps

1. **Review** this document with team
2. **Prioritize** features based on workflow needs
3. **Implement** P0 features (File Correlation, Git Integration)
4. **Test** in real workflows
5. **Iterate** based on feedback

---

## Appendix: Plugin API Reference

### Available Client APIs

```typescript
// Toast notifications
client.tui.showToast({ body: { message: string, variant: "success"|"error"|"info"|"warning" } });

// Append to prompt
client.tui.appendPrompt({ body: { text: string } });

// Logging
client.app.log({ body: { service: string, level: string, message: string } });
```

### Available Input Shapes

```typescript
// tool.execute.after
input: { tool: string, args: Record<string, any> }
output: { result: string }

// shell.execute.after (proposed)
input: { command: string, cwd: string }
output: { stdout: string, stderr: string, exitCode: number }

// chat.message
input: { sessionID: string, agent: { mode: string }, message: { role: string, parts: any[] } }
output: { parts: any[] }

// chat.system.transform
input: { sessionID: string, agent: { mode: string } }
output: { system: string[] }
```

---

**Document Status:** Ready for implementation  
**Last Updated:** 2026-03-25  
**Author:** exarp-go research (Sisyphus)
