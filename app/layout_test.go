package app

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// checkFrame asserts the two invariants every view depends on: it fills the
// alt screen exactly, and no row runs past the terminal width. Both are easy
// to break by adding a line to a view, and neither is visible until the
// terminal starts scrolling or wrapping.
func checkFrame(t *testing.T, name string, m model) {
	t.Helper()
	lines := strings.Split(m.View(), "\n")
	if len(lines) != m.height {
		t.Errorf("%s: rendered %d rows, want %d", name, len(lines), m.height)
	}
	for i, l := range lines {
		if w := lipgloss.Width(l); w > m.width {
			t.Errorf("%s: row %d is %d cols, want <= %d", name, i, w, m.width)
		}
	}
}

func testModel(w, h int) model {
	m := initialModel()
	m.width, m.height = w, h
	m.uatVersion = "1.3.3"
	m.resizePanes()
	return m
}

func TestEveryViewFillsTheFrame(t *testing.T) {
	for _, size := range [][2]int{{120, 40}, {80, 24}, {64, 20}, {200, 60}} {
		base := testModel(size[0], size[1])

		for _, st := range []viewState{
			viewMain, viewCategory, viewAssetOps, viewZen,
			viewPak, viewJson, viewNiagara, viewSettings,
		} {
			m := base
			m.state = st
			m.cursor = len(m.currentMenu().items) - 1
			checkFrame(t, "menu", m)
		}

		m := base
		mm, _ := m.openForm("zen", 1)
		m = mm.(model)
		m.formCursor = len(m.formInputs) - 1
		checkFrame(t, "form", m)

		m = base
		m.state = viewPreview
		m.previewCommand = strings.Repeat("create_iostore --input D:/mod ", 8)
		checkFrame(t, "preview", m)

		m = base
		m.state = viewRunning
		m.runningOutput = strings.Repeat("scanning D:/a/very/long/path/to/a/file.uasset\n", 80)
		m.setLogContent()
		checkFrame(t, "running", m)

		m = base
		m.state = viewOutput
		m.output = strings.Repeat("wrote chunk\n", 80)
		m.setOutputContent()
		checkFrame(t, "output", m)

		m = base
		m.state = viewDownloading
		m.dlProgress = downloadProgressMsg{
			phase: "downloading", bytesDownloaded: 41e6, totalBytes: 73e6,
			speed: 3.5e6, eta: 9 * time.Second,
		}
		checkFrame(t, "download", m)

		m = base
		m.state = viewPrompt
		m.prompt = &updatePromptSpec{
			title:   "UAssetTool update available",
			body:    []string{"Installed: v1.3.2", "Latest: v1.3.3"},
			confirm: "Update", cancel: "Skip",
		}
		checkFrame(t, "prompt", m)

		m = base
		mm, _ = m.openSettingInput("GamePaksDir", "Game Paks Dir")
		m = mm.(model)
		checkFrame(t, "setting-input", m)
	}
}

// TestMenuWindowKeepsCursorVisible guards the windowing that keeps long menus
// from being silently clipped on short terminals.
func TestMenuWindowKeepsCursorVisible(t *testing.T) {
	m := testModel(120, 16)
	m.state = viewZen
	n := len(m.currentMenu().items)

	for m.cursor = 0; m.cursor < n; m.cursor++ {
		start, count := m.menuWindow(n)
		if m.cursor < start || m.cursor >= start+count {
			t.Fatalf("cursor %d outside window [%d,%d)", m.cursor, start, start+count)
		}
		if start < 0 || start+count > n {
			t.Fatalf("window [%d,%d) outside 0..%d", start, start+count, n)
		}
	}
}

func TestProgressBarKeepsItsWidth(t *testing.T) {
	for _, pct := range []float64{-1, 0, 0.013, 0.5, 0.999, 1, 2} {
		if w := lipgloss.Width(renderProgressBar(pct, 40)); w != 40 {
			t.Errorf("pct %v: bar is %d cols, want 40", pct, w)
		}
	}
}
