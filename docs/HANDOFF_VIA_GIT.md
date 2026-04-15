# Handoff via Git so Remote Gets exarp Task List

**Purpose:** How to hand off work so a **remote machine** can get the current list of exarp (Todo2) tasks using **git** (pull), without access to your local `.todo2/` data.

**See also:** [session-handoff skill](../.cursor/skills/session-handoff/SKILL.md), [CURSOR_API_AND_CLI_INTEGRATION.md](CURSOR_API_AND_CLI_INTEGRATION.md).

---

## Why git alone is not enough

- **`.todo2/` is gitignored.** Task storage (SQLite `todo2.db` and fallback `state.todo2.json`) and handoffs (`handoffs.json`) live under `.todo2/`, so they are **not** in the repo. A remote that only runs `git pull` will **not** get your task list or handoff notes.

- To give the remote the **current list of exarp tasks** and handoff context via git, you must **export that data into tracked files**, commit, and push. The remote then pulls and reads those files.

---

## Recommended flow (source machine → git → remote)

Do this on the **machine that has the tasks** (where you're ending the session), then commit and push. On the **remote**, pull and read the exported files.

### 1. End session with handoff (includes in-progress tasks)

Create a handoff note so the remote knows what was done and what’s next. Include tasks so the handoff carries context.

**MCP (preferred):**
```json
{
  "action": "handoff",
  "sub_action": "end",
  "summary": "Short summary of what was done and what’s left",
  "include_tasks": true,
  "include_git_status": true
}
```

**CLI:**
```bash
exarp-go -tool session -args '{"action":"handoff","sub_action":"end","summary":"Your summary","include_tasks":true}'
```

**Special handoffs: task journal and point-in-time snapshot**

You can attach two optional extras to the handoff:

| Option | Param | Description |
|--------|--------|-------------|
| **Task journal** | `modified_task_ids` or `task_journal` | List of tasks modified this session. `modified_task_ids`: array of task IDs (e.g. `["T-1","T-2"]`). `task_journal`: JSON array of objects `{ "id": "T-1", "action": "updated" }` (or pass as string). Stored in handoff as `task_journal`. |
| **Point-in-time snapshot** | `include_point_in_time_snapshot: true` | Full task list at handoff time, stored as **gzip+base64** in the handoff (`point_in_time_snapshot`). Format marker: `point_in_time_snapshot_format`: `"gz+b64"`. Decode to get state.todo2.json-shaped JSON (see below). |

**Example: handoff with journal and snapshot**
```json
{
  "action": "handoff",
  "sub_action": "end",
  "summary": "Completed T-1 and T-2; T-3 in progress",
  "include_tasks": true,
  "include_point_in_time_snapshot": true,
  "modified_task_ids": ["T-1", "T-2", "T-3"]
}
```

**CI:** The **Go CI/CD** workflow (`.github/workflows/go.yml`) runs a full handoff after build and uploads **`docs/CI_HANDOFF.json`** as the **`ci-handoff`** artifact (7-day retention). The **Pre-release (full)** job runs on `workflow_dispatch` or push to `main` and runs `make pre-release`, uploading binary + handoff as **`pre-release-artifacts`** (14-day retention). **Agentic CI** (`.github/workflows/agentic-ci.yml`) also exports the handoff and uploads it as **`ci-handoff`**. Locally, `make pre-release` or `make handoff-export` writes **`docs/CI_HANDOFF.json`**; commit it to give the remote a single handoff + snapshot blob.

**Decoding the point-in-time snapshot**

The snapshot is gzip-compressed, then base64-encoded. To decode:

- **In code (Go):** Use `tools.DecodePointInTimeSnapshot(encoded)` to get raw JSON bytes; then `tools.ParseTasksFromJSON(bytes)` to get `[]Todo2Task`.
- **By hand:** Base64-decode the `point_in_time_snapshot` string, then gzip-decompress the result. You get JSON in state.todo2.json shape (top-level `"todos"` array). You can write that to `.todo2/state.todo2.json` (and run exarp-go migrate if you use DB) to restore that point in time.

### 2. Export handoff to a tracked file

Write the handoff data into a file under `docs/` (or another tracked path) so it’s committed and the remote can read it after `git pull`.

**MCP:**
```json
{
  "action": "handoff",
  "sub_action": "export",
  "output_path": "docs/HANDOFF_LATEST.json",
  "export_latest": true
}
```

**CLI:**
```bash
exarp-go -tool session -args '{"action":"handoff","sub_action":"export","output_path":"docs/HANDOFF_LATEST.json","export_latest":true}'
```

### 3. (Optional) Export full task list to a tracked file

So the remote has the **current list of exarp tasks** in one place (not only what’s in the handoff), dump the task list to a tracked file. Use JSON for structure; the CLI supports `--json`.

**CLI (all tasks, JSON):**
```bash
exarp-go task list --status all --json > docs/TASKS_SNAPSHOT.json
```

To limit size you can filter, e.g. open tasks only:
```bash
exarp-go task list --json > docs/TASKS_SNAPSHOT.json
```

### 4. Commit and push

```bash
git add docs/HANDOFF_LATEST.json docs/TASKS_SNAPSHOT.json   # add only what you created
git commit -m "Handoff: export handoff and task list for remote"
git push
```

(Or use `make p` from repo root.)

---

## On the remote machine

1. **Pull:**  
   `git pull` (or `make pl` from repo root).

2. **Read handoff:**  
   Open `docs/HANDOFF_LATEST.json` for summary, blockers, next steps, and any in-progress tasks included in the handoff.

3. **Read task list:**  
   Open `docs/TASKS_SNAPSHOT.json` for the full list of exarp tasks as of the export.

4. **Resume (optional):**  
   If you use exarp-go on the remote and want to “resume” from the handoff in the TUI or via MCP, you still don’t have `.todo2/` state there. The handoff **content** is in `docs/HANDOFF_LATEST.json`; use it as the prompt/context. To have exarp-go **task** commands see the same tasks on the remote, you’d need either to commit `.todo2/` (see below) or implement an import-from-snapshot flow.

---

## Summary

| Goal | Action |
|------|--------|
| Remote gets **handoff context** (summary, next steps, in-progress tasks) via git | Export handoff to `docs/HANDOFF_LATEST.json`, commit, push. Remote pulls and reads the file. |
| Remote gets **current list of exarp tasks** via git | Export task list to `docs/TASKS_SNAPSHOT.json`, or use **point-in-time snapshot** in the handoff (gzip+base64) and export the handoff to a tracked file. |
| **Task journal** (which tasks were modified this session) | Pass `modified_task_ids` or `task_journal` when ending the handoff; stored in handoff and included in exported JSON. |
| **Point-in-time snapshot** (full task state, compressed) | Pass `include_point_in_time_snapshot: true` when ending the handoff; decode with `DecodePointInTimeSnapshot` or base64 then gzip. |
| Remote runs **exarp-go task commands** against that same data | Not supported by default: `.todo2/` is gitignored. Options: commit `.todo2/`, or decode the point-in-time snapshot and write to `.todo2/state.todo2.json` (then sync/migrate as needed). |

**Best practice:** Export handoff to a tracked path (e.g. `docs/HANDOFF_LATEST.json`). For a single portable blob, use `include_point_in_time_snapshot: true` and export the handoff; the remote can decode the snapshot to get the full task state.
