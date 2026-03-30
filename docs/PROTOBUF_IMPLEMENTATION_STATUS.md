# Protobuf implementation — status and plan

**Last updated:** 2026-03-30  

## Document map (read this first)

| Document | Role |
|----------|------|
| **This file** | Single source of truth for **what is implemented** and **what is optional next**. |
| [PROTOBUF_USAGE.md](PROTOBUF_USAGE.md) | **Build:** `make proto`, `proto-check`, `proto-clean`, buf, prerequisites. |
| [PROTOBUF_INTEGRATION.md](PROTOBUF_INTEGRATION.md) | **Where it plugs in:** Makefile, buf, Ansible, tests, handler pattern. |
| [PROTOBUF_REMAINING_WORK.md](PROTOBUF_REMAINING_WORK.md) | Short **optional backlog** + links (no duplicate checklists). |
| [archive/protobuf/README.md](archive/protobuf/README.md) | **Historical:** pre-migration analysis, simplification rationale, Phase 1 snapshot, TUI future notes. |
| [PROTOBUF_ANALYSIS.md](PROTOBUF_ANALYSIS.md) | **Stub** → archive (stable URL for old links). |
| [PROTOBUF_IMPLEMENTATION_PROGRESS.md](PROTOBUF_IMPLEMENTATION_PROGRESS.md) | **Stub** → archive. |
| [PROTOBUF_IMPLEMENTATION_SUMMARY.md](PROTOBUF_IMPLEMENTATION_SUMMARY.md) | **Stub** → archive. |
| [PROTOBUF_SIMPLIFICATION_OPPORTUNITIES.md](PROTOBUF_SIMPLIFICATION_OPPORTUNITIES.md) | **Stub** → archive. |
| [PROTOBUF_TUI_FUTURE_IMPROVEMENTS.md](PROTOBUF_TUI_FUTURE_IMPROVEMENTS.md) | **Stub** → archive (TUI3270 + proto ideas). |
| [CONFIGURATION_PROTOBUF_INTEGRATION.md](CONFIGURATION_PROTOBUF_INTEGRATION.md) | Config-specific protobuf details. |

---

## Current implementation (what is done)

### 1. Proto schemas and generated code

- **proto/tools.proto** — Tool request/response messages. → `proto/tools.pb.go`
- **proto/todo2.proto** — Todo2Task, Todo2State. → `proto/todo2.pb.go`; used by `internal/models/todo2_protobuf.go` and DB metadata.
- **proto/config.proto** — FullConfig. → `proto/config.pb.go`
- **proto/bridge.proto** — ToolRequest / ToolResponse. → `proto/bridge.pb.go`; Python `bridge/proto/bridge_pb2.py`

### 2. Tool request parsing

- **internal/tools/protobuf_helpers.go** — Per-tool `Parse*Request` + `*RequestToParams`; uses **mcp-go-core** `request.ParseRequest[T]()` where applicable (protobuf-first, JSON fallback).
- **internal/tools/handlers_wrap.go** — `WrapHandler`: parse → convert → `ApplyDefaults` → native handler.
- **internal/tools/handlers.go** — Handlers wired to the above pattern (some tools use inline parse where historically duplicated).

### 3. Config

- **internal/config/protobuf.go** — `ToProtobuf` / `FromProtobuf`
- **internal/config/loader.go** — Prefers `.exarp/config.pb`, YAML fallback
- **internal/cli/config.go** — Export / yaml ↔ protobuf

### 4. Database (task metadata)

- **migrations/003_add_protobuf_support.sql** — `metadata_protobuf`, `metadata_format`
- **internal/database/tasks.go** — `SerializeTaskMetadata`; protobuf when format is `protobuf`, JSON fallback
- **internal/models/todo2_protobuf.go** — Serialize/deserialize, tests and benches

### 5. Memory tool

- **internal/tools/memory.go** — Writes `.pb`; loads `.pb` then `.json`

### 6. Python bridge

- **internal/bridge/python.go** — Binary `ToolRequest` on stdin when `--protobuf`
- **bridge/execute_tool.py** — Parses protobuf when available

### 7. Report / scorecard

- Overview and scorecard paths use proto-backed aggregation and conversion helpers (see git history / `report.go` and related).

### 8. Schema version

- **internal/database/schema.go** — `SchemaVersion` matches applied migrations (see that file for the current number).

---

## Remaining / optional work

| Area | Priority | Notes |
|------|----------|--------|
| **Benchmarks + doc** | Low | Run `internal/models` protobuf benches; summarize in STATUS or `docs/OPTIMIZATION_RESULTS.md`. |
| **CI / dev env** | Medium | Document or gate `make proto` in contributor docs / CI if `.pb.go` drift is a problem. |
| **Doc drift** | Medium | When adding a tool: extend `tools.proto`, `make proto`, document in USAGE if new messages. |
| **Bridge responses** | Low | Optional: always emit binary `ToolResponse` end-to-end if measured win. |

---

## Action plan (suggested order)

1. **Treat [PROTOBUF_IMPLEMENTATION_STATUS.md](PROTOBUF_IMPLEMENTATION_STATUS.md) as canonical** — Link new PRs or tasks here instead of spawning parallel “progress” files.
2. **Optional: add a CI check** — Fail PR if `proto/*.proto` changed but generated `*.pb.go` not updated (`make proto` + `git diff --exit-code`), if the team wants strict enforcement.
3. **Run and record one benchmark pass** — Close the “benchmarks” row in the table above with real numbers or “no significant win vs JSON for N tasks.”
4. **On new MCP tools** — Add messages to `tools.proto`, regenerate, add `Parse*` / `*ToParams` in `protobuf_helpers.go`, register handler with `WrapHandler` or the same parse pattern as siblings.
5. **MCP client JSON** — For stdio result shape, follow **mcp-go-core** release notes / adapter middleware; exarp-go stays focused on **request** proto + **native** JSON params.

---

## Historical note

Older internal Todo2 items (“rename migration 002”, “migrate handler X”) are **obsolete** if already shipped. Track new work in Todo2 with references to this file instead of duplicating checklists in [PROTOBUF_REMAINING_WORK.md](PROTOBUF_REMAINING_WORK.md).

---

## Quick reference (code)

| Concern | Location |
|---------|----------|
| Regenerate Go | `make proto` ([PROTOBUF_USAGE.md](PROTOBUF_USAGE.md)) |
| Parse tool args | `internal/tools/protobuf_helpers.go`, mcp-go-core `request.ParseRequest` |
| Task metadata | `internal/database/tasks.go`, `internal/models/todo2_protobuf.go` |
| Config | `internal/config/protobuf.go`, `loader.go` |
| Memory files | `internal/tools/memory.go` |
| Bridge | `internal/bridge/python.go`, `bridge/execute_tool.py` |
