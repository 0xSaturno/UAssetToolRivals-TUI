package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ── color palette ───────────────────────────────────────────────────────────
//
// AdaptiveColor picks the Light entry on light-background terminals. The dim
// greys are the reason this matters: #555555 is invisible on white.

var (
	colorCyan    = lipgloss.AdaptiveColor{Light: "#0076A3", Dark: "#00D7FF"}
	colorGreen   = lipgloss.AdaptiveColor{Light: "#00875F", Dark: "#00FF87"}
	colorYellow  = lipgloss.AdaptiveColor{Light: "#9A6D00", Dark: "#FFD700"}
	colorRed     = lipgloss.AdaptiveColor{Light: "#C41E1E", Dark: "#FF5555"}
	colorMagenta = lipgloss.AdaptiveColor{Light: "#A8005C", Dark: "#FF79C6"}
	colorBlue    = lipgloss.AdaptiveColor{Light: "#0050C8", Dark: "#6C9EFF"}

	colorText   = lipgloss.AdaptiveColor{Light: "#1C1C1C", Dark: "#E0E0E0"}
	colorSubtle = lipgloss.AdaptiveColor{Light: "#5F5F5F", Dark: "#888888"}
	colorDim    = lipgloss.AdaptiveColor{Light: "#8A8A8A", Dark: "#555555"}
	colorBorder = lipgloss.AdaptiveColor{Light: "#C4C4CE", Dark: "#3A3A5C"}

	// colorOnAccent is the foreground used on top of a saturated accent fill.
	colorOnAccent = lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#101020"}

	colorBg        = lipgloss.AdaptiveColor{Light: "#FAFAFA", Dark: "#0D0D1A"}
	colorCardBg    = lipgloss.AdaptiveColor{Light: "#F0F0F5", Dark: "#14142B"}
	colorHighlight = lipgloss.AdaptiveColor{Light: "#DFE6EE", Dark: "#242449"}
)

// gradientStops returns the hex endpoints used by the progress bar, matched to
// the terminal background so the ramp stays visible either way.
func gradientStops() (string, string) {
	if lipgloss.HasDarkBackground() {
		return "#00D7FF", "#00FF87"
	}
	return "#0076A3", "#00875F"
}

// ── text styles ─────────────────────────────────────────────────────────────

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorCyan)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(colorSubtle).
			Italic(true)

	accentGreen   = lipgloss.NewStyle().Foreground(colorGreen)
	accentYellow  = lipgloss.NewStyle().Foreground(colorYellow)
	accentRed     = lipgloss.NewStyle().Foreground(colorRed)
	accentMagenta = lipgloss.NewStyle().Foreground(colorMagenta)
	accentBlue    = lipgloss.NewStyle().Foreground(colorBlue)
	accentCyan    = lipgloss.NewStyle().Foreground(colorCyan)

	textStyle = lipgloss.NewStyle().Foreground(colorText)
	dimStyle  = lipgloss.NewStyle().Foreground(colorDim)
)

// ── chrome ──────────────────────────────────────────────────────────────────

var (
	headerBar = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorCyan).
			Padding(0, 1)

	headerMeta = lipgloss.NewStyle().Foreground(colorSubtle)

	footerRule = lipgloss.NewStyle().Foreground(colorBorder)

	cardBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(1, 2)
)

// ── menu rows ───────────────────────────────────────────────────────────────

var (
	rowBG = lipgloss.NewStyle().Background(colorHighlight)

	rowSelLabel = lipgloss.NewStyle().
			Background(colorHighlight).
			Foreground(colorText).
			Bold(true)

	rowSelDesc = lipgloss.NewStyle().
			Background(colorHighlight).
			Foreground(colorSubtle)

	itemNormal = lipgloss.NewStyle().Foreground(colorText)

	itemDim = lipgloss.NewStyle().Foreground(colorDim)

	itemSelected = lipgloss.NewStyle().
			Foreground(colorOnAccent).
			Background(colorCyan).
			Bold(true).
			Padding(0, 1)
)

// ── detail pane (drawn on a card background) ────────────────────────────────

var (
	paneBG = lipgloss.NewStyle().Background(colorCardBg)

	paneTitle = lipgloss.NewStyle().
			Background(colorCardBg).
			Foreground(colorCyan).
			Bold(true)

	paneText = lipgloss.NewStyle().
			Background(colorCardBg).
			Foreground(colorText)

	paneDim = lipgloss.NewStyle().
		Background(colorCardBg).
		Foreground(colorSubtle)

	paneLabel = lipgloss.NewStyle().
			Background(colorCardBg).
			Foreground(colorYellow)

	paneBorder = lipgloss.NewStyle().Foreground(colorBorder)
)

// ── form fields ─────────────────────────────────────────────────────────────

var (
	fieldFocused = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorCyan).
			Padding(0, 1)

	fieldBlurred = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	fieldLabelFocused = lipgloss.NewStyle().Foreground(colorCyan).Bold(true)
	fieldLabelBlurred = lipgloss.NewStyle().Foreground(colorSubtle)
)

// ── misc ────────────────────────────────────────────────────────────────────

var (
	progressBarEmpty = lipgloss.NewStyle().Foreground(colorBorder)

	keyHintStyle = lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true)

	keyDescStyle = lipgloss.NewStyle().Foreground(colorDim)

	breadcrumbStyle = lipgloss.NewStyle().Foreground(colorSubtle)

	breadcrumbActive = lipgloss.NewStyle().
				Foreground(colorCyan).
				Bold(true)

	tagOn = lipgloss.NewStyle().
		Foreground(colorOnAccent).
		Background(colorGreen).
		Bold(true).
		Padding(0, 1)

	tagOff = lipgloss.NewStyle().
		Foreground(colorText).
		Background(colorDim).
		Padding(0, 1)

	previewCmdStyle = lipgloss.NewStyle().
			Foreground(colorYellow).
			Bold(true)
)

// padBG right-pads s to width using spaces painted with style, so a highlight
// runs the full width of the row instead of stopping at the text.
func padBG(s string, width int, style lipgloss.Style) string {
	pad := width - lipgloss.Width(s)
	if pad <= 0 {
		return s
	}
	return s + style.Render(strings.Repeat(" ", pad))
}
