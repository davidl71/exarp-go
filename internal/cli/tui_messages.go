// tui_messages.go — TUI tea.Msg types for all async operations.
package cli

import (
	"time"

	"github.com/davidl71/exarp-go/internal/database"
)

// BackgroundJob represents a launched child agent (or other background job).
type BackgroundJob struct {
	Kind      ChildAgentKind
	Prompt    string
	StartedAt time.Time
	Pid       int
	Output    string
}

type jobCompletedMsg struct {
	Pid      int
	Output   string
	ExitCode int
	Err      error
}

type configSection struct {
	name        string
	description string
	keys        []string
}

type taskLoadedMsg struct {
	tasks []*database.Todo2Task
	err   error
}

type scorecardLoadedMsg struct {
	text            string
	recommendations []string
	err             error
}

type handoffLoadedMsg struct {
	text    string
	entries []map[string]interface{}
	err     error
}

type handoffActionDoneMsg struct {
	action  string
	updated int
	err     error
}

type runRecommendationResultMsg struct {
	output string
	err    error
}

type configSectionDetailMsg struct {
	text string
}

type configSaveResultMsg struct {
	message string
	success bool
}

type wavesRefreshDoneMsg struct {
	err error
}

type taskAnalysisLoadedMsg struct {
	text   string
	action string
	err    error
}

type taskAnalysisApproveDoneMsg struct {
	message string
	err     error
}

type moveTaskToWaveDoneMsg struct {
	taskID string
	err    error
}

type updateWavesFromPlanDoneMsg struct {
	message string
	err     error
}

type enqueueWaveDoneMsg struct {
	waveLevel int
	enqueued  int
	err       error
}

type childAgentResultMsg struct {
	Result ChildAgentRunResult
}

type tickMsg time.Time

type configSavedMsg struct{}

// statusUpdateDoneMsg is sent after inline status change (d/i/t/r) so the task list is reloaded.
type statusUpdateDoneMsg struct {
	err error
}

// taskCreatedMsg is sent after inline task creation completes.
type taskCreatedMsg struct {
	taskID string
	err    error
}

// bulkStatusUpdateDoneMsg is sent after bulk status update (D) so the task list is reloaded.
type bulkStatusUpdateDoneMsg struct {
	updated int
	total   int
	err     error
}
