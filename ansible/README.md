# Ansible Configuration for exarp-go

This directory contains Ansible playbooks and roles for automating the setup of the exarp-go MCP server in development and production environments.

## Structure

```
ansible/
├── ansible.cfg              # Ansible configuration
├── inventories/             # Environment inventories
│   ├── development/
│   │   ├── hosts           # Development hosts
│   │   └── group_vars/
│   │       └── all.yml     # Development variables
│   └── production/
│       ├── hosts           # Production hosts
│       └── group_vars/
│           └── all.yml     # Production variables
├── playbooks/
│   ├── development.yml     # Development environment setup
│   ├── production.yml      # Production environment setup
│   └── templates/
│       └── exarp-go.service.j2  # Systemd service template
└── roles/
    ├── common/             # Base system setup
    ├── golang/             # Go installation
    ├── python/             # Python and uv installation
    └── linters/            # Development linters
```

## Prerequisites

### Install Ansible

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install -y ansible-core
```

**macOS:**
```bash
brew install ansible
```

**Via pip (user install):**
```bash
python3 -m pip install --user ansible-core
export PATH="$HOME/.local/bin:$PATH"
```

### Verify Installation

```bash
ansible --version
ansible-playbook --version
```

## Usage

### Development Environment

```bash
cd ansible
ansible-playbook -i inventories/development playbooks/development.yml
```

This will:
- Install Go 1.24.0
- Install Python 3.10 and uv
- Install all development linters (golangci-lint, shellcheck, shfmt, gomarklint, markdownlint-cli, cspell)
- Install file watchers (fswatch/inotify-tools)
- Optionally install GitHub CLI (gh) if `install_gh: true` is set
- Set up project directories

### Production Environment

```bash
cd ansible
ansible-playbook -i inventories/production playbooks/production.yml
```

This will:
- Install Go 1.24.0
- Install Python 3.10 and uv
- Create systemd service for exarp-go
- Set up project directories (no linters or dev tools)

### Validate Configuration

```bash
cd ansible

# Check syntax
ansible-playbook --syntax-check -i inventories/development playbooks/development.yml
ansible-playbook --syntax-check -i inventories/production playbooks/production.yml

# Dry run (check what would change)
ansible-playbook --check -i inventories/development playbooks/development.yml
```

## Roles

### common
Base system setup:
- Updates package cache
- Installs base packages (curl, wget, git, build tools)
- Creates project user and directories

### golang
Go installation:
- Downloads and installs Go 1.24.0
- Sets up Go workspace (GOPATH, GOROOT)
- Configures environment variables

### python
Python and uv installation:
- Installs Python 3.10
- Upgrades pip to latest
- Installs uv (Python package manager)

### linters
Development linters (development only):
- golangci-lint
- shellcheck
- shfmt
- gomarklint
- markdownlint-cli
- cspell

## Configuration Variables

### Development (`inventories/development/group_vars/all.yml`)

- `go_version`: "1.24.0"
- `python_version`: "3.10"
- `project_path`: "/opt/mcp-stdio-tools"
- `install_linters`: true
- `install_dev_tools`: true
- `install_file_watchers`: true

### Production (`inventories/production/group_vars/all.yml`)

- `go_version`: "1.24.0"
- `python_version`: "3.10"
- `project_path`: "/opt/mcp-stdio-tools"
- `project_user`: "deploy"
- `install_linters`: false
- `install_dev_tools`: false

## Systemd Service

The production playbook creates a systemd service file at `/etc/systemd/system/exarp-go.service`:

```bash
# After deployment, enable and start the service:
sudo systemctl daemon-reload
sudo systemctl enable exarp-go
sudo systemctl start exarp-go
sudo systemctl status exarp-go
```

## Troubleshooting

### Ansible Not Found

If `ansible` command is not found after installation:
```bash
# Check if it's in PATH
which ansible

# If installed via pip --user, add to PATH:
export PATH="$HOME/.local/bin:$PATH"
```

### Permission Errors

If you get permission errors:
- Ensure you have sudo access
- Check that `become: yes` is set in playbooks (it is by default)
- Verify SSH keys are set up for remote hosts

### Go Installation Fails

If Go installation fails:
- Check internet connectivity
- Verify Go version exists at https://go.dev/dl/
- Check disk space: `df -h`

### Python/uv Installation Issues

If Python or uv installation fails:
- Ensure package manager is updated
- Check Python version compatibility
- For uv, verify curl is available

## Status

✅ **All roles created:**
- common
- golang
- python
- linters

✅ **All playbooks validated:**
- development.yml
- production.yml

✅ **Systemd template created:**
- exarp-go.service.j2

✅ **YAML syntax validated:**
- All playbooks
- All roles
- All inventory files

## Next Steps

1. **Install Ansible** (if not already installed)
2. **Update inventory hosts** with your actual server IPs/hostnames
3. **Review and adjust variables** in `group_vars/all.yml` files
4. **Run syntax check**: `ansible-playbook --syntax-check ...`
5. **Test with dry run**: `ansible-playbook --check ...`
6. **Deploy**: `ansible-playbook ...`

