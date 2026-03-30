# Protobuf — remaining work (optional)

**Last updated:** 2026-03-30  

This file previously duplicated handler migration checklists that **contradicted** the “100% complete” summary (stale items 9–20 listed helpers that already exist). **Do not use this file as a migration guide.**

## Canonical references

| Need | Document |
|------|----------|
| **Current implementation + code map** | [PROTOBUF_IMPLEMENTATION_STATUS.md](PROTOBUF_IMPLEMENTATION_STATUS.md) |
| **Regenerating `.pb.go`, Makefile targets** | [PROTOBUF_USAGE.md](PROTOBUF_USAGE.md) |
| **Integration overview (handlers, bridge, tests)** | [PROTOBUF_INTEGRATION.md](PROTOBUF_INTEGRATION.md) |
| **Historical research & snapshots** | [archive/protobuf/README.md](archive/protobuf/README.md) (or stub [PROTOBUF_ANALYSIS.md](PROTOBUF_ANALYSIS.md)) |

## Migration status (summary)

- **Tool handlers:** Protobuf-first args with JSON fallback; `internal/tools/protobuf_helpers.go` + **mcp-go-core** `request.ParseRequest[T]()`. Native handlers still receive `json.RawMessage` after params map marshal.
- **Task DB metadata, config, memory files, report/scorecard internals:** Implemented per [PROTOBUF_IMPLEMENTATION_STATUS.md](PROTOBUF_IMPLEMENTATION_STATUS.md).
- **Python bridge:** Go↔Python protobuf path exists (`internal/bridge/python.go`, `bridge/execute_tool.py`). Optional: return more fields as binary `ToolResponse` everywhere if profiling shows benefit.

## Optional follow-ups (backlog)

1. **MCP wire JSON** — Tool *results* normalized for stdio clients live in **mcp-go-core** (not exarp-go). No extra proto layer required unless you want typed response messages.
2. **Docs** — Keep STATUS + USAGE + INTEGRATION in sync when adding tools or proto files.
3. **Benchmarks** — Run/document `todo2_protobuf` benches vs JSON; attach numbers to STATUS or `docs/OPTIMIZATION_RESULTS.md`.
4. **Automation** — Ensure CI/dev images document `make proto` / `protoc` (Ansible role already mentioned in PROTOBUF_INTEGRATION).

For a **prioritized action plan**, see **“Action plan”** in [PROTOBUF_IMPLEMENTATION_STATUS.md](PROTOBUF_IMPLEMENTATION_STATUS.md).
