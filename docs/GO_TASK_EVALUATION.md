# go-task/task Evaluation for exarp-go

**Date:** 2026-02-20
**Status:** Evaluation Complete
**Recommendation:** Hybrid adoption (Taskfile alongside Makefile, phased migration)

## Executive Summary

[go-task/task](https://github.com/go-task/task) (v3.48.0, 14.9k stars, MIT license) is a modern YAML-based task runner that could replace or complement the existing 1349-line Makefile. After analyzing feature parity, DX, and migration cost, the recommendation is a **phased hybrid approach**: introduce a `Taskfile.yml` for the most-used developer workflows while keeping the Makefile for CI and backward compatibility.

---

## 1. Current Makefile Analysis

### Scale

- **1349 lines**, ~100+ targets across 18 `.PHONY` groups
- 6 categories: build, test, quality, sprint automation, task management, exarp-go tools
- Heavy use of shell conditionals (platform detection, CGO, tool detection)
- `make config` generates `.make.config` for tool availability

### Complexity Hotspots

| Area | Lines | Complexity |
|------|-------|------------|
| `build` target (CGO/platform detection) | ~80 | High — nested `if/else` for Darwin/arm64/CGO fallback |
| `build-apple-fm` + `build-swift-bridge` | ~90 | High — vendor detection, Swift compilation |
| Sprint automation targets | ~60 | Medium — exarp-go binary dependency |
| Task management targets | ~80 | Medium — variable passing, grep filtering |
| Code quality (`fmt`/`lint`/`check`) | ~100 | Medium — tool detection, exarp-go delegation |
| Ansible targets | ~30 | Low |

### Shortcut Aliases

The project relies heavily on `make b`, `make p`, `make st`, etc. — documented in `.cursor/rules/make-shortcuts.mdc`. Any migration must preserve these ergonomics.

---

## 2. Feature Comparison

### Full Parity

| Feature | Make (current) | go-task equivalent | Verdict |
|---------|---------------|-------------------|---------|
| `CGO_ENABLED=1 go build -ldflags "..." -o bin/exarp-go ./cmd/server` | Inline shell | `env: {CGO_ENABLED: '1'}` + `cmds:` | **Full parity** |
| Dynamic vars (`git describe`, `date`) | `$(shell cmd)` | `sh: cmd` under `vars:` | **Full parity** |
| Platform conditional (Darwin + arm64) | Nested `if/else` in shell | `platforms: [darwin/arm64]` or `if:` | **Cleaner** |
| `.env` file loading | Manual `include` | Native `dotenv:` | **Better** |
| Colored output / ANSI | Manual `\033[0;32m` constants | Built-in + customizable | **Better** |
| `make -j N` parallel | Opt-in flag | Default for `deps:` | **Better** |
| File change detection | Timestamp only | Checksum (default) | **Better** |
| Watch mode | External (`fswatch`/`inotifywait`) | Built-in `--watch` | **Better** |
| Self-documenting help | Manual `grep -hE` target | Built-in `task --list` | **Better** |
| Task dependencies | Implicit via file targets | Explicit `deps:` | **Equivalent** |
| Multi-file organization | `include file.mk` (flat) | `includes:` with namespacing | **Better** |
| IDE integration | None | VS Code/Cursor extension + schema | **Better** |

### Gaps (Make can, go-task cannot easily)

| Feature | Impact on exarp-go |
|---------|-------------------|
| Pattern rules (`%.o: %.c`) | **None** — project doesn't use compiled C targets |
| Ubiquity (pre-installed) | **Low** — team is small, install is `brew install go-task` |
| `.make.config` include with conditional tool detection | **Medium** — would need redesign using `preconditions:` or `status:` |
| `define` macros (`exarp_run`, `exarp_ensure_binary`) | **Low** — replaceable with `internal:` helper tasks |
| Recursive Make (`$(MAKE) target`) | **Low** — `task: target` is the equivalent |

---

## 3. DX Comparison

### Before (Makefile)

```makefile
build: ## Build the Go server
	$(Q)echo "$(BLUE)Building $(PROJECT_NAME) v$(VERSION)...$(NC)"
	@if ! command -v $(GO) >/dev/null 2>&1 && [ ! -x "$(GO)" ]; then \
		echo "$(RED)❌ Go not found$(NC)"; exit 1; \
	fi
	@if [ "$$(uname -s)" = "Darwin" ] && [ "$$(uname -m)" = "arm64" ]; then \
		if command -v gcc >/dev/null 2>&1 || command -v clang >/dev/null 2>&1; then \
			CGO_ENABLED=1 $(GO) build -ldflags "..." -o $(BINARY_PATH) ./cmd/server 2>&1 || \
			(CGO_ENABLED=0 $(GO) build ... || exit 1); \
		else \
			CGO_ENABLED=0 $(GO) build ... || exit 1; \
		fi; \
	else \
		CGO_ENABLED=0 $(GO) build ... || exit 1; \
	fi
```

### After (Taskfile.yml)

```yaml
tasks:
  build:
    desc: Build the Go server
    aliases: [b]
    deps: [swift-bridge]
    vars:
      CGO_FLAG: '{{if and (eq OS "darwin") (eq ARCH "arm64")}}1{{else}}0{{end}}'
    env:
      CGO_ENABLED: '{{.CGO_FLAG}}'
    cmds:
      - go build -ldflags "{{.LDFLAGS}}" -o bin/exarp-go ./cmd/server
    sources:
      - '**/*.go'
      - go.mod
    generates:
      - bin/exarp-go
    preconditions:
      - sh: command -v go
        msg: "Go not found. Install Go first."

  swift-bridge:
    internal: true
    platforms: [darwin/arm64]
    status:
      - test -f vendor/github.com/blacktop/go-foundationmodels/libFMShim.a
    cmds:
      - swiftc -sdk $(xcrun --show-sdk-path) ...
```

**DX improvements:**
- 80 lines of shell → ~25 lines of YAML
- Platform logic is declarative, not imperative
- `sources:` / `generates:` give free incremental builds
- `aliases: [b]` preserves `task b` shortcut
- `preconditions:` with messages replace manual error handling

---

## 4. Migration Strategy: Phased Hybrid

### Phase 1: Developer Workflows (Week 1)

Create `Taskfile.yml` covering the most-used targets:

| Makefile target | Taskfile task | Notes |
|-----------------|--------------|-------|
| `build` / `b` | `build` (alias: `b`) | With incremental builds |
| `test` / `test-go` | `test` | Parallel by default |
| `fmt` | `fmt` | Direct go-fmt |
| `lint` / `lint-fix` | `lint`, `lint:fix` | Via exarp-go or golangci-lint |
| `dev` / `dev-watch` | `dev` | Built-in watch mode |
| `tidy` | `tidy` | go mod tidy + vendor |

**Backward compatibility:** Keep `Makefile` fully functional. Developers choose which to use.

### Phase 2: Task Management + Quality (Week 2-3)

Add remaining workflows:

| Category | Tasks |
|----------|-------|
| Task management | `task:list`, `task:show`, `task:update`, `task:create` |
| Code quality | `check`, `check:fix`, `check:all`, `scorecard` |
| Sprint automation | `sprint:start`, `sprint:end`, `pre-sprint` |
| CI/CD | `ci`, `validate`, `pre-commit`, `pre-release` |

Use `includes:` to split into `taskfiles/Build.yml`, `taskfiles/Test.yml`, etc.

### Phase 3: CI Migration + Makefile Deprecation (Week 4+)

- Update `.github/workflows/` to use `task` commands
- Update `.cursor/rules/make-shortcuts.mdc` to reference both
- Add deprecation notices to Makefile targets
- Eventually make Makefile a thin wrapper: each target calls `task <name>`

### Rollback Strategy

- `Taskfile.yml` and `Makefile` coexist indefinitely
- CI runs Makefile targets (proven, stable)
- Developers can use either
- Remove Makefile only when team is fully on Taskfile (no hard deadline)

---

## 5. Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Team unfamiliarity with YAML task syntax | Medium | Low | Good docs, Go template syntax is learnable |
| Subtle behavior differences from Make | Low | Medium | Parallel rollout, both systems available |
| Watch mode edge cases with long processes | Low | Low | Use `go build && ./binary` pattern |
| CI/CD migration breaks | Low | High | Keep Makefile for CI until Taskfile is proven |
| go-task project abandoned | Very Low | High | 14.9k stars, MIT, active team; fallback to Make |

---

## 6. Installation

```bash
# macOS
brew install go-task

# Go install
go install github.com/go-task/task/v3/cmd/task@latest

# Direct download
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b ~/.local/bin
```

Cursor/VS Code extension: [task.vscode-task](https://marketplace.visualstudio.com/items?itemName=task.vscode-task)

Schema validation: Add Red Hat YAML extension + configure `yaml.schemas` for `taskfile.dev/schema.json`.

---

## 7. Prototype: Minimal Taskfile.yml

```yaml
# yaml-language-server: $schema=https://taskfile.dev/schema.json
version: '3'

dotenv: ['.env']

vars:
  PROJECT_NAME: exarp-go
  BINARY_PATH: bin/exarp-go
  VERSION:
    sh: git describe --tags --always --dirty 2>/dev/null || echo "dev"
  BUILD_TIME:
    sh: date -u +"%Y-%m-%dT%H:%M:%SZ"
  GIT_COMMIT:
    sh: git rev-parse --short HEAD 2>/dev/null || echo "unknown"
  LDFLAGS: >-
    -X main.Version={{.VERSION}}
    -X main.BuildTime={{.BUILD_TIME}}
    -X main.GitCommit={{.GIT_COMMIT}}
  CGO_FLAG: '{{if and (eq OS "darwin") (eq ARCH "arm64")}}1{{else}}0{{end}}'

tasks:
  default:
    desc: Build and run sanity checks
    deps: [build, sanity-check]

  build:
    desc: Build the Go server
    aliases: [b]
    env:
      CGO_ENABLED: '{{.CGO_FLAG}}'
    cmds:
      - go build -ldflags "{{.LDFLAGS}}" -o {{.BINARY_PATH}} ./cmd/server
    sources:
      - '**/*.go'
      - go.mod
      - go.sum
    generates:
      - '{{.BINARY_PATH}}'
    preconditions:
      - sh: command -v go
        msg: "Go not found. Install Go first."

  test:
    desc: Run all Go tests
    env:
      CGO_ENABLED: '0'
    cmds:
      - go test ./... -parallel 4

  test:fast:
    desc: Fast test (pre-push)
    env:
      CGO_ENABLED: '0'
    cmds:
      - go test ./... -parallel 4 -short

  fmt:
    desc: Format Go code
    cmds:
      - go fmt ./...

  lint:
    desc: Lint code
    deps: [build]
    cmds:
      - '{{.BINARY_PATH}} -tool lint -args ''{"action":"run","linter":"auto","path":"."}'''''

  lint:fix:
    desc: Lint and auto-fix
    deps: [build]
    cmds:
      - '{{.BINARY_PATH}} -tool lint -args ''{"action":"run","linter":"auto","path":".","fix":true}'''''

  tidy:
    desc: Clean up dependencies
    cmds:
      - go mod tidy
      - cmd: go mod vendor
        if: test -d vendor

  dev:
    desc: Development mode (watch + rebuild)
    watch: true
    sources:
      - '**/*.go'
    cmds:
      - task: build

  sanity-check:
    desc: Verify tools/resources/prompts counts
    cmds:
      - go build -o bin/sanity-check cmd/sanity-check/main.go
      - ./bin/sanity-check

  clean:
    desc: Clean build artifacts
    cmds:
      - rm -f coverage-go.out coverage-go.html
      - rm -f bin/sanity-check bin/migrate

  # Git shortcuts
  push:
    desc: Git push
    aliases: [p]
    cmds:
      - git push

  pull:
    desc: Git pull
    aliases: [pl]
    cmds:
      - git pull

  status:
    desc: Git status
    aliases: [st]
    cmds:
      - git -c advice.statusHints=false status
```

---

## 8. Verdict

| Criterion | Score (1-5) | Notes |
|-----------|------------|-------|
| Feature parity | 4.5 | All exarp-go needs are covered; minor gaps in macro system |
| DX improvement | 4.0 | YAML is cleaner; templates slightly verbose |
| Migration effort | 3.5 | ~2-3 days for Phase 1, ~1 week for full |
| Risk | 4.0 | Low risk with hybrid approach |
| Long-term maintainability | 4.5 | Declarative > imperative for build systems |
| **Overall** | **4.1** | **Recommended with phased hybrid adoption** |

### Decision: Proceed with Phase 1

Create `Taskfile.yml` with core developer workflows (build, test, fmt, lint, dev). Keep Makefile for CI and advanced targets. Evaluate after 2 weeks of usage before proceeding to Phase 2.

---

## References

- [go-task/task GitHub](https://github.com/go-task/task) — 14.9k stars, v3.48.0
- [Taskfile.dev Documentation](https://taskfile.dev/)
- [VS Code Extension](https://marketplace.visualstudio.com/items?itemName=task.vscode-task)
- [Schema](https://taskfile.dev/schema.json)
- [LLM-friendly docs](https://taskfile.dev/llms-full.txt)
- [DevOpsian: Why I Switched from Makefile to Taskfile](https://devopsian.net/p/why-i-switched-from-makefile-to-taskfile/)
- [AppliedGo: Just Make a Task](https://appliedgo.net/spotlight/just-make-a-task/)
