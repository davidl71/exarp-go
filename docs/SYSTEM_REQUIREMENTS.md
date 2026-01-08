# System Requirements - exarp-go MCP Server

## Disk Space Requirements

### Minimum Installation (Production)

**Runtime Requirements:**
- **Binary:** ~8 MB (compiled `exarp-go` binary)
- **Project files:** ~1 MB (source code, configs, bridge scripts)
- **Logs directory:** ~10-50 MB (varies with usage)
- **Total runtime:** **~20-60 MB**

**With Dependencies:**
- **Go 1.24.0 installation:** ~300-400 MB (`/usr/local/go`)
- **Python 3.10 installation:** ~50-100 MB (system Python)
- **uv (Python package manager):** ~10-20 MB
- **Go module cache:** ~50-200 MB (first build, cached in `$GOPATH/pkg/mod`)
- **Total with dependencies:** **~410-780 MB**

### Development Installation

**Additional Development Space:**
- **Source code:** ~55 MB (current project size)
- **Build artifacts:** ~23 MB (`bin/` directory)
- **Test files:** ~72 KB
- **Documentation:** ~736 KB
- **Ansible configs:** ~124 KB
- **Go build cache:** ~100-300 MB (temporary, can be cleaned)
- **Python virtual environment:** ~20-50 MB (if using `uv sync --dev`)
- **Development tools (linters):** ~50-100 MB
  - golangci-lint: ~30-50 MB
  - shellcheck: ~5-10 MB
  - shfmt: ~2-5 MB
  - markdownlint-cli: ~10-20 MB
  - cspell: ~5-10 MB
- **Total development:** **~250-500 MB additional**

**Full Development Setup:** **~660-1,280 MB**

### Production Deployment (via Ansible)

**System Requirements:**
- **Base OS:** ~2-5 GB (Ubuntu/Debian minimal)
- **Go 1.24.0:** ~300-400 MB
- **Python 3.10:** ~50-100 MB
- **uv:** ~10-20 MB
- **Project deployment:** ~60 MB
- **Systemd service:** <1 KB
- **Logs (initial):** ~10 MB
- **Total production:** **~3.4-5.6 GB** (including OS)

**Production Runtime (excluding OS):** **~430-590 MB**

## Detailed Breakdown

### Current Project Size

```
Project root:           55 MB
├── bin/                23 MB  (compiled binaries)
│   └── exarp-go        7.8 MB (main binary)
├── bridge/             32 KB  (Python bridge scripts)
├── cmd/                32 KB  (Go source)
├── internal/          276 KB (Go source)
├── docs/              736 KB  (documentation)
├── ansible/           124 KB (Ansible configs)
├── tests/             72 KB  (test files)
└── Other files:       ~30 MB (go.mod, go.sum, configs, etc.)
```

### Go Installation

**Go 1.24.0 Download:**
- Linux amd64: ~150-200 MB (tar.gz archive)
- Linux arm64: ~150-200 MB (tar.gz archive)
- macOS amd64: ~150-200 MB (pkg installer)
- macOS arm64: ~150-200 MB (pkg installer)

**Go Installation (extracted):**
- `/usr/local/go`: ~300-400 MB
- Includes: compiler, standard library, tools, documentation

**Go Module Cache (first build):**
- `$GOPATH/pkg/mod`: ~50-200 MB
- Cached dependencies (MCP SDK, etc.)
- Grows with more dependencies

**Go Build Cache:**
- `$GOCACHE`: ~100-300 MB (temporary)
- Can be cleaned: `go clean -cache`

### Python Installation

**Python 3.10:**
- System package: ~50-100 MB
- Includes: interpreter, standard library, pip

**uv (Python package manager):**
- Binary: ~10-20 MB
- Installed via: `curl -LsSf https://astral.sh/uv/install.sh | sh`
- Location: `~/.cargo/bin/uv` or system PATH

**Python Virtual Environment (optional):**
- `.venv/`: ~20-50 MB (if using `uv sync --dev`)
- Only needed for development dependencies

### Development Tools (Optional)

**Linters (development only):**
- golangci-lint: ~30-50 MB
- shellcheck: ~5-10 MB
- shfmt: ~2-5 MB
- gomarkdoc: ~5-10 MB
- markdownlint-cli: ~10-20 MB (via npm)
- cspell: ~5-10 MB (via npm)
- **Total linters:** ~57-105 MB

**File Watchers:**
- fswatch (macOS): ~5-10 MB
- inotify-tools (Linux): ~1-2 MB

## Memory (RAM) Requirements

### Runtime Memory

**Minimum:**
- Binary execution: ~10-20 MB
- Python bridge processes: ~5-10 MB per process
- **Total minimum:** **~15-30 MB**

**Typical Usage:**
- Binary: ~20-50 MB
- Python bridge: ~10-20 MB
- **Total typical:** **~30-70 MB**

