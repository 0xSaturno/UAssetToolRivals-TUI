package app

// ── form types ──────────────────────────────────────────────────────────────

type formField struct {
	label      string
	configKey  string
	defaultVal string
	optional   bool
	boolToggle bool
}

type commandForm struct {
	command string
	fields  []formField
}

// ── form definitions ────────────────────────────────────────────────────────

func getFormForCommand(menuPath string, choice int) *commandForm {
	switch menuPath {
	case "asset":
		return assetForm(choice)
	case "zen":
		return zenForm(choice)
	case "pak":
		return pakForm(choice)
	case "json":
		return jsonForm(choice)
	case "niagara":
		return niagaraForm(choice)
	}
	return nil
}

func assetForm(choice int) *commandForm {
	switch choice {
	case 0:
		return &commandForm{command: "detect", fields: []formField{
			{label: "Asset Path (.uasset) or Directory"},
			{label: "Mappings Path (.usmap)", configKey: "UsmapPath", optional: true},
		}}
	case 1:
		return &commandForm{command: "batch_detect", fields: []formField{
			{label: "Directory"},
			{label: "Mappings Path (.usmap)", configKey: "UsmapPath", optional: true},
		}}
	case 2:
		return &commandForm{command: "fix", fields: []formField{
			{label: "Asset Path (.uasset)"},
			{label: "Mappings Path (.usmap)", configKey: "UsmapPath", optional: true},
		}}
	case 3:
		return &commandForm{command: "dump", fields: []formField{
			{label: "Asset Path (.uasset)"},
			{label: "Mappings Path (.usmap)", configKey: "UsmapPath"},
		}}
	case 4:
		return &commandForm{command: "skeletal_mesh_info", fields: []formField{
			{label: "Asset Path (.uasset)"},
			{label: "Mappings Path (.usmap)", configKey: "UsmapPath"},
		}}
	case 5:
		// pixel format comes from the base asset; no --format option
		return &commandForm{command: "inject_texture", fields: []formField{
			{label: "Base UAsset (.uasset)"},
			{label: "Image File (png/tga/dds/bmp/jpeg)"},
			{label: "Output UAsset (.uasset)"},
			{label: "Disable Mipmaps?", boolToggle: true, optional: true},
			{label: "Mappings Path (.usmap)", configKey: "UsmapPath", optional: true},
		}}
	case 6:
		return &commandForm{command: "batch_inject_texture", fields: []formField{
			{label: "UAsset Directory"},
			{label: "Image Directory"},
			{label: "Output Directory"},
			{label: "Disable Mipmaps?", boolToggle: true, optional: true},
			{label: "Mappings Path (.usmap)", configKey: "UsmapPath", optional: true},
		}}
	case 7:
		return &commandForm{command: "extract_texture", fields: []formField{
			{label: "Texture UAsset (.uasset)"},
			{label: "Output Image Path"},
			{label: "Output Format (PNG/TGA/DDS/BMP)", optional: true},
			{label: "Mip Index", optional: true},
			{label: "Mappings Path (.usmap)", configKey: "UsmapPath", optional: true},
		}}
	case 8:
		return &commandForm{command: "batch_extract_texture", fields: []formField{
			{label: "UAsset Directory"},
			{label: "Output Directory"},
			{label: "Output Format (PNG/TGA/DDS/BMP)", optional: true},
			{label: "Mip Index", optional: true},
			{label: "Mappings Path (.usmap)", configKey: "UsmapPath", optional: true},
		}}
	}
	return nil
}

func zenForm(choice int) *commandForm {
	switch choice {
	case 0:
		return &commandForm{command: "to_zen", fields: []formField{
			{label: "Asset Path (.uasset)"},
			{label: "Disable Material Tags?", boolToggle: true, optional: true},
		}}
	case 1:
		return &commandForm{command: "create_iostore_bundle", fields: []formField{
			{label: "Output Base Name"},
			{label: "Input Files (space sep)"},
			{label: "Mount Point (default ../../../)", optional: true},
			{label: "Compress?", boolToggle: true, optional: true},
			{label: "Enable Encryption?", boolToggle: true, optional: true},
			{label: "AES Key", configKey: "AesKey", optional: true},
		}}
	case 2:
		return &commandForm{command: "create_mod_iostore", fields: []formField{
			{label: "Output Base Name"},
			{label: "Mount Point (default ../../../)", optional: true},
			{label: "Game Path (default Marvel/Content/)", optional: true},
			{label: "Input Files/Dirs/.pak (space sep)"},
			{label: "Compress?", boolToggle: true, optional: true},
			{label: "Enable Obfuscation?", boolToggle: true, optional: true},
			{label: "PAK AES Key", configKey: "AesKey", optional: true},
			{label: "Disable Material Tags?", boolToggle: true, optional: true},
			{label: "Hybrid (embed non-Unreal files)?", boolToggle: true, optional: true},
		}}
	case 3:
		return &commandForm{command: "extract_iostore", fields: []formField{
			{label: "UTOC Path"},
			{label: "Output Directory"},
			{label: "Package Name", optional: true},
			{label: "Chunk ID (hex)", optional: true},
			{label: "AES Key", configKey: "AesKey", optional: true},
		}}
	case 4:
		return &commandForm{command: "extract_iostore_legacy", fields: []formField{
			{label: "Paks Directory", configKey: "GamePaksDir"},
			{label: "Output Directory", configKey: "OutputExtractionDir"},
			{label: "Mod Container Path(s)", optional: true},
			{label: "Filter Patterns (space sep or .txt file)", optional: true},
			{label: "AES Key", configKey: "AesKey", optional: true},
			{label: "Extra Container Path(s)", optional: true},
			{label: "Extract dependencies?", boolToggle: true, optional: true, defaultVal: "N"},
		}}
	case 5:
		return &commandForm{command: "inspect_zen", fields: []formField{
			{label: "Zen File (.ucas/.zen)"},
		}}
	case 6:
		return &commandForm{command: "list_iostore", fields: []formField{
			{label: "UTOC Path or Directory"},
			{label: "AES Key", configKey: "AesKey", optional: true},
			{label: "Filter Pattern (single)", optional: true},
			{label: "Asset Type Filter (comma sep)", optional: true},
			{label: "Report Asset Types?", boolToggle: true, optional: true},
			{label: "ScriptObjects.bin Path", optional: true},
			{label: "Game Paks Dir (for script objects)", optional: true},
		}}
	case 7:
		return &commandForm{command: "dump_zen_from_game", fields: []formField{
			{label: "Paks Directory", configKey: "GamePaksDir"},
			{label: "Package Path (/Game/...)"},
			{label: "Output File", optional: true},
		}}
	case 8:
		return &commandForm{command: "is_iostore_compressed", fields: []formField{
			{label: "UTOC Path"},
		}}
	case 9:
		return &commandForm{command: "is_iostore_encrypted", fields: []formField{
			{label: "UTOC Path"},
		}}
	case 10:
		return &commandForm{command: "recompress_iostore", fields: []formField{
			{label: "UTOC Path"},
		}}
	case 11:
		return &commandForm{command: "extract_script_objects", fields: []formField{
			{label: "Paks Directory", configKey: "GamePaksDir"},
			{label: "Output File (ScriptObjects.bin)"},
		}}
	case 12:
		return &commandForm{command: "cityhash", fields: []formField{
			{label: "Path/String for cityhash"},
		}}
	}
	return nil
}

