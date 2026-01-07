#!/bin/bash
# Go Project Health Check Script
# Performs comprehensive Go-specific health checks

set -e

PROJECT_ROOT="${1:-.}"
cd "$PROJECT_ROOT"

echo "🔍 Go Project Health Check"
echo "=========================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PASS=0
FAIL=0
WARN=0

# Check go.mod
echo -n "Checking go.mod... "
if [ -f "go.mod" ]; then
    echo -e "${GREEN}✅${NC}"
    PASS=$((PASS + 1))
else
    echo -e "${RED}❌${NC}"
    FAIL=$((FAIL + 1))
fi

# Check go.sum
echo -n "Checking go.sum... "
if [ -f "go.sum" ]; then
    echo -e "${GREEN}✅${NC}"
    PASS=$((PASS + 1))
else
    echo -e "${YELLOW}⚠️${NC}  (run 'go mod tidy')"
    WARN=$((WARN + 1))
fi

# Check go mod tidy
echo -n "Checking go mod tidy... "
if go mod tidy -e 2>&1 | grep -q "go: downloading\|go: finding"; then
    echo -e "${YELLOW}⚠️${NC}  (dependencies need updating)"
    WARN=$((WARN + 1))
else
    echo -e "${GREEN}✅${NC}"
    PASS=$((PASS + 1))
fi

# Check go build
echo -n "Checking go build... "
if go build ./... > /dev/null 2>&1; then
    echo -e "${GREEN}✅${NC}"
    PASS=$((PASS + 1))
else
    echo -e "${RED}❌${NC}"
    FAIL=$((FAIL + 1))
fi

# Check go vet
echo -n "Checking go vet... "
if go vet ./... > /dev/null 2>&1; then
    echo -e "${GREEN}✅${NC}"
    PASS=$((PASS + 1))
else
    echo -e "${RED}❌${NC}"
    FAIL=$((FAIL + 1))
fi

# Check go fmt
echo -n "Checking go fmt... "
UNFORMATTED=$(gofmt -l . 2>/dev/null | grep -v "^vendor/" | head -5)
if [ -z "$UNFORMATTED" ]; then
    echo -e "${GREEN}✅${NC}"
    PASS=$((PASS + 1))
else
    echo -e "${YELLOW}⚠️${NC}  (run 'go fmt ./...')"
    WARN=$((WARN + 1))
fi

# Check golangci-lint
echo -n "Checking golangci-lint... "
if command -v golangci-lint > /dev/null 2>&1; then
    if [ -f ".golangci.yml" ] || [ -f ".golangci.yaml" ]; then
        if golangci-lint run --timeout=30s > /dev/null 2>&1; then
            echo -e "${GREEN}✅${NC}"
            PASS=$((PASS + 1))
        else
            echo -e "${RED}❌${NC}"
            FAIL=$((FAIL + 1))
        fi
    else
        echo -e "${YELLOW}⚠️${NC}  (not configured)"
        WARN=$((WARN + 1))
    fi
else
    echo -e "${YELLOW}⚠️${NC}  (not installed)"
    WARN=$((WARN + 1))
fi

# Check go test
echo -n "Checking go test... "
if go test ./... > /dev/null 2>&1; then
    echo -e "${GREEN}✅${NC}"
    PASS=$((PASS + 1))
else
    echo -e "${RED}❌${NC}"
    FAIL=$((FAIL + 1))
fi

# Check test coverage
echo -n "Checking test coverage... "
if go test ./... -coverprofile=coverage.out > /dev/null 2>&1; then
    COVERAGE=$(go tool cover -func=coverage.out 2>/dev/null | grep total | awk '{print $3}')
    rm -f coverage.out
    if [ -n "$COVERAGE" ]; then
        COVERAGE_NUM=${COVERAGE//%/}
        if (( $(echo "$COVERAGE_NUM >= 80" | bc -l) )); then
            echo -e "${GREEN}✅${NC}  ($COVERAGE)"
            PASS=$((PASS + 1))
        elif (( $(echo "$COVERAGE_NUM >= 50" | bc -l) )); then
            echo -e "${YELLOW}⚠️${NC}  ($COVERAGE - target: 80%)"
            WARN=$((WARN + 1))
        else
            echo -e "${RED}❌${NC}  ($COVERAGE - target: 80%)"
            FAIL=$((FAIL + 1))
        fi
    else
        echo -e "${YELLOW}⚠️${NC}  (coverage unknown)"
        WARN=$((WARN + 1))
    fi
else
    echo -e "${RED}❌${NC}"
    FAIL=$((FAIL + 1))
fi

# Check govulncheck
echo -n "Checking govulncheck... "
if command -v govulncheck > /dev/null 2>&1; then
    if govulncheck ./... > /dev/null 2>&1; then
        echo -e "${GREEN}✅${NC}"
        PASS=$((PASS + 1))
    else
        echo -e "${RED}❌${NC}"
        FAIL=$((FAIL + 1))
    fi
else
    echo -e "${YELLOW}⚠️${NC}  (not installed - run 'go install golang.org/x/vuln/cmd/govulncheck@latest')"
    WARN=$((WARN + 1))
fi

# Summary
echo ""
echo "=========================="
echo "Summary:"
echo "  ✅ Pass: $PASS"
echo "  ⚠️  Warn: $WARN"
echo "  ❌ Fail: $FAIL"
echo ""

if [ $FAIL -eq 0 ]; then
    echo -e "${GREEN}✅ All critical checks passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some checks failed${NC}"
    exit 1
fi

