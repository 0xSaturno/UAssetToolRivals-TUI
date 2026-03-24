# UAssetTool TUI

A terminal UI wrapper for [UAssetToolRivals](https://github.com/XzantGaming/UAssetToolRivals), built in Go with Bubble Tea. It provides a keyboard-and-mouse driven interface for browsing commands, filling in arguments, previewing the generated CLI, streaming live output, and managing persistent settings without memorizing raw command syntax.

## Features

- **Terminal UI workflow**: Navigate command categories, forms, previews, running logs, and results inside a full-screen TUI.
- **Auto download / update**: Fetches the latest platform-matching `UAssetTool` release from GitHub and extracts it beside the TUI executable.
- **Persistent settings**: Stores defaults like `GamePaksDir`, `UsmapPath`, `AesKey`, and `OutputExtractionDir` in `config.json` next to the executable.
- **Command preview**: Lets you review the generated UAssetTool CLI command before running it.
- **Live output streaming**: Shows UAT stdout/stderr live while the command is running.
- **Scrollable output panes**: Running logs and result views support keyboard scrolling, mouse wheel scrolling, track jumping, and draggable scrollbars.
- **Clipboard support**: `Ctrl+C` copies the active input value, running log, or result output instead of closing the app.
- **Smart wrapping / truncation**: Long paths and output lines are wrapped or truncated to fit the terminal cleanly.

## Command Categories

- **Asset Operations**
  - `detect`
  - `batch_detect`
  - `fix`
  - `dump`
  - `skeletal_mesh_info`

- **Zen / IoStore**
  - `to_zen`
  - `create_mod_iostore`
  - `extract_iostore_legacy`
  - `inspect_zen`
  - `is_iostore_compressed`
  - `is_iostore_encrypted`
  - `recompress_iostore`
  - `extract_script_objects`
  - `cityhash`

- **PAK Operations**
  - `create_pak`
  - `create_companion_pak`
  - `extract_pak`

- **JSON Conversion**
  - `to_json`
  - `from_json`

- **Niagara / Other**
  - `niagara_list`
  - `niagara_details`
  - `niagara_edit`
  - `modify_colors`
  - `scan_childbp_isenemy`

## Requirements

- **Windows or Linux**
- **Go** to build from source
- A matching `UAssetTool` binary for your platform

The TUI can download the correct upstream `UAssetTool` release asset for the current OS/architecture if it is not already present.

## Project Layout

- `tui/` - Go TUI source and build output
- `tui/UAssetTool.exe` - the CLI tool executed by the wrapper on Windows
- `tui/UAssetTool` - the CLI tool executed by the wrapper on Linux
- `tui/config.json` - saved settings written by the app

## Build

From the `tui` directory:

### Dev run

```powershell
go run .
```

```bash
go run .
```

### Dev build

```powershell
go build -o uassettool-tui.exe .
```

```bash
go build -o uassettool-tui .
```

### Release build

```powershell
go build -trimpath -ldflags="-s -w" -o uassettool-tui.exe .
```

```bash
go build -trimpath -ldflags="-s -w" -o uassettool-tui .
```

## Usage

Run the TUI from the `tui` directory:

```powershell
.\uassettool-tui.exe
```

```bash
./uassettool-tui
```

Typical flow:

1. Open **Download / Update** if `UAssetTool` has not been downloaded yet for your platform.
2. Open **Settings** and configure your default paths and options.
3. Open **Run Command** and choose a command category.
4. Fill in the generated form fields.
5. Review the command preview if enabled.
6. Run the command and inspect the live output or final result pane.

## Controls

- **Arrow keys / Enter / Esc**: standard navigation
- **Mouse**: select items, scroll output, and drag the scrollbar thumb in running/result panes
- **PgUp / PgDn / Home / End**: faster output navigation
- **Ctrl+C**: copy the active form value, setting value, running log, or result output

## Configuration

The app stores its settings in `config.json` beside the executable. Current settings include:

- `GamePaksDir`
- `UsmapPath`
- `AesKey`
- `OutputExtractionDir`
- `PreviewCommand`
- `EnableAdvancedExtractIoStoreArgs`
- `ToolVersion`

Editing `config.json` manually is possible, but using the in-app **Settings** menu is the intended workflow.

## Notes

- The TUI runs in an alternate screen and captures mouse input for scrolling and dragging.
- `Ctrl+C` is intentionally reserved for copy behavior inside the app and does not quit the program.
- On Windows, `UAssetTool.exe` is expected to live beside `uassettool-tui.exe`.
- On Linux, `UAssetTool` is expected to live beside `uassettool-tui` and is marked executable after download.
