# Auto-Progression Setup - Phase 2 Task Automation

**Created:** 2026-01-10  
**Purpose:** Automatically start Phase 2 tasks when Phase 1 research tasks complete

## Overview

This system monitors Phase 1 research task completion and automatically starts Phase 2 tasks when their dependencies are satisfied.

## Phase 1 Research Tasks (Currently In Progress)

The following 8 research tasks are being executed in parallel:

1. **T-0**: Research task data structure and design task resources API (2.9h, high)
2. **T-5**: Research common Go code patterns between exarp-go and devwisdom-go (2.9h, high)
3. **T-6**: Analyze pros and cons of extracting common code to shared library (2.9h, high)
4. **T-7**: Recommend shared library strategy with implementation approach (2.9h, high)
5. **T-8**: Analyze MCP protocol implementations for unified approach (2.9h, high)
6. **T-16**: (2.9h, high)
7. **T-17**: (2.9h, high)
8. **T-18**: (2.9h, high)

**Total Phase 1 Time:** 2.9 hours (parallel execution of 8 tasks)

## Auto-Progression Script

**Location:** `scripts/auto_progress_phase2.py`

### Usage

**Manual Check:**
```bash
cd /Users/davidl/Projects/exarp-go
python3 scripts/auto_progress_phase2.py
```

**Automated Monitoring (recommended):**
```bash
# Check every 5 minutes
cd /Users/davidl/Projects/exarp-go
watch -n 300 python3 scripts/auto_progress_phase2.py

# Or add to crontab (check every 10 minutes)
*/10 * * * * cd /Users/davidl/Projects/exarp-go && python3 scripts/auto_progress_phase2.py >> logs/phase2_progression.log 2>&1
```

### How It Works

1. **Monitors Phase 1 Tasks**: Checks completion status of 8 Phase 1 research tasks
2. **Identifies Ready Tasks**: Finds Phase 2 tasks that:
   - Depend on Phase 1 research tasks (T-0, T-5, T-6, T-7, T-8, T-16, T-17, T-18)
   - Have ALL dependencies satisfied (Phase 1 + any other dependencies)
   - Are not already completed or in progress
3. **Starts Tasks Automatically**: 
   - Sorts by priority (high → medium → low)
   - Starts up to 8 tasks in parallel
   - Updates task status to "In Progress"

### Output Format

**When Phase 2 tasks are ready:**
```json
{
  "started": 6,
  "tasks": [
    {"id": "T-NaN", "name": "Document MCP server configuration..."},
    {"id": "T-20", "name": "..."},
    ...
  ],
  "message": "Started 6 Phase 2 tasks"
}
```

**When no tasks ready yet:**
```json
{
  "started": 0,
  "message": "No Phase 2 tasks ready yet"
}
```

## Phase 2 Task Identification

Phase 2 tasks are identified by:
- **Dependency on Phase 1**: Task depends on one or more Phase 1 research tasks
- **All Dependencies Satisfied**: ALL task dependencies are completed (Phase 1 + any other dependencies)
- **Status**: Task is in "Todo" status (not "In Progress" or "Done")

### Task Categories

**Category 1: Phase 1 Only Dependencies**
- Tasks that ONLY depend on Phase 1 research tasks
- Will automatically start when Phase 1 completes
- **Example:** T-NaN (depends only on T-0)

**Category 2: Phase 1 + Other Dependencies**
- Tasks that depend on Phase 1 research AND other tasks
- Will start when Phase 1 completes AND other dependencies are satisfied
- **Example:** T-23 (depends on T-16, T-17, T-19)

## Current Status

**Phase 1:** 8 research tasks in progress (0/8 completed)  
**Phase 2 Ready:** 6 tasks identified that depend on Phase 1  
**Auto-Start When Phase 1 Completes:** Tasks will automatically start when dependencies are satisfied

### Tasks Waiting for Phase 1

