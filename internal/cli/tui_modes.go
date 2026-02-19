// tui_modes.go — TUI mode constants and sort order constants.
// Replaces magic strings throughout TUI code with compile-time checked constants.
package cli

// TUI modes - representing different views/screens in the interface.
const (
	ModeTasks         = "tasks"         // Main task list view
	ModeTaskDetail    = "taskDetail"    // Task detail overlay
	ModeConfig        = "config"        // Configuration view
	ModeConfigSection = "configSection" // Config section detail overlay
	ModeScorecard     = "scorecard"     // Project health scorecard view
	ModeHandoffs      = "handoffs"      // Session handoff notes view
	ModeWaves         = "waves"         // Dependency waves view
	ModeJobs          = "jobs"          // Background jobs view
	ModeTaskAnalysis  = "taskAnalysis"  // Task analysis results view
)

// Sort orders for task list.
const (
	SortByID        = "id"
	SortByStatus    = "status"
	SortByPriority  = "priority"
	SortByUpdated   = "updated"
	SortByHierarchy = "hierarchy"
)
