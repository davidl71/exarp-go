# Project Structure and File Organization

**Purpose:** Separate build artifacts, temporary files, and runtime data from source code

---

## Current Structure

```
exarp-go/
├── bin/              # Build artifacts (binaries) - IGNORED ✅
├── .exarp/           # Runtime state (workflow mode, memories) - IGNORED ✅
│   ├── memories/     # Memory storage
│   └── workflow_mode.json
├── .todo2/           # Todo2 state - IGNORED ✅
├── test-results/     # Test output - IGNORED ✅
├── coverage-report/  # Coverage reports - IGNORED ✅
├── coverage.out      # Coverage data - IGNORED ✅
├── docs/             # Documentation (some generated reports here)
├── internal/         # Source code
├── cmd/              # Source code
└── ...
```

---

## File Categories

### ✅ Build Artifacts (Ignored)
- `bin/` - All compiled binaries
- `build/` - Alternative build directory (if used)
- `*.test` - Test binaries
- `*.out` - Coverage and test output files

### ✅ Runtime State (Ignored)
- `.exarp/` - All exarp runtime files
  - `memories/` - Memory storage
  - `workflow_mode.json` - Workflow state
  - (Future: `cache/`, `logs/`, `tmp/` subdirectories)
- `.todo2/` - Todo2 state and database

### ✅ Test Artifacts (Ignored)
- `test-results/` - Test output
- `coverage-report/` - Coverage HTML reports
- `coverage.out` - Coverage data

### ✅ Temporary Files (Ignored)
- `tmp/`, `.tmp/`, `temp/` - Temporary directories
- `*.tmp` - Temporary files
- `*.log` - Log files

### ✅ Python Artifacts (Ignored)
- `.venv/`, `venv/` - Virtual environments
- `__pycache__/` - Python cache
- `*.pyc`, `*.pyo`, `*.pyd` - Compiled Python files

---

## Recommended Organization

### Current Approach (Working Well)

**Build Artifacts:**
- `bin/` - All compiled binaries (gitignored)

**Runtime State:**
- `.exarp/` - All exarp runtime files (gitignored)
  - `memories/` - Memory storage
  - `workflow_mode.json` - Workflow state
  - (Future: `cache/`, `logs/`, `tmp/` subdirectories)

**Task Management:**
- `.todo2/` - Todo2 state and database (gitignored)

**Test Artifacts:**
- `test-results/` - Test output (gitignored)
- `coverage-report/` - Coverage HTML reports (gitignored)
- `coverage.out` - Coverage data (gitignored)

**Temporary Files:**
- `tmp/`, `.tmp/`, `temp/` - Temporary directories (gitignored)

---

## Future Enhancement: Organize `.exarp/` Subdirectories

For better organization, could restructure `.exarp/`:

```
.exarp/
├── state/           # Runtime state files
│   └── workflow_mode.json
├── memories/        # Memory storage (already exists)
│   └── *.json
├── cache/           # Cache files (future)
├── logs/            # Log files (future)
└── tmp/             # Temporary files (future)
```

**Benefits:**
- Better organization within `.exarp/`
- Clear separation of concerns
- Easy to find specific file types

**Migration:** Would require updating code paths, but provides better structure.

---

## .gitignore Coverage

✅ **All Build/Runtime Files Properly Ignored:**

```gitignore
# Exarp runtime directories
.exarp/
.todo2/

# Build artifacts
bin/
build/
*.test
*.out
coverage*.out
coverage*.html
coverage.out

# Test artifacts
test-results/
coverage-report/

# Temporary files
tmp/
.tmp/
temp/
*.tmp
*.log

# Python
.venv/
venv/
__pycache__/
*.pyc
*.pyo
*.pyd
.Python

# Go workspace
go.work
go.work.sum
```

---

## Benefits

1. **Clear Separation:** Build/runtime files separate from source code
2. **Easy Cleanup:** All runtime files in `.exarp/`, build artifacts in `bin/`
3. **Git Friendly:** Everything properly ignored
4. **Consistent:** Matches `.todo2/` pattern
5. **Scalable:** Easy to add new subdirectories

---

## Summary

**Current Status:** ✅ Well organized and properly gitignored

- Build artifacts: `bin/` (ignored)
- Runtime state: `.exarp/` (ignored)
- Test artifacts: `test-results/`, `coverage-report/` (ignored)
- Temporary files: `tmp/`, `.tmp/`, `temp/` (ignored)

**All build environment and temporary files are separated from source code and properly ignored.**
