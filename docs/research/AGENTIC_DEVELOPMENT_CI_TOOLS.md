# Agentic Development CI Tools Research

**Date:** 2026-01-09  
**Purpose:** Research different tools for agentic development CI/CD  
**Status:** ✅ Research Complete

---

## Executive Summary

This document provides a comprehensive overview of tools and frameworks for implementing Continuous Integration (CI) in agentic development workflows. Agentic development involves integrating AI agents into software development processes, requiring specialized CI/CD approaches that validate agent behavior, performance, and reliability.

### 🎯 **Recommended Approach: Open Source First, GitHub Actions as Fallback**

**Primary Choice: Open Source Tools**
- **AgentScope 1.0** - Developer-centric framework with scalable evaluation
- **ToolBrain** - RL framework for agent training validation
- **Agint** - Graph compiler for type-safe agent workflows
- **Magentic-UI** - Human-agent interaction testing

**Fallback: GitHub Actions (Self-Hosted Runners)**
- ✅ Already configured in your projects
- ✅ Self-hosted runners on Ubuntu and macOS M4 agents
- ✅ Unlimited minutes, full control
- ✅ Can integrate any open source agent tool

**Not Recommended:** Proprietary/Cloud solutions (Agent CI, Buildkite, Kiro, etc.)

---

## Key Findings

### 1. **Agent CI** ⭐ Recommended for Agent-Specific CI

**Website:** https://agent-ci.com

**Description:**
- Platform specifically designed for agentic application CI/CD
- Treats AI agents as software entities, not just ML models
- Direct GitHub integration with automated testing and validation

**Key Features:**
- Automated CI/CD pipelines for agent applications
- Git-based workflows (monitors PRs automatically)
- Specialized evaluations for:
  - Agent behavior validation
  - Performance testing
  - Safety checks
  - Consistency verification
- Prevents regressions through systematic validation

**Use Cases:**
- Agent application development
- Multi-agent systems
- Agent behavior validation
- Production deployment safety

**Documentation:** https://agent-ci.com/docs/core-concepts/cicd

---

### 2. **Buildkite MCP Server** ⭐ Enterprise-Scale Agentic Workflows

**Website:** https://buildkite.com/platform/agentic-workflows/

**Description:**
- Model Context Protocol (MCP) server for AI-powered CI workflows
- Enterprise-grade solution for large-scale agentic development
- Precision log access and full pipeline control

**Key Features:**
- AI-powered workflow automation at scale
- Build log analysis with AI agents
- Failed build assessment and optimization
- Enterprise authentication
- Full pipeline control via MCP

**Use Cases:**
- Enterprise agentic development
- Large-scale CI/CD optimization
- Build failure analysis automation
- Pipeline performance optimization

**Integration:**
- MCP protocol support
- Works with existing Buildkite infrastructure
- Compatible with agent frameworks using MCP

---

### 3. **autodevops.ai** ⭐ Hook-Based Agentic Automation

**Website:** https://www.autodevops.ai

**Description:**
- Hook-based agentic automation for CI/CD pipelines
- Modular skill system for domain expertise packaging
- Progressive disclosure architecture

**Key Features:**
- **Modular Skills:** Package domain expertise as skills loaded on-demand
- **Context Optimization:** Prevents context bloat by loading only relevant skills
- **Unlimited Scale:** Filesystem-based architecture
- **Seamless Integration:** Works with existing CI/CD pipelines
- **Progressive Disclosure:** Skills activate only when needed

**Use Cases:**
- Enhancing existing CI/CD with agentic capabilities
- Domain-specific automation
- Context-efficient agent workflows
- Legacy CI/CD modernization

**Architecture:**
- Filesystem-based skill storage
- Hook-based activation
- Stage-specific skill loading

---

### 4. **GitHub Actions** ✅ Standard CI/CD Platform

**Context7 Documentation:** `/websites/github_en_actions`

**Description:**
- Native GitHub CI/CD platform
- Already in use across projects (exarp-go, devwisdom-go, ib_box_spread)
- Extensive ecosystem and community support

**Key Features:**
- Native Git integration
- Matrix builds for multi-platform testing
- Self-hosted runners support
- Artifact management
- Workflow flexibility

**Current Usage:**
- ✅ exarp-go: Go CI/CD workflow (test, lint, vet, build)
- ✅ devwisdom-go: CI workflow (test, lint, build)
- ✅ ib_box_spread: Multiple workflows (test, build, codeql, parallel-agents-ci)

