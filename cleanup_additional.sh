#!/bin/bash
#
# Additional Cleanup Script
# Cleans caches, logs, and optional packages
#
# Usage:
#   ./cleanup_additional.sh [--dry-run] [--skip-debs] [--skip-packages] [--help]
#
# Options:
#   --dry-run         Show what would be cleaned without actually cleaning
#   --skip-debs       Skip removing downloaded .deb files
#   --skip-packages   Skip package removal (only clean caches/logs)
#   --help            Show this help message

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

DRY_RUN=false
SKIP_DEBS=false
SKIP_PACKAGES=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --skip-debs)
            SKIP_DEBS=true
            shift
            ;;
        --skip-packages)
            SKIP_PACKAGES=true
            shift
            ;;
        --help)
            echo "Usage: $0 [--dry-run] [--skip-debs] [--skip-packages] [--help]"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage"
            exit 1
            ;;
    esac
done

# Helper function to get size
get_size() {
    local path="$1"
    du -sh "$path" 2>/dev/null | cut -f1 || echo "0"
}

# Helper function to run commands
run_cmd() {
    local cmd="$1"
    local desc="$2"
    local size="$3"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        echo -e "${BLUE}[DRY RUN]${NC} Would run: $cmd"
        echo "         Description: $desc ($size)"
        return 0
    else
        echo -e "${GREEN}[RUNNING]${NC} $desc ($size)"
        eval "$cmd" || {
            echo -e "${RED}[ERROR]${NC} Failed: $desc"
            return 1
        }
    fi
}

echo -e "${BLUE}=== Additional Cleanup Script ===${NC}"
echo ""

if [[ "$DRY_RUN" == "true" ]]; then
    echo -e "${YELLOW}DRY RUN MODE - No changes will be made${NC}"
    echo ""
fi

TOTAL_FREED=0

# ============================================================================
# 1. GO BUILD CACHE CLEANUP
# ============================================================================
echo -e "${BLUE}--- Go Build Cache Cleanup ---${NC}"

if command -v go &> /dev/null; then
    GO_CACHE_DIR="$HOME/.cache/go-build"
    if [[ -d "$GO_CACHE_DIR" ]]; then
        GO_CACHE_SIZE=$(get_size "$GO_CACHE_DIR")
        echo "Go build cache: $GO_CACHE_SIZE"
        
        # Check actual size in bytes for calculation
        GO_CACHE_BYTES=$(du -sb "$GO_CACHE_DIR" 2>/dev/null | cut -f1 || echo "0")
        GO_CACHE_MB=$((GO_CACHE_BYTES / 1024 / 1024))
        
        if [[ "$GO_CACHE_BYTES" -gt 0 ]]; then
            run_cmd "go clean -cache" "Clean Go build cache" "$GO_CACHE_SIZE"
            TOTAL_FREED=$((TOTAL_FREED + GO_CACHE_MB))
        else
            echo "Go build cache is empty"
        fi
    else
        echo "Go build cache directory not found"
    fi
else
    echo "Go not found (using /usr/local/go or not in PATH)"
    # Try direct cleanup if cache exists
    GO_CACHE_DIR="$HOME/.cache/go-build"
    if [[ -d "$GO_CACHE_DIR" ]]; then
        GO_CACHE_SIZE=$(get_size "$GO_CACHE_DIR")
        echo "Found Go build cache: $GO_CACHE_SIZE"
        GO_CACHE_BYTES=$(du -sb "$GO_CACHE_DIR" 2>/dev/null | cut -f1 || echo "0")
        GO_CACHE_MB=$((GO_CACHE_BYTES / 1024 / 1024))
        if [[ "$GO_CACHE_BYTES" -gt 0 ]]; then
            run_cmd "rm -rf $GO_CACHE_DIR" "Remove Go build cache" "$GO_CACHE_SIZE"
            TOTAL_FREED=$((TOTAL_FREED + GO_CACHE_MB))
        fi
    fi
fi

echo ""

# ============================================================================
# 2. GOOGLE CHROME CACHE CLEANUP
# ============================================================================
echo -e "${BLUE}--- Google Chrome Cache Cleanup ---${NC}"

