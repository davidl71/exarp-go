# Protobuf Integration

**Status:** Implemented  
**Tasks:** T-1768316817909, T-1768317405631, T-1768319001461  

**Canonical doc map:** [PROTOBUF_IMPLEMENTATION_STATUS.md](PROTOBUF_IMPLEMENTATION_STATUS.md) (this file = **where** proto plugs into build, Ansible, and tests).

## Overview

exarp-go uses Protocol Buffers for:
- **Tool request** parsing (`internal/tools/protobuf_helpers.go`, `WrapHandler` in `handlers_wrap.go`; **mcp-go-core** `request.ParseRequest[T]()` for protobuf/JSON dual decode)
- **Config** serialization
- **Todo2** task metadata (binary/JSON round-trip in DB)

## Build Tooling

### Makefile Targets

| Target | Description |
|--------|-------------|
| `make proto` | Generate Go from .proto using protoc |
| `make proto-buf` | Generate using buf (falls back to proto if buf unavailable) |
| `make proto-check` | Validate .proto syntax |
| `make proto-clean` | Remove generated code |
| `make install-tools` | Install protoc-gen-go |

### buf.yaml

Uses buf for linting and generation:
- Remote plugin: `buf.build/protocolbuffers/go`
- Output: `proto/` with `paths=source_relative`
- Lint: DEFAULT rules

### Ansible (golang role)

- **protoc**: Installed via apt (Debian) or Homebrew (macOS)
- **protoc-gen-go**: Installed via `go install`

## Proto Files

| File | Purpose |
|------|---------|
| `proto/todo2.proto` | Todo2 task messages |
| `proto/tools.proto` | Tool request/response (task_workflow, etc.) |
| `proto/config.proto` | Config schema |
| `proto/bridge.proto` | Bridge messages |

## Testing

- **Unit tests:** `internal/models/todo2_protobuf_test.go` – mock tasks
- **Integration:** `internal/tools/protobuf_integration_test.go` – real Todo2 tasks
  - Loads from .todo2 (DB or JSON)
  - Serializes first 10 tasks to protobuf and back
  - Verifies round-trip
  - Skips if PROJECT_ROOT not set or no tasks

## Handler Integration

Handlers use `Parse*Request` helpers that:
1. Try protobuf unmarshal first
2. Fall back to JSON for backward compatibility
3. Convert to params map for native handlers

See `internal/tools/protobuf_helpers.go` for converters.
