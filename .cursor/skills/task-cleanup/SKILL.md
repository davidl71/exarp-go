---
name: task-cleanup
description: Bulk remove one-off or performance tasks from Todo2. Use when the user asks to remove one-off tasks, performance tasks, or says they "reappeared." Prefer task_workflow batch delete (task_ids) for speed.
---

# Task cleanup (one-off / performance)

Use this skill when the user wants to **remove one-off or performance tasks** from the database and Todo2, or when those tasks have "reappeared" (e.g. after a restore or sync).

## Fast path: batch delete

Use **one** `task_workflow` MCP call with `action=delete` and **`task_ids`** (comma-separated) instead of looping.

- **One-off tasks** (if they reappear):  
  `task_ids`: `T-1769716109198,T-1768318259581,T-321,T-401,T-402,T-403`

- **Performance tasks** (if they reappear):  
  `task_ids`: `T-1768325753706,T-1768325747897,T-1768325741814,T-1768325734895,T-1768325728818,T-1768325715211,T-247,T-177,T-226,T-81,T-166,T-359,T-222,T-200,T-1768317407961`

Example (MCP):

```json
{"action":"delete","task_ids":"T-321,T-401,T-402,T-403,T-1768325753706,T-1768325747897,T-1768325741814"}
```

## Fallback: CLI

If MCP is not available, use the loop from `docs/TASKS_AS_CI_AND_AUTOMATION.md` §3.4 / §3.5, or:

```bash
./bin/exarp-go -tool task_workflow -args '{"action":"delete","task_ids":"T-321,T-401,T-402"}'
```

## Reference

- ID lists and rationale: `docs/TASKS_AS_CI_AND_AUTOMATION.md` (§3.4 One-off, §3.5 Performance).
- task-workflow skill: batch delete is documented there (`task_ids` for delete).
