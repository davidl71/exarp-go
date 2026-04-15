#!/usr/bin/env bash
# check-navigability.sh — Detect missing AI navigability aids in Go source files.
# Checks: file-level comments, package docs, magic status/priority strings, large files.
# Exit 0 = clean, exit 1 = issues found (with summary).
#
# Usage: scripts/check-navigability.sh [--fix-hint] [path...]
#   --fix-hint  Show suggested fix for each issue
#   path...     Limit scan to these directories (default: internal/ cmd/)

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

FIX_HINT=false
SCAN_DIRS=()

for arg in "$@"; do
  case "$arg" in
    --fix-hint) FIX_HINT=true ;;
    *) SCAN_DIRS+=("$arg") ;;
  esac
done

if [ ${#SCAN_DIRS[@]} -eq 0 ]; then
  SCAN_DIRS=("internal/" "cmd/")
fi

ISSUES=0
WARNINGS=0

# Track package dirs that have docs (use temp files for bash 3 compat)
PKG_DOC_FOUND=$(mktemp)
PKG_ALL_DIRS=$(mktemp)
trap 'rm -f "$PKG_DOC_FOUND" "$PKG_ALL_DIRS"' EXIT

issue() {
  local file="$1" line="$2" msg="$3"
  echo "ERROR: $file:$line: $msg"
  ISSUES=$((ISSUES + 1))
}

warn() {
  local file="$1" msg="$2"
  echo "WARN:  $file: $msg"
  WARNINGS=$((WARNINGS + 1))
}

hint() {
  if $FIX_HINT; then
    echo "  FIX: $1"
  fi
}

# ---------- Check 1: File-level comments ----------
check_file_comment() {
  local file="$1"
  local basename
  basename="$(basename "$file")"

  case "$basename" in
    *_test.go|*_bench_test.go|*.pb.go|*.pb.*.go|doc.go) return ;;
  esac

  local line_count
  line_count=$(wc -l < "$file" | tr -d ' ')
  [ "$line_count" -le 50 ] && return

  local has_comment=false
  local i=0
  while IFS= read -r line && [ $i -lt 5 ]; do
    if echo "$line" | grep -qE '^[[:space:]]*//' ; then
      has_comment=true
      break
    fi
    i=$((i + 1))
  done < "$file"

  if ! $has_comment; then
    issue "$file" 1 "missing file-level comment (file has $line_count lines)"
    hint "Add a comment like: // ${basename} — <purpose>. before the package line."
  fi
}

# ---------- Check 2: Package doc comments ----------
check_package_doc() {
  local file="$1"
  local dir
  dir="$(dirname "$file")"

  [[ "$(basename "$file")" == *_test.go ]] && return

  echo "$dir" >> "$PKG_ALL_DIRS"

  if grep -qE '^// Package [a-zA-Z]' "$file" 2>/dev/null; then
    echo "$dir" >> "$PKG_DOC_FOUND"
  fi
}

report_missing_package_docs() {
  if [ ! -s "$PKG_ALL_DIRS" ]; then
    return
  fi
  local all_dirs found_dirs
  all_dirs=$(sort -u "$PKG_ALL_DIRS")
  found_dirs=$(sort -u "$PKG_DOC_FOUND" 2>/dev/null || echo "")

  while IFS= read -r dir; do
    [ -z "$dir" ] && continue
    if ! echo "$found_dirs" | grep -qxF "$dir"; then
      warn "$dir/" "missing package doc comment (// Package <name> ...)"
      hint "Add '// Package <name> ...' in the primary .go file of $dir/."
    fi
  done <<< "$all_dirs"
}

# ---------- Check 3: Magic status/priority strings ----------
check_magic_strings() {
  local file="$1"
  local basename
  basename="$(basename "$file")"

  [[ "$file" != internal/tools/* ]] && return
  case "$basename" in
    *_test.go|*_bench_test.go|constants.go|registry.go) return ;;
  esac

  # Use grep for speed: find candidate lines, then filter false positives
  local matches
  matches=$(grep -nE '"(Todo|In Progress|Done|Review|Cancelled|Blocked)"' "$file" 2>/dev/null || true)
  [ -z "$matches" ] && return

  while IFS= read -r match_line; do
    [ -z "$match_line" ] && continue
    local line_num="${match_line%%:*}"
    local content="${match_line#*:}"

    # Skip comments, model refs, format strings
    echo "$content" | grep -qE '^[[:space:]]*//' && continue
    echo "$content" | grep -qF 'models.' && continue
    echo "$content" | grep -qE 'fmt\.(Sprintf|Errorf|Printf)' && continue

    local pat
    for pat in '"Todo"' '"In Progress"' '"Done"' '"Review"' '"Cancelled"' '"Blocked"'; do
      if echo "$content" | grep -qF "$pat"; then
        warn "$file:$line_num" "possible magic string $pat — prefer models.Status* constant"
        break
      fi
    done
  done <<< "$matches"
}

# ---------- Check 4: Large files ----------
check_large_files() {
  local file="$1"
  local basename
  basename="$(basename "$file")"

  case "$basename" in
    *_test.go|*_bench_test.go) return ;;
  esac

  local line_count
  line_count=$(wc -l < "$file" | tr -d ' ')

  if [ "$line_count" -gt 1500 ]; then
    warn "$file" "very large file ($line_count lines) — consider splitting by concern"
  fi
}

# ---------- Main ----------
echo "=== AI Navigability Check ==="
echo "Scanning: ${SCAN_DIRS[*]}"
echo ""

cd "$ROOT_DIR"

FILE_COUNT=0
for scan_dir in "${SCAN_DIRS[@]}"; do
  [ -d "$scan_dir" ] || continue
  while IFS= read -r file; do
    [ -z "$file" ] && continue
    [[ "$file" == *vendor/* ]] && continue
    FILE_COUNT=$((FILE_COUNT + 1))

    check_file_comment "$file"
    check_package_doc "$file"
    check_magic_strings "$file"
    check_large_files "$file"
  done < <(find "$scan_dir" -name '*.go' -not -path '*/vendor/*' | sed 's|//|/|g' | sort)
done

report_missing_package_docs

echo ""
echo "=== Summary ==="
echo "Files scanned: $FILE_COUNT"
echo "Errors:   $ISSUES"
echo "Warnings: $WARNINGS"

if [ "$ISSUES" -gt 0 ]; then
  echo ""
  echo "Run with --fix-hint for suggested fixes."
  exit 1
fi

if [ "$WARNINGS" -gt 0 ]; then
  echo "(Warnings are advisory; exit code 0)"
fi

exit 0
