#!/bin/bash
#
# Ubuntu Cleanup Script
# Analyzes and optionally removes unnecessary packages, caches, and files
#
# Usage:
#   ./cleanup_ubuntu.sh [--dry-run] [--aggressive] [--help]
#
# Options:
#   --dry-run      Show what would be removed without actually removing
#   --aggressive   More aggressive cleanup (removes more packages)
#   --help         Show this help message

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

DRY_RUN=false
AGGRESSIVE=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --aggressive)
            AGGRESSIVE=true
            shift
            ;;
        --help)
            echo "Usage: $0 [--dry-run] [--aggressive] [--help]"
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

# Helper function to get package size in MB
get_pkg_size_mb() {
    local pkg="$1"
    dpkg-query -Wf '${Installed-Size}' "$pkg" 2>/dev/null | awk '{printf "%.1f", $1/1024}' || echo "0"
}

echo -e "${BLUE}=== Ubuntu Cleanup Script ===${NC}"
echo ""

if [[ "$DRY_RUN" == "true" ]]; then
    echo -e "${YELLOW}DRY RUN MODE - No changes will be made${NC}"
    echo ""
fi

# ============================================================================
# 1. ANALYZE DOCUMENTATION PACKAGES
# ============================================================================
echo -e "${BLUE}--- Analyzing Documentation Packages ---${NC}"

DOC_PKGS=$(dpkg-query -Wf '${Package}\n' 2>/dev/null | grep -E 'doc$|-doc' || true)
DOC_COUNT=$(echo "$DOC_PKGS" | grep -c . || echo "0")
DOC_SIZE_MB=0

if [[ "$DOC_COUNT" -gt 0 ]]; then
    echo "Found $DOC_COUNT documentation packages:"
    while IFS= read -r pkg; do
        [[ -z "$pkg" ]] && continue
        size=$(get_pkg_size_mb "$pkg")
        doc_size=$(echo "$size" | awk '{print $1}')
        DOC_SIZE_MB=$(echo "$DOC_SIZE_MB $doc_size" | awk '{print $1 + $2}')
        echo "  - $pkg (${size} MB)"
    done <<< "$DOC_PKGS"
    echo "Total documentation size: ~${DOC_SIZE_MB} MB"
else
    echo "No documentation packages found"
fi

echo ""

# ============================================================================
# 2. ANALYZE DEVELOPMENT PACKAGES
# ============================================================================
echo -e "${BLUE}--- Analyzing Development Packages ---${NC}"

DEV_PKGS=$(dpkg-query -Wf '${Package}\n' 2>/dev/null | grep '\-dev$' || true)
if [[ -z "$DEV_PKGS" ]]; then
    DEV_COUNT=0
else
    DEV_COUNT=$(echo "$DEV_PKGS" | wc -l)
fi

# Check which dev packages are potentially needed
NEEDED_DEV=()
POTENTIALLY_UNNEEDED=()

