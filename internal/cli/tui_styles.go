// tui_styles.go — TUI lipgloss styles and column width constants.
package cli

import "github.com/charmbracelet/lipgloss"

var (
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#00FF00")).
			Bold(true).
			Padding(0, 1)

	headerLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#000000")).
				Background(lipgloss.Color("#00FF00")).
				Bold(true)

	headerValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#000000")).
				Background(lipgloss.Color("#00FF00"))

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#000000")).
			Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#008080")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#FFFF00")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#808080"))

	oldIDStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	highPriorityStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF0000")).
				Bold(true)

	mediumPriorityStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFF00"))

	lowPriorityStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00"))

	// Status colors (mirrors 3270 TUI statusColor palette)
	statusDoneStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")) // green
	statusInProgressStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFF00")) // yellow
	statusTodoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")) // cyan/turquoise
	statusReviewStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF69B4")) // pink

	borderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#808080"))
)

const (
	colCursor   = 3
	colIDMedium = 18
	colStatus   = 12
	colPriority = 10
	colPRI      = 4
	colOLD      = 4

	minTermWidth  = 80
	minTermHeight = 24

	// Wide-layout constants: compact column widths to maximize space for Description.
	wideColCursor          = 3
	wideColID              = 18
	wideColStatus          = 11
	wideColPriority        = 8
	wideColPRI             = 3
	wideColOLD             = 3
	wideColSpaces          = 6
	wideFixedColsTotal     = wideColCursor + wideColID + wideColStatus + wideColPriority + wideColPRI + wideColOLD + wideColSpaces
	wideMinDescWidth       = 50
	wideTagsColMin         = 24
	wideShowTagsThreshold  = 160
	wideFixedPlusDescSpace = wideFixedColsTotal + 1
)
