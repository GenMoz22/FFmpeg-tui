package tui

import (
	"context"
	"strings"
	"time"

	"ffmpeg-tui/src/ffmpeg"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type ActivePanel int

const (
	PanelSidebar ActivePanel = iota
	PanelSubOptions
	PanelConsole
)

type MsgMediaProbed struct{ Info *ffmpeg.MediaInfo }
type MsgFFmpegProgress ffmpeg.ProgressMessage
type MsgResetProgressBar struct{}
type MsgValidationError struct{ Message string }
type MsgError struct{ Err error }

type HistoryItem struct {
	Action string
	Target string
}

type Model struct {
	ActivePanel  ActivePanel
	MediaInfo    *ffmpeg.MediaInfo
	EditOpts     ffmpeg.EditOptions
	CmdInput     textinput.Model
	ProgressBar  progress.Model
	CurrentCmd   string
	ValidationError string

	SidebarIdx      int
	SidebarItems    []string
	SubOptionsIdx   int
	SubOptionsItems []string

	SubMenuFocusIdx int

	ParamInput1 textinput.Model
	ParamInput2 textinput.Model
	ParamInput3 textinput.Model
	ParamInput4 textinput.Model
	ParamInput5 textinput.Model

	History []HistoryItem

	ProgressChan <-chan ffmpeg.ProgressMessage
	CtxCancel    context.CancelFunc
}

func InitialModel(filepath string) Model {
	ti := textinput.New()
	ti.Placeholder = "Edit raw command line flags..."
	ti.Width = 115

	pi1 := textinput.New()
	pi1.Width = 25
	pi2 := textinput.New()
	pi2.Width = 25
	pi3 := textinput.New()
	pi3.Width = 25
	pi4 := textinput.New()
	pi4.Width = 25
	pi5 := textinput.New()
	pi5.Width = 25

	p := progress.New(progress.WithDefaultGradient())

	m := Model{
		ActivePanel:  PanelSidebar,
		SidebarItems: []string{"Crop Video", "Trim Segment", "Split Video", "Audio Normalization", "Export Single Frame", "Burn Subtitles", "Convert Format"},
		CmdInput:     ti,
		ParamInput1:  pi1,
		ParamInput2:  pi2,
		ParamInput3:  pi3,
		ParamInput4:  pi4,
		ParamInput5:  pi5,
		ProgressBar:  p,
		EditOpts: ffmpeg.EditOptions{
			InputFiles: []string{filepath},
		},
		History: []HistoryItem{},
	}
	m.UpdateLiveCommand()
	return m
}

func (m *Model) UpdateLiveCommand() {
	args := ffmpeg.BuildCommand(m.EditOpts)
	m.CurrentCmd = "ffmpeg " + strings.Join(args, " ")
	if !m.CmdInput.Focused() {
		m.CmdInput.SetValue(m.CurrentCmd)
	}
}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		info, err := ffmpeg.ProbeMedia(ctx, m.EditOpts.InputFiles[0])
		if err != nil {
			return MsgError{Err: err}
		}
		return MsgMediaProbed{Info: info}
	}
}
