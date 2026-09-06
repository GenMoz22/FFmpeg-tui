package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the TUI layout and UI components.
func (m Model) View() string {
	widthColLeft := 35
	widthColMid := 45
	widthColRight := 55
	totalWidth := widthColLeft + widthColMid + widthColRight + 4

	// 1. Sidebar - Core Modes
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
	sidebarView := sbStyle.Width(widthColLeft).Height(16).Render(
		TitleStyle.Render("CORE MODES") + "\n\n" + strings.Join(sbLines, "\n"),
	)

	// 2. Middle Column - Parameter Controls
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

	switch m.EditOpts.ActiveTool {
		case "trim":
			subContent += "\n\nStart Position:\n" + m.ParamInput1.View() + "\nEnd Position:\n" + m.ParamInput2.View()
		case "split":
			subContent += "\n\nSplit Timestamp:\n" + m.ParamInput1.View()
		case "frame":
			subContent += "\n\nExtract Timestamp:\n" + m.ParamInput1.View()
		case "subtitles":
			subContent += "\n\nSRT File:\n" + m.ParamInput1.View() +
			"\nPosition (top/bottom/center):\n" + m.ParamInput2.View() +
			"\nOffset (px):\n" + m.ParamInput3.View() +
			"\nBackground Color:\n" + m.ParamInput4.View() +
			"\nText Color:\n" + m.ParamInput5.View()
		case "gif":
			subContent += "\n\nStart (Optional):\n" + m.ParamInput1.View() + "\nEnd (Optional):\n" + m.ParamInput2.View()
		case "compress":
			subContent += "\n\n" + lipgloss.NewStyle().Foreground(ColorInactive).Render("CRF 23: Balanced quality / visually lossless.\nCRF 28: High compression for web/social sharing.")
		case "replaceaudio":
			subContent += "\n\nAudio File Path:\n" + m.ParamInput1.View()
	}

	if m.ValidationError != "" {
		subContent += "\n\n" + ErrorBannerStyle.Render("[!] "+m.ValidationError)
	}

	subView := subStyle.Width(widthColMid).Height(16).Render(
		SubTitleStyle.Render("PARAMETER CONTROLS") + "\n\n" + subContent,
	)

	// 3. Right Column - Target Analyzer (ffprobe metadata)
	var infoLines []string
	if m.MediaInfo != nil {
		infoLines = append(infoLines, fmt.Sprintf("[-] File: %s", m.MediaInfo.Path))
		infoLines = append(infoLines, fmt.Sprintf("[-] Duration: %s", m.MediaInfo.FormatDuration()))
		infoLines = append(infoLines, fmt.Sprintf("[-] Size: %.2f MB", float64(m.MediaInfo.Size)/(1024*1024)))
		infoLines = append(infoLines, fmt.Sprintf("[-] Total Bitrate: %d kbps", m.MediaInfo.Bitrate/1000))

		if m.MediaInfo.Video != nil {
			infoLines = append(infoLines, lipgloss.NewStyle().Foreground(ColorAccent).Render("--- Video Stream ---"))
			infoLines = append(infoLines, fmt.Sprintf("  • Codec: %s", strings.ToUpper(m.MediaInfo.Video.Codec)))
			infoLines = append(infoLines, fmt.Sprintf("  • Resolution: %dx%d", m.MediaInfo.Video.Width, m.MediaInfo.Video.Height))
			infoLines = append(infoLines, fmt.Sprintf("  • Framerate: %s FPS", m.MediaInfo.Video.FPS))
			infoLines = append(infoLines, fmt.Sprintf("  • Aspect Ratio: %s", m.MediaInfo.Video.AspectRatio))
		}
		if m.MediaInfo.Audio != nil {
			infoLines = append(infoLines, lipgloss.NewStyle().Foreground(ColorAccent).Render("--- Audio Stream ---"))
			infoLines = append(infoLines, fmt.Sprintf("  • Codec: %s", strings.ToUpper(m.MediaInfo.Audio.Codec)))
			infoLines = append(infoLines, fmt.Sprintf("  • Channels: %d", m.MediaInfo.Audio.Channels))
			infoLines = append(infoLines, fmt.Sprintf("  • Sample Rate: %s Hz", m.MediaInfo.Audio.SampleRate))
		}
	} else {
		infoLines = append(infoLines, "Probing media info with ffprobe...")
	}
	infoView := BoxStyle.Width(widthColRight).Height(16).Render(
		TitleStyle.Render("FFPROBE TARGET ANALYZER") + "\n\n" + strings.Join(infoLines, "\n"),
	)

	// 4. Lower Console Panel
	var consoleStyle = BoxStyle
	if m.ActivePanel == PanelConsole {
		consoleStyle = BoxFocusStyle
	}

	statusText := "READY"
	if m.IsRunning {
		statusText = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB86C")).Bold(true).Render("PROCESSING... (Please wait)")
	}

	helpText := lipgloss.NewStyle().Foreground(ColorInactive).Render(
		"[Tab] Switch Windows  |  [Enter] Confirm & Run  |  [Esc] Quit",
	)
	consoleView := consoleStyle.Width(totalWidth).Height(7).Render(
		TitleStyle.Render("LIVE COMMAND") + "\n\n" +
		m.CmdInput.View() + "\n\n" +
		"Status: " + statusText + "  " + m.ProgressBar.View() + "\n\n" +
		helpText,
	)

	// 5. Session History Panel
	var historyLines []string
	if len(m.History) == 0 {
		historyLines = append(historyLines, lipgloss.NewStyle().Foreground(ColorInactive).Render("No operations completed in this session."))
	} else {
		for _, h := range m.History {
			if h.Success {
				historyLines = append(historyLines, fmt.Sprintf("[+] [%s] Executed successfully ➔ Generated file: %s",
										lipgloss.NewStyle().Foreground(ColorSuccess).Render(h.Action), h.Target))
			} else {
				historyLines = append(historyLines, fmt.Sprintf("[-] [%s] Operation failed ➔ %s",
										lipgloss.NewStyle().Foreground(ColorError).Render(h.Action), h.Target))
			}
		}
	}
	historyView := BoxStyle.Width(totalWidth).Height(5).Render(
		TitleStyle.Render("SESSION HISTORY LOG") + "\n\n" +
		strings.Join(historyLines, "\n"),
	)

	topRow := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, subView, infoView)
	return lipgloss.JoinVertical(lipgloss.Left, topRow, consoleView, historyView)
}
