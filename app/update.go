package app

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func scrollBy(scroll *int, delta int) {
	if *scroll < 0 {
		*scroll = 0
	}
	*scroll += delta
	if *scroll < 0 {
		*scroll = 0
	}
}

func (m model) runningBoxRect() (left, top, width, visibleLines int) {
	left = 0
	top = 5
	width = m.width - 8
	if width < 30 {
		width = 30
	}
	if width > 120 {
		width = 120
	}
	visibleLines = m.height - 12
	if visibleLines < 6 {
		visibleLines = 6
	}
	return
}

func (m model) outputBoxRect() (left, top, width, visibleLines int) {
	left = 0
	top = 3
	if m.dlInfo != nil && !m.outputErr {
		top += 1
		bodyLines := 0
		if m.dlInfo.Body != "" {
			bodyLines = len(strings.Split(m.dlInfo.Body, "\n"))
			if bodyLines > 12 {
				bodyLines = 13
			}
			bodyLines += 2
		}
		top += 3 + bodyLines
	}
	width = m.width - 6
	if width < 30 {
		width = 30
	}
	if width > 140 {
		width = 140
	}
	visibleLines = m.height - 14
	if visibleLines < 8 {
		visibleLines = 8
	}
	return
}

func mouseScrollTarget(totalLines, visibleLines, row int) int {
	if totalLines <= visibleLines {
		return 0
	}
	if row < 0 {
		row = 0
	}
	if row >= visibleLines {
		row = visibleLines - 1
	}
	maxScroll := totalLines - visibleLines
	if visibleLines <= 1 {
		return maxScroll
	}
	return (row * maxScroll) / (visibleLines - 1)
}

func scrollbarThumbMetrics(totalLines, visibleLines, scroll int) (thumbStart, thumbSize int) {
	if visibleLines <= 0 {
		return 0, 0
	}
	if totalLines <= visibleLines {
		return 0, 0
	}
	thumbSize = (visibleLines * visibleLines) / totalLines
	if thumbSize < 1 {
		thumbSize = 1
	}
	if thumbSize > visibleLines {
		thumbSize = visibleLines
	}
	trackRange := visibleLines - thumbSize
	maxScroll := totalLines - visibleLines
	if trackRange <= 0 || maxScroll <= 0 {
		return 0, thumbSize
	}
	scroll = clampScroll(totalLines, visibleLines, scroll)
	thumbStart = (scroll * trackRange) / maxScroll
	return thumbStart, thumbSize
}

func mouseDragScrollTarget(totalLines, visibleLines, thumbSize, thumbRow int) int {
	if totalLines <= visibleLines {
		return 0
	}
	trackRange := visibleLines - thumbSize
	maxScroll := totalLines - visibleLines
	if trackRange <= 0 || maxScroll <= 0 {
		return 0
	}
	if thumbRow < 0 {
		thumbRow = 0
	}
	if thumbRow > trackRange {
		thumbRow = trackRange
	}
	return (thumbRow * maxScroll) / trackRange
}

func (m model) normalizedRunningLines() []string {
	logText := normalizeBoxText(strings.TrimRight(m.runningOutput, "\n"))
	logLines := strings.Split(logText, "\n")
	if strings.TrimSpace(m.runningOutput) == "" {
		return []string{dimStyle.Render("Waiting for UAssetTool output...")}
	}
	_, _, width, _ := m.runningBoxRect()
	contentWidth := width - cardBox.GetHorizontalFrameSize()
	if contentWidth < 10 {
		contentWidth = 10
	}
	for i, line := range logLines {
		logLines[i] = smartTruncateMiddle(line, contentWidth)
	}
	return hardWrapLines(logLines, contentWidth)
}

func (m model) normalizedOutputLines() []string {
	out := normalizeBoxText(m.output)
	lines := strings.Split(out, "\n")
	_, _, width, _ := m.outputBoxRect()
	contentWidth := width - successBox.GetHorizontalFrameSize()
	if contentWidth < 10 {
		contentWidth = 10
	}
	return hardWrapLines(lines, contentWidth)
}

func copyTextToClipboard(text string) string {
	text = normalizeBoxText(text)
	if strings.TrimSpace(text) == "" {
		return "Nothing to copy"
	}
	if err := clipboard.WriteAll(text); err != nil {
		return "Copy failed: " + err.Error()
	}
	return "Copied to clipboard"
}

