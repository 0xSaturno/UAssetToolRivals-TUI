package app

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

var ansiEscapePattern = regexp.MustCompile("\x1b\\[[0-9;?]*[ -/]*[@-~]")

// ── view router ─────────────────────────────────────────────────────────────

func (m model) View() string {
	w := m.contentWidth()
	switch m.state {
	case viewForm:
		return m.frame(m.formSubhead(), m.viewForm(), formKeybar(w))
	case viewOutput:
		return m.frame(m.outputSubhead(), m.viewOutput(), outputKeybar(w))
	case viewRunning:
		return m.frame(m.runningSubhead(), m.viewRunning(), runningKeybar(w))
	case viewDownloading:
		return m.frame(crumbLine("Download"), m.viewDownload(), "")
	case viewSettingInput:
		return m.frame(crumbLine("Settings", m.settingLabel), m.viewSettingInput(), settingKeybar(w))
	case viewPreview:
		return m.frame(crumbLine("Commands", "Preview"), m.viewPreview(), previewKeybar(w))
	case viewPrompt:
		return m.frame("", m.viewPrompt(), promptKeybar(w))
	default:
		menu := m.currentMenu()
		return m.frame(m.menuSubhead(menu), m.viewMenu(), menuKeybar(w))
	}
}

// ── breadcrumbs ─────────────────────────────────────────────────────────────

// crumbLine renders "Home › a › b" with the last crumb highlighted.
func crumbLine(tail ...string) string {
	crumbs := append([]string{"Home"}, tail...)
	parts := make([]string, len(crumbs))
	for i, c := range crumbs {
		if i == len(crumbs)-1 {
			parts[i] = breadcrumbActive.Render(c)
		} else {
			parts[i] = breadcrumbStyle.Render(c)
		}
	}
	return strings.Join(parts, dimStyle.Render(" › "))
}

func (m model) breadcrumb() string {
	switch m.state {
	case viewCategory:
		return crumbLine("Commands")
	case viewAssetOps:
		return crumbLine("Commands", "Asset Ops")
	case viewZen:
		return crumbLine("Commands", "Zen/IoStore")
	case viewPak:
		return crumbLine("Commands", "PAK Ops")
	case viewJson:
		return crumbLine("Commands", "JSON")
	case viewNiagara:
		return crumbLine("Commands", "Niagara")
	case viewSettings:
		return crumbLine("Settings")
	}
	return ""
}

// menuSubhead is the breadcrumb on sub-menus and the status subtitle at home.
func (m model) menuSubhead(menu menuDef) string {
	if bc := m.breadcrumb(); bc != "" {
		return bc
	}
	if menu.subtitle != "" {
		return subtitleStyle.Render(menu.subtitle)
	}
	return ""
}

func (m model) formSubhead() string {
	if m.form == nil {
		return ""
	}
	return crumbLine("Commands", m.form.command)
}

func (m model) runningSubhead() string {
	return crumbLine("Commands", "Running")
}

func (m model) outputSubhead() string {
	if m.outputErr {
		return crumbLine("Commands", accentRed.Render("Failed"))
	}
	return crumbLine("Commands", accentGreen.Render("Result"))
}

// ── key hints ───────────────────────────────────────────────────────────────

func keyHint(key, desc string) string {
	return keyHintStyle.Render(key) + " " + keyDescStyle.Render(desc)
}

// keybar joins hints with a separator, dropping trailing ones that do not fit
// the width. Pass them in descending order of importance: on a narrow terminal
// the tail is what disappears.
func keybar(width int, hints ...string) string {
	sep := keyDescStyle.Render("  ·  ")
	sepW := lipgloss.Width(sep)

	var b strings.Builder
	used := 0
	for i, h := range hints {
		w := lipgloss.Width(h)
		if i > 0 {
			w += sepW
		}
		if used+w > width {
			break
		}
		if i > 0 {
			b.WriteString(sep)
		}
		b.WriteString(h)
		used += w
	}
	return b.String()
}

func menuKeybar(w int) string {
	return keybar(w,
		keyHint("↑↓", "navigate"),
		keyHint("enter", "select"),
		keyHint("esc", "back"),
		keyHint("q", "quit"),
	)
}