CHROME_CACHE_DIR="$HOME/.cache/google-chrome"
if [[ -d "$CHROME_CACHE_DIR" ]]; then
    CHROME_CACHE_SIZE=$(get_size "$CHROME_CACHE_DIR")
    echo "Chrome cache: $CHROME_CACHE_SIZE"
    
    CHROME_CACHE_BYTES=$(du -sb "$CHROME_CACHE_DIR" 2>/dev/null | cut -f1 || echo "0")
    CHROME_CACHE_MB=$((CHROME_CACHE_BYTES / 1024 / 1024))
    
    if [[ "$CHROME_CACHE_BYTES" -gt 0 ]]; then
        if [[ "$DRY_RUN" == "true" ]]; then
            echo -e "${BLUE}[DRY RUN]${NC} Would remove: $CHROME_CACHE_DIR"
            echo "         Description: Clean Chrome cache ($CHROME_CACHE_SIZE)"
        else
            echo -e "${YELLOW}Note: Chrome should be closed before cleaning cache${NC}"
            read -p "Remove Chrome cache? (y/N): " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                run_cmd "rm -rf $CHROME_CACHE_DIR" "Clean Chrome cache" "$CHROME_CACHE_SIZE"
                TOTAL_FREED=$((TOTAL_FREED + CHROME_CACHE_MB))
            else
                echo -e "${YELLOW}Skipping Chrome cache cleanup${NC}"
            fi
        fi
    else
        echo "Chrome cache is empty"
    fi
else
    echo "Chrome cache directory not found"
fi

echo ""

# ============================================================================
# 3. JOURNAL LOGS CLEANUP
# ============================================================================
echo -e "${BLUE}--- Journal Logs Cleanup ---${NC}"

if command -v journalctl &> /dev/null; then
    # Get journal size before cleanup
    JOURNAL_SIZE_BEFORE=$(journalctl --disk-usage 2>/dev/null | awk '{print $1}' || echo "0")
    
    if [[ "$JOURNAL_SIZE_BEFORE" != "0" ]] && [[ -n "$JOURNAL_SIZE_BEFORE" ]]; then
        echo "Journal logs: $JOURNAL_SIZE_BEFORE"
        run_cmd "sudo journalctl --vacuum-time=7d" "Clean journal logs (keep last 7 days)" "$JOURNAL_SIZE_BEFORE"
    else
        echo "Cannot determine journal size (may need sudo)"
        run_cmd "sudo journalctl --vacuum-time=7d" "Clean journal logs (keep last 7 days)" "?"
    fi
else
    echo "journalctl not found"
fi

echo ""

