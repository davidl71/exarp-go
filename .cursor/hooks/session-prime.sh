#!/usr/bin/env bash
# sessionStart hook: run exarp-go session prime and inject result as additional_context.
# Cursor calls this when a new composer conversation is created; we add task/handoff context.
# Input (stdin): JSON with session_id, is_background_agent, composer_mode.
# Output (stdout): JSON with additional_context (string), optional continue (bool).

set -e
# Consume stdin (Cursor sends sessionStart payload)
INPUT=$(cat 2>/dev/null || true)

# Run from project root (Cursor runs project hooks with cwd = project root)
EXARP_GO="${EXARP_GO:-}"
if [[ -z "$EXARP_GO" ]]; then
  if [[ -x "./bin/exarp-go" ]]; then
    EXARP_GO="./bin/exarp-go"
  else
    EXARP_GO="exarp-go"
  fi
fi

PRIME_JSON=""
if command -v "$EXARP_GO" >/dev/null 2>&1; then
  RAW=$("$EXARP_GO" -tool session -args '{"action":"prime","include_tasks":true,"include_hints":true,"compact":true,"include_cli_command":false}' -json -quiet 2>/dev/null) || true
  if [[ -n "$RAW" ]]; then
    # CLI returns [{"type":"text","text":"<prime JSON>"}]; extract inner JSON
    if command -v jq >/dev/null 2>&1; then
      PRIME_JSON=$(printf '%s' "$RAW" | jq -r '.[0].text // empty' 2>/dev/null) || true
    fi
  fi
fi

# Build additional_context string from prime result
ADDITIONAL=""
if [[ -n "$PRIME_JSON" ]] && command -v jq >/dev/null 2>&1; then
  STATUS_CTX=$(printf '%s' "$PRIME_JSON" | jq -r '.status_context // ""')
  STATUS_LABEL=$(printf '%s' "$PRIME_JSON" | jq -r '.status_label // ""')
  SUGGESTED=$(printf '%s' "$PRIME_JSON" | jq -r '.suggested_next[0].content // ""')
  CLI_SUG=$(printf '%s' "$PRIME_JSON" | jq -r '.cursor_cli_suggestion // ""')
  HAS_HANDOFF=$(printf '%s' "$PRIME_JSON" | jq -r 'if .handoff_alert != null then "1" else "" end')
  LINES=()
  [[ -n "$STATUS_LABEL" ]] && LINES+=("Session: $STATUS_LABEL")
  [[ -n "$STATUS_CTX" ]] && LINES+=("Context: $STATUS_CTX")
  [[ -n "$HAS_HANDOFF" ]] && LINES+=("Review handoff from previous developer before starting.")
  [[ -n "$SUGGESTED" ]] && LINES+=("Suggested next: $SUGGESTED")
  [[ -n "$CLI_SUG" ]] && LINES+=("CLI: $CLI_SUG")
  ADDITIONAL=$(IFS=$'\n'; echo "${LINES[*]}")
fi

if [[ -z "$ADDITIONAL" ]]; then
  ADDITIONAL="exarp-go session prime unavailable (install exarp-go and ensure jq in PATH for rich context)."
fi

# Cursor sessionStart output schema (exit 0 so session is created)
if command -v jq >/dev/null 2>&1; then
  jq -n --arg ctx "$ADDITIONAL" '{ "additional_context": $ctx, "continue": true }'
else
  # Minimal JSON if jq not available (escape quotes in ADDITIONAL)
  ESCAPED=$(printf '%s' "$ADDITIONAL" | sed 's/\\/\\\\/g; s/"/\\"/g; s/\n/\\n/g')
  printf '{"additional_context":"%s","continue":true}\n' "$ESCAPED"
fi
