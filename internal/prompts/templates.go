// templates.go — MCP prompt template strings (workflow, persona, advisor, etc.).
package prompts

import (
	"fmt"
	"strings"
)

// promptTemplates contains all prompt templates as string constants.
var promptTemplates = map[string]string{
	"align": `Analyze Todo2 task alignment with project goals.

This prompt will:
1. Evaluate task alignment with project objectives
2. Identify misaligned or out-of-scope tasks
3. Calculate alignment scores for each task
4. Optionally create follow-up tasks for misaligned items

Use: analyze_alignment(action="todo2", create_followup_tasks=True)`,

	"discover": `Discover tasks from various sources in the codebase.

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

This helps ensure no tasks slip through the cracks.`,

	"config": `Generate IDE configuration files for optimal AI assistance.

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
- overwrite=True: Replace existing files`,

	"scan": `Scan all project dependencies for security vulnerabilities.

This prompt will:
1. Scan Python, Rust, and npm dependencies
2. Identify known vulnerabilities
3. Prioritize by severity (critical, high, medium, low)
4. Provide remediation recommendations

Use: security(action="scan") for local pip-audit,
     security(action="alerts") for GitHub Dependabot,
     security(action="report") for combined report.`,

	"scorecard": `Generate a comprehensive project health scorecard with trusted advisor wisdom.

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
- report(action="briefing", overall_score=<score>)`,

	"overview": `Generate a one-page project overview for stakeholders.

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
- output_path="docs/OVERVIEW.md" `,

	"plan": `Generate a Cursor-style plan document (.plan.md) for the project.

Use: report(action="plan")

Structure (see https://cursor.com/learn/creating-plans):
1. Purpose & Success Criteria - What we're building and how we know it's working
2. Technical Foundation - Architecture, tech stack, invariants
3. Iterative Milestones - Checkbox list from Todo2 next actions / backlog order
4. Open Questions - Blockers or decisions needed

Options:
- output_path omitted - Write to .cursor/plans/<project-slug>.plan.md (Cursor discovers it; frontmatter includes status: draft for Build/Built)
- output_path=".cursor/plans/my-feature.plan.md" - Per-feature plan
- plan_title="My Feature" - Override title (default: project name from go.mod)

Use when starting a feature, epic, or sprint to align agents and humans on the plan.`,

	"dashboard": `Display a comprehensive project dashboard with key metrics and status overview.

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
Perfect for: daily check-ins, status updates, project reviews, stakeholder briefings.`,

	"remember": `Use AI session memory to persist insights across sessions.

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

Memory persists between sessions - like having a colleague who remembers!`,

	"daily_checkin": `Daily check-in workflow for project health monitoring.

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

Tip: Use report(action="briefing") with your scorecard metrics for a focused morning briefing.`,

	"sprint_start": `Sprint start workflow for preparing a clean backlog.

Run these tools at the beginning of each sprint:
1. task_analysis(action="duplicates") - Clean up duplicate tasks
2. analyze_alignment(action="todo2") - Ensure tasks align with goals
3. task_workflow(action="approve") - Queue ready tasks for automation
4. task_workflow(action="clarify", sub_action="list") - Identify blocked tasks
5. consult_advisor(stage="planning") - Strategic wisdom from Sun Tzu ⚔️

This ensures a clean, prioritized backlog before starting sprint work.`,

	"sprint_end": `Sprint end workflow for quality assurance and documentation.

Run these tools at the end of each sprint:
1. testing(action="run", coverage=True) - Verify test coverage
2. testing(action="coverage") - Identify coverage gaps
3. health(action="docs") - Ensure docs are updated
4. security(action="report") - Security check before release
5. consult_advisor(stage="review") - Stoic wisdom for accepting results 🏛️

This ensures quality standards are met before sprint completion.`,

	"pre_sprint": `Pre-sprint cleanup workflow.

Run these tools in sequence:
1. task_analysis(action="duplicates") - Find and consolidate duplicates
2. analyze_alignment(action="todo2") - Check task alignment
3. health(action="docs") - Ensure docs are up to date

This ensures a clean task list and aligned goals before starting new work.`,

	"post_impl": `Post-implementation review workflow.

Run these tools after completing a feature:
1. health(action="docs") - Update documentation
2. security(action="report") - Check for new vulnerabilities
3. run_automation(action="discover") - Discover new automation needs

This ensures quality and identifies follow-up work.`,

	"sync": `Synchronize tasks between shared TODO table and Todo2.

This prompt will:
1. Compare tasks across systems
2. Identify missing or out-of-sync tasks
3. Preview or apply changes

Use: task_workflow(action="sync", dry_run=True) first to preview,
     then task_workflow(action="sync", dry_run=False) to apply.`,

	"dups": `Find and consolidate duplicate Todo2 tasks.

This prompt will:
1. Detect duplicate tasks using similarity analysis
2. Group similar tasks together
3. Provide recommendations for consolidation
4. Optionally auto-fix duplicates (merge/delete)

Use: task_analysis(action="duplicates", similarity_threshold=0.85, auto_fix=False)`,

	"context": "Strategically manage LLM context to reduce token usage.\n\n**Tools Available:**\n\n1. **context(action=\"summarize\")** - Compress verbose outputs\n   ```\n   context(action=\"summarize\", data=json_output, level=\"brief\")\n   → \"Health: 85/100, 3 issues, 2 actions\"\n   \n   Levels:\n   - brief: One-line key metrics\n   - detailed: Multi-line with categories\n   - key_metrics: Numbers only\n   - actionable: Recommendations/tasks only\n   \n   ⚡ **Performance Note:** On Apple Silicon Macs (macOS 26+), summarization uses \n   Apple Foundation Models for on-device, privacy-focused processing. Falls back \n   to Python implementation on other platforms.\n   ```\n\n2. **context(action=\"budget\")** - Analyze token usage\n   ```\n   context(action=\"budget\", items=json_array, budget_tokens=4000)\n   → Shows which items to summarize to fit budget\n   ```\n\n3. **context(action=\"batch\")** - Summarize multiple items\n   ```\n   context(action=\"batch\", items=json_array, level=\"brief\")\n   → Combined summaries of multiple items\n   ```\n\n4. **workflow_mode(action=\"focus\")** - Reduce visible tools\n   ```\n   workflow_mode(action=\"focus\", mode=\"security_review\")\n   → 74% fewer tools shown = less context\n   ```\n\n**Context Reduction Strategy:**\n\n| Method | Reduction | Best For |\n|--------|-----------|----------|\n| workflow_mode(action=\"focus\") | 50-80% tools | Start of task |\n| context(action=\"summarize\", level=\"brief\") | 70-90% data | Tool results |\n| context(action=\"summarize\", level=\"key_metrics\") | 80-95% data | Numeric data |\n| context(action=\"budget\") | Planning | Multiple results |\n\n**Example Workflow:**\n\n1. Start task → `workflow_mode(action=\"focus\", mode=\"security_review\")`\n2. Run tool → Get large JSON output\n3. Compress → `context(action=\"summarize\", data=output, level=\"brief\")`\n4. Continue → Reduced context, same key info\n\n**Token Estimation:**\n- ~4 chars per token (rough estimate)\n- Brief summary: 50-100 tokens\n- Full tool output: 500-2000 tokens\n\n**When to Summarize:**\n- After any tool returns >500 tokens\n- Before adding multiple results to context\n- When approaching context limits\n- Before asking follow-up questions",

	"mode": "Suggest the optimal Cursor IDE mode (Agent vs Ask) for a task.\n\n**When to use this prompt:**\n- Before starting a new task to choose the right mode\n- When you're unsure which mode is more efficient\n- To explain mode differences to users\n\n**Usage:**\nrecommend_workflow_mode(task_description=\"your task here\")\n\n**Mode Guidelines:**\n\n🤖 **AGENT Mode** - Best for:\n- Multi-file changes and refactoring\n- Feature implementation from scratch\n- Scaffolding and code generation\n- Automated workflows with many steps\n- Infrastructure and deployment tasks\n\n💬 **ASK Mode** - Best for:\n- Questions and explanations\n- Code review and understanding\n- Single-file edits and fixes\n- Debugging with user guidance\n- Learning and documentation\n\n**Example Analysis:**\n```\nTask: \"Implement user authentication with OAuth2\"\n→ Recommends AGENT (keywords: implement, OAuth2 integration)\n→ Confidence: 85%\n→ Reason: Multi-file implementation task\n\nTask: \"Explain how the auth module works\"\n→ Recommends ASK (keywords: explain, understand)\n→ Confidence: 90%\n→ Reason: Question/explanation request\n```\n\n**How to Switch Modes:**\n1. Look at the top of the chat window\n2. Click the mode selector dropdown\n3. Choose \"Agent\" or \"Ask\"\n\n**Note:** MCP cannot programmatically change your mode - this is advisory only.",

	"task_update": "Update Todo2 task status using proper MCP tools.\n\n⚠️ **CRITICAL: ALWAYS use MCP tools - NEVER edit .todo2/state.todo2.json directly**\n\n**Why use MCP tools:**\n- ✅ Automatic actualHours calculation when marking as Done\n- ✅ Status normalization (Title Case: Todo, In Progress, Done)\n- ✅ Automatic completedAt timestamps\n- ✅ Status change history tracking\n- ✅ Validation and error handling\n- ✅ Proper state machine transitions\n\n**Method 1: Batch Update via MCP Tool (Recommended)**\n```\nmcp_exarp-go_task_workflow(\n    action=\"approve\",\n    status=\"Todo\",           # Filter by current status\n    new_status=\"Done\",       # New status to set\n    task_ids='[\"T-0\", \"T-1\", \"T-2\"]',  # JSON array string\n    clarification_none=True,  # Optional: only tasks with descriptions\n    filter_tag=None,          # Optional: filter by tag\n    dry_run=False            # Preview without applying\n)\n```\n\n**Method 2: Python Utilities**\n```python\nfrom project_management_automation.utils.todo2_mcp_client import (\n    update_todos_mcp,\n    add_comments_mcp\n)\n\n# Batch update status\nupdates = [\n    {\"id\": \"T-0\", \"status\": \"Done\"},\n    {\"id\": \"T-1\", \"status\": \"Done\"},\n]\nupdate_todos_mcp(updates=updates)\n\n# Add result comments\nadd_comments_mcp(\n    todo_id=\"T-0\",\n    comments=[{\"type\": \"result\", \"content\": \"Task completed...\"}]\n)\n```\n\n**Method 3: CLI Command**\n```bash\n./bin/exarp-go -tool task_workflow -args '{\n  \"action\": \"approve\",\n  \"status\": \"Todo\",\n  \"new_status\": \"Done\",\n  \"task_ids\": \"[\\\"T-0\\\", \\\"T-1\\\"]\"\n}'\n```\n\n**🚫 FORBIDDEN:**\n- ❌ Editing .todo2/state.todo2.json directly\n- ❌ Manually modifying task status, timestamps, or metadata\n- ❌ Bypassing the abstraction layer\n\n**Remember: MCP tools exist for a reason. Use them.**",

	// Migrated from Python (docs/PROMPTS_MIGRATION_AND_OBSOLESCENCE_PLAN.md) - Batch 1: docs, automation, workflow
	"docs": `Analyze the project documentation health and identify issues.

This prompt will:
1. Check documentation structure and organization
2. Validate internal and external links
3. Identify broken references and formatting issues
4. Generate a health score (0-100)
5. Optionally create Todo2 tasks for issues found

Use: health(action="docs", create_tasks=True)`,

	"automation_discover": `Discover new automation opportunities in the codebase.

This prompt will:
1. Analyze codebase for repetitive patterns
2. Identify high-value automation opportunities
3. Score opportunities by value and effort
4. Generate recommendations for automation

Use: automation(action="discover")`,

	"weekly_maintenance": `Weekly maintenance workflow.

Run these tools weekly:
1. health(action="docs") - Keep docs healthy
2. task_analysis(action="duplicates") - Clean up duplicates
3. security(action="scan") - Check security
4. task_workflow(action="sync") - Sync across systems

This maintains project health and keeps systems in sync.`,

	"task_review": `Comprehensive task review workflow for backlog hygiene.

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
- Blocked → Resolve dependencies`,

	"project_health": `Comprehensive project health assessment.

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

Use this before major releases or quarterly reviews.`,

	"automation_setup": `One-time automation setup workflow.

Run these tools to enable automated project management:

1. setup_hooks(action="git") - Configure automatic checks on commits
   - pre-commit: docs health, security scan (blocking)
   - pre-push: task alignment, comprehensive security (blocking)
   - post-commit: no-op (run automation discover manually if needed)
   - post-merge: duplicate detection, task sync (non-blocking)

2. setup_hooks(action="patterns") - Configure file change triggers
   - docs/**/*.md → documentation health check
   - src/**/*.py → run tests
   - requirements.txt → security scan

3. Configure cron jobs (manual):
   - Daily: automation(action="daily")
   - Weekly: security(action="scan")
   - See Makefile targets (e.g. make sprint)

After setup, exarp-go will automatically maintain project health.`,

	// Batch 2: advisor
	"advisor_consult": `Consult a trusted advisor for wisdom on your current work.

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
- consult_advisor(tool="testing") - Tool guidance`,

	"advisor_briefing": `Get a morning briefing from trusted advisors based on project health.

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
2. report(action="briefing", overall_score=<score>, ...) - Get focused guidance`,

	// Batch 3: personas
	"persona_developer": `Developer daily workflow for writing quality code.

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
- Bandit Findings: 0 high/critical`,

	"persona_project_manager": `Project Manager workflow for delivery tracking.

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
- Cycle Time: Should be consistent`,

	"persona_code_reviewer": `Code Reviewer workflow for quality gates.

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
- Bandit Findings: 0 in new code`,

	"persona_executive": `Executive/Stakeholder workflow for strategic view.

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
  Review: Uniqueness, Architecture health, Security posture`,

	"persona_security": `Security Engineer workflow for risk management.

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
- Security Score: >90%`,

	"persona_architect": `Architect workflow for system design.

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
- Distance from Main Sequence: <0.3`,

	"persona_qa": `QA Engineer workflow for quality assurance.

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
- First Pass Yield: >85%`,

	"persona_tech_writer": `Technical Writer workflow for documentation.

**Weekly Doc Health (~5 min):**
- health(action="docs") - Full docs analysis
  Check: Broken links, Stale documents, Missing docs
- consult_advisor(metric="documentation") - Confucius wisdom 🎓

**Key Targets:**
- Broken Links: 0
- Stale Docs (>30 days): 0
- Comment Density: 10-30%
- Docstring Coverage: >90%`,

	"tractatus_decompose": `Decompose the given concept using Tractatus Thinking before implementation or design.

1. Call the tractatus_thinking MCP tool with operation="start" and concept set to the user's concept (e.g. "What is X?" or the design question). Obtain session_id from the response.

2. Optionally add propositions with operation="add" (session_id, content, decomposition_type such as "clarification" or "analysis", and optional parent_number). Build the logical structure.

3. Call operation="export" with the session_id and format="markdown" to get the decomposition.

4. Use the exported structure to inform your implementation plan or answer: summarize key propositions and dependencies, then proceed with implementation or design steps.

If the user provided a {concept} placeholder, use it as the concept; otherwise prompt for a concept to decompose.`,

	// Prompt optimization (internal use; see docs/PROMPT_OPTIMIZATION_TEMPLATE_SPEC.md)
	"prompt_optimization_analysis": `Analyze the following prompt for quality. Evaluate it on five dimensions (0.0 to 1.0 each): clarity, specificity, completeness, structure, actionability.

**Prompt to analyze:**
{prompt}

**Context (optional):** {context}

**Task type hint (optional):** {task_type}

**Task type guidance (apply when task_type is set):**
- **code**: Emphasize file paths, APIs, test requirements, concrete inputs/outputs
- **docs**: Emphasize structure, audience, scope, clarity for readers
- **general** (or empty): Balanced evaluation across all dimensions

**Dimensions:**
- **Clarity** (0.0-1.0): Is the task clearly defined? Unambiguous wording?
- **Specificity** (0.0-1.0): Are requirements specific enough? Concrete vs vague?
- **Completeness** (0.0-1.0): Are all necessary details included? No critical gaps?
- **Structure** (0.0-1.0): Is the prompt well-organized? Logical flow?
- **Actionability** (0.0-1.0): Can an AI execute this without clarification?

**Respond with JSON only:**
{
  "clarity": <float>,
  "specificity": <float>,
  "completeness": <float>,
  "structure": <float>,
  "actionability": <float>,
  "summary": "<brief overall assessment>",
  "notes": ["<optional per-dimension notes>"]
}`,

	"prompt_optimization_suggestions": `Given the original prompt and its analysis, generate concrete improvement suggestions for each dimension that scored below 0.8.

**Original prompt:**
{prompt}

**Analysis (JSON):**
{analysis}

**Context (optional):** {context}

**Task type hint (optional):** {task_type}

**Task type guidance (apply when task_type is set):**
- **code**: Focus recommendations on file paths, APIs, test requirements
- **docs**: Focus on structure, audience, scope
- **general** (or empty): Balanced suggestions

For each low-scoring dimension (clarity, specificity, completeness, structure, actionability), identify:
1. **issue** — What is missing or vague in the current prompt
2. **recommendation** — Concrete, actionable improvement (rewritten phrase or instruction)

**Respond with JSON only:**
{
  "suggestions": [
    {
      "dimension": "specificity",
      "issue": "<what is missing or vague>",
      "recommendation": "<concrete improvement>"
    }
  ]
}

If all dimensions score >= 0.8, return an empty suggestions array. Otherwise include one suggestion per dimension below 0.8.`,

	"prompt_optimization_refinement": `Apply the following improvement suggestions to the original prompt. Produce an optimized version that preserves intent while improving clarity, specificity, completeness, structure, and actionability.

**Original prompt:**
{prompt}

**Suggestions to apply (JSON):**
{suggestions}

**Context (optional):** {context}

**Task type hint (optional):** {task_type}

**Task type guidance (apply when task_type is set):**
- **code**: Ensure refined prompt includes file paths, APIs, test requirements
- **docs**: Ensure structure, audience, and scope are clear
- **general** (or empty): Balanced refinement

**Instructions:**
- Preserve the original intent and scope
- Incorporate each recommendation from the suggestions
- Improve wording for clarity and specificity
- Ensure logical structure and flow
- Make the prompt actionable (AI can execute without clarification)
- Return the refined prompt text only — no JSON, no commentary, no before/after labels`,

	// Task breakdown (MODEL_ASSISTED_WORKFLOW Phase 3; see docs/MODEL_ASSISTED_WORKFLOW.md)
	"task_breakdown": `Analyze this task and break it down into 3-8 subtasks.

**Task:**
{task_description}

**Acceptance criteria (if provided):**
{acceptance_criteria}

**Context (optional):** {context}

**Requirements:**
- Each subtask should be independently executable where possible
- Subtasks should have clear acceptance criteria
- Consider dependencies between subtasks (list dependencies explicitly)
- Estimate complexity (simple/medium/complex) for each subtask

**Respond with JSON only:**
{
  "subtasks": [
    {
      "name": "<subtask title>",
      "description": "<brief description>",
      "acceptance_criteria": ["<criterion 1>", "<criterion 2>"],
      "complexity": "simple|medium|complex",
      "dependencies": ["<name of subtask this depends on>"]
    }
  ]
}`,

	// Task breakdown brief (MODEL_ASSISTED_WORKFLOW Phase 3; T-209 — minimal variant for quick breakdowns)
	"task_breakdown_brief": `Break this task into 3-6 subtasks. Task: {task_description}. Context: {context}. Reply with JSON only: {"subtasks":[{"name":"...","description":"...","complexity":"simple|medium|complex","dependencies":[]}]}.`,

	// Auto-execution (MODEL_ASSISTED_WORKFLOW Phase 4; see docs/MODEL_ASSISTED_WORKFLOW.md)
	"task_execution": `Execute this task and propose concrete code or file changes. You are acting as an auto-execution agent for a simple, well-scoped task.

**Task name:** {task_name}

**Task description:**
{task_description}

**Relevant code context (files or snippets):**
{code_context}

**Guidelines or requirements (optional):**
{guidelines}

**Instructions:**
- Produce exactly the changes needed to complete the task
- Each change must specify: file path, location (line or range), and the new content or a clear before/after
- Include a brief explanation and a confidence score (0.0 to 1.0)
- If the task is ambiguous or you cannot safely apply changes, return lower confidence and explain

**Respond with JSON only:**
{
  "changes": [
    {
      "file": "<path>",
      "location": "<line or start_line-end_line>",
      "old_content": "<optional: exact text to replace>",
      "new_content": "<new text or full replacement>"
    }
  ],
  "explanation": "<brief summary of what was done>",
  "confidence": <float 0.0-1.0>
}`,
}

// GetPromptTemplate retrieves a prompt template by name.
func GetPromptTemplate(name string) (string, error) {
	template, ok := promptTemplates[name]
	if !ok {
		return "", fmt.Errorf("unknown prompt: %s", name)
	}

	return template, nil
}

// SubstituteTemplate replaces {variable_name} placeholders with values from args.
// Exported for use by prompt_analyzer and other packages that need template substitution.
func SubstituteTemplate(template string, args map[string]interface{}) string {
	return substituteTemplate(template, args)
}

// substituteTemplate replaces {variable_name} placeholders with values from args.
func substituteTemplate(template string, args map[string]interface{}) string {
	if len(args) == 0 {
		return template // No substitution needed
	}

	result := template

	for key, value := range args {
		placeholder := "{" + key + "}"
		// Convert value to string safely
		var valueStr string

		switch v := value.(type) {
		case string:
			valueStr = v
		case int, int64, float64:
			valueStr = fmt.Sprintf("%v", v)
		case bool:
			valueStr = fmt.Sprintf("%t", v)
		default:
			valueStr = fmt.Sprintf("%v", v)
		}

		result = strings.ReplaceAll(result, placeholder, valueStr)
	}

	return result
}
