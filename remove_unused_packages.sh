#!/bin/bash
#
# Remove Unused Packages Script
# Safely removes packages identified as unused based on usage analysis
#
# Usage:
#   ./remove_unused_packages.sh [--dry-run] [--skip-qemu] [--help]
#
# Options:
#   --dry-run      Show what would be removed without actually removing
#   --skip-qemu    Skip QEMU non-x86 packages (keep all QEMU packages)
#   --help         Show this help message

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

DRY_RUN=false
SKIP_QEMU=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --skip-qemu)
            SKIP_QEMU=true
            shift
            ;;
        --help)
            echo "Usage: $0 [--dry-run] [--skip-qemu] [--help]"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage"
            exit 1
            ;;
    esac
done

# Helper function to run commands
run_cmd() {
    local cmd="$1"
    local desc="$2"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        echo -e "${BLUE}[DRY RUN]${NC} Would run: $cmd"
        echo "         Description: $desc"
        return 0
    else
        echo -e "${GREEN}[RUNNING]${NC} $desc"
        eval "$cmd" || {
            echo -e "${RED}[ERROR]${NC} Failed: $desc"
            return 1
        }
    fi
}

echo -e "${BLUE}=== Remove Unused Packages Script ===${NC}"
echo ""

if [[ "$DRY_RUN" == "true" ]]; then
    echo -e "${YELLOW}DRY RUN MODE - No changes will be made${NC}"
    echo ""
fi

# ============================================================================
# 1. VERIFY CURRENT SYSTEM STATE
# ============================================================================
echo -e "${BLUE}--- Verifying Current System State ---${NC}"

# Check current kernel
CURRENT_KERNEL=$(uname -r)
echo "Current kernel: $CURRENT_KERNEL"

# Check QEMU VM status
if command -v virsh &> /dev/null; then
    echo "Checking QEMU/libvirt status..."
    VM_COUNT=$(virsh list --all 2>/dev/null | grep -c "macOS" || echo "0")
    if [[ "$VM_COUNT" -gt 0 ]]; then
        echo "  - Found macOS VM"
    else
        echo "  - No VMs found"
    fi
fi

# Check Go installation
if command -v go &> /dev/null; then
    GO_VERSION=$(go version 2>/dev/null | awk '{print $3}' || echo "unknown")
    echo "Go version: $GO_VERSION"
fi

echo ""

# ============================================================================
# 2. PURGE LIBREOFFICE CONFIG FILES (Already removed)
# ============================================================================
echo -e "${BLUE}--- Purge LibreOffice Config Files ---${NC}"

