package app

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// ── menu types ──────────────────────────────────────────────────────────────

type menuItem struct {
	label string
	icon  string
	color lipgloss.Style
	desc  string
}

type menuDef struct {
	title    string
	subtitle string
	items    []menuItem
}

// ── menu definitions ────────────────────────────────────────────────────────

func mainMenuDef(version string) menuDef {
	sub := "UAssetTool not found — download it below"
	if version != "" {
		sub = "Running UAssetTool " + version
	}
	return menuDef{
		title:    "UAssetTool CLI",
		subtitle: sub,
		items: []menuItem{
			{"Run Command", "▶", accentGreen, "Execute UAssetTool CLI commands"},
			{"Download / Update", "⬇", accentBlue, "Fetch latest release from GitHub"},
			{"Settings", "⚙", accentYellow, "Configure paths and preferences"},
			{"Exit", "✕", dimStyle, ""},
		},
	}
}

var categoryMenu = menuDef{
	title:    "📂 Command Category",
	subtitle: "Choose an operation type",
	items: []menuItem{
		{"Asset Operations", "🔍", accentGreen, "detect, fix, dump, skeletal mesh"},
		{"Zen / IoStore", "📦", accentBlue, "convert, extract, create, inspect"},
		{"PAK Operations", "🗃", accentYellow, "create, extract, companion pak"},
		{"JSON Conversion", "🔄", accentGreen, "uasset ↔ json"},
		{"Niagara / Other", "✨", accentMagenta, "niagara, colors, scan"},
		{"← Back", " ", dimStyle, ""},
	},
}

var assetOpsMenu = menuDef{
	title:    "🔍 Asset Operations",
	subtitle: "Inspect and fix UAsset files",
	items: []menuItem{
		{"Detect Type", "•", accentCyan, "detect"},
		{"Batch Detect Type", "•", accentCyan, "batch_detect"},
		{"Fix SerializeSize", "•", accentCyan, "fix"},
		{"Dump Info", "•", accentCyan, "dump"},
		{"Skeletal Mesh Info", "•", accentCyan, "skeletal_mesh_info"},
		{"← Back", " ", dimStyle, ""},
	},
}

var zenMenu = menuDef{
	title:    "📦 Zen / IoStore Operations",
	subtitle: "IoStore conversion and inspection",
	items: []menuItem{
		{"Legacy to Zen", "•", accentBlue, "to_zen"},
		{"Create Mod IoStore", "•", accentBlue, "create_mod_iostore"},
		{"Extract IoStore Legacy", "•", accentBlue, "extract_iostore_legacy"},
		{"Inspect Zen", "•", accentBlue, "inspect_zen"},
		{"Check Compression", "•", accentBlue, "is_iostore_compressed"},
		{"Check Encryption", "•", accentBlue, "is_iostore_encrypted"},
		{"Recompress IoStore", "•", accentBlue, "recompress_iostore"},
		{"Extract ScriptObjects.bin", "•", accentBlue, "extract_script_objects"},
		{"CityHash Path/String", "•", accentBlue, "cityhash"},
		{"← Back", " ", dimStyle, ""},
	},
}

var pakMenu = menuDef{
	title:    "🗃 PAK Operations",
	subtitle: "Create and extract PAK archives",
	items: []menuItem{
		{"Create PAK", "•", accentYellow, "create_pak"},
		{"Create Companion PAK", "•", accentYellow, "create_companion_pak"},
		{"Extract/List PAK", "•", accentYellow, "extract_pak"},
		{"← Back", " ", dimStyle, ""},
	},
}

var jsonMenu = menuDef{
	title:    "🔄 JSON Conversion",
	subtitle: "Convert between UAsset and JSON",
	items: []menuItem{
		{"UAsset to JSON", "•", accentGreen, "to_json"},
		{"JSON to UAsset", "•", accentGreen, "from_json"},
		{"← Back", " ", dimStyle, ""},
	},
}

var niagaraMenu = menuDef{
	title:    "✨ Niagara / Other",
	subtitle: "Niagara assets and batch utilities",
	items: []menuItem{
		{"Niagara List", "•", accentMagenta, "niagara_list"},
		{"Niagara Details", "•", accentMagenta, "niagara_details"},
		{"Niagara Edit", "•", accentMagenta, "niagara_edit"},
		{"Modify Colors Batch", "•", accentMagenta, "modify_colors"},
		{"Scan ChildBP IsEnemy", "•", accentMagenta, "scan_childbp_isenemy"},
		{"← Back", " ", dimStyle, ""},
	},
}

// ── settings menu (dynamic) ─────────────────────────────────────────────────

func settingsMenuDef(cfg Config) menuDef {
	return menuDef{
		title:    "⚙ Settings",
		subtitle: "Configure defaults and preferences",
		items: []menuItem{
			{fmt.Sprintf("Game Paks Dir  %s", dimVal(cfg.GamePaksDir)), "📁", accentCyan, ""},
			{fmt.Sprintf("USMAP Path  %s", dimVal(cfg.UsmapPath)), "📄", accentCyan, ""},
			{fmt.Sprintf("AES Key  %s", dimVal(cfg.AesKey)), "🔑", accentCyan, ""},
			{fmt.Sprintf("Output Dir  %s", dimVal(cfg.OutputExtractionDir)), "📂", accentCyan, ""},
			{toggleLabel("Command Preview", cfg.PreviewCommand), "👁", accentYellow, ""},
			{toggleLabel("Advanced Extract IoStore Args", cfg.EnableAdvancedExtractIoStoreArgs), "🧪", accentYellow, ""},
			{"← Back", " ", dimStyle, ""},
		},
	}
}

func toggleLabel(label string, on bool) string {
	if on {
		return label + "  ON"
	}
	return label + "  OFF"
}

func dimVal(v string) string {
	if v == "" {
		return dimStyle.Render("(not set)")
	}
	if len(v) > 50 {
		return v[:47] + "..."
	}
	return v
}
