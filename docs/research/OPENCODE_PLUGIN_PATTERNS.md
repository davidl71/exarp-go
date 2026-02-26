# OpenCode Plugin Ecosystem — Pattern Analysis

**Date:** 2026-02-26
**Purpose:** Catalog architectural patterns from 16 OpenCode plugins/projects and 1 distributed ML survey, identify actionable improvements for the exarp-go OpenCode plugin and core system.

---

## Projects Analyzed

### OpenCode Plugins (True Plugins)

| Project | Stars | Type | Key Innovation |
|---------|-------|------|----------------|
| [subtask2](https://github.com/spoons-and-mirrors/subtask2) | — | Plugin | Command chaining, loops, parallel execution, `$RESULT[name]` capture |
| [plannotator](https://github.com/backnotprop/plannotator) | — | Plugin | Interactive plan review UI, `primary_tools`, agent detection, `session.prompt()` |
| [opencode-supermemory](https://github.com/supermemoryai/opencode-supermemory) | 694 | Plugin | Persistent semantic memory, `synthetic: true`, preemptive compaction at 80% token limit |
| [opencode-skillful](https://github.com/zenobi-us/opencode-skillful) | — | Plugin | Skills system, `noReply: true` silent injection, `small_model` metadata generation |
| [opencode-scheduler](https://github.com/different-ai/opencode-scheduler) | — | Plugin | Embedded Perl supervisor, `OPENCODE_PERMISSION` deny for non-interactive, agent filtering |
| [micode](https://github.com/vtemian/micode) | — | Plugin | Tool output appending, dynamic context injection |
| [opencode-background-agents](https://github.com/kdcokenny/opencode-background-agents) | — | Plugin | Parallel sub-agent spawning, batched notifications, `client.app.log()` |
| [opencode-notify](https://github.com/kdcokenny/opencode-notify) | — | Plugin | OS notifications (macOS/Linux/Windows), `session.error` + `permission.updated` events |
| [opencode-workspace](https://github.com/kdcokenny/opencode-workspace) | — | Plugin | Bundle/meta-plugin, `tool.execute.after` output appending, agent-specific system prompts |
| [opencode-worktree](https://github.com/kdcokenny/opencode-worktree) | — | Plugin | Tmux mutex, self-cleaning temp scripts, cross-platform terminal detection (37+ terminals), Zod env validation |

### OpenCode Platforms (Not Plugins)

| Project | Stars | Type | Key Innovation |
|---------|-------|------|----------------|
| [openchamber](https://github.com/btriapitsyn/openchamber) | 1.1k | Web/Desktop GUI | SSE event-driven cache, AI-powered session summarization, Radix UI |
| [openwork](https://github.com/different-ai/openwork) | 10.5k | Full platform | Scoped token RBAC, Promise-based approval service, JSONL audit, OS-native scheduling |
| [ocx](https://github.com/kdcokenny/ocx) | 313 | Package manager | Atomic file ops (`process.pid`), `fetchWithCache` request dedup, registry-based distribution, profiles |
| [agentic](https://github.com/Cluster444/agentic) | 352 | Workflow CLI | Agent templates with model substitution, SHA-256 content comparison, upward directory search, phased workflow |
| [OpenAgentsControl](https://github.com/darrenhinde/OpenAgentsControl) | 2.2k | Context framework | ContextScout sub-agent, MVI (Minimal Viable Information), pattern-first development, interactive installer |

### Cross-Domain

| Project | Stars | Type | Key Innovation |
|---------|-------|------|----------------|
| [exo](https://github.com/exo-explore/exo) | 41.8k | Distributed inference | `rustworkx` topology graph, leader election (gossipsub), topic-based pub/sub, placement algorithm, typed channels, event-sourced state |
| KDNuggets survey (2025) | — | Article | PyTorch DDP, TensorFlow Strategy, Ray Train/Tune/Serve, Spark MLlib, Dask task graph |

---

## Top Patterns — Prioritized for exarp-go Adoption

### Priority 1: Immediate Plugin Improvements

#### 1.1 Sub-agent Filtering
**Source:** Plannotator, Scheduler, Workspace
**Pattern:** Skip task injection for non-primary agents to avoid noise in sub-agent contexts.

```typescript
"experimental.chat.system.transform": async (input, output) => {
  const agent = input.agent;
  if (agent?.mode === "subagent" || agent?.mode === "title") return;
  // ... inject task context only for primary agent
}
```

**Impact:** Reduces token waste when OpenCode spawns sub-agents for parallel work.

#### 1.2 `synthetic: true` on Injected Parts
**Source:** Supermemory
**Pattern:** Mark plugin-generated message parts so OpenCode can distinguish them from user input.

```typescript
output.parts.unshift({
  type: "text",
  text: `[exarp-go tasks: ${formatTasksCompact(cache.tasks)}]`,
  synthetic: true,
});
```

**Impact:** Prevents OpenCode from treating injected task context as user intent.

#### 1.3 `noReply: true` for Silent Injection
**Source:** Skillful, Background Agents
**Pattern:** Inject persistent context without triggering an agent response cycle.

**Impact:** Useful for fire-and-forget status messages (e.g., "task T-xxx marked Done").

#### 1.4 `tool.execute.after` for Task Reminders
**Source:** Workspace, micode
**Pattern:** Append task-relevant reminders to tool output, keeping tasks visible without system prompt bloat.

```typescript
"tool.execute.after": async (input, output) => {
  if (inProgressTasks.length > 0) {
    output.result += `\n\n[Reminder: ${inProgressTasks.length} tasks in progress]`;
  }
}
```

**Impact:** Keeps task awareness in the agent's working memory during tool use.

#### 1.5 `primary_tools` Configuration
**Source:** Plannotator
**Pattern:** Restrict `exarp_update_task` to primary agent only, preventing sub-agents from accidentally changing task status.

```typescript
config.experimental = config.experimental ?? {};
config.experimental.primary_tools = [
  ...(config.experimental.primary_tools || []),
  "exarp_update_task",
];
```

**Impact:** Prevents accidental task status changes from parallel sub-agents.

### Priority 2: Near-Term Enhancements

#### 2.1 Preemptive Compaction Awareness
**Source:** Supermemory
**Pattern:** Monitor token usage via `message.updated` events. When context approaches model limit (~80%), proactively summarize tasks and inject condensed context.

**Impact:** Prevents task context from being lost during context window overflow.

#### 2.2 `session.error` Handling
**Source:** Notify
**Pattern:** Show toast or desktop notification when a session hits an error.

```typescript
if (event.type === "session.error") {
  await showToast(client, "Session error — check results", "error");
}
```

**Impact:** Better error visibility for long-running sessions.

#### 2.3 Agent-Specific System Prompts
**Source:** Workspace
**Pattern:** Inject different task context depending on the agent mode (build vs plan vs review).

**Impact:** More relevant context per agent role.

#### 2.4 `session.prompt()` for Programmatic Messaging
**Source:** Plannotator
**Pattern:** Use `client.session.prompt()` to send messages back into the session, with agent targeting.

**Impact:** Enables automated task completion summaries or follow-up prompts.

### Priority 3: Architectural Patterns for exarp-go Core

#### 3.1 Promise-Based Approval Service
**Source:** OpenWork
**Pattern:** In-memory Promise queue for human-in-the-loop approvals with timeout.

```typescript
class ApprovalService {
  async requestApproval(input): Promise<ApprovalResult> {
    return new Promise((resolve) => {
      const timeout = setTimeout(() => resolve({ allowed: false }), ms);
      this.pending.set(id, { resolve, timeout });
    });
  }
}
```

**Relevance:** Directly applicable to exarp-go's task Review workflow gates. Could make `Review → Done` transitions await actual human input instead of polling.

#### 3.2 JSONL Audit Trail
**Source:** OpenWork
**Pattern:** Append-only JSONL log for significant actions (task creates, status changes, lock acquisitions).

**Relevance:** Todo2 tracks task changes but lacks a general audit log. JSONL is schema-free, grep-friendly, and zero-migration.

#### 3.3 Debounced Event-Driven Cache
**Source:** OpenWork, OpenChamber
**Pattern:** Cursor-based reload events with debouncing instead of TTL-based cache invalidation.

**Relevance:** Our plugin uses 30s TTL cache. Cursor + debounce is more precise — only sends "changed" when something actually changed, and clients fetch only what's new.

#### 3.4 Topology-Aware Task Scheduling
**Source:** exo
**Pattern:** Graph-based placement algorithm with progressive constraint filtering (memory, sharding, RDMA preference, leaf nodes).

**Relevance:** Conceptual parallel to exarp-go's wave-based execution plan. Could inform smarter task dependency resolution — e.g., placing tasks on agents with the right capabilities/context.

#### 3.5 Event-Sourced State Transitions
**Source:** exo
**Pattern:** Compute state changes as diffs and emit typed events (`InstanceCreated`, `TaskStatusUpdated`).

**Relevance:** Currently task state changes are direct mutations. Event sourcing would enable undo, replay, and audit trail natively.

#### 3.6 Typed Pub/Sub for Agent Communication
**Source:** exo
**Pattern:** Topic-based router with typed messages and local/network routing.

**Relevance:** If exarp-go ever coordinates multiple agents, a typed pub/sub beats ad-hoc message passing.

### Priority 4: Reference Patterns (Future Use)

| Pattern | Source | When Useful |
|---------|--------|-------------|
| Scoped token RBAC | OpenWork | Multi-user or remote access scenarios |
| OS-native job scheduling (launchd/systemd) | OpenWork | Persistent scheduled health checks without Redis |
| OTel-compatible structured logging | OpenWork | JSON log output for observability |
| Registry-based component distribution | OCX | If exarp-go plugins become distributable |
| Agent template system with model substitution | Agentic | Multi-model task execution |
| Interactive bash installer (platform detection) | OAC | Cross-platform distribution |
| Cross-platform terminal detection (37+) | Worktree | Terminal-aware notifications |
| Dynamic thinking budget (`chat.params` hook) | Scheduler | Adjusting LLM effort per task complexity |

---

## Hook Usage Across Plugins (Compatibility Matrix)

| Hook | subtask2 | plannotator | supermemory | skillful | scheduler | micode | BG agents | notify | workspace |
|------|----------|-------------|-------------|----------|-----------|--------|-----------|--------|-----------|
| `shell.env` | — | — | — | — | — | — | — | — | ✓ |
| `event` | — | ✓ | ✓ | — | ✓ | — | ✓ | ✓ | ✓ |
| `chat.message` | — | — | ✓ | ✓ | — | — | ✓ | — | ✓ |
| `chat.params` | — | — | — | — | ✓ | — | — | — | — |
| `system.transform` | — | ✓ | — | ✓ | ✓ | ✓ | ✓ | — | ✓ |
| `session.compacting` | — | — | ✓ | — | ✓ | — | — | — | ✓ |
| `tool` | — | ✓ | ✓ | — | — | — | ✓ | — | ✓ |
| `tool.execute.before` | — | — | — | — | — | — | — | — | ✓ |
| `tool.execute.after` | — | — | — | — | — | ✓ | — | — | ✓ |
| `config` | ✓ | ✓ | — | ✓ | ✓ | — | ✓ | ✓ | ✓ |

**Our plugin currently uses:** `shell.env`, `event`, `chat.message`, `system.transform`, `session.compacting`, `tool`, `config`.
**Missing hooks worth adding:** `tool.execute.after`, `chat.params` (for thinking budget).

---

## Distributed ML Framework Relevance (KDNuggets Survey)

| Framework | exarp-go Parallel |
|-----------|------------------|
| **Dask task graph** | Wave-based execution plan (`task_analysis` parallelization) |
| **TorchElastic fault tolerance** | Agent locking with lease renewal + stale lock cleanup |
| **Ray Train/Tune/Serve pipeline** | Sprint automation (`sprint_start` → `sprint_end` → `pre_sprint`) |
| **Spark in-memory processing** | `TaskCache` with TTL-based invalidation |
| **PyTorch DDP gradient sync** | Multi-agent task synchronization via Todo2 DB |

---

## References

- exarp-go plugin: `.opencode/plugins/exarp-go.ts`
- Plugin docs: `docs/OPENCODE_INTEGRATION.md`
- OpenCode plugin API: https://opencode.ai/docs/plugins/
- exo distributed inference: https://github.com/exo-explore/exo
- KDNuggets survey: https://www.kdnuggets.com/top-5-frameworks-for-distributed-machine-learning