func formKeybar(w int) string {
	return keybar(w,
		keyHint("tab", "next"),
		keyHint("↑↓", "field"),
		keyHint("enter", "submit"),
		keyHint("esc", "cancel"),
		keyHint("ctrl+←→", "word"),
		keyHint("ctrl+⌫/del", "delete word"),
		keyHint("ctrl+c", "copy"),
	)
}

func outputKeybar(w int) string {
	return keybar(w,
		keyHint("↑↓/pgup/pgdn", "scroll"),
		keyHint("home/end", "jump"),
		keyHint("ctrl+c", "copy"),
		keyHint("esc", "back"),
	)
}

func runningKeybar(w int) string {
	return keybar(w,
		keyHint("↑↓/pgup/pgdn", "scroll"),
		keyHint("end", "follow"),
		keyHint("ctrl+c", "copy"),
		keyHint("ctrl+x", "stop"),
	)
}

func previewKeybar(w int) string {
	return keybar(w,
		keyHint("←→/tab", "choose"),
		keyHint("Y/enter", "run"),
		keyHint("N/esc", "cancel"),
	)
}

func promptKeybar(w int) string {
	return keybar(w,
		keyHint("←→/tab", "choose"),
		keyHint("Y/enter", "confirm"),
		keyHint("N/esc", "skip"),
	)
}

func settingKeybar(w int) string {
	return keybar(w,
		keyHint("enter", "save"),
		keyHint("esc", "cancel"),
		keyHint("ctrl+←→", "word"),
		keyHint("ctrl+⌫/del", "delete word"),
		keyHint("ctrl+c", "copy"),
	)
}

// ── progress bar ────────────────────────────────────────────────────────────

// eighths are the horizontal partial-block glyphs, so the bar advances a
// fraction of a cell at a time instead of jumping a whole one.
var eighths = []string{"", "▏", "▎", "▍", "▌", "▋", "▊", "▉"}

// renderProgressBar draws a gradient bar with sub-cell precision.
func renderProgressBar(pct float64, width int) string {
	if width < 10 {
		width = 40
	}
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}

	startHex, endHex := gradientStops()
	start, _ := colorful.Hex(startHex)
	end, _ := colorful.Hex(endHex)

	shade := func(i int) lipgloss.Style {
		t := 0.0
		if width > 1 {
			t = float64(i) / float64(width-1)
		}
		return lipgloss.NewStyle().Foreground(lipgloss.Color(start.BlendLuv(end, t).Hex()))
	}

	exact := pct * float64(width)
	full := int(exact)
	if full > width {
		full = width
	}
	frac := exact - float64(full)

	var b strings.Builder
	for i := 0; i < full; i++ {
		b.WriteString(shade(i).Render("█"))
	}

	rest := width - full
	if rest > 0 {
		if idx := int(frac * 8); idx > 0 {
			b.WriteString(shade(full).Render(eighths[idx]))
			rest--
		}
		if rest > 0 {
			b.WriteString(progressBarEmpty.Render(strings.Repeat("░", rest)))
		}
	}
	return b.String()
}

// ── text helpers ────────────────────────────────────────────────────────────

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

// ── scrollable pane ─────────────────────────────────────────────────────────
//
// The viewport owns scrolling and wrapping; this only draws the frame and the
// scrollbar gutter around whatever it hands back.

// paneContentWidth is the usable text width inside a pane of the given outer
// width: two borders, two pad columns, a gap and the scrollbar column.
func paneContentWidth(outerWidth int) int {
	w := outerWidth - 6
	if w < 10 {
		w = 10
	}
	return w
}

func scrollbarCell(row, height, total, offset int) string {
	if height <= 0 || total <= height {
		return " "
	}
	thumbStart, thumbSize := thumbMetrics(total, height, offset)
	if row >= thumbStart && row < thumbStart+thumbSize {
		return accentCyan.Render("█")
	}
	return dimStyle.Render("│")
}

func renderPane(vp viewport.Model, outerWidth int, borderStyle lipgloss.Style) string {
	inner := outerWidth - 2
	if inner < 4 {
		inner = 4
	}

	lines := strings.Split(vp.View(), "\n")
	total := vp.TotalLineCount()

	rule := strings.Repeat("─", inner)
	var b strings.Builder
	b.WriteString(borderStyle.Render("╭" + rule + "╮"))
	b.WriteString("\n")
	for i, line := range lines {
		row := " " + padLineRight(line, vp.Width) + " " +
			scrollbarCell(i, vp.Height, total, vp.YOffset) + " "
		b.WriteString(borderStyle.Render("│"))
		b.WriteString(padLineRight(row, inner))
		b.WriteString(borderStyle.Render("│"))
		b.WriteString("\n")
	}
	b.WriteString(borderStyle.Render("╰" + rule + "╯"))
	return b.String()
}

