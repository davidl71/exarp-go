"""
MCP Prompts for Project Management Automation

Reusable prompt templates that guide users through common workflows
using the consolidated project management automation tools.

Updated for consolidated tool API (all tools use action= parameter).
"""

# ═══════════════════════════════════════════════════════════════════════════════
# DOCUMENTATION PROMPTS
# ═══════════════════════════════════════════════════════════════════════════════

DOCUMENTATION_HEALTH_CHECK = """Analyze the project documentation health and identify issues.

This prompt will:
1. Check documentation structure and organization
2. Validate internal and external links
3. Identify broken references and formatting issues
4. Generate a health score (0-100)
5. Optionally create Todo2 tasks for issues found

Use: health(action="docs", create_tasks=True)"""

DOCUMENTATION_QUICK_CHECK = """Quick documentation health check without creating tasks.

Use: health(action="docs", create_tasks=False)"""

# ═══════════════════════════════════════════════════════════════════════════════
# TASK MANAGEMENT PROMPTS
# ═══════════════════════════════════════════════════════════════════════════════

TASK_ALIGNMENT_ANALYSIS = """Analyze Todo2 task alignment with project goals.

This prompt will:
1. Evaluate task alignment with project objectives
2. Identify misaligned or out-of-scope tasks
3. Calculate alignment scores for each task
4. Optionally create follow-up tasks for misaligned items

Use: analyze_alignment(action="todo2", create_followup_tasks=True)"""

DUPLICATE_TASK_CLEANUP = """Find and consolidate duplicate Todo2 tasks.

This prompt will:
1. Detect duplicate tasks using similarity analysis
2. Group similar tasks together
3. Provide recommendations for consolidation
4. Optionally auto-fix duplicates (merge/delete)

Use: task_analysis(action="duplicates", similarity_threshold=0.85, auto_fix=False)"""

TASK_SYNC = """Synchronize tasks between shared TODO table and Todo2.

This prompt will:
1. Compare tasks across systems
2. Identify missing or out-of-sync tasks
3. Preview or apply changes

Use: task_workflow(action="sync", dry_run=True) first to preview,
     then task_workflow(action="sync", dry_run=False) to apply."""

# ═══════════════════════════════════════════════════════════════════════════════
# SECURITY PROMPTS
# ═══════════════════════════════════════════════════════════════════════════════

SECURITY_SCAN_ALL = """Scan all project dependencies for security vulnerabilities.

This prompt will:
1. Scan Python, Rust, and npm dependencies
2. Identify known vulnerabilities
3. Prioritize by severity (critical, high, medium, low)
4. Provide remediation recommendations

Use: security(action="scan") for local pip-audit,
     security(action="alerts") for GitHub Dependabot,
     security(action="report") for combined report."""

SECURITY_SCAN_PYTHON = """Scan Python dependencies for security vulnerabilities.

Use: security(action="scan", languages=["python"])"""

SECURITY_SCAN_RUST = """Scan Rust dependencies for security vulnerabilities.

Use: security(action="scan", languages=["rust"])"""

# ═══════════════════════════════════════════════════════════════════════════════
# AUTOMATION PROMPTS
# ═══════════════════════════════════════════════════════════════════════════════

AUTOMATION_DISCOVERY = """Discover new automation opportunities in the codebase.

This prompt will:
1. Analyze codebase for repetitive patterns
2. Identify high-value automation opportunities
3. Score opportunities by value and effort
4. Generate recommendations for automation

Use: run_automation(action="discover", min_value_score=0.7)"""

AUTOMATION_HIGH_VALUE = """Find only high-value automation opportunities (score >= 0.8).

Use: run_automation(action="discover", min_value_score=0.8)"""

# ═══════════════════════════════════════════════════════════════════════════════
# WORKFLOW PROMPTS
# ═══════════════════════════════════════════════════════════════════════════════

PRE_SPRINT_CLEANUP = """Pre-sprint cleanup workflow.

Run these tools in sequence:
1. task_analysis(action="duplicates") - Find and consolidate duplicates
2. analyze_alignment(action="todo2") - Check task alignment
3. health(action="docs") - Ensure docs are up to date

This ensures a clean task list and aligned goals before starting new work."""

POST_IMPLEMENTATION_REVIEW = """Post-implementation review workflow.

Run these tools after completing a feature:
1. health(action="docs") - Update documentation
2. security(action="report") - Check for new vulnerabilities
3. run_automation(action="discover") - Discover new automation needs

This ensures quality and identifies follow-up work."""

