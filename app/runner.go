package app

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

var programRef *tea.Program

var appVersion = "dev"

const (
	uatLatestReleaseAPI = "https://api.github.com/repos/XzantGaming/UassetToolRivals/releases/latest"
	tuiLatestReleaseAPI = "https://api.github.com/repos/0xSaturno/UAssetToolRivals-TUI/releases/latest"
)

var runningCmdMu sync.Mutex
var runningCmd *exec.Cmd

func Run() {
	fmt.Println("[debug] app version:", currentTUIVersion())
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	programRef = p

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func setRunningCmd(cmd *exec.Cmd) {
	runningCmdMu.Lock()
	runningCmd = cmd
	runningCmdMu.Unlock()
}

func clearRunningCmd(cmd *exec.Cmd) {
	runningCmdMu.Lock()
	if runningCmd == cmd {
		runningCmd = nil
	}
	runningCmdMu.Unlock()
}

func stopRunningTool() error {
	runningCmdMu.Lock()
	cmd := runningCmd
	runningCmdMu.Unlock()
	if cmd == nil || cmd.Process == nil {
		return fmt.Errorf("no running UAssetTool process")
	}
	if programRef != nil {
		programRef.Send(toolOutputMsg{chunk: "\n[panic] stop requested, terminating UAssetTool...\n"})
	}
	if runtime.GOOS == "windows" {
		return exec.Command("taskkill", "/PID", fmt.Sprintf("%d", cmd.Process.Pid), "/T", "/F").Run()
	}
	return cmd.Process.Kill()
}

func exePath() string {
	exe, _ := os.Executable()
	name := "UAssetTool"
	if runtime.GOOS == "windows" {
		name = "UAssetTool.exe"
	}
	return filepath.Join(filepath.Dir(exe), name)
}

func tuiExePath() string {
	exe, _ := os.Executable()
	return exe
}

func currentTUIVersion() string {
	if appVersion != "" && appVersion != "dev" {
		return appVersion
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		fmt.Println("[debug] ReadBuildInfo unavailable")
		return "dev"
	}
	if info != nil && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}

func normalizeVersionTag(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "refs/tags/")
	v = strings.TrimPrefix(v, "v")
	return v
}

func isVersionNewer(current, latest string) bool {
	currentNorm := normalizeVersionTag(current)
	latestNorm := normalizeVersionTag(latest)
	if latestNorm == "" {
		return false
	}
	if currentNorm == "" || currentNorm == "dev" || currentNorm == "(devel)" {
		return latestNorm != currentNorm
	}
	if currentNorm == latestNorm {
		return false
	}
	currentParts := strings.Split(currentNorm, ".")
	latestParts := strings.Split(latestNorm, ".")
	maxLen := len(currentParts)
	if len(latestParts) > maxLen {
		maxLen = len(latestParts)
	}
	for len(currentParts) < maxLen {
		currentParts = append(currentParts, "0")
	}
	for len(latestParts) < maxLen {
		latestParts = append(latestParts, "0")
	}
	for i := 0; i < maxLen; i++ {
		var curN, latN int
		fmt.Sscanf(currentParts[i], "%d", &curN)
		fmt.Sscanf(latestParts[i], "%d", &latN)
		if latN > curN {
			return true
		}
		if latN < curN {
			return false
		}
	}
	return latestNorm > currentNorm
}

func candidateTUIAssetNames() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{
			"uassettool-tui-windows-amd64.exe",
			"uassettool-tui-win-amd64.exe",
			"uassettool-tui-windows-x64.exe",
		}
	case "linux":
		return []string{
			"uassettool-tui-linux-amd64",
			"uassettool-tui-linux-x64",
		}
	}
	return []string{
		fmt.Sprintf("uassettool-tui-%s-%s", runtime.GOOS, runtime.GOARCH),
	}
}

func selectTUIReleaseAsset(info *ReleaseInfo) (string, int64, string, error) {
	if info == nil {
		return "", 0, "", fmt.Errorf("release metadata unavailable")
	}
	for _, candidate := range candidateTUIAssetNames() {
		for _, a := range info.Assets {
			if strings.EqualFold(a.Name, candidate) {
				return a.BrowserDownloadURL, a.Size, a.Name, nil
			}
		}
	}
	needle := strings.ToLower("uassettool-tui")
	for _, a := range info.Assets {
		name := strings.ToLower(a.Name)
		if strings.Contains(name, needle) && strings.Contains(name, runtime.GOOS) {
			return a.BrowserDownloadURL, a.Size, a.Name, nil
		}
	}
	return "", 0, "", fmt.Errorf("no TUI release asset found for %s/%s", runtime.GOOS, runtime.GOARCH)
}

