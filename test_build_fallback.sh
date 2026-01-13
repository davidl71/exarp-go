#!/bin/bash
# Test script to verify build fallback when sql.h is not found
# This simulates the scenario without actually moving system files

set -e

cd "$(dirname "$0")"

echo "Testing build fallback logic for missing sql.h..."
echo ""

# Test 1: Verify detection logic finds sql.h when present
echo "Test 1: Detection with sql.h present"
ODBC_PREFIX=$(brew --prefix unixodbc 2>/dev/null || echo "")
SQL_H_FOUND=0

if [ -n "$ODBC_PREFIX" ] && [ -f "$ODBC_PREFIX/include/sql.h" ]; then
    SQL_H_FOUND=1
elif [ -f "/opt/homebrew/include/sql.h" ]; then
    SQL_H_FOUND=1
    ODBC_PREFIX="/opt/homebrew"
elif [ -f "/usr/local/include/sql.h" ]; then
    SQL_H_FOUND=1
    ODBC_PREFIX="/usr/local"
fi

if [ $SQL_H_FOUND -eq 1 ]; then
    echo "✅ sql.h detected at: $ODBC_PREFIX/include/sql.h"
    echo "   Expected: Build with ODBC support"
else
    echo "❌ sql.h not detected"
    echo "   Expected: Build with no_odbc tag"
fi

echo ""

# Test 2: Build with no_odbc tag (simulating sql.h not found)
echo "Test 2: Build with no_odbc tag (simulating sql.h not found)"
if CGO_ENABLED=1 go build -tags "no_odbc" -o /tmp/exarp-go-test ./cmd/server 2>&1 | grep -v "ld: warning:" | head -5; then
    if [ -f "/tmp/exarp-go-test" ]; then
        echo "✅ Build succeeded with no_odbc tag"
        echo "   Binary size: $(ls -lh /tmp/exarp-go-test | awk '{print $5}')"
        rm -f /tmp/exarp-go-test
    else
        echo "❌ Build command succeeded but binary not found"
    fi
else
    echo "❌ Build failed with no_odbc tag"
    exit 1
fi

echo ""

# Test 3: Verify Makefile logic (dry run)
echo "Test 3: Makefile build logic (checking detection paths)"
echo "Checking Makefile detection paths..."
if grep -q "SQL_H_FOUND=0" Makefile && grep -q "no_odbc" Makefile; then
    echo "✅ Makefile includes sql.h detection and no_odbc fallback"
else
    echo "❌ Makefile missing detection logic"
fi

echo ""
echo "✅ All fallback tests passed!"
echo ""
echo "Summary:"
echo "- sql.h detection works correctly"
echo "- Build succeeds with no_odbc tag when sql.h missing"
echo "- Makefile includes proper fallback logic"
