---
name: report-scorecard
description: Run exarp-go report and scorecard. Use when the user asks for project overview, scorecard, briefing, or status; after significant changes; or before reviews or standups.
---

# exarp-go Report and Scorecard

Apply this skill when you need a structured project overview, scorecard, or briefing from exarp-go.

## Report Tool Actions

Use the exarp-go `report` tool with one of:

| Action | Use when |
|--------|----------|
| **overview** | You need a high-level project summary, metrics, or health snapshot. |
| **scorecard** | You need a scorecard (alignment, completion, docs, testing, security). Use `fast_mode=true` for Go projects when quick feedback is enough. |
| **briefing** | You need a short briefing (e.g. for standups or handoffs). |

## When to Suggest or Run Report

- User asks for “project status”, “overview”, “scorecard”, or “briefing”.
- After major changes (e.g. migration, big refactor) to re-check alignment and quality.
- Before PR review or standup to get a concise project snapshot.
- When debugging “why does the project look inconsistent?” – overview/scorecard can clarify gaps.

## Example Usage

- **Overviews:** `report` with `action=overview`, and optionally `include_metrics=true`, `include_architecture=true`.
- **Scorecards:** `report` with `action=scorecard`; for Go, `fast_mode=true` is often enough.
- **Briefings:** `report` with `action=briefing`.

## Integration with Other exarp-go Tools

- **health** – Use for docs/CI/git checks; use **report** when you need narrative overview, scorecard, or briefing.
- **session** – Use `session` with `action=prime` at conversation start; use **report** when you need a shareable or persistent snapshot (overview/scorecard/briefing).
