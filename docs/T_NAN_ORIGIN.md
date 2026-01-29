# Why T-NaN Exists

## Cause

**T-NaN** is an invalid task ID that appears when the **Todo2 Cursor/VS Code extension** (or any JavaScript code) generates a new task ID as:

```js
id = "T-" + number
```

In JavaScript, if `number` is `NaN` (Not a Number), the result is the string `"T-NaN"`. That can happen when:

- `number` is `undefined` (e.g. missing timestamp)
- `parseInt(x)` or `Number(x)` returns `NaN` for invalid input
- A date/timestamp calculation yields `NaN` in an edge case

So **T-NaN is produced by the extension/JS layer**, not by exarp-go.

## exarp-go behavior

exarp-go generates task IDs in Go with `generateEpochTaskID()`:

- Uses `time.Now().UnixMilli()` (int64), which cannot be NaN
- Format: `T-{epoch_milliseconds}` (e.g. `T-1769716109198`)

So exarp-go never produces `T-NaN`.

## Detection and repair

- **sanity_check** reports invalid task IDs (e.g. `T-NaN` or any ID where the part after `T-` is not a valid integer).
- **fix_invalid_ids** (task_workflow action) finds tasks with invalid IDs, assigns new epoch IDs, updates dependencies, deletes the old DB row (if any), and saves so both DB and JSON stay consistent.

To fix an existing T-NaN task:

```bash
exarp-go -tool task_workflow -args '{"action":"fix_invalid_ids"}'
```

Prefer creating tasks via exarp-go (`action=create` or `exarp-go task create`) so IDs are always valid.
