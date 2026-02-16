# Jenkins Integration with exarp-go

This document describes integration options between Jenkins and exarp-go (agentic CI, report/scorecard, task workflow, automation). Use it to add quality gates, agent validation, and task-aware builds to Jenkins pipelines.

## Overview

exarp-go provides:

- **Report & scorecard** — project health overview and metrics (JSON/text).
- **Agentic validation** — same checks as [Agentic Development CI](.github/workflows/agentic-ci.yml): tool workflow, automation, testing, task_workflow sync.
- **Task workflow** — Todo2 list/update/sync; can drive “what to build” or “what failed.”
- **Automation** — daily/nightly/sprint workflows.

Jenkins can run exarp-go as a build step, consume its output for gates, and optionally trigger jobs from exarp-go (webhook or CLI).

---

## 1. Report/Scorecard as a Pipeline Step

**Idea:** Run report (overview or scorecard) in a pipeline stage and archive or gate on the result.

**Steps:**

1. Build exarp-go (or use a pre-built image/binary).
2. In a pipeline stage, run:
   - `./bin/exarp-go -tool report -args '{"action":"overview","include_metrics":true}'`
   - `./bin/exarp-go -tool report -args '{"action":"scorecard"}'`
3. Archive the JSON/text output as a build artifact.
4. Optionally fail the build if a threshold is not met (e.g. parse JSON and check `overall_score` or critical checks).

**Use case:** Every build gets a project health snapshot; no code changes required.

---

## 2. Agentic Validation in Jenkins (Mirror agentic-ci)

**Idea:** Replicate the [agentic-ci](.github/workflows/agentic-ci.yml) workflow inside Jenkins.

**Steps:**

1. Build exarp-go (one stage).
2. Run the same validation commands used in GitHub Actions:
   - `./bin/exarp-go -test tool_workflow`
   - `./bin/exarp-go -test automation`
   - `./bin/exarp-go -tool testing -args '{"action":"validate",...}'`
   - `./bin/exarp-go -tool task_workflow -args '{"action":"sync"}'`
   - `./bin/exarp-go -tool report -args '{"action":"overview",...}'`
3. Capture stdout/stderr and report output; archive as “Agent Validation Report.”
4. Fail the job if any validation step fails.

**Use case:** Run agentic validation in Jenkins instead of (or in addition to) GitHub Actions.

---

## 3. Pipeline Triggered by exarp-go (Webhook or CLI)

**Idea:** Jenkins runs a job when exarp-go (or another service) requests it.

**Options:**

- **Webhook:** Jenkins job triggered by HTTP POST (e.g. from a small HTTP endpoint or script that exarp-go or another tool calls after certain events).
- **CLI from Jenkins:** Jenkins runs on a schedule or SCM; the pipeline calls exarp-go (e.g. `automation`, `task_workflow`) and uses the output to decide which job or parameters to run (e.g. suggested next tasks).

**Use case:** Event-driven or logic-driven CI from exarp-go.

---

## 4. Scorecard/Report as a Quality Gate

**Idea:** Use report/scorecard as a promotion or deploy gate.

**Steps:**

1. In a promotion/deploy pipeline, add a stage that runs `report` (overview or scorecard).
2. Parse the JSON (or a thin wrapper script) and enforce a rule (e.g. `overall_score >= 70`, or “no critical failures”).
3. Pass/fail the stage based on that; block deploy if the gate fails.

**Use case:** Deploy only when project health is above a defined threshold.

---

## 5. Jenkinsfile / Pipeline Generator from exarp-go

**Idea:** exarp-go (or a script) emits a suggested Jenkinsfile or pipeline stages from project context.

**Steps:**

1. Add a tool or script that, given project context (from `report`, `task_workflow`, or `automation`), outputs a suggested Jenkinsfile or pipeline snippet.
2. Snippet includes: build (e.g. `make build`), test (e.g. `make test` + optional `-tool testing`), report/scorecard, optional agent validation.
3. Users copy the snippet into their Jenkins job.

**Use case:** One-click-style pipeline aligned with agentic CI and report.

---

## 6. Daily/Weekly Report or Briefing Job

**Idea:** Scheduled Jenkins job that runs report (and optionally DevWisdom briefing) and publishes the result.

**Steps:**

1. Create a Jenkins job (e.g. daily or weekly).
2. Run exarp-go report (overview/briefing) and optionally DevWisdom daily briefing if available.
3. Publish output to Jenkins run description, workspace file, and/or email/Slack/Confluence.

**Use case:** Recurring project health and “wisdom” digest for the team.

---

## 7. Task-Aware Builds (Todo2 + Jenkins)

**Idea:** Link Jenkins builds to Todo2 tasks.

**Steps:**

1. Pipeline (or a wrapper) calls exarp-go to list “in progress” or “suggested next” tasks (e.g. `task_workflow` list).
2. Use that to:
   - Tag/label the build (e.g. “T-123”, feature name).
   - Notify or assign on failure (e.g. “Build for T-123 failed”).
   - Or drive a parameterized job: “Build and test for task T-XXX” with options from exarp-go.

**Use case:** Traceability from builds to tasks and clearer ownership.

---

## 8. Shared exarp-go Binary in Jenkins

**Idea:** One job builds exarp-go and publishes the binary; other jobs consume it.

**Steps:**

1. Dedicated job (or stage) builds exarp-go (e.g. `make build` or `make build-apple-fm` if needed).
2. Publish the binary to an artifact repository or shared path.
3. Downstream jobs download that binary and run `exarp-go -tool ...` so all pipelines use the same agent version.

**Use case:** Consistent agent version across many Jenkins jobs.

---

## Quick Wins (Recommended First)

| Priority | Item | Description |
|----------|------|-------------|
| 1 | Scorecard stage | Single pipeline stage: run `report` scorecard, archive result, optionally fail on low score. |
| 2 | Agentic validation pipeline | Replicate agentic-ci in Jenkins: build + validation + report. |
| 3 | Scheduled report job | Scheduled job: run report/overview (and optional briefing), email or post summary. |

---

## Tasks (Todo2)

| Task ID | Summary |
|---------|---------|
| T-1771253063737 | Jenkins: Add pipeline stage for exarp-go scorecard (quick win #1) |
| T-1771253067039 | Jenkins: Replicate agentic-ci pipeline (quick win #2) |
| T-1771253068151 | Jenkins: Scheduled job for report/briefing (quick win #3) |
| T-1771253070385 | Jenkins: Scorecard as quality gate for deploy |
| T-1771253072479 | Jenkins: Task-aware builds (Todo2 + Jenkins) |

---

## References

- [Agentic CI Setup](AGENTIC_CI_SETUP.md) — GitHub Actions workflow and usage.
- [.github/workflows/agentic-ci.yml](../.github/workflows/agentic-ci.yml) — Reference workflow to mirror in Jenkins.
- [.cursor/rules/agentic-ci.mdc](../.cursor/rules/agentic-ci.mdc) — When to use agentic CI.