func normalizeInputValue(val string) string {
	val = strings.TrimSpace(strings.Trim(val, `"`))
	if val == "" {
		return ""
	}
	val = strings.ReplaceAll(val, `\`, `/`)
	val = strings.ReplaceAll(val, `//`, `/`)
	if strings.Contains(val, ":/") || strings.HasPrefix(val, "/") || strings.HasPrefix(val, "./") || strings.HasPrefix(val, "../") {
		val = filepath.Clean(val)
	}
	return val
}

func quoteArg(val string) string {
	if val == "" {
		return ""
	}
	val = strings.ReplaceAll(val, `"`, `\"`)
	return fmt.Sprintf(`"%s"`, val)
}

func quotePathArg(val string) string {
	return quoteArg(normalizeInputValue(val))
}

func isPathLikeField(label string) bool {
	label = strings.ToLower(label)
	return strings.Contains(label, "path") || strings.Contains(label, "directory") || strings.Contains(label, "file") || strings.Contains(label, "folder")
}

func shouldQuoteField(f formField, command string, index int) bool {
	if isPathLikeField(f.label) {
		return true
	}
	switch command {
	case "create_mod_iostore":
		return index == 0 || index == 1 || index == 2 || index == 3 || index == 6
	case "extract_iostore_legacy":
		return index == 0 || index == 1 || index == 2 || index == 3
	case "extract_pak":
		return index != 4
	case "create_pak":
		return index == 0 || index == 2
	case "create_companion_pak":
		return index == 0
	case "to_json", "from_json", "batch_detect", "dump", "skeletal_mesh_info", "to_zen", "niagara_list", "niagara_details", "modify_colors", "scan_childbp_isenemy", "extract_script_objects", "inspect_zen", "is_iostore_compressed", "is_iostore_encrypted", "recompress_iostore", "detect", "fix":
		return true
	}
	return false
}

func stopToolCmd() tea.Cmd {
	return func() tea.Msg {
		return toolStopMsg{err: stopRunningTool()}
	}
}

// ── update ──────────────────────────────────────────────────────────────────

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinTickMsg:
		if m.state == viewRunning || m.state == viewDownloading {
			m.spinFrame++
			return m, spinTick()
		}
		return m, nil

	case toolOutputMsg:
		m.runningOutput += msg.chunk
		m.runningScroll = -1
		return m, nil

	case toolStopMsg:
		if msg.err != nil {
			m.status = "Stop failed: " + msg.err.Error()
		} else {
			m.status = "Panic stop sent"
		}
		return m, nil

	case toolDoneMsg:
		m.state = viewOutput
		m.output = msg.output
		if m.output == "" {
			m.output = m.runningOutput
		}
		m.outputErr = msg.err != nil
		if msg.err != nil && m.output == "" {
			m.output = msg.err.Error()
		}
		m.outputScroll = -1
		m.runningOutput = ""
		m.runningScroll = 0
		return m, nil

	case downloadProgressMsg:
		m.dlProgress = msg
		return m, nil

	case downloadCompleteMsg:
		m.dlInfo = msg.info
		m.state = viewOutput
		if msg.err != nil {
			m.output = "Download failed: " + msg.err.Error()
			m.outputErr = true
		} else {
			m.output = msg.output
			m.outputErr = false
			if msg.info != nil && msg.info.TagName != "" {
				m.config.ToolVersion = msg.info.TagName
				saveConfig(m.config)
			}
		}
		m.outputScroll = -1
		return m, nil

	case tea.MouseMsg:
		return m.handleMouse(msg)

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	if m.state == viewForm && len(m.formInputs) > 0 {
		var cmd tea.Cmd
		m.formInputs[m.formCursor], cmd = m.formInputs[m.formCursor].Update(msg)
		return m, cmd
	}
	if m.state == viewSettingInput {
		var cmd tea.Cmd
		m.settingInput, cmd = m.settingInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

// ── mouse ───────────────────────────────────────────────────────────────────

func (m model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case viewForm, viewSettingInput, viewDownloading:
		if msg.Action == tea.MouseActionRelease {
			m.draggingScroll = ""
			m.dragOffsetY = 0
		}
		return m, nil
	case viewOutput:
		left, top, width, visibleLines := m.outputBoxRect()
		lines := m.normalizedOutputLines()
		right := left + width - 1
		contentTop := top + 2
		contentBottom := contentTop + visibleLines - 1
		scrollbarX := right - 2
		thumbStart, thumbSize := scrollbarThumbMetrics(len(lines), visibleLines, m.outputScroll)
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			scrollBy(&m.outputScroll, -3)
			return m, nil
		case tea.MouseButtonWheelDown:
			scrollBy(&m.outputScroll, 3)
			return m, nil
		}
		if msg.Action == tea.MouseActionMotion && m.draggingScroll == "output" {
			thumbRow := msg.Y - contentTop - m.dragOffsetY
			m.outputScroll = mouseDragScrollTarget(len(lines), visibleLines, thumbSize, thumbRow)
			return m, nil
		}
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			if msg.X == scrollbarX && msg.Y >= contentTop && msg.Y <= contentBottom {
				row := msg.Y - contentTop
				if row >= thumbStart && row < thumbStart+thumbSize {
					m.draggingScroll = "output"
					m.dragOffsetY = row - thumbStart
				} else {
					m.outputScroll = mouseScrollTarget(len(lines), visibleLines, row)
				}
			}
			return m, nil
		}
		if msg.Action == tea.MouseActionRelease {
			m.draggingScroll = ""
			m.dragOffsetY = 0
		}
		return m, nil
	case viewRunning:
		left, top, width, visibleLines := m.runningBoxRect()
		lines := m.normalizedRunningLines()
		right := left + width - 1
		contentTop := top + 2
		contentBottom := contentTop + visibleLines - 1
		scrollbarX := right - 2
		thumbStart, thumbSize := scrollbarThumbMetrics(len(lines), visibleLines, m.runningScroll)
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			scrollBy(&m.runningScroll, -3)
			return m, nil
		case tea.MouseButtonWheelDown:
			scrollBy(&m.runningScroll, 3)
			return m, nil
		}
		if msg.Action == tea.MouseActionMotion && m.draggingScroll == "running" {
			thumbRow := msg.Y - contentTop - m.dragOffsetY
			m.runningScroll = mouseDragScrollTarget(len(lines), visibleLines, thumbSize, thumbRow)
			return m, nil
		}
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			if msg.X == scrollbarX && msg.Y >= contentTop && msg.Y <= contentBottom {
				row := msg.Y - contentTop
				if row >= thumbStart && row < thumbStart+thumbSize {
					m.draggingScroll = "running"
					m.dragOffsetY = row - thumbStart
				} else {
					m.runningScroll = mouseScrollTarget(len(lines), visibleLines, row)
				}
			}
			return m, nil
		}
		if msg.Action == tea.MouseActionRelease {
			m.draggingScroll = ""
			m.dragOffsetY = 0
		}
		return m, nil
	case viewPreview:
		if msg.Action == tea.MouseActionRelease {
			m.draggingScroll = ""
			m.dragOffsetY = 0
		}
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			if msg.Y >= 8 && msg.Y <= 9 {
				return m.runPreviewedCommand()
			}
			if msg.Y >= 10 {
				m.state = viewCategory
				m.cursor = 0
				return m, nil
			}
		}
		return m, nil
	}

	menu := m.currentMenu()
	headerLines := m.menuHeaderLines()
	idx := msg.Y - headerLines + 1
	if idx < 0 || idx >= len(menu.items) {
		return m, nil
	}

	switch msg.Action {
	case tea.MouseActionPress:
		if msg.Button == tea.MouseButtonLeft {
			m.cursor = idx
			return m.selectCurrentItem()
		}
	case tea.MouseActionMotion:
		m.cursor = idx
	}

	return m, nil
}

