# Interactive UAssetTool Menu
# Allows building, updating, and running UAssetTool CLI easily.

$ScriptDir = $PSScriptRoot
$ConfigFile = Join-Path $ScriptDir "config.json"
$SubmoduleDir = Join-Path $ScriptDir "UassetToolRivals"
$ProjectFile = Join-Path $SubmoduleDir "src/UAssetTool/UAssetTool.csproj"
$ExePath = Join-Path $ScriptDir "UAssetTool.exe"

# --- Config Management ---
$Config = @{}

function Load-Config {
    if (Test-Path $ConfigFile) {
        try {
            $json = Get-Content $ConfigFile -Raw | ConvertFrom-Json
            # Copy properties to hash table
            foreach ($prop in $json.PSObject.Properties) {
                $Config[$prop.Name] = $prop.Value
            }
        }
        catch {
            Write-Warning "Failed to load config.json"
        }
    }
}

function Save-Config {
    try {
        $Config | ConvertTo-Json -Depth 2 | Set-Content $ConfigFile
        Write-Host "Settings saved." -ForegroundColor Green
    }
    catch {
        Write-Warning "Failed to save config.json"
    }
}

function Get-Config {
    param($Key)
    return $Config[$Key]
}

function Get-ConfigBool {
    param($Key, $Default = $false)

    $value = Get-Config $Key
    if ($null -eq $value) { return $Default }

    $normalized = $value.ToString().Trim().ToLowerInvariant()
    return $normalized -in @("1", "true", "yes", "y", "on")
}

function Set-Config {
    param($Key, $Value)
    $Config[$Key] = $Value
    Save-Config
}

# --- Infrastructure ---

function Show-Header {
    Clear-Host
    Write-Host "============================" -ForegroundColor Cyan
    Write-Host "   UAssetTool Manager" -ForegroundColor White
    Write-Host "============================" -ForegroundColor Cyan
    Write-Host ""
}

function Write-ColorDivider {
    param(
        [string]$Color = "DarkGray",
        [int]$Width = 44
    )
    Write-Host ("=" * $Width) -ForegroundColor $Color
}

function Show-MenuHeader {
    param(
        [string]$Title,
        [string]$Color = "Cyan"
    )

    Clear-Host
    Write-ColorDivider $Color
    Write-Host " $Title" -ForegroundColor $Color
    Write-ColorDivider "DarkGray"
}

function Check-Prerequisites {
    if (-not (Get-Command dotnet -ErrorAction SilentlyContinue)) {
        Write-Warning "dotnet SDK is not found in PATH. Please install .NET 8 SDK."
        return $false
    }
    if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
        Write-Warning "git is not found in PATH."
        return $false
    }
    return $true
}

