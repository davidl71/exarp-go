# Ansible Quick Start Guide

## Quick Setup (Local Development)

### Option 1: Direct Ansible command

```bash
cd ansible
ansible-playbook -i inventories/development playbooks/development.yml
```

With tags:
```bash
# Full setup (includes linters)
ansible-playbook playbooks/development.yml

# Without optional tools
ansible-playbook playbooks/development.yml --skip-tags optional

# Only Go and Python
ansible-playbook playbooks/development.yml --tags golang,python

# Only linters
ansible-playbook playbooks/development.yml --tags linters

# Ollama (fixes ollama tool / native tests; then run ollama serve)
ansible-playbook playbooks/development.yml --tags ollama
```

### Option 2: Use the script

```bash
cd ansible
./run-dev-setup.sh
```

The script checks syntax, lists tasks, prompts for confirmation, then runs the playbook. Use `--ask-become-pass` if you need to pass a sudo password (or run with `ansible-playbook ... --ask-become-pass` manually).

### Prerequisites

**Install Ansible**

macOS:
```bash
brew install ansible
```

Linux:
```bash
# Debian/Ubuntu
sudo apt install ansible

# RedHat/CentOS
sudo yum install ansible
```

**Install Galaxy dependencies (optional)**

```bash
cd ansible
ansible-galaxy install -r requirements.yml
```

## What gets installed

### Development (full)
- Go 1.24.0 (~300–400 MB)
- Python 3.10+ with pip and uv (~60–120 MB)
- Node.js & npm
- Linters: golangci-lint, shellcheck, gomarklint, shfmt, markdownlint-cli, cspell (~50–100 MB)
- File watchers: fswatch (macOS) / inotify-tools (Linux) (~1–10 MB)

**Total:** ~410–630 MB

### Production (minimal)
- Go 1.24.0, Python 3.10+, pip, uv, Node.js & npm
- No linters or dev tools

## Verify after installation

```bash
/usr/local/go/bin/go version
python3 --version
uv --version
golangci-lint --version   # if linters were installed
gh --version             # if install_gh: true
```

## Common commands

```bash
# Dry run (check mode)
ansible-playbook playbooks/development.yml --check

# Verbose output
ansible-playbook playbooks/development.yml -v

# Test connectivity
ansible all -i inventories/development/hosts -m ping

# Syntax check
ansible-playbook playbooks/development.yml --syntax-check
```

## Customization

Edit `inventories/development/group_vars/all.yml` to:
- Change Go version
- Enable/disable linters (`install_linters`, `linters` list)
- Enable GitHub CLI (`install_gh`)
- Enable Ollama for ollama tool / native tests (`install_ollama: true`; then run `ollama serve`)

## Troubleshooting

```bash
# Verbose debugging
ansible-playbook playbooks/development.yml -vvv

# If sudo is required
ansible-playbook -i inventories/development playbooks/development.yml --ask-become-pass
```

---

For full documentation, see [ansible/README.md](README.md).
