# Protobuf Implementation Status and Plan

*Generated from implementation review (2026-01-29).*

## Current Implementation (What’s Done)

### 1. Proto schemas and generated code

- **proto/tools.proto** – Request/response messages for memory, context, report, task_workflow, health, security, testing, automation, session, lint, estimation, git_tools, and others. Generated `proto/tools.pb.go`.
- **proto/todo2.proto** – Todo2Task, Todo2State. Generated `proto/todo2.pb.go`. Used by `internal/models/todo2_protobuf.go` and database task metadata.
- **proto/config.proto** – FullConfig and nested config messages. Generated `proto/config.pb.go`.
- **proto/bridge.proto** – ToolRequest, ToolResponse for Go↔Python bridge. Generated `proto/bridge.pb.go`; Python `bridge/proto/bridge_pb2.py`.

### 2. Tool request parsing (mcp-go-core)

- **internal/tools/protobuf_helpers.go** – All tool requests use `request.ParseRequest[T]()` from mcp-go-core (Memory, Context, Report, TaskWorkflow, Health, Security, InferSessionMode, ToolCatalog, WorkflowMode, Estimation, Session, GitTools, MemoryMaint, TaskAnalysis, etc.). Protobuf-first, JSON fallback. No remaining duplicate ParseXRequest logic to migrate.

### 3. Handlers

- **internal/tools/handlers.go** – Every handler uses “try protobuf first, fall back to JSON” and the shared ParseXRequest + XRequestToParams pattern. Handlers already use protobuf request types where available.

### 4. Config

- **internal/config/protobuf.go** – ToProtobuf / FromProtobuf for FullConfig and nested structs; duration and zero-value handling.
- **internal/config/loader.go** – LoadConfig prefers `.exarp/config.pb`, then YAML. LoadConfigProtobuf reads binary protobuf.
- **internal/cli/config.go** – `config export protobuf`, `config convert yaml↔protobuf`.

### 5. Database (task metadata)

- **migrations/003_add_protobuf_support.sql** – Adds `metadata_protobuf` (BLOB), `metadata_format` (TEXT), index. Already applied in sequence (001 → 002 → 003 → 004 → 005).
- **internal/database/tasks.go** – CreateTask/UpdateTask serialize task metadata via `models.SerializeTaskToProtobuf`; GetTask/ListTasks deserialize from protobuf when `metadata_format='protobuf'`. JSON fallback for legacy rows and if protobuf columns are missing.
- **internal/models/todo2_protobuf.go** – Todo2TaskToProto, ProtoToTodo2Task, SerializeTaskToProtobuf, DeserializeTaskFromProtobuf. Benchmarks in `internal/models/todo2_protobuf_bench_test.go`.

### 6. Memory tool (file format)

- **internal/tools/memory.go** – saveMemory writes `.pb` (protobuf binary). LoadAllMemories reads `.pb` first, then `.json`. deleteMemoryFile tries `.pb` then `.json`. Memory↔proto in protobuf_helpers (MemoryToProto, ProtoToMemory, SerializeMemoryToProtobuf, DeserializeMemoryFromProtobuf).

### 7. Python bridge

- **internal/bridge/python.go** – Builds proto.ToolRequest, marshals to binary, invokes Python with `--protobuf`, stdin = protobuf; reads stdout as proto.ToolResponse or JSON.
- **bridge/execute_tool.py** – Accepts `--protobuf`, reads stdin as protobuf, uses bridge_pb2.ToolRequest/ToolResponse when HAVE_PROTOBUF. Full protobuf path implemented.

---

## Schema version constant

- **internal/database/schema.go** – `SchemaVersion` updated to **5** (matches migrations 001–005). No duplicate migration numbers; no file renames needed.

---

## Remaining / optional work

| Area | Priority | Notes |
|------|----------|--------|
| ~~**SchemaVersion constant**~~ | ~~High~~ | Done: set to 5 in schema.go. |
| ~~**Report/scorecard internals**~~ | ~~Low~~ | Done: overview uses aggregateProjectDataProto + formatOverview*Proto; scorecard uses GoScorecardResultToProto + ProtoToScorecardMap (commit a7a8aea). |
| **Protobuf build tooling / Ansible** | Medium | Document or automate protoc/Makefile/Ansible for proto generation if not already. |
| **Documentation** | Medium | Update docs to describe protobuf usage (config, tasks, memory, bridge). T1.5.5. |
| **Benchmarks / tests** | Low | Run and document protobuf vs JSON benchmarks; add/run tests for protobuf serialization with real tasks. |

---

## Task list changes (from review)

- **Mark Done (implemented):** Migrate memory to protobuf file format; Create protobuf schemas (tools + high-priority); Update tool handlers to protobuf; T1.5.1 Conversion Layer; T1.5.2 Loader; T1.5.3 CLI; Run migration 003; Update ListTasks/DB to protobuf; Phase 1 Task Management protobuf; Migrate to mcp-go-core request parsing; Migrate Python bridge to protobuf.
- **Update:** “Fix schema versioning and duplicate migration versions” → scope to “Update SchemaVersion constant from 3 to 5” (no 002→003 rename).
- **Done:** Migrate report/scorecard to protobuf (T-1768318391643) — overview and scorecard now use proto internally.
- **Keep:** Measure performance (benchmarks); Test protobuf serialization with real tasks; Set up protobuf build tooling/Ansible; T1.5.5 docs; Protobuf Integration Epic (update description to reflect status).
- **Obsolete/remove:** Any task that still says “Rename 002_add_protobuf_support to 003” (already 003).

---

## Quick reference

- **Proto files:** proto/*.proto → proto/*.pb.go (Go), bridge/proto/bridge_pb2.py (Python).
- **Parsing:** internal/tools/protobuf_helpers.go + mcp-go-core request.ParseRequest[T].
- **Config:** internal/config/protobuf.go, loader.go; CLI config export/convert.
- **Tasks DB:** internal/database/tasks.go + internal/models/todo2_protobuf.go; migration 003.
- **Memory:** internal/tools/memory.go (save .pb, load .pb/.json).
- **Bridge:** internal/bridge/python.go + bridge/execute_tool.py (--protobuf).
