#!/bin/bash
# Run Ansible Development Setup
# This script runs the Ansible development playbook

set -e

cd "$(dirname "$0")"

echo "=== Ansible Development Setup ==="
echo ""

# Check if Ansible is installed
if ! command -v ansible-playbook &> /dev/null; then
    echo "❌ Ansible not found!"
    echo ""
    echo "Please install Ansible first:"
    echo "  sudo apt install ansible-core"
    echo ""
    echo "Or via pip:"
    echo "  python3 -m pip install --user ansible-core"
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
    exit 1
fi

echo "✅ Ansible found: $(ansible-playbook --version | head -1)"
echo ""

# Check syntax
echo "1. Checking playbook syntax..."
if ansible-playbook --syntax-check -i inventories/development playbooks/development.yml; then
    echo "   ✅ Syntax check passed"
else
    echo "   ❌ Syntax check failed"
    exit 1
fi
echo ""

# Show what will be installed
echo "2. Tasks that will run:"
ansible-playbook --list-tasks -i inventories/development playbooks/development.yml | grep -E "^(  |    )" | head -20
echo ""

# Ask for confirmation
read -p "Continue with installation? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled."
    exit 0
fi

# Run the playbook
echo ""
echo "3. Running development setup..."
echo ""
ansible-playbook -i inventories/development playbooks/development.yml

echo ""
echo "=== Setup Complete ==="
echo ""
echo "Verify installation:"
echo "  /usr/local/go/bin/go version"
echo "  python3 --version"
echo "  uv --version"
echo "  gh --version 2>/dev/null || echo 'gh not in PATH (set install_gh: true to install)'"
echo "  golangci-lint --version 2>/dev/null || echo 'golangci-lint not installed (may have been skipped gracefully)'"

