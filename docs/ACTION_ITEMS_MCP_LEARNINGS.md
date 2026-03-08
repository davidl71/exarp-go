# Action items: MCP core learnings (exarp tasks)

Concrete action items from [../../docs/exarp-devwisdom-mcp-core-learnings.md](../../docs/exarp-devwisdom-mcp-core-learnings.md). Use this list as a checklist or import into the task store (e.g. batch create).

**Execution order:** See [../../docs/PLAN_ALL_ACTION_ITEMS.md](../../docs/PLAN_ALL_ACTION_ITEMS.md) — Phase A (item 8) first, then B (1,2), C (6,12), D (5,7), E (3,4,9,10,11).

---

## Plan mapping (item → phase → owner)

| Item | Phase | Owner | Order |
|------|--------|--------|-------|
| 8 | A | exarp-go | 1 – Do once |
| 1 | B | devwisdom-go | 2 – Incremental |
| 2 | B | devwisdom-go | 3 – When adding proto/small payloads |
| 6 | C | mcp-go-core / apps | 4 – As needed |
| 12 | C | testing | 5 – As needed |
| 5 | D | mcp-go-core | 6 – Optional |
| 7 | D | mcp-go-core | 7 – Optional |
| 3 | E | devwisdom-go | 8 – When-needed |
| 4 | E | devwisdom-go | 9 – When-needed (depends on 3) |
| 9 | E | exarp-go / devwisdom-go | 10 – When-needed |
| 10 | E | any | 11 – Guideline |
| 11 | E | apps | 12 – Ongoing |

---

## Tasks (name + description)

1. **Align devwisdom-go tool handlers to framework.ToolHandler**  
   Align new tool handlers to `(ctx, json.RawMessage) → ([]types.TextContent, error)`; migrate existing handlers when touching them so hooks/filter apply. Owner: devwisdom-go.

2. **Use ParseRequest and FormatResultCompact in devwisdom-go**  
   When adding proto or smaller payloads: use `request.ParseRequest` and `response.FormatResultCompact` where appropriate. Owner: devwisdom-go.

3. **Consider optional .proto for devwisdom-go tools**  
   If tool contracts grow: consider optional `.proto` for tools; keep JSON-only until then. Owner: devwisdom-go.

4. **Add proto build step in devwisdom-go**  
   If introducing proto: add a Makefile/script to generate Go from `.proto`. Owner: devwisdom-go.

5. **Optional validate-params helper in mcp-go-core**  
   Add validate-params helper using go-playground/validator or go-jsonschema; integrate with `WrapMapToolHandler` or adapter. Owner: mcp-go-core.

6. **Add rate limiting or auth middleware**  
   Add rate limiting (e.g. golang.org/x/time/rate) or auth as additional middleware. Owner: mcp-go-core or apps.

7. **Optional slog/zerolog backend in mcp-go-core**  
   Use slog or zerolog as logging backend behind existing `logging.Logger` and `WithRequestID`. Owner: mcp-go-core.

8. **Refresh vendor when bumping mcp-go-core**  
   When bumping mcp-go-core dependency: run `go mod vendor` in exarp-go. Owner: exarp-go.

9. **Optional proto for large/binary resources**  
   For large or binary resource payloads: consider optional proto resource MIME type. Owner: exarp-go or devwisdom-go.

10. **Keep MCP tool args JSON or proto-at-the-edge for gRPC**  
    For gRPC or other RPC: use proto there; keep MCP tool args as JSON or proto-at-the-edge. Owner: any.

11. **Keep ctxcache in apps**  
    Keep ctxcache in apps; only add cache-in-context API in core if standardizing the pattern. Owner: apps.

12. **Use httptest for transport tests**  
    Use httptest for HTTP/SSE transport tests; keep existing request/response/protocol tests in core. Owner: testing.

---

## Batch-create JSON (for task_create batch)

Use with the task batch-create tool if your workflow accepts an array of `{name, description, ...}`. Create tasks in plan order (phase A first); filter by tag `mcp-learnings` or `phase-a` / `phase-b` / etc.

**Create all 12 tasks from CLI (from exarp-go repo):**
```bash
exarp-go task create-batch --file docs/tasks_mcp_learnings_batch.json
```
Or use the JSON file `docs/tasks_mcp_learnings_batch.json` (same content as "With plan tags" below).

**With plan tags** (phase + owner):

