# Ansible Playbooks for exarp-go

Ansible automation for setting up development and production environments for the exarp-go Go MCP server.

## Structure

```
ansible/
├── ansible.cfg              # Ansible configuration
├── inventories/
│   ├── development/        # Development environment
│   │   ├── hosts
│   │   └── group_vars/
│   └── production/         # Production environment
│       ├── hosts
│       └── group_vars/
├── roles/
│   ├── common/             # Base system setup
│   ├── golang/             # Go installation
│   ├── python/             # Python and package managers
│   ├── linters/            # Optional linters
│   ├── ollama/             # Optional Ollama (fixes ollama tool / native tests)
│   └── redis/              # Optional Redis (queue/worker: make queue-enqueue-wave, queue-worker)
├── playbooks/
│   ├── development.yml     # Development setup
│   └── production.yml      # Production setup
└── requirements.yml        # Ansible Galaxy dependencies
```

## Prerequisites

- Ansible 2.9+ installed
- SSH access to target hosts (for remote deployment)
- Sudo/root access on target hosts

## Installation

### Install Ansible

**macOS:**
```bash
brew install ansible
```

**Linux (Debian/Ubuntu):**
```bash
sudo apt update
sudo apt install ansible
```

**Linux (RedHat/CentOS):**
```bash
sudo yum install ansible
```

### Install Ansible Galaxy Dependencies

```bash
cd ansible
ansible-galaxy install -r requirements.yml
```

## Usage

For a short path, see **[QUICKSTART.md](QUICKSTART.md)** or run `./run-dev-setup.sh` from the ansible directory.

### Development Environment

Setup local development environment with all optional tools. The **common** role also configures git for less verbose output (advice hints off) for the dev user when `configure_git_quiet: true` (default); see [docs/DEV_ENV_GIT_CONFIG.md](../docs/DEV_ENV_GIT_CONFIG.md).

```bash
cd ansible
ansible-playbook playbooks/development.yml
```

With specific tags:
```bash
# Only install Go
ansible-playbook playbooks/development.yml --tags golang

# Only install linters
ansible-playbook playbooks/development.yml --tags linters

# Only install Redis (queue/worker)
ansible-playbook playbooks/development.yml --tags redis

# Skip optional tools
ansible-playbook playbooks/development.yml --skip-tags optional
```

### Production Environment

Setup production environment (minimal, no linters):

```bash
cd ansible
ansible-playbook -i inventories/production/hosts playbooks/production.yml
```

### Dry Run (Check Mode)

Test playbooks without making changes:

```bash
ansible-playbook playbooks/development.yml --check
```

## Configuration

### Development Variables

Edit `inventories/development/group_vars/all.yml`:

```yaml
# Enable/disable optional tools
install_linters: true
install_dev_tools: true
install_file_watchers: true
install_ollama: true   # Ollama (ollama serve + models)
install_redis: false  # Redis for queue/worker; set REDIS_ADDR=127.0.0.1:6379 (see docs/EXARP_CLI_SHORTCUTS.md)

# Select specific linters
linters:
  - golangci-lint
  - shellcheck
  - gomarklint
```

### Production Variables

Edit `inventories/production/group_vars/all.yml`:

```yaml
# Production typically doesn't need linters
install_linters: false
install_dev_tools: false
```

## What Gets Installed

### Always Installed (Both Environments)

- **Go 1.24.0** - Go programming language
- **Protocol Buffers (protoc)** - Installed by both **common** (base packages) and **golang** (protoc + protoc-gen-go). Required for `make proto` to regenerate Go from `.proto` files.
- **Python 3.10+** - Python runtime
- **pip** - Python package manager
- **uv** - Fast Python package manager
- **Node.js & npm** - For MCP server dependencies
- **Base tools** - git, curl, wget, make, build tools

### Optional (Development Only)