if [[ "$DEV_COUNT" -gt 0 ]]; then
    echo "Found $DEV_COUNT development packages:"
    
    while IFS= read -r pkg; do
        [[ -z "$pkg" ]] && continue
        size=$(get_pkg_size_mb "$pkg")
        
        # Check if this dev package might be needed for Go projects
        case "$pkg" in
            *python*dev*)
                # Python dev packages - not needed for Go projects (unless compiling Python extensions)
                POTENTIALLY_UNNEEDED+=("$pkg (${size} MB) - Python headers, not needed for Go projects")
                ;;
            *stdc++*dev*)
                # C++ standard library dev - not needed for Go projects (unless using CGO with C++)
                POTENTIALLY_UNNEEDED+=("$pkg (${size} MB) - C++ headers, not needed for Go projects")
                ;;
            systemd-dev)
                # Systemd dev - not needed for Go projects
                POTENTIALLY_UNNEEDED+=("$pkg (${size} MB) - Systemd headers, not needed for Go projects")
                ;;
            autotools-dev)
                # Autotools - not needed for Go projects
                POTENTIALLY_UNNEEDED+=("$pkg (${size} MB) - Autotools, not needed for Go projects")
                ;;
            libc6-dev|linux-libc-dev)
                # C library headers - needed for CGO builds (keep)
                NEEDED_DEV+=("$pkg (${size} MB) - C headers, needed for CGO builds")
                ;;
            libgcc-*-dev|libcrypt-dev|libexpat1-dev|zlib1g-dev)
                # GCC/Crypto libraries - needed for CGO builds (keep)
                NEEDED_DEV+=("$pkg (${size} MB) - Library headers, needed for CGO builds")
                ;;
            dpkg-dev|manpages-dev)
                # Build tools - might be useful but not essential
                NEEDED_DEV+=("$pkg (${size} MB) - Build tools, useful but not essential")
                ;;
            *)
                NEEDED_DEV+=("$pkg (${size} MB) - Unknown if needed (keeping to be safe)")
                ;;
        esac
    done <<< "$DEV_PKGS"
    
    if [[ ${#NEEDED_DEV[@]} -gt 0 ]]; then
        echo ""
        echo "Potentially needed dev packages:"
        for pkg_info in "${NEEDED_DEV[@]}"; do
            echo "  - $pkg_info"
        done
    fi
    
    if [[ ${#POTENTIALLY_UNNEEDED[@]} -gt 0 ]]; then
        echo ""
        echo "Potentially unneeded dev packages:"
        for pkg_info in "${POTENTIALLY_UNNEEDED[@]}"; do
            echo "  - $pkg_info"
        done
    fi
else
    echo "No development packages found"
fi

echo ""

# ============================================================================
# 3. SAFE CLEANUP (Always safe)
# ============================================================================
echo -e "${BLUE}--- Safe Cleanup Operations ---${NC}"

# APT cleanup
run_cmd "sudo apt autoremove -y" "Remove unused packages and dependencies"
run_cmd "sudo apt autoclean" "Clean package cache"
run_cmd "sudo apt clean" "Remove all package downloads"

# User cache cleanup
CACHE_SIZE=$(du -sh ~/.cache 2>/dev/null | cut -f1 || echo "0")
if [[ "$CACHE_SIZE" != "0" ]] && [[ "$CACHE_SIZE" != "" ]]; then
    echo "User cache size: $CACHE_SIZE"
    run_cmd "rm -rf ~/.cache/*" "Clean user cache (~/.cache)"
fi

# Go cache cleanup (if Go is installed)
if command -v go &> /dev/null; then
    GO_CACHE_SIZE=$(du -sh "$(go env GOCACHE)" 2>/dev/null | cut -f1 || echo "0")
    if [[ "$GO_CACHE_SIZE" != "0" ]] && [[ "$GO_CACHE_SIZE" != "" ]]; then
        echo "Go cache size: $GO_CACHE_SIZE"
        run_cmd "go clean -cache -modcache -testcache" "Clean Go build cache"
    fi
fi

# Python cache cleanup
if command -v pip &> /dev/null; then
    run_cmd "pip cache purge" "Clean pip cache"
fi

if command -v pipx &> /dev/null; then
    run_cmd "pipx cache clean" "Clean pipx cache"
fi

# npm cache cleanup
if command -v npm &> /dev/null; then
    run_cmd "npm cache clean --force" "Clean npm cache"
fi

# Docker cleanup (if Docker is installed)
if command -v docker &> /dev/null; then
    run_cmd "docker system prune -a --volumes -f" "Clean Docker system (images, containers, volumes)"
fi

# Journal logs cleanup
run_cmd "sudo journalctl --vacuum-time=3d" "Clean systemd journal logs (keep last 3 days)"

echo ""

# ============================================================================
# 4. DOCUMENTATION CLEANUP (Optional)
# ============================================================================
if [[ "$DOC_COUNT" -gt 0 ]]; then
    echo -e "${BLUE}--- Documentation Package Cleanup ---${NC}"
    echo "Found documentation packages that can be removed:"
    
    DOC_TO_REMOVE=()
    while IFS= read -r pkg; do
        [[ -z "$pkg" ]] && continue
        # Skip essential documentation
        if [[ ! "$pkg" =~ ^(libc6-doc|man-db|manpages)$ ]]; then
            DOC_TO_REMOVE+=("$pkg")
        fi
    done <<< "$DOC_PKGS"
    
    if [[ ${#DOC_TO_REMOVE[@]} -gt 0 ]]; then
        echo "Packages to remove: ${DOC_TO_REMOVE[*]}"
        if [[ "$DRY_RUN" == "true" ]]; then
            echo -e "${BLUE}[DRY RUN]${NC} Would remove: ${DOC_TO_REMOVE[*]}"
        else
            read -p "Remove documentation packages? (y/N): " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                run_cmd "sudo apt remove --purge -y ${DOC_TO_REMOVE[*]}" "Remove documentation packages"
            fi
        fi
    fi
    echo ""
fi

# ============================================================================
# 5. DEVELOPMENT PACKAGE CLEANUP (Optional, more aggressive)
# ============================================================================
if [[ ${#POTENTIALLY_UNNEEDED[@]} -gt 0 ]]; then
    echo -e "${BLUE}--- Development Package Cleanup ---${NC}"
    echo "Found development packages that might not be needed:"
    
    DEV_TO_REMOVE=()
    for pkg_info in "${POTENTIALLY_UNNEEDED[@]}"; do
        pkg=$(echo "$pkg_info" | cut -d' ' -f1)
        DEV_TO_REMOVE+=("$pkg")
    done
    
    if [[ ${#DEV_TO_REMOVE[@]} -gt 0 ]]; then
        echo "⚠️  WARNING: Removing dev packages might break compilation of some software"
        echo "Packages to consider removing:"
        for pkg_info in "${POTENTIALLY_UNNEEDED[@]}"; do
            echo "  - $pkg_info"
        done
        
        if [[ "$DRY_RUN" == "true" ]]; then
            echo -e "${BLUE}[DRY RUN]${NC} Would remove: ${DEV_TO_REMOVE[*]}"
        else
            if [[ "$AGGRESSIVE" == "true" ]]; then
                read -p "Remove development packages? (y/N): " -n 1 -r
                echo
                if [[ $REPLY =~ ^[Yy]$ ]]; then
                    run_cmd "sudo apt remove --purge -y ${DEV_TO_REMOVE[*]}" "Remove development packages"
                fi
            else
                echo -e "${YELLOW}Skipped (use --aggressive to enable)${NC}"
            fi
        fi
    fi
    echo ""
fi

# ============================================================================
# 6. FIND LARGE FILES/DIRECTORIES
# ============================================================================
echo -e "${BLUE}--- Large Files and Directories ---${NC}"

echo "Top 10 largest directories (requiring sudo):"
if [[ "$DRY_RUN" == "true" ]]; then
    echo -e "${BLUE}[DRY RUN]${NC} Would check: sudo du -h --max-depth=1 / 2>/dev/null | sort -hr | head -11"
else
    sudo du -h --max-depth=1 / 2>/dev/null | sort -hr | head -11 || true
fi

echo ""
echo "Large files (>100MB):"
if [[ "$DRY_RUN" == "true" ]]; then
    echo -e "${BLUE}[DRY RUN]${NC} Would check: sudo find / -type f -size +100M 2>/dev/null | head -20"
else
    sudo find / -type f -size +100M 2>/dev/null | head -20 || true
fi

echo ""

# ============================================================================
# SUMMARY
# ============================================================================
echo -e "${GREEN}=== Cleanup Summary ===${NC}"
if [[ "$DRY_RUN" == "true" ]]; then
    echo -e "${YELLOW}This was a dry run. Run without --dry-run to perform cleanup.${NC}"
else
    echo "Cleanup completed!"
fi
echo ""
echo "To free more space:"
echo "  1. Review large files/directories listed above"
echo "  2. Remove unused snap packages: snap list && sudo snap remove <package>"
echo "  3. Remove old kernels: sudo apt autoremove --purge"
echo "  4. Check /var/log for large log files"
echo ""
