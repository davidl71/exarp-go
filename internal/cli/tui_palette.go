// tui_palette.go — Terminal-adaptive color palette for TUI.
//
// Colors are selected based on whether the terminal has a dark or light
// background, using lipgloss.LightDark. All semantic color names are defined
// here so that styles throughout the TUI reference this palette rather than
// hardcoded hex strings.
package cli

import (
	"image/color"
	"os"

	"charm.land/lipgloss/v2"
)

// tuiLightDark is a LightDarkFunc initialised once at startup from the
// detected terminal background. It defaults to dark when detection fails
// (e.g. when stdout is not a TTY, as in tests or pipes).
var tuiLightDark = lipgloss.LightDark(lipgloss.HasDarkBackground(os.Stdin, os.Stdout))

// Semantic palette — foreground / text colors.
//
// Dark values are chosen for dark terminals (bright, readable on dark BG).
// Light values are chosen for light terminals (deeper, readable on light BG).
var (
	// ColorPrimary is the main accent color used for headers and key elements.
	ColorPrimary color.Color = tuiLightDark(lipgloss.Color("#005f87"), lipgloss.Color("#87ceeb"))

	// ColorSecondary is used for secondary labels and status bars.
	ColorSecondary color.Color = tuiLightDark(lipgloss.Color("#005f5f"), lipgloss.Color("#87d7d7"))

	// ColorAccent is a vivid highlight for selected items and borders.
	ColorAccent color.Color = tuiLightDark(lipgloss.Color("#875f00"), lipgloss.Color("#ffd700"))

	// ColorMuted is a subdued color for help text and less important info.
	ColorMuted color.Color = tuiLightDark(lipgloss.Color("#585858"), lipgloss.Color("#9e9e9e"))

	// ColorSuccess indicates a successful or done state (green family).
	ColorSuccess color.Color = tuiLightDark(lipgloss.Color("#005f00"), lipgloss.Color("#5faf00"))

	// ColorWarning indicates an in-progress or caution state (yellow family).
	ColorWarning color.Color = tuiLightDark(lipgloss.Color("#875f00"), lipgloss.Color("#d7af00"))

	// ColorDanger is used for high priority, errors, and stale IDs (red family).
	ColorDanger color.Color = tuiLightDark(lipgloss.Color("#870000"), lipgloss.Color("#ff5f5f"))

	// ColorInfo is used for Todo/new items (cyan family).
	ColorInfo color.Color = tuiLightDark(lipgloss.Color("#005f87"), lipgloss.Color("#00afaf"))

	// ColorReview is used for the Review status (pink/magenta family).
	ColorReview color.Color = tuiLightDark(lipgloss.Color("#875f87"), lipgloss.Color("#ff87af"))
)

// Semantic palette — background colors.
var (
	// BgHeader is the background for the header bar.
	BgHeader color.Color = tuiLightDark(lipgloss.Color("#d7ff00"), lipgloss.Color("#005f00"))

	// BgHeaderText is the foreground for text on BgHeader.
	BgHeaderText color.Color = tuiLightDark(lipgloss.Color("#000000"), lipgloss.Color("#d7ff00"))

	// BgSelected is the background for a selected/highlighted row.
	BgSelected color.Color = tuiLightDark(lipgloss.Color("#ffffaf"), lipgloss.Color("#5f5f00"))

	// BgSelectedText is the foreground for text on BgSelected.
	BgSelectedText color.Color = tuiLightDark(lipgloss.Color("#000000"), lipgloss.Color("#ffffaf"))

	// BgStatusBar is the background for the bottom status bar.
	BgStatusBar color.Color = tuiLightDark(lipgloss.Color("#dadada"), lipgloss.Color("#1c1c1c"))

	// BgStatusBarText is the foreground for text on BgStatusBar.
	BgStatusBarText color.Color = tuiLightDark(lipgloss.Color("#000000"), lipgloss.Color("#ffffff"))

	// BgStatus is the background for the mode/status badge in the status bar.
	BgStatus color.Color = tuiLightDark(lipgloss.Color("#0087af"), lipgloss.Color("#008080"))

	// BgNormal is the default foreground for non-selected rows.
	BgNormal color.Color = tuiLightDark(lipgloss.Color("#1c1c1c"), lipgloss.Color("#eeeeee"))
)
