# Suggested Cursor heredocs

Heredoc patterns useful when working in Cursor’s terminal or in scripts: creating project files, embedding multi-line JSON for tools, and keeping context tidy.

## 1. Creating / overwriting project files (terminal)

Use a quoted delimiter so the shell doesn’t expand `$VAR` or backticks.

```bash
# Create or overwrite .cursorignore (from repo root)
cd "$(make r)"
cat > .cursorignore << 'CURSORIGNORE_EOF'
vendor/
bin/
build/
.todo2/
.exarp/
CURSORIGNORE_EOF
```

```bash
# Append a single block to an existing file
cat >> .cursorignore << 'EOF'
# My custom pattern
my-dir/
EOF
```

## 2. Multi-line JSON for exarp-go tools

For `exarp-go -tool … -args '...'`, use heredocs to keep JSON readable and avoid quoting issues.

```bash
# Session prime (compact)
./bin/exarp-go -tool session -args "$(cat << 'JSON'
{"action":"prime","include_hints":true,"include_tasks":true,"compact":true}
JSON
)"

# task_workflow list with filters (multi-line)
./bin/exarp-go -tool task_workflow -args "$(cat << 'JSON'
{
  "action": "sync",
  "sub_action": "list",
  "status": "Todo",
  "output_format": "json",
  "compact": true
}
JSON
)"
```

In scripts, you can use a variable:

```bash
PRIME_ARGS=$(cat << 'JSON'
{"action":"prime","include_tasks":true,"compact":true}
JSON
)
./bin/exarp-go -tool session -args "$PRIME_ARGS"
```

## 3. Cursor Composer “context” (not literal heredocs)

In Composer, use `@` to attach context instead of pasting huge blocks:

- `@Folder` – e.g. `@internal/tools` for the tools package.
- `@Code` – select a symbol or range.
- `@Docs` – pull in docs.
- `@Web` – reference a URL.

Combining `@Files` / `@Folders` with a short prompt often works better than one long heredoc-style paste.

## 4. Makefile / script snippets

Embedding small config or JSON in Makefile targets:

```makefile
# Example: prime and show output (POSIX shell)
prime:
	./bin/exarp-go -tool session -args '{"action":"prime","compact":true}'
```

For multi-line JSON in Make, use a single line (no real heredoc) or a shell heredoc inside the recipe:

```makefile
prime-verbose:
	./bin/exarp-go -tool session -args "$$(cat << 'JSON'\
{\"action\":\"prime\",\"include_tasks\":true,\"compact\":true}\
JSON)"
```

## Quick reference

| Goal | Pattern |
|------|--------|
| Create file, no expansion | `cat > file << 'EOF'` … `EOF` |
| Append block | `cat >> file << 'EOF'` … `EOF` |
| JSON for -tool -args | `-args "$(cat << 'JSON'` … `JSON")"` |
| Composer context | Prefer `@Files` / `@Folders` / `@Code` over huge pastes |