**Peak Usage (with all tools active):**
- Binary: ~50-100 MB
- Multiple Python processes: ~50-100 MB
- **Total peak:** **~100-200 MB**

### Development Memory

**Build Process:**
- Go compiler: ~200-500 MB (during build)
- Test execution: ~100-300 MB
- **Build peak:** **~300-800 MB**

## CPU Requirements

### Runtime

**Minimum:**
- Single core, 1.0 GHz
- Suitable for: low-traffic MCP server

**Recommended:**
- Dual core, 2.0+ GHz
- Suitable for: typical usage, multiple concurrent requests

**High Performance:**
- Quad core, 3.0+ GHz
- Suitable for: high-traffic, multiple concurrent tools

### Development

**Recommended:**
- Dual core, 2.5+ GHz
- Faster builds and test execution

## Storage Recommendations

### Production Server

**Minimum Disk Space:**
- **5 GB** free space (including OS)
- Allows for: binary, logs, temporary files

**Recommended Disk Space:**
- **10-20 GB** free space
- Allows for: logs rotation, updates, backups

**High Availability:**
- **50+ GB** free space
- Allows for: extensive logging, multiple versions, backups

### Development Machine

**Minimum Disk Space:**
- **2 GB** free space
- Allows for: source code, build artifacts, dependencies

**Recommended Disk Space:**
- **5-10 GB** free space
- Allows for: full development environment, test data, documentation

## Disk Usage by Component

### Essential Components

| Component | Size | Required |
|-----------|------|----------|
| exarp-go binary | 7.8 MB | ✅ Yes |
| Bridge scripts | 32 KB | ✅ Yes |
| Go source | 308 KB | ✅ Yes (build time) |
| Config files | <1 MB | ✅ Yes |
| **Total essential** | **~9 MB** | **Runtime** |

### Dependencies

| Component | Size | Required |
|-----------|------|----------|
| Go 1.24.0 | 300-400 MB | ✅ Yes (build) |
| Python 3.10 | 50-100 MB | ✅ Yes (runtime) |
| uv | 10-20 MB | ✅ Yes (optional) |
| Go module cache | 50-200 MB | ✅ Yes (first build) |
| **Total dependencies** | **~410-720 MB** | **First install** |

### Development Tools

| Component | Size | Required |
|-----------|------|----------|
| Linters | 57-105 MB | ❌ Optional |
| File watchers | 1-10 MB | ❌ Optional |
| Test dependencies | 20-50 MB | ❌ Optional |
| **Total dev tools** | **~78-165 MB** | **Development only** |

## Disk Space Summary

### Production Deployment

```
Minimum:      ~430 MB  (runtime + dependencies)
Recommended:  ~1 GB    (with buffer for logs/updates)
High-avail:   ~5 GB    (with extensive logging)
```

### Development Setup

```
Minimum:      ~660 MB  (source + dependencies)
Recommended:  ~1.5 GB  (with dev tools)
Full setup:   ~2 GB    (with all optional tools)
```

### Cleanup Commands

**Free up Go build cache:**
```bash
go clean -cache
go clean -modcache  # Removes module cache (will re-download)
```

**Free up Python cache:**
```bash
find . -type d -name __pycache__ -exec rm -r {} +
find . -type f -name "*.pyc" -delete
```

**Remove build artifacts:**
```bash
rm -rf bin/
make clean
```

## Monitoring Disk Usage

### Check Current Usage

```bash
# Project size
du -sh /home/dlowes/projects/exarp-go

# Binary size
du -sh bin/exarp-go

# Go module cache
du -sh $GOPATH/pkg/mod 2>/dev/null || echo "Not set"

# Go build cache
du -sh $(go env GOCACHE) 2>/dev/null || echo "Not set"

# Python virtual environment
du -sh .venv 2>/dev/null || echo "Not found"
```

### Ansible Disk Usage Check

The Ansible playbooks include disk space checks. Before deployment:

```bash
cd ansible
ansible-playbook --check -i inventories/production playbooks/production.yml
```

## Recommendations

### For Production

1. **Minimum:** 5 GB free disk space
2. **Recommended:** 10-20 GB free disk space
3. **Log rotation:** Configure log rotation to prevent disk fill
4. **Monitoring:** Set up disk usage alerts at 80% capacity

### For Development

1. **Minimum:** 2 GB free disk space
2. **Recommended:** 5-10 GB free disk space
3. **Regular cleanup:** Run `go clean -cache` periodically
4. **Build artifacts:** Remove old binaries when not needed

## Notes

- **Go module cache** is shared across all Go projects
- **Python virtual environments** are project-specific
- **Build artifacts** can be regenerated, safe to delete
- **Logs** should be rotated regularly
- **Documentation** can be excluded in production deployments

---

**Last Updated:** 2026-01-07  
**Based on:** Current project analysis and standard Go/Python installation sizes

