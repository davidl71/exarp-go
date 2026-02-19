# Task content hash / checksum (design)

**Purpose:** Use a content checksum (hash) on tasks to support **quick duplicate detection** and **conflict detection** when syncing or merging task state (e.g. DB↔JSON, or future multi-source merge). No schema change required if we store the hash in task `metadata`.

---

## 1. What to hash

- **Content fingerprint:** Hash the same normalized string used for content comparison today: `normalizeTaskContent(Content, LongDescription)` (trim, lowercase, collapse spaces). That matches `tasksMatchByContent` / duplicate logic.
- **Optional “identity” fingerprint:** For conflict detection, some implementations hash a canonical representation of the task (e.g. `content + long_description + status + priority`, normalized) so “same task, different edits” can be detected by hash mismatch.

**Recommendation:** Start with **content-only** hash (content + long_description, normalized). Same hash ⇒ same meaning for duplicate detection; different hash on same task ID after merge ⇒ possible conflict to report.

---

## 2. Where to store

- **Option A – Metadata:** `task.Metadata["content_hash"] = "<hex>"`. No DB migration; works with existing JSON and DB metadata column. Easy to add and backfill.
- **Option B – New column:** `content_hash TEXT` on `tasks` table + index. Slightly faster for “find by hash” and keeps metadata smaller; requires a migration.

**Recommendation:** **Metadata** for v1. Backfill on load if missing (compute and set when reading from DB/JSON); write on create/update. Add a DB column later if we need indexed hash lookups.

---

## 3. Hash algorithm

- **SHA-256** of the normalized content string (hex-encoded, e.g. 64 chars). Good for dedup and conflict checks; no need for crypto security, but SHA-256 is fast and standard.
- Alternative: **xxhash** or **fnv** for speed if we only care about duplicates in-process; SHA-256 keeps hashes comparable across tools/languages if we ever export or compare externally.

**Recommendation:** **SHA-256** (hex) for simplicity and portability.

---

## 4. Use cases

### 4.1 Quick duplicate detection (exact content)

- **Today:** `task_analysis` duplicates use Jaccard similarity over words (and `findTaskByContent` uses full normalized string comparison). Pairwise comparison is O(n²) or O(n) with a content map.
- **With hash:** Compute `content_hash` for each task (or read from metadata). Group by `content_hash`. Tasks in the same group with different IDs are **exact duplicate candidates**. No need to run similarity for those; optionally still run similarity for “near” duplicates (e.g. same hash = auto-group; different hash = skip or run similarity only for small subsets).
- **Implementation:** In `findDuplicateTasks` (or a new path), build `map[contentHash][]taskID`. Any key with `len > 1` is an exact-duplicate group. Backfill hash in metadata when loading if missing.

### 4.2 Conflict resolution (sync / merge)

- **Today:** `SyncTodo2Tasks` merges by task ID; DB overwrites JSON. No detection of “same ID, different content” (e.g. edited on two branches).
- **With hash:** When merging two sources (e.g. DB vs JSON, or local vs remote state):
  - For each task ID present in both, compare `content_hash`.
  - If hashes match ⇒ same content, no conflict.
  - If hashes differ ⇒ both sides changed; report as conflict (and optionally surface in session prime or a dedicated “sync conflicts” report). No automatic overwrite; user or tool can choose “keep DB”, “keep JSON”, or merge manually.
- **Optional:** Store `last_content_hash` (or “base” hash) when last synced, to support 3-way “ours / theirs / base” style resolution later.

---

## 5. Implementation sketch

1. **Hash helper (shared):**
   - `ContentHash(task Todo2Task) string` → normalize content + long_description, SHA-256, return hex.
   - `SetContentHash(task *Todo2Task)` → set `task.Metadata["content_hash"] = ContentHash(*task)` (and optionally ensure Metadata non-nil).

2. **Backfill on load:**
   - Wherever tasks are loaded from DB or JSON (e.g. in `LoadTodo2Tasks`, or inside store adapter), if `task.Metadata["content_hash"]` is empty, set it from `ContentHash(task)`. Persist on next save so we gradually backfill.

3. **Write path:**
   - On create and update (e.g. in `CreateTask` / `UpdateTask` and in `task_workflow` create/update handlers), call `SetContentHash` before save so new and updated tasks always have a hash.

4. **Duplicate detection:**
   - In `findDuplicateTasks` or a new helper: build `map[contentHash][]string` (hash → task IDs). Groups with `len > 1` are exact-duplicate sets. Optionally merge with existing similarity-based groups (e.g. “exact” vs “similar” in report).

5. **Sync / conflict reporting:**
   - In `SyncTodo2Tasks` (or a separate “merge report” step): when building merged list, for each task ID that appears in both DB and JSON, compare `content_hash`; if different, append to a `conflicts` list and return or log it (e.g. `sync_result.conflicts = [{task_id, db_hash, json_hash}]`). Caller can then warn or block overwrite.

6. **API surface:**
   - Optional `task_workflow` or `task_analysis` flag: e.g. `action=sync` with `report_conflicts=true`, or `action=duplicates` with `use_content_hash=true` (default true once implemented). No breaking change if hash is optional and fallback is current behavior.

---

## 6. Other places in exarp where hashes help

