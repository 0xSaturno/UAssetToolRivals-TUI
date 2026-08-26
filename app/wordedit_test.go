package app

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const testPath = "D:/Games/MarvelRivals/Content/Paks"

func newTestInput(value string) textinput.Model {
	ti := textinput.New()
	ti.KeyMap = inputKeyMap()
	ti.Focus()
	ti.SetValue(value)
	return ti
}

// TestWordDeleteWalksPathSegments is the behaviour that motivated the custom
// word ops: bubbles breaks on whitespace only, so on a path its ctrl+backspace
// clears the entire field.
func TestWordDeleteWalksPathSegments(t *testing.T) {
	ti := newTestInput(testPath)
	for _, want := range []string{
		"D:/Games/MarvelRivals/Content/",
		"D:/Games/MarvelRivals/",
		"D:/Games/",
		"D:/",
		"",
		"",
	} {
		if !handleWordKey(&ti, "ctrl+backspace") {
			t.Fatal("ctrl+backspace was not consumed")
		}
		if got := ti.Value(); got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	}
}

func TestWordDeleteForward(t *testing.T) {
	ti := newTestInput(testPath)
	ti.SetCursor(0)
	if !handleWordKey(&ti, "ctrl+delete") {
		t.Fatal("ctrl+delete was not consumed")
	}
	// "D" plus the ":" that is not a separator, then "/" is taken with it
	if got, want := ti.Value(), "Games/MarvelRivals/Content/Paks"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWordMotion(t *testing.T) {
	ti := newTestInput(testPath)

	handleWordKey(&ti, "ctrl+left")
	if got, want := ti.Position(), len("D:/Games/MarvelRivals/Content/"); got != want {
		t.Errorf("ctrl+left: cursor at %d, want %d", got, want)
	}
	handleWordKey(&ti, "ctrl+left")
	if got, want := ti.Position(), len("D:/Games/MarvelRivals/"); got != want {
		t.Errorf("ctrl+left twice: cursor at %d, want %d", got, want)
	}
	handleWordKey(&ti, "ctrl+right")
	if got, want := ti.Position(), len("D:/Games/MarvelRivals/Content/"); got != want {
		t.Errorf("ctrl+right: cursor at %d, want %d", got, want)
	}
}

// TestWordKeysAtBoundaries makes sure the ops are inert rather than panicking
// at either end of the value, including on an empty field.
func TestWordKeysAtBoundaries(t *testing.T) {
	for _, value := range []string{"", "a", testPath} {
		for _, pos := range []int{0, len(value)} {
			for _, k := range []string{"ctrl+left", "ctrl+right", "ctrl+backspace", "ctrl+delete"} {
				ti := newTestInput(value)
				ti.SetCursor(pos)
				if !handleWordKey(&ti, k) {
					t.Fatalf("%q at %d: %s was not consumed", value, pos, k)
				}
			}
		}
	}
}

func TestHandleWordKeyIgnoresOtherKeys(t *testing.T) {
	for _, k := range []string{"a", "enter", "esc", "tab", "backspace", "delete", "left", "right", "ctrl+c"} {
		ti := newTestInput(testPath)
		if handleWordKey(&ti, k) {
			t.Errorf("%s was consumed as a word key", k)
		}
		if ti.Value() != testPath {
			t.Errorf("%s modified the value", k)
		}
	}
}

// TestBubblesWordBindingsAreOff guards against both implementations running:
// bubbles' whitespace-only version must never fire.
func TestBubblesWordBindingsAreOff(t *testing.T) {
	km := inputKeyMap()
	for name, b := range map[string]bool{
		"WordForward":        km.WordForward.Enabled(),
		"WordBackward":       km.WordBackward.Enabled(),
		"DeleteWordForward":  km.DeleteWordForward.Enabled(),
		"DeleteWordBackward": km.DeleteWordBackward.Enabled(),
	} {
		if b {
			t.Errorf("%s is still enabled in the shared keymap", name)
		}
	}

	ti := newTestInput(testPath)
	ti, _ = ti.Update(tea.KeyMsg{Type: tea.KeyBackspace, Alt: true})
	if ti.Value() == "" {
		t.Error("alt+backspace reached bubbles and cleared the whole path")
	}
}

// TestInputsUseTheSharedKeyMap catches a new input being created without it.
func TestInputsUseTheSharedKeyMap(t *testing.T) {
	m := testModel(120, 40)
	mm, _ := m.openForm("zen", 1)
	m = mm.(model)
	if len(m.formInputs) == 0 {
		t.Fatal("form has no inputs")
	}
	for i, ti := range m.formInputs {
		if ti.KeyMap.DeleteWordBackward.Enabled() {
			t.Errorf("form input %d is not using the shared keymap", i)
		}
	}

	mm, _ = m.openSettingInput("GamePaksDir", "Game Paks Dir")
	m = mm.(model)
	if m.settingInput.KeyMap.DeleteWordBackward.Enabled() {
		t.Error("setting input is not using the shared keymap")
	}
}

// TestCtrlWordKeyLeavesOrdinaryKeysAlone guards the Windows ctrl-recovery from
// hijacking keys the input needs verbatim. On other platforms it is a no-op.
func TestCtrlWordKeyLeavesOrdinaryKeysAlone(t *testing.T) {
	for _, msg := range []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'a'}},
		{Type: tea.KeyEnter},
		{Type: tea.KeyEsc},
		{Type: tea.KeyTab},
		{Type: tea.KeyLeft},
		{Type: tea.KeyBackspace, Alt: true},
		{Type: tea.KeyDelete, Alt: true},
	} {
		if got := ctrlWordKey(msg); got != "" {
			t.Errorf("%s was rewritten to %q", msg.String(), got)
		}
	}
}
