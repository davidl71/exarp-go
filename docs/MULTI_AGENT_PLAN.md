# Multi-Agent Coordination Plan

**Date:** 2026-01-07  
**Status:** 📋 Planning  
**Purpose:** Define strategy for coordinating multiple AI agents working on the exarp-go project

---

## Executive Summary

This plan outlines how to coordinate multiple AI agents (AI assistants, specialized tools, MCP servers) to work efficiently on the Go SDK migration and ongoing development. The strategy leverages existing parallel research workflows, Todo2 task management, and MCP server infrastructure.

**Key Goals:**
- ✅ Efficient task distribution across agents
- ✅ Clear communication protocols
- ✅ Conflict resolution mechanisms
- ✅ Scalable architecture
- ✅ Monitoring and evaluation

---

## Current State Analysis

### Existing Infrastructure

**MCP Servers (4 configured):**
1. **advisor** - DevWisdom Go MCP Server (Crew Role: Advisor - wisdom quotes, advisors)
2. **coordinator** - Project Management Automation (Crew Role: Coordinator - 24 tools, 8 prompts, 6 resources)
3. **researcher** - Context7 MCP Server (Crew Role: Researcher - advanced context management and documentation retrieval)
4. **analyst** - Tractatus Thinking MCP Server (Crew Role: Analyst - structured reasoning and logical decomposition)

**Parallel Research Workflow:**
- ✅ CodeLlama (Ollama) or FM chain for code analysis
- ✅ Context7 for documentation
- ✅ Tractatus Thinking for logical reasoning
- ✅ Web search for latest information

**Task Management:**
- ✅ Todo2 system with 8 migration tasks
- ✅ Task dependencies defined
- ✅ Research workflow integrated

**Development Workflow:**
- ✅ Makefile with 30+ targets
- ✅ Auto-reload, auto-test, auto-coverage
- ✅ Continuous feedback loop

---

## Conflict Detection

Multi-agent conflict detection identifies when multiple agents may be working on conflicting work. Two detectors are implemented and surfaced in automation and session prime.

### 1. Task-overlap detection

**Goal:** Detect when two or more In Progress tasks have a dependency relationship (e.g. A blocks B, both In Progress = conflict).

**Implementation:** `internal/tools/conflict_detection.go` — `DetectTaskOverlapConflicts(tasks)`. Loads all tasks, filters to In Progress, and for each In Progress task checks if any of its dependencies is also In Progress. Returns `TaskOverlapConflict` (TaskID, DepTaskID, Reason).

**Integration:** Automation (daily, nightly, sprint) and session prime. Results include `conflicts.task_overlap`; session prime adds `conflict_hints` with messages like "Task overlap: T-X blocks T-Y; both In Progress".

### 2. File-level conflict detection

**Goal:** Detect when multiple In Progress tasks list the same file(s), suggesting concurrent edits.

**Implementation:** `DetectFileConflicts(tasks)` parses long_description for file paths (pattern: `- Update: path`, `- Create: path`, etc. from Files/Components). Builds file → task IDs map; for any file referenced by ≥2 In Progress tasks, emits a `FileConflict` (TaskIDs, Files).

**Integration:** Same as task-overlap: automation results `conflicts.file`, session prime `conflict_hints` (e.g. "File conflict: tasks T-1, T-2 share file(s): internal/foo.go").

### 3. Integration points

| Place | Behavior |
|-------|----------|
| **Automation** (`internal/tools/automation_native.go`) | Before running daily/nightly/sprint tasks, calls `DetectConflicts(ctx, projectRoot)`. If any overlaps or file conflicts, sets `results["conflicts"]` with `task_overlap` and `file` arrays. |
| **Session prime** (`internal/tools/session.go`) | After handoff check, calls `DetectConflicts`. If any, sets `result["conflict_hints"]` to a list of human-readable strings for the AI to consider. |

**Usage:** Automation can skip or flag conflicting tasks when assigning work; session prime surfaces hints so the AI is aware of existing conflicts at session start.

---

## Task dependency analysis

Task dependencies are used to order work and to group tasks into waves for parallel execution.

**Storage:** Todo2 tasks have optional `dependencies` (array of task IDs). Stored in SQLite (`.todo2/todo2.db`) or legacy JSON.

**Analysis tools:**
- **task_analysis** (exarp-go): `action=dependencies` (full graph), `action=dependencies_summary` (critical path + parallel groups), `action=execution_plan` (ordered backlog), `action=parallelization` (groups by dependency level).
- **report** (exarp-go): `action=plan` and `action=parallel_execution_plan` use `BacklogExecutionOrder()` to compute waves (same wave = no dependency between tasks = can run in parallel).

