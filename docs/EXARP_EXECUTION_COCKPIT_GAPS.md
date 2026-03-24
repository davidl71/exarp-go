## exarp-go Execution Cockpit Gaps

**Status:** Active research and modification note  
**Context:** Observed while using exarp-go to drive multi-step implementation work from Aether

### Summary

exarp-go is already strong at backlog governance:

- task creation and updates
- dependency analysis
- execution-plan generation
- reporting and session handoff

The main gap is not planning. The main gap is execution ergonomics during active coding work.

In practice, the tool is still weaker as an execution cockpit than as a backlog manager. The missing pieces are mostly around:

- active task ownership
- partial-progress tracking
- child-task ergonomics
- structured execution logs
- richer status semantics
- tighter linkage between task, code, verification, and outcome

Important nuance:

- exarp-go already has underlying execution-adjacent primitives such as task assignee/lease locking, `parent_id`, dependencies, and session assignee operations
- the gap is that these primitives do not yet add up to a cohesive, session-centered execution cockpit
- this note is therefore mostly about surfacing, unifying, and streamlining existing capabilities plus adding a few missing first-class records

### Real Usage Friction

These are the recurrent problems hit in real implementation sessions:

1. There is no strong session-level active-task model.
   exarp-go already has assignee / lease locking and can move a task to `In Progress`, but that still does not create a durable sense of "this is the task currently being executed in this session" with explicit session focus, start time, and execution context.

2. Partial completion is awkward.
   A coding session often completes one slice of a larger task, not the whole task. Today this gets pushed into free-form comments like `note` or `result`, but the model does not treat "slice done, parent still active" as first-class state.

3. Subtask creation is too manual for execution waves.
   During real work, a parent task often needs to be split into 2-5 concrete execution children with inherited context, dependencies, and tags. exarp-go already has `parent_id` and dependency support, but the workflow to create and wire these children is still not fast enough.

4. Execution logs are comment-shaped instead of structured.
   The natural unit of coding progress is something like:
   - files touched
   - commands run
   - compile/lint/test result
   - follow-up risk

   exarp-go stores this as comments, but does not yet make it a first-class structured run record.

5. Status semantics are too coarse for coding work.
   `Todo`, `In Progress`, `Review`, `Done`, `Blocked`, `Cancelled` are useful, but coding execution often wants narrower distinctions such as:
   - active
   - partially_done
   - implemented_not_verified
   - verified_waiting_cleanup
   - blocked_external

6. Dependency editing is still too expensive mid-flight.
   The dependency model is already good enough for planning and persistence, but it is not fluid enough for fast updates while tasks are being split, merged, or superseded during implementation.

7. Task-to-code linkage is weak.
   A task should be able to accumulate:
   - touched files
   - compile/lint/test evidence
   - related docs
   - commits or diffs

   exarp-go supports fragments of this, but not as one coherent execution record.

8. Parallel execution support is planner-heavy, executor-light.
   exarp-go can identify waves and dependency order, but it does not yet feel like a native coordinator for multiple live workers executing sibling slices.

### Highest-Value Modifications

These are the best next changes in descending order of practical value.

### 1. Active Task Claim

Add a first-class `active task` concept at the session level, ideally layered on top of the existing assignee / lease model instead of replacing it.

Required behavior:

- mark one task as the session's active execution target
- record start timestamp
- record assignee / actor
- expose this in session prime, task list, and suggested-next views
- allow `claim`, `release`, and `switch`

Why this matters:

- reduces drift between chat intent and task state
- makes "what am I working on right now?" explicit
- improves handoff quality

Suggested shape:

- build on existing `session action=assignee` / database claim-release semantics
- optionally add a clearer task-centric surface such as `task_workflow action=claim task_id=...`
- `session prime` returns `active_task`

### 2. Structured Execution Runs

Add a lightweight execution-run model attached to a task.

Required fields:

- `run_id`
- `task_id`
- `started_at`
- `ended_at`
- `actor`
- `status`
- `summary`
- `files_touched`
- `commands_run`
- `verification`
- `notes`

Why this matters:

- turns free-form result comments into durable evidence
- enables better summaries, audits, and review preparation
- supports automatic handoff generation

This does not need to replace comments. It should complement them.

### 3. Partial Completion / Slice Tracking

Add first-class support for partial progress on a parent task.

Required behavior:

- record a completed slice without closing the parent
- optionally attach the slice to a child task
- expose this in task views and planning output

Suggested shape:

- `task_workflow action=add_progress task_id=...`
- fields: `summary`, `files`, `verification`, `remaining_work`

### 4. Fast Child Task Creation

Add a helper for splitting a task into execution children.

Required behavior:

- create several child tasks in one call
- inherit parent tags and planning context
- optionally add sequential or parallel dependencies automatically

Suggested shape:

- `task_workflow action=split task_id=... children=[...]`

Implementation note:

- this should create normal child tasks using existing `parent_id`, tags, and dependencies rather than introducing a second hierarchy model

This should be optimized for implementation waves, not just epic planning.

### 5. Better Status Semantics for Execution

Keep the current top-level statuses for compatibility, but add execution-phase metadata.

Suggested execution states:

- `active`
- `partial`
- `implemented`
- `verified`
- `blocked_external`

This can be stored separately from Todo2 status if needed.

### 6. Structured Verification Evidence

Add first-class verification recording.

Required behavior:

- attach commands and results to a task or execution run
- distinguish lint / compile / test / manual verification
- make this queryable

Suggested shape:

- `task_workflow action=verify task_id=...`
- payload includes `kind`, `command`, `result`, `details`

### 7. Better Dependency Mutation UX

Add more fluid dependency operations for execution-time graph changes.

Needed operations:

- append dependency
- remove dependency
- replace dependencies
- mark dependency superseded

This should avoid forcing direct DB edits for normal graph maintenance.

This is primarily a workflow/surface improvement, not a data-model redesign.

### 8. Parallel Worker Coordination

Add explicit parallel execution support at the task layer.

Useful capabilities:

- assign multiple child tasks to one parent execution wave
- mark worker ownership
- record which worker touched which files
- aggregate child completion into parent progress

This would let exarp-go move from "planner that suggests parallelism" to "coordinator that tracks it".

### Recommended Implementation Order

1. active task claim
2. structured execution runs
3. partial progress / slice tracking
4. fast child task creation
5. structured verification evidence
6. dependency mutation UX
7. parallel worker coordination

### Proposed Initial Scope

The smallest meaningful first batch is:

- expose a clearer active-task claim/release flow on top of existing assignee locking
- add structured execution runs
- expose both in `task_workflow`, `session`, and task list/show output

That would solve the biggest real-world gap immediately without redesigning the whole task system.

### Non-Goals

These are not required for the first batch:

- full durable command bus
- full UI dashboard
- replacement of existing comments/history
- replacing Todo2 status model

### Bottom Line

exarp-go already plans well enough.

The next improvement should be execution-state fidelity:

- what is actively being worked on
- what slice just completed
- what evidence exists
- what remains
- how parallel workers map back to task state

That is the shortest path from "good backlog manager" to "useful execution cockpit".
