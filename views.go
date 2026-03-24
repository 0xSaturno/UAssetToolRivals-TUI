package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var ansiEscapePattern = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

// ── view router ─────────────────────────────────────────────────────────────

func (m model) View() string {
	switch m.state {
	case viewForm:
		return m.viewForm()
	case viewOutput:
		return m.viewOutput()
	case viewRunning:
		return m.viewRunning()
	case viewDownloading:
		return m.viewDownload()
	case viewSettingInput:
		return m.viewSettingInput()
	case viewPreview:
		return m.viewPreview()
	default:
		return m.viewMenu()
	}
}

// ── breadcrumbs ─────────────────────────────────────────────────────────────

func (m model) breadcrumb() string {
	crumbs := []string{"Home"}
	switch m.state {
	case viewCategory:
		crumbs = append(crumbs, "Commands")
	case viewAssetOps:
		crumbs = append(crumbs, "Commands", "Asset Ops")
	case viewZen:
		crumbs = append(crumbs, "Commands", "Zen/IoStore")
	case viewPak:
		crumbs = append(crumbs, "Commands", "PAK Ops")
	case viewJson:
		crumbs = append(crumbs, "Commands", "JSON")
	case viewNiagara:
		crumbs = append(crumbs, "Commands", "Niagara")
	case viewSettings:
		crumbs = append(crumbs, "Settings")
	}
	if len(crumbs) <= 1 {
		return ""
	}
	var parts []string
	for i, c := range crumbs {
		if i == len(crumbs)-1 {
			parts = append(parts, breadcrumbActive.Render(c))
		} else {
			parts = append(parts, breadcrumbStyle.Render(c))
		}
	}
	return "  " + strings.Join(parts, dimStyle.Render(" › ")) + "\n"
}

// ── key hints ───────────────────────────────────────────────────────────────

func keyHint(key, desc string) string {
	return keyHintStyle.Render(key) + " " + keyDescStyle.Render(desc)
}

func menuKeybar() string {
	return "  " + strings.Join([]string{
		keyHint("↑↓", "navigate"),
		keyHint("⏎", "select"),
		keyHint("esc", "back"),
		keyHint("q", "quit"),
	}, "  ")
}

func formKeybar() string {
	return "  " + strings.Join([]string{
		keyHint("tab", "next"),
		keyHint("↑↓", "navigate"),
		keyHint("⏎", "submit"),
		keyHint("esc", "cancel"),
	}, "  ")
}

// ── spinner ─────────────────────────────────────────────────────────────────

func (m model) spinner() string {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	return accentCyan.Render(frames[m.spinFrame%len(frames)])
}

// ── progress bar ────────────────────────────────────────────────────────────