```json
[
  {"name": "[A-8] Refresh vendor when bumping mcp-go-core", "long_description": "Phase A. When bumping mcp-go-core: run go mod vendor in exarp-go. Owner: exarp-go. Plan: docs/PLAN_ALL_ACTION_ITEMS.md", "tags": ["mcp-learnings", "phase-a", "exarp-go"]},
  {"name": "[B-1] Align devwisdom-go tool handlers to framework.ToolHandler", "long_description": "Phase B. Align new handlers to (ctx, json.RawMessage) → ([]types.TextContent, error); migrate when touching. Owner: devwisdom-go.", "tags": ["mcp-learnings", "phase-b", "devwisdom-go"]},
  {"name": "[B-2] Use ParseRequest and FormatResultCompact in devwisdom-go", "long_description": "Phase B. When adding proto or smaller payloads use request.ParseRequest and response.FormatResultCompact. Owner: devwisdom-go.", "tags": ["mcp-learnings", "phase-b", "devwisdom-go"]},
  {"name": "[C-6] Add rate limiting or auth middleware", "long_description": "Phase C. Add rate limiting (golang.org/x/time/rate) or auth middleware. Owner: mcp-go-core or apps.", "tags": ["mcp-learnings", "phase-c", "mcp-go-core"]},
  {"name": "[C-12] Use httptest for transport tests", "long_description": "Phase C. Use httptest for HTTP/SSE transport tests; keep request/response/protocol tests in core. Owner: testing.", "tags": ["mcp-learnings", "phase-c", "testing"]},
  {"name": "[D-5] Optional validate-params helper in mcp-go-core", "long_description": "Phase D. Add validate-params helper (validator or go-jsonschema); integrate with WrapMapToolHandler. Owner: mcp-go-core.", "tags": ["mcp-learnings", "phase-d", "mcp-go-core"]},
  {"name": "[D-7] Optional slog/zerolog backend in mcp-go-core", "long_description": "Phase D. Use slog or zerolog behind logging.Logger and WithRequestID. Owner: mcp-go-core.", "tags": ["mcp-learnings", "phase-d", "mcp-go-core"]},
  {"name": "[E-3] Consider optional .proto for devwisdom-go tools", "long_description": "Phase E. If tool contracts grow consider .proto; keep JSON-only until then. Owner: devwisdom-go.", "tags": ["mcp-learnings", "phase-e", "devwisdom-go"]},
  {"name": "[E-4] Add proto build step in devwisdom-go", "long_description": "Phase E. If introducing proto add Makefile/script to generate Go. Depends on E-3. Owner: devwisdom-go.", "tags": ["mcp-learnings", "phase-e", "devwisdom-go"]},
  {"name": "[E-9] Optional proto for large/binary resources", "long_description": "Phase E. For large/binary resource payloads consider proto resource MIME type. Owner: exarp-go or devwisdom-go.", "tags": ["mcp-learnings", "phase-e"]},
  {"name": "[E-10] Keep MCP tool args JSON or proto-at-the-edge for gRPC", "long_description": "Phase E (guideline). For gRPC use proto there; keep MCP tool args JSON or proto-at-the-edge. Owner: any.", "tags": ["mcp-learnings", "phase-e", "guideline"]},
  {"name": "[E-11] Keep ctxcache in apps", "long_description": "Phase E. Keep ctxcache in apps; add cache-in-context API in core only if standardizing. Owner: apps.", "tags": ["mcp-learnings", "phase-e", "apps"]}
]
```

**Plain (name + long_description only)** — same 12 tasks in plan execution order (A→B→C→D→E):
```json
[
  {"name": "Refresh vendor when bumping mcp-go-core", "long_description": "Phase A. When bumping mcp-go-core dependency: run go mod vendor in exarp-go. Owner: exarp-go."},
  {"name": "Align devwisdom-go tool handlers to framework.ToolHandler", "long_description": "Phase B. Align new tool handlers to (ctx, json.RawMessage) → ([]types.TextContent, error); migrate existing handlers when touching them so hooks/filter apply. Owner: devwisdom-go."},
  {"name": "Use ParseRequest and FormatResultCompact in devwisdom-go", "long_description": "Phase B. When adding proto or smaller payloads: use request.ParseRequest and response.FormatResultCompact where appropriate. Owner: devwisdom-go."},
  {"name": "Add rate limiting or auth middleware", "long_description": "Phase C. Add rate limiting (e.g. golang.org/x/time/rate) or auth as additional middleware. Owner: mcp-go-core or apps."},
  {"name": "Use httptest for transport tests", "long_description": "Phase C. Use httptest for HTTP/SSE transport tests; keep existing request/response/protocol tests in core. Owner: testing."},
  {"name": "Optional validate-params helper in mcp-go-core", "long_description": "Phase D. Add validate-params helper using go-playground/validator or go-jsonschema; integrate with WrapMapToolHandler or adapter. Owner: mcp-go-core."},
  {"name": "Optional slog/zerolog backend in mcp-go-core", "long_description": "Phase D. Use slog or zerolog as logging backend behind existing logging.Logger and WithRequestID. Owner: mcp-go-core."},
  {"name": "Consider optional .proto for devwisdom-go tools", "long_description": "Phase E. If tool contracts grow: consider optional .proto for tools; keep JSON-only until then. Owner: devwisdom-go."},
  {"name": "Add proto build step in devwisdom-go", "long_description": "Phase E. If introducing proto: add a Makefile/script to generate Go from .proto. Owner: devwisdom-go."},
  {"name": "Optional proto for large/binary resources", "long_description": "Phase E. For large or binary resource payloads: consider optional proto resource MIME type. Owner: exarp-go or devwisdom-go."},
  {"name": "Keep MCP tool args JSON or proto-at-the-edge for gRPC", "long_description": "Phase E (guideline). For gRPC or other RPC: use proto there; keep MCP tool args as JSON or proto-at-the-edge. Owner: any."},
  {"name": "Keep ctxcache in apps", "long_description": "Phase E. Keep ctxcache in apps; only add cache-in-context API in core if standardizing the pattern. Owner: apps."}
]
```
