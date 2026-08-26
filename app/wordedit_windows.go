//go:build windows

package app

import (
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
)

// The Windows console delivers Backspace and Delete through ReadConsoleInput
// as VK_BACK / VK_DELETE, and bubbletea's console reader builds its KeyMsg
// from the virtual key code alone — the ctrl flag in ControlKeyState never
// reaches us. bubbletea v1 has no key type for ctrl+delete either, so the
// chord cannot be recovered from the message.
//
// Asking the OS whether Ctrl is physically down when we handle the key gets it
// back. The gap between the keypress landing in the console buffer and this
// check is a few milliseconds and a chord means Ctrl is still held, so this is
// reliable in practice. Worst case it misses and the key deletes one
// character, which is what would have happened anyway.

var procGetAsyncKeyState = syscall.NewLazyDLL("user32.dll").NewProc("GetAsyncKeyState")

const vkControl = 0x11

func ctrlHeld() bool {
	if err := procGetAsyncKeyState.Find(); err != nil {
		return false
	}
	state, _, _ := procGetAsyncKeyState.Call(uintptr(vkControl))
	return state&0x8000 != 0
}

// ctrlWordKey reports the word-edit key a bare Backspace or Delete should be
// treated as because Ctrl is held, or "" to leave the key alone.
func ctrlWordKey(msg tea.KeyMsg) string {
	if msg.Alt {
		return ""
	}
	switch msg.Type {
	case tea.KeyBackspace:
		if ctrlHeld() {
			return "ctrl+backspace"
		}
	case tea.KeyDelete:
		if ctrlHeld() {
			return "ctrl+delete"
		}
	}
	return ""
}