# ============================================================================
# 4. DOWNLOADED .DEB FILES CLEANUP
# ============================================================================
if [[ "$SKIP_DEBS" != "true" ]]; then
    echo -e "${BLUE}--- Downloaded .deb Files Cleanup ---${NC}"
    
    DEB_FILES=($(find ~/Downloads -maxdepth 1 -name "*.deb" -type f 2>/dev/null || true))
    
    if [[ ${#DEB_FILES[@]} -gt 0 ]]; then
        echo "Found ${#DEB_FILES[@]} .deb files in ~/Downloads:"
        DEB_TOTAL_BYTES=0
        for deb in "${DEB_FILES[@]}"; do
            DEB_SIZE=$(du -sb "$deb" 2>/dev/null | cut -f1 || echo "0")
            DEB_TOTAL_BYTES=$((DEB_TOTAL_BYTES + DEB_SIZE))
            DEB_SIZE_MB=$((DEB_SIZE / 1024 / 1024))
            echo "  - $(basename "$deb") ($(get_size "$deb"))"
        done
        
        DEB_TOTAL_MB=$((DEB_TOTAL_BYTES / 1024 / 1024))
        DEB_TOTAL_SIZE=$(du -sh ~/Downloads/*.deb 2>/dev/null | awk '{sum+=$1} END {if (sum) print sum " total"}')
        
        if [[ "$DRY_RUN" == "true" ]]; then
            echo -e "${BLUE}[DRY RUN]${NC} Would remove: ${#DEB_FILES[@]} .deb files (~${DEB_TOTAL_MB} MB)"
        else
            echo -e "${YELLOW}Note: These are installer files. Only remove if packages are already installed.${NC}"
            read -p "Remove downloaded .deb files? (y/N): " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                run_cmd "rm ~/Downloads/*.deb" "Remove downloaded .deb files" "~${DEB_TOTAL_MB} MB"
                TOTAL_FREED=$((TOTAL_FREED + DEB_TOTAL_MB))
            else
                echo -e "${YELLOW}Skipping .deb file removal${NC}"
            fi
        fi
    else
        echo "No .deb files found in ~/Downloads"
    fi
    
    echo ""
fi

# ============================================================================
# 5. PACKAGE CLEANUP
# ============================================================================
if [[ "$SKIP_PACKAGES" != "true" ]]; then
    echo -e "${BLUE}--- Package Cleanup ---${NC}"
    
    # xubuntu-docs
    if dpkg-query -Wf '${Package}\n' xubuntu-docs 2>/dev/null | grep -q '^xubuntu-docs$'; then
        XUBUNTU_DOCS_SIZE=$(dpkg-query -Wf '${Installed-Size}' xubuntu-docs 2>/dev/null | awk '{printf "%.1f MB", $1/1024}' || echo "?")
        echo "Found xubuntu-docs ($XUBUNTU_DOCS_SIZE) - documentation"
        
        if [[ "$DRY_RUN" == "true" ]]; then
            echo -e "${BLUE}[DRY RUN]${NC} Would remove: xubuntu-docs"
        else
            read -p "Remove xubuntu-docs? (y/N): " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                XUBUNTU_DOCS_MB=$(dpkg-query -Wf '${Installed-Size}' xubuntu-docs 2>/dev/null | awk '{print int($1/1024)}' || echo "0")
                run_cmd "sudo apt remove --purge -y xubuntu-docs" "Remove xubuntu-docs" "$XUBUNTU_DOCS_SIZE"
                TOTAL_FREED=$((TOTAL_FREED + XUBUNTU_DOCS_MB))
            fi
        fi
    else
        echo "xubuntu-docs not installed"
    fi
    
    # GIMP data (if GIMP not installed)
    if dpkg-query -Wf '${Package}\n' gimp-data 2>/dev/null | grep -q '^gimp-data$'; then
        if ! dpkg-query -Wf '${Package}\n' 2>/dev/null | grep -q '^gimp$'; then
            GIMP_DATA_SIZE=$(dpkg-query -Wf '${Installed-Size}' gimp-data 2>/dev/null | awk '{printf "%.1f MB", $1/1024}' || echo "?")
            echo "Found gimp-data ($GIMP_DATA_SIZE) - GIMP not installed"
            
            if [[ "$DRY_RUN" == "true" ]]; then
                echo -e "${BLUE}[DRY RUN]${NC} Would remove: gimp-data"
            else
                read -p "Remove gimp-data? (y/N): " -n 1 -r
                echo
                if [[ $REPLY =~ ^[Yy]$ ]]; then
                    GIMP_DATA_MB=$(dpkg-query -Wf '${Installed-Size}' gimp-data 2>/dev/null | awk '{print int($1/1024)}' || echo "0")
                    run_cmd "sudo apt remove --purge -y gimp-data" "Remove gimp-data" "$GIMP_DATA_SIZE"
                    TOTAL_FREED=$((TOTAL_FREED + GIMP_DATA_MB))
                fi
            fi
        else
            echo "GIMP is installed, keeping gimp-data"
        fi
    fi
    
    echo ""
fi

# ============================================================================
# 6. CLEANUP
# ============================================================================
echo -e "${BLUE}--- Final Cleanup ---${NC}"

run_cmd "sudo apt autoremove -y" "Remove unused dependencies" "varies"
run_cmd "sudo apt autoclean" "Clean package cache" "varies"

echo ""

# ============================================================================
# SUMMARY
# ============================================================================
echo -e "${GREEN}=== Cleanup Summary ===${NC}"
if [[ "$DRY_RUN" == "true" ]]; then
    echo -e "${YELLOW}This was a dry run. Run without --dry-run to perform cleanup.${NC}"
    echo ""
    echo "Expected space savings:"
    echo "  - Go build cache: ~1.3 GB"
    echo "  - Chrome cache: ~432 MB"
    echo "  - Journal logs: ~336 MB"
    echo "  - Downloaded .deb files: ~510 MB"
    echo "  - Packages: ~160 MB"
    echo ""
    echo "Total: ~2.7 GB"
else
    TOTAL_FREED_GB=$((TOTAL_FREED / 1024))
    TOTAL_FREED_MB_REMAINDER=$((TOTAL_FREED % 1024))
    
    if [[ "$TOTAL_FREED_GB" -gt 0 ]]; then
        echo "Space freed: ~${TOTAL_FREED_GB}.${TOTAL_FREED_MB_REMAINDER} GB (${TOTAL_FREED} MB)"
    else
        echo "Space freed: ~${TOTAL_FREED} MB"
    fi
fi

echo ""
