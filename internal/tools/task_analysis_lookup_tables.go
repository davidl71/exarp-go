// task_analysis_lookup_tables.go — Read-only maps/slices for task_analysis (sort/noise/ownership lane globs).
// Hoisted from hot paths to avoid per-compare allocations (sort.Slice) and per-call literal churn.
package tools

import "github.com/davidl71/exarp-go/internal/models"

var (
	// parallelGroupPriorityRank orders PriorityHigh < PriorityMedium < PriorityLow for sort.Slice.
	parallelGroupPriorityRank = map[string]int{
		models.PriorityHigh:   0,
		models.PriorityMedium: 1,
		models.PriorityLow:    2,
	}

	// ownershipConfidenceOrder orders suggestion confidence for sort.Slice (high first).
	ownershipConfidenceOrder = map[string]int{
		"high": 0, "medium": 1, "low": 2, "none": 3, "": 4,
	}

	// noiseFragmentStarters: mid-sentence phrases that suggest non-actionable titles (findNoiseTasks).
	noiseFragmentStarters = []string{
		"that ", "and ", "which ", "the ", "this ", "it ", "these ", "those ",
		"of ", "in ", "for ", "to ", "by ", "is ", "are ", "was ", "were ",
	}

	// noiseActionVerbs: verbs that suggest real tasks.
	noiseActionVerbs = []string{
		"add", "allow", "audit", "build", "check", "clean", "complete", "create",
		"document", "enable", "ensure", "expose", "extract", "fetch", "fix",
		"handle", "implement", "improve", "integrate", "investigate", "list",
		"migrate", "optimize", "refactor", "remove", "replace", "retrieve",
		"review", "run", "scan", "support", "test", "update", "validate",
		"verify", "wire", "write",
	}

	// noiseStatePhrases: completion/state words in title text.
	noiseStatePhrases = []string{
		"pass", "passes", "done", "complete", "completed", "cancelled", "cancellation",
		"required", "requires", "supported", "available",
	}

	// noiseMeaningfulTags: tags that indicate substantive work even if title looks weak.
	noiseMeaningfulTags = map[string]bool{
		"testing": true, "mcp": true, "migration": true, "docs": true, "documentation": true,
		"formid": true, "integration": true, "api": true, "ci": true, "database": true,
	}

	// ownershipLaneToDirPattern: infer_ownership globsForLane — lane → directory name hints.
	ownershipLaneToDirPattern = map[string][]string{
		"backend-auth":        {"auth", "authentication"},
		"backend-api":         {"api", "routes", "handlers"},
		"backend-runtime":     {"server", "service", "backend"},
		"tui-shell":           {"tui", "ui", "shell"},
		"tui-pane":            {"panes", "pane", "views"},
		"docs":                {"docs", "doc", "documentation"},
		"testing":             {"test", "tests", "__tests__"},
		"config":              {"config", ".cursor", ".github"},
		"database":            {"db", "database", "models", "schema"},
		"source-architecture": {"proto", "api"},
	}

	// ownershipLaneKeywords: infer_ownership filesForLane — lane → path substring hints.
	ownershipLaneKeywords = map[string][]string{
		"backend-auth": {"auth"},
		"backend-api":  {"api", "route", "handler"},
		"tui-shell":    {"shell", "app", "input"},
		"tui-pane":     {"pane", "alert", "log", "setting"},
		"docs":         {"doc"},
		"testing":      {"test"},
	}
)
