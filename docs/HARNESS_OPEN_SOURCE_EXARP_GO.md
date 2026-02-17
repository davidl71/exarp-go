# Harness Open Source with exarp-go

**Purpose:** Use Harness Open Source as CI/CD and dev platform alongside exarp-go, and how it can improve the development process.  
**Last updated:** 2026-02-17

---

## 1. Overview

**Harness Open Source** is a source-available CI/CD and software delivery platform (Apache 2.0) that provides:

- Source/SCM (repos, PRs, code review, secret scanning)
- CI/CD pipelines (YAML-based; Go, Node, Rust, etc.)
- Cloud dev environments (Gitspaces)
- Artifact management (Docker, Helm)
- 100+ integrations (e.g. GitHub, GitLab, Bitbucket)

exarp-go already runs in **GitHub Actions** (agentic-ci, go.yml) and can be invoked from **any** CI that can run a binary and Makefile targets. Harness fits the same pattern: run `./bin/exarp-go -tool ...` and `make` in pipeline steps.

---

## 2. Using Harness with exarp-go

### 2.1 Same pattern as existing CI

exarp-go is CI-agnostic. Current usage (see [CLI_MAKE_CI_USAGE.md](CLI_MAKE_CI_USAGE.md) and [AGENTIC_CI_SETUP.md](AGENTIC_CI_SETUP.md)):

- Build: `make build` (or `go build -o bin/exarp-go ./cmd/server`)
- Agent validation: `./bin/exarp-go -test tool_workflow`, `-test automation`, `-tool testing -args '{"action":"validate"}'`
- Todo2: `./bin/exarp-go -tool task_workflow -args '{"action":"sync"}'`
- Report: `./bin/exarp-go -tool report -args '{"action":"overview"}'` or `"scorecard"`
- Task sanity: `make task-sanity-check`

In Harness you run these same commands inside pipeline steps (e.g. a “Run” or “Script” step with Go environment).

### 2.2 Example Harness pipeline (conceptual)

Harness pipelines are YAML-based. A minimal pipeline that mirrors agentic CI could look like:

```yaml
# Conceptual Harness pipeline - run exarp-go validation
steps:
  - name: checkout
    type: clone
    spec:
      connector: your-git-connector
      repo: exarp-go

  - name: setup-go
    type: run
    spec:
      image: golang:1.25-bookworm
      shell: sh
      command: |
        go version
        go mod download

  - name: build-exarp-go
    type: run
    spec:
      image: golang:1.25-bookworm
      shell: sh
      command: |
        go build -o bin/exarp-go ./cmd/server

  - name: agent-validation
    type: run
    spec:
      image: golang:1.25-bookworm
      shell: sh
      command: |
        ./bin/exarp-go -test tool_workflow
        ./bin/exarp-go -test automation
        ./bin/exarp-go -tool task_workflow -args '{"action":"sync"}'
        ./bin/exarp-go -tool report -args '{"action":"overview","include_metrics":true}'
        make task-sanity-check
```

(Exact step types and syntax depend on Harness Open Source’s current pipeline DSL; the important part is running the same exarp-go and Make commands.)

### 2.3 Where exarp-go runs

- **In Harness:** As pipeline steps (build + validation + report + task sanity), same as in [agentic-ci](.github/workflows/agentic-ci.yml) and [JENKINS_INTEGRATION.md](JENKINS_INTEGRATION.md).
- **Locally / in Cursor:** Via MCP (exarp-go as MCP server) and CLI (`exarp-go -tool ...`, `make ...`).
- **Sprint/backlog automation:** exarp-go’s `automation` tool (daily/nightly/sprint) can be triggered by Harness scheduled pipelines or by a separate scheduler; no change to exarp-go’s behavior.

---

## 3. How Harness can improve the development process

### 3.1 Single platform (optional consolidation)

- **Today:** GitHub (repo + Actions) + local Make + exarp-go MCP/CLI.
- **With Harness:** You can centralize SCM, CI, and (optionally) artifacts in one place. exarp-go remains the “agentic” layer (tasks, report, validation, automation); Harness runs exarp-go inside its pipelines.

