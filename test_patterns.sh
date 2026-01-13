#!/bin/bash
# Quick test of setup_hooks patterns action

# Test 1: Dry run with defaults
echo "=== Test 1: Dry run with default patterns ==="
echo '{"action":"patterns","dry_run":true}' | ./bin/exarp-go 2>&1 | head -30

# Test 2: Check if config file would be created
echo ""
echo "=== Test 2: Check config file location ==="
ls -la .cursor/exarp_patterns.json 2>/dev/null || echo "Config file doesn't exist yet (expected for dry run)"

# Test 3: Actual installation (if we want to test)
# echo ""
# echo "=== Test 3: Actual installation ==="
# echo '{"action":"patterns","dry_run":false}' | ./bin/exarp-go 2>&1 | head -30
