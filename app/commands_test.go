package app

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
)

// sampleFor fills a field the way a user would.
func sampleFor(f formField) string {
	if f.boolToggle {
		return "Y"
	}
	switch f.label {
	case "Mip Index", "Path Hash Seed":
		return "0"
	case "Output Format (PNG/TGA/DDS/BMP)":
		return "png"
	case "Package Path (/Game/...)":
		return "/Game/Marvel/Foo"
	case "Path/String for cityhash":
		return "/Game/Marvel/Foo"
	case "Asset Type Filter (comma sep)":
		return "Texture2D,Material"
	case "Filter Pattern (single)":
		return "SK_1014"
	case "Filter Patterns (space sep)", "Filter Patterns (space sep or .txt file)":
		return "SK_1014 SK_1057"
	case "Mount Point (default ../../../)":
		return "../../../"
	case "Game Path (default Marvel/Content/)":
		return "Marvel/Content/"
	case "Chunk ID (hex)":
		return "AABBCC"
	case "Package Name":
		return "/Game/Marvel/Foo"
	case "Edits JSON":
		return `[{"exportIndex":4}]`
	case "AES Key", "PAK AES Key":
		return "0x0C263D8C22DCB085894899C3A3796383E9BF9DE0CBFB08C9BF2DEF2E84F29D74"
	case "Input Files (space sep)", "File Paths (space sep)":
		return "a/one.uasset b/two.uasset"
	case "Namespace Filter":
		return "601_HeroUIAsset_1011_ST"
	case "Key Lookup (needs namespace)":
		return "HeroInfo_TName"
	case "Search Term":
		return "BRUCE BANNER"
	}
	return "C:/work/thing"
}

func argsFor(t *testing.T, menuPath string, choice int) string {
	t.Helper()
	form := getFormForCommand(menuPath, choice)
	if form == nil {
		t.Fatalf("%s[%d]: no form defined", menuPath, choice)
	}
	m := model{form: form, formInputs: make([]textinput.Model, len(form.fields))}
	for i, f := range form.fields {
		ti := textinput.New()
		ti.SetValue(sampleFor(f))
		m.formInputs[i] = ti
	}
	return buildPreviewCommand(m.buildArgList())
}

// UAssetTool ignores flags it does not know, so a wrong flag fails silently at
// runtime. These assertions are where it shows up instead.
func TestFormsMatchToolCommandLines(t *testing.T) {
	const path = `C:\work\thing`
	const aes = "0C263D8C22DCB085894899C3A3796383E9BF9DE0CBFB08C9BF2DEF2E84F29D74"

	cases := []struct {
		menu   string
		choice int
		want   string
	}{
		{"asset", 0, "detect " + path + " " + path},
		{"asset", 1, "batch_detect " + path + " " + path},
		{"asset", 2, "fix " + path + " " + path},
		{"asset", 3, "dump " + path + " " + path},
		{"asset", 4, "skeletal_mesh_info " + path + " " + path},
		// inject_texture takes no --format: the pixel format comes from the base asset.
		{"asset", 5, "inject_texture " + path + " " + path + " " + path + " --no-mips --usmap " + path},
		{"asset", 6, "batch_inject_texture " + path + " " + path + " " + path + " --no-mips --usmap " + path},
		{"asset", 7, "extract_texture " + path + " " + path + " --format PNG --mip 0 --usmap " + path},
		{"asset", 8, "batch_extract_texture " + path + " " + path + " --format PNG --mip 0 --usmap " + path},

		// to_zen takes no usmap argument.
		{"zen", 0, "to_zen " + path + " --no-material-tags"},
		{"zen", 1, "create_iostore_bundle " + path + " a/one.uasset b/two.uasset --mount-point ../../../ --compress --encrypt --aes-key " + aes},
		// "Output Base Name" is not path-like by label, so it stays raw here
		// while create_iostore_bundle normalizes its own. Both work.
		{"zen", 2, "create_mod_iostore C:/work/thing --mount-point ../../../ --game-path Marvel/Content/ " + path + " --compress --obfuscate --pak-aes " + aes + " --no-material-tags --hybrid"},
		{"zen", 3, "extract_iostore " + path + " " + path + " --package /Game/Marvel/Foo --chunk-id AABBCC --aes " + aes},
		{"zen", 4, "extract_iostore_legacy " + path + " " + path + " --mod " + path + " --filter SK_1014 SK_1057 --aes " + aes + " --container " + path + " --with-deps"},
		{"zen", 5, "inspect_zen " + path},
		{"zen", 6, "list_iostore " + path + " --aes " + aes + " --filter SK_1014 --type Texture2D,Material --types --script-objects " + path + " --game-paks " + path},
		{"zen", 7, "dump_zen_from_game " + path + " /Game/Marvel/Foo " + path},
		{"zen", 8, "is_iostore_compressed " + path},
		{"zen", 9, "is_iostore_encrypted " + path},
		{"zen", 10, "recompress_iostore " + path},
		{"zen", 11, "extract_script_objects " + path + " " + path},
		// The hash is over the exact string, so the path must not be normalized.
		{"zen", 12, "cityhash /Game/Marvel/Foo"},

		// create_pak takes no --compress; --root must precede the inputs it scopes.
		{"pak", 0, "create_pak " + path + " --root " + path + " " + path + " --mount-point ../../../ --aes-key " + aes},
		{"pak", 1, "create_companion_pak " + path + " a/one.uasset b/two.uasset --mount-point ../../../ --path-hash-seed 0 --aes-key " + aes},
		{"pak", 2, "extract_pak " + path + " " + path + " --aes " + aes + " --list --filter SK_1014 SK_1057"},
		{"pak", 3, "list_pak " + path + " --aes " + aes + " --paths-only --filter SK_1014 SK_1057"},

		{"json", 0, "to_json " + path + " " + path + " " + path + " --compact"},
		{"json", 1, "from_json " + path + " " + path + " " + path},

		{"niagara", 0, "niagara_details " + path + " --usmap " + path},
		{"niagara", 1, "niagara_edit " + path + " --usmap " + path + " --output " + path + ` --edits "[{\"exportIndex\":4}]" --edits-file ` + path},
		{"niagara", 2, "niagara_audit " + path + " " + path},
		// namespace/key/search are literal strings, never path-normalized
		{"niagara", 3, "parse_locres " + path + " --output " + path + " --namespace 601_HeroUIAsset_1011_ST --key HeroInfo_TName --search \"BRUCE BANNER\" --stats"},
	}

	for _, c := range cases {
		if got := argsFor(t, c.menu, c.choice); got != c.want {
			t.Errorf("%s[%d]\n got: %s\nwant: %s", c.menu, c.choice, got, c.want)
		}
	}
}

