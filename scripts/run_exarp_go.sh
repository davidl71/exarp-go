#!/usr/bin/env bash
# Portable exarp-go runner for use by Cursor, OpenCode, and other MCP clients.
#
# Copy this script into your project (e.g. scripts/run_exarp_go.sh) and set your
# MCP server "command" to it. It sets PROJECT_ROOT to your project and resolves
# the exarp-go binary (sibling repo, PATH, or EXARP_GO_ROOT).
#
# Resolution order:
#   1. EXARP_GO_ROOT/bin/exarp-go (if set and executable)
#   2. Walk up from CWD for exarp-go repo (go.mod + cmd or bin/exarp-go); use bin or "go run ."
#   3. exarp-go on PATH
#   4. PROJECT_ROOT/../exarp-go/bin/exarp-go
#   5. ~/go/bin/exarp-go, ~/Projects/exarp-go/bin/exarp-go, /usr/local/bin/exarp-go
#
# Env: PROJECT_ROOT (project exarp-go serves); EXARP_GO_ROOT (exarp-go repo);
#      EXARP_GO_VERBOSE=1 to log which binary is used.
# When PROJECT_ROOT is unset: if the script lives inside a project (e.g. project/scripts/), that project is used;
# otherwise (e.g. script in GOPATH/bin) the current working directory is used.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# Project root: prefer env; else derive from script path if script lives inside a project; else use CWD (e.g. global install).
if [[ -n "${PROJECT_ROOT:-}" ]] && [[ "${PROJECT_ROOT}" != "{{PROJECT_ROOT}}" ]]; then
  : # use PROJECT_ROOT as set by Cursor/OpenCode or caller
elif [[ -d "${SCRIPT_DIR}/../.todo2" ]] || ( [[ -f "${SCRIPT_DIR}/../go.mod" ]] && ! grep -q 'exarp-go' "${SCRIPT_DIR}/../go.mod" 2>/dev/null ); then
  # Script is inside a project (e.g. project/scripts/run_exarp_go.sh)
  PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
else
  # Global install (e.g. GOPATH/bin) or no project around script: use current directory
  PROJECT_ROOT="$(pwd)"
fi
export PROJECT_ROOT

cd "${PROJECT_ROOT}"

is_exarp_go_repo() {
  local dir="${1:-.}"
  [[ -f "${dir}/go.mod" ]] || return 1
  grep -q 'exarp-go' "${dir}/go.mod" 2>/dev/null || return 1
  [[ -f "${dir}/cmd/server/main.go" ]] || [[ -x "${dir}/bin/exarp-go" ]] || return 1
  return 0
}

find_exarp_go_root() {
  local start_dir="${1:-$(pwd)}"
  local d
  d="$(cd "${start_dir}" 2>/dev/null && pwd)" || return 1
  while [[ -n "${d}" ]] && [[ "${d}" != "/" ]]; do
    if is_exarp_go_repo "${d}"; then
      echo "${d}"
      return 0
    fi
    d="$(dirname "${d}")"
  done
  return 1
}

resolve_exarp_go_bin() {
  local cwd_root fallback_root

  if [[ -n "${EXARP_GO_ROOT:-}" ]] && is_exarp_go_repo "${EXARP_GO_ROOT}"; then
    if [[ -x "${EXARP_GO_ROOT}/bin/exarp-go" ]]; then
      EXARP_GO_BIN="${EXARP_GO_ROOT}/bin/exarp-go"
      [[ -n "${EXARP_GO_VERBOSE:-}" ]] && echo "[exarp-go] using EXARP_GO_ROOT bin: ${EXARP_GO_BIN}" >&2
      return 0
    fi
    if command -v go &>/dev/null; then
      EXARP_GO_RUN_GO=1
      export EXARP_GO_ROOT
      [[ -n "${EXARP_GO_VERBOSE:-}" ]] && echo "[exarp-go] using EXARP_GO_ROOT with go run: ${EXARP_GO_ROOT}" >&2
      return 0
    fi
  fi

  cwd_root="$(find_exarp_go_root "$(pwd)" 2>/dev/null)" || true
  if [[ -n "${cwd_root}" ]]; then
    if [[ -x "${cwd_root}/bin/exarp-go" ]]; then
      EXARP_GO_BIN="${cwd_root}/bin/exarp-go"
      export EXARP_GO_ROOT="${cwd_root}"
      [[ -n "${EXARP_GO_VERBOSE:-}" ]] && echo "[exarp-go] using CWD exarp-go repo bin: ${EXARP_GO_BIN}" >&2
      return 0
    fi
    if command -v go &>/dev/null; then
      export EXARP_GO_ROOT="${cwd_root}"
      EXARP_GO_RUN_GO=1
      [[ -n "${EXARP_GO_VERBOSE:-}" ]] && echo "[exarp-go] using CWD exarp-go repo with go run: ${EXARP_GO_ROOT}" >&2
      return 0
    fi
  fi

  if command -v exarp-go &>/dev/null; then
    EXARP_GO_BIN="exarp-go"
    [[ -n "${EXARP_GO_VERBOSE:-}" ]] && echo "[exarp-go] using PATH exarp-go" >&2
    return 0
  fi

  fallback_root=""
  if [[ -n "${EXARP_GO_ROOT:-}" ]] && [[ -x "${EXARP_GO_ROOT}/bin/exarp-go" ]]; then
    fallback_root="${EXARP_GO_ROOT}"
  elif [[ -x "${PROJECT_ROOT}/../exarp-go/bin/exarp-go" ]]; then
    fallback_root="$(cd "${PROJECT_ROOT}/../exarp-go" && pwd)"
  fi
  if [[ -n "${fallback_root}" ]]; then
    EXARP_GO_BIN="${fallback_root}/bin/exarp-go"
    export EXARP_GO_ROOT="${fallback_root}"
    [[ -n "${EXARP_GO_VERBOSE:-}" ]] && echo "[exarp-go] using fallback working dir: ${EXARP_GO_BIN}" >&2
    return 0
  fi

  for candidate in \
    "${HOME}/go/bin/exarp-go" \
    "${HOME}/Projects/exarp-go/bin/exarp-go" \
    "${HOME}/Projects/mcp/exarp-go/bin/exarp-go" \
    "/usr/local/bin/exarp-go"; do
    if [[ -x "${candidate}" ]]; then
      EXARP_GO_BIN="${candidate}"
      [[ -n "${EXARP_GO_VERBOSE:-}" ]] && echo "[exarp-go] using fallback path: ${EXARP_GO_BIN}" >&2
      return 0
    fi
  done

  return 1
}

EXARP_GO_BIN=""
EXARP_GO_RUN_GO=""
if ! resolve_exarp_go_bin; then
  echo "exarp-go not found. Install it, set PATH, or set EXARP_GO_ROOT to repo (with bin/exarp-go built)." >&2
  exit 1
fi

if [[ -n "${EXARP_GO_ROOT:-}" ]]; then
  export EXARP_MIGRATIONS_DIR="${EXARP_MIGRATIONS_DIR:-${EXARP_GO_ROOT}/migrations}"
fi

if [[ -n "${EXARP_GO_RUN_GO:-}" ]]; then
  cd "${EXARP_GO_ROOT}"
  exec go run ./cmd/server "$@"
fi

exec "${EXARP_GO_BIN}" "$@"