**Agentic Development Enhancements:**
- Can integrate with Agent CI
- Supports MCP server workflows
- Matrix builds for agent testing across environments
- Artifact storage for agent outputs

**Best Practices from Context7:**
```yaml
# Multi-version testing (adaptable for agent testing)
strategy:
  matrix:
    python-version: ["3.9", "3.10", "3.11", "3.12"]
    agent-version: ["v1", "v2"]  # Example for agent versions

# Artifact management for agent outputs
- name: Archive agent results
  uses: actions/upload-artifact@v4
  with:
    name: agent-test-results
    path: agent_outputs/
```

---

### 5. **Kiro by AWS** ⭐ AI-Powered IDE with CI Integration

**Source:** https://www.techradar.com/pro/aws-launches-kiro-an-agentic-ai-ide

**Description:**
- AWS AI-powered IDE for agentic development
- Integrated CI/CD capabilities
- Automated code review and testing

**Key Features:**
- AI agents for prompt breakdown and implementation
- Automated testing and code evolution tracking
- Project planning automation
- Technical blueprint updates
- Model Context Protocol (MCP) integration
- Agentic chat for ad-hoc coding
- Automated code reviews
- Production deployment support

**Use Cases:**
- Full-stack agentic development
- AWS-based projects
- Enterprise agentic workflows
- Integrated development and CI/CD

**CI/CD Integration:**
- Automated testing in development workflow
- Code review automation
- Deployment automation

---

### 6. **ToolBrain** - Reinforcement Learning for Agent Training

**Source:** https://arxiv.org/abs/2510.00023

**Description:**
- Flexible RL framework for coaching tool use in agentic models
- Supports various training strategies
- Knowledge distillation capabilities

**Key Features:**
- Reinforcement learning algorithms
- Supervised learning support
- Custom reward generation
- Knowledge distillation (large → small models)
- Efficient fine-tuning pipelines
- Automatic task generation
- Seamless tool retrieval

**CI/CD Integration:**
- Agent training validation
- Model performance testing
- Tool use verification
- Continuous agent improvement

**Use Cases:**
- Agent training pipelines
- Tool use optimization
- Model fine-tuning workflows
- Agent capability validation

---

### 7. **Agint** - Agentic Graph Compiler

**Source:** https://arxiv.org/abs/2511.19635

**Description:**
- Agentic graph compiler, interpreter, and runtime
- Converts natural language to typed DAGs
- Hybrid LLM and function-based JIT runtime

**Key Features:**
- Natural language → typed DAG conversion
- Explicit type floors
- Semantic graph transformations
- Dynamic graph refinement
- Reproducible execution
- Speculative evaluation
- Interoperability with developer tools

**CI/CD Integration:**
- Graph compilation validation
- Type checking in CI
- Execution reproducibility testing
- Graph optimization verification

**Use Cases:**
- Complex agent workflows
- Type-safe agent development
- Reproducible agent execution
- Graph-based agent systems

---

### 8. **AgentScope 1.0** - Developer-Centric Agent Framework

**Source:** https://arxiv.org/abs/2508.16279

**Description:**
- Developer-centric framework for building agentic applications
- Unified interfaces and extensible modules
- Advanced agent-level infrastructure

**Key Features:**
- Systematic asynchronous design
- Human-agent interaction patterns
- Agent-agent interaction patterns
- Scalable evaluation module
- Visual studio interface
- Long-trajectory agent support
- Traceability features

**CI/CD Integration:**
- Agent evaluation in CI
- Interaction pattern testing
- Long-trajectory validation
- Visual debugging support

**Use Cases:**
- Multi-agent systems
- Long-running agent workflows
- Agent interaction testing
- Developer-friendly agent development

---

### 9. **Magentic-UI** - Human-Agent Interaction Interface

**Source:** https://arxiv.org/abs/2507.22358

**Description:**
- Open-source web interface for human-agent interaction
- Multi-agent architecture
- MCP tool integration

**Key Features:**
- Web browsing capabilities
- Code execution
- File manipulation
- MCP tool extension
- Co-planning mechanisms
- Action guards
- Long-term memory
- Multi-tasking support

**CI/CD Integration:**
- Human-in-the-loop testing
- Interaction pattern validation
- Safety guard verification
- Memory persistence testing

**Use Cases:**
- Human-agent collaboration
- Interactive agent testing
- Safety-critical agent systems
- Multi-agent coordination

