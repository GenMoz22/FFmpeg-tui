package main

import (
	"fmt"
	"os"

	"ffmpeg-tui/src/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Error: Missing input media file.")
		fmt.Println("Usage: ffmpeg-tui <path_to_media_file>")
		os.Exit(1)
	}

	filePath := os.Args[1]
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("Error: The file '%s' does not exist.\n", filePath)
		os.Exit(1)
	}

	p := tea.NewProgram(tui.InitialModel(filePath), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Fatal TUI runtime error: %v\n", err)
		os.Exit(1)
	}
}
