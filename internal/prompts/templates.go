package prompts

import (
	"fmt"
	"strings"
)

// promptTemplates contains all prompt templates as string constants
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
}

// GetPromptTemplate retrieves a prompt template by name
func GetPromptTemplate(name string) (string, error) {
	template, ok := promptTemplates[name]
	if !ok {
		return "", fmt.Errorf("unknown prompt: %s", name)
	}
	return template, nil
}

// substituteTemplate replaces {variable_name} placeholders with values from args
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