function Download-PrecompiledTool {
    Write-Host "Downloading precompiled UAssetTool from latest release..." -ForegroundColor Yellow
    $zipUrl = "https://github.com/XzantGaming/UassetToolRivals/releases/latest/download/UAssetTool-win-x64.zip"
    $tempZip = Join-Path $env:TEMP "UAssetTool-win-x64.zip"
    $targetExe = Join-Path $ScriptDir "UAssetTool.exe"
    
    try {
        [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
        Invoke-WebRequest -Uri $zipUrl -OutFile $tempZip -UseBasicParsing
        
        Write-Host "Extracting..." -ForegroundColor Yellow
        $extractDir = Join-Path $env:TEMP "UAssetTool-Extract"
        if (Test-Path $extractDir) { Remove-Item $extractDir -Recurse -Force -ErrorAction SilentlyContinue }
        
        Expand-Archive -Path $tempZip -DestinationPath $extractDir -Force
        
        $extractedExe = Join-Path $extractDir "UAssetTool.exe"
        if (Test-Path $extractedExe) {
            Move-Item -Path $extractedExe -Destination $targetExe -Force
            # Update LastWriteTime to ensure Get-ExePath prioritizes it if it's newer
            (Get-Item $targetExe).LastWriteTime = Get-Date
            Write-Host "Download and extraction complete! Saved to $targetExe" -ForegroundColor Green
        }
        else {
            Write-Warning "UAssetTool.exe not found in the downloaded zip."
        }
    }
    catch {
        Write-Error "Failed to download or extract the precompiled tool. $_"
    }
    finally {
        if (Test-Path $tempZip) { Remove-Item $tempZip -Force -ErrorAction SilentlyContinue }
        if (Test-Path $extractDir) { Remove-Item $extractDir -Recurse -Force -ErrorAction SilentlyContinue }
    }
    Pause
}

function Run-Process-Wait {
    param($Arguments)
    
    if (-not (Test-Path $ExePath)) {
        Write-Warning "UAssetTool.exe not found."
        Write-Warning "Please download the precompiled version first (or build using UAssetTool-Dev.ps1)."
        Pause
        return
    }

    Write-Host ""
    Write-ColorDivider "Cyan"
    Write-Host "Running: UAssetTool.exe $Arguments" -ForegroundColor White
    Write-ColorDivider "DarkGray"

    if (Get-ConfigBool "PreviewCommand" $false) {
        Write-Host "[PREVIEW] Full Command" -ForegroundColor Magenta
        Write-Host "`"$ExePath`" $Arguments" -ForegroundColor Yellow
        $confirm = Read-Host "Proceed? (Y/N, default Y)"
        if ($confirm -eq "N") {
            Write-Host "Command canceled." -ForegroundColor Yellow
            Pause
            return
        }
        Write-ColorDivider "DarkMagenta"
    }

    Write-Host "[DEBUG] Start: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')" -ForegroundColor DarkGray
    
    $pinfo = New-Object System.Diagnostics.ProcessStartInfo
    $pinfo.FileName = $ExePath
    $pinfo.Arguments = $Arguments
    $pinfo.UseShellExecute = $false
    $pinfo.RedirectStandardOutput = $false 
    $pinfo.RedirectStandardError = $false
    
    $p = [System.Diagnostics.Process]::Start($pinfo)
    $p.WaitForExit()
    
    Write-Host ""
    Write-Host "[DEBUG] Exit Code: $($p.ExitCode)" -ForegroundColor DarkGray
    Write-ColorDivider "DarkGray"
    Write-Host "Finished." -ForegroundColor Green
    Write-ColorDivider "Green"
    Pause
}

function Get-Input {
    param($Prompt, $ConfigKey = $null, $Mandatory = $true)
    
    $default = if ($ConfigKey) { Get-Config $ConfigKey } else { $null }
    
    $displayPrompt = $Prompt
    if ($default) {
        $displayPrompt += " [Default: $default]"
    }
    
    $val = Read-Host "$displayPrompt"
    
    if ([string]::IsNullOrWhiteSpace($val)) {
        if ($default) { return $default }
        if ($Mandatory) {
            Write-Warning "Input is required."
            return $null
        }
    }
    return $val
}

function Get-File {
    param($Prompt, $ConfigKey = $null)
    $val = Get-Input "$Prompt (Path)" $ConfigKey $false
    if ([string]::IsNullOrWhiteSpace($val)) { return $null }
    return $val.Trim('"')
}

function Convert-ToQuotedArgList {
    param([string]$InputText)

    if ([string]::IsNullOrWhiteSpace($InputText)) {
        return @()
    }

    # Supports tokens separated by whitespace while preserving single/double-quoted groups.
    $pattern = '"[^\"]*"|''[^'']*''|\S+'
    $regexMatches = [regex]::Matches($InputText, $pattern)
    $quotedArgs = @()

    foreach ($match in $regexMatches) {
        $token = $match.Value.Trim()
        if ([string]::IsNullOrWhiteSpace($token)) { continue }

        if (($token.StartsWith('"') -and $token.EndsWith('"')) -or ($token.StartsWith("'") -and $token.EndsWith("'"))) {
            if ($token.Length -ge 2) {
                $token = $token.Substring(1, $token.Length - 2)
            }
        }

        if (-not [string]::IsNullOrWhiteSpace($token)) {
            $escaped = $token.Replace('"', '\\"')
            $quotedArgs += "`"$escaped`""
        }
    }

    return $quotedArgs
}

# --- Command Categories ---

