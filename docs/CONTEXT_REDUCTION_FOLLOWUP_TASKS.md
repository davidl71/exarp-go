# Context reduction follow-up tasks

Create these Todo2 tasks (e.g. `exarp-go task create "..."` or task_workflow action=create) when the task store is available. See [CONTEXT_REDUCTION_OPTIONS.md](CONTEXT_REDUCTION_OPTIONS.md) for implementation details.

## Implementation tasks

| # | Title | Description | Priority | Tags |
|---|--------|-------------|----------|------|
| 1 | **Content hashing: task list and report** | Add optional `include_hash` / `content_hash` param; return short hash (e.g. SHA256 first 8 chars) on task summaries and report/overview so clients can cache. §2 in CONTEXT_REDUCTION_OPTIONS.md. | medium | #context #performance |
| 2 | **Gzip+base64 for large JSON payloads** | Add `compress=true` (or size-threshold auto); return `{"compressed":"gzip+base64","data":"..."}` for large tool responses; document client decompression. §3. | medium | #context #performance #mcp |
| 3 | **Compact/summary-only modes** | Session prime: `include_tasks=compact` (counts + suggested_next only). Task list: minimal fields mode (id, status, priority, content truncated, omit long_description). Report: `level=brief` or `compact=true`. §4. | medium | #context #session #task_workflow #report |
| 4 | **DB compression for long text** | Optional gzip of `long_description` / large metadata in SQLite; transparent decompress in GetTask/ListTasks; migration for existing rows. §5. | medium | #database #performance |
| 5 | **Wire compact=true in more tools** | Add `compact` param and use `FormatResultOptionalCompact` in task_analysis, health, and other high-volume JSON tools (report/overview/scorecard/briefing already wired). | low | #context #mcp |
| 6 | **Resource compression (optional)** | If any MCP resource body becomes large, serve gzip+base64 with header for clients that support it. §6. | low | #context #mcp |

## Cursor / automation

| # | Title | Description | Priority | Tags |
|---|--------|-------------|----------|------|
| 7 | **Cursor rules: compact and hashing** | Ensure exarp-mcp-output.mdc and session-prime.mdc (and any agent rules) recommend `compact=true` and document `content_hash` when available. | low | #docs #cursor |
| 8 | **Makefile: use compact for report/context** | Pass `compact=true` in exarp-report-scorecard, exarp-report-overview, and exarp-context-budget args where JSON is consumed by scripts. | low | #build #context |
| 9 | **Scripts: document compact for CI/automation** | In scripts that call exarp-go -tool report or session, add COMPACT=1 or use compact in args; document in CONTEXT_REDUCTION_OPTIONS.md or README. | low | #docs #scripts |

---

**Create all:** After fixing DB/migrations, run create for each row (name = Title, long_description = Description, priority/tags as listed).
