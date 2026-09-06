package tui

import (
	"context"
	"fmt"
	"strings"

	"ffmpeg-tui/internal/ffmpeg"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Panel int

const (
	PanelSidebar Panel = iota
	PanelSubOptions
	PanelConsole
)

type HistoryItem struct {
	Action  string
	Target  string
	Success bool
}

type Model struct {
	SidebarItems    []string
	SidebarIdx      int
	SubOptionsItems []string
	SubOptionsIdx   int
	SubMenuFocusIdx int

	ActivePanel Panel

	ParamInput1 textinput.Model
	ParamInput2 textinput.Model
	ParamInput3 textinput.Model
	ParamInput4 textinput.Model
	ParamInput5 textinput.Model

	CmdInput textinput.Model

	EditOpts  ffmpeg.EditOptions
	MediaInfo *ffmpeg.MediaInfo

	ProgressBar progress.Model
	IsRunning   bool
	History     []HistoryItem

	ValidationError string
	ProgressChan    <-chan ffmpeg.ProgressMessage
	CtxCancel       func()
}

// InitialModel constructs and initializes the primary application model state.
func InitialModel(filePath string) Model {
	p1 := textinput.New()
	p2 := textinput.New()
	p3 := textinput.New()
	p4 := textinput.New()
	p5 := textinput.New()

	cmdIn := textinput.New()
	cmdIn.CharLimit = 1024

	pb := progress.New(progress.WithDefaultGradient())

	sidebar := []string{
		"Crop Video",
		"Trim Segment",
		"Split Video",
		"Audio Normalization",
		"Frame Export",
		"Burn Subtitles",
		"Convert Format",
		"CRF Compression",
		"Separate Video/Audio",
		"Strip Metadata",
		"Replace Audio Track",
	}

	opts := ffmpeg.EditOptions{
		InputFiles: []string{filePath},
		ActiveTool: "crop",
		CropPreset: "9:16",
	}

	m := Model{
		SidebarItems:    sidebar,
		SidebarIdx:      0,
		SubOptionsItems: []string{"9:16 (Shorts/Reels)", "1:1 (Square)", "16:9 (Widescreen)", "Auto Crop Black Bars"},
		SubOptionsIdx:   0,
		ActivePanel:     PanelSidebar,
		ParamInput1:     p1,
		ParamInput2:     p2,
		ParamInput3:     p3,
		ParamInput4:     p4,
		ParamInput5:     p5,
		CmdInput:        cmdIn,
		EditOpts:        opts,
		ProgressBar:     pb,
		IsRunning:       false,
		History:         make([]HistoryItem, 0),
	}

	m.UpdateLiveCommand()
	return m
}

// Init initializes the Bubble Tea model and kicks off media probing asynchronously.
func (m Model) Init() tea.Cmd {
	filePath := ""
	if len(m.EditOpts.InputFiles) > 0 {
		filePath = m.EditOpts.InputFiles[0]
	}

	return tea.Batch(
		textinput.Blink,
		probeMediaCmd(filePath),
	)
}

// probeMediaCmd triggers asynchronous FFprobe inspection upon startup.
func probeMediaCmd(path string) tea.Cmd {
	return func() tea.Msg {
		if path == "" {
			return nil
		}
		info, err := ffmpeg.ProbeMedia(context.Background(), path)
		if err != nil {
			return MsgError{Err: err}
		}
		return MsgMediaProbed{Info: info}
	}
}

// UpdateLiveCommand synchronizes the active command string shown in the UI with current options.
func (m *Model) UpdateLiveCommand() {
	args := ffmpeg.BuildCommand(m.EditOpts)

	if len(args) == 0 {
		m.CmdInput.SetValue("ffmpeg")
		return
	}

	fullArgs := strings.Join(args, " ")

	if strings.Contains(fullArgs, "&&") {
		m.CmdInput.SetValue(fmt.Sprintf("ffmpeg %s", fullArgs))
	} else {
		m.CmdInput.SetValue(fmt.Sprintf("ffmpeg %s", fullArgs))
	}
}
