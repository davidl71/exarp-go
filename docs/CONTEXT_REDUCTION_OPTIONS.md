# Context reduction: base64, hashing, compression

Ways to reduce MCP/LLM context size and improve speed in exarp-go.

## 1. Compact JSON (no indentation)

**Where:** All tool responses that use `response.FormatResult()` (JSON is `MarshalIndent(..., "  ")`).

**Idea:** When the client doesn’t need pretty-printing (e.g. Cursor parsing JSON), return compact JSON to cut size and parsing cost.

**Implementation:**
- Support a param like `output_compact=true` or `compact=true` in tools that return JSON.
- Use a helper that marshals with `json.Marshal` (no indent) when compact is set; otherwise keep using `response.FormatResult()`.

**Impact:** Typically 15–30% smaller JSON for nested structures (task lists, report overview, session prime). No behavior change beyond size.

**Status:** Helper added in `internal/tools/response_compact.go`; wired for **session** (action=prime) and **task_workflow** (sync sub_action=list, output_format=json) when `compact=true`.

---

## 2. Content hashing (cache / dedup)

**Where:** Task payloads (e.g. `task_workflow` list, `session` prime), report/overview payloads.

**Ideas:**
- **Task content hash:** Store or return a short hash (e.g. SHA256 first 8 chars) of `content` + `long_description` (and maybe `status`). Clients can skip re-processing or re-sending when hash is unchanged.
- **Report/overview hash:** Return `content_hash` with overview/scorecard so the client can cache and only refetch when hash changes.

**Implementation:**
- Add optional `content_hash` (or `hash`) to task summary objects when a param like `include_hash=true` is set.
- For report, compute hash of the serialized report and include it in the response; client can send `If-None-Match`-style hint in a future request (if we add conditional GET).

**Impact:** Fewer redundant tool calls and less context when the client uses the hash for caching.

---

## 3. Gzip + base64 for large payloads

**Where:** Very large JSON responses (e.g. full task list with 100+ tasks, full report overview with many metrics).

**Idea:** For `output_format=json` and size above a threshold (e.g. 8KB), optionally return:
```json
{"compressed": "gzip+base64", "data": "<base64-encoded-gzipped-json>"}
```
Client (or Cursor) decompresses and parses JSON. MCP content is still text (so no protocol change).

**Implementation:**
- Add param e.g. `compress=true` (or auto when payload size > threshold).
- In a response helper: gzip the JSON bytes, base64-encode, wrap in the envelope above.
- Document that when `compressed` is present, the client must decompress before use.

**Impact:** Large payloads can be 3–5x smaller on the wire. Only useful when payloads are big enough that gzip wins (e.g. > ~1–2KB).

**Caveat:** Cursor/IDE must support decompression or we only use this for CLI/file output.

---

## 4. Compact / summary-only modes

**Where:** Session prime, task_workflow list, report overview.

**Ideas:**
- **Session prime:** `include_tasks=compact` returns only counts + `suggested_next` (id, one-line content), no full `tasks.recent` or full task bodies. Cursor can then call `task_workflow` with `task_id` when it needs details.
- **Task list:** `compact=true` (or a dedicated mode) returns minimal fields: `id`, `status`, `priority`, `content` (truncated to e.g. 80 chars), optional `content_hash`; omit `long_description` and large metadata until requested via `task show`.
- **Report:** `level=brief` or `compact=true` returns a short summary (e.g. health score + top 3 issues) instead of full metrics.

**Implementation:** Already partially there (session has `getTasksSummaryFromTasks` with limited recent; task list has `limit` and text truncation). Add explicit `compact`/`summary_only` params and reduce fields in JSON for those modes.

**Impact:** Large reduction in context for prime and list; report becomes a one-line or short blob.

---

## 5. Database storage compression

**Where:** SQLite task storage (e.g. `long_description`, large `metadata`).

**Idea:** Store long text columns as gzipped blobs; decompress on read. SQLite doesn’t require schema change if we store bytes in a blob column or encode compressed bytes in a text column (e.g. base64).

**Implementation:** Optional compression in database layer for fields above a length threshold; transparent decompression in `GetTask`/`ListTasks`. Migration path for existing rows.