func renderProgressBar(pct float64, width int) string {
	if width < 10 {
		width = 40
	}
	filled := int(pct * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	bar := progressBarFull.Render(strings.Repeat("█", filled))
	bar += progressBarEmpty.Render(strings.Repeat("░", empty))
	return bar
}

func smartTruncateMiddle(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	if width == 1 {
		return "…"
	}
	r := []rune(s)
	if len(r) <= width {
		return s
	}
	keep := width - 1
	left := keep / 2
	right := keep - left
	if left == 0 {
		return "…" + string(r[len(r)-right:])
	}
	if right == 0 {
		return string(r[:left]) + "…"
	}
	return string(r[:left]) + "…" + string(r[len(r)-right:])
}

func hardWrapLine(s string, width int) []string {
	if width <= 0 {
		return []string{""}
	}
	if s == "" {
		return []string{""}
	}
	if lipgloss.Width(s) <= width {
		return []string{s}
	}

	r := []rune(s)
	var out []string
	start := 0
	for start < len(r) {
		end := start + width
		if end > len(r) {
			end = len(r)
		}
		out = append(out, string(r[start:end]))
		start = end
	}
	return out
}

func hardWrapLines(lines []string, width int) []string {
	if width <= 0 {
		return lines
	}
	wrapped := make([]string, 0, len(lines))
	for _, line := range lines {
		wrapped = append(wrapped, hardWrapLine(line, width)...)
	}
	return wrapped
}

func normalizeBoxText(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s = ansiEscapePattern.ReplaceAllString(s, "")
	return s
}

func padLineRight(s string, width int) string {
	pad := width - lipgloss.Width(s)
	if pad <= 0 {
		return s
	}
	return s + strings.Repeat(" ", pad)
}

func manualBoxContentWidth(outerWidth int) int {
	if outerWidth < 8 {
		outerWidth = 8
	}
	innerWidth := outerWidth - 2
	padX := 2
	scrollbarWidth := 2
	contentWidth := innerWidth - (padX * 2) - scrollbarWidth
	if contentWidth < 1 {
		contentWidth = 1
	}
	return contentWidth
}

func clampScroll(totalLines, visibleLines, scroll int) int {
	maxScroll := totalLines - visibleLines
	if maxScroll < 0 {
		maxScroll = 0
	}
	if scroll < 0 {
		return maxScroll
	}
	if scroll > maxScroll {
		return maxScroll
	}
	return scroll
}

func sliceLinesForScroll(lines []string, visibleLines, scroll int) ([]string, int, int) {
	if visibleLines < 1 {
		visibleLines = 1
	}
	total := len(lines)
	if total == 0 {
		return []string{""}, 0, 1
	}
	scroll = clampScroll(total, visibleLines, scroll)
	end := scroll + visibleLines
	if end > total {
		end = total
	}
	return lines[scroll:end], scroll, total
}

func scrollbarColumn(row, visibleLines, totalLines, scroll int) string {
	if visibleLines <= 0 {
		return " "
	}
	if totalLines <= visibleLines {
		return " "
	}
	thumbSize := (visibleLines * visibleLines) / totalLines
	if thumbSize < 1 {
		thumbSize = 1
	}
	if thumbSize > visibleLines {
		thumbSize = visibleLines
	}
	trackRange := visibleLines - thumbSize
	thumbStart := 0
	maxScroll := totalLines - visibleLines
	if trackRange > 0 && maxScroll > 0 {
		thumbStart = (scroll * trackRange) / maxScroll
	}
	if row >= thumbStart && row < thumbStart+thumbSize {
		return accentCyan.Render("█")
	}
	return dimStyle.Render("│")
}

func renderManualBox(lines []string, outerWidth int, borderStyle lipgloss.Style, visibleLines int, scroll int) string {
	if outerWidth < 8 {
		outerWidth = 8
	}
	innerWidth := outerWidth - 2
	padX := 2
	contentWidth := manualBoxContentWidth(outerWidth)

	lines = hardWrapLines(lines, contentWidth)
	if len(lines) == 0 {
		lines = []string{""}
	}
	visible, scroll, total := sliceLinesForScroll(lines, visibleLines, scroll)

	horizontal := strings.Repeat("─", innerWidth)
	top := borderStyle.Render("╭" + horizontal + "╮")
	bottom := borderStyle.Render("╰" + horizontal + "╯")
	empty := borderStyle.Render("│") + strings.Repeat(" ", innerWidth) + borderStyle.Render("│")

	var b strings.Builder
	b.WriteString(top)
	b.WriteString("\n")
	b.WriteString(empty)
	b.WriteString("\n")
	for i, line := range visible {
		content := strings.Repeat(" ", padX) + padLineRight(line, contentWidth) + strings.Repeat(" ", padX) + scrollbarColumn(i, len(visible), total, scroll) + " "
		b.WriteString(borderStyle.Render("│"))
		b.WriteString(content)
		b.WriteString(borderStyle.Render("│"))
		if i < len(visible)-1 {
			b.WriteString("\n")
		}
	}
	b.WriteString("\n")
	b.WriteString(empty)
	b.WriteString("\n")
	b.WriteString(bottom)
	return b.String()
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return "<1s"
	}
	s := int(d.Seconds())
	if s < 60 {
		return fmt.Sprintf("%ds", s)
	}
	return fmt.Sprintf("%dm%02ds", s/60, s%60)
}

