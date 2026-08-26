package app

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func allMenus() []menuDef {
	return []menuDef{
		mainMenuDef("1.3.3", ""),
		categoryMenu,
		assetOpsMenu,
		zenMenu,
		pakMenu,
		jsonMenu,
		niagaraMenu,
		settingsMenuDef(Config{}),
		settingsMenuDef(Config{PreviewCommand: true, EnableAdvancedExtractIoStoreArgs: true}),
	}
}

// safeIconRune reports whether r is a text symbol that terminals render in one
// cell. The blocks below are all Emoji=No, so no font substitutes a
// double-width color glyph for them.
//
// The excluded ranges are where the trouble lives: Miscellaneous Symbols
// (U+2600–U+26FF) and Miscellaneous Symbols and Arrows (U+2B00–U+2BFF) hold
// characters like ⬇ that go-runewidth measures as one cell while terminals
// draw them as two, and everything from U+1F000 up is pictographic.
func safeIconRune(r rune) bool {
	switch {
	case r == ' ':
		return true
	case r < 0x0100: // ASCII and Latin-1
		return true
	case r >= 0x2000 && r <= 0x206F: // General Punctuation
		return true
	case r >= 0x2100 && r <= 0x21FF: // Letterlike Symbols, Arrows
		return true
	case r >= 0x2200 && r <= 0x22FF: // Mathematical Operators
		return true
	case r >= 0x2300 && r <= 0x23FF: // Miscellaneous Technical
		return true
	case r >= 0x2500 && r <= 0x25FF: // Box Drawing, Block Elements, Geometric Shapes
		return true
	case r >= 0x2700 && r <= 0x27BF: // Dingbats
		return true
	}
	return false
}

// TestMenuIconsAreSingleCellText is the guard for the alignment bug that emoji
// icons caused: 🗃 and 👁 measure as one cell but render as two, so every row
// carrying one sat a column off from its neighbours.
func TestMenuIconsAreSingleCellText(t *testing.T) {
	for _, menu := range allMenus() {
		for _, item := range menu.items {
			runes := []rune(item.icon)
			if len(runes) != 1 {
				t.Errorf("%q: icon %q is %d runes, want exactly 1",
					item.label, item.icon, len(runes))
				continue
			}
			if w := lipgloss.Width(item.icon); w != 1 {
				t.Errorf("%q: icon %q measures %d cells, want 1", item.label, item.icon, w)
			}
			if !safeIconRune(runes[0]) {
				t.Errorf("%q: icon %q (U+%04X) is outside the safe text-symbol blocks; "+
					"terminals may render it double-width", item.label, item.icon, runes[0])
			}
		}
	}
}

// TestMenuTitlesAndLabelsAreText catches emoji sneaking into the strings that
// share a line with an icon.
func TestMenuTitlesAndLabelsAreText(t *testing.T) {
	check := func(what, s string) {
		for _, r := range s {
			if r >= 0x1F000 || (r >= 0x2600 && r <= 0x26FF) || (r >= 0x2B00 && r <= 0x2BFF) {
				t.Errorf("%s contains emoji %q (U+%04X): %q", what, string(r), r, s)
			}
		}
	}
	for _, menu := range allMenus() {
		check("menu title", menu.title)
		check("menu subtitle", menu.subtitle)
		for _, item := range menu.items {
			check("item label", item.label)
			check("item desc", item.desc)
		}
	}
}

// TestToggleIconTracksState covers the settings toggles showing their state in
// the icon column as well as the label.
func TestToggleIconTracksState(t *testing.T) {
	on := settingsMenuDef(Config{PreviewCommand: true})
	off := settingsMenuDef(Config{PreviewCommand: false})
	if on.items[4].icon == off.items[4].icon {
		t.Errorf("Command Preview shows %q whether it is on or off", on.items[4].icon)
	}
}
