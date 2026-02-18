# Ansible Setup Summary

**Date:** 2026-01-07  
**Status:** ✅ Complete

## Overview

Created comprehensive Ansible playbooks for setting up both development and production environments for the exarp-go Go MCP server, with optional linters and dependencies.

## Structure Created

```
ansible/
├── ansible.cfg                    # Ansible configuration
├── README.md                      # Full documentation
├── QUICKSTART.md                  # Quick start guide
├── requirements.yml               # Ansible Galaxy dependencies
├── inventories/
│   ├── development/
│   │   ├── hosts                  # Development hosts
│   │   └── group_vars/
│   │       └── all.yml            # Development variables
│   └── production/
│       ├── hosts                  # Production hosts
│       └── group_vars/
│           └── all.yml            # Production variables
├── playbooks/
│   ├── development.yml            # Development setup playbook
│   ├── production.yml             # Production setup playbook
│   └── templates/
│       └── exarp-go.service.j2    # Systemd service template (used by production.yml)
└── roles/
    ├── common/                    # Base system setup
    │   └── tasks/main.yml
    ├── golang/                    # Go installation
    │   └── tasks/main.yml
    ├── python/                    # Python & package managers
    │   └── tasks/main.yml
    ├── linters/                   # Optional linters
    │   └── tasks/main.yml
    ├── ollama/                    # Optional Ollama
    │   └── tasks/main.yml
    └── redis/                     # Optional Redis (queue/worker)
        └── tasks/main.yml
```

## Features

### Development Environment

**Always Installed:**
- Go 1.24.0
- Protocol Buffers (protoc + protoc-gen-go) — in **common** and **golang** roles; use `make proto` after setup
- Python 3.10+
- pip, uv (Python package manager)
- Node.js & npm
- Base tools (git, curl, wget, make, build tools)
- **Git config:** Less verbose output (advice hints off) for the dev user when `configure_git_quiet: true` (default). See [docs/DEV_ENV_GIT_CONFIG.md](DEV_ENV_GIT_CONFIG.md).

**Optional (Configurable):**
- golangci-lint - Comprehensive Go linter
- gomarklint - Go markdown linter
- shellcheck - Shell script linter
- shfmt - Shell script formatter
- markdownlint-cli - Markdown linter
- cspell - Code spell checker
- fswatch (macOS) / inotify-tools (Linux) - File watchers
- **Redis** - For exarp-go queue/worker (Asynq). Set `install_redis: true` in group_vars; then use `REDIS_ADDR=127.0.0.1:6379` with `make queue-enqueue-wave` and `make queue-worker`. See `docs/EXARP_CLI_SHORTCUTS.md`.

### Production Environment

**Installed:**
- Go 1.24.0
- Python 3.10+
- pip, uv
- Node.js & npm
- Base tools
- Systemd service for exarp-go

**Not Installed:**
- Linters (not needed in production)
- Development tools
- File watchers

## Usage

### Development Setup

```bash
cd ansible
ansible-playbook playbooks/development.yml
```

### Production Setup

```bash
cd ansible
ansible-playbook -i inventories/production/hosts playbooks/production.yml
```

### With Tags

```bash
# Only install Go
ansible-playbook playbooks/development.yml --tags golang

# Only install linters
ansible-playbook playbooks/development.yml --tags linters

# Skip optional tools
ansible-playbook playbooks/development.yml --skip-tags optional
```

## Configuration

### Enable/Disable Optional Tools

Edit `inventories/development/group_vars/all.yml`:

```yaml
install_linters: true      # Install linters
install_dev_tools: true    # Install dev tools
install_file_watchers: true # Install file watchers
install_ollama: true       # Ollama (ollama serve + models)
install_redis: false       # Redis for queue/worker (REDIS_ADDR=127.0.0.1:6379)
```

### Select Specific Linters

```yaml
linters:
  - golangci-lint
  - shellcheck
  - gomarklint
  # Add more as needed
```

### Change Versions

```yaml
go_version: "latest"   # or pin e.g. "1.25.6"
python_version: "3.10"
```

## Verification

✅ **Ansible Installed:** Version 2.20.1  
✅ **Syntax Check:** All playbooks pass  
✅ **Structure:** Complete and organized  
✅ **Documentation:** README and QUICKSTART created

## Next Steps

1. **Test Development Setup:**
   ```bash
   cd ansible
   ansible-playbook playbooks/development.yml --check
   ```

2. **Run Development Setup:**
   ```bash
   ansible-playbook playbooks/development.yml
   ```

3. **Customize as Needed:**
   - Edit `inventories/development/group_vars/all.yml`
   - Add custom linters to `roles/linters/tasks/main.yml`

4. **Configure Production:**
   - Update `inventories/production/hosts` with actual hosts
   - Customize `inventories/production/group_vars/all.yml`

## Documentation

- **Full Guide:** `ansible/README.md`
- **Quick Start:** `ansible/QUICKSTART.md`
- **This Summary:** `docs/ANSIBLE_SETUP.md`

---

**Last Updated:** 2026-01-07  
**Status:** ✅ Ready to use

