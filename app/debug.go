package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Diagnostics used to go to stdout, which the alt-screen UI owns — they left
// stray lines on the menu and in the terminal after exit. They now go to a log
// file next to the executable, and only when UAT_TUI_DEBUG is set.

var (
	debugOnce sync.Once
	debugFile *os.File
)

func debugEnabled() bool {
	v := os.Getenv("UAT_TUI_DEBUG")
	return v != "" && v != "0"
}

func debugWriter() *os.File {
	debugOnce.Do(func() {
		exe, err := os.Executable()
		if err != nil {
			return
		}
		path := filepath.Join(filepath.Dir(exe), "uassettool-tui-debug.log")
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return
		}
		debugFile = f
		fmt.Fprintf(f, "\n=== session started %s ===\n", time.Now().Format(time.RFC3339))
	})
	return debugFile
}

func debugln(args ...any) {
	if !debugEnabled() {
		return
	}
	w := debugWriter()
	if w == nil {
		return
	}
	fmt.Fprint(w, time.Now().Format("15:04:05.000")+" ")
	fmt.Fprintln(w, args...)
}