Improvement: One place to see builds, pipelines, and (if you use it) repos and PRs, while keeping exarp-go as the source of truth for task workflow and project health.

### 3.2 Consistent dev environments (Gitspaces)

- **Problem:** “Works on my machine” — different Go versions, CGO, or missing tools across developers.
- **Harness:** Gitspaces (or equivalent) can provide a preconfigured dev environment (Go 1.25, make, optional CGO) so everyone runs the same stack.

Improvement: Fewer env-related failures; same environment for local dev and CI.

### 3.3 Stronger CI without leaving exarp-go

- Reuse the **same** agent validation and task sanity that agentic CI runs (see [TASKS_AS_CI_AND_AUTOMATION.md](TASKS_AS_CI_AND_AUTOMATION.md)).
- Add Harness-native steps around exarp-go: security scanning, container builds, deployment — while exarp-go continues to own task workflow, report, and automation logic.

Improvement: Richer CI (containers, deployments) without duplicating exarp-go’s validation logic.

### 3.4 Scheduled and event-driven automation

- **exarp-go today:** `automation` tool (daily/nightly/sprint); `make sprint-start`, `pre-sprint`, `check-tasks`, `update-completed-tasks`.
- **With Harness:** Harness can run pipelines on a schedule (e.g. nightly) or on events. Each pipeline can call exarp-go (e.g. `./bin/exarp-go -tool automation -args '{"action":"nightly"}'` or run `make pre-sprint`).

Improvement: Clear audit trail and retries in Harness, while exarp-go still does the actual task/backlog work.

### 3.5 Migration and flexibility

- Harness supports migrations from GitHub, GitLab, Bitbucket, CircleCI, etc. If you ever move SCM or want to try Harness without dropping GitHub Actions, you can run both (e.g. Harness for a subset of pipelines) and keep agentic-ci as the reference for “what exarp-go must run.”

Improvement: Option to gradually adopt Harness or use it for specific pipelines (e.g. release, deploy) while keeping GitHub Actions for PR checks.

---

## 4. What stays in exarp-go

Regardless of CI platform:

- **Task workflow and Todo2** — exarp-go (and MCP) remain the interface; CI only invokes `task_workflow`, `task_analysis`, etc.
- **Report and scorecard** — exarp-go generates them; CI runs the binary and archives/fails on result.
- **Automation semantics** — daily/nightly/sprint logic lives in exarp-go; Harness (or any scheduler) only triggers the binary.
- **Agent validation** — the set of `-test` and `-tool` calls defined in agentic-ci and [JENKINS_INTEGRATION.md](JENKINS_INTEGRATION.md) stay the same; Harness just runs them.

---

## 5. Practical next steps

1. **Try Harness Open Source** (e.g. [harness.io/open-source](https://harness.io/open-source), [developer.harness.io/docs/open-source](https://developer.harness.io/docs/open-source)).
2. **Add one pipeline** that: checkout → `go build` → run the same exarp-go validation + report + task-sanity steps as agentic-ci.
3. **Keep GitHub Actions** as the primary PR gate until Harness is proven; then you can duplicate or replace agentic-ci in Harness.
4. **Optionally** use Gitspaces for a standard dev environment and document it in the repo (e.g. in README or CONTRIBUTING).

---

## 6. References

- [AGENTIC_CI_SETUP.md](AGENTIC_CI_SETUP.md) — Current agentic CI (build + validation + report).
- [CLI_MAKE_CI_USAGE.md](CLI_MAKE_CI_USAGE.md) — Invoking exarp-go from Make/CI.
- [TASKS_AS_CI_AND_AUTOMATION.md](TASKS_AS_CI_AND_AUTOMATION.md) — What is already CI vs exarp automation.
- [JENKINS_INTEGRATION.md](JENKINS_INTEGRATION.md) — Same “run exarp-go in pipeline” pattern for Jenkins.
- Harness: [harness.io/open-source](https://harness.io/open-source), [developer.harness.io/docs/open-source](https://developer.harness.io/docs/open-source), [GitHub](https://github.com/harness).
