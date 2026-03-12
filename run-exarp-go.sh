#!/bin/bash
# Cursor MCP wrapper: delegates to start.sh (one script to rule them all).
# Preserves PROJECT_ROOT, EXARP_WATCH, EXARP_DEBUG_LOG.
# When PROJECT_ROOT is unset or the literal placeholder {{PROJECT_ROOT}}, prefer the
# active workspace and fall back to the script directory so the server still starts.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -z "${PROJECT_ROOT:-}" || "${PROJECT_ROOT}" == "{{PROJECT_ROOT}}" ]]; then
  if [[ -n "${PWD:-}" && -d "${PWD}" && "${PWD}" != "$SCRIPT_DIR" ]]; then
    export PROJECT_ROOT="${PWD}"
  else
    export PROJECT_ROOT="$SCRIPT_DIR"
  fi
fi
# Ensure migrations run when PROJECT_ROOT is another project (no migrations/ there)
export EXARP_MIGRATIONS_DIR="${EXARP_MIGRATIONS_DIR:-$SCRIPT_DIR/migrations}"
exec "$SCRIPT_DIR/start.sh" "$@"