func expectedToolName() string {
	if runtime.GOOS == "windows" {
		return "UAssetTool.exe"
	}
	return "UAssetTool"
}

func candidateAssetNames() []string {
	base := []string{
		fmt.Sprintf("UAssetTool-%s-%s.zip", runtime.GOOS, runtime.GOARCH),
		fmt.Sprintf("UAssetTool-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH),
		fmt.Sprintf("UAssetTool-%s-%s.tgz", runtime.GOOS, runtime.GOARCH),
	}

	switch runtime.GOOS {
	case "windows":
		base = append(base,
			"UAssetTool-win-x64.zip",
			"UAssetTool-windows-x64.zip",
			"UAssetTool-win-amd64.zip",
			"UAssetTool-windows-amd64.zip",
		)
	case "linux":
		base = append(base,
			"UAssetTool-linux-x64.zip",
			"UAssetTool-linux-amd64.zip",
			"UAssetTool-linux-x64.tar.gz",
			"UAssetTool-linux-amd64.tar.gz",
		)
	}

	return base
}

func selectReleaseAsset(info *ReleaseInfo) (string, int64, error) {
	if info == nil {
		return "", 0, fmt.Errorf("release metadata unavailable")
	}

	candidates := candidateAssetNames()
	for _, candidate := range candidates {
		for _, a := range info.Assets {
			if strings.EqualFold(a.Name, candidate) {
				return a.BrowserDownloadURL, a.Size, nil
			}
		}
	}

	toolName := strings.ToLower(expectedToolName())
	osName := runtime.GOOS
	archNames := []string{runtime.GOARCH}
	if runtime.GOARCH == "amd64" {
		archNames = append(archNames, "x64")
	}
	if runtime.GOARCH == "386" {
		archNames = append(archNames, "x86")
	}

	for _, a := range info.Assets {
		name := strings.ToLower(a.Name)
		if !strings.Contains(name, "uassettool") || !strings.Contains(name, osName) {
			continue
		}
		archMatch := false
		for _, arch := range archNames {
			if strings.Contains(name, strings.ToLower(arch)) {
				archMatch = true
				break
			}
		}
		if !archMatch {
			continue
		}
		if strings.HasSuffix(name, ".zip") || strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".tgz") {
			return a.BrowserDownloadURL, a.Size, nil
		}
		if strings.Contains(name, toolName) {
			return a.BrowserDownloadURL, a.Size, nil
		}
	}

	return "", 0, fmt.Errorf("no UAssetTool release asset found for %s/%s", runtime.GOOS, runtime.GOARCH)
}

func ensureExecutable(path string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	return os.Chmod(path, 0755)
}

func extractToolFromZip(archivePath, target string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if filepath.Base(f.Name) != expectedToolName() {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}

		out, err := os.Create(target)
		if err != nil {
			rc.Close()
			return err
		}

		_, err = io.Copy(out, rc)
		rc.Close()
		closeErr := out.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
		return ensureExecutable(target)
	}

	return fmt.Errorf("%s not found in archive", expectedToolName())
}

func extractToolFromTarGz(archivePath, target string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gz, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if hdr == nil || filepath.Base(hdr.Name) != expectedToolName() {
			continue
		}

		out, err := os.Create(target)
		if err != nil {
			return err
		}
		_, err = io.Copy(out, tr)
		closeErr := out.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
		return ensureExecutable(target)
	}

	return fmt.Errorf("%s not found in archive", expectedToolName())
}

func extractToolArchive(archivePath, target, sourceURL string) error {
	lowerURL := strings.ToLower(sourceURL)
	lowerPath := strings.ToLower(archivePath)
	if strings.HasSuffix(lowerURL, ".tar.gz") || strings.HasSuffix(lowerURL, ".tgz") || strings.HasSuffix(lowerPath, ".tar.gz") || strings.HasSuffix(lowerPath, ".tgz") {
		return extractToolFromTarGz(archivePath, target)
	}
	return extractToolFromZip(archivePath, target)
}