// ── view: menu ──────────────────────────────────────────────────────────────

// menuPaneWidths splits the body into a list column and a detail column. The
// detail column is dropped on narrow terminals and on menus without
// descriptions.
func (m model) menuPaneWidths(menu menuDef) (listW, detailW int) {
	w := m.contentWidth()
	if w < 78 || !menuHasDesc(menu) {
		return w, 0
	}
	detailW = w * 2 / 5
	if detailW > 46 {
		detailW = 46
	}
	if detailW < 30 {
		detailW = 30
	}
	return w - detailW - 2, detailW
}

func menuHasDesc(menu menuDef) bool {
	for _, it := range menu.items {
		if it.desc != "" {
			return true
		}
	}
	return false
}

// menuListTop is the screen row of the first visible menu item, used by the
// mouse handler. The list starts one blank row into the body.
func (m model) menuListTop() int {
	return m.bodyTop() + 1
}

// menuWindow is the slice of items that fits the body, scrolled to keep the
// cursor visible. Long menus would otherwise be clipped without a trace.
func (m model) menuWindow(n int) (start, count int) {
	fit := m.bodyHeight() - 2
	if fit < 1 {
		fit = 1
	}
	if fit >= n {
		return 0, n
	}
	start = m.cursor - fit/2
	if start < 0 {
		start = 0
	}
	if start+fit > n {
		start = n - fit
	}
	return start, fit
}

func (m model) viewMenu() string {
	menu := m.currentMenu()
	listW, detailW := m.menuPaneWidths(menu)

	list := "\n" + m.renderMenuList(menu, listW, detailW > 0)
	if detailW == 0 {
		return list
	}

	detail := "\n" + m.renderMenuDetail(menu, detailW, m.bodyHeight()-1)
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(listW).MaxWidth(listW).Render(list),
		lipgloss.NewStyle().Width(2).Render(""),
		detail,
	)
}

func (m model) renderMenuList(menu menuDef, width int, twoPane bool) string {
	start, count := m.menuWindow(len(menu.items))

	var b strings.Builder
	for i := start; i < start+count; i++ {
		item := menu.items[i]
		icon := item.icon
		labelW := width - 4 - lipgloss.Width(icon)

		var row string
		if i == m.cursor {
			bar := lipgloss.NewStyle().
				Foreground(item.color.GetForeground()).
				Background(colorHighlight).
				Render("▌")
			iconSeg := lipgloss.NewStyle().
				Foreground(item.color.GetForeground()).
				Background(colorHighlight).
				Render(icon)
			row = bar + rowBG.Render(" ") + iconSeg +
				rowSelLabel.Render(" "+smartTruncateMiddle(item.label, labelW))
			if !twoPane && item.desc != "" {
				if space := width - lipgloss.Width(row) - 4; space > 14 {
					row += rowSelDesc.Render("  " + smartTruncateMiddle(item.desc, space))
				}
			}
			row = padBG(row, width, rowBG)
		} else {
			row = "  " + item.color.Render(icon) + " " +
				itemNormal.Render(smartTruncateMiddle(item.label, labelW))
		}
		b.WriteString(row)
		b.WriteString("\n")
	}

	if count < len(menu.items) {
		b.WriteString(dimStyle.Render(fmt.Sprintf("  %d of %d",
			m.cursor+1, len(menu.items))))
	}
	return b.String()
}

