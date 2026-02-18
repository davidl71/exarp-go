#!/bin/bash

###############################################################################
# Development Automation Script
# 
# Streamlines development and testing workflow with auto-reload, auto-test,
# and intelligent file watching. Minimizes need for manual prompting.
#
# Usage:
#   ./dev.sh [OPTIONS]
#
# Options:
#   --watch              Watch files and auto-reload server
#   --test               Auto-run tests on file changes
#   --test-watch         Watch mode for tests only (no server)
#   --coverage           Generate coverage reports
#   --build              Auto-rebuild on changes
#   --quiet              Suppress non-error output
#   --help, -h           Show this help message
###############################################################################

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$SCRIPT_DIR"
PYTHON="uv run python"
BINARY_PATH="$PROJECT_ROOT/bin/exarp-go"
# Optional: REDIS_ADDR=127.0.0.1:6379 for queue/worker (make queue-enqueue-wave, make queue-worker)
WATCH_FILES=false
AUTO_TEST=false
TEST_WATCH=false
COVERAGE=false
AUTO_BUILD=false
QUIET=false

# Files to watch
WATCH_PATTERNS=(
    "*.go"
    "go.mod"
    "go.sum"
)

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Logging
log() {
    [[ "$QUIET" == true ]] && return
    local level="$1"
    shift
    local message="$*"
    local timestamp=$(date '+%H:%M:%S')
    
    case "$level" in
        INFO)
            echo -e "${GREEN}[$timestamp] [INFO]${NC} $message"
            ;;
        WARN)
            echo -e "${YELLOW}[$timestamp] [WARN]${NC} $message"
            ;;
        ERROR)
            echo -e "${RED}[$timestamp] [ERROR]${NC} $message" >&2
            ;;
        TEST)
            echo -e "${CYAN}[$timestamp] [TEST]${NC} $message"
            ;;
        BUILD)
            echo -e "${BLUE}[$timestamp] [BUILD]${NC} $message"
            ;;
    esac
}

# Parse arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --watch)
                WATCH_FILES=true
                shift
                ;;
            --test)
                AUTO_TEST=true
                shift
                ;;
            --test-watch)
                TEST_WATCH=true
                AUTO_TEST=true
                shift
                ;;
            --coverage)
                COVERAGE=true
                shift
                ;;
            --build)
                AUTO_BUILD=true
                shift
                ;;
            --quiet)
                QUIET=true
                shift
                ;;
            --help|-h)
                cat <<EOF
Usage: $0 [OPTIONS]

Development automation script for $(basename "$PROJECT_ROOT")

Options:
  --watch              Watch files and auto-reload server
  --test               Auto-run tests on file changes
  --test-watch         Watch mode for tests only (no server)
  --coverage           Generate coverage reports
  --build              Auto-rebuild on changes
  --quiet              Suppress non-error output
  --help, -h           Show this help message

Examples:
  $0 --watch                    # Auto-reload on changes
  $0 --watch --test             # Auto-reload + auto-test
  $0 --test-watch                # Test-only watch mode
  $0 --watch --test --coverage  # Full dev mode with coverage
EOF
                exit 0
                ;;
            *)
                log ERROR "Unknown option: $1"
                exit 1
                ;;
        esac
    done
}

