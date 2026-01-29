# Agentic CI Setup Guide for exarp-go

**Date:** 2026-01-09  
**Purpose:** Guide for setting up agentic development CI with GitHub Actions and exarp-go  
**Status:** ✅ Setup Complete

**Note (2026-01-29):** The AgentScope evaluation job and `bridge/agent_evaluation.py` were removed from agentic CI. CI now runs build → agent-validation (exarp-go) → combined-validation only.

---

## Overview

This guide explains how to set up agentic development CI for exarp-go using:
1. **GitHub Actions** - Primary CI/CD platform (already configured)
2. **exarp-go Tools** - Native agent validation (immediate, no dependencies)
3. **AgentScope** - Removed from CI 2026-01-29; optional local use only if reinstated

---

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    GitHub Actions CI                      │
├─────────────────────────────────────────────────────────┤
│                                                           │
│  ┌──────────────────┐      ┌──────────────────────┐    │
│  │  Build exarp-go  │ ────▶│ Agent Validation     │    │
│  │   (Required)     │      │ (exarp-go tools)     │    │
│  └──────────────────┘      └──────────┬───────────┘    │
│         │                               │                 │
│         └───────────────────────────────┼─────────────────┘
│                                         ▼                 │
│                          ┌──────────────────────┐       │
│                          │ Combined Report      │       │
│                          └──────────────────────┘       │
└─────────────────────────────────────────────────────────┘
```
(AgentScope job removed 2026-01-29.)

---

## Quick Start

### 1. Basic Setup (No Dependencies)

The workflow is already configured and will work immediately:

```bash
# Just push to main/develop or create a PR
git push origin main
```

**What runs:**
- ✅ Builds exarp-go binary
- ✅ Validates MCP tools
- ✅ Tests agent behavior
- ✅ Checks Todo2 integration
- ✅ Generates agent report

### 2. Enhanced Setup (With AgentScope)

To enable AgentScope evaluation:

```bash
# Install AgentScope (optional)
pip install agentscope

# The workflow will automatically use AgentScope if available
# No additional configuration needed
```

---

## Workflow Jobs

### Job 1: Build exarp-go

**Purpose:** Build the exarp-go binary for other jobs

**Runs on:** `ubuntu-latest` (GitHub-hosted)

**Steps:**
1. Checkout code
2. Set up Go 1.25
3. Cache Go modules
4. Build binary
5. Upload binary artifact

**Output:** `exarp-go-binary` artifact

---

### Job 2: Agent Validation (exarp-go)

**Purpose:** Validate agent behavior using exarp-go's native tools

**Runs on:** `self-hosted` (falls back to `ubuntu-latest`)

**Dependencies:** Requires `build` job

**Steps:**
1. Download exarp-go binary
2. Validate MCP tools (`tool_workflow`, `automation`, `testing`)
3. Test agent behavior using `testing` tool
4. Check Todo2 integration using `task_workflow` tool
5. Generate agent report using `report` tool
6. Upload validation report

**Output:** `agent-validation-report` artifact

**Key Features:**
- ✅ Uses existing exarp-go tools (no new dependencies)
- ✅ Validates all 24 MCP tools
- ✅ Tests agent behavior
- ✅ Checks Todo2 integration
- ✅ Generates comprehensive report

---

### Job 3: AgentScope Evaluation (Optional)

**Purpose:** Advanced agent evaluation using AgentScope framework

**Runs on:** `self-hosted` (falls back to `ubuntu-latest`)

**Dependencies:** Requires `build` job

**Condition:** Only runs on pushes (not on external PRs)

**Steps:**
1. Set up Python 3.11
2. Download exarp-go binary
3. Install AgentScope
4. Run AgentScope evaluation
5. Generate evaluation report
6. Upload evaluation results

**Output:** `agentscope-evaluation-results` artifact

**Key Features:**
- ✅ Advanced evaluation framework
- ✅ Configurable test cases
- ✅ Metrics tracking
- ✅ Detailed evaluation reports

**Configuration:** `.github/agentscope_eval.yaml`

---

### Job 4: Combined Validation

**Purpose:** Combine results from both validation approaches

**Runs on:** `self-hosted`

**Dependencies:** Requires both `agent-validation` and `agentscope-evaluation` jobs

**Steps:**
1. Download all validation reports
2. Combine reports into single document
3. Upload combined report

**Output:** `combined-agent-validation-report` artifact

---

## Configuration Files

### `.github/workflows/agentic-ci.yml`

Main GitHub Actions workflow file. Defines all CI jobs.

**Key Features:**
- Builds exarp-go binary
- Runs agent validation
- Optionally runs AgentScope evaluation
- Combines results

---

### `.github/agentscope_eval.yaml`

AgentScope evaluation configuration.

**Configuration Options:**
- `evaluator`: Evaluator settings
- `test_cases`: List of test cases to run
- `metrics`: Evaluation metrics and thresholds
- `output`: Output format and options

**Example Test Cases:**
- MCP tool validation
- Task workflow validation
- Automation workflow validation
- Testing tool validation

---

### ~~`bridge/agent_evaluation.py`~~ (Removed 2026-01-29)

AgentScope evaluation was removed from agentic CI. Use exarp-go validation (agent-validation job) and combined report only. To run AgentScope locally you would need to re-add a script and config; the workflow no longer references it.

---

## Self-Hosted Runners

### Using Self-Hosted Runners

The workflow is configured to use self-hosted runners when available:

```yaml
runs-on: self-hosted  # Falls back to ubuntu-latest if not available
```

**Benefits:**
- ✅ Unlimited CI minutes
- ✅ Faster builds (local network)
- ✅ Access to local dependencies
- ✅ Test on real hardware

**Setup:**
See `ib_box_spread_full_universal/docs/SELF_HOSTED_RUNNER_SETUP.md` for setup instructions.

**Current Runners:**
- Ubuntu agent: `192.168.192.57`
- macOS M4 agent: `192.168.192.141`

---

## Running Locally

### Test Agent Validation

```bash
# Build exarp-go
make build