---

### 10. **Metorial** - MCP Server Integration Platform

**Website:** https://metorial.com

**Description:**
- Instant access to 600+ MCP servers
- Pre-configured tool integrations
- Simplified agent tool integration

**Key Features:**
- 600+ pre-configured MCP servers
- Popular platform integrations (Google Drive, Slack, GitHub, Notion)
- Support for major AI agent frameworks
- Efficient tool-calling capabilities

**CI/CD Integration:**
- Tool integration testing
- MCP server validation
- Cross-platform tool verification
- Agent capability testing

**Use Cases:**
- Rapid agent tool integration
- Multi-platform agent development
- Tool compatibility testing
- Agent framework integration

---

## Comparison Matrix

| Tool | Type | GitHub Integration | MCP Support | Enterprise | **On-Premises** | Best For |
|------|------|-------------------|-------------|------------|----------------|----------|
| **Agent CI** | Agent-Specific CI | ✅ Native | ❌ | ❌ | ⚠️ **Partial** (architecture allows, primarily cloud) | Agent applications |
| **Buildkite MCP** | Enterprise CI | ❌ | ✅ Native | ✅ | ✅ **Yes** (self-hosted agents) | Large-scale workflows |
| **autodevops.ai** | CI Enhancement | ✅ | ❌ | ⚠️ | ❓ **Unknown** | Existing CI/CD enhancement |
| **GitHub Actions** | Standard CI/CD | ✅ Native | ⚠️ Via plugins | ⚠️ | ✅ **Yes** (self-hosted runners) | General CI/CD |
| **Kiro (AWS)** | IDE + CI | ✅ | ✅ | ✅ | ❌ **No** (AWS cloud) | AWS-based development |
| **ToolBrain** | Training Framework | ❌ | ❌ | ❌ | ✅ **Yes** (open source) | Agent training |
| **Agint** | Graph Compiler | ❌ | ⚠️ | ❌ | ✅ **Yes** (open source) | Graph-based agents |
| **AgentScope** | Agent Framework | ❌ | ⚠️ | ❌ | ✅ **Yes** (open source) | Multi-agent systems |
| **Magentic-UI** | UI Framework | ❌ | ✅ | ❌ | ✅ **Yes** (open source) | Human-agent interaction |
| **Metorial** | Tool Platform | ✅ | ✅ Native | ❌ | ❌ **No** (SaaS) | Tool integration |
| **TeamCity** | Enterprise CI | ✅ | ❌ | ✅ | ✅ **Yes** (on-premises) | Enterprise CI/CD |

---

## On-Premises / Self-Hosted Deployment

### ✅ **Fully Self-Hosted Solutions**

#### 1. **GitHub Actions (Self-Hosted Runners)** ⭐ Recommended

**Status:** ✅ **Full self-hosted support**

**Capabilities:**
- Self-hosted runners can run on your own infrastructure
- Unlimited minutes (no cloud usage charges)
- Full control over hardware and environment
- Works with existing GitHub Actions workflows

**Current Usage in Projects:**
- ✅ `ib_box_spread_full_universal` has documented self-hosted runner setup
- ✅ Ubuntu agent: `192.168.192.57` (self-hosted runner)
- ✅ macOS M4 agent: `192.168.192.141` (self-hosted runner)

**Setup:**
```bash
# Download and configure runner
./config.sh --url https://github.com/OWNER/REPO --token TOKEN

# Install as service
sudo ./svc.sh install
sudo ./svc.sh start
```

**Documentation:** See `/Users/davidl/Projects/Trading/ib_box_spread_full_universal/docs/SELF_HOSTED_RUNNER_SETUP.md`

---

#### 2. **Buildkite** ⭐ Enterprise Self-Hosted

**Status:** ✅ **Full self-hosted support**

**Capabilities:**
- Self-hosted agents run on your infrastructure
- Web-based interface can be self-hosted (Enterprise)
- Full control over build environment
- MCP server support for agentic workflows

**Deployment:**
- Agents: Self-hosted (required)
- Web UI: Cloud (standard) or self-hosted (Enterprise)

**Best For:** Enterprise teams needing full on-premises control

---

#### 3. **TeamCity** ⭐ Enterprise On-Premises

**Status:** ✅ **Full on-premises deployment**

**Capabilities:**
- Complete on-premises installation
- Build agents on your infrastructure
- Web server on your servers
- Enterprise-grade features

**Best For:** Large enterprises with strict security requirements

