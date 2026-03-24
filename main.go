package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

var programRef *tea.Program

func main() {
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