func runTool(args string) (string, error) {
	exe := exePath()
	if _, err := os.Stat(exe); os.IsNotExist(err) {
		return "", fmt.Errorf("%s not found. Please download it first", filepath.Base(exe))
	}

	cmd := exec.Command(exe)
	cmd.Args = append([]string{exe}, splitArgs(args)...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}
	setRunningCmd(cmd)
	defer clearRunningCmd(cmd)

	p := programRef
	var combined bytes.Buffer
	var mu sync.Mutex
	var wg sync.WaitGroup

	streamReader := func(r io.Reader) {
		defer wg.Done()
		reader := bufio.NewReader(r)
		for {
			chunk, readErr := reader.ReadString('\n')
			if len(chunk) > 0 {
				mu.Lock()
				combined.WriteString(chunk)
				mu.Unlock()
				if p != nil {
					p.Send(toolOutputMsg{chunk: chunk})
				}
			}
			if readErr != nil {
				if readErr == io.EOF {
					return
				}
				if p != nil {
					p.Send(toolOutputMsg{chunk: fmt.Sprintf("\n[stream error] %v\n", readErr)})
				}
				return
			}
		}
	}

	wg.Add(2)
	go streamReader(stdout)
	go streamReader(stderr)

	err = cmd.Wait()
	wg.Wait()

	mu.Lock()
	out := combined.String()
	mu.Unlock()
	return out, err
}

