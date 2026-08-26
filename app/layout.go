package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ── frame geometry ──────────────────────────────────────────────────────────
//
// Every view is composed as header / body / footer and padded out to the full
// alt-screen height, so the keybar sits on the last row instead of floating
// wherever the content happened to end. The header and footer are a fixed
// number of rows, which is also what makes mouse hit-testing tractable.

const (
	maxContentWidth = 110
	minContentWidth = 40

	headerRows = 5 // border, title, border, subhead, blank
	footerRows = 4 // blank, rule, keybar, status
)

// contentWidth is the width of the centered column every view draws into.
func (m model) contentWidth() int {
	w := m.width - 4
	if w > maxContentWidth {
		w = maxContentWidth
	}
	if w < minContentWidth {
		w = minContentWidth
	}
	if m.width > 0 && w > m.width {
		w = m.width
	}
	return w
}

// contentLeft is the screen column the content column starts at. Mouse
// handlers subtract it before hit-testing.
func (m model) contentLeft() int {
	off := (m.width - m.contentWidth()) / 2
	if off < 0 {
		return 0
	}
	return off
}

// bodyHeight is how many rows the flexible middle section gets.
func (m model) bodyHeight() int {
	h := m.height - headerRows - footerRows
	if h < 3 {
		h = 3
	}
	return h
}

// bodyTop is the screen row the body starts on.
func (m model) bodyTop() int {
	return headerRows
}

// ── header ──────────────────────────────────────────────────────────────────

// appTitle is the product name shown in the header bar, with the TUI version
// appended when the binary was stamped with one.
func appTitle() string {
	title := "UAssetTool TUI"
	v := normalizeVersionTag(currentTUIVersion())
	if v != "" && v != "dev" && v != "(devel)" {
		title += " v" + v
	}
	return title
}

// headerMetaText is the right-hand side of the header bar: which UAssetTool
// build we are driving.
func (m model) headerMetaText() string {
	if m.uatMissing {
		return "UAT not installed"
	}
	if v := normalizeVersionDisplay(m.installedToolVersion()); v != "" {
		return "UAT " + v
	}
	return ""
}

// renderHeader draws the full-width title bar plus a sub-heading row. Always
// headerRows tall.
func (m model) renderHeader(subhead string) string {
	w := m.contentWidth()
	inner := w - headerBar.GetHorizontalFrameSize()
	if inner < 1 {
		inner = 1
	}

	left := titleStyle.Render(appTitle())
	right := headerMeta.Render(m.headerMetaText())
	gap := inner - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		right = ""
		gap = inner - lipgloss.Width(left)
		if gap < 0 {
			left = truncateEnd(left, inner)
			gap = 0
		}
	}
	bar := headerBar.Width(w - headerBar.GetHorizontalBorderSize()).Render(left + strings.Repeat(" ", gap) + right)

	subhead = truncateEnd(subhead, w)
	return bar + "\n" + subhead + "\n" + "\n"
}

// ── footer ──────────────────────────────────────────────────────────────────

// statusTone picks the color for a status line from its wording; the model
// only carries a plain string.
func statusTone(s string) lipgloss.Style {
	low := strings.ToLower(s)
	switch {
	case strings.Contains(low, "failed"), strings.Contains(low, "error"),
		strings.Contains(low, "required"), strings.Contains(low, "not found"):
		return accentRed
	case strings.Contains(low, "copied"), strings.Contains(low, "saved"),
		strings.Contains(low, "success"):
		return accentGreen
	default:
		return accentYellow
	}
}

// renderFooter draws the rule, the keybar and the status row. Always
// footerRows tall.
func (m model) renderFooter(keys string) string {
	w := m.contentWidth()
	rule := footerRule.Render(strings.Repeat("─", w))

	status := ""
	if m.status != "" {
		status = statusTone(m.status).Render(truncateEnd(m.status, w))
	}

	return "\n" + "\n" + rule + "\n" + truncateEnd(keys, w) + "\n" + status
}

// ── composition ─────────────────────────────────────────────────────────────

// frame stacks header, body and footer, pads the body so the footer lands on
// the last row, and centers the whole column horizontally.
func (m model) frame(subhead, body, keys string) string {
	w := m.contentWidth()

	bodyH := m.bodyHeight()
	body = lipgloss.NewStyle().
		Width(w).
		MaxWidth(w).
		Height(bodyH).
		MaxHeight(bodyH).
		Render(body)

	stacked := m.renderHeader(subhead) + body + m.renderFooter(keys)

	if pad := m.contentLeft(); pad > 0 {
		indent := strings.Repeat(" ", pad)
		lines := strings.Split(stacked, "\n")
		for i, l := range lines {
			lines[i] = indent + l
		}
		stacked = strings.Join(lines, "\n")
	}
	return stacked
}

// truncateEnd clips a rendered string to width, appending an ellipsis. It is
// ANSI-naive, so only pass it single-style text or text already sized.
func truncateEnd(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	return lipgloss.NewStyle().MaxWidth(width).Render(s)
}