// A menu row with no form silently does nothing when selected.
func TestNoFormsBeyondMenus(t *testing.T) {
	menus := []struct {
		path string
		menu menuDef
	}{
		{"asset", assetOpsMenu},
		{"zen", zenMenu},
		{"pak", pakMenu},
		{"json", jsonMenu},
		{"niagara", niagaraMenu},
	}
	for _, mn := range menus {
		// last entry is always "← Back"
		commands := len(mn.menu.items) - 1
		for i := 0; i < commands; i++ {
			if getFormForCommand(mn.path, i) == nil {
				t.Errorf("%s: menu item %d (%q) has no form", mn.path, i, mn.menu.items[i].label)
			}
		}
		if getFormForCommand(mn.path, commands) != nil {
			t.Errorf("%s: form %d exists but no menu item selects it", mn.path, commands)
		}
	}
}

func TestParseToolVersionOutput(t *testing.T) {
	cases := []struct {
		out             string
		version, commit string
		stamped         bool
	}{
		// what the released binaries actually print
		{"UAssetTool v1.0.0+952bd331976c6f28efb36ca320c82c27e2456023\r\n", "1.0.0", "952bd331976c6f28efb36ca320c82c27e2456023", false},
		{"UAssetTool v1.5.6\n", "1.5.6", "", true},
		{"UAssetTool 1.5.6+abc\r\n", "1.5.6", "abc", true},
		{"UAssetTool v1.5.6-rc1\n", "1.5.6-rc1", "", true},
		{"UAssetTool vunknown\n", "", "", false},
	}
	for _, c := range cases {
		m := toolVersionOutput.FindStringSubmatch(c.out)
		if c.version == "" {
			if m != nil {
				t.Errorf("%q: expected no match, got %q", c.out, m[0])
			}
			continue
		}
		if m == nil {
			t.Errorf("%q: no match", c.out)
			continue
		}
		if m[1] != c.version || m[2] != c.commit {
			t.Errorf("%q: got version=%q commit=%q", c.out, m[1], m[2])
		}
		if stamped := normalizeVersionTag(m[1]) != unstampedToolVersion; stamped != c.stamped {
			t.Errorf("%q: stamped=%v want %v", c.out, stamped, c.stamped)
		}
	}
}

func TestAesKeyArgStripsHexPrefix(t *testing.T) {
	for in, want := range map[string]string{
		"0xDEADBEEF":   "DEADBEEF",
		"0XDEADBEEF":   "DEADBEEF",
		"DEADBEEF":     "DEADBEEF",
		`"0xDEADBEEF"`: "DEADBEEF",
		"":             "",
		"0x":           "0x",
	} {
		if got := aesKeyArg(in); got != want {
			t.Errorf("aesKeyArg(%q) = %q, want %q", in, got, want)
		}
	}
}
