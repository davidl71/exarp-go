# AGENTS.md

## Cursor Cloud specific instructions

This is a **Go library** (not a standalone application). There are no services to start, no databases, and no Docker dependencies.

### Quick reference

| Action | Command |
|--------|---------|
| Build | `go build ./...` |
| Test | `go test ./... -v` |
| Lint | `golangci-lint run ./...` |
| Vet | `go vet ./...` |
| Format | `go fmt ./...` |
| Full check | `make check` |

See `Makefile` for all targets (`make help`).

### Pre-existing build issues

Several packages have pre-existing compilation errors and will fail `go build ./...` / `go test ./...`:
- `pkg/mcp/platform` — redeclared types (`OS`, `Architecture`)
- `pkg/mcp/client` — API drift with `mcp-golang` dependency
- `pkg/mcp/cli` (tests only) — `GetFlag` signature mismatch in tests
- `pkg/mcp/factory` (tests only) — nil pointer dereference in `TestNewServerFromConfig/nil_config`
- `examples/` — various unused imports and API mismatches

Core packages that build and test cleanly: `config`, `logging`, `security`, `protocol`, `types`, `request`, `response`, `framework` (builds but has vet warning in tests).

### Tooling

- Go 1.24+ is required (`go.mod` specifies `go 1.24.0`, toolchain `go1.24.11`).
- `golangci-lint` is installed at `~/go/bin/golangci-lint`. Ensure `~/go/bin` is on `PATH`.
- `go mod tidy` may be needed after fresh clone if `go.sum` is incomplete.