// ── keyboard ────────────────────────────────────────────────────────────────

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "ctrl+c":
		switch m.state {
		case viewForm:
			if len(m.formInputs) > 0 && m.formCursor >= 0 && m.formCursor < len(m.formInputs) {
				m.status = copyTextToClipboard(m.formInputs[m.formCursor].Value())
			} else {
				m.status = "Nothing to copy"
			}
		case viewSettingInput:
			m.status = copyTextToClipboard(m.settingInput.Value())
		case viewRunning:
			m.status = copyTextToClipboard(m.runningOutput)
		case viewOutput:
			m.status = copyTextToClipboard(m.output)
		default:
			m.status = "Ctrl+C copy is available in running/output views"
		}
		return m, nil
	}

	switch m.state {
	case viewOutput:
		switch key {
		case "up", "k":
			if m.outputScroll > 0 {
				m.outputScroll--
			}
			return m, nil
		case "down", "j":
			m.outputScroll++
			return m, nil
		case "pgup", "b":
			m.outputScroll -= 8
			if m.outputScroll < 0 {
				m.outputScroll = 0
			}
			return m, nil
		case "pgdown", "f":
			m.outputScroll += 8
			return m, nil
		case "home":
			m.outputScroll = 0
			return m, nil
		case "end":
			m.outputScroll = -1
			return m, nil
		case "enter", "esc", "backspace":
			m.state = viewMain
			m.output = ""
			m.outputScroll = 0
			m.cursor = 0
			m.dlInfo = nil
		}
		return m, nil

	case viewRunning:
		switch key {
		case "ctrl+x", "ctrl+z":
			m.status = "Stopping UAssetTool..."
			return m, stopToolCmd()
		case "up", "k":
			if m.runningScroll > 0 {
				m.runningScroll--
			}
		case "down", "j":
			m.runningScroll++
		case "pgup", "b":
			m.runningScroll -= 8
			if m.runningScroll < 0 {
				m.runningScroll = 0
			}
		case "pgdown", "f":
			m.runningScroll += 8
		case "home":
			m.runningScroll = 0
		case "end":
			m.runningScroll = -1
		}
		return m, nil

	case viewDownloading:
		return m, nil

	case viewPreview:
		return m.handlePreviewKey(msg)

	case viewForm:
		return m.handleFormKey(msg)

	case viewSettingInput:
		return m.handleSettingInputKey(msg)

	default:
		return m.handleMenuKey(msg)
	}
}