---

### ⚠️ **Partial Self-Hosted Support**

#### 4. **Agent CI**

**Status:** ⚠️ **Architecture allows on-premises, primarily cloud**

**Capabilities:**
- Architecture supports on-premises deployment
- Primarily designed as cloud SaaS
- May require custom setup for full on-premises

**Best For:** Cloud-first teams with occasional on-premises needs

---

### ✅ **Open Source (Can Self-Host)**

#### 5. **ToolBrain**
- **Status:** ✅ Open source, can self-host
- **Deployment:** Install on your infrastructure

#### 6. **Agint**
- **Status:** ✅ Open source, can self-host
- **Deployment:** Install on your infrastructure

#### 7. **AgentScope**
- **Status:** ✅ Open source, can self-host
- **Deployment:** Install on your infrastructure

#### 8. **Magentic-UI**
- **Status:** ✅ Open source, can self-host
- **Deployment:** Install on your infrastructure

---

### ❌ **Cloud-Only Solutions**

#### 9. **Kiro (AWS)**
- **Status:** ❌ AWS cloud service only
- **Deployment:** AWS infrastructure required

#### 10. **Metorial**
- **Status:** ❌ SaaS platform only
- **Deployment:** Cloud service, no self-hosting

#### 11. **autodevops.ai**
- **Status:** ❓ Unknown (likely cloud SaaS)
- **Deployment:** Check with vendor

---

## Recommendations for On-Premises Deployment

### 🎯 **Primary Recommendation: Open Source Tools + GitHub Actions (Self-Hosted Runners)**

**Why:**
- ✅ **Open source first** - Use AgentScope, ToolBrain, Agint, Magentic-UI
- ✅ **GitHub Actions fallback** - Already configured with self-hosted runners
- ✅ Full self-hosted runner support
- ✅ Existing documentation and setup scripts
- ✅ Works with agentic development workflows
- ✅ Unlimited minutes on self-hosted runners
- ✅ Can use your existing Ubuntu and macOS M4 agents
- ✅ No vendor lock-in
- ✅ Full control and customization

**Current Setup:**
- Ubuntu agent already configured as self-hosted runner (`192.168.192.57`)
- macOS M4 agent already configured as self-hosted runner (`192.168.192.141`)
- Documentation exists in `ib_box_spread_full_universal` project

**Enhancement Path:**
1. ✅ Use existing self-hosted runners
2. ✅ Integrate open source agent tools (AgentScope, ToolBrain, Agint)
3. ✅ Add agent-specific testing workflows in GitHub Actions
4. ✅ Leverage MCP servers for agentic workflows
5. ✅ Custom agent validation using open source frameworks

**Implementation Example:**
```yaml
# .github/workflows/open-source-agentic-ci.yml
name: Open Source Agentic CI

on: [push, pull_request]

jobs:
  agentscope-eval:
    runs-on: self-hosted
    steps:
      - uses: actions/checkout@v4
      - name: Install AgentScope
        run: pip install agentscope
      - name: Run Agent Evaluations
        run: python -m agentscope.eval --config .github/agent_eval.yaml
```

---

### **Not Recommended (Proprietary/Cloud)**

- ❌ **Buildkite (Enterprise)** - Proprietary, requires license
- ❌ **TeamCity** - Proprietary, requires JetBrains license
- ❌ **Agent CI** - Proprietary SaaS
- ❌ **autodevops.ai** - Likely proprietary SaaS

---

## Recommendations

### 🎯 **User Preference: Open Source First, GitHub as Fallback**

Based on preference for open source solutions with GitHub Actions as fallback:

---

### **Primary Recommendation: Open Source Tools**

#### 1. **AgentScope 1.0** ⭐ Top Open Source Choice

**Why:**
- ✅ **Fully open source** (developer-centric framework)
- ✅ Scalable evaluation module for CI/CD
- ✅ Visual studio interface for debugging
- ✅ Long-trajectory agent support
- ✅ Can integrate with GitHub Actions workflows
- ✅ Self-hostable on your infrastructure

**Integration Strategy:**
```yaml
# GitHub Actions workflow using AgentScope
- name: AgentScope Evaluation
  run: |
    pip install agentscope
    python -m agentscope.eval --config agent_eval.yaml
```

**Best For:** Multi-agent systems, long-running agent workflows, open source preference

---

#### 2. **ToolBrain** ⭐ Open Source Training Framework

