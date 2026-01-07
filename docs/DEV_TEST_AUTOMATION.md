# Development & Testing Automation

**Date:** 2026-01-07  
**Purpose:** Streamline development and testing workflow to minimize manual prompting

---

## Overview

This document describes the automated development and testing workflow that eliminates the need for manual prompting during development. The system provides:

1. **Auto-reload** - Server automatically restarts on file changes
2. **Auto-test** - Tests run automatically on file changes
3. **Auto-build** - Build verification on changes
4. **Coverage tracking** - Automatic coverage reports
5. **Intelligent watching** - Smart file change detection

---

## Quick Start

**🚀 RECOMMENDED: Start with `make dev-full` for full automation!**

### Full Development Mode (Recommended)

```bash
# Full automation: auto-reload + auto-test + coverage
make dev-full

# This is the ONE command you need for daily development!
# Just run it once, then edit files - everything happens automatically.
```

### Basic Development Mode

```bash
# Auto-reload server on file changes
make dev

# Or use the script directly
./dev.sh --watch
```

### Development with Auto-Testing

```bash
# Auto-reload + auto-test on changes
make dev-test

# Or use the script directly
./dev.sh --watch --test
```

### Full Development Mode

```bash
# Auto-reload + auto-test + coverage
make dev-full

# Or use the script directly
./dev.sh --watch --test --coverage
```

### Test-Only Watch Mode

```bash
# Only run tests on changes (no server)
./dev.sh --test-watch
```

---

## Makefile Targets

### Development Targets

| Target | Description | Command |
|--------|-------------|---------|
| `make dev` | Start dev mode (auto-reload) | `./dev.sh --watch` |
| `make dev-test` | Dev mode with auto-test | `./dev.sh --watch --test` |
| `make dev-full` | Full dev mode (watch + test + coverage) | `./dev.sh --watch --test --coverage` |

### Testing Targets

| Target | Description | Command |
|--------|-------------|---------|
| `make test` | Run all tests | `pytest tests/` |
| `make test-watch` | Test watch mode | `./dev.sh --test-watch` |
| `make test-coverage` | Tests with coverage | `pytest --cov` |
| `make test-html` | HTML coverage report | `pytest --cov --cov-report=html` |
| `make test-tools` | Test tool imports | Quick import test |
| `make test-prompts` | Test prompt imports | Quick import test |
| `make test-resources` | Test resource imports | Quick import test |
| `make test-all` | Run all import tests | All quick tests |

### Code Quality Targets

| Target | Description | Command |
|--------|-------------|---------|
| `make fmt` | Format code | `ruff format .` |
| `make lint` | Lint code | `ruff check .` |
| `make lint-fix` | Lint and auto-fix | `ruff check --fix .` |

### Quick Commands

| Target | Description |
|--------|-------------|
| `make quick-test` | Quick test (tools only) |
| `make quick-dev` | Quick dev mode (watch only) |
| `make quick-build` | Quick build (verify imports) |

---

## Development Script (`dev.sh`)

### Options

```bash
./dev.sh [OPTIONS]
```

**Options:**
- `--watch` - Watch files and auto-reload server
- `--test` - Auto-run tests on file changes
- `--test-watch` - Watch mode for tests only (no server)
- `--coverage` - Generate coverage reports
- `--build` - Auto-rebuild on changes
- `--quiet` - Suppress non-error output
- `--help, -h` - Show help message

### Examples

```bash
# Basic auto-reload
./dev.sh --watch

# Auto-reload + auto-test
./dev.sh --watch --test

# Test-only watch mode
./dev.sh --test-watch

# Full dev mode with coverage
./dev.sh --watch --test --coverage

# Quiet mode (minimal output)
./dev.sh --watch --test --quiet
```

### File Watching

The script automatically detects the best file watching method:

1. **fswatch** (macOS) - Fast, efficient
2. **inotifywait** (Linux) - Native Linux support
3. **Polling** (Fallback) - Works everywhere, less efficient

**Watched Files:**
- `*.py` - Python source files
- `pyproject.toml` - Project configuration
- `uv.lock` - Dependency lock file
- `mcp_stdio_tools/` - Source directory

