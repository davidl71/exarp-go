# Quick Start - Run Development Setup

## Option 1: Direct Ansible Command

```bash
cd /home/dlowes/projects/exarp-go/ansible
ansible-playbook -i inventories/development playbooks/development.yml
```

## Option 2: Use the Script

If the script is visible:

```bash
cd /home/dlowes/projects/exarp-go/ansible
./start-dev-setup.sh
```

Or:

```bash
cd /home/dlowes/projects/exarp-go/ansible
bash start-dev-setup.sh
```

## Option 3: Check if Script Exists

```bash
cd /home/dlowes/projects/exarp-go/ansible
ls -la *.sh
```

You should see:
- `start-dev-setup.sh`
- `run-dev-setup.sh`

## Prerequisites

Make sure Ansible is installed:

```bash
# Check if installed
ansible-playbook --version

# If not installed:
sudo apt update
sudo apt install -y ansible-core
```

## What Gets Installed

- Go 1.24.0 (~300-400 MB)
- Python 3.10 + uv (~60-120 MB)
- All 6 linters (~50-100 MB)
- File watchers (~1-10 MB)

**Total:** ~410-630 MB

## Verify After Installation

```bash
/usr/local/go/bin/go version
python3 --version
uv --version
golangci-lint --version
```

