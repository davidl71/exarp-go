// enum_compat.go — Proto enum ↔ legacy string compatibility helpers.
package tools

import "github.com/davidl71/exarp-go/proto"

func outputFormatEnumToString(f proto.OutputFormat) string {
	switch f {
	case proto.OutputFormat_OUTPUT_FORMAT_TEXT:
		return "text"
	case proto.OutputFormat_OUTPUT_FORMAT_MARKDOWN:
		return "markdown"
	case proto.OutputFormat_OUTPUT_FORMAT_JSON:
		return "json"
	default:
		return ""
	}
}

func reportActionEnumToString(a proto.ReportAction) string {
	switch a {
	case proto.ReportAction_REPORT_ACTION_OVERVIEW:
		return "overview"
	case proto.ReportAction_REPORT_ACTION_SCORECARD:
		return "scorecard"
	case proto.ReportAction_REPORT_ACTION_BRIEFING:
		return "briefing"
	case proto.ReportAction_REPORT_ACTION_PRD:
		return "prd"
	default:
		return ""
	}
}

func taskWorkflowActionEnumToString(a proto.TaskWorkflowAction) string {
	switch a {
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_SYNC:
		return "sync"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_APPROVE:
		return "approve"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_CLARIFY:
		return "clarify"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_CLARITY:
		return "clarity"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_CLEANUP:
		return "cleanup"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_CREATE:
		return "create"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_UPDATE:
		return "update"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_DELETE:
		return "delete"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_LIST:
		return "list"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_SHOW:
		return "show"
	default:
		return ""
	}
}

func taskAnalysisActionEnumToString(a proto.TaskAnalysisAction) string {
	switch a {
	case proto.TaskAnalysisAction_TASK_ANALYSIS_ACTION_DUPLICATES:
		return "duplicates"
	case proto.TaskAnalysisAction_TASK_ANALYSIS_ACTION_TAGS:
		return "tags"
	case proto.TaskAnalysisAction_TASK_ANALYSIS_ACTION_DISCOVER_TAGS:
		return "discover_tags"
	case proto.TaskAnalysisAction_TASK_ANALYSIS_ACTION_HIERARCHY:
		return "hierarchy"
	case proto.TaskAnalysisAction_TASK_ANALYSIS_ACTION_DEPENDENCIES:
		return "dependencies"
	case proto.TaskAnalysisAction_TASK_ANALYSIS_ACTION_PARALLELIZATION:
		return "parallelization"
	case proto.TaskAnalysisAction_TASK_ANALYSIS_ACTION_CONFLICTS:
		return "conflicts"
	case proto.TaskAnalysisAction_TASK_ANALYSIS_ACTION_SUGGEST_DEPS:
		return "suggest_deps"
	case proto.TaskAnalysisAction_TASK_ANALYSIS_ACTION_STALE:
		return "stale"
	case proto.TaskAnalysisAction_TASK_ANALYSIS_ACTION_COMPLETABLE:
		return "completable"
	default:
		return ""
	}
}

func sessionActionEnumToString(a proto.SessionAction) string {
	switch a {
	case proto.SessionAction_SESSION_ACTION_PRIME:
		return "prime"
	case proto.SessionAction_SESSION_ACTION_HANDOFF:
		return "handoff"
	case proto.SessionAction_SESSION_ACTION_PROMPTS:
		return "prompts"
	case proto.SessionAction_SESSION_ACTION_ASSIGNEE:
		return "assignee"
	default:
		return ""
	}
}

func healthActionEnumToString(a proto.HealthAction) string {
	switch a {
	case proto.HealthAction_HEALTH_ACTION_SERVER:
		return "server"
	case proto.HealthAction_HEALTH_ACTION_GIT:
		return "git"
	case proto.HealthAction_HEALTH_ACTION_DOCS:
		return "docs"
	case proto.HealthAction_HEALTH_ACTION_DOD:
		return "dod"
	case proto.HealthAction_HEALTH_ACTION_CICD:
		return "cicd"
	case proto.HealthAction_HEALTH_ACTION_TOOLS:
		return "tools"
	case proto.HealthAction_HEALTH_ACTION_CTAGS:
		return "ctags"
	default:
		return ""
	}
}

func securityActionEnumToString(a proto.SecurityAction) string {
	switch a {
	case proto.SecurityAction_SECURITY_ACTION_SCAN:
		return "scan"
	case proto.SecurityAction_SECURITY_ACTION_ALERTS:
		return "alerts"
	case proto.SecurityAction_SECURITY_ACTION_REPORT:
		return "report"
	default:
		return ""
	}
}

func lintActionEnumToString(a proto.LintAction) string {
	switch a {
	case proto.LintAction_LINT_ACTION_RUN:
		return "run"
	case proto.LintAction_LINT_ACTION_ANALYZE:
		return "analyze"
	default:
		return ""
	}
}

func testingActionEnumToString(a proto.TestingAction) string {
	switch a {
	case proto.TestingAction_TESTING_ACTION_RUN:
		return "run"
	case proto.TestingAction_TESTING_ACTION_COVERAGE:
		return "coverage"
	case proto.TestingAction_TESTING_ACTION_SUGGEST:
		return "suggest"
	case proto.TestingAction_TESTING_ACTION_VALIDATE:
		return "validate"
	default:
		return ""
	}
}

func automationActionEnumToString(a proto.AutomationAction) string {
	switch a {
	case proto.AutomationAction_AUTOMATION_ACTION_DAILY:
		return "daily"
	case proto.AutomationAction_AUTOMATION_ACTION_NIGHTLY:
		return "nightly"
	case proto.AutomationAction_AUTOMATION_ACTION_SPRINT:
		return "sprint"
	case proto.AutomationAction_AUTOMATION_ACTION_DISCOVER:
		return "discover"
	default:
		return ""
	}
}

func gitToolsActionEnumToString(a proto.GitToolsAction) string {
	switch a {
	case proto.GitToolsAction_GIT_TOOLS_ACTION_COMMITS:
		return "commits"
	case proto.GitToolsAction_GIT_TOOLS_ACTION_BRANCHES:
		return "branches"
	case proto.GitToolsAction_GIT_TOOLS_ACTION_TASKS:
		return "tasks"
	case proto.GitToolsAction_GIT_TOOLS_ACTION_DIFF:
		return "diff"
	case proto.GitToolsAction_GIT_TOOLS_ACTION_GRAPH:
		return "graph"
	case proto.GitToolsAction_GIT_TOOLS_ACTION_MERGE:
		return "merge"
	case proto.GitToolsAction_GIT_TOOLS_ACTION_SET_BRANCH:
		return "set_branch"
	default:
		return ""
	}
}