func splitArgs(s string) []string {
	var args []string
	var current strings.Builder
	inQuote := false
	quoteChar := byte(0)

	for i := 0; i < len(s); i++ {
		c := s[i]
		if inQuote {
			if c == quoteChar {
				inQuote = false
			} else {
				current.WriteByte(c)
			}
		} else if c == '"' || c == '\'' {
			inQuote = true
			quoteChar = c
		} else if c == ' ' || c == '\t' {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		} else {
			current.WriteByte(c)
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args
}

// ── release metadata ────────────────────────────────────────────────────────

type ReleaseInfo struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
	Assets      []struct {
		Name               string `json:"name"`
		Size               int64  `json:"size"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func fetchReleaseInfo() (*ReleaseInfo, error) {
	return fetchReleaseInfoFromURL(uatLatestReleaseAPI)
}

func fetchTUIReleaseInfo() (*ReleaseInfo, error) {
	return fetchReleaseInfoFromURL(tuiLatestReleaseAPI)
}

func fetchReleaseInfoFromURL(apiURL string) (*ReleaseInfo, error) {
	fmt.Println("[debug] fetching release info:", apiURL)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var info ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	fmt.Println("[debug] release info fetched:", info.TagName, info.Name)
	return &info, nil
}

func autoCheckUpdatesCmd(cfg Config) tea.Cmd {
	return func() tea.Msg {
		fmt.Println("[debug] auto update check starting")
		state := updateCheckState{
			UATCurrentVersion: cfg.ToolVersion,
			TUICurrentVersion: currentTUIVersion(),
		}
		uatInfo, uatErr := fetchReleaseInfo()
		if uatErr != nil {
			fmt.Println("[debug] UAT update check failed:", uatErr)
		} else {
			state.UATLatest = uatInfo
			state.UATNeedsUpdate = isVersionNewer(cfg.ToolVersion, uatInfo.TagName)
			fmt.Println("[debug] UAT current/latest/update:", cfg.ToolVersion, uatInfo.TagName, state.UATNeedsUpdate)
		}
		tuiInfo, tuiErr := fetchTUIReleaseInfo()
		if tuiErr != nil {
			fmt.Println("[debug] TUI update check failed:", tuiErr)
		} else {
			state.TUILatest = tuiInfo
			state.TUINeedsUpdate = isVersionNewer(state.TUICurrentVersion, tuiInfo.TagName)
			fmt.Println("[debug] TUI current/latest/update:", state.TUICurrentVersion, tuiInfo.TagName, state.TUINeedsUpdate)
		}
		if uatErr != nil && tuiErr != nil {
			return updateCheckMsg{state: state, err: fmt.Errorf("UAT: %v | TUI: %v", uatErr, tuiErr)}
		}
		return updateCheckMsg{state: state}
	}
}

func performTUISelfUpdate(info *ReleaseInfo) (string, error) {
	if info == nil {
		return "", fmt.Errorf("missing TUI release metadata")
	}
	assetURL, _, assetName, err := selectTUIReleaseAsset(info)
	if err != nil {
		return "", err
	}
	fmt.Println("[debug] TUI self-update asset selected:", assetName, assetURL)
	resp, err := http.Get(assetURL)
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("download returned %d", resp.StatusCode)
	}
	currentExe := tuiExePath()
	tmpNew := currentExe + ".new"
	tmpBak := currentExe + ".old"
	fmt.Println("[debug] TUI self-update current/new/bak:", currentExe, tmpNew, tmpBak)
	out, err := os.Create(tmpNew)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		out.Close()
		return "", err
	}
	if err := out.Close(); err != nil {
		return "", err
	}
	if err := ensureExecutable(tmpNew); err != nil {
		return "", err
	}
	if runtime.GOOS == "windows" {
		cmdLine := fmt.Sprintf("ping 127.0.0.1 -n 3 > nul && copy /Y %q %q > nul && del /F /Q %q > nul 2>nul && start \"\" %q", tmpNew, currentExe, tmpNew, currentExe)
		fmt.Println("[debug] TUI self-update restart command:", cmdLine)
		cmd := exec.Command("cmd", "/C", cmdLine)
		if err := cmd.Start(); err != nil {
			return "", err
		}
		return "TUI updated. Restarting...", nil
	}
	if err := os.Rename(tmpNew, currentExe); err != nil {
		return "", err
	}
	cmd := exec.Command(currentExe)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return "", err
	}
	return "TUI updated. Restarting...", nil
}

func performPromptAction(spec *updatePromptSpec) tea.Cmd {
	return func() tea.Msg {
		if spec == nil {
			return updatePromptResultMsg{action: "none", err: fmt.Errorf("missing prompt spec")}
		}
		fmt.Println("[debug] performing prompt action:", spec.action)
		switch spec.action {
		case "update-uat":
			info, _ := fetchReleaseInfo()
			assetURL, totalSize, err := selectReleaseAsset(info)
			if err != nil {
				return updatePromptResultMsg{action: spec.action, err: err}
			}
			fmt.Println("[debug] UAT prompt accepted; switching to manual update command flow")
			_ = assetURL
			_ = totalSize
			return updatePromptResultMsg{action: spec.action, text: "run-uat-update"}
		case "update-tui":
			text, err := performTUISelfUpdate(spec.release)
			return updatePromptResultMsg{action: spec.action, err: err, text: text}
		default:
			return updatePromptResultMsg{action: spec.action, err: fmt.Errorf("unknown action %s", spec.action)}
		}
	}
}

// ── progress-tracked download ───────────────────────────────────────────────

type downloadProgressMsg struct {
	bytesDownloaded int64
	totalBytes      int64
	speed           float64 // bytes/sec
	eta             time.Duration
	phase           string // "downloading" | "extracting" | "done"
}

type downloadCompleteMsg struct {
	output string
	err    error
	info   *ReleaseInfo
}

func downloadToolCmd() tea.Cmd {
	return func() tea.Msg {
		info, _ := fetchReleaseInfo()

		assetURL, totalSize, err := selectReleaseAsset(info)
		if err != nil {
			return downloadCompleteMsg{err: err, info: info}
		}

		p := programRef

		target := exePath()

		resp, err := http.Get(assetURL)
		if err != nil {
			return downloadCompleteMsg{err: fmt.Errorf("download failed: %w", err), info: info}
		}
		defer resp.Body.Close()

		if totalSize == 0 && resp.ContentLength > 0 {
			totalSize = resp.ContentLength
		}

		tmpFile, err := os.CreateTemp("", "uassettool-*.zip")
		if err != nil {
			return downloadCompleteMsg{err: err, info: info}
		}
		tmpPath := tmpFile.Name()
		defer os.Remove(tmpPath)

		// download with progress
		var downloaded int64
		buf := make([]byte, 32*1024)
		startTime := time.Now()
		lastReport := time.Now()

		for {
			n, readErr := resp.Body.Read(buf)
			if n > 0 {
				_, writeErr := tmpFile.Write(buf[:n])
				if writeErr != nil {
					tmpFile.Close()
					return downloadCompleteMsg{err: writeErr, info: info}
				}
				downloaded += int64(n)

				if time.Since(lastReport) > 80*time.Millisecond {
					elapsed := time.Since(startTime).Seconds()
					speed := float64(downloaded) / elapsed
					var eta time.Duration
					if totalSize > 0 && speed > 0 {
						remaining := float64(totalSize-downloaded) / speed
						eta = time.Duration(remaining * float64(time.Second))
					}
					p.Send(downloadProgressMsg{
						bytesDownloaded: downloaded,
						totalBytes:      totalSize,
						speed:           speed,
						eta:             eta,
						phase:           "downloading",
					})
					lastReport = time.Now()
				}
			}
			if readErr != nil {
				if readErr == io.EOF {
					break
				}
				tmpFile.Close()
				return downloadCompleteMsg{err: readErr, info: info}
			}
		}
		tmpFile.Close()

		// extracting phase
		p.Send(downloadProgressMsg{
			bytesDownloaded: downloaded,
			totalBytes:      totalSize,
			speed:           0,
			phase:           "extracting",
		})

		if err := extractToolArchive(tmpPath, target, assetURL); err != nil {
			return downloadCompleteMsg{err: err, info: info}
		}

		return downloadCompleteMsg{
			output: "Download complete! Saved to " + target,
			info:   info,
		}
	}
}