**Impact:** Smaller DB and less I/O; slight CPU cost on read/write. Best for projects with many tasks with long descriptions.

---

## 6. Resource content

**Where:** MCP resources (e.g. `stdio://models`, `stdio://tool_catalog`).

**Idea:** If a resource body is large and rarely changes, serve it gzip+base64 with a small header so clients that support it can decompress. Most resources are small; only consider for large dynamic resources.

**Impact:** Low unless we add large resource payloads.

---

## Summary

| Option              | Effort | Context reduction      | Best for                |
|---------------------|--------|------------------------|-------------------------|
| Compact JSON        | Low    | ~15–30%                | All JSON tool responses |
| Content hashing     | Medium | Fewer duplicate calls  | task list, report, prime|
| Gzip+base64         | Medium | 3–5x for large blobs   | Large lists/reports     |
| Compact/summary mode| Low–med| Large (fewer fields)    | Prime, list, report     |
| DB compression      | Medium | Storage + I/O          | Many tasks, long descs  |
| Resource compression| Low    | Only if resources grow | Large resources         |

**Recommended order:** (1) Compact JSON and (4) compact/summary modes first; then (2) content hashing for caching; then (3) gzip+base64 for very large responses if needed.

---

## Agent hints: what’s safe to remove from context

exarp-go can give agents (Cursor, Claude, Codex, OpenCode, Ollama) explicit hints about what is **safe to summarize or drop** so they can reduce context without guessing.

### 1. Session prime

With **`include_hints: true`**, prime returns a **`context_reduction`** hint telling the agent to:

- Use **`compact: true`** on prime, task_workflow (list), and report when consuming JSON.
- When context is large, call **`context(action=budget, items=[...], budget_tokens=N)`** to get **`safe_to_summarize`** and **`agent_hint`**.

So the agent sees at session start that it *can* ask exarp for “what’s safe to remove” by calling the context tool.

### 2. context(action=budget)

The agent passes its current context **items** (e.g. report overview, task list, handoff text) and an optional **`budget_tokens`**. exarp returns:

| Field | Purpose |
|-------|--------|
| **`items`** | Per-item token estimate and **`recommendation`**: `summarize_brief`, `summarize_key_metrics`, `keep_detailed`, `keep_full`. |
| **`safe_to_summarize`** | Indices of items that are **safe to replace with a summary** (no need to keep full text). |
| **`agent_hint`** | One-line instruction, e.g. *“Safe to summarize items [0, 1, 2] (use context action=summarize or batch with level=brief) to fit budget.”* |
| **`strategy`** | Human-readable reduction strategy and estimated savings. |

The agent can then:

- **Summarize** the listed items via `context(action=summarize, ...)` or `context(action=batch, ...)` with `level=brief`.
- **Drop** full payloads and keep only summaries when over budget.

So “safe to remove” is expressed as “safe to **summarize**” (replace with a short version) rather than “delete”; the agent keeps a compressed version instead of losing information.

### 3. Static best practices (rules)

- **`.cursor/rules/context-reduction.mdc`** and **`exarp-mcp-output.mdc`** tell the AI to use **`compact: true`** and, when available, **`content_hash`** to avoid re-sending unchanged data. That reduces context size without exarp having to list “safe to remove” items.

---

## Scripts and Makefile

- **Makefile:** `exarp-report-scorecard` and `exarp-report-overview` use `output_format=json` and `compact=true` in their args (compact takes effect once the report tool wires `FormatResultOptionalCompact`). See `_exarp_report_scorecard_args` and `_exarp_report_overview_args`.
- **Shell scripts / CI:** When invoking exarp-go with `-tool report`, `-tool session`, or `-tool task_workflow` and consuming JSON, pass `"compact":true` and (when available) `"output_format":"json"` to reduce payload size. Example:  
  `exarp-go -tool session -args '{"action":"prime","include_tasks":true,"include_hints":true,"compact":true}'`
- **Follow-up tasks:** See [CONTEXT_REDUCTION_FOLLOWUP_TASKS.md](CONTEXT_REDUCTION_FOLLOWUP_TASKS.md) for the full list of implementation and automation tasks.
