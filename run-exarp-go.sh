#!/bin/bash
# Cursor MCP wrapper: delegates to start.sh (one script to rule them all).
# Preserves PROJECT_ROOT, EXARP_WATCH, EXARP_DEBUG_LOG.
# When PROJECT_ROOT is unset or the literal placeholder {{PROJECT_ROOT}}, use script dir
# so the server starts and migrations are found (EXARP_MIGRATIONS_DIR).
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -z "${PROJECT_ROOT}" || "${PROJECT_ROOT}" == "{{PROJECT_ROOT}}" ]]; then
  export PROJECT_ROOT="$SCRIPT_DIR"
fi
# Ensure migrations run when PROJECT_ROOT is another project (no migrations/ there)
export EXARP_MIGRATIONS_DIR="${EXARP_MIGRATIONS_DIR:-$SCRIPT_DIR/migrations}"
exec "$SCRIPT_DIR/start.sh" "$@"
