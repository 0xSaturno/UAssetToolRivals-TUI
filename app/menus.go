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
	title := "UAssetTool TUI"
	tuiVersion := normalizeVersionTag(currentTUIVersion())
	if tuiVersion != "" && tuiVersion != "dev" && tuiVersion != "(devel)" {
		title = "UAssetTool TUI v" + tuiVersion
	}
	return menuDef{
		title:    title,
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
		{"Asset Operations", "🔍", accentGreen, "Detect, inspect, repair, and edit asset files"},
		{"Zen / IoStore", "📦", accentBlue, "Convert, extract, build, and inspect IoStore data"},
		{"PAK Operations", "🗃", accentYellow, "Create, extract, and package PAK archives"},
		{"JSON Conversion", "🔄", accentGreen, "Convert assets to and from editable JSON"},
		{"Niagara / Other", "✨", accentMagenta, "Inspect Niagara data and run batch utilities"},
		{"← Back", " ", dimStyle, ""},
	},
}

var assetOpsMenu = menuDef{
	title:    "🔍 Asset Operations",
	subtitle: "Inspect, fix, and edit UAsset files",
	items: []menuItem{
		{"Detect Type", "•", accentCyan, "Identify what kind of asset a file contains"},
		{"Batch Detect Type", "•", accentCyan, "Scan a folder and identify asset types in bulk"},
		{"Fix SerializeSize", "•", accentCyan, "Repair broken serialize size values in an asset"},
		{"Dump Info", "•", accentCyan, "Print detailed metadata and structure information"},
		{"Skeletal Mesh Info", "•", accentCyan, "Show skeletal mesh sections, LODs, and related data"},
		{"Inject Texture", "•", accentCyan, "Inject an image into a Texture2D asset"},
		{"Batch Inject Textures", "•", accentCyan, "Inject matching replacement images into Texture2D assets in bulk"},
		{"Extract Texture", "•", accentCyan, "Extract a Texture2D asset to an image file"},
		{"← Back", " ", dimStyle, ""},
	},
}

var zenMenu = menuDef{
	title:    "📦 Zen / IoStore Operations",
	subtitle: "IoStore conversion and inspection",
	items: []menuItem{
		{"Legacy to Zen", "•", accentBlue, "Convert a legacy asset into Zen format"},
		{"Create IoStore Bundle", "•", accentBlue, "Create a complete IoStore bundle from files"},
		{"Create Mod IoStore", "•", accentBlue, "Build a mod-ready IoStore container"},
		{"Extract IoStore Raw", "•", accentBlue, "Extract or list raw chunks from an IoStore container"},
		{"Extract IoStore Legacy", "•", accentBlue, "Extract files from IoStore into legacy assets"},
		{"Inspect Zen", "•", accentBlue, "View details and contents of a Zen container"},
		{"List IoStore", "•", accentBlue, "List packages and chunk types inside IoStore containers"},
		{"Dump Zen From Game", "•", accentBlue, "Dump a raw Zen package from the game Paks directory"},
		{"Check Compression", "•", accentBlue, "Check whether an IoStore container is compressed"},
		{"Check Encryption", "•", accentBlue, "Check whether an IoStore container is encrypted"},
		{"Recompress IoStore", "•", accentBlue, "Recompress an existing IoStore container"},
		{"Extract ScriptObjects.bin", "•", accentBlue, "Export ScriptObjects.bin from the game Paks folder"},
		{"CityHash Path/String", "•", accentBlue, "Generate CityHash values for paths or text"},
		{"← Back", " ", dimStyle, ""},
	},
}

var pakMenu = menuDef{
	title:    "🗃 PAK Operations",
	subtitle: "Create and extract PAK archives",
	items: []menuItem{
		{"Create PAK", "•", accentYellow, "Package files into a new PAK archive"},
		{"Create Companion PAK", "•", accentYellow, "Build the companion PAK file used with a mod container"},
		{"Extract/List PAK", "•", accentYellow, "Extract files from a PAK or list its contents"},
		{"← Back", " ", dimStyle, ""},
	},
}

var jsonMenu = menuDef{
	title:    "🔄 JSON Conversion",
	subtitle: "Convert between UAsset and JSON",
	items: []menuItem{
		{"UAsset to JSON", "•", accentGreen, "Convert a binary asset into editable JSON"},
		{"JSON to UAsset", "•", accentGreen, "Build a binary asset from edited JSON"},
		{"← Back", " ", dimStyle, ""},
	},
}

var niagaraMenu = menuDef{
	title:    "✨ Niagara / Other",
	subtitle: "Niagara assets and batch utilities",
	items: []menuItem{
		{"Niagara Details", "•", accentMagenta, "Inspect Niagara color curves and asset details"},
		{"Niagara Edit", "•", accentMagenta, "Apply JSON-based edits to selected Niagara exports"},
		{"Niagara Audit", "•", accentMagenta, "Deep-scan Niagara exports for color-related data"},
		{"Scan ChildBP IsEnemy", "•", accentMagenta, "Scan ChildBP assets for IsEnemy parameter redirects"},
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
