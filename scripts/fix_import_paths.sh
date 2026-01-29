#!/bin/bash
# Fix Python Import Paths for exarp-go
# Copies missing scripts directory and files from project-management-automation
#
# Obsolete for exarp-go MLX-only (2026-01-29): project_management_automation/scripts
# was removed. Python is deprecated; only MLX bridge path remains. This script is
# for reference or external sync only; exarp-go no longer uses these scripts.

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Paths
EXARP_GO_ROOT="/Users/davidl/Projects/exarp-go"
SOURCE_ROOT="/Users/davidl/Projects/project-management-automation"
EXARP_SCRIPTS_DIR="${EXARP_GO_ROOT}/project_management_automation/scripts"
SOURCE_SCRIPTS_DIR="${SOURCE_ROOT}/project_management_automation/scripts"

echo -e "${GREEN}🔧 Fixing Python Import Paths for exarp-go${NC}"
echo ""

# Check if source directory exists
if [ ! -d "$SOURCE_SCRIPTS_DIR" ]; then
    echo -e "${RED}❌ Error: Source scripts directory not found: $SOURCE_SCRIPTS_DIR${NC}"
    exit 1
fi

# Create scripts directory structure
echo -e "${YELLOW}📁 Creating scripts directory structure...${NC}"
mkdir -p "${EXARP_SCRIPTS_DIR}/base"

# Copy __init__.py files
echo -e "${YELLOW}📄 Copying __init__.py files...${NC}"
if [ -f "${SOURCE_SCRIPTS_DIR}/__init__.py" ]; then
    cp "${SOURCE_SCRIPTS_DIR}/__init__.py" "${EXARP_SCRIPTS_DIR}/__init__.py"
    echo "  ✅ Copied scripts/__init__.py"
else
    touch "${EXARP_SCRIPTS_DIR}/__init__.py"
    echo "  ✅ Created scripts/__init__.py"
fi

if [ -f "${SOURCE_SCRIPTS_DIR}/base/__init__.py" ]; then
    cp "${SOURCE_SCRIPTS_DIR}/base/__init__.py" "${EXARP_SCRIPTS_DIR}/base/__init__.py"
    echo "  ✅ Copied scripts/base/__init__.py"
else
    touch "${EXARP_SCRIPTS_DIR}/base/__init__.py"
    echo "  ✅ Created scripts/base/__init__.py"
fi

# Copy base classes
echo -e "${YELLOW}📦 Copying base classes...${NC}"
for file in "intelligent_automation_base.py" "mcp_client.py"; do
    if [ -f "${SOURCE_SCRIPTS_DIR}/base/${file}" ]; then
        cp "${SOURCE_SCRIPTS_DIR}/base/${file}" "${EXARP_SCRIPTS_DIR}/base/${file}"
        echo "  ✅ Copied base/${file}"
    else
        echo -e "  ${RED}⚠️  Warning: ${file} not found in source${NC}"
    fi
done

# List of scripts needed (from grep results). automate_daily.py removed 2026-01-29; use exarp-go automation.
NEEDED_SCRIPTS=(
    "automate_docs_health_v2.py"
    "automate_todo2_alignment_v2.py"
    "automate_todo2_duplicate_detection.py"
    "automate_dependency_security.py"
    "automate_automation_opportunities.py"
    "automate_todo_sync.py"
    "automate_external_tool_hints.py"
    "automate_stale_task_cleanup.py"
    "automate_sprint.py"
    "automate_run_tests.py"
    "automate_test_coverage.py"
    "automate_attribution_check.py"
)

# Copy needed scripts
echo -e "${YELLOW}📜 Copying automation scripts...${NC}"
COPIED=0
MISSING=0

for script in "${NEEDED_SCRIPTS[@]}"; do
    if [ -f "${SOURCE_SCRIPTS_DIR}/${script}" ]; then
        cp "${SOURCE_SCRIPTS_DIR}/${script}" "${EXARP_SCRIPTS_DIR}/${script}"
        echo "  ✅ Copied ${script}"
        ((COPIED++))
    else
        echo -e "  ${RED}⚠️  Warning: ${script} not found in source${NC}"
        ((MISSING++))
    fi
done

# Copy any additional scripts that might be needed
echo -e "${YELLOW}📜 Copying additional scripts...${NC}"
for script in "${SOURCE_SCRIPTS_DIR}"/automate_*.py; do
    if [ -f "$script" ]; then
        script_name=$(basename "$script")
        if [ ! -f "${EXARP_SCRIPTS_DIR}/${script_name}" ]; then
            cp "$script" "${EXARP_SCRIPTS_DIR}/${script_name}"
            echo "  ✅ Copied additional script: ${script_name}"
            ((COPIED++))
        fi
    fi
done

# Summary
echo ""
echo -e "${GREEN}✅ Import path fix complete!${NC}"
echo "  📊 Scripts copied: ${COPIED}"
if [ $MISSING -gt 0 ]; then
    echo -e "  ${YELLOW}⚠️  Missing scripts: ${MISSING}${NC}"
fi
echo ""
echo "📁 Directory structure:"
echo "  ${EXARP_SCRIPTS_DIR}/"
echo "  ${EXARP_SCRIPTS_DIR}/base/"
echo ""
echo -e "${GREEN}✨ Ready to test daily automation!${NC}"

