#!/bin/bash
# Script to check APT repositories for issues

set -euo pipefail

echo "=== APT Repository Health Check ==="
echo ""

# Run apt update and capture output
echo "Running: sudo apt update"
echo "---"
sudo apt update 2>&1 | tee /tmp/apt-update-output.log
EXIT_CODE=${PIPESTATUS[0]}

echo ""
echo "---"
echo ""

if [ $EXIT_CODE -eq 0 ]; then
    echo "✓ apt update completed successfully"
    echo ""
    
    # Check for warnings
    if grep -i "warning" /tmp/apt-update-output.log > /dev/null; then
        echo "⚠ Warnings found:"
        grep -i "warning" /tmp/apt-update-output.log | head -10
        echo ""
    fi
else
    echo "✗ apt update failed with exit code: $EXIT_CODE"
    echo ""
    echo "Errors found:"
    grep -iE "error|fail|404|403|timeout|could not|unable to" /tmp/apt-update-output.log | head -20
    echo ""
fi

# Summary of repositories
echo "=== Repository Summary ==="
echo ""
echo "Active repositories:"
grep -h "^deb\|^URIs:" /etc/apt/sources.list /etc/apt/sources.list.d/*.list /etc/apt/sources.list.d/*.sources 2>/dev/null | \
    grep -v "^#" | grep -v "^$" | \
    sed 's/^deb \[[^]]*\] //; s/^deb //; s/^URIs: //' | \
    awk '{print $1}' | sort -u | nl

echo ""
echo "Full output saved to: /tmp/apt-update-output.log"