// renderMenuDetail draws the right-hand pane: what the highlighted entry does
// and, when it maps to a command, the arguments it will ask for.
func (m model) renderMenuDetail(menu menuDef, width, height int) string {
	if m.cursor < 0 || m.cursor >= len(menu.items) {
		return ""
	}
	item := menu.items[m.cursor]

	cw := width - 4
	if cw < 8 {
		cw = 8
	}
	blank := paneBG.Render(strings.Repeat(" ", cw))

	var lines []string
	addWrapped := func(text string, style lipgloss.Style, indent string) {
		if text == "" {
			return
		}
		w := cw - lipgloss.Width(indent)
		if w < 4 {
			w = 4
		}
		for i, l := range strings.Split(style.Width(w).Render(text), "\n") {
			prefix := indent
			if i > 0 {
				prefix = strings.Repeat(" ", lipgloss.Width(indent))
			}
			lines = append(lines, padBG(paneBG.Render(prefix)+l, cw, paneBG))
		}
	}

	addWrapped(strings.TrimSpace(item.label), paneTitle, "")
	lines = append(lines, blank)
	addWrapped(item.desc, paneDim, "")

	if path := menuPathForState(m.state); path != "" {
		if form := getFormForCommand(path, m.cursor); form != nil {
			lines = append(lines, blank)
			addWrapped("UAssetTool "+form.command, paneLabel, "")
			lines = append(lines, blank)
			for _, f := range form.fields {
				label := f.label
				if f.optional {
					label += " (optional)"
				}
				addWrapped(label, paneDim, paneText.Render("• "))
			}
		}
	}

	// Hug the content: a card stretched to the full body height is mostly
	// empty space.
	if h := len(lines) + 2; h < height {
		height = h
	}
	return renderCard(lines, width, height, paneBorder)
}

// renderCard frames pre-styled, card-background lines so the fill runs edge to
// edge instead of stopping where the text does.
func renderCard(lines []string, width, height int, borderStyle lipgloss.Style) string {
	inner := width - 2
	if inner < 2 {
		inner = 2
	}
	rows := height - 2
	if rows < 1 {
		rows = 1
	}

	rule := strings.Repeat("─", inner)
	var b strings.Builder
	b.WriteString(borderStyle.Render("╭" + rule + "╮"))
	b.WriteString("\n")
	for i := 0; i < rows; i++ {
		content := ""
		if i < len(lines) {
			content = lines[i]
		}
		row := paneBG.Render(" ") + padBG(content, inner-1, paneBG)
		b.WriteString(borderStyle.Render("│"))
		b.WriteString(row)
		b.WriteString(borderStyle.Render("│"))
		b.WriteString("\n")
	}
	b.WriteString(borderStyle.Render("╰" + rule + "╯"))
	return b.String()
}

func menuPathForState(s viewState) string {
	switch s {
	case viewAssetOps:
		return "asset"
	case viewZen:
		return "zen"
	case viewPak:
		return "pak"
	case viewJson:
		return "json"
	case viewNiagara:
		return "niagara"
	}
	return ""
}

// ── view: form ──────────────────────────────────────────────────────────────

const formRowsPerField = 4 // top border, input, bottom border, gap

// formWindow is the slice of fields that fits the body, scrolled to keep the
// focused field visible.
func (m model) formWindow() (start, count int) {
	n := len(m.formInputs)
	fit := (m.bodyHeight() - 2) / formRowsPerField
	if fit < 1 {
		fit = 1
	}
	if fit >= n {
		return 0, n
	}
	start = m.formCursor - fit/2
	if start < 0 {
		start = 0
	}
	if start+fit > n {
		start = n - fit
	}
	return start, fit
}

func (m model) viewForm() string {
	if m.form == nil {
		return ""
	}
	w := m.contentWidth()
	start, count := m.formWindow()

	var b strings.Builder
	b.WriteString("\n")
	for i := start; i < start+count && i < len(m.formInputs); i++ {
		f := m.form.fields[i]
		label := f.label
		if f.boolToggle {
			label += " [Y/N]"
		}
		if f.optional {
			label += " (optional)"
		}
		b.WriteString(renderField(label, m.formInputs[i].View(), w, i == m.formCursor))
		b.WriteString("\n")
	}

	if count < len(m.formInputs) {
		b.WriteString(dimStyle.Render(fmt.Sprintf("  field %d of %d",
			m.formCursor+1, len(m.formInputs))))
	}
	return b.String()
}