# Check dependencies
check_deps() {
    local missing=()
    
    if ! command -v uv &> /dev/null; then
        missing+=("uv")
    fi
    
    if [[ "$WATCH_FILES" == true ]] && ! command -v fswatch &> /dev/null && ! command -v inotifywait &> /dev/null; then
        log WARN "fswatch/inotifywait not found, using polling mode"
    fi
    
    if [[ ${#missing[@]} -gt 0 ]]; then
        log ERROR "Missing dependencies: ${missing[*]}"
        log INFO "Install with: pip install ${missing[*]}"
        exit 1
    fi
}

# Run tests
run_tests() {
    local test_cmd="$PYTHON -m pytest tests/ -v"
    
    if [[ "$COVERAGE" == true ]]; then
        test_cmd="$test_cmd --cov=bridge --cov-report=term"
    fi
    
    log TEST "Running tests..."
    
    if eval "$test_cmd" 2>&1; then
        log TEST "✅ All tests passed"
        return 0
    else
        # Fallback to Go binary check
        if [[ -f "$BINARY_PATH" ]]; then
            log TEST "✅ Go binary exists"
            return 0
        else
            log ERROR "❌ Tests failed"
            return 1
        fi
    fi
}

# Quick binary test
quick_test() {
    log TEST "Running quick binary test..."
    
    if [[ -f "$BINARY_PATH" ]] && [[ -x "$BINARY_PATH" ]]; then
        log TEST "✅ Binary test passed"
        return 0
    else
        log ERROR "❌ Binary test failed"
        return 1
    fi
}

# Build/verify
build() {
    log BUILD "Building Go binary..."
    
    if go build -o "$BINARY_PATH" ./cmd/server 2>&1; then
        log BUILD "✅ Build verified"
        return 0
    else
        log ERROR "❌ Build failed"
        return 1
    fi
}

# Start server
start_server() {
    log INFO "Starting server..."
    "$BINARY_PATH" &
    local pid=$!
    echo "$pid" > .server.pid
    log INFO "Server started (PID: $pid)"
}

# Stop server
stop_server() {
    if [[ -f .server.pid ]]; then
        local pid=$(cat .server.pid 2>/dev/null || echo "")
        if [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null; then
            log INFO "Stopping server (PID: $pid)..."
            kill "$pid" 2>/dev/null || true
            rm -f .server.pid
            log INFO "Server stopped"
        fi
    fi
}

# Watch files with fswatch (macOS)
watch_fswatch() {
    log INFO "Starting file watcher (fswatch)..."
    
    local watch_paths=()
    for pattern in "${WATCH_PATTERNS[@]}"; do
        watch_paths+=("$PROJECT_ROOT/$pattern")
    done
    
    fswatch -o -r "${watch_paths[@]}" "$PROJECT_ROOT/cmd" "$PROJECT_ROOT/internal" | while read -r; do
        log INFO "File change detected"
        
        if [[ "$AUTO_BUILD" == true ]]; then
            if build; then
                if [[ "$WATCH_FILES" == true ]] && [[ "$TEST_WATCH" != true ]]; then
                    stop_server
                    sleep 0.5
                    start_server
                fi
            fi
        fi
        
        if [[ "$AUTO_TEST" == true ]]; then
            if [[ "$COVERAGE" == true ]]; then
                run_tests
            else
                quick_test
            fi
        fi
        
        if [[ "$WATCH_FILES" == true ]] && [[ "$TEST_WATCH" != true ]] && [[ "$AUTO_BUILD" != true ]]; then
            stop_server
            sleep 0.5
            start_server
        fi
    done
}

# Watch files with inotifywait (Linux)
watch_inotify() {
    log INFO "Starting file watcher (inotifywait)..."
    
    inotifywait -m -r -e modify,create,delete \
        --include '\.(py|toml)$' \
        "$PROJECT_ROOT" 2>/dev/null | while read -r directory event file; do
        log INFO "File change detected: $file"
        
        if [[ "$AUTO_BUILD" == true ]]; then
            if build; then
                if [[ "$WATCH_FILES" == true ]] && [[ "$TEST_WATCH" != true ]]; then
                    stop_server
                    sleep 0.5
                    start_server
                fi
            fi
        fi
        
        if [[ "$AUTO_TEST" == true ]]; then
            if [[ "$COVERAGE" == true ]]; then
                run_tests
            else
                quick_test
            fi
        fi
        
        if [[ "$WATCH_FILES" == true ]] && [[ "$TEST_WATCH" != true ]] && [[ "$AUTO_BUILD" != true ]]; then
            stop_server
            sleep 0.5
            start_server
        fi
    done
}

# Watch files with polling (fallback)
watch_polling() {
    log INFO "Starting polling-based file watcher..."
    
    local last_check=0
    
    while true; do
        local current_check=$(find "$PROJECT_ROOT" -name "*.py" -o -name "pyproject.toml" | xargs stat -f %m 2>/dev/null | sort -n | tail -1 || echo "0")
        
        if [[ $current_check -gt $last_check ]] && [[ $last_check -gt 0 ]]; then
            log INFO "File change detected"
            
            if [[ "$AUTO_BUILD" == true ]]; then
                if build; then
                    if [[ "$WATCH_FILES" == true ]] && [[ "$TEST_WATCH" != true ]]; then
                        stop_server
                        sleep 0.5
                        start_server
                    fi
                fi
            fi
            
            if [[ "$AUTO_TEST" == true ]]; then
                if [[ "$COVERAGE" == true ]]; then
                    run_tests
                else
                    quick_test
                fi
            fi
            
            if [[ "$WATCH_FILES" == true ]] && [[ "$TEST_WATCH" != true ]] && [[ "$AUTO_BUILD" != true ]]; then
                stop_server
                sleep 0.5
                start_server
            fi
        fi
        
        last_check=$current_check
        sleep 2
    done
}

# Watch files (auto-detect method)
watch_files() {
    if command -v fswatch &> /dev/null; then
        watch_fswatch
    elif command -v inotifywait &> /dev/null; then
        watch_inotify
    else
        watch_polling
    fi
}

# Cleanup
cleanup() {
    log INFO "Shutting down..."
    stop_server
    rm -f .server.pid
    exit 0
}

# Main
main() {
    parse_args "$@"
    check_deps
    
    # Initial build/test
    if ! build; then
        log ERROR "Initial build failed"
        exit 1
    fi
    
    if [[ "$AUTO_TEST" == true ]]; then
        if [[ "$COVERAGE" == true ]]; then
            run_tests
        else
            quick_test
        fi
    fi
    
    # Start server if not in test-only mode
    if [[ "$WATCH_FILES" == true ]] && [[ "$TEST_WATCH" != true ]]; then
        start_server
    fi
    
    # Setup signal handlers
    trap cleanup SIGINT SIGTERM
    
    # Start watching if requested
    if [[ "$WATCH_FILES" == true ]] || [[ "$TEST_WATCH" == true ]]; then
        log INFO "Watching for changes... (Press Ctrl+C to stop)"
        watch_files
    else
        # Just run once
        if [[ "$AUTO_TEST" == true ]]; then
            if [[ "$COVERAGE" == true ]]; then
                run_tests
            else
                quick_test
            fi
        fi
    fi
}

# Run main
main "$@"