// ── menu header line count (for mouse offset) ──────────────────────────────

func (m model) menuHeaderLines() int {
	menu := m.currentMenu()
	lines := 1 // leading \n
	// headerBox: border-top + content + border-bottom = 3 lines
	lines += 3
	// headerBox MarginBottom(1)
	lines += 1
	// extra \n after headerBox
	lines += 1
	// breadcrumb (only on sub-menus)
	if bc := m.breadcrumb(); bc != "" {
		lines++
	}
	// subtitle
	if menu.subtitle != "" {
		lines++
	}
	// blank line before items
	lines++
	return lines
}

// ── view: menu ──────────────────────────────────────────────────────────────

func (m model) viewMenu() string {
	menu := m.currentMenu()
	var b strings.Builder

	b.WriteString("\n")

	title := titleStyle.Render(menu.title)
	b.WriteString(headerBox.Render(title))
	b.WriteString("\n")

	bc := m.breadcrumb()
	if bc != "" {
		b.WriteString(bc)
	}

	if menu.subtitle != "" {
		b.WriteString("  " + subtitleStyle.Render(menu.subtitle) + "\n")
	}
	b.WriteString("\n")

	menuWidth := m.width - 6
	if menuWidth < 20 {
		menuWidth = 20
	}

	for i, item := range menu.items {
		if i == m.cursor {
			label := smartTruncateMiddle(item.label, menuWidth-6)
			sel := fmt.Sprintf(" %s %s ", item.icon, label)
			b.WriteString(itemSelected.Render(sel))
			if item.desc != "" {
				descWidth := menuWidth - lipgloss.Width(sel) - 4
				if descWidth < 12 {
					descWidth = 12
				}
				b.WriteString("  " + dimStyle.Render(smartTruncateMiddle(item.desc, descWidth)))
			}
		} else {
			icon := item.color.Render(item.icon)
			label := item.color.Render(smartTruncateMiddle(item.label, menuWidth-6))
			b.WriteString(fmt.Sprintf("   %s %s", icon, label))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(menuKeybar())
	b.WriteString("\n")

	if m.status != "" {
		b.WriteString("\n  " + accentYellow.Render(m.status) + "\n")
	}

	return b.String()
}

// ── view: form ──────────────────────────────────────────────────────────────

func (m model) viewForm() string {
	var b strings.Builder

	b.WriteString("\n")
	title := accentGreen.Bold(true).Render("▶ " + m.form.command)
	b.WriteString(headerBox.Render(title))
	b.WriteString("\n")

	for i, f := range m.form.fields {
		label := f.label
		optTag := ""
		if f.optional {
			optTag = " " + dimStyle.Render("(optional)")
		}
		if f.boolToggle {
			label += " [Y/N]"
		}

		if i == m.formCursor {
			b.WriteString(accentCyan.Render("  ▸ ") + accentCyan.Bold(true).Render(label) + optTag)
		} else {
			b.WriteString(dimStyle.Render("    "+label) + optTag)
		}
		b.WriteString("\n")
		b.WriteString("    " + m.formInputs[i].View())
		b.WriteString("\n\n")
	}

	b.WriteString(formKeybar())
	b.WriteString("\n")

	if m.status != "" {
		b.WriteString("\n  " + accentRed.Render("⚠ "+m.status) + "\n")
	}

	return b.String()
}

// ── view: preview ───────────────────────────────────────────────────────────

func (m model) viewPreview() string {
	var b strings.Builder

	b.WriteString("\n")
	title := accentYellow.Bold(true).Render("👁 Command Preview")
	b.WriteString(headerBox.Render(title))
	b.WriteString("\n")

	cmdLine := fmt.Sprintf("UAssetTool.exe %s", m.previewArgs)
	cardStyle := cardBox
	innerWidth := 0
	if m.width > 12 {
		cardWidth := m.width - 4
		if cardWidth > 100 {
			cardWidth = 100
		}
		cardStyle = cardStyle.Width(cardWidth).MaxWidth(cardWidth)
		innerWidth = cardWidth - cardStyle.GetHorizontalFrameSize()
	}
	prompt := dimStyle.Render("$") + " "
	availableWidth := innerWidth - lipgloss.Width(prompt) - 1
	if availableWidth <= 0 {
		availableWidth = 20
	}
	previewLine := prompt + previewCmdStyle.Render(smartTruncateMiddle(cmdLine, availableWidth))
	b.WriteString(cardStyle.Render(previewLine))
	b.WriteString("\n\n")

	b.WriteString("  " + keyHint("Y/⏎", "run command") + "    " + keyHint("N/esc", "cancel"))
	b.WriteString("\n")

	return b.String()
}

// ── view: running ───────────────────────────────────────────────────────────

func (m model) viewRunning() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(headerBox.Render(accentCyan.Bold(true).Render("▶ Running Command")))
	b.WriteString("\n\n")

	b.WriteString("  " + m.spinner() + "  " + accentCyan.Render("Executing UAssetTool..."))
	b.WriteString("\n\n")

	logText := normalizeBoxText(strings.TrimRight(m.runningOutput, "\n"))
	logLines := strings.Split(logText, "\n")
	if strings.TrimSpace(m.runningOutput) == "" {
		logLines = []string{dimStyle.Render("Waiting for UAssetTool output...")}
	}

	maxLines := m.height - 12
	if maxLines < 6 {
		maxLines = 6
	}

	logWidth := m.width - 6
	logWidth -= 2
	if logWidth < 30 {
		logWidth = 30
	}
	if logWidth > 120 {
		logWidth = 120
	}

	contentWidth := manualBoxContentWidth(logWidth)

	for i, line := range logLines {
		logLines[i] = smartTruncateMiddle(line, contentWidth)
	}
	b.WriteString(renderManualBox(logLines, logWidth, lipgloss.NewStyle().Foreground(lipgloss.Color(colorBorder)), maxLines, m.runningScroll))
	b.WriteString("\n")
	b.WriteString("  " + dimStyle.Render("Live debug log from UAT  •  ↑↓ scroll  PgUp/PgDn jump  Home/End  •  Ctrl+C copy"))
	b.WriteString("\n")

	return b.String()
}

// ── view: download ──────────────────────────────────────────────────────────

func (m model) viewDownload() string {
	var b strings.Builder
	p := m.dlProgress

	b.WriteString("\n")
	b.WriteString(headerBox.Render(accentBlue.Bold(true).Render("⬇ Download / Update")))
	b.WriteString("\n")

	if p.phase == "" {
		b.WriteString("  " + m.spinner() + "  " + accentCyan.Render("Fetching release info..."))
		b.WriteString("\n")
		return b.String()
	}

	if p.phase == "downloading" {
		pct := float64(0)
		if p.totalBytes > 0 {
			pct = float64(p.bytesDownloaded) / float64(p.totalBytes)
		}

		b.WriteString("  " + m.spinner() + "  " + accentCyan.Render("Downloading UAssetTool..."))
		b.WriteString("\n\n")

		barWidth := m.width - 10
		if barWidth < 20 {
			barWidth = 40
		}
		if barWidth > 60 {
			barWidth = 60
		}
		b.WriteString("  " + renderProgressBar(pct, barWidth))
		b.WriteString("\n\n")

		stats := fmt.Sprintf("  %s / %s", formatBytes(p.bytesDownloaded), formatBytes(p.totalBytes))
		if p.totalBytes > 0 {
			stats += fmt.Sprintf("  (%d%%)", int(pct*100))
		}
		b.WriteString(accentCyan.Render(stats))
		b.WriteString("\n")

		speedStr := formatBytes(int64(p.speed)) + "/s"
		etaStr := formatDuration(p.eta)
		b.WriteString("  " + dimStyle.Render("Speed: ") + accentGreen.Render(speedStr))
		b.WriteString("    " + dimStyle.Render("ETA: ") + accentYellow.Render(etaStr))
		b.WriteString("\n")

	} else if p.phase == "extracting" {
		b.WriteString("  " + m.spinner() + "  " + accentYellow.Render("Extracting UAssetTool.exe..."))
		b.WriteString("\n\n")
		b.WriteString("  " + renderProgressBar(1.0, 40))
		b.WriteString("\n\n")
		b.WriteString("  " + accentGreen.Render(formatBytes(p.bytesDownloaded)) + dimStyle.Render(" downloaded"))
		b.WriteString("\n")
	}

	return b.String()
}

// ── view: output ────────────────────────────────────────────────────────────

func (m model) viewOutput() string {
	var b strings.Builder

	b.WriteString("\n")

	icon := "✓"
	hdrStyle := accentGreen
	if m.outputErr {
		icon = "✗"
		hdrStyle = accentRed
	}

	b.WriteString(headerBox.Render(hdrStyle.Bold(true).Render(fmt.Sprintf("%s Result", icon))))
	b.WriteString("\n")

	if m.dlInfo != nil && !m.outputErr {
		var info strings.Builder
		info.WriteString(accentCyan.Bold(true).Render("Release: ") + accentGreen.Render(m.dlInfo.TagName))
		if m.dlInfo.Name != "" && m.dlInfo.Name != m.dlInfo.TagName {
			info.WriteString("  " + dimStyle.Render(m.dlInfo.Name))
		}
		info.WriteString("\n")
		info.WriteString(dimStyle.Render("Published: ") + accentYellow.Render(m.dlInfo.PublishedAt.Format("Jan 02, 2006 15:04")))
		info.WriteString("\n")
		if m.dlInfo.Body != "" {
			body := m.dlInfo.Body
			bodyLines := strings.Split(body, "\n")
			maxBodyLines := 12
			if len(bodyLines) > maxBodyLines {
				bodyLines = bodyLines[:maxBodyLines]
				bodyLines = append(bodyLines, dimStyle.Render(fmt.Sprintf("  ... (%d more lines)", len(strings.Split(body, "\n"))-maxBodyLines)))
			}
			info.WriteString("\n" + dimStyle.Render("Release Notes:") + "\n")
			for _, line := range bodyLines {
				info.WriteString("  " + line + "\n")
			}
		}
		b.WriteString(cardBox.Render(info.String()))
		b.WriteString("\n")
	}

	out := normalizeBoxText(m.output)
	lines := strings.Split(out, "\n")
	maxLines := m.height - 14
	if maxLines < 8 {
		maxLines = 8
	}
	boxWidth := m.width - 6
	if boxWidth < 30 {
		boxWidth = 30
	}
	if boxWidth > 140 {
		boxWidth = 140
	}
	contentWidth := manualBoxContentWidth(boxWidth)
	lines = hardWrapLines(lines, contentWidth)
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorGreen))
	if m.outputErr {
		borderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colorRed))
	}
	b.WriteString(renderManualBox(lines, boxWidth, borderStyle, maxLines, m.outputScroll))
	b.WriteString("\n\n")

	b.WriteString("  " + keyHint("↑↓/PgUp/PgDn/Home/End", "scroll") + "    " + keyHint("Ctrl+C", "copy") + "    " + keyHint("⏎/esc", "back to menu"))
	b.WriteString("\n")

	return b.String()
}

// ── view: setting input ─────────────────────────────────────────────────────

func (m model) viewSettingInput() string {
	var b strings.Builder

	b.WriteString("\n")
	title := accentYellow.Bold(true).Render("✏ " + m.settingLabel)
	b.WriteString(headerBox.Render(title))
	b.WriteString("\n\n")
	b.WriteString("  " + m.settingInput.View())
	b.WriteString("\n\n")
	b.WriteString("  " + keyHint("⏎", "save") + "    " + keyHint("esc", "cancel"))
	b.WriteString("\n")

	return b.String()
}
