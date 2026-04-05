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
	case proto.ReportAction_REPORT_ACTION_EXECUTION_BRIEFING:
		return "execution_briefing"
	case proto.ReportAction_REPORT_ACTION_PLAN:
		return "plan"
	case proto.ReportAction_REPORT_ACTION_SCORECARD_PLANS:
		return "scorecard_plans"
	case proto.ReportAction_REPORT_ACTION_PARALLEL_EXECUTION_PLAN:
		return "parallel_execution_plan"
	case proto.ReportAction_REPORT_ACTION_UPDATE_WAVES_FROM_PLAN:
		return "update_waves_from_plan"
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
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_FIX_DATES:
		return "fix_dates"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_FIX_EMPTY_DESCRIPTIONS:
		return "fix_empty_descriptions"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_FIX_EMPTY_NAMES:
		return "fix_empty_names"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_SANITY_CHECK:
		return "sanity_check"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_FIX_INVALID_IDS:
		return "fix_invalid_ids"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_LINK_PLANNING:
		return "link_planning"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_ADD_COMMENT:
		return "add_comment"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_REQUEST_APPROVAL:
		return "request_approval"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_SYNC_APPROVALS:
		return "sync_approvals"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_APPLY_APPROVAL_RESULT:
		return "apply_approval_result"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_SYNC_FROM_PLAN:
		return "sync_from_plan"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_SYNC_PLAN_STATUS:
		return "sync_plan_status"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_SUMMARIZE:
		return "summarize"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_RUN_WITH_AI:
		return "run_with_ai"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_ENRICH_TOOL_HINTS:
		return "enrich_tool_hints"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_CLAIM:
		return "claim"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_BATCH_CLAIM:
		return "batch_claim"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_RELEASE:
		return "release"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_AGENT_STATUS:
		return "agent_status"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_START_RUN:
		return "start_run"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_END_RUN:
		return "end_run"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_LIST_RUNS:
		return "list_runs"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_SHOW_RUN:
		return "show_run"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_VERIFY:
		return "verify"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_ADD_PROGRESS:
		return "add_progress"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_SPLIT:
		return "split"
	case proto.TaskWorkflowAction_TASK_WORKFLOW_ACTION_IMPORT_SQLITE:
		return "import_sqlite"
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
		return "suggest_dependencies"
	case proto.TaskAnalysisAction_TASK_ANALYSIS_ACTION_STALE:
		return "stale"
	case proto.TaskAnalysisAction_TASK_ANALYSIS_ACTION_COMPLETABLE:
		return "completable"
	case proto.TaskAnalysisAction_TASK_ANALYSIS_ACTION_NEXT_BATCH:
		return "next_batch"
	case proto.TaskAnalysisAction_TASK_ANALYSIS_ACTION_FIX_MISSING_DEPS:
		return "fix_missing_deps"
	case proto.TaskAnalysisAction_TASK_ANALYSIS_ACTION_VALIDATE:
		return "validate"
	case proto.TaskAnalysisAction_TASK_ANALYSIS_ACTION_EXECUTION_PLAN:
		return "execution_plan"
	case proto.TaskAnalysisAction_TASK_ANALYSIS_ACTION_COMPLEXITY:
		return "complexity"
	case proto.TaskAnalysisAction_TASK_ANALYSIS_ACTION_DEPENDENCIES_SUMMARY:
		return "dependencies_summary"
	case proto.TaskAnalysisAction_TASK_ANALYSIS_ACTION_SUGGEST_DEPENDENCIES:
		return "suggest_dependencies"
	case proto.TaskAnalysisAction_TASK_ANALYSIS_ACTION_NOISE:
		return "noise"
	case proto.TaskAnalysisAction_TASK_ANALYSIS_ACTION_INFER_OWNERSHIP:
		return "infer_ownership"
	case proto.TaskAnalysisAction_TASK_ANALYSIS_ACTION_HOTSPOTS:
		return "hotspots"
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
	case proto.GitToolsAction_GIT_TOOLS_ACTION_LOCAL_COMMITS:
		return "local_commits"
	default:
		return ""
	}
}

