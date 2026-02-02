---
name: lint-docs
description: Lint markdown and check for broken references/links. Use when the user asks to check broken references, validate doc links, lint docs, or run markdown link check.
---

# Lint Docs and Broken References

Apply this skill when the user wants to **check for broken references**, validate markdown links, or lint documentation.

## Built-in behavior

exarp-go **already includes** markdown link checking:

- **Linter:** gomarklint (native Go) with **link check enabled** via `.gomarklint.json` (`enableLinkCheck: true`).
- **How to run:** Use the exarp-go **lint** tool (or `make lint`) and target markdown paths. Link checking runs automatically when markdown is linted.

## When to use

- User asks to "check for broken references", "validate doc links", "check markdown links", or "lint docs".
- Before doc-heavy PRs or when fixing documentation.
- When you suspect dead links in README or `docs/`.

## How to run

| Method | Command / tool |
|--------|-----------------|
| **exarp-go lint tool** | `lint` with `path` set to `docs` or a specific `.md` file; use `linter=markdownlint` (or `auto`) so markdown is linted. Link check is included via gomarklint config. |
| **Makefile** | `make lint` (lints project; markdown under project root uses gomarklint with link check). |
| **CLI** | `exarp-go -tool lint -args '{"path":"docs","linter":"markdownlint"}'` or equivalent. |

## Config

- **`.gomarklint.json`** – `enableLinkCheck: true` turns on broken link detection for all gomarklint runs.
- No extra flags needed: the lint tool invokes gomarklint, which reads this config.

## Summary

**Broken reference checking is built-in.** Prefer the **lint** tool (or `make lint`) with docs/markdown paths; no separate "link check" tool is required. Point the user to this when they ask for broken reference or link validation.
