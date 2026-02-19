# Ansible Setup Summary

**Date:** 2026-02-19  
**Status:** ✅ Complete

## Overview

Ansible playbooks for setting up development and production environments for the exarp-go Go MCP server. Roles use OS-specific variable files (`vars/Darwin.yml`, `vars/Debian.yml`, `vars/RedHat.yml`) to reduce duplication, and independent installs run in parallel via `async`.

## Structure

```
ansible/
├── ansible.cfg                    # Ansible configuration
├── README.md                      # Full documentation
├── QUICKSTART.md                  # Quick start guide
├── requirements.yml               # Ansible Galaxy dependencies
├── run-dev-setup.sh               # Interactive dev setup script
├── inventories/
│   ├── development/
│   │   ├── hosts                  # Development hosts
│   │   └── group_vars/all.yml     # Development variables
│   └── production/
│       ├── hosts                  # Production hosts
│       └── group_vars/all.yml     # Production variables
├── playbooks/
│   ├── development.yml            # Development setup playbook
│   ├── production.yml             # Production setup playbook
│   └── templates/
│       └── exarp-go.service.j2    # Systemd service template
└── roles/
    ├── common/                    # Base system setup, SQLite, CA certs
    │   ├── tasks/main.yml
    │   └── vars/{Darwin,Debian,RedHat}.yml
    ├── golang/                    # Go, protoc, protoc-gen-go
    │   ├── tasks/main.yml
    │   ├── defaults/main.yml
    │   └── vars/{Darwin,Debian,RedHat}.yml
    ├── linters/                   # Optional linters (includes Node.js for npm tools)
    │   └── tasks/main.yml
    ├── ollama/                    # Optional Ollama + model pulls
    │   └── tasks/main.yml
    └── redis/                     # Optional Redis (queue/worker)
        └── tasks/main.yml
```

## Features

### Development Environment

**Always Installed:**
- Go (latest stable or pinned) with SSL fallback for macOS
- Protocol Buffers (protoc + protoc-gen-go) — use `make proto` after setup
- Base tools (git, curl, wget, jq, make, build tools)
- SQLite runtime
- **Git config:** Less verbose output (advice hints off) when `configure_git_quiet: true` (default)

**Optional (Configurable):**
- golangci-lint, govulncheck — Go linters
- gomarklint, shellcheck, shfmt — Code quality tools
- yamllint, ansible-lint — YAML/Ansible linting (ansible-lint on macOS: brew; Linux: pip3)
- actionlint — GitHub Actions linting (macOS: brew; add to `linters` list if desired)
- markdownlint-cli, cspell — Markdown/spelling (installs Node.js automatically)
- fswatch (macOS) / inotify-tools (Linux) — File watchers
- **Ollama** — Local LLM server + models for native tests
- **Redis** — For queue/worker (Asynq). Set `install_redis: true`; use `REDIS_ADDR=127.0.0.1:6379`

### Production Environment

**Installed:**
- Go (latest stable or pinned)
- Base tools
- Systemd service for exarp-go (Linux only)

**Not Installed:**
- Linters, dev tools, file watchers

## Usage

### Development Setup (Makefile)

```bash
make ansible-check    # Syntax check
make ansible-dev      # Interactive setup (recommended)
make ansible-list     # List tasks
make ansible-galaxy   # Install Galaxy requirements
```

### Development Setup (Manual)

```bash
cd ansible
ansible-playbook -i inventories/development playbooks/development.yml --ask-become-pass
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
  - yamllint
  - ansible-lint
  - gomarklint
  # Add more as needed (e.g. actionlint on macOS)
```

### Change Go Version

```yaml
go_version: "latest"   # or pin e.g. "1.26.0"
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

**Last Updated:** 2026-02-19  
**Status:** ✅ Ready to use