func gitMergeConflictStrategyEnumToString(s proto.GitMergeConflictStrategy) string {
	switch s {
	case proto.GitMergeConflictStrategy_GIT_MERGE_CONFLICT_STRATEGY_NEWER:
		return "newer"
	case proto.GitMergeConflictStrategy_GIT_MERGE_CONFLICT_STRATEGY_SOURCE:
		return "source"
	case proto.GitMergeConflictStrategy_GIT_MERGE_CONFLICT_STRATEGY_TARGET:
		return "target"
	default:
		return ""
	}
}

func sessionHandoffSubActionEnumToString(s proto.SessionHandoffSubAction) string {
	switch s {
	case proto.SessionHandoffSubAction_SESSION_HANDOFF_SUB_ACTION_END:
		return "end"
	case proto.SessionHandoffSubAction_SESSION_HANDOFF_SUB_ACTION_RESUME:
		return "resume"
	case proto.SessionHandoffSubAction_SESSION_HANDOFF_SUB_ACTION_LATEST:
		return "latest"
	case proto.SessionHandoffSubAction_SESSION_HANDOFF_SUB_ACTION_LIST:
		return "list"
	case proto.SessionHandoffSubAction_SESSION_HANDOFF_SUB_ACTION_SYNC:
		return "sync"
	case proto.SessionHandoffSubAction_SESSION_HANDOFF_SUB_ACTION_EXPORT:
		return "export"
	case proto.SessionHandoffSubAction_SESSION_HANDOFF_SUB_ACTION_CLOSE:
		return "close"
	case proto.SessionHandoffSubAction_SESSION_HANDOFF_SUB_ACTION_APPROVE:
		return "approve"
	case proto.SessionHandoffSubAction_SESSION_HANDOFF_SUB_ACTION_DELETE:
		return "delete"
	default:
		return ""
	}
}

func sessionSyncDirectionEnumToString(d proto.SessionSyncDirection) string {
	switch d {
	case proto.SessionSyncDirection_SESSION_SYNC_DIRECTION_BOTH:
		return "both"
	case proto.SessionSyncDirection_SESSION_SYNC_DIRECTION_PULL:
		return "pull"
	case proto.SessionSyncDirection_SESSION_SYNC_DIRECTION_PUSH:
		return "push"
	default:
		return ""
	}
}

func workflowModeActionEnumToString(a proto.WorkflowModeAction) string {
	switch a {
	case proto.WorkflowModeAction_WORKFLOW_MODE_ACTION_FOCUS:
		return "focus"
	case proto.WorkflowModeAction_WORKFLOW_MODE_ACTION_SUGGEST:
		return "suggest"
	case proto.WorkflowModeAction_WORKFLOW_MODE_ACTION_STATS:
		return "stats"
	default:
		return ""
	}
}

func setupHooksActionEnumToString(a proto.SetupHooksAction) string {
	switch a {
	case proto.SetupHooksAction_SETUP_HOOKS_ACTION_GIT:
		return "git"
	case proto.SetupHooksAction_SETUP_HOOKS_ACTION_PATTERNS:
		return "patterns"
	default:
		return ""
	}
}

func memoryMaintActionEnumToString(a proto.MemoryMaintAction) string {
	switch a {
	case proto.MemoryMaintAction_MEMORY_MAINT_ACTION_HEALTH:
		return "health"
	case proto.MemoryMaintAction_MEMORY_MAINT_ACTION_GC:
		return "gc"
	case proto.MemoryMaintAction_MEMORY_MAINT_ACTION_PRUNE:
		return "prune"
	case proto.MemoryMaintAction_MEMORY_MAINT_ACTION_CONSOLIDATE:
		return "consolidate"
	case proto.MemoryMaintAction_MEMORY_MAINT_ACTION_DREAM:
		return "dream"
	default:
		return ""
	}
}