// renderField draws one input in a rounded box with its label inlaid in the
// top border, so the focused field is obvious without costing an extra row.
func renderField(label, content string, width int, focused bool) string {
	boxStyle, labelStyle := fieldBlurred, fieldLabelBlurred
	if focused {
		boxStyle, labelStyle = fieldFocused, fieldLabelFocused
	}
	border := lipgloss.NewStyle().Foreground(boxStyle.GetBorderTopForeground())

	inner := width - 2
	if inner < 6 {
		inner = 6
	}
	tag := " " + label + " "
	if lipgloss.Width(tag) > inner-2 {
		tag = " " + smartTruncateMiddle(label, inner-4) + " "
	}
	dashes := inner - 1 - lipgloss.Width(tag)
	if dashes < 0 {
		dashes = 0
	}

	var b strings.Builder
	b.WriteString(border.Render("╭─") + labelStyle.Render(tag) +
		border.Render(strings.Repeat("─", dashes)+"╮"))
	b.WriteString("\n")
	b.WriteString(border.Render("│") + " " + padLineRight(content, inner-2) + " " +
		border.Render("│"))
	b.WriteString("\n")
	b.WriteString(border.Render("╰" + strings.Repeat("─", inner) + "╯"))
	b.WriteString("\n")
	return b.String()
}

// fieldInputWidth is the value width a textinput gets inside renderField.
func fieldInputWidth(outerWidth int) int {
	w := outerWidth - 8
	if w < 10 {
		w = 10
	}
	return w
}

// ── view: preview ───────────────────────────────────────────────────────────

// previewButtons are the labels on the preview confirm row; the mouse
// handler needs their rendered widths to hit-test.
var previewButtons = [2]string{"Run", "Cancel"}

// previewButtonHit maps a click on the confirm row to a button index, or -1.
func (m model) previewButtonHit(x, y int) int {
	// body rows: blank, title, blank, three card rows, blank, buttons
	if y != m.bodyTop()+7 {
		return -1
	}
	col := m.contentLeft() + 2
	for i, label := range previewButtons {
		w := lipgloss.Width(itemSelected.Render(label))
		if x >= col && x < col+w {
			return i
		}
		col += w + 3
	}
	return -1
}

func (m model) viewPreview() string {
	w := m.contentWidth()
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(accentYellow.Bold(true).Render("Command Preview"))
	b.WriteString("\n\n")

	card := cardBox.Width(w-cardBox.GetHorizontalBorderSize()).MaxWidth(w).Padding(0, 2)
	inner := w - card.GetHorizontalFrameSize()
	prompt := dimStyle.Render("$") + " "
	avail := inner - lipgloss.Width(prompt)
	if avail < 10 {
		avail = 10
	}
	cmdLine := "UAssetTool.exe " + m.previewCommand
	b.WriteString(card.Render(prompt + previewCmdStyle.Render(smartTruncateMiddle(cmdLine, avail))))
	b.WriteString("\n\n")
	b.WriteString(renderChoice(previewButtons[0], previewButtons[1], m.previewCursor))
	b.WriteString("\n")

	return b.String()
}

// renderChoice draws a two-button row with the active one filled.
func renderChoice(a, c string, cursor int) string {
	btn := func(s string, on bool) string {
		if on {
			return itemSelected.Render(s)
		}
		return lipgloss.NewStyle().Foreground(colorSubtle).Padding(0, 1).Render(s)
	}
	return "  " + btn(a, cursor == 0) + "   " + btn(c, cursor != 0)
}

func (m model) viewPrompt() string {
	w := m.contentWidth()
	var b strings.Builder

	title := "Update Available"
	if m.prompt != nil && m.prompt.title != "" {
		title = m.prompt.title
	}
	b.WriteString("\n")
	b.WriteString(accentYellow.Bold(true).Render("⚠  " + title))
	b.WriteString("\n\n")

	var body strings.Builder
	if m.prompt != nil {
		for _, line := range m.prompt.body {
			body.WriteString(line)
			body.WriteString("\n")
		}
	}
	b.WriteString(cardBox.Width(w - cardBox.GetHorizontalBorderSize()).MaxWidth(w).
		Render(strings.TrimSpace(body.String())))
	b.WriteString("\n\n")

	confirmLabel, cancelLabel := "Confirm", "Cancel"
	if m.prompt != nil {
		if m.prompt.confirm != "" {
			confirmLabel = m.prompt.confirm
		}
		if m.prompt.cancel != "" {
			cancelLabel = m.prompt.cancel
		}
	}
	b.WriteString(renderChoice(confirmLabel, cancelLabel, m.promptCursor))
	b.WriteString("\n")

	return b.String()
}