WEEKLY_MAINTENANCE = """Weekly maintenance workflow.

Run these tools weekly:
1. health(action="docs") - Keep docs healthy
2. task_analysis(action="duplicates") - Clean up duplicates
3. security(action="scan") - Check security
4. task_workflow(action="sync") - Sync across systems

This maintains project health and keeps systems in sync."""

# ═══════════════════════════════════════════════════════════════════════════════
# DAILY WORKFLOW PROMPTS
# ═══════════════════════════════════════════════════════════════════════════════

DAILY_CHECKIN = """Daily check-in workflow for project health monitoring.

Run these tools every morning (5 min):
1. health(action="server") - Verify server is operational
2. report(action="scorecard") - Get current health metrics and overall score
3. consult_advisor(stage="daily_checkin") - Get wisdom from Pistis Sophia 📜
4. task_workflow(action="clarify", sub_action="list") - Identify blockers
5. health(action="git") - Verify Git status across agents

The advisor will provide wisdom matched to your current project health:
- Score < 30%: 🔥 Chaos mode - urgent guidance for every action
- Score 30-60%: 🏗️ Building mode - focus on fundamentals
- Score 60-80%: 🌱 Maturing mode - strategic advice
- Score 80-100%: 🎯 Mastery mode - reflective wisdom

For automated daily maintenance, use run_automation(action="daily").

Tip: Use report(action="briefing") with your scorecard metrics for a focused morning briefing."""

# ═══════════════════════════════════════════════════════════════════════════════
# SPRINT WORKFLOW PROMPTS
# ═══════════════════════════════════════════════════════════════════════════════

SPRINT_START = """Sprint start workflow for preparing a clean backlog.

Run these tools at the beginning of each sprint:
1. task_analysis(action="duplicates") - Clean up duplicate tasks
2. analyze_alignment(action="todo2") - Ensure tasks align with goals
3. task_workflow(action="approve") - Queue ready tasks for automation
4. task_workflow(action="clarify", sub_action="list") - Identify blocked tasks
5. consult_advisor(stage="planning") - Strategic wisdom from Sun Tzu ⚔️

This ensures a clean, prioritized backlog before starting sprint work."""

SPRINT_END = """Sprint end workflow for quality assurance and documentation.

Run these tools at the end of each sprint:
1. testing(action="run", coverage=True) - Verify test coverage
2. testing(action="coverage") - Identify coverage gaps
3. health(action="docs") - Ensure docs are updated
4. security(action="report") - Security check before release
5. consult_advisor(stage="review") - Stoic wisdom for accepting results 🏛️

This ensures quality standards are met before sprint completion."""

# ═══════════════════════════════════════════════════════════════════════════════
# TASK REVIEW WORKFLOW
# ═══════════════════════════════════════════════════════════════════════════════

TASK_REVIEW = """Comprehensive task review workflow for backlog hygiene.

Run monthly or after major project changes:
1. task_analysis(action="duplicates") - Find and merge duplicates
2. analyze_alignment(action="todo2") - Check task-goal alignment  
3. task_workflow(action="clarify", sub_action="list") - Review blocked tasks
4. task_analysis(action="hierarchy") - Review task structure
5. task_workflow(action="approve") - Queue reviewed tasks

Categories to evaluate:
- Duplicates → Merge or remove
- Misaligned → Re-scope or cancel
- Obsolete → Cancel if work already done
- Stale (>30 days) → Review priority or cancel
- Blocked → Resolve dependencies"""

# ═══════════════════════════════════════════════════════════════════════════════
# PROJECT HEALTH PROMPT
# ═══════════════════════════════════════════════════════════════════════════════

PROJECT_HEALTH = """Comprehensive project health assessment.

Run these tools for a full health check:
1. health(action="server") - Server operational status
2. health(action="docs") - Documentation score
3. testing(action="run", coverage=True) - Test results and coverage
4. testing(action="coverage") - Coverage gap analysis
5. security(action="report") - Security vulnerabilities
6. health(action="cicd") - CI/CD pipeline status
7. analyze_alignment(action="todo2") - Task alignment with goals

This provides a complete picture of:
- Code quality (tests, coverage)
- Documentation health
- Security posture
- CI/CD reliability
- Project management state

Use this before major releases or quarterly reviews."""

# ═══════════════════════════════════════════════════════════════════════════════
# AUTOMATION SETUP PROMPT
# ═══════════════════════════════════════════════════════════════════════════════

AUTOMATION_SETUP = """One-time automation setup workflow.

Run these tools to enable automated project management:

1. setup_hooks(action="git") - Configure automatic checks on commits
   - pre-commit: docs health, security scan (blocking)
   - pre-push: task alignment, comprehensive security (blocking)
   - post-commit: automation discovery (non-blocking)
   - post-merge: duplicate detection, task sync (non-blocking)

2. setup_hooks(action="patterns") - Configure file change triggers
   - docs/**/*.md → documentation health check
   - src/**/*.py → run tests
   - requirements.txt → security scan

3. Configure cron jobs (manual):
   - Daily: run_automation(action="daily")
   - Weekly: security(action="scan")
   - See scripts/cron/*.sh for examples

After setup, Exarp will automatically maintain project health."""