**Outputs:**
- **Critical path:** Longest dependency chain (e.g. T-214 → T-215 → T-221). Defines minimum sequential order.
- **Dependency levels:** Level 0 = no dependencies; level N = depends only on tasks at level &lt; N.
- **Waves:** Tasks in the same wave have no dependency ordering among them and can be run in parallel (e.g. Wave 0 in `.cursor/plans/parallel-execution-subagents.plan.md`).

**Usage:** Use `task_analysis` with `action=dependencies_summary` or `action=execution_plan` to inspect order; use the plan’s waves to run subagents in parallel per wave.

---

## Agent Types and Roles

### 1. Primary AI Assistant (Cursor AI)

**Role:** Project coordinator and primary implementer  
**Responsibilities:**
- Task creation and management (Todo2)
- Code implementation
- Architecture decisions
- Task distribution to specialized agents
- Result synthesis and integration
- Conflict resolution

**Capabilities:**
- Full codebase access
- File editing and creation
- Terminal command execution
- MCP tool invocation
- Web search

**Limitations:**
- Single-threaded execution (one task at a time)
- Context window limits
- No parallel code execution

### 2. Specialized Research Agents

#### CodeLlama (Ollama / local)
**Role:** Code analysis and architecture review  
**Responsibilities:**
- Code review and pattern analysis
- Architecture assessment
- Design pattern recommendations
- Code quality analysis

**Invocation:**
- Via **`text_generate`** (`provider=ollama` / `fm` / `auto`) or **`ollama`** MCP tool
- Local execution (Apple Silicon when using FM chain)

#### Context7 Agent
**Role:** Documentation and library research  
**Responsibilities:**
- Library documentation retrieval
- API reference lookup
- Code example discovery
- Version compatibility checking

**Invocation:**
- Via Context7 MCP server
- `resolve-library-id` → `query-docs`

#### Tractatus Thinking Agent
**Role:** Logical reasoning and problem decomposition  
**Responsibilities:**
- Concept analysis
- Problem decomposition
- Logical structure analysis
- Dependency identification

**Invocation:**
- Via Tractatus Thinking MCP server
- `tractatus_thinking` operation

#### Web Search Agent
**Role:** Latest information and best practices  
**Responsibilities:**
- Finding 2026 patterns and practices
- Community insights
- Security advisories
- Performance benchmarks

**Invocation:**
- Via web search tools
- Direct API calls

### 3. Task-Specific Agents (Future)

**Potential Specialized Agents:**
- **Testing Agent** - Automated test generation and execution
- **Documentation Agent** - Auto-generate docs from code
- **Security Agent** - Vulnerability scanning and remediation
- **Performance Agent** - Profiling and optimization
- **Migration Agent** - Automated code migration

---

## Communication Protocols

### 1. Task Distribution Protocol

**Pattern:** Hierarchical delegation with clear ownership

```
Primary AI Assistant
    ↓ (delegates research)
    ├── CodeLlama → Code analysis
    ├── Context7 → Documentation
    ├── Tractatus → Logical reasoning
    └── Web Search → Latest info
    ↓ (synthesizes results)
Primary AI Assistant → Implementation
```

**Task Assignment Rules:**
- **Code/Architecture** → CodeLlama
- **Library Docs** → Context7
- **Logical Analysis** → Tractatus
- **Latest Info** → Web Search
- **Implementation** → Primary AI Assistant

### 2. Result Aggregation Protocol

**Format:** Standardized research comment structure

```markdown
**MANDATORY RESEARCH COMPLETED** ✅

**Local Codebase Analysis:**
[code snippets and patterns]

**Internet Research (2026):**
[verified links and findings]

**CodeLlama Analysis:**
[code review and recommendations]

**Context7 Documentation:**
[library docs and examples]

**Tractatus Reasoning:**
[logical decomposition]

**Synthesis & Recommendation:**
[combined analysis and decision]
```

### 3. Conflict Resolution Protocol

**When agents disagree:**
1. **Document all perspectives** - Record each agent's recommendation
2. **Analyze source reliability** - Prioritize authoritative sources
3. **Check codebase patterns** - Prefer existing patterns
4. **Human escalation** - Mark task as "Review" for critical conflicts
5. **Consensus building** - Use majority opinion with rationale

**Priority Order:**
1. Codebase patterns (highest)
2. Official documentation (Context7)
3. CodeLlama analysis
4. Web search results
5. Tractatus reasoning

### 4. State Synchronization

**Shared State:**
- Todo2 tasks (via MCP tools)
- Codebase (via git)
- Documentation (via files)

**Synchronization Points:**
- Before task start: Read current state
- After task completion: Update state
- On conflicts: Resolve before proceeding

---

## Task Distribution Framework

### Phase 1: Research Distribution