func (m model) handleMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	menu := m.currentMenu()
	key := msg.String()

	switch key {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		} else {
			m.cursor = len(menu.items) - 1
		}
	case "down", "j":
		if m.cursor < len(menu.items)-1 {
			m.cursor++
		} else {
			m.cursor = 0
		}
	case "enter", " ":
		return m.selectCurrentItem()
	case "esc", "backspace":
		return m.goBack()
	case "q":
		if m.state == viewMain {
			return m, tea.Quit
		}
		return m.goBack()
	}
	return m, nil
}

func (m model) handlePreviewKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "y", "Y", "enter":
		return m.runPreviewedCommand()
	case "n", "N", "esc":
		m.state = viewCategory
		m.cursor = 0
		return m, nil
	}
	return m, nil
}

func (m model) runPreviewedCommand() (tea.Model, tea.Cmd) {
	args := m.previewArgs
	m.state = viewRunning
	m.status = "Running..."
	m.runningOutput = ""
	m.runningScroll = -1
	return m, tea.Batch(spinTick(), func() tea.Msg {
		out, err := runTool(args)
		return toolDoneMsg{out, err}
	})
}

// ── navigation ──────────────────────────────────────────────────────────────

func (m model) goBack() (tea.Model, tea.Cmd) {
	switch m.state {
	case viewMain:
		return m, tea.Quit
	case viewCategory:
		m.state = viewMain
	case viewAssetOps, viewZen, viewPak, viewJson, viewNiagara:
		m.state = viewCategory
	case viewSettings:
		m.state = viewMain
	default:
		m.state = viewMain
	}
	m.cursor = 0
	return m, nil
}

func (m model) selectCurrentItem() (tea.Model, tea.Cmd) {
	menu := m.currentMenu()
	if m.cursor >= len(menu.items) {
		return m, nil
	}

	switch m.state {
	case viewMain:
		switch m.cursor {
		case 0:
			m.state = viewCategory
			m.cursor = 0
		case 1:
			m.state = viewDownloading
			m.status = "Fetching release info..."
			m.dlProgress = downloadProgressMsg{}
			return m, tea.Batch(spinTick(), downloadToolCmd())
		case 2:
			m.state = viewSettings
			m.cursor = 0
		case 3:
			return m, tea.Quit
		}

	case viewCategory:
		switch m.cursor {
		case 0:
			m.state = viewAssetOps
		case 1:
			m.state = viewZen
		case 2:
			m.state = viewPak
		case 3:
			m.state = viewJson
		case 4:
			m.state = viewNiagara
		case 5:
			m.state = viewMain
		}
		m.cursor = 0

	case viewAssetOps:
		if m.cursor == len(assetOpsMenu.items)-1 {
			return m.goBack()
		}
		return m.openForm("asset", m.cursor)

	case viewZen:
		if m.cursor == len(zenMenu.items)-1 {
			return m.goBack()
		}
		return m.openForm("zen", m.cursor)

	case viewPak:
		if m.cursor == len(pakMenu.items)-1 {
			return m.goBack()
		}
		return m.openForm("pak", m.cursor)

	case viewJson:
		if m.cursor == len(jsonMenu.items)-1 {
			return m.goBack()
		}
		return m.openForm("json", m.cursor)

	case viewNiagara:
		if m.cursor == len(niagaraMenu.items)-1 {
			return m.goBack()
		}
		return m.openForm("niagara", m.cursor)

	case viewSettings:
		if m.cursor == len(settingsMenuDef(m.config).items)-1 {
			return m.goBack()
		}
		return m.handleSettingsSelect()
	}

	return m, nil
}

