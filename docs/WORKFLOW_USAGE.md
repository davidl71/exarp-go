# Development Workflow Usage Guide

**Date:** 2026-01-07  
**Purpose:** Quick reference for using the automated development workflow

---

## 🚀 Quick Start

**For daily development, use this ONE command:**

```bash
make dev-full
```

This starts:
- ✅ Auto-reload server on file changes
- ✅ Auto-run tests on file changes  
- ✅ Auto-generate coverage reports
- ✅ Continuous feedback loop

**Just run it once, then edit files - everything happens automatically!**

---

## 📋 Available Commands

### Development Modes

| Command | What It Does | When to Use |
|---------|--------------|-------------|
| `make dev-full` | Full automation (watch + test + coverage) | **Daily development** |
| `make dev-test` | Watch + auto-test (no coverage) | Faster feedback, less overhead |
| `make dev` | Watch only (auto-reload) | When you don't need tests |
| `make test-watch` | Test-only watch mode | Testing without server |

### Quick Commands

| Command | Purpose |
|---------|---------|
| `make quick-test` | Fast import verification |
| `make quick-build` | Verify imports work |
| `make test` | Run all tests once |
| `make test-coverage` | Run tests with coverage |

### Code Quality

| Command | Purpose |
|---------|---------|
| `make fmt` | Format code |
| `make lint` | Check code quality |
| `make lint-fix` | Auto-fix linting issues |

---

## 🎯 Recommended Workflow

### Starting Your Day

```bash
# 1. Start full development mode
make dev-full

# 2. Edit files - everything happens automatically!
# - Server auto-reloads
# - Tests auto-run
# - Coverage updates
```

### During Development

- **Edit files** - Changes trigger automatic reloads and tests
- **Watch terminal** - See test results and coverage updates
- **Fix issues** - Tests re-run automatically when you save

### Before Committing

```bash
# Stop dev-full (Ctrl+C)

# Run full test suite with coverage
make test-coverage

# Format and lint
make fmt
make lint-fix

# Commit your changes
```

---

## 🔧 Troubleshooting

### Server Not Reloading

```bash
# Check if dev.sh is running
ps aux | grep dev.sh

# Restart manually
make dev-full
```

### Tests Not Running

```bash
# Run tests manually to check for errors
make test

# Check test output
make test-coverage
```

### File Watcher Issues

The script automatically detects the best file watching method:
- **macOS**: Uses `fswatch` (install: `brew install fswatch`)
- **Linux**: Uses `inotifywait` (usually pre-installed)
- **Fallback**: Uses polling (works everywhere, less efficient)

---

## 📚 More Information

- **Full Documentation**: [docs/DEV_TEST_AUTOMATION.md](DEV_TEST_AUTOMATION.md)
- **Quick Summary**: [docs/STREAMLINED_WORKFLOW_SUMMARY.md](STREAMLINED_WORKFLOW_SUMMARY.md)
- **Makefile**: See `make help` for all targets
- **Script**: See `./dev.sh --help` for script options

---

## ✅ Benefits

**Before (Manual):**
- Edit file → Run server → Edit file → Run tests → Repeat...

**Now (Automated):**
- Run `make dev-full` once → Edit files → Everything happens automatically!

**Time Saved:** ~80% reduction in manual commands during development.

