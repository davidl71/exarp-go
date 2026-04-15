# Protobuf Integration

**Status:** Implemented  
**Tasks:** T-1768316817909, T-1768317405631, T-1768319001461

## Overview

exarp-go uses Protocol Buffers for:
- **Tool request/response** parsing (task_workflow, task_analysis, etc.)
- **Config** serialization
- **Todo2** task serialization (binary/JSON round-trip)

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