LIBREOFFICE_PKGS=($(dpkg -l | grep '^rc.*libreoffice' | awk '{print $2}' || true))
if [[ ${#LIBREOFFICE_PKGS[@]} -gt 0 ]]; then
    echo "Found LibreOffice packages with config files:"
    for pkg in "${LIBREOFFICE_PKGS[@]}"; do
        echo "  - $pkg"
    done
    echo ""
    run_cmd "sudo apt purge -y ${LIBREOFFICE_PKGS[*]}" "Purge LibreOffice config files (~144 MB)"
else
    echo "No LibreOffice config files to purge"
fi

echo ""

# ============================================================================
# 3. REMOVE GOLANG SOURCE CODE
# ============================================================================
echo -e "${BLUE}--- Remove Go Source Code ---${NC}"

if dpkg-query -Wf '${Package}\n' golang-1.22-src 2>/dev/null | grep -q '^golang-1.22-src$'; then
    SIZE=$(dpkg-query -Wf '${Installed-Size}' golang-1.22-src 2>/dev/null | awk '{printf "%.1f MB", $1/1024}' || echo "?")
    echo "Found golang-1.22-src ($SIZE) - source code not needed for building"
    run_cmd "sudo apt remove --purge -y golang-1.22-src" "Remove Go source code (keep compiler)"
else
    echo "golang-1.22-src not installed"
fi

echo ""

# ============================================================================
# 4. REMOVE OLD KERNEL HEADERS
# ============================================================================
echo -e "${BLUE}--- Remove Old Kernel Headers ---${NC}"

OLD_HEADERS="linux-headers-6.8.0-90 linux-headers-6.8.0-90-generic"
HEADERS_TO_REMOVE=()

for header in $OLD_HEADERS; do
    if dpkg-query -Wf '${Package}\n' "$header" 2>/dev/null | grep -q "^${header}$"; then
        HEADERS_TO_REMOVE+=("$header")
        size=$(dpkg-query -Wf '${Installed-Size}' "$header" 2>/dev/null | awk '{printf "%.1f MB", $1/1024}' || echo "?")
        echo "  - $header ($size)"
    fi
done

if [[ ${#HEADERS_TO_REMOVE[@]} -gt 0 ]]; then
    echo "Current kernel: $CURRENT_KERNEL (headers for 6.8.0-90 not needed)"
    run_cmd "sudo apt remove --purge -y ${HEADERS_TO_REMOVE[*]}" "Remove old kernel headers (~82 MB)"
else
    echo "No old kernel headers found"
fi

echo ""

# ============================================================================
# 5. REMOVE QEMU NON-X86 ARCHITECTURES (Optional)
# ============================================================================
if [[ "$SKIP_QEMU" != "true" ]]; then
    echo -e "${BLUE}--- Remove QEMU Non-x86 Architectures ---${NC}"
    
    QEMU_TO_REMOVE=(
        "qemu-efi-aarch64"
        "qemu-efi-arm"
        "qemu-system-arm"
        "qemu-system-mips"
        "qemu-system-ppc"
        "qemu-system-s390x"
        "qemu-system-sparc"
        "qemu-system-misc"
    )
    
    FOUND_QEMU=()
    for pkg in "${QEMU_TO_REMOVE[@]}"; do
        if dpkg-query -Wf '${Package}\n' "$pkg" 2>/dev/null | grep -q "^${pkg}$"; then
            FOUND_QEMU+=("$pkg")
            size=$(dpkg-query -Wf '${Installed-Size}' "$pkg" 2>/dev/null | awk '{printf "%.1f MB", $1/1024}' || echo "?")
            echo "  - $pkg ($size)"
        fi
    done
    
    if [[ ${#FOUND_QEMU[@]} -gt 0 ]]; then
        echo "Warning: These are needed for non-x86_64 VMs (ARM, ARM64, MIPS, etc.)"
        echo "You're using x86_64 macOS VM, so these are not needed"
        echo ""
        
        if [[ "$DRY_RUN" == "true" ]]; then
            echo -e "${BLUE}[DRY RUN]${NC} Would remove: ${FOUND_QEMU[*]}"
        else
            read -p "Remove QEMU non-x86 packages? (y/N): " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                run_cmd "sudo apt remove --purge -y ${FOUND_QEMU[*]}" "Remove QEMU non-x86 packages (~765 MB)"
            else
                echo -e "${YELLOW}Skipping QEMU package removal${NC}"
            fi
        fi
    else
        echo "No QEMU non-x86 packages found"
    fi
else
    echo -e "${YELLOW}--- Skipping QEMU Package Removal (--skip-qemu) ---${NC}"
fi

echo ""

# ============================================================================
# 6. CLEANUP
# ============================================================================
echo -e "${BLUE}--- Cleanup ---${NC}"

run_cmd "sudo apt autoremove -y" "Remove unused dependencies"
run_cmd "sudo apt autoclean" "Clean package cache"

echo ""

# ============================================================================
# 7. VERIFICATION
# ============================================================================
echo -e "${BLUE}--- Verification ---${NC}"

# Verify Go still works
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}')
    echo -e "${GREEN}✓${NC} Go compiler still works: $GO_VERSION"
else
    echo -e "${RED}✗${NC} Go compiler not found!"
fi

# Verify QEMU x86_64 still works (if not skipped)
if [[ "$SKIP_QEMU" != "true" ]] && command -v qemu-system-x86_64 &> /dev/null; then
    QEMU_VERSION=$(qemu-system-x86_64 --version 2>/dev/null | head -1 || echo "unknown")
    echo -e "${GREEN}✓${NC} QEMU x86_64 still works: $QEMU_VERSION"
fi

# Verify VM still accessible
if command -v virsh &> /dev/null; then
    if virsh list --all &> /dev/null; then
        echo -e "${GREEN}✓${NC} libvirt still accessible"
    else
        echo -e "${YELLOW}⚠${NC} libvirt not accessible (may need to check)"
    fi
fi

echo ""

# ============================================================================
# SUMMARY
# ============================================================================
echo -e "${GREEN}=== Cleanup Summary ===${NC}"
if [[ "$DRY_RUN" == "true" ]]; then
    echo -e "${YELLOW}This was a dry run. Run without --dry-run to perform removal.${NC}"
else
    echo "Cleanup completed!"
    echo ""
    echo "Expected space savings:"
    echo "  - LibreOffice config: ~144 MB"
    echo "  - Go source code: ~119 MB"
    echo "  - Old kernel headers: ~82 MB"
    if [[ "$SKIP_QEMU" != "true" ]] && [[ ${#FOUND_QEMU[@]} -gt 0 ]]; then
        echo "  - QEMU non-x86 packages: ~765 MB"
    fi
    echo ""
    echo "Total: ~1.1 GB (or ~350 MB if QEMU skipped)"
fi
echo ""