# ═══════════════════════════════════════════════════════════════════════════════
# MODE SUGGESTION PROMPT
# ═══════════════════════════════════════════════════════════════════════════════

# ═══════════════════════════════════════════════════════════════════════════════
# SESSION HANDOFF PROMPTS
# ═══════════════════════════════════════════════════════════════════════════════

END_OF_DAY = """End your work session and create a handoff for other developers.

This prompt will:
1. Summarize tasks you worked on today
2. Note any blockers or issues encountered
3. Unassign your tasks so others can pick them up
4. Create a handoff note visible to all machines
5. Warn about uncommitted git changes

Use: session_handoff(action="end", summary="What I worked on", blockers=["Any blockers"], next_steps=["Suggested next steps"])"""

RESUME_SESSION = """Resume work by reviewing the latest handoff from another developer.

This prompt will:
1. Show what the previous developer was working on
2. List any blockers they noted
3. Show unassigned tasks available to pick up
4. Recommend high-priority work to continue

Use: session_handoff(action="resume")"""

VIEW_HANDOFFS = """View recent handoff notes from all developers.

Use: session_handoff(action="list", limit=10)"""

# ═══════════════════════════════════════════════════════════════════════════════
# WORKFLOW PROMPTS
# ═══════════════════════════════════════════════════════════════════════════════

MODE_SUGGESTION = """Suggest the optimal Cursor IDE mode (Agent vs Ask) for a task.

**When to use this prompt:**
- Before starting a new task to choose the right mode
- When you're unsure which mode is more efficient
- To explain mode differences to users

**Usage:**
recommend_workflow_mode(task_description="your task here")

**Mode Guidelines:**

🤖 **AGENT Mode** - Best for:
- Multi-file changes and refactoring
- Feature implementation from scratch
- Scaffolding and code generation
- Automated workflows with many steps
- Infrastructure and deployment tasks

💬 **ASK Mode** - Best for:
- Questions and explanations
- Code review and understanding
- Single-file edits and fixes
- Debugging with user guidance
- Learning and documentation

**Example Analysis:**
```
Task: "Implement user authentication with OAuth2"
→ Recommends AGENT (keywords: implement, OAuth2 integration)
→ Confidence: 85%
→ Reason: Multi-file implementation task

Task: "Explain how the auth module works"
→ Recommends ASK (keywords: explain, understand)
→ Confidence: 90%
→ Reason: Question/explanation request
```

**How to Switch Modes:**
1. Look at the top of the chat window
2. Click the mode selector dropdown
3. Choose "Agent" or "Ask"

**Note:** MCP cannot programmatically change your mode - this is advisory only."""

MODE_SELECT = """Mode-aware workflow selection guide for Cursor IDE sessions.

This prompt helps users select the appropriate workflow mode based on their task type
and current session context. It references the session mode inference system (MODE-002)
and provides decision guidance.

**When to use this prompt:**
- At the start of a new session to choose workflow mode
- When transitioning between different types of work
- To understand mode differences and best practices
- To align with inferred session mode from tool patterns

**Usage:**
1. Check current inferred mode: automation://session/mode resource
2. Use recommend_workflow_mode(task_description="...") for task-specific guidance
3. Reference this prompt in daily_checkin, sprint_start, and planning workflows

**Mode Decision Tree:**

**Task Type → Mode Selection:**
- Multi-file refactoring → AGENT mode
- Feature implementation → AGENT mode  
- Code generation/scaffolding → AGENT mode
- Questions and explanations → ASK mode
- Quick fixes/single file → ASK mode
- Code review → ASK mode
- Manual editing → MANUAL mode (no AI assistance)

**Session Patterns → Mode Inference:**
- High tool frequency (>5/min) + multi-file → AGENT
- Moderate frequency (1-3/min) + single file → ASK
- Low frequency (<1/min) + direct edits → MANUAL

**Integration Points:**
- daily_checkin: Include mode selection guidance
- sprint_start: Recommend mode for sprint tasks
- planning: Suggest mode based on task breakdown
- advisor consultations: Mode-aware advisor selection (MODE-003)

**References:**
- MODE-002: Session mode inference from tool patterns
- MODE-003: Mode-aware advisor guidance
- recommend_workflow_mode(): Task-based mode recommendation"""