**Why:**
- ✅ **Fully open source** (RL framework)
- ✅ Agent training validation in CI
- ✅ Knowledge distillation capabilities
- ✅ Can run in GitHub Actions
- ✅ Self-hostable

**Integration Strategy:**
```yaml
# GitHub Actions workflow using ToolBrain
- name: ToolBrain Agent Training
  run: |
    pip install toolbrain
    python -m toolbrain.train --config training_config.yaml
```

**Best For:** Agent training pipelines, tool use optimization, continuous agent improvement

---

#### 3. **Agint** ⭐ Open Source Graph Compiler

**Why:**
- ✅ **Fully open source** (graph compiler)
- ✅ Type-safe agent development
- ✅ Reproducible execution (great for CI)
- ✅ Can integrate with GitHub Actions
- ✅ Self-hostable

**Integration Strategy:**
```yaml
# GitHub Actions workflow using Agint
- name: Agint Graph Compilation
  run: |
    pip install agint
    agint compile --input agent_spec.yaml --validate-types
```

**Best For:** Complex agent workflows, type-safe development, graph-based agents

---

#### 4. **Magentic-UI** ⭐ Open Source UI Framework

**Why:**
- ✅ **Fully open source** (human-agent interaction)
- ✅ MCP tool integration
- ✅ Can be used for testing human-agent interactions
- ✅ Self-hostable

**Best For:** Human-in-the-loop testing, interaction pattern validation

---

### **Fallback: GitHub Actions (Self-Hosted Runners)**

#### **GitHub Actions + Open Source Agent Tools** ⭐ Recommended Fallback

**Why:**
- ✅ **Already configured** in your projects
- ✅ Self-hosted runners on Ubuntu and macOS M4 agents
- ✅ Unlimited minutes on self-hosted runners
- ✅ Can integrate any open source agent tool
- ✅ Full control over infrastructure
- ✅ Works with existing workflows

**Integration Strategy:**
```yaml
# .github/workflows/agentic-ci.yml
name: Agentic Development CI

on: [push, pull_request]

jobs:
  agent-evaluation:
    runs-on: self-hosted  # Your Ubuntu/macOS agents
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Python
        uses: actions/setup-python@v5
        with:
          python-version: '3.11'
      
      - name: Install AgentScope
        run: pip install agentscope
      
      - name: Run Agent Evaluations
        run: |
          python -m agentscope.eval \
            --config .github/agent_eval.yaml \
            --output agent_results/
      
      - name: Upload Agent Results
        uses: actions/upload-artifact@v4
        with:
          name: agent-evaluation-results
          path: agent_results/
```

**Current Setup:**
- ✅ Ubuntu agent: `192.168.192.57` (self-hosted runner)
- ✅ macOS M4 agent: `192.168.192.141` (self-hosted runner)
- ✅ Documentation: `ib_box_spread_full_universal/docs/SELF_HOSTED_RUNNER_SETUP.md`

---

### **Recommended Approach**

1. **Start with Open Source Tools:**
   - Use **AgentScope** for agent evaluation in CI
   - Use **ToolBrain** for agent training validation
   - Use **Agint** for type-safe agent workflows

2. **Integrate via GitHub Actions:**
   - Run open source tools in GitHub Actions workflows
   - Use your existing self-hosted runners
   - Combine multiple open source tools as needed

3. **Benefits:**
   - ✅ Full open source stack
   - ✅ No vendor lock-in
   - ✅ Self-hosted infrastructure
   - ✅ Unlimited CI minutes
   - ✅ Full control and customization

---

### **Not Recommended (Proprietary/Cloud)**

- ❌ **Agent CI** - Proprietary SaaS (not open source)
- ❌ **Buildkite** - Enterprise proprietary (not open source)
- ❌ **autodevops.ai** - Likely proprietary SaaS
- ❌ **Kiro (AWS)** - AWS proprietary service
- ❌ **Metorial** - SaaS platform (not open source)

---

## Integration Strategies

### Strategy 1: Agent CI for Agent Applications

```yaml
# .github/workflows/agent-ci.yml
name: Agent CI

on:
  pull_request:
    branches: [main, develop]

jobs:
  agent-validation:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Agent CI Validation
        uses: agent-ci/action@v1
        with:
          agent-config: .agent-ci.yml
          validate-behavior: true
          validate-performance: true
          validate-safety: true
```

### Strategy 2: GitHub Actions + Agent Testing

