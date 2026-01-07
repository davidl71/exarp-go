# Ansible Quick Start Guide

## Quick Setup (Local Development)

### 1. Install Ansible

**macOS:**
```bash
brew install ansible
```

**Linux:**
```bash
# Debian/Ubuntu
sudo apt install ansible

# RedHat/CentOS
sudo yum install ansible
```

### 2. Install Dependencies

```bash
cd ansible
ansible-galaxy install -r requirements.yml
```

### 3. Run Development Playbook

```bash
# Full setup (includes linters)
ansible-playbook playbooks/development.yml

# Without optional tools
ansible-playbook playbooks/development.yml --skip-tags optional

# Only Go and Python
ansible-playbook playbooks/development.yml --tags golang,python
```

## What Gets Installed

### Development (Full)
- ✅ Go 1.24.0
- ✅ Python 3.10+
- ✅ pip, uv
- ✅ Node.js & npm
- ✅ golangci-lint
- ✅ shellcheck
- ✅ gomarklint
- ✅ shfmt
- ✅ markdownlint-cli
- ✅ cspell
- ✅ File watchers (fswatch/inotify-tools)

### Production (Minimal)
- ✅ Go 1.24.0
- ✅ Python 3.10+
- ✅ pip, uv
- ✅ Node.js & npm
- ❌ No linters
- ❌ No dev tools

## Common Commands

```bash
# Check what would change (dry run)
ansible-playbook playbooks/development.yml --check

# Verbose output
ansible-playbook playbooks/development.yml -v

# Install specific components
ansible-playbook playbooks/development.yml --tags linters
ansible-playbook playbooks/development.yml --tags golang
```

## Customization

Edit `inventories/development/group_vars/all.yml` to:
- Change Go version
- Enable/disable linters
- Add custom tools

## Troubleshooting

```bash
# Test connectivity
ansible all -i inventories/development/hosts -m ping

# Check syntax
ansible-playbook playbooks/development.yml --syntax-check

# Verbose debugging
ansible-playbook playbooks/development.yml -vvv
```

---

For full documentation, see [ansible/README.md](README.md)

