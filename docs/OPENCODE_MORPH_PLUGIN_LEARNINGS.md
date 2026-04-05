# OpenCode Morph plugin — learnings for exarp-go

Patterns from the **OpenCode + Morph** plugin ([`@morphllm/opencode-morph-plugin`](https://github.com/morphllm/opencode-morph-plugin)), also vendored alongside this monorepo at `../opencode-morph-plugin` when using the `mcp` workspace layout. Not a dependency of exarp-go; this doc captures **product and safety patterns** for future work (OpenCode docs, session compaction, agent edit flows).

**See also:** [OPENCODE_INTEGRATION.md](OPENCODE_INTEGRATION.md), [CONTEXT_REDUCTION_OPTIONS.md](CONTEXT_REDUCTION_OPTIONS.md), [AGNO_GO_LEARNINGS.md](AGNO_GO_LEARNINGS.md).

---

## Summary table

| Area | Morph / OpenCode pattern | exarp-go relevance |
|------|-------------------------|-------------------|
| **Fast apply** | Partial edits with `// ... existing code ...`, API merge, markdown-fence stripping | No equivalent today; native edits are exact/replace. Optional future: merge-based patch tool or stricter pre-write guards. |
| **Edit safety** | Preflight: missing markers on large files, marker leakage in output, catastrophic size drop → refuse write | Transferable to any agent-driven file write path (MCP helpers, batch edits). |
| **Agent modes** | `morph_edit` blocked for readonly agents (`plan`, `explore`) unless env override | Same principle as lane/gating: exploration vs implementation separation. |
| **Search routing** | WarpGrep for exploratory NL queries; plain `grep` for exact symbols | Parallels “task_analysis / discovery vs ripgrep” guidance; keep routing docs explicit. |
| **Public repos** | `warpgrep_github_search` without clone | Complements “fetch README / API” flows; optional doc note in research helpers. |
| **Compaction** | `messages.transform`: threshold from model context, **frozen** prefix bytes for cache stability, re-compact only tail (no double-compact) | Relevant if exarp-go adds **conversation** or **handoff** compression beyond token_estimate; see CONTEXT_REDUCTION_OPTIONS / Agno table. |
| **Policy vs code** | `instructions/morph-tools.md` always loaded in `opencode.json` for tool choice | Parallels **skills + `.mdc` rules**; canonical “which tool when” should stay discoverable. |
| **Config** | Env flags per feature (`MORPH_EDIT`, `MORPH_COMPACT`, …) | Pattern for optional backends (already used across LLM providers). |

---

## Implementation pointers (local tree)

When the repo is checked out next to exarp-go:

| Asset | Purpose |
|--------|---------|
| `opencode-morph-plugin/index.ts` | Plugin entry: shared `MorphClient` / timeouts, tool definitions, hooks |
| `opencode-morph-plugin/instructions/morph-tools.md` | Always-on routing policy for the model |
| `opencode-morph-plugin/README.md` | Install, compaction tuning, tool summaries |

---

## Follow-up work (tracked in Todo2)

Tasks created from this doc should use tags like `opencode`, `morph`, or `agent-ux` and reference this file in **planning_doc** where applicable. IDs are listed in **Todo2 tracking** below after creation.

Suggested themes (non-exhaustive):

1. **Examples** — Optional `docs/examples/` OpenCode JSON snippet showing `plugin` + `instructions` for Morph alongside exarp-go MCP.
2. **Context / session** — Spike: frozen-prefix or semantic compaction for long sessions or handoff bundles (align with Agno/compression plans).
3. **Edit / MCP safety** — Audit: apply Morph-like guards (truncation, marker/literal leakage) to any centralized write helpers if introduced.
4. **Documentation** — Keep [OPENCODE_INTEGRATION.md](OPENCODE_INTEGRATION.md) as the primary OpenCode doc; link here for Morph-specific patterns.

---

## Todo2 tracking

| Task ID | Title (short) |
|---------|----------------|
| `T-1775401689648347000` | OpenCode example: Morph plugin + exarp-go MCP |
| `T-1775401689674803000` | Session/handoff: spike frozen-prefix compaction pattern |
| `T-1775401689696819000` | Audit: agent file-write paths for merge safety rails |
| `T-1775401689719490000` | Docs: canonical tool-routing guidance (skills vs always-on rules) |

---

## References

- Morph plugin (upstream): [github.com/morphllm/opencode-morph-plugin](https://github.com/morphllm/opencode-morph-plugin)
- Morph docs: [docs.morphllm.com](https://docs.morphllm.com/quickstart)
- OpenCode config: [opencode.ai/docs/config](https://opencode.ai/docs/config/)
