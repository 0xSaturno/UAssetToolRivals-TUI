//go:build !windows

package app

import tea "github.com/charmbracelet/bubbletea"

// ctrlWordKey is a no-op outside Windows: terminals that support the chord
// report it as a distinct key, which handleWordKey already recognises.
func ctrlWordKey(tea.KeyMsg) string { return "" }