1. **T-NaN**: Document MCP server configuration analysis (depends on T-0)
2. **T-20**: (depends on T-16)
3. **T-21**: (depends on T-16, T-17)
4. **T-22**: (depends on T-16, T-18)
5. **T-23**: (depends on T-16, T-17, T-19)
6. **T-24**: (depends on T-16, T-17, T-18)

## Execution Strategy

- **Parallel Execution**: Up to 8 Phase 2 tasks started simultaneously
- **Priority-Based**: Higher priority tasks started first
- **Dependency-Aware**: Only starts tasks when ALL dependencies are satisfied
- **Automatic Progression**: No manual intervention needed

## Monitoring

### Check Current Status

```bash
# Check Phase 1 completion
cd /Users/davidl/Projects/exarp-go
python3 -c "
from project_management_automation.utils.todo2_mcp_client import Todo2Client
from project_management_automation.utils.todo2_utils import is_completed_status
client = Todo2Client()
todos = client.list_todos()
phase1_ids = ['T-0', 'T-5', 'T-6', 'T-7', 'T-8', 'T-16', 'T-17', 'T-18']
all_tasks = {t.get('id'): t for t in todos.get('todos', [])}
completed = [tid for tid in phase1_ids if tid in all_tasks and is_completed_status(all_tasks[tid].get('status', 'Todo'))]
print(f'Phase 1: {len(completed)}/{len(phase1_ids)} completed')
for tid in phase1_ids:
    if tid in all_tasks:
        status = all_tasks[tid].get('status', 'Todo')
        print(f'  {tid}: {status}')
"

# Check Phase 2 readiness
python3 scripts/auto_progress_phase2.py
```

## Integration with Parallel Execution

This auto-progression system works with the parallel execution plan:

1. **Phase 1** (Research): 8 research tasks running in parallel → 2.9 hours
2. **Phase 2** (Implementation): Automatically starts when Phase 1 completes (6 tasks identified)
3. **Subsequent Phases**: Continue automatically as dependencies are satisfied

## Benefits

- ✅ **No Manual Monitoring**: Automatically progresses through phases
- ✅ **Dependency-Aware**: Only starts tasks when ready
- ✅ **Parallel Execution**: Maximizes parallelism for efficiency
- ✅ **Priority-Based**: Ensures high-priority tasks start first
- ✅ **Research-First**: Prioritizes research tasks to inform implementation

## Next Steps

1. **Complete Phase 1 Research**: Work on the 8 research tasks in parallel
2. **Auto-Progress**: Script will automatically start Phase 2 tasks when Phase 1 completes
3. **Continue Progress**: System will automatically progress through all phases

## Troubleshooting

**Phase 2 tasks not starting?**
- Check Phase 1 completion: Run `python3 scripts/auto_progress_phase2.py`
- Verify dependencies: Some Phase 2 tasks may have additional dependencies beyond Phase 1
- Check task status: Ensure tasks are in "Todo" status, not blocked

**Want to manually start Phase 2?**
```python
from project_management_automation.utils.todo2_mcp_client import Todo2Client, update_todos_mcp
from project_management_automation.utils.project_root import find_project_root

client = Todo2Client()
project_root = find_project_root()

# Get ready tasks
ready = client.list_todos(ready_only=True)

# Start specific tasks
updates = [{'id': 'T-NaN', 'status': 'In Progress'}]
update_todos_mcp(updates, project_root=project_root)
```

## Automation Setup

To enable automatic monitoring, add to crontab:

```bash
# Edit crontab
crontab -e

# Add line (check every 10 minutes)
*/10 * * * * cd /Users/davidl/Projects/exarp-go && python3 scripts/auto_progress_phase2.py >> logs/phase2_progression.log 2>&1
```

Or use a file watcher to check when tasks are updated:

```bash
# Install fswatch (if not installed)
brew install fswatch

# Watch Todo2 state file and auto-progress
fswatch .todo2/state.todo2.json | while read; do
    cd /Users/davidl/Projects/exarp-go
    python3 scripts/auto_progress_phase2.py
done
```
