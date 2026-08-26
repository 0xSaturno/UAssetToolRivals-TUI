package app

import "strings"

// ── pane sizing and content ─────────────────────────────────────────────────
//
// The output and log panes are bubbles viewports. Everything here keeps them
// sized to the current frame and re-wraps their content when the terminal
// resizes, since the wrapping is done up front rather than by the viewport.

// resizePanes fits both viewports to the current frame and re-wraps whatever
// they are holding.
func (m *model) resizePanes() {
	w := paneContentWidth(m.contentWidth())
	body := m.bodyHeight()

	outH := body - 3 // leading blank + two border rows
	if outH < 3 {
		outH = 3
	}
	logH := body - 5 // spinner, blank, two border rows
	if logH < 3 {
		logH = 3
	}

	m.outVP.Width, m.outVP.Height = w, outH
	m.logVP.Width, m.logVP.Height = w, logH

	m.setOutputContent()
	m.setLogContent()
}

// setOutputContent rebuilds the result pane: release metadata first when a
// download produced it, then the command output.
func (m *model) setOutputContent() {
	w := m.outVP.Width
	lines := m.releaseHeaderLines(w)
	lines = append(lines, hardWrapLines(strings.Split(normalizeBoxText(m.output), "\n"), w)...)
	m.outVP.SetContent(strings.Join(lines, "\n"))
}

// setLogContent rebuilds the live log pane, preserving follow-the-tail unless
// the user has scrolled up.
func (m *model) setLogContent() {
	w := m.logVP.Width
	var lines []string
	if strings.TrimSpace(m.runningOutput) == "" {
		lines = []string{dimStyle.Render("Waiting for UAssetTool output…")}
	} else {
		text := normalizeBoxText(strings.TrimRight(m.runningOutput, "\n"))
		lines = hardWrapLines(strings.Split(text, "\n"), w)
	}
	m.logVP.SetContent(strings.Join(lines, "\n"))
	if m.logFollow {
		m.logVP.GotoBottom()
	}
}

// resizeForm keeps the text inputs matched to the frame width.
func (m *model) resizeForm() {
	w := fieldInputWidth(m.contentWidth())
	for i := range m.formInputs {
		if m.formInputs[i].CharLimit == 1 {
			continue
		}
		m.formInputs[i].Width = w
	}
	m.settingInput.Width = w
}

// ── scrollbar geometry ──────────────────────────────────────────────────────

// thumbMetrics maps a scroll offset onto the scrollbar track.
func thumbMetrics(total, height, offset int) (thumbStart, thumbSize int) {
	if height <= 0 || total <= height {
		return 0, 0
	}
	thumbSize = (height * height) / total
	if thumbSize < 1 {
		thumbSize = 1
	}
	if thumbSize > height {
		thumbSize = height
	}
	trackRange := height - thumbSize
	maxOffset := total - height
	if trackRange <= 0 || maxOffset <= 0 {
		return 0, thumbSize
	}
	if offset < 0 {
		offset = 0
	}
	if offset > maxOffset {
		offset = maxOffset
	}
	return (offset * trackRange) / maxOffset, thumbSize
}

// offsetForTrackRow converts a click on the track into a scroll offset.
func offsetForTrackRow(total, height, row int) int {
	if total <= height || height <= 1 {
		return 0
	}
	if row < 0 {
		row = 0
	}
	if row >= height {
		row = height - 1
	}
	return (row * (total - height)) / (height - 1)
}

// offsetForThumbRow converts a dragged thumb position into a scroll offset.
func offsetForThumbRow(total, height, thumbSize, thumbRow int) int {
	trackRange := height - thumbSize
	maxOffset := total - height
	if trackRange <= 0 || maxOffset <= 0 {
		return 0
	}
	if thumbRow < 0 {
		thumbRow = 0
	}
	if thumbRow > trackRange {
		thumbRow = trackRange
	}
	return (thumbRow * maxOffset) / trackRange
}

// ── pane hit-testing ────────────────────────────────────────────────────────

// paneRect reports where a pane's scrollable content sits on screen: the
// first content row, and the column its scrollbar is drawn in.
func (m model) paneRect(state viewState) (contentTop, scrollbarX, height int) {
	switch state {
	case viewOutput:
		// body: blank row, then the box border
		return m.bodyTop() + 2, m.contentLeft() + m.outVP.Width + 3, m.outVP.Height
	case viewRunning:
		// body: spinner row, blank row, then the box border
		return m.bodyTop() + 3, m.contentLeft() + m.logVP.Width + 3, m.logVP.Height
	}
	return 0, 0, 0
}