# Validate MCP tools
./bin/exarp-go -test tool_workflow
./bin/exarp-go -test automation
./bin/exarp-go -test testing

# Test agent behavior
./bin/exarp-go -tool testing \
  -args '{"action":"validate","test_path":"./internal/tools","framework":"go"}'

# Check Todo2 integration
./bin/exarp-go -tool task_workflow \
  -args '{"action":"sync"}'

# Generate report
./bin/exarp-go -tool report \
  -args '{"action":"overview","include_metrics":true}'
```

### ~~Test AgentScope Evaluation~~ (Removed 2026-01-29)

AgentScope job and `bridge/agent_evaluation.py` were removed. Use exarp-go validation steps above; combined report contains exarp-go validation only.

---

## Customization

### Adding Custom Test Cases

Edit `.github/agentscope_eval.yaml`:

```yaml
test_cases:
  - name: "custom_test"
    description: "Your custom test case"
    test_type: "functional"
    expected_behavior:
      - "Expected behavior 1"
      - "Expected behavior 2"
```

### Modifying Validation Steps

Edit `.github/workflows/agentic-ci.yml`:

```yaml
- name: Custom Validation Step
  run: |
    ./bin/exarp-go -tool your_tool \
      -args '{"action":"your_action"}'
```

### Adjusting Metrics Thresholds

Edit `.github/agentscope_eval.yaml`:

```yaml
metrics:
  - name: "tool_success_rate"
    threshold: 0.98  # 98% success rate required
```

---

## Troubleshooting

### AgentScope Not Available

**Error:** `AgentScope not available`

**Solution:**
```bash
pip install agentscope
```

**Note:** The workflow will continue without AgentScope - it's optional.

---

### Self-Hosted Runner Not Available

**Error:** Workflow runs on `ubuntu-latest` instead of `self-hosted`

**Solution:**
- Check runner is online in GitHub Settings → Actions → Runners
- Verify runner labels match workflow requirements
- Workflow will automatically fall back to `ubuntu-latest`

---

### Evaluation Fails

**Error:** AgentScope evaluation fails

**Solution:**
1. Check `.github/agentscope_eval.yaml` syntax
2. Verify AgentScope is installed
3. Check evaluation logs in GitHub Actions
4. Evaluation failures don't block the workflow (uses `|| true`)

---

## Best Practices

### 1. Start Simple

Begin with basic agent validation (Job 2) before adding AgentScope.

### 2. Use Self-Hosted Runners

Set up self-hosted runners for faster builds and unlimited minutes.

### 3. Monitor Metrics

Review evaluation reports regularly to track agent behavior over time.

### 4. Customize Test Cases

Add test cases specific to your agent's use cases.

### 5. Combine Approaches

Use both exarp-go tools and AgentScope for comprehensive validation.

---

## Next Steps

1. ✅ **Workflow is ready** - Push to trigger CI
2. **Optional:** Install AgentScope for advanced evaluation
3. **Optional:** Set up self-hosted runners
4. **Customize:** Add your own test cases and metrics
5. **Monitor:** Review evaluation reports regularly

---

## References

- **Workflow File:** `.github/workflows/agentic-ci.yml`
- **AgentScope Config:** `.github/agentscope_eval.yaml`
- **Bridge Script:** `bridge/agent_evaluation.py`
- **Integration Analysis:** `docs/research/AGENTIC_CI_EXARP_GO_INTEGRATION.md`
- **Tool Research:** `docs/research/AGENTIC_DEVELOPMENT_CI_TOOLS.md`