# ═══════════════════════════════════════════════════════════════════════════════
# CONTEXT MANAGEMENT PROMPT
# ═══════════════════════════════════════════════════════════════════════════════

CONTEXT_MANAGEMENT = """Strategically manage LLM context to reduce token usage.

**Tools Available:**

1. **context(action="summarize")** - Compress verbose outputs
   ```
   context(action="summarize", data=json_output, level="brief")
   → "Health: 85/100, 3 issues, 2 actions"
   
   Levels:
   - brief: One-line key metrics
   - detailed: Multi-line with categories
   - key_metrics: Numbers only
   - actionable: Recommendations/tasks only
   
   ⚡ **Performance Note:** On Apple Silicon Macs (macOS 26+), summarization uses 
   Apple Foundation Models for on-device, privacy-focused processing. Falls back 
   to Python implementation on other platforms.
   ```

2. **context(action="budget")** - Analyze token usage
   ```
   context(action="budget", items=json_array, budget_tokens=4000)
   → Shows which items to summarize to fit budget
   ```

3. **context(action="batch")** - Summarize multiple items
   ```
   context(action="batch", items=json_array, level="brief")
   → Combined summaries of multiple items
   ```

4. **workflow_mode(action="focus")** - Reduce visible tools
   ```
   workflow_mode(action="focus", mode="security_review")
   → 74% fewer tools shown = less context
   ```

**Context Reduction Strategy:**

| Method | Reduction | Best For |
|--------|-----------|----------|
| workflow_mode(action="focus") | 50-80% tools | Start of task |
| context(action="summarize", level="brief") | 70-90% data | Tool results |
| context(action="summarize", level="key_metrics") | 80-95% data | Numeric data |
| context(action="budget") | Planning | Multiple results |

**Example Workflow:**

1. Start task → `workflow_mode(action="focus", mode="security_review")`
2. Run tool → Get large JSON output
3. Compress → `context(action="summarize", data=output, level="brief")`
4. Continue → Reduced context, same key info

**Token Estimation:**
- ~4 chars per token (rough estimate)
- Brief summary: 50-100 tokens
- Full tool output: 500-2000 tokens

**When to Summarize:**
- After any tool returns >500 tokens
- Before adding multiple results to context
- When approaching context limits
- Before asking follow-up questions"""

# ═══════════════════════════════════════════════════════════════════════════════
# PROJECT REPORTS
# ═══════════════════════════════════════════════════════════════════════════════

PROJECT_OVERVIEW = """Generate a one-page project overview for stakeholders.

Use: report(action="overview")

Sections included:
- Project Info: name, version, type, status
- Health Scorecard: overall score + component breakdown
- Codebase Metrics: files, lines, tools, prompts
- Task Status: total, pending, remaining work
- Project Phases: progress on each phase
- Risks & Blockers: critical issues to address
- Next Actions: prioritized tasks with estimates

Output formats:
- output_format="text" - Terminal-friendly ASCII (default)
- output_format="html" - Styled HTML page
- output_format="markdown" - For GitHub/documentation
- output_format="json" - Structured data

Save to file:
- output_path="docs/OVERVIEW.html"
- output_path="docs/OVERVIEW.md" """

PROJECT_DASHBOARD = """Display a comprehensive project dashboard with key metrics and status overview.

This prompt generates an at-a-glance view of the project including:

1. Overall Health Score (from scorecard)
   - Use: report(action="scorecard") to get overall score and component breakdown

2. Project Overview Summary
   - Use: report(action="overview", output_format="text") for formatted overview
   - Includes: codebase metrics, task status, project phases, risks, next actions

3. Task Priorities Breakdown
   - Query Todo2 tasks by priority and status
   - Show high/medium/low priority tasks that are Todo or In Progress
   - Display task IDs and names

4. Component Health Scores
   - Display all component scores (security, testing, docs, etc.)
   - Highlight areas needing attention (scores < 60%)
   - Show visual progress bars

5. Recent Activity
   - Recent git commits (last 8-10)
   - Current branch status
   - Git remote status

6. Key Metrics Summary
   - Total tasks, completed, pending, blocked
   - Codebase stats (files, lines, tools, prompts)
   - Remaining work estimates

Output Format:
- Display in a readable, organized format suitable for terminal/viewing
- Use tables, sections, and visual indicators (🟢🟡🔴) for status
- Group related information together
- Highlight important findings and recommendations

Use this when you want a quick, comprehensive view of project status and health.
Perfect for: daily check-ins, status updates, project reviews, stakeholder briefings."""

