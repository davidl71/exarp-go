# Protobuf usage and build tooling

exarp-go uses Protocol Buffers for config (`.exarp/config.pb`), task metadata, memory, tool request/response (`tools.proto`, `todo2.proto`, `bridge.proto`), and the Go↔Python bridge.

**Document map:** [PROTOBUF_IMPLEMENTATION_STATUS.md](PROTOBUF_IMPLEMENTATION_STATUS.md) lists all protobuf-related docs and their roles (this file = **build / regenerate**).

## Regenerating Go code from .proto files

**Using protoc (Makefile):**

```bash
make proto
```

Requires `protoc` and `protoc-gen-go` on PATH. Install:

- macOS: `brew install protobuf`; `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`
- Linux: `apt-get install protobuf-compiler` (or equivalent); install protoc-gen-go as above

**Using buf (optional):**

```bash
make proto-buf
```

Uses [buf](https://buf.build/) if available; otherwise falls back to `make proto`.

**Generated files:** `proto/tools.pb.go`, `proto/todo2.pb.go`, `proto/config.pb.go`, `proto/bridge.pb.go`. Do not edit these by hand; change the `.proto` sources and re-run `make proto`.

## Validation and cleanup

```bash
make proto-check   # Validate .proto syntax without generating code
make proto-clean   # Remove generated .pb.go files (use before regenerating if needed)
```

## References

- [PROTOBUF_IMPLEMENTATION_STATUS.md](PROTOBUF_IMPLEMENTATION_STATUS.md) — Canonical status, code map, action plan
- [archive/protobuf/README.md](archive/protobuf/README.md) — Historical protobuf research & snapshots
- [PROTOBUF_REMAINING_WORK.md](PROTOBUF_REMAINING_WORK.md) — Optional backlog (no duplicate checklists)
- [PROTOBUF_INTEGRATION.md](PROTOBUF_INTEGRATION.md) — Makefile, buf, Ansible, tests
- [CONFIGURATION_REFERENCE.md](CONFIGURATION_REFERENCE.md) — Config (including protobuf export/convert)
