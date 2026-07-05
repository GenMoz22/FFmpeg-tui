package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	widthColLeft := 35
	widthColMid := 45
	widthColRight := 51
	totalWidth := widthColLeft + widthColMid + widthColRight + 4

	// Left Column: Core modes navigation menu
	var sbStyle = BoxStyle
	if m.ActivePanel == PanelSidebar {
		sbStyle = BoxFocusStyle
	}
	var sbLines []string
	for i, item := range m.SidebarItems {
		prefix := "  "
		if i == m.SidebarIdx {
			prefix = lipgloss.NewStyle().Foreground(ColorActive).Render(">> ")
		}
		sbLines = append(sbLines, fmt.Sprintf("%s%s", prefix, item))
	}
	sidebarView := sbStyle.Width(widthColLeft).Height(13).Render(
		TitleStyle.Render("CORE MODES") + "\n\n" + strings.Join(sbLines, "\n"),
	)

	// Middle Column: Parameter selection / Form structures
	var subStyle = BoxStyle
	if m.ActivePanel == PanelSubOptions {
		subStyle = BoxFocusStyle
	}
	var subLines []string
	for i, item := range m.SubOptionsItems {
		prefix := "  "
		if i == m.SubOptionsIdx {
			prefix = lipgloss.NewStyle().Foreground(ColorAccent).Render("* ")
		}
		subLines = append(subLines, fmt.Sprintf("%s%s", prefix, item))
	}

	subContent := strings.Join(subLines, "\n")
	if m.EditOpts.ActiveTool == "trim" {
		subContent += "\n\n" +
		"Start Position:\n" + m.ParamInput1.View() + "\n" +
		"End Position:\n" + m.ParamInput2.View()
	} else if m.EditOpts.ActiveTool == "split" {
		subContent += "\n\n" +
		"Split At Timestamp:\n" + m.ParamInput1.View()
	} else if m.EditOpts.ActiveTool == "frame" {
		subContent += "\n\n" +
		"Frame Extraction Target:\n" + m.ParamInput1.View()
	} else if m.EditOpts.ActiveTool == "subtitles" {
		subContent += "\n\n" +
		"SRT File Path: " + m.ParamInput1.View() + "\n" +
		"Pos (top/bot): " + m.ParamInput2.View() + "\n" +
		"Offset (px):   " + m.ParamInput3.View() + "\n" +
		"Bg Color:      " + m.ParamInput4.View() + "\n" +
		"Text Color:    " + m.ParamInput5.View()
	}

	if m.ValidationError != "" {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Bold(true)
		subContent += "\n\n" + errStyle.Render("[!] "+m.ValidationError)
	}

	subView := subStyle.Width(widthColMid).Height(13).Render(
		SubTitleStyle.Render("PARAMETER CONTROLS") + "\n\n" + subContent,
	)

	// Right Column: Target analysis panel info tracer
	var infoLines []string
	if m.MediaInfo != nil {
		infoLines = append(infoLines, fmt.Sprintf("[-] File: %s", m.MediaInfo.Path))
		infoLines = append(infoLines, fmt.Sprintf("[-] Duration: %s", m.MediaInfo.FormatDuration()))
		if m.MediaInfo.Video != nil {
			infoLines = append(infoLines, fmt.Sprintf("[-] Video: %s (%dx%d)", strings.ToUpper(m.MediaInfo.Video.Codec), m.MediaInfo.Video.Width, m.MediaInfo.Video.Height))
		}
		if m.MediaInfo.Audio != nil {
			infoLines = append(infoLines, fmt.Sprintf("[-] Audio: %s", strings.ToUpper(m.MediaInfo.Audio.Codec)))
		}
	} else {
		infoLines = append(infoLines, "Parsing resource stream metadata...")
	}
	infoView := BoxStyle.Width(widthColRight).Height(13).Render(
		TitleStyle.Render("TARGET ANALYZER") + "\n\n" + strings.Join(infoLines, "\n"),
	)

	// Lower Panel 1: Live Command Instruction Engine Console
	var consoleStyle = BoxStyle
	if m.ActivePanel == PanelConsole {
		consoleStyle = BoxFocusStyle
	}
	helpText := lipgloss.NewStyle().Foreground(ColorInactive).Render(
		"[Tab] Switch Windows  |  [Enter] Select Field or Execute Pipeline  |  [q] Abort Task",
	)
	consoleView := consoleStyle.Width(totalWidth).Height(7).Render(
		TitleStyle.Render("LIVE COMMAND") + "\n\n" +
		m.CmdInput.View() + "\n\n" +
		"Status: " + m.ProgressBar.View() + "\n\n" +
		helpText,
	)

	// Lower Panel 2: Operational Log History Tracker
	var historyLines []string
	if len(m.History) == 0 {
		historyLines = append(historyLines, lipgloss.NewStyle().Foreground(ColorInactive).Render("No actions executed in this session."))
	} else {
		for _, h := range m.History {
			historyLines = append(historyLines, fmt.Sprintf("[+] [%s] Job closed ➔ Target: %s",
									lipgloss.NewStyle().Foreground(ColorSuccess).Render(h.Action), h.Target))
		}
	}
	historyView := BoxStyle.Width(totalWidth).Height(5).Render(
		TitleStyle.Render("SESSION RUNTIME HISTORY") + "\n\n" +
		strings.Join(historyLines, "\n"),
	)

	topRow := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, subView, infoView)
	return lipgloss.JoinVertical(lipgloss.Left, topRow, consoleView, historyView)
}