PROJECT_SCORECARD = """Generate a comprehensive project health scorecard with trusted advisor wisdom.

Use: report(action="scorecard")

Metrics evaluated (each with a trusted advisor):
- Security (😈 BOFH) - Paranoid security checks
- Testing (🏛️ Stoics) - Discipline through test coverage
- Documentation (🎓 Confucius) - Teaching through docs
- Completion (⚔️ Sun Tzu) - Strategic task execution
- Alignment (☯️ Tao) - Balance and purpose
- Clarity (🎭 Gracián) - Pragmatic task clarity
- CI/CD (⚗️ Kybalion) - Cause and effect automation
- Dogfooding (🔧 Murphy) - Use your own tools!
- Overall score (0-100%) with production readiness

After running, consult advisors for lowest-scoring metrics:
- consult_advisor(metric="<lowest_metric>", score=<score>)
- report(action="briefing", overall_score=<score>)"""

# ═══════════════════════════════════════════════════════════════════════════════
# WISDOM ADVISOR PROMPTS
# ═══════════════════════════════════════════════════════════════════════════════

ADVISOR_CONSULT = """Consult a trusted advisor for wisdom on your current work.

Each project metric has an assigned advisor with unique perspective:

📊 **Metric Advisors:**
- security → 😈 BOFH: Paranoid, expects users to break everything
- testing → 🏛️ Stoics: Discipline through adversity
- documentation → 🎓 Confucius: Teaching and transmitting wisdom
- completion → ⚔️ Sun Tzu: Strategy and decisive execution
- alignment → ☯️ Tao: Balance, flow, and purpose
- clarity → 🎭 Gracián: Models of clarity and pragmatism
- ci_cd → ⚗️ Kybalion: Cause and effect, mental models
- dogfooding → 🔧 Murphy: Expect failure, plan for it

⏰ **Stage Advisors:**
- daily_checkin → 📜 Pistis Sophia: Enlightenment journey
- planning → ⚔️ Sun Tzu: Strategic planning
- implementation → 💻 Tao of Programming: Natural flow
- debugging → 😈 BOFH: Knows all the ways things break
- review → 🏛️ Stoics: Accepting harsh truths

Use:
- consult_advisor(metric="security", score=75.0) - Metric advice
- consult_advisor(stage="daily_checkin") - Stage advice
- consult_advisor(tool="testing") - Tool guidance"""

ADVISOR_BRIEFING = """Get a morning briefing from trusted advisors based on project health.

Use: report(action="briefing", overall_score=75.0, security_score=80.0, ...)

The briefing focuses on your lowest-scoring metrics, providing:
1. Advisor wisdom matched to score tier
2. Specific encouragement for improvement
3. Context-aware guidance

Score tiers affect advisor tone:
- 🔥 < 30%: Chaos - urgent, every-action guidance
- 🏗️ 30-60%: Building - focus on fundamentals
- 🌱 60-80%: Maturing - strategic advice
- 🎯 80-100%: Mastery - reflective wisdom

Combine with scorecard for context:
1. report(action="scorecard") - Get current scores
2. report(action="briefing", **scores) - Get focused guidance"""

# ADVISOR_AUDIO removed - audio tools migrated to devwisdom-go MCP server

# ═══════════════════════════════════════════════════════════════════════════════
# PERSONA-BASED WORKFLOWS
# ═══════════════════════════════════════════════════════════════════════════════

PERSONA_DEVELOPER = """Developer daily workflow for writing quality code.

**Morning Checkin (~2 min):**
1. report(action="scorecard") - Quick health check
2. task_workflow(action="clarify", sub_action="list") - Any blockers?
3. consult_advisor(stage="daily_checkin") - Morning wisdom 📜

**Before Committing:**
- health(action="docs") if you touched docs
- lint(action="run") - Code quality check
- Git pre-commit hook runs automatically

**Before PR/Push:**
- analyze_alignment(action="todo2") - Is work aligned with goals?
- consult_advisor(stage="review") - Stoic wisdom 🏛️
- Git pre-push hook runs full checks

**During Debugging:**
- consult_advisor(stage="debugging") - BOFH knows breakage 😈
- memory(action="save", category="debug") - Save discoveries

**Key Targets:**
- Cyclomatic Complexity: <10 per function
- Test Coverage: >80%
- Bandit Findings: 0 high/critical"""

