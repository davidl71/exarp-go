#!/bin/bash
# Run Ansible Development Setup
# Installs galaxy requirements, validates playbook, then runs it.

set -e

cd "$(dirname "$0")"

echo "=== Ansible Development Setup ==="
echo ""

# --- Check Ansible ---
if ! command -v ansible-playbook &> /dev/null; then
    echo "❌ Ansible not found!"
    echo ""
    if [[ "$(uname)" == "Darwin" ]]; then
        echo "Install via Homebrew:"
        echo "  brew install ansible"
    else
        echo "Install via package manager:"
        echo "  sudo apt install ansible-core   # Debian/Ubuntu"
        echo "  sudo dnf install ansible-core   # Fedora/RHEL"
    fi
    echo ""
    echo "Or via pip/uv:"
    echo "  uv tool install ansible-core"
    echo "  python3 -m pip install --user ansible-core"
    exit 1
fi

echo "✅ Ansible found: $(ansible-playbook --version | head -1)"
echo ""

# --- Install Galaxy requirements ---
if [ -f requirements.yml ]; then
    echo "1. Installing Ansible Galaxy requirements..."
    if ansible-galaxy collection install -r requirements.yml --force-with-deps 2>/dev/null; then
        echo "   ✅ Galaxy requirements installed"
    else
        echo "   ⚠️  Galaxy install had issues (continuing anyway)"
    fi
    echo ""
else
    echo "1. No requirements.yml found, skipping galaxy install"
    echo ""
fi

# --- Syntax check ---
echo "2. Checking playbook syntax..."
if ansible-playbook --syntax-check -i inventories/development playbooks/development.yml; then
    echo "   ✅ Syntax check passed"
else
    echo "   ❌ Syntax check failed"
    exit 1
fi
echo ""

# --- Show tasks ---
echo "3. Tasks that will run:"
ansible-playbook --list-tasks -i inventories/development playbooks/development.yml | grep -E "^(  |    )" | head -25
echo ""

# --- Confirm ---
read -p "Continue with installation? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled."
    exit 0
fi

# --- Run ---
echo ""
echo "4. Running development setup..."
echo ""

EXTRA_ARGS=""
if [[ "$(uname)" == "Darwin" ]]; then
    # macOS: localhost, may not need become password for Homebrew
    EXTRA_ARGS="--ask-become-pass"
else
    EXTRA_ARGS="--ask-become-pass"
fi

ansible-playbook -i inventories/development playbooks/development.yml $EXTRA_ARGS

echo ""
echo "=== Setup Complete ==="
echo ""
echo "Verify installation:"
echo "  go version"
echo "  gh --version 2>/dev/null || echo 'gh: not installed (set install_gh: true)'"
echo "  golangci-lint --version 2>/dev/null || echo 'golangci-lint: not installed (set install_linters: true)'"
echo "  redis-cli ping 2>/dev/null && echo 'Redis: OK (REDIS_ADDR=127.0.0.1:6379)' || echo 'Redis: not installed (set install_redis: true)'"
echo "  ollama --version 2>/dev/null || echo 'Ollama: not installed (set install_ollama: true)'"

