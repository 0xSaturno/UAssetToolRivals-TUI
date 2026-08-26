package app

import (
	"unicode"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
)

// ── word-wise editing in text inputs ────────────────────────────────────────
//
// bubbles breaks words on whitespace only. Nearly every value this tool asks
// for is a path, which has no spaces in it, so its ctrl+backspace wipes the
// whole field in one keystroke. These ops break on path separators too, the
// way an editor's address bar does.

// isWordSep reports whether r ends a word. Path separators, the extension dot
// and the underscore are boundaries, so a path can be walked a segment at a
// time and a name like SK_Character_Body a part at a time; letters, digits and
// dashes stay inside one word.
func isWordSep(r rune) bool {
	return unicode.IsSpace(r) || r == '/' || r == '\\' || r == '.' || r == '_'
}

// wordLeft returns the position one word to the left of pos: back over any
// separators, then back over the word itself.
func wordLeft(v []rune, pos int) int {
	if pos > len(v) {
		pos = len(v)
	}
	for pos > 0 && isWordSep(v[pos-1]) {
		pos--
	}
	for pos > 0 && !isWordSep(v[pos-1]) {
		pos--
	}
	return pos
}

// wordRight returns the position one word to the right of pos: forward over
// the word, then over the separators that follow it, so that deleting forward
// takes the trailing slash with it.
func wordRight(v []rune, pos int) int {
	if pos < 0 {
		pos = 0
	}
	for pos < len(v) && !isWordSep(v[pos]) {
		pos++
	}
	for pos < len(v) && isWordSep(v[pos]) {
		pos++
	}
	return pos
}

// handleWordKey applies word motion and deletion to ti, reporting whether it
// consumed the key. Callers must not pass a consumed key on to the input.
func handleWordKey(ti *textinput.Model, key string) bool {
	v := []rune(ti.Value())
	pos := ti.Position()

	switch key {
	case "ctrl+left", "alt+left", "alt+b":
		ti.SetCursor(wordLeft(v, pos))

	case "ctrl+right", "alt+right", "alt+f":
		ti.SetCursor(wordRight(v, pos))

	case "ctrl+backspace", "alt+backspace", "ctrl+w":
		start := wordLeft(v, pos)
		if start == pos {
			return true
		}
		ti.SetValue(string(v[:start]) + string(v[pos:]))
		ti.SetCursor(start)

	case "ctrl+delete", "alt+delete", "alt+d":
		end := wordRight(v, pos)
		if end == pos {
			return true
		}
		ti.SetValue(string(v[:pos]) + string(v[end:]))
		ti.SetCursor(pos)

	default:
		return false
	}
	return true
}

// inputKeyMap is the keymap every input in the app uses. It hands the word
// bindings over to handleWordKey by disabling the built-in ones, so there is
// exactly one implementation rather than two that disagree on what a word is.
func inputKeyMap() textinput.KeyMap {
	km := textinput.DefaultKeyMap
	km.WordForward = key.NewBinding(key.WithDisabled())
	km.WordBackward = key.NewBinding(key.WithDisabled())
	km.DeleteWordForward = key.NewBinding(key.WithDisabled())
	km.DeleteWordBackward = key.NewBinding(key.WithDisabled())
	return km
}