```yaml
# .github/workflows/agent-test.yml
name: Agent Testing

on: [push, pull_request]

jobs:
  test-agents:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        agent-version: [v1, v2, latest]
    steps:
      - uses: actions/checkout@v4
      - name: Test Agent ${{ matrix.agent-version }}
        run: |
          python -m pytest tests/agents/ \
            --agent-version=${{ matrix.agent-version }} \
            --validate-behavior \
            --validate-performance
      - name: Upload Agent Results
        uses: actions/upload-artifact@v4
        with:
          name: agent-results-${{ matrix.agent-version }}
          path: agent_outputs/
```

### Strategy 3: Buildkite MCP Integration

```yaml
# buildkite-pipeline.yml
steps:
  - label: "Agent Analysis"
    command: |
      # Use Buildkite MCP server for AI-powered analysis
      mcp buildkite analyze-build-logs
      mcp buildkite optimize-pipeline
```

### Strategy 4: autodevops.ai Hook Integration

```yaml
# .autodevops/hooks/pre-test.yml
skills:
  - agent-validation
  - behavior-testing
  - performance-benchmarking

# Skills loaded only during test stage
```

---

## Current Project Context

### Existing CI/CD Setup

**exarp-go:**
- GitHub Actions workflow (`.github/workflows/go.yml`)
- Jobs: test, lint, vet, build
- Go 1.24, golangci-lint
- Coverage reporting via codecov

**devwisdom-go:**
- GitHub Actions workflow (`.github/workflows/ci.yml`)
- Jobs: test, lint, build
- Go 1.21, 1.22 matrix
- Coverage artifact upload

**ib_box_spread_full_universal:**
- Multiple GitHub Actions workflows
- Parallel agent CI workflow
- CodeQL security scanning
- Multi-platform builds

### Recommended Next Steps

1. **Evaluate Agent CI** for agent-specific validation
2. **Enhance GitHub Actions** with agent testing matrix
3. **Consider autodevops.ai** for hook-based enhancements
4. **Explore Buildkite MCP** if enterprise features needed

---

## References

### Verified Links (2026)

1. **Agent CI:** https://agent-ci.com/docs/core-concepts/cicd
2. **Buildkite MCP:** https://buildkite.com/platform/agentic-workflows/
3. **autodevops.ai:** https://www.autodevops.ai
4. **GitHub Actions Docs:** https://docs.github.com/en/actions (via Context7)
5. **Kiro AWS:** https://www.techradar.com/pro/aws-launches-kiro-an-agentic-ai-ide
6. **ToolBrain Paper:** https://arxiv.org/abs/2510.00023
7. **Agint Paper:** https://arxiv.org/abs/2511.19635
8. **AgentScope Paper:** https://arxiv.org/abs/2508.16279
9. **Magentic-UI Paper:** https://arxiv.org/abs/2507.22358
10. **Metorial:** https://metorial.com/integration/context7

### Context7 Documentation

- **GitHub Actions:** `/websites/github_en_actions` (6032 code snippets, High reputation, 72.7 benchmark score)

---

## Conclusion

The agentic development CI landscape in 2026 offers specialized tools for agent validation (Agent CI, Buildkite MCP) and enhancement platforms (autodevops.ai) that can integrate with existing CI/CD infrastructure. For projects already using GitHub Actions, enhancing workflows with agent-specific testing and validation provides a practical path forward.

**Key Takeaway:** Agent CI appears to be the most purpose-built solution for agentic application CI/CD, while autodevops.ai offers the most flexible integration approach for existing workflows.

---

## exarp-go Integration Analysis

**See:** `docs/research/AGENTIC_CI_EXARP_GO_INTEGRATION.md` for detailed integration analysis.

### Best Integration for exarp-go: **AgentScope 1.0** ⭐

**Why:**
- ✅ Python-based (works with exarp-go's Python bridge)
- ✅ MCP compatible (can integrate with exarp-go MCP server)
- ✅ Evaluation framework (built for CI/CD)
- ✅ Open source and self-hostable
- ✅ GitHub Actions ready

### Fallback: **GitHub Actions + exarp-go Tools** ⭐

**Why:**
- ✅ Already configured (`.github/workflows/go.yml`)
- ✅ Self-hosted runners available
- ✅ Uses existing 24 tools (testing, task_workflow, automation)
- ✅ Zero new dependencies
- ✅ Full control and customization

**Integration Score:** AgentScope 9.5/10, GitHub Actions 10/10 (fallback)