// ── form handling ───────────────────────────────────────────────────────────

func (m model) openForm(menuPath string, choice int) (tea.Model, tea.Cmd) {
	form := getFormForCommand(menuPath, choice)
	if form == nil {
		return m, nil
	}
	m.state = viewForm
	m.form = form
	m.formMenuPath = menuPath
	m.formCursor = 0
	m.status = ""

	m.formInputs = make([]textinput.Model, len(form.fields))
	for i, f := range form.fields {
		ti := textinput.New()
		ti.Placeholder = f.label
		if f.optional {
			ti.Placeholder += " (optional)"
		}
		if f.boolToggle {
			ti.Placeholder = f.label + " [Y/N]"
			ti.CharLimit = 1
		}
		if f.defaultVal != "" {
			ti.SetValue(f.defaultVal)
		}
		if f.configKey != "" {
			ti.SetValue(m.getConfigVal(f.configKey))
		}
		if i == 0 {
			ti.Focus()
		}
		ti.Width = 60
		ti.PromptStyle = accentCyan
		ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colorText))
		m.formInputs[i] = ti
	}
	return m, m.formInputs[0].Focus()
}

func (m model) handleFormKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc":
		m.state = viewCategory
		m.cursor = 0
		m.status = ""
		return m, nil

	case "tab", "down":
		if m.formCursor < len(m.formInputs)-1 {
			m.formInputs[m.formCursor].Blur()
			m.formCursor++
			return m, m.formInputs[m.formCursor].Focus()
		}

	case "shift+tab", "up":
		if m.formCursor > 0 {
			m.formInputs[m.formCursor].Blur()
			m.formCursor--
			return m, m.formInputs[m.formCursor].Focus()
		}

	case "enter":
		if m.formCursor >= len(m.formInputs)-1 || len(m.formInputs) == 1 {
			return m.submitForm()
		}
		m.formInputs[m.formCursor].Blur()
		m.formCursor++
		return m, m.formInputs[m.formCursor].Focus()

	case "ctrl+enter":
		return m.submitForm()
	}

	var cmd tea.Cmd
	m.formInputs[m.formCursor], cmd = m.formInputs[m.formCursor].Update(msg)
	return m, cmd
}

func (m model) submitForm() (tea.Model, tea.Cmd) {
	for i, f := range m.form.fields {
		if !f.optional && strings.TrimSpace(m.formInputs[i].Value()) == "" {
			m.status = fmt.Sprintf("Required: %s", f.label)
			return m, nil
		}
	}

	args := m.buildArgs()

	if m.config.PreviewCommand {
		m.state = viewPreview
		m.previewArgs = args
		return m, nil
	}

	m.state = viewRunning
	m.status = "Running..."
	m.runningOutput = ""
	return m, tea.Batch(spinTick(), func() tea.Msg {
		out, err := runTool(args)
		return toolDoneMsg{out, err}
	})
}