// ── view: running ───────────────────────────────────────────────────────────

func (m model) viewRunning() string {
	w := m.contentWidth()
	var b strings.Builder

	label := accentCyan.Render("Executing UAssetTool…")
	if !m.logFollow {
		label += dimStyle.Render("   scrolled back — press end to follow")
	}
	b.WriteString(m.spin.View() + "  " + label)
	b.WriteString("\n\n")
	b.WriteString(renderPane(m.logVP, w, lipgloss.NewStyle().Foreground(colorBorder)))

	return b.String()
}

// ── view: download ──────────────────────────────────────────────────────────

func (m model) viewDownload() string {
	var b strings.Builder
	p := m.dlProgress

	b.WriteString("\n")

	if p.phase == "" {
		b.WriteString(m.spin.View() + "  " + accentCyan.Render("Fetching release info…"))
		return b.String()
	}

	barWidth := m.contentWidth() - 4
	if barWidth > 64 {
		barWidth = 64
	}
	if barWidth < 20 {
		barWidth = 20
	}

	switch p.phase {
	case "downloading":
		pct := 0.0
		if p.totalBytes > 0 {
			pct = float64(p.bytesDownloaded) / float64(p.totalBytes)
		}

		b.WriteString(m.spin.View() + "  " + accentCyan.Render("Downloading UAssetTool…"))
		b.WriteString("\n\n")
		b.WriteString(renderProgressBar(pct, barWidth))
		b.WriteString("\n\n")

		stats := fmt.Sprintf("%s / %s", formatBytes(p.bytesDownloaded), formatBytes(p.totalBytes))
		if p.totalBytes > 0 {
			stats += fmt.Sprintf("   %d%%", int(pct*100))
		}
		b.WriteString(textStyle.Render(stats))
		b.WriteString("\n\n")
		b.WriteString(dimStyle.Render("speed  ") + accentGreen.Render(formatBytes(int64(p.speed))+"/s"))
		b.WriteString(dimStyle.Render("    eta  ") + accentYellow.Render(formatDuration(p.eta)))

	case "extracting":
		b.WriteString(m.spin.View() + "  " + accentYellow.Render("Extracting UAssetTool.exe…"))
		b.WriteString("\n\n")
		b.WriteString(renderProgressBar(1.0, barWidth))
		b.WriteString("\n\n")
		b.WriteString(accentGreen.Render(formatBytes(p.bytesDownloaded)) + dimStyle.Render(" downloaded"))
	}

	return b.String()
}

// ── view: output ────────────────────────────────────────────────────────────

func (m model) viewOutput() string {
	w := m.contentWidth()
	border := lipgloss.NewStyle().Foreground(colorGreen)
	if m.outputErr {
		border = lipgloss.NewStyle().Foreground(colorRed)
	}
	return "\n" + renderPane(m.outVP, w, border)
}

// releaseHeaderLines renders download release metadata into the same scroll
// pane as the output, so the pane height no longer depends on it.
func (m model) releaseHeaderLines(width int) []string {
	if m.dlInfo == nil || m.outputErr {
		return nil
	}
	out := []string{
		accentCyan.Bold(true).Render("Release  ") + accentGreen.Render(m.dlInfo.TagName),
	}
	if m.dlInfo.Name != "" && m.dlInfo.Name != m.dlInfo.TagName {
		out[0] += dimStyle.Render("  " + m.dlInfo.Name)
	}
	out = append(out, dimStyle.Render("Published  ")+
		accentYellow.Render(m.dlInfo.PublishedAt.Format("Jan 02, 2006 15:04")))

	if m.dlInfo.Body != "" {
		out = append(out, "", dimStyle.Render("Release notes"))
		notes := strings.Split(normalizeBoxText(m.dlInfo.Body), "\n")
		for _, line := range hardWrapLines(notes, width-2) {
			out = append(out, "  "+line)
		}
	}
	return append(out, "", footerRule.Render(strings.Repeat("─", width)), "")
}

// ── view: setting input ─────────────────────────────────────────────────────

func (m model) viewSettingInput() string {
	w := m.contentWidth()
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(renderField(m.settingLabel, m.settingInput.View(), w, true))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  Leave blank to clear this setting."))
	return b.String()
}