PERSONA_PROJECT_MANAGER = """Project Manager workflow for delivery tracking.

**Daily Standup Prep (~3 min):**
1. report(action="scorecard") - Overall health
2. task_workflow(action="clarify", sub_action="list") - What needs decisions?
3. consult_advisor(stage="planning") - Strategic wisdom ⚔️

**Sprint Planning (~15 min):**
1. report(action="overview", output_format="markdown") - Current state
2. task_analysis(action="duplicates") - Clean up backlog
3. analyze_alignment(action="todo2") - Prioritize aligned work
4. analyze_alignment(action="prd") - Check PRD persona mapping

**Sprint Retrospective (~20 min):**
1. report(action="scorecard") - Full analysis
2. consult_advisor(metric="completion") - Sun Tzu on execution ⚔️
3. Review: Cycle time, First pass yield, Estimation accuracy

**Weekly Status Report (~5 min):**
- report(action="overview", output_format="html", output_path="docs/WEEKLY_STATUS.html")

**Key Metrics:**
- Task Completion %: Per sprint goal
- Blocked Tasks: Target 0
- Cycle Time: Should be consistent"""

PERSONA_CODE_REVIEWER = """Code Reviewer workflow for quality gates.

**Pre-Review Check (~1 min):**
- report(action="scorecard") - Changed since main?
- consult_advisor(stage="review") - Stoic equanimity 🏛️

**During Review:**
For complexity concerns:
  lint(action="run")
For security concerns:
  security(action="scan")
For architecture concerns:
  report(action="scorecard") - Full coupling/cohesion

**Review Checklist:**
- [ ] Complexity acceptable? (CC < 10)
- [ ] Tests added/updated?
- [ ] No security issues?
- [ ] Documentation updated?

**Key Targets:**
- Cyclomatic Complexity: <10 new, <15 existing
- Bandit Findings: 0 in new code"""

PERSONA_EXECUTIVE = """Executive/Stakeholder workflow for strategic view.

**Weekly Check (~2 min):**
- report(action="overview", output_format="html")
  One-page summary: health, risks, progress, blockers

**Monthly Review (~10 min):**
- report(action="scorecard", output_format="markdown")
  Review GQM goal achievement
- report(action="briefing") - Advisor wisdom summary

**Executive Dashboard Metrics:**
| Metric | What It Tells You |
|--------|-------------------|
| Health Score (0-100) | Overall project health |
| Goal Alignment % | Building the right things? |
| Security Score | Risk exposure |
| Velocity Trend | Speeding up or slowing? |

**Quarterly Strategy (~30 min):**
- report(action="scorecard")
  Review: Uniqueness, Architecture health, Security posture"""

PERSONA_SECURITY_ENGINEER = """Security Engineer workflow for risk management.

**Daily Scan (~5 min):**
- security(action="scan") - Dependency vulnerabilities
- consult_advisor(metric="security") - BOFH paranoia 😈

**Weekly Deep Scan (~15 min):**
1. security(action="report") - Full combined report
2. report(action="scorecard") - Security score trend
3. consult_advisor(metric="security", score=<current_score>)

**Security Audit (~1 hour):**
- report(action="scorecard") - Full analysis
  Review: All findings, Dependency tree, Security hotspots

**Key Targets:**
- Critical Vulns: 0
- High Vulns: 0
- Bandit High/Critical: 0
- Security Score: >90%"""

PERSONA_ARCHITECT = """Architect workflow for system design.

**Weekly Architecture Review (~15 min):**
- report(action="scorecard")
  Focus: Coupling matrix, Cohesion scores
- consult_advisor(metric="alignment") - Tao balance ☯️

**Before Major Changes:**
1. report(action="scorecard", output_path="before.json")
2. [Make changes]
3. report(action="scorecard", output_path="after.json")
4. Compare architecture impact

**Tech Debt Prioritization (~30 min):**
- report(action="scorecard")
  Review: High complexity, Dead code, Coupling hotspots
- task_analysis(action="hierarchy") - Task structure

**Key Targets:**
- Avg Cyclomatic Complexity: <5
- Max Complexity: <15
- Distance from Main Sequence: <0.3"""

PERSONA_QA_ENGINEER = """QA Engineer workflow for quality assurance.

**Daily Testing Status (~3 min):**
1. testing(action="run") - Run test suite
2. testing(action="coverage") - Coverage report
3. consult_advisor(metric="testing") - Stoic discipline 🏛️

**Sprint Testing Review (~20 min):**
- report(action="scorecard")
  Review: Test coverage %, Test ratio, Failing tests

**Key Targets:**
- Test Coverage: >80%
- Tests Passing: 100%
- Defect Density: <5 per KLOC
- First Pass Yield: >85%"""

PERSONA_TECH_WRITER = """Technical Writer workflow for documentation.

**Weekly Doc Health (~5 min):**
- health(action="docs") - Full docs analysis
  Check: Broken links, Stale documents, Missing docs
- consult_advisor(metric="documentation") - Confucius wisdom 🎓

**Key Targets:**
- Broken Links: 0
- Stale Docs (>30 days): 0
- Comment Density: 10-30%
- Docstring Coverage: >90%"""