- **golangci-lint** - Comprehensive Go linter
- **gomarklint** - Go markdown linter
- **shellcheck** - Shell script linter
- **shfmt** - Shell script formatter
- **markdownlint-cli** - Markdown linter
- **cspell** - Code spell checker
- **fswatch** (macOS) / **inotify-tools** (Linux) - File watchers
- **Ollama** - For the ollama tool and native tests (set `install_ollama: true`; then run `ollama serve`)

## Tags

Use tags to run specific parts of the playbook:

- `common` - Base system setup
- `golang` - Go installation
- `python` - Python installation
- `linters` - Linter installation
- `ollama` - Ollama + models (fixes ollama tool / native tests)
- `dev_tools` - Development tools
- `optional` - All optional tools
- `always` - Always run (default)

### Fix environment for ollama native tests

To fix the environment so ollama native tests pass (server + models):

1. Set `install_ollama: true` in `inventories/development/group_vars/all.yml`.
2. Run the **ollama-only** playbook (recommended; does not run other roles):
   ```bash
   ansible-playbook -i inventories/development playbooks/ollama.yml
   ```
   Or run the full development playbook with the ollama tag:
   ```bash
   ansible-playbook -i inventories/development playbooks/development.yml --tags ollama
   ```
3. Start the Ollama server: `ollama serve` (or run the Ollama app).
4. Run tests: `make test-go` or `make test-ollama`. Ollama tests use a **light default model** (`qwen2.5:1.5b`) for speed; pull it with `ollama pull qwen2.5:1.5b`. For larger families use **quantized tags** (e.g. `qwen2.5:7b-q4_0`) to keep runs fast and memory low. Override with `OLLAMA_TEST_MODEL` / `OLLAMA_TEST_CODE_MODEL` (and `OLLAMA_DEFAULT_MODEL` / `OLLAMA_CODE_MODEL` for real-models tests) to match your installed models.

## Examples

### Install Everything (Development)

```bash
ansible-playbook playbooks/development.yml
```

### Install Only Go and Python

```bash
ansible-playbook playbooks/development.yml --tags golang,python
```

### Install Only Linters

```bash
ansible-playbook playbooks/development.yml --tags linters
```

### Production Setup (No Linters)

```bash
ansible-playbook -i inventories/production/hosts playbooks/production.yml
```

## Troubleshooting

### Permission Issues

If you get permission errors, ensure sudo access:

```bash
ansible-playbook playbooks/development.yml --ask-become-pass
```

### Connection Issues

For local setup, ensure `ansible_connection=local` in inventory.

For remote hosts, ensure SSH access:

```bash
ansible all -i inventories/development/hosts -m ping
```

### Go Installation Issues

If Go installation fails, check:
- Internet connectivity
- Disk space
- Write permissions to `/usr/local`

## Customization

### Add Custom Linters

Edit `inventories/development/group_vars/all.yml`:

```yaml
linters:
  - golangci-lint
  - shellcheck
  - your-custom-linter
```

Then add installation task in `roles/linters/tasks/main.yml`.

### Change Go Version

Edit `inventories/development/group_vars/all.yml`:

```yaml
# Latest stable (default), or pin e.g. "1.25.6"
go_version: "latest"
```

## Ansible 2.24 migration

`inject_facts_as_vars` defaults to `False` in Ansible 2.24. This project sets it to `True` explicitly and silences deprecation warnings in `ansible.cfg`. Before upgrading to 2.24, migrate tasks to use `ansible_facts['ansible_*']` instead of top-level `ansible_*` variables.

## Best Practices

1. **Always test in development first** - Use `--check` mode
2. **Use tags** - Install only what you need
3. **Version control** - Commit inventory changes
4. **Separate environments** - Never mix dev/prod configs
5. **Review changes** - Use `--diff` to see file changes

## Support

For issues or questions:
- Check Ansible logs: `ansible-playbook ... -v` (verbose)
- Review playbook syntax: `ansible-playbook ... --syntax-check`
- Test connectivity: `ansible all -i inventories/development/hosts -m ping`

---

**Last Updated:** 2026-01-07