func (m model) buildArgs() string {
	if m.form == nil {
		return ""
	}
	var parts []string
	parts = append(parts, m.form.command)

	for i, f := range m.form.fields {
		val := normalizeInputValue(m.formInputs[i].Value())
		if val == "" {
			continue
		}
		if f.boolToggle {
			switch m.form.command {
			case "to_zen":
				if strings.EqualFold(val, "y") {
					parts = append(parts, "--no-material-tags")
				}
			case "create_mod_iostore":
				switch i {
				case 4:
					if strings.EqualFold(val, "n") {
						parts = append(parts, "--no-compress")
					} else {
						parts = append(parts, "--compress")
					}
				case 5:
					if strings.EqualFold(val, "y") {
						parts = append(parts, "--obfuscate")
					}
				case 7:
					if strings.EqualFold(val, "y") {
						parts = append(parts, "--no-material-tags")
					}
				}
			case "extract_iostore_legacy":
				if strings.EqualFold(val, "y") {
					parts = append(parts, "--with-deps")
				}
			case "create_pak":
				if strings.EqualFold(val, "y") {
					parts = append(parts, "--compress")
				} else {
					parts = append(parts, "--no-compress")
				}
			case "extract_pak":
				if i == 3 && strings.EqualFold(val, "y") {
					parts = append(parts, "--list")
				}
			case "niagara_details":
				if strings.EqualFold(val, "y") {
					parts = append(parts, "--full")
				}
			case "scan_childbp_isenemy":
				if strings.EqualFold(val, "y") {
					parts = append(parts, "--extracted")
				}
			}
			continue
		}
		switch m.form.command {
		case "detect", "fix":
			if i == 1 {
				parts = append(parts, quotePathArg(val))
				continue
			}
		case "create_mod_iostore":
			if i == 1 {
				parts = append(parts, "--mount-point", quoteArg(val))
				continue
			}
			if i == 2 {
				parts = append(parts, "--game-path", quoteArg(val))
				continue
			}
			if i == 6 {
				parts = append(parts, "--pak-aes", quoteArg(val))
				continue
			}
		case "extract_pak":
			if i == 2 {
				parts = append(parts, "--aes", quoteArg(val))
				continue
			}
			if i == 4 {
				parts = append(parts, "--filter", quoteArg(val))
				continue
			}
		case "extract_iostore_legacy":
			if i == 2 {
				parts = append(parts, "--mod", quotePathArg(val))
				continue
			}
			if i == 3 {
				parts = append(parts, "--filter", quoteArg(val))
				continue
			}
		case "create_pak":
			if i == 2 {
				parts = append(parts, "--mount-point", quoteArg(val))
				continue
			}
		case "scan_childbp_isenemy":
			if i == 1 {
				parts = append(parts, "--aes", quoteArg(val))
				continue
			}
		case "niagara_edit":
			if i == 2 {
				parts = append(parts, val)
				continue
			}
		}
		if shouldQuoteField(f, m.form.command, i) {
			if isPathLikeField(f.label) {
				parts = append(parts, quotePathArg(val))
			} else {
				parts = append(parts, quoteArg(val))
			}
		} else {
			parts = append(parts, val)
		}
	}

	return strings.Join(parts, " ")
}

// ── settings handling ───────────────────────────────────────────────────────

func (m model) handleSettingsSelect() (tea.Model, tea.Cmd) {
	switch m.cursor {
	case 0:
		return m.openSettingInput("GamePaksDir", "Game Paks Directory")
	case 1:
		return m.openSettingInput("UsmapPath", "Default USMAP Path")
	case 2:
		return m.openSettingInput("AesKey", "Default AES Key")
	case 3:
		return m.openSettingInput("OutputExtractionDir", "Output Extraction Directory")
	case 4:
		m.config.PreviewCommand = !m.config.PreviewCommand
		saveConfig(m.config)
		return m, nil
	case 5:
		m.config.EnableAdvancedExtractIoStoreArgs = !m.config.EnableAdvancedExtractIoStoreArgs
		saveConfig(m.config)
		return m, nil
	}
	return m, nil
}

func (m model) openSettingInput(key, label string) (tea.Model, tea.Cmd) {
	m.state = viewSettingInput
	m.settingKey = key
	m.settingLabel = label
	ti := textinput.New()
	ti.Placeholder = label
	ti.SetValue(m.getConfigVal(key))
	ti.Width = 60
	ti.PromptStyle = accentYellow
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colorText))
	ti.Focus()
	m.settingInput = ti
	return m, ti.Focus()
}

func (m model) handleSettingInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "enter":
		val := strings.Trim(m.settingInput.Value(), `"`)
		switch m.settingKey {
		case "GamePaksDir":
			m.config.GamePaksDir = val
		case "UsmapPath":
			m.config.UsmapPath = val
		case "AesKey":
			m.config.AesKey = val
		case "OutputExtractionDir":
			m.config.OutputExtractionDir = val
		}
		saveConfig(m.config)
		m.state = viewSettings
		m.cursor = 0
		return m, nil
	case "esc":
		m.state = viewSettings
		m.cursor = 0
		return m, nil
	}

	var cmd tea.Cmd
	m.settingInput, cmd = m.settingInput.Update(msg)
	return m, cmd
}