# ═══════════════════════════════════════════════════════════════════════════════
# TASK DISCOVERY PROMPTS
# ═══════════════════════════════════════════════════════════════════════════════

TASK_DISCOVERY = """Discover tasks from various sources in the codebase.

Use: task_discovery(action="all") for comprehensive scan

Sources:
- action="comments": Find TODO/FIXME in code files
- action="markdown": Find task lists in *.md files  
- action="orphans": Find orphaned Todo2 tasks
- action="all": All sources combined

Options:
- create_tasks=True: Auto-create Todo2 tasks from discoveries
- file_patterns='["*.py", "*.ts"]': Limit code scanning
- include_fixme=False: Skip FIXME comments

This helps ensure no tasks slip through the cracks."""

# ═══════════════════════════════════════════════════════════════════════════════
# CONFIG GENERATION PROMPTS
# ═══════════════════════════════════════════════════════════════════════════════

CONFIG_GENERATION = """Generate IDE configuration files for optimal AI assistance.

Use: generate_config(action="rules|ignore|simplify")

Actions:
- action="rules": Generate .cursor/rules/*.mdc files
  Tailored rules for your project type and frameworks

- action="ignore": Generate .cursorignore/.cursorindexingignore
  Optimize AI context by excluding noise

- action="simplify": Simplify existing rule files
  Remove redundancy and improve clarity

Options:
- dry_run=True: Preview changes without writing
- analyze_only=True: Only analyze project structure
- overwrite=True: Replace existing files"""

# ═══════════════════════════════════════════════════════════════════════════════
# MEMORY SYSTEM PROMPTS
# ═══════════════════════════════════════════════════════════════════════════════

MEMORY_SYSTEM = """Use AI session memory to persist insights across sessions.

Use: memory(action="save|recall|search")

Actions:
- action="save": Store an insight
  memory(action="save", title="...", content="...", category="debug")
  
- action="recall": Get context for a task before starting
  memory(action="recall", task_id="T-123")
  
- action="search": Find past insights
  memory(action="search", query="authentication flow")

Categories:
- debug: Error solutions, workarounds, root causes
- research: Pre-implementation findings
- architecture: Component relationships, dependencies
- preference: Coding style, workflow preferences
- insight: Sprint patterns, blockers, optimizations

Memory persists between sessions - like having a colleague who remembers!"""

# ═══════════════════════════════════════════════════════════════════════════════
# PROMPT METADATA
# ═══════════════════════════════════════════════════════════════════════════════