**Strategy:** Parallel research execution

```python
# Pseudo-code for parallel research
async def distribute_research(task):
    results = await asyncio.gather(
        codellama_analyze(task.code_context),
        context7_get_docs(task.library_requirements),
        tractatus_decompose(task.problem_statement),
        web_search(task.latest_patterns)
    )
    return synthesize_results(results)
```

**Benefits:**
- ✅ Faster research (parallel vs sequential)
- ✅ Specialized analysis from each agent
- ✅ Comprehensive coverage

### Phase 2: Implementation Distribution

**Strategy:** Sequential with parallel validation

```
Primary AI → Implementation
    ↓
Parallel Validation:
    ├── CodeLlama → Code review
    ├── Testing → Auto-test
    └── Linting → Auto-lint
    ↓
Primary AI → Integration
```

**Current Limitation:**
- Primary AI is single-threaded
- Can't run multiple implementations simultaneously
- **Solution:** Use task dependencies to enable parallel work on independent tasks

### Phase 3: Independent Task Parallelization

**Strategy:** Identify independent tasks for parallel execution

**Example:**
```
T-3 (Batch 1 Tools) and T-8 (MCP Config) can run in parallel
    ↓
Agent 1 → T-3 (Tool migration)
Agent 2 → T-8 (Config setup)
    ↓
Both complete → Continue with dependent tasks
```

**Requirements:**
- Clear task dependencies
- No shared resource conflicts
- Independent code paths

---

## Workflow Orchestration

### 1. Task Lifecycle with Multi-Agent Support

```
[PLANNED]
    ↓
[RESEARCH PHASE - Parallel Agent Execution]
    ├── CodeLlama → Code analysis
    ├── Context7 → Documentation
    ├── Tractatus → Reasoning
    └── Web Search → Latest info
    ↓
[RESEARCHED] (with aggregated results)
    ↓
[IMPLEMENTATION PHASE - Primary AI]
    ↓
[VALIDATION PHASE - Parallel Checks]
    ├── CodeLlama → Review
    ├── Tests → Auto-run
    └── Lint → Auto-check
    ↓
[DONE]
```

### 2. Agent Coordination Workflow

**Step 1: Task Assessment**
- Primary AI analyzes task complexity
- Identifies research needs
- Determines agent requirements

**Step 2: Research Delegation**
- Primary AI delegates to specialized agents
- Agents execute in parallel
- Results aggregated

**Step 3: Implementation**
- Primary AI synthesizes research
- Implements solution
- Updates codebase

**Step 4: Validation**
- Parallel validation (tests, lint, review)
- Primary AI integrates feedback
- Task completion

### 3. Multi-Task Coordination

**Independent Tasks:**
- Can be worked on by different agents simultaneously
- Requires clear ownership
- Shared state synchronization

**Dependent Tasks:**
- Sequential execution required
- Dependency resolution before start
- Result propagation

---

## Infrastructure Requirements

### 1. Communication Infrastructure

**Current:**
- ✅ MCP protocol for tool invocation
- ✅ Todo2 for task management
- ✅ File system for state sharing

**Needed:**
- ⏳ Agent status tracking
- ⏳ Result aggregation system
- ✅ Conflict detection mechanism (task-overlap and file-level; see Conflict Detection section)

### 2. Monitoring and Evaluation

**Metrics to Track:**
- Task completion time
- Agent utilization
- Conflict frequency
- Research quality
- Implementation success rate

**Tools:**
- Todo2 task tracking
- Git commit analysis
- Test coverage reports
- Performance benchmarks

### 3. Scalability Considerations

**Current Limitations:**
- Single primary AI (Cursor)
- Sequential implementation
- Manual task distribution

**Future Enhancements:**
- Multiple primary AI instances
- Automated task distribution
- Agent pool management
- Load balancing

---

## Implementation Roadmap

### Phase 1: Foundation (Week 1)

**Goals:**
- ✅ Document current agent capabilities
- ✅ Define communication protocols
- ✅ Establish task distribution rules

**Deliverables:**
- This plan document
- Agent role definitions
- Communication protocol specs

**Status:** ✅ **Complete** (this document)

### Phase 2: Enhanced Research Workflow (Week 2)

**Goals:**
- Improve parallel research execution
- Add result aggregation automation
- Enhance conflict resolution

**Tasks:**
- [ ] Create research result aggregator
- [x] Implement conflict detection (task-overlap + file-level; automation + session prime)
- [ ] Add agent performance tracking

### Phase 3: Task Parallelization (Week 3-4)

**Goals:**
- Identify independent tasks
- Enable parallel task execution
- Implement state synchronization

