package main

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
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func exePath() string {
	exe, _ := os.Executable()
	name := "UAssetTool"
	if runtime.GOOS == "windows" {
		name = "UAssetTool.exe"
	}
	return filepath.Join(filepath.Dir(exe), name)
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
	apiURL := "https://api.github.com/repos/XzantGaming/UassetToolRivals/releases/latest"
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
	return &info, nil
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
