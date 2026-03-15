// tui_styles.go — TUI lipgloss styles and column width constants.
package cli

import (
	"os"
	"strings"

	"charm.land/bubbles/v2/spinner"
	"charm.land/lipgloss/v2"
)

// detectTerminalColorDepth returns the terminal's color capability.
// 0 = no colors, 8 = basic colors, 256 = 256 colors, 16777216 = truecolor.
// Returns 0 when NO_COLOR is set (https://no-color.org).
func detectTerminalColorDepth() int {
	if os.Getenv("NO_COLOR") != "" {
		return 0
	}
	term := os.Getenv("TERM")
	switch {
	case term == "" || term == "dumb":
		return 0
	case term == "ansi" || term == "vt100":
		return 8
	case term == "xterm" || term == "screen" || term == "tmux":
		return 256
	case term == "xterm-256color" || term == "screen-256color" || term == "tmux-256color":
		return 16777216
	case len(term) > 0 && term[:6] == "xterm":
		return 256
	default:
		return 256 // Default to 256 for modern terminals
	}
}

var terminalColorDepth = detectTerminalColorDepth()

func adaptiveColor(hex16, hex256, hexTrue string) lipgloss.Style {
	switch terminalColorDepth {
	case 0:
		return lipgloss.NewStyle()
	case 8:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(hex16))
	case 256:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(hex256))
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(hexTrue))
	}
}

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

	softBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#808080")).
			Padding(1, 2)

	softBorderHeaderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#00FF00")).
				BorderTop(false).
				Padding(1, 2)

	softBorderSelectedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#FFFF00")).
				Padding(1, 2)

	// SoftBorder uses Unicode rounded box-drawing characters for a softer visual appearance.
	SoftBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "╰",
		BottomRight: "╯",
	}

	// SoftBorderStyle is a panel style with soft rounded borders for unfocused containers.
	SoftBorderStyle = lipgloss.NewStyle().
			Border(SoftBorder).
			BorderForeground(lipgloss.Color("#666666")).
			Padding(0, 1)

	// SoftBorderFocusedStyle is a panel style with soft rounded borders for the active panel.
	SoftBorderFocusedStyle = lipgloss.NewStyle().
				Border(SoftBorder).
				BorderForeground(lipgloss.Color("#999999")).
				Padding(0, 1)

	// SoftBoxStyle is a lightweight container for panel/section grouping (no padding).
	SoftBoxStyle = lipgloss.NewStyle().
			Border(SoftBorder).
			BorderForeground(lipgloss.Color("#555555"))
)

// SpinnerStyleType is a named string type for spinner style configuration.
type SpinnerStyleType string

const (
	SpinnerStyleDefault  SpinnerStyleType = "default"
	SpinnerStyleLine     SpinnerStyleType = "line"
	SpinnerStyleMiniDot  SpinnerStyleType = "minidot"
	SpinnerStyleJump     SpinnerStyleType = "jump"
	SpinnerStylePulse    SpinnerStyleType = "pulse"
	SpinnerStylePoints   SpinnerStyleType = "points"
	SpinnerStyleMeter    SpinnerStyleType = "meter"
	SpinnerStyleEllipsis SpinnerStyleType = "ellipsis"
)

// SpinnerStyleFromString maps a style name to its bubbles/spinner.Spinner value.
// Returns spinner.Line for unknown or empty values.
func SpinnerStyleFromString(s string) spinner.Spinner {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "dot", "default", "":
		return spinner.Dot
	case "line":
		return spinner.Line
	case "minidot":
		return spinner.MiniDot
	case "jump":
		return spinner.Jump
	case "pulse":
		return spinner.Pulse
	case "points":
		return spinner.Points
	case "meter":
		return spinner.Meter
	case "ellipsis":
		return spinner.Ellipsis
	default:
		return spinner.Line
	}
}

const (
	colCursor   = 3
	colIDMedium = 22
	colStatus   = 12
	colPriority = 10
	colPRI      = 4
	colOLD      = 4

	minTermWidth  = 80
	minTermHeight = 24

	// Wide-layout constants: compact column widths to maximize space for Description.
	wideColCursor          = 3
	wideColID              = 22
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
