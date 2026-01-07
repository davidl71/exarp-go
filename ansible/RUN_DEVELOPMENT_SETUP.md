# Running Development Setup

## Quick Answer: Does the estimation include linters?

**YES** - The disk usage estimation (~660-1,280 MB for full development setup) **includes linters**:

- **Development tools (linters):** ~50-100 MB
  - golangci-lint: ~30-50 MB
  - shellcheck: ~5-10 MB
  - shfmt: ~2-5 MB
  - markdownlint-cli: ~10-20 MB
  - cspell: ~5-10 MB

The development playbook installs linters by default because `install_linters: true` is set in `inventories/development/group_vars/all.yml`.

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
ansible-playbook --version
```

## Running the Development Setup

### Step 1: Check Syntax

```bash
cd /home/dlowes/projects/exarp-go/ansible
ansible-playbook --syntax-check -i inventories/development playbooks/development.yml
```

### Step 2: Dry Run (Check What Would Change)

```bash
ansible-playbook --check -i inventories/development playbooks/development.yml
```

### Step 3: Run the Playbook

```bash
ansible-playbook -i inventories/development playbooks/development.yml
```

**Note:** This will:
- Install Go 1.24.0
- Install Python 3.10 and uv
- Install all linters (golangci-lint, shellcheck, shfmt, gomarklint, markdownlint-cli, cspell)
- Install file watchers (inotify-tools on Linux, fswatch on macOS)
- Create project directories
- Set up environment variables

### Step 4: Verify Installation

```bash
# Check Go
/usr/local/go/bin/go version

# Check Python
python3 --version

# Check uv
uv --version

# Check linters
golangci-lint --version
shellcheck --version
shfmt --version
markdownlint --version
cspell --version
```

## Disk Usage After Installation

After running the development setup, you should see approximately:

- **Go 1.24.0:** ~300-400 MB (`/usr/local/go`)
- **Python 3.10:** ~50-100 MB (system package)
- **uv:** ~10-20 MB
- **Linters:** ~50-100 MB
  - golangci-lint: ~30-50 MB
  - shellcheck: ~5-10 MB
  - shfmt: ~2-5 MB
  - markdownlint-cli: ~10-20 MB
  - cspell: ~5-10 MB
- **Go module cache:** ~50-200 MB (after first build)
- **Total:** **~460-820 MB** (excluding OS)

## Troubleshooting

### Ansible Not Found

If `ansible-playbook` command is not found:
```bash
# Check if installed
which ansible-playbook

# If installed via pip --user, add to PATH
export PATH="$HOME/.local/bin:$PATH"

# Or install system-wide
sudo apt install ansible-core
```

### Permission Errors

The playbook uses `become: yes` (sudo). Ensure you have sudo access:
```bash
sudo -v
```

### Skip Linters Installation

If you want to skip linters (to save ~50-100 MB), modify the inventory:
```bash
# Edit inventories/development/group_vars/all.yml
install_linters: false
```

Or run with tags to skip linters:
```bash
ansible-playbook -i inventories/development playbooks/development.yml --skip-tags linters
```

### Check What Will Be Installed

```bash
# List all tasks
ansible-playbook --list-tasks -i inventories/development playbooks/development.yml

# List only linter tasks
ansible-playbook --list-tasks -i inventories/development playbooks/development.yml | grep -i lint
```

## Summary

✅ **Linters ARE included** in the disk usage estimation  
✅ **Linters ARE installed** by default in development setup  
✅ **Total development setup:** ~660-1,280 MB (including linters)

See `docs/SYSTEM_REQUIREMENTS.md` for complete disk usage breakdown.