- **Sanity check (duplicate content)** — [internal/tools/task_workflow_common.go](internal/tools/task_workflow_common.go) already uses `normalizeTaskContent` to build a `contentToIDs` map (lines 1097–1119). Use the same shared normalizer; optionally use **content hash** as the map key instead of the full normalized string (equivalent behavior, shorter keys).
- **Memory duplicate detection** — [internal/tools/memory_maint.go](internal/tools/memory_maint.go) has `similarityRatio` and `findDuplicateGroups` by title. For **exact** duplicate titles: normalize with the shared helper, then group by hash. Same hash ⇒ exact duplicate; keep existing similarity for “near” duplicates. Optional follow-up: add hash-based exact-dedup pass before or alongside `findDuplicateGroups`.
- **Tasksync matching** — [internal/tasksync/google_tasks.go](internal/tasksync/google_tasks.go) (and similar) uses `Content == Title`. Use shared **NormalizeForComparison** (and optionally hash equality) so matching is robust to whitespace/case. Optional follow-up.
- **Report / overview cache** — [docs/CONTEXT_REDUCTION_OPTIONS.md](docs/CONTEXT_REDUCTION_OPTIONS.md) and [docs/CONTEXT_REDUCTION_FOLLOWUP_TASKS.md](docs/CONTEXT_REDUCTION_FOLLOWUP_TASKS.md) already mention returning `content_hash` (or short hash) on task summaries and report so clients can cache. Reuse the same task content hash when adding `include_hash` / cache support; no second hash type.
- **Estimation history** — [internal/tools/estimation_shared.go](internal/tools/estimation_shared.go) does word-overlap similarity; a content hash would help only for **exact** “same task as before.” Optional later improvement.

---

## 7. Code sharing

Centralize normalize + hash so tasks, sanity check, memory, and tasksync can share one implementation.

- **Shared package:** Put helpers in **internal/models** (or a small **internal/pkg/textfinger** if you prefer to keep models minimal). Both `internal/tools` and `internal/tasksync` can import `internal/models` without cycles; `database` already uses `models.Todo2Task`.
- **API to add (in models):**
  - **NormalizeForComparison(content, description string) string** — Trim, lowercase, collapse spaces (same semantics as current `normalizeTaskContent`). Used for: task content, sanity-check key, memory title (description can be `""`), tasksync title/content.
  - **ContentHashFromString(s string) string** — SHA-256(s) hex. Used for: task `content_hash`, sanity-check map key (optional), memory exact-dedup key (optional), tasksync “same content” (optional).
  - **ContentHash(task \*Todo2Task) string** — `ContentHashFromString(NormalizeForComparison(task.Content, task.LongDescription))`.
  - **SetContentHash(task \*Todo2Task)** — Set `task.Metadata["content_hash"]`; ensure Metadata non-nil.
  - **EnsureContentHash(task \*Todo2Task)** — If `content_hash` missing or empty, call SetContentHash.
- **Migration:** Move the current `normalizeTaskContent` logic from [internal/tools/todo2_db_adapter.go](internal/tools/todo2_db_adapter.go) into `models.NormalizeForComparison`. In tools, replace usages with `models.NormalizeForComparison` (or a thin wrapper). Implement task content hash and duplicate/sync logic on top of the shared helpers so all “content identity” and “duplicate/conflict” behavior uses the same normalization and hashing.

---

## 8. Files to touch (summary)

| Area              | File(s) | Change |
|-------------------|---------|--------|
| Hash + normalize (shared) | **internal/models/** (new file, e.g. `task_hash.go`) | Add `NormalizeForComparison`, `ContentHashFromString`, `ContentHash`, `SetContentHash`, `EnsureContentHash`. |
| tools: use shared | `internal/tools/todo2_db_adapter.go` | Replace local `normalizeTaskContent` with `models.NormalizeForComparison` (or wrapper). |
| Backfill          | `internal/tools/todo2_utils.go`, `internal/tools/todo2_db_adapter.go` (load paths) | Call `EnsureContentHash` when loading from DB/JSON. |
| Write path        | `internal/tools/todo2_db_adapter.go` (`saveTodo2TasksToDB`), `internal/tools/task_store.go` (CreateTask/UpdateTask) | Call `SetContentHash` before persist. |
| Duplicates        | `internal/tools/task_analysis_shared.go` | Group by `content_hash` for exact dupes; combine with existing similarity. |
| Sync/conflict     | `internal/tools/todo2_utils.go` (`SyncTodo2Tasks`) | Compare hashes when same ID in both sources; report conflicts (e.g. optional `*SyncResult` out param). |
| Sanity check      | `internal/tools/task_workflow_common.go` | Use shared normalizer (and optionally hash as map key) for duplicate-content check. |
| Optional follow-ups | `internal/tools/memory_maint.go`, `internal/tasksync/*.go`, report/overview API | Memory: exact-dedup by normalized title/hash. Tasksync: normalize/hash for matching. Report: return task `content_hash` when `include_hash` set. |

---

## 9. Docs and rules

- Document `content_hash` in task metadata (e.g. in `docs/TASK_TOOL_ENRICHMENT_DESIGN.md` or a “Task metadata” section) so agents and tools know it’s stable and comparable.
- Session prime or report could mention “sync conflicts” when `sync_result.conflicts` is non-empty (e.g. “2 tasks have conflicting edits in DB vs JSON”).

---

**Summary:** Adding a **content hash** (SHA-256 of normalized content) in task metadata gives a cheap, consistent way to find exact duplicates and to detect conflicts when merging DB↔JSON (or future sources). **Shared code** in **internal/models** (NormalizeForComparison, ContentHashFromString, ContentHash, SetContentHash, EnsureContentHash) is used by tasks, sanity check, duplicate detection, and sync conflict reporting; the same helpers can be reused for memory exact-dedup, tasksync matching, and report/overview cache (optional follow-ups). Implementing the shared layer, backfill, write-path set, duplicate grouping, and sync conflict reporting covers both “quick checks for duplicates” and “conflict resolution” without a schema change.