PROMPTS = {
    # Documentation
    "doc_health_check": {
        "name": "Documentation Health Check",
        "description": DOCUMENTATION_HEALTH_CHECK,
        "category": "documentation",
        "arguments": []
    },
    "doc_quick_check": {
        "name": "Quick Documentation Check",
        "description": DOCUMENTATION_QUICK_CHECK,
        "category": "documentation",
        "arguments": []
    },
    # Task Management
    "task_alignment": {
        "name": "Task Alignment Analysis",
        "description": TASK_ALIGNMENT_ANALYSIS,
        "category": "tasks",
        "arguments": []
    },
    "duplicate_cleanup": {
        "name": "Duplicate Task Cleanup",
        "description": DUPLICATE_TASK_CLEANUP,
        "category": "tasks",
        "arguments": []
    },
    "task_sync": {
        "name": "Task Synchronization",
        "description": TASK_SYNC,
        "category": "tasks",
        "arguments": []
    },
    "task_discovery": {
        "name": "Task Discovery",
        "description": TASK_DISCOVERY,
        "category": "tasks",
        "arguments": []
    },
    # Security
    "security_scan_all": {
        "name": "Security Scan (All Languages)",
        "description": SECURITY_SCAN_ALL,
        "category": "security",
        "arguments": []
    },
    "security_scan_python": {
        "name": "Security Scan (Python)",
        "description": SECURITY_SCAN_PYTHON,
        "category": "security",
        "arguments": []
    },
    "security_scan_rust": {
        "name": "Security Scan (Rust)",
        "description": SECURITY_SCAN_RUST,
        "category": "security",
        "arguments": []
    },
    # Automation
    "automation_discovery": {
        "name": "Automation Discovery",
        "description": AUTOMATION_DISCOVERY,
        "category": "automation",
        "arguments": []
    },
    "automation_high_value": {
        "name": "High-Value Automation Discovery",
        "description": AUTOMATION_HIGH_VALUE,
        "category": "automation",
        "arguments": []
    },
    # Workflows - Sprint
    "pre_sprint_cleanup": {
        "name": "Pre-Sprint Cleanup Workflow",
        "description": PRE_SPRINT_CLEANUP,
        "category": "workflow",
        "arguments": []
    },
    "post_implementation_review": {
        "name": "Post-Implementation Review Workflow",
        "description": POST_IMPLEMENTATION_REVIEW,
        "category": "workflow",
        "arguments": []
    },
    "weekly_maintenance": {
        "name": "Weekly Maintenance Workflow",
        "description": WEEKLY_MAINTENANCE,
        "category": "workflow",
        "arguments": []
    },
    "daily_checkin": {
        "name": "Daily Check-in Workflow",
        "description": DAILY_CHECKIN,
        "category": "workflow",
        "arguments": []
    },
    "sprint_start": {
        "name": "Sprint Start Workflow",
        "description": SPRINT_START,
        "category": "workflow",
        "arguments": []
    },
    "sprint_end": {
        "name": "Sprint End Workflow",
        "description": SPRINT_END,
        "category": "workflow",
        "arguments": []
    },
    "task_review": {
        "name": "Task Review Workflow",
        "description": TASK_REVIEW,
        "category": "workflow",
        "arguments": []
    },
    "project_health": {
        "name": "Project Health Assessment",
        "description": PROJECT_HEALTH,
        "category": "workflow",
        "arguments": []
    },
    "automation_setup": {
        "name": "Automation Setup Workflow",
        "description": AUTOMATION_SETUP,
        "category": "config",
        "arguments": []
    },
    # Reports
    "project_scorecard": {
        "name": "Project Scorecard",
        "description": PROJECT_SCORECARD,
        "category": "reports",
        "arguments": []
    },
    "project_overview": {
        "name": "Project Overview",
        "description": PROJECT_OVERVIEW,
        "category": "reports",
        "arguments": []
    },
    "project_dashboard": {
        "name": "Project Dashboard",
        "description": PROJECT_DASHBOARD,
        "category": "reports",
        "arguments": []
    },
    # Wisdom Advisors
    "advisor_consult": {
        "name": "Consult Trusted Advisor",
        "description": ADVISOR_CONSULT,
        "category": "wisdom",
        "arguments": []
    },
    "advisor_briefing": {
        "name": "Advisor Morning Briefing",
        "description": ADVISOR_BRIEFING,
        "category": "wisdom",
        "arguments": []
    },
    # advisor_audio removed - migrated to devwisdom-go MCP server
    # Memory System
    "memory_system": {
        "name": "AI Session Memory",
        "description": MEMORY_SYSTEM,
        "category": "memory",
        "arguments": []
    },
    # Config Generation
    "config_generation": {
        "name": "Config File Generation",
        "description": CONFIG_GENERATION,
        "category": "config",
        "arguments": []
    },
    # Mode Suggestion
    "mode_suggestion": {
        "name": "Mode Suggestion (Agent vs Ask)",
        "description": MODE_SUGGESTION,
        "category": "workflow",
        "arguments": []
    },
    "mode_select": {
        "name": "Mode-Aware Workflow Selection",
        "description": MODE_SELECT,
        "category": "workflow",
        "arguments": []
    },
    # Context Management
    "context_management": {
        "name": "Context Management & Summarization",
        "description": CONTEXT_MANAGEMENT,
        "category": "workflow",
        "arguments": []
    },
    # Persona-based workflows
    "persona_developer": {
        "name": "Developer Workflow",
        "description": PERSONA_DEVELOPER,
        "category": "persona",
        "arguments": []
    },
    "persona_project_manager": {
        "name": "Project Manager Workflow",
        "description": PERSONA_PROJECT_MANAGER,
        "category": "persona",
        "arguments": []
    },
    "persona_code_reviewer": {
        "name": "Code Reviewer Workflow",
        "description": PERSONA_CODE_REVIEWER,
        "category": "persona",
        "arguments": []
    },
    "persona_executive": {
        "name": "Executive/Stakeholder Workflow",
        "description": PERSONA_EXECUTIVE,
        "category": "persona",
        "arguments": []
    },
    "persona_security": {
        "name": "Security Engineer Workflow",
        "description": PERSONA_SECURITY_ENGINEER,
        "category": "persona",
        "arguments": []
    },
    "persona_architect": {
        "name": "Architect Workflow",
        "description": PERSONA_ARCHITECT,
        "category": "persona",
        "arguments": []
    },
    "persona_qa": {
        "name": "QA Engineer Workflow",
        "description": PERSONA_QA_ENGINEER,
        "category": "persona",
        "arguments": []
    },
    "persona_tech_writer": {
        "name": "Technical Writer Workflow",
        "description": PERSONA_TECH_WRITER,
        "category": "persona",
        "arguments": []
    },
}
