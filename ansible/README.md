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
│   └── linters/            # Optional linters
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

Setup local development environment with all optional tools:

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

## Tags

Use tags to run specific parts of the playbook:

- `common` - Base system setup
- `golang` - Go installation
- `python` - Python installation
- `linters` - Linter installation
- `dev_tools` - Development tools
- `optional` - All optional tools
- `always` - Always run (default)

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
go_version: "1.25.0"
```

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