---

## Workflow Examples

### Daily Development

```bash
# Start full dev mode
make dev-full

# Edit files...
# - Server auto-reloads
# - Tests auto-run
# - Coverage updates

# Press Ctrl+C to stop
```

### Testing Only

```bash
# Watch for changes and test
make test-watch

# Edit test files...
# - Tests auto-run on save
```

### Quick Verification

```bash
# Quick build check
make quick-build

# Quick test
make quick-test

# Quick dev (watch only)
make quick-dev
```

---

## Integration with Model-Assisted Testing

### Future: Intelligent Test Generation

When model-assisted testing is implemented:

```bash
# Auto-generate tests for changed code
make test-models

# Test task breakdown
make test-breakdown

# Test auto-execution
make test-auto-exec
```

### Model-Assisted Test Workflow

1. **Code changes detected** → Model analyzes changes
2. **Test suggestions** → Model suggests test cases
3. **Auto-generation** → Tests generated automatically
4. **Auto-execution** → Tests run automatically
5. **Results** → Coverage and results displayed

---

## Continuous Integration

### Pre-Commit Hooks

Add to `.git/hooks/pre-commit`:

```bash
#!/bin/bash
make lint-fix
make test
```

### GitHub Actions

Example workflow (`.github/workflows/ci.yml`):

```yaml
name: CI

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-python@v4
      - run: make install
      - run: make test-coverage
      - run: make lint
```

---

## Performance

### File Watching Performance

| Method | Latency | CPU Usage | Platform |
|--------|---------|-----------|----------|
| fswatch | < 100ms | Low | macOS |
| inotifywait | < 100ms | Low | Linux |
| Polling | 2s | Medium | All |

### Test Execution Time

- **Quick test:** < 1 second
- **Full test suite:** 5-10 seconds
- **With coverage:** 10-20 seconds

---

## Troubleshooting

### Server Not Reloading

1. Check file watcher is running: `ps aux | grep dev.sh`
2. Verify file changes are detected: Check logs
3. Try manual restart: `make run`

### Tests Not Running

1. Check test files exist: `ls tests/`
2. Verify pytest is installed: `uv run pytest --version`
3. Run tests manually: `make test`

### Coverage Not Generating

1. Install coverage: `uv add --dev pytest-cov`
2. Run with coverage: `make test-coverage`
3. Check HTML report: `open htmlcov/index.html`

---

## Best Practices

### Development Workflow

1. **Start with `make dev-full`** - Full automation from the start
2. **Edit files** - Changes trigger auto-reload and tests
3. **Check coverage** - Ensure tests cover new code
4. **Fix issues** - Auto-test catches problems immediately

### Testing Strategy

1. **Quick tests first** - Use `make quick-test` for fast feedback
2. **Full tests before commit** - Use `make test-coverage`
3. **Coverage goals** - Maintain >80% coverage

### Code Quality

1. **Auto-format** - Use `make fmt` before committing
2. **Auto-lint** - Use `make lint-fix` to fix issues
3. **Manual review** - Review auto-fixes before committing

---

## Future Enhancements

### Planned Features

1. **Model-Assisted Testing**
   - Auto-generate tests for new code
   - Suggest test improvements
   - Identify untested code paths

2. **Intelligent Test Selection**
   - Run only affected tests
   - Parallel test execution
   - Test result caching

3. **Performance Monitoring**
   - Track test execution time
   - Identify slow tests
   - Optimize test suite

4. **Integration with IDE**
   - VS Code integration
   - Cursor integration
   - Real-time test results

---

## See Also

- [Makefile](../Makefile) - All available targets
- [dev.sh](../dev.sh) - Development script
- [DEVWISDOM_GO_LESSONS.md](./DEVWISDOM_GO_LESSONS.md) - Lessons from devwisdom-go
- [MODEL_ASSISTED_WORKFLOW.md](./MODEL_ASSISTED_WORKFLOW.md) - Model integration

---

**Status:** Ready for use. Start with `make dev-full` for full automation.