function Menu-AssetOps {
    while ($true) {
        Show-MenuHeader "Asset Operations" "Cyan"
        Write-Host "1. Detect Type (detect)"
        Write-Host "2. Batch Detect Type (batch_detect)"
        Write-Host "3. Fix SerializeSize (fix)"
        Write-Host "4. Dump Info (dump)"
        Write-Host "5. Skeletal Mesh Info (skeletal_mesh_info)"
        Write-ColorDivider "DarkGray" 36
        Write-Host "6. Back" -ForegroundColor Gray
        
        $c = Read-Host "Choose"
        switch ($c) {
            "1" { 
                $f = Get-File "Asset Path (.uasset)"
                if ($f) { Run-Process-Wait "detect `"$f`"" }
            }
            "2" {
                $d = Get-File "Directory"
                $m = Get-File "Mappings Path (.usmap) [Optional]" "UsmapPath"
                $procArgs = "batch_detect `"$d`""
                if ($m) { $procArgs += " `"$m`"" }
                if ($d) { Run-Process-Wait $procArgs }
            }
            "3" {
                $f = Get-File "Asset Path (.uasset)"
                if ($f) { Run-Process-Wait "fix `"$f`"" }
            }
            "4" {
                $f = Get-File "Asset Path (.uasset)"
                $m = Get-File "Mappings Path (.usmap)" "UsmapPath"
                $procArgs = "dump `"$f`""
                if ($m) { $procArgs += " `"$m`"" }
                if ($f) { Run-Process-Wait $procArgs }
            }
            "5" {
                $f = Get-File "Asset Path (.uasset)"
                $m = Get-File "Mappings Path (.usmap)" "UsmapPath"
                if ($f -and $m) { Run-Process-Wait "skeletal_mesh_info `"$f`" `"$m`"" }
            }
            "6" { return }
        }
    }
}

function Menu-ZenIoStore {
    while ($true) {
        Show-MenuHeader "Zen / IoStore Operations" "Blue"
        Write-Host "1. Legacy to Zen (to_zen)"
        Write-Host "2. Create Mod IoStore (create_mod_iostore)"
        Write-Host "3. Extract IoStore Legacy (extract_iostore_legacy)"
        Write-Host "4. Inspect Zen (inspect_zen)"
        Write-Host "5. Check Compression (is_iostore_compressed)"
        Write-Host "6. Check Encryption (is_iostore_encrypted)"
        Write-Host "7. Recompress IoStore (recompress_iostore)"
        Write-Host "8. Extract ScriptObjects.bin (extract_script_objects)"
        Write-Host "9. CityHash Path/String (cityhash)"
        Write-ColorDivider "DarkBlue" 40
        Write-Host "10. Back" -ForegroundColor Gray
        
        $c = Read-Host "Choose"
        switch ($c) {
            "1" {
                $f = Get-File "Asset Path (.uasset)"
                $m = Get-File "Mappings Path (.usmap) [Optional]" "UsmapPath"
                $disableTags = Read-Host "Disable Material Tags? (Y/N, default N)"
                $procArgs = "to_zen `"$f`""
                if ($m) { $procArgs += " `"$m`"" }
                if ($disableTags -eq "Y") { $procArgs += " --no-material-tags" }
                if ($f) { Run-Process-Wait $procArgs }
            }
            "2" {
                $o = Get-Input "Output Base Name"
                $u = Get-File "Mappings Path (.usmap) [Optional]" "UsmapPath"
                $f = Get-Input "UAsset Files (space sep or directory)"
                $cmp = Read-Host "Compress? (Y/N, default Y)"
                $obf = Read-Host "Enable Obfuscation? (Y/N, default N)"
                $pakAes = Get-Input "PAK AES Key (optional, for encrypted input .pak)" "AesKey" $false
                $noTags = Read-Host "Disable Material Tag injection? (Y/N, default N)"
                $procArgs = "create_mod_iostore `"$o`""
                if ($u) { $procArgs += " --usmap `"$u`"" }
                if ($cmp -eq "N") { $procArgs += " --no-compress" } else { $procArgs += " --compress" }
                if ($obf -eq "Y") { $procArgs += " --obfuscate" }
                if ($pakAes) { $procArgs += " --pak-aes `"$pakAes`"" }
                if ($noTags -eq "Y") { $procArgs += " --no-material-tags" }
                $inputArgs = Convert-ToQuotedArgList $f
                if ($inputArgs.Count -gt 0) { $procArgs += " " + ($inputArgs -join " ") }
                if ($o -and $f) { Run-Process-Wait $procArgs }
            }
            "3" {
                $p = Get-File "Paks Directory (contains .utoc)" "GamePaksDir"
                if (-not $p) { return }
                
                $o = Get-Input "Output Directory" "OutputExtractionDir" $true
                $filt = Get-Input "Filter Patterns (optional, space sep)" $null $false
                $deps = Read-Host "Extract dependencies too? (Y/N, default N)"
                $mod = Get-Input "Mod Container Path(s) (optional, space sep)" $null $false
                $scriptObj = $null
                $globalUtoc = $null

                if (Get-ConfigBool "EnableAdvancedExtractIoStoreArgs" $false) {
                    $scriptObj = Get-File "ScriptObjects.bin Path (optional)"
                    $globalUtoc = Get-File "global.utoc Path (optional)"
                }
                
                $procArgs = "extract_iostore_legacy `"$p`" `"$o`""
                if ($filt) { $procArgs += " --filter $filt" }
                if ($deps -eq "Y") { $procArgs += " --with-deps" }
                if ($mod) {
                    $modArgs = Convert-ToQuotedArgList $mod
                    if ($modArgs.Count -gt 0) { $procArgs += " --mod " + ($modArgs -join " ") }
                }
                if ($scriptObj) { $procArgs += " --script-objects `"$scriptObj`"" }
                if ($globalUtoc) { $procArgs += " --global `"$globalUtoc`"" }
                
                if ($p -and $o) { Run-Process-Wait $procArgs }
            }
            "4" {
                $f = Get-File "Zen File (.ucas/.zen)"
                if ($f) { Run-Process-Wait "inspect_zen `"$f`"" }
            }
            "5" {
                $f = Get-File "UTOC Path"
                if ($f) { Run-Process-Wait "is_iostore_compressed `"$f`"" }
            }
            "6" {
                $f = Get-File "UTOC Path"
                if ($f) { Run-Process-Wait "is_iostore_encrypted `"$f`"" }
            }
            "7" {
                $f = Get-File "UTOC Path"
                if ($f) { Run-Process-Wait "recompress_iostore `"$f`"" }
            }
            "8" {
                $p = Get-File "Paks Directory" "GamePaksDir"
                $o = Get-File "Output File (ScriptObjects.bin)"
                if ($p -and $o) { Run-Process-Wait "extract_script_objects `"$p`" `"$o`"" }
            }
            "9" {
                $inputText = Get-Input "Path/String for cityhash"
                if ($inputText) { Run-Process-Wait "cityhash `"$inputText`"" }
            }
            "10" { return }
        }
    }
}

function Menu-PakOps {
    while ($true) {
        Show-MenuHeader "PAK Operations" "Yellow"
        Write-Host "1. Create PAK (create_pak)"
        Write-Host "2. Create Companion PAK (create_companion_pak)"
        Write-Host "3. Extract/List PAK (extract_pak)"
        Write-ColorDivider "DarkYellow" 36
        Write-Host "4. Back" -ForegroundColor Gray
        
        $c = Read-Host "Choose"
        switch ($c) {
            "1" {
                $o = Get-Input "Output PAK Path"
                $f = Get-Input "Files (space sep)"
                $mount = Get-Input "Mount Point (optional, default ../../../)" $null $false
                $cmp = Read-Host "Enable compression? (Y/N, default N)"
                $procArgs = "create_pak `"$o`""
                $fileArgs = Convert-ToQuotedArgList $f
                if ($fileArgs.Count -gt 0) { $procArgs += " " + ($fileArgs -join " ") }
                if ($mount) { $procArgs += " --mount-point `"$mount`"" }
                if ($cmp -eq "Y") { $procArgs += " --compress" } else { $procArgs += " --no-compress" }
                if ($o -and $f) { Run-Process-Wait $procArgs }
            }
            "2" {
                $o = Get-Input "Output PAK Path"
                $f = Get-Input "File Paths (space sep)"
                $procArgs = "create_companion_pak `"$o`""
                $fileArgs = Convert-ToQuotedArgList $f
                if ($fileArgs.Count -gt 0) { $procArgs += " " + ($fileArgs -join " ") }
                if ($o -and $f) { Run-Process-Wait $procArgs }
            }
            "3" {
                $p = Get-File "PAK File"
                $o = Get-Input "Output Directory" "OutputExtractionDir" $true
                $aes = Get-Input "AES Key (Optional)" "AesKey" $false
                $listOnly = Read-Host "List only (no extraction)? (Y/N, default N)"
                $filt = Get-Input "Filter Patterns (optional, space sep)" $null $false
                
                $procArgs = "extract_pak `"$p`" `"$o`""
                if ($aes) { $procArgs += " --aes `"$aes`"" }
                if ($listOnly -eq "Y") { $procArgs += " --list" }
                if ($filt) {
                    $filterArgs = Convert-ToQuotedArgList $filt
                    if ($filterArgs.Count -gt 0) { $procArgs += " --filter " + ($filterArgs -join " ") }
                }
                
                if ($p -and $o) { Run-Process-Wait $procArgs }
            }
            "4" { return }
        }
    }
}

function Menu-Json {
    while ($true) {
        Show-MenuHeader "JSON Conversion" "Green"
        Write-Host "1. UAsset to JSON (to_json)"
        Write-Host "2. JSON to UAsset (from_json)"
        Write-ColorDivider "DarkGreen" 34
        Write-Host "3. Back" -ForegroundColor Gray
        
        $c = Read-Host "Choose"
        switch ($c) {
            "1" {
                $f = Get-File "Asset Path or Directory"
                $m = Get-File "Mappings Path (.usmap) [Optional]" "UsmapPath"
                $o = Get-File "Output Directory [Optional]"
                $procArgs = "to_json `"$f`""
                if ($m) { $procArgs += " `"$m`"" }
                if ($o) { $procArgs += " `"$o`"" }
                if ($f) { Run-Process-Wait $procArgs }
            }
            "2" {
                $j = Get-File "JSON File or Directory"
                $o = Get-File "Output UAsset Path or Directory"
                $m = Get-File "Mappings Path (.usmap) [Optional]" "UsmapPath"
                $procArgs = "from_json `"$j`" `"$o`""
                if ($m) { $procArgs += " `"$m`"" }
                if ($j -and $o) { Run-Process-Wait $procArgs }
            }
            "3" { return }
        }
    }
}

function Menu-Niagara {
    while ($true) {
        Show-MenuHeader "Niagara / Other" "Magenta"
        Write-Host "1. Niagara List (niagara_list)"
        Write-Host "2. Niagara Details (niagara_details)"
        Write-Host "3. Niagara Edit (niagara_edit)"
        Write-Host "4. Modify Colors Batch (modify_colors)"
        Write-Host "5. Scan ChildBP IsEnemy (scan_childbp_isenemy)"
        Write-ColorDivider "DarkMagenta" 38
        Write-Host "6. Back" -ForegroundColor Gray
        
        $c = Read-Host "Choose"
        switch ($c) {
            "1" {
                $d = Get-Input "Directory"
                $m = Get-File "Mappings Path (.usmap) [Optional]" "UsmapPath"
                $procArgs = "niagara_list `"$d`""
                if ($m) { $procArgs += " `"$m`"" }
                if ($d) { Run-Process-Wait $procArgs }
            }
            "2" {
                $f = Get-File "Asset Path (.uasset)"
                $m = Get-File "Mappings Path (.usmap) [Optional]" "UsmapPath"
                $full = Read-Host "Show full details? (Y/N, default N)"
                $procArgs = "niagara_details `"$f`""
                if ($m) { $procArgs += " `"$m`"" }
                if ($full -eq "Y") { $procArgs += " --full" }
                if ($f) { Run-Process-Wait $procArgs }
            }
            "3" {
                $f = Get-File "Asset Path (.uasset)"
                $rgba = Get-Input "R G B A (space sep)"
                $extra = Get-Input "Extra options (optional, e.g. --export-name Glow --channels rgb)" $null $false
                $m = Get-File "Mappings Path (.usmap) [Optional]" "UsmapPath"
                $procArgs = "niagara_edit `"$f`" $rgba"
                if ($extra) { $procArgs += " $extra" }
                if ($m) { $procArgs += " `"$m`"" }
                if ($f -and $rgba) { Run-Process-Wait $procArgs }
            }
            "4" {
                $d = Get-Input "Directory"
                $m = Get-File "Mappings Path (.usmap)" "UsmapPath"
                $rgba = Get-Input "R G B A (space sep)"
                if ($d -and $m -and $rgba) { Run-Process-Wait "modify_colors `"$d`" `"$m`" $rgba" }
            }
            "5" {
                $p = Get-Input "Paks Directory or Extracted Folder" "GamePaksDir"
                $ext = Read-Host "Is this Extracted? (Y/N)"
                $procArgs = "scan_childbp_isenemy `"$p`""
                if ($ext -eq "Y") { $procArgs += " --extracted" }
                if ($p) { Run-Process-Wait $procArgs }
            }
            "6" { return }
        }
    }
}

function Test-GlobalCommandInstalled {
    $profilePath = $PROFILE
    if (Test-Path $profilePath) {
        $content = Get-Content $profilePath -Raw
        if ($content -match "function uasset-tool") {
            return $true
        }
    }
    return $false
}

function Install-GlobalCommand {
    $profilePath = $PROFILE
    $profileDir = Split-Path $profilePath -Parent
    
    if (-not (Test-Path $profileDir)) {
        New-Item -Path $profileDir -ItemType Directory -Force | Out-Null
    }
    
    if (-not (Test-Path $profilePath)) {
        New-Item -Path $profilePath -ItemType File -Force | Out-Null
    }
    
    if (Test-GlobalCommandInstalled) {
        Write-Warning "Global command 'uasset-tool' is already installed."
        Pause
        return
    }
    
    $ScriptFullPath = Join-Path $ScriptDir "UAssetTool-Shell.ps1"
    $cmd = "`nfunction uasset-tool {`n    & `"$ScriptFullPath`"`n}`n"
    Add-Content -Path $profilePath -Value $cmd
    
    Write-Host "Global command 'uasset-tool' installed successfully!" -ForegroundColor Green
    Write-Host "You may need to restart PowerShell or run '. `$PROFILE' to use it." -ForegroundColor Yellow
    Pause
}

function Menu-Settings {
    while ($true) {
        Show-MenuHeader "Settings" "Cyan"
        Write-Host "1. Set Game Paks Directory"
        Write-Host "   (Current: " -NoNewline -ForegroundColor DarkGray
        Write-Host "$(Get-Config 'GamePaksDir')" -NoNewline -ForegroundColor Yellow
        Write-Host ")" -ForegroundColor DarkGray

        Write-Host "2. Set Default USMAP Path"
        Write-Host "   (Current: " -NoNewline -ForegroundColor DarkGray
        Write-Host "$(Get-Config 'UsmapPath')" -NoNewline -ForegroundColor Yellow
        Write-Host ")" -ForegroundColor DarkGray

        Write-Host "3. Set Default AES Key"
        Write-Host "   (Current: " -NoNewline -ForegroundColor DarkGray
        Write-Host "$(Get-Config 'AesKey')" -NoNewline -ForegroundColor Yellow
        Write-Host ")" -ForegroundColor DarkGray

        Write-Host "4. Set Default Output Extraction Dir"
        Write-Host "   (Current: " -NoNewline -ForegroundColor DarkGray
        Write-Host "$(Get-Config 'OutputExtractionDir')" -NoNewline -ForegroundColor Yellow
        Write-Host ")" -ForegroundColor DarkGray

        $previewStatus = if (Get-ConfigBool "PreviewCommand" $false) { "[Enabled]" } else { "[Disabled]" }
        $previewColor = if (Get-ConfigBool "PreviewCommand" $false) { "Green" } else { "DarkYellow" }
        Write-Host "5. Toggle Command Preview Before Run " -NoNewline
        Write-Host "$previewStatus" -ForegroundColor $previewColor

        $advancedExtractStatus = if (Get-ConfigBool "EnableAdvancedExtractIoStoreArgs" $false) { "[Enabled]" } else { "[Disabled]" }
        $advancedExtractColor = if (Get-ConfigBool "EnableAdvancedExtractIoStoreArgs" $false) { "Green" } else { "DarkYellow" }
        Write-Host "6. Toggle Advanced Extract IoStore Args (script/global) " -NoNewline
        Write-Host "$advancedExtractStatus" -ForegroundColor $advancedExtractColor
        
        $status = if (Test-GlobalCommandInstalled) { "[Installed]" } else { "[Not Installed]" }
        $color = if (Test-GlobalCommandInstalled) { "Green" } else { "Yellow" }
        Write-Host "7. Install Global Command 'uasset-tool' " -NoNewline
        Write-Host "$status" -ForegroundColor $color
        
        Write-ColorDivider "DarkGray" 40
        Write-Host "8. Back" -ForegroundColor Gray
        
        $c = Read-Host "Choose"
        switch ($c) {
            "1" { 
                $v = Read-Host "Enter Game Paks Directory"
                Set-Config "GamePaksDir" $v.Trim('"') 
            }
            "2" { 
                $v = Read-Host "Enter Default USMAP Path"
                Set-Config "UsmapPath" $v.Trim('"') 
            }
            "3" { 
                $v = Read-Host "Enter Default AES Key"
                Set-Config "AesKey" $v.Trim('"') 
            }
            "4" { 
                $v = Read-Host "Enter Default Output Extraction Directory"
                Set-Config "OutputExtractionDir" $v.Trim('"') 
            }
            "5" {
                $enabled = Get-ConfigBool "PreviewCommand" $false
                Set-Config "PreviewCommand" (-not $enabled)
            }
            "6" {
                $enabled = Get-ConfigBool "EnableAdvancedExtractIoStoreArgs" $false
                Set-Config "EnableAdvancedExtractIoStoreArgs" (-not $enabled)
            }
            "7" { Install-GlobalCommand }
            "8" { return }
        }
    }
}

function Run-Menu-Selector {
    while ($true) {
        Show-MenuHeader "Select Command Category" "Cyan"
        Write-Host "1. Asset Operations (detect, fix, dump)"
        Write-Host "2. Zen / IoStore (convert, extract, create)"
        Write-Host "3. PAK Operations (create, extract)"
        Write-Host "4. JSON Conversion"
        Write-Host "5. Niagara / Other"
        Write-ColorDivider "DarkGray" 40
        Write-Host "6. Back to Main Menu" -ForegroundColor Gray
        
        $c = Read-Host "Choose"
        switch ($c) {
            "1" { Menu-AssetOps }
            "2" { Menu-ZenIoStore }
            "3" { Menu-PakOps }
            "4" { Menu-Json }
            "5" { Menu-Niagara }
            "6" { return }
        }
    }
}

function Pause {
    Write-Host ""
    Read-Host "Press Enter to continue..."
}

# Add initial load
Load-Config

# Main Loop
if (-not (Check-Prerequisites)) {
    Pause
    exit
}

do {
    Show-Header
    Write-Host "1. Run Command" -ForegroundColor Green
    Write-Host "2. Download/Update UAssetTool" -ForegroundColor Blue
    Write-Host "3. Settings" -ForegroundColor Cyan
    Write-Host "4. Exit" -ForegroundColor Gray
    Write-Host ""
    $choice = Read-Host "Choose an option (1-4)"

    switch ($choice) {
        "1" { Run-Menu-Selector }
        "2" { Download-PrecompiledTool }
        "3" { Menu-Settings }
        "4" { exit }
        default { Write-Warning "Invalid option. Please try again."; Start-Sleep -Seconds 1 }
    }
} while ($true)