func memoryMaintMergeStrategyEnumToString(s proto.MemoryMaintMergeStrategy) string {
	switch s {
	case proto.MemoryMaintMergeStrategy_MEMORY_MAINT_MERGE_STRATEGY_NEWEST:
		return "newest"
	case proto.MemoryMaintMergeStrategy_MEMORY_MAINT_MERGE_STRATEGY_OLDEST:
		return "oldest"
	case proto.MemoryMaintMergeStrategy_MEMORY_MAINT_MERGE_STRATEGY_LONGEST:
		return "longest"
	default:
		return ""
	}
}

func memoryMaintScopeEnumToString(s proto.MemoryMaintScope) string {
	switch s {
	case proto.MemoryMaintScope_MEMORY_MAINT_SCOPE_DAY:
		return "day"
	case proto.MemoryMaintScope_MEMORY_MAINT_SCOPE_WEEK:
		return "week"
	case proto.MemoryMaintScope_MEMORY_MAINT_SCOPE_MONTH:
		return "month"
	case proto.MemoryMaintScope_MEMORY_MAINT_SCOPE_ALL:
		return "all"
	default:
		return ""
	}
}

func memoryToolActionEnumToString(a proto.MemoryToolAction) string {
	switch a {
	case proto.MemoryToolAction_MEMORY_TOOL_ACTION_SAVE:
		return "save"
	case proto.MemoryToolAction_MEMORY_TOOL_ACTION_RECALL:
		return "recall"
	case proto.MemoryToolAction_MEMORY_TOOL_ACTION_SEARCH:
		return "search"
	case proto.MemoryToolAction_MEMORY_TOOL_ACTION_LIST:
		return "list"
	default:
		return ""
	}
}

func contextToolActionEnumToString(a proto.ContextToolAction) string {
	switch a {
	case proto.ContextToolAction_CONTEXT_TOOL_ACTION_SUMMARIZE:
		return "summarize"
	case proto.ContextToolAction_CONTEXT_TOOL_ACTION_BUDGET:
		return "budget"
	case proto.ContextToolAction_CONTEXT_TOOL_ACTION_BATCH:
		return "batch"
	default:
		return ""
	}
}

func localLLMSummaryLevelEnumToString(l proto.LocalLLMSummaryLevel) string {
	switch l {
	case proto.LocalLLMSummaryLevel_LOCAL_LLM_SUMMARY_LEVEL_BRIEF:
		return "brief"
	case proto.LocalLLMSummaryLevel_LOCAL_LLM_SUMMARY_LEVEL_DETAILED:
		return "detailed"
	case proto.LocalLLMSummaryLevel_LOCAL_LLM_SUMMARY_LEVEL_KEY_METRICS:
		return "key_metrics"
	case proto.LocalLLMSummaryLevel_LOCAL_LLM_SUMMARY_LEVEL_ACTIONABLE:
		return "actionable"
	default:
		return ""
	}
}

func localLLMDocstringStyleEnumToString(s proto.LocalLLMDocstringStyle) string {
	switch s {
	case proto.LocalLLMDocstringStyle_LOCAL_LLM_DOCSTRING_STYLE_GOOGLE:
		return "google"
	case proto.LocalLLMDocstringStyle_LOCAL_LLM_DOCSTRING_STYLE_NUMPY:
		return "numpy"
	case proto.LocalLLMDocstringStyle_LOCAL_LLM_DOCSTRING_STYLE_SPHINX:
		return "sphinx"
	default:
		return ""
	}
}

func localLLMBackendEnumToString(b proto.LocalLLMBackend) string {
	switch b {
	case proto.LocalLLMBackend_LOCAL_LLM_BACKEND_FM:
		return "fm"
	case proto.LocalLLMBackend_LOCAL_LLM_BACKEND_OLLAMA:
		return "ollama"
	case proto.LocalLLMBackend_LOCAL_LLM_BACKEND_AUTO:
		return "auto"
	default:
		return ""
	}
}
