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
			{label: "Asset Path (.uasset)"},
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
			{label: "Mappings Path (.usmap)", configKey: "UsmapPath", optional: true},
		}}
	case 4:
		return &commandForm{command: "skeletal_mesh_info", fields: []formField{
			{label: "Asset Path (.uasset)"},
			{label: "Mappings Path (.usmap)", configKey: "UsmapPath"},
		}}
	case 5:
		return &commandForm{command: "inject_texture", fields: []formField{
			{label: "Base UAsset (.uasset)"},
			{label: "Image File (png/tga/dds/bmp/jpeg)"},
			{label: "Output UAsset (.uasset)"},
			{label: "Compression Format (BC7/BC3/BC1/BC5/BC4/BGRA8)", optional: true},
			{label: "Disable Mipmaps?", boolToggle: true, optional: true},
			{label: "Mappings Path (.usmap)", configKey: "UsmapPath", optional: true},
		}}
	case 6:
		return &commandForm{command: "extract_texture", fields: []formField{
			{label: "Texture UAsset (.uasset)"},
			{label: "Output Image Path"},
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
			{label: "Mappings Path (.usmap)", configKey: "UsmapPath", optional: true},
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
			{label: "UAsset Files (space sep or directory)"},
			{label: "Compress?", boolToggle: true, optional: true},
			{label: "Enable Obfuscation?", boolToggle: true, optional: true},
			{label: "PAK AES Key", configKey: "AesKey", optional: true},
			{label: "Disable Material Tags?", boolToggle: true, optional: true},
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
			{label: "Filter Patterns (space sep)", optional: true},
			{label: "AES Key", configKey: "AesKey", optional: true},
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
			{label: "Filter Pattern", optional: true},
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
		return &commandForm{command: "create_pak", fields: []formField{
			{label: "Output PAK Path"},
			{label: "Files (space sep)"},
			{label: "Mount Point (default ../../../)", optional: true},
			{label: "Enable compression?", boolToggle: true, optional: true},
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
		return &commandForm{command: "scan_childbp_isenemy", fields: []formField{
			{label: "Paks Directory or Extracted Folder", configKey: "GamePaksDir"},
			{label: "AES Key", configKey: "AesKey", optional: true},
			{label: "Is Extracted?", boolToggle: true, optional: true},
		}}
	}
	return nil
}