**Tasks:**
- [x] Analyze task dependencies — documented in "Task dependency analysis" above; use task_analysis (dependencies_summary, execution_plan) and report (plan, parallel_execution_plan).
- [x] Create parallel execution framework — see [docs/PARALLEL_EXECUTION_FRAMEWORK.md](PARALLEL_EXECUTION_FRAMEWORK.md) (waves, runParallelTasks, task_analysis parallelization).
- [x] Implement shared state management — documented in PARALLEL_EXECUTION_FRAMEWORK.md § Shared state management (Todo2, handoffs, task locks).

### Phase 4: Advanced Coordination (Week 5+)

**Goals:**
- Multi-agent implementation
- Automated task distribution
- Performance optimization

**Tasks:**
- [ ] Design agent pool system
- [ ] Implement load balancing
- [ ] ~~Add monitoring dashboard~~ — **Future improvement** *(T-232 removed)*: implies a long-lived web server. Defer until HTTP/SSE server deployment. See `docs/STALE_LOCK_DETECTION_AND_HANDLING.md`.

---

## Best Practices

### 1. Agent Selection

**Choose the right agent for the task:**
- Code analysis → CodeLlama
- Documentation → Context7
- Logical reasoning → Tractatus
- Latest info → Web Search
- Implementation → Primary AI

### 2. Result Synthesis

**Always synthesize before implementing:**
- Combine all agent perspectives
- Resolve conflicts explicitly
- Document decision rationale
- Reference source agents

### 3. Conflict Resolution

**When agents disagree:**
- Don't ignore conflicts
- Analyze source reliability
- Prefer codebase patterns
- Escalate critical decisions

### 4. State Management

**Keep state synchronized:**
- Read state before starting
- Update state after completion
- Resolve conflicts immediately
- Document state changes

### 5. Performance Optimization

**Maximize parallel execution:**
- Delegate research in parallel
- Identify independent tasks
- Minimize sequential dependencies
- Use async operations

---

## Risk Mitigation

### 1. Agent Conflicts

**Risk:** Agents provide conflicting recommendations  
**Mitigation:**
- Document all perspectives
- Use priority-based resolution
- Escalate critical conflicts
- Maintain decision log

### 2. State Inconsistencies

**Risk:** Agents work with outdated state  
**Mitigation:**
- Always read state before starting
- Update state immediately after changes
- Use version control (git)
- Implement state validation

### 3. Resource Contention

**Risk:** Multiple agents access same resources  
**Mitigation:**
- Clear ownership per task
- Use task dependencies
- Implement locking mechanisms
- Coordinate file access

### 4. Agent Failures

**Risk:** Specialized agent fails or unavailable  
**Mitigation:**
- Graceful degradation (continue with available agents)
- Fallback to alternative agents
- Manual intervention option
- Error logging and monitoring

---

## Success Metrics

### Quantitative Metrics

- **Task Completion Time:** Reduce by 30-50% with parallel research
- **Agent Utilization:** 80%+ utilization of specialized agents
- **Conflict Rate:** <5% of tasks have unresolvable conflicts
- **Research Quality:** 90%+ of research results are actionable

### Qualitative Metrics

- **Code Quality:** Maintained or improved
- **Documentation:** Comprehensive and up-to-date
- **Developer Experience:** Smooth workflow, minimal friction
- **System Reliability:** Stable, predictable behavior

---

## Next Steps

### Immediate Actions (This Week)

1. ✅ **Complete this plan** - Document multi-agent strategy
2. ⏳ **Enhance research workflow** - Improve result aggregation
3. ⏳ **Test parallel execution** - Validate independent task execution
4. ⏳ **Monitor agent performance** - Track utilization and effectiveness

### Short-Term (Next 2 Weeks)

1. **Implement conflict resolution** - Automated conflict detection and resolution
2. **Create agent dashboard** - Visualize agent activity and performance
3. **Optimize task distribution** - Improve agent selection and delegation
4. **Document agent capabilities** - Comprehensive agent reference guide

### Long-Term (Next Month+)

1. **Multi-agent implementation** - Enable true parallel implementation
2. **Automated task distribution** - AI-driven task assignment
3. **Agent pool management** - Dynamic agent allocation
4. **Performance optimization** - Continuous improvement

---

## References

- [Parallel Research Workflow](./PARALLEL_RESEARCH_WORKFLOW.md) - Current parallel research implementation
- [Model-Assisted Workflow](./MODEL_ASSISTED_WORKFLOW.md) - Local LLMs (Ollama, FM chain)
- [Migration Status](./MIGRATION_STATUS.md) - Current project state
- [Todo2 Workflow](../.cursor/rules/todo2.mdc) - Task management system

---

**Status:** ✅ **Plan Complete** - Ready for implementation

**Last Updated:** 2026-01-07

