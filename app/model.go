package app

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// ── view state ──────────────────────────────────────────────────────────────

type viewState int

const (
	viewMain viewState = iota
	viewCategory
	viewAssetOps
	viewZen
	viewPak
	viewJson
	viewNiagara
	viewSettings
	viewForm
	viewOutput
	viewRunning
	viewDownloading
	viewSettingInput
	viewPreview
)

// ── model ───────────────────────────────────────────────────────────────────

type model struct {
	state  viewState
	cursor int
	config Config
	width  int
	height int

	// form
	form         *commandForm
	formInputs   []textinput.Model
	formCursor   int
	formMenuPath string

	// output
	output         string
	outputErr      bool
	runningOutput  string
	runningScroll  int
	outputScroll   int
	draggingScroll string
	dragOffsetY    int

	// settings input
	settingKey   string
	settingLabel string
	settingInput textinput.Model

	// status
	status string

	// download progress
	dlProgress downloadProgressMsg
	dlInfo     *ReleaseInfo

	// preview
	previewArgs string

	// animation
	spinFrame int
}

// ── messages ────────────────────────────────────────────────────────────────

type toolDoneMsg struct {
	output string
	err    error
}

type toolOutputMsg struct {
	chunk string
}

type toolStopMsg struct {
	err error
}

type spinTickMsg struct{}

func spinTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return spinTickMsg{}
	})
}

// ── init ────────────────────────────────────────────────────────────────────

func initialModel() model {
	return model{
		state:  viewMain,
		config: loadConfig(),
		width:  80,
		height: 30,
	}
}

func (m model) Init() tea.Cmd {
	return tea.SetWindowTitle("UAssetTool Manager")
}

// ── helpers ─────────────────────────────────────────────────────────────────

func (m model) currentMenu() menuDef {
	switch m.state {
	case viewMain:
		return mainMenuDef(m.config.ToolVersion)
	case viewCategory:
		return categoryMenu
	case viewAssetOps:
		return assetOpsMenu
	case viewZen:
		return zenMenu
	case viewPak:
		return pakMenu
	case viewJson:
		return jsonMenu
	case viewNiagara:
		return niagaraMenu
	case viewSettings:
		return settingsMenuDef(m.config)
	default:
		return mainMenuDef(m.config.ToolVersion)
	}
}

func (m model) getConfigVal(key string) string {
	switch key {
	case "GamePaksDir":
		return m.config.GamePaksDir
	case "UsmapPath":
		return m.config.UsmapPath
	case "AesKey":
		return m.config.AesKey
	case "OutputExtractionDir":
		return m.config.OutputExtractionDir
	}
	return ""
}
