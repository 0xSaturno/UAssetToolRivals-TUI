package main

import "github.com/charmbracelet/lipgloss"

// ── color palette ───────────────────────────────────────────────────────────

const (
	colorCyan      = "#00D7FF"
	colorGreen     = "#00FF87"
	colorYellow    = "#FFD700"
	colorRed       = "#FF5555"
	colorMagenta   = "#FF79C6"
	colorBlue      = "#6C9EFF"
	colorDim       = "#555555"
	colorText      = "#E0E0E0"
	colorDark      = "#1A1A2E"
	colorBorder    = "#3A3A5C"
	colorSubtle    = "#888888"
	colorBg        = "#0D0D1A"
	colorCardBg    = "#14142B"
	colorHighlight = "#2E2E5E"
)

// ── styles ──────────────────────────────────────────────────────────────────

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorCyan))

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorSubtle)).
			Italic(true)

	itemNormal = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorText)).
			PaddingLeft(3)

	itemSelected = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorDark)).
			Background(lipgloss.Color(colorCyan)).
			Bold(true).
			PaddingLeft(1).
			PaddingRight(1)

	itemDim = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorDim)).
		PaddingLeft(3)

	accentGreen   = lipgloss.NewStyle().Foreground(lipgloss.Color(colorGreen))
	accentYellow  = lipgloss.NewStyle().Foreground(lipgloss.Color(colorYellow))
	accentRed     = lipgloss.NewStyle().Foreground(lipgloss.Color(colorRed))
	accentMagenta = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMagenta))
	accentBlue    = lipgloss.NewStyle().Foreground(lipgloss.Color(colorBlue))
	accentCyan    = lipgloss.NewStyle().Foreground(lipgloss.Color(colorCyan))

	dimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colorDim))

	headerBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colorCyan)).
			Padding(0, 2).
			MarginBottom(1)

	cardBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colorBorder)).
		Padding(1, 2).
		Margin(0, 0, 1, 0)

	successBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colorGreen)).
			Padding(1, 2).
			Margin(1, 0)

	errorBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colorRed)).
			Padding(1, 2).
			Margin(1, 0)

	progressBarFull = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorCyan)).
			Bold(true)

	progressBarEmpty = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorBorder))

	keyHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorCyan)).
			Bold(true)

	keyDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorDim))

	breadcrumbStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorSubtle))

	breadcrumbActive = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorCyan)).
				Bold(true)

	tagOn = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorDark)).
		Background(lipgloss.Color(colorGreen)).
		Bold(true).
		Padding(0, 1)

	tagOff = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorText)).
		Background(lipgloss.Color(colorDim)).
		Padding(0, 1)

	previewCmdStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorYellow)).
			Bold(true)
)
