# Docs and Code Index: Purpose and Agent Impact

**Purpose:** Define what we mean by a "docs/code index," document current state, and evaluate whether introducing or improving an index helps Cursor and other AI agents.

---

## What Is an Index Here?

- **Docs index:** A single, curated list of documentation files (or sections) with short descriptions and optional categories, so humans and agents can find the right doc without scanning the tree. Example: `docs/README.md` in this repo.
- **Code index:** A short, high-level map of where important behavior lives (packages, key files, entry points). Not a full API reference—that’s `go doc` and the codebase. Example: a "Code map" section in CLAUDE.md or a dedicated one-pager.

---

## Current State

### Docs

- **`docs/README.md`** is the official "Documentation Index." It lists active docs by category (Architecture, Cursor & AI, Active Workflows, etc.), points to the archive, and states doc standards. It is intended to be updated when adding major sections.
- **Gaps:** The index can drift (many files in `docs/` are not listed; new docs are added without always updating the index). There is no machine-readable index (e.g. JSON) for agents.

### Code

- **No formal code index.** Agents and humans rely on:
  - **CLAUDE.md** – project summary, key patterns, where to add things (e.g. task_workflow: `task_workflow_native.go`, `task_workflow_common.go`, `registry.go`).
  - **.cursorrules** – skills table (paths to SKILL.md), plan paths, task/build rules.
  - **.cursor/rules/*.mdc** – detailed rules (Makefile, Todo2, agent locking, LLM tools, etc.) that often reference specific files or packages.
  - **Skills** – e.g. use-exarp-tools points to MCP resources (`stdio://tools`, `stdio://models`) and doc references.
- **Discovery:** Agents use semantic search, grep, and the above hints; there is no single "code map" file.

---

## Would an Index Help Cursor or Other Agents?

### Short answer

- **Docs index: yes, if kept up to date and easy to load.** A single, current list of docs with one-line descriptions helps agents choose which doc to open (or which resource to fetch) without scanning hundreds of files. It reduces noise and wrong-file opens.
- **Code index: marginally.** A very short "code map" (entry points, main packages, where to add tools/rules) can help onboarding and reduce one or two searches per session. It does not replace semantic search or rules; it complements them.

### Docs index – why it helps

| Benefit | Explanation |
|--------|-------------|
| **Fewer irrelevant reads** | Agents get a curated list instead of every `.md` under `docs/` (and archive). |
| **Faster "where is X documented?"** | Categories and descriptions make it easier to pick e.g. "Cursor & AI" or "Build/CI" and open the right file. |
| **Stable entry point** | One canonical place to look; rules/skills can say "see docs/README.md for doc index." |
| **Consistency** | If the index is part of the workflow ("add new major sections here"), it stays more accurate over time. |

**Caveats:** The index must be maintained. An outdated index (wrong or missing files) can misdirect agents. Prefer a short, categorized list over an exhaustive one if maintenance is limited.

### Code index – why the gain is smaller

| Benefit | Explanation |
|--------|-------------|
| **Onboarding / first touch** | A one-page "code map" (cmd vs internal, main tools, where handlers live) can save a few searches for new agents or contributors. |
| **Anchor for rules** | Rules like "add task_workflow actions in X and Y" can point to the code index for the full picture. |

| Limitation | Explanation |
|------------|-------------|
| **Go is already navigable** | Package structure, `go doc`, and grep give agents strong signals; Cursor also does semantic search. |
| **Details change** | File paths and "main" files change; a long code index will drift unless maintained. |
| **Rules already encode the important bits** | CLAUDE.md and .cursor/rules already say where to add tools, where the task store lives, etc. A code index would duplicate some of that. |

So: a **short** code map (e.g. one section in CLAUDE.md or a small CODE_MAP.md) can help a bit; a long, file-by-file index is likely not worth the maintenance for agent use.

---

## Recommendations

1. **Docs index**
   - **Keep and treat as the canonical doc index:** `docs/README.md`.
   - **Update it** when adding or retiring major docs (or at least when doing doc cleanup).
   - **Optional:** Add a "Last updated" or "Index scope" line so agents know whether to trust it for "all active docs" or "high-level only."
   - **Optional (later):** A machine-readable slice (e.g. a small JSON or YAML listing path + title + category) could be generated from the same source for tools that want to filter by category; not required for Cursor to benefit.

2. **Code index**
   - **Prefer enhancing CLAUDE.md** with a short "Code map" subsection (entry points, `internal/` layout, where tools/handlers live) over a separate CODE_INDEX.md, so one file stays the main agent entry point.
   - **Do not** maintain a long file-by-file code index for agents; semantic search and rules are more effective and stay accurate as the code changes.

3. **Cursor and other agents**
   - **Reference the docs index in rules or CLAUDE.md:** e.g. "For full doc list and categories, see docs/README.md."
   - **Session prime / tool hints:** Already surface tasks and tools; no need for the index to be in the prime payload unless you add a "suggested docs" feature later.

---

## Summary

| Index type | Introduce / improve? | Rationale |
|------------|----------------------|-----------|
| **Docs** | **Yes – maintain and use** | Single, curated doc list reduces irrelevant reads and speeds "where is X documented?" for Cursor and other agents. Keep it up to date. |
| **Code** | **Optional – short code map** | A brief "code map" (e.g. in CLAUDE.md) can help a bit; a long code index is not worth it. Rely on rules + semantic search for the rest. |

Introducing a **maintained docs index** (which you already have; the main lever is keeping it current and referenced) helps Cursor and other agents. Introducing a **short code map** in the main agent-facing doc is optional and gives a small benefit; a full code index is not recommended.
