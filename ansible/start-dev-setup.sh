#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "=========================================="
echo "  Ansible Development Setup"
echo "=========================================="
echo ""

# Check for Ansible
if ! command -v ansible-playbook >/dev/null 2>&1; then
    echo "❌ ERROR: ansible-playbook not found"
    echo ""
    echo "Please install Ansible first:"
    echo "  sudo apt install ansible-core"
    echo ""
    echo "Or via pip:"
    echo "  python3 -m pip install --user ansible-core"
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
    exit 1
fi

echo "✅ Found: $(ansible-playbook --version | head -1)"
echo ""

# Syntax check
echo "Step 1: Checking playbook syntax..."
if ansible-playbook --syntax-check -i inventories/development playbooks/development.yml; then
    echo "✅ Syntax OK"
else
    echo "❌ Syntax errors found"
    exit 1
fi
echo ""

# List tasks
echo "Step 2: Tasks to be executed:"
ansible-playbook --list-tasks -i inventories/development playbooks/development.yml 2>/dev/null | grep -E "^  " | head -15
echo ""

# Run playbook
echo "Step 3: Running development setup..."
echo "=========================================="
echo ""
echo "Note: You will be prompted for your sudo password"
echo ""
ansible-playbook -i inventories/development playbooks/development.yml --ask-become-pass

echo ""
echo "=========================================="
echo "✅ Development setup complete!"
echo "=========================================="
echo ""
echo "Verify installation:"
echo "  /usr/local/go/bin/go version"
echo "  python3 --version"
echo "  uv --version 2>/dev/null || echo 'uv not in PATH'"
echo "  gh --version 2>/dev/null || echo 'gh not in PATH (set install_gh: true to install)'"
echo "  golangci-lint --version 2>/dev/null || echo 'golangci-lint not in PATH'"

