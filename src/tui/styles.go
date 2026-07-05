package tui

import "github.com/charmbracelet/lipgloss"

var (
	ColorActive   = lipgloss.Color("#25D366") // Green focus
	ColorInactive = lipgloss.Color("#6272A4") // Muted
	ColorBorder   = lipgloss.Color("#44475A")
	ColorAccent   = lipgloss.Color("#8BE9FD") // Cyan for submenus
	ColorSuccess  = lipgloss.Color("#50FA7B") // History success

	BoxStyle = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder()).
	BorderForeground(ColorBorder).
	Padding(0, 1)

	BoxFocusStyle = BoxStyle.Copy().
	BorderForeground(ColorActive)

	TitleStyle = lipgloss.NewStyle().
	Foreground(ColorActive).
	Bold(true)

	SubTitleStyle = lipgloss.NewStyle().
	Foreground(ColorAccent).
	Bold(true)
)