func pakForm(choice int) *commandForm {
	switch choice {
	case 0:
		// --root only scopes the inputs after it, so it is asked first
		return &commandForm{command: "create_pak", fields: []formField{
			{label: "Output PAK Path"},
			{label: "Root Directory (base for in-PAK paths)", optional: true},
			{label: "Files or Directories (space sep)"},
			{label: "Mount Point (default ../../../)", optional: true},
			{label: "AES Key", configKey: "AesKey", optional: true},
		}}
	case 1:
		return &commandForm{command: "create_companion_pak", fields: []formField{
			{label: "Output PAK Path"},
			{label: "File Paths (space sep)"},
			{label: "Mount Point (default ../../../)", optional: true},
			{label: "Path Hash Seed", optional: true},
			{label: "AES Key", configKey: "AesKey", optional: true},
		}}
	case 2:
		return &commandForm{command: "extract_pak", fields: []formField{
			{label: "PAK File"},
			{label: "Output Directory", configKey: "OutputExtractionDir"},
			{label: "AES Key", configKey: "AesKey", optional: true},
			{label: "List only?", boolToggle: true, optional: true},
			{label: "Filter Patterns (space sep)", optional: true},
		}}
	case 3:
		// --filter is greedy, so it stays last
		return &commandForm{command: "list_pak", fields: []formField{
			{label: "PAK File"},
			{label: "AES Key", configKey: "AesKey", optional: true},
			{label: "Paths only?", boolToggle: true, optional: true},
			{label: "Filter Patterns (space sep)", optional: true},
		}}
	}
	return nil
}

func jsonForm(choice int) *commandForm {
	switch choice {
	case 0:
		return &commandForm{command: "to_json", fields: []formField{
			{label: "Asset Path or Directory"},
			{label: "Mappings Path (.usmap)", configKey: "UsmapPath", optional: true},
			{label: "Output Directory", optional: true},
			{label: "Compact (read-only) JSON?", boolToggle: true, optional: true},
		}}
	case 1:
		return &commandForm{command: "from_json", fields: []formField{
			{label: "JSON File or Directory"},
			{label: "Output UAsset Path or Directory"},
			{label: "Mappings Path (.usmap)", configKey: "UsmapPath", optional: true},
		}}
	}
	return nil
}

func niagaraForm(choice int) *commandForm {
	switch choice {
	case 0:
		return &commandForm{command: "niagara_details", fields: []formField{
			{label: "Asset Path (.uasset)"},
			{label: "Mappings Path (.usmap)", configKey: "UsmapPath", optional: true},
		}}
	case 1:
		return &commandForm{command: "niagara_edit", fields: []formField{
			{label: "Asset Path (.uasset)"},
			{label: "Mappings Path (.usmap)", configKey: "UsmapPath", optional: true},
			{label: "Output Path (.uasset)", optional: true},
			{label: "Edits JSON", optional: true},
			{label: "Edits JSON File", optional: true},
		}}
	case 2:
		return &commandForm{command: "niagara_audit", fields: []formField{
			{label: "Asset Path (.uasset)"},
			{label: "Mappings Path (.usmap)", configKey: "UsmapPath", optional: true},
		}}
	case 3:
		return &commandForm{command: "parse_locres", fields: []formField{
			{label: "Locres File or Directory"},
			{label: "Output JSON Path (recommended for full dumps)", optional: true},
			{label: "Namespace Filter", optional: true},
			{label: "Key Lookup (needs namespace)", optional: true},
			{label: "Search Term", optional: true},
			{label: "Stats only?", boolToggle: true, optional: true},
		}}
	}
	return nil
}
