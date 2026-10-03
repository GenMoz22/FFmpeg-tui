package tui

import (
	"context"
	"fmt"
	"strconv"
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

type ModeKind int

const (
	ModeCrop ModeKind = iota
	ModeTrim
	ModeSplit
	ModeAudioNorm
	ModeFrame
	ModeSubtitles
	ModeConvert
	ModeCompress
	ModeSepAudio
	ModeMetadata
	ModeReplaceAudio
)

type SidebarOption struct {
	Title string
	Kind  ModeKind
}

var allSidebarOptions = []SidebarOption{
	{Title: "Crop Video", Kind: ModeCrop},
	{Title: "Trim Segment", Kind: ModeTrim},
	{Title: "Split Video", Kind: ModeSplit},
	{Title: "Audio Normalization", Kind: ModeAudioNorm},
	{Title: "Frame Export", Kind: ModeFrame},
	{Title: "Burn Subtitles", Kind: ModeSubtitles},
	{Title: "Convert Format", Kind: ModeConvert},
	{Title: "CRF Compression", Kind: ModeCompress},
	{Title: "Separate Video/Audio", Kind: ModeSepAudio},
	{Title: "Strip Metadata", Kind: ModeMetadata},
	{Title: "Replace Audio Track", Kind: ModeReplaceAudio},
}

type Model struct {
	SidebarItems    []SidebarOption
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

	opts := ffmpeg.EditOptions{
		InputFiles: []string{filePath},
		ActiveTool: "crop",
		CropPreset: "9:16",
	}

	m := Model{
		SidebarItems:    allSidebarOptions,
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

// FormatCommandLine quotes arguments containing spaces to build a valid CLI command representation.
func FormatCommandLine(args []string) string {
	quoted := make([]string, len(args))
	for i, arg := range args {
		if strings.Contains(arg, " ") && !strings.HasPrefix(arg, "\"") && !strings.HasPrefix(arg, "'") {
			quoted[i] = fmt.Sprintf("%q", arg)
		} else {
			quoted[i] = arg
		}
	}
	return strings.Join(quoted, " ")
}

// UpdateLiveCommand synchronizes the active command string shown in the UI with current options.
func (m *Model) UpdateLiveCommand() {
	args := ffmpeg.BuildCommand(m.EditOpts)

	if len(args) == 0 {
		m.CmdInput.SetValue("ffmpeg")
		return
	}

	if len(args) == 2 && args[0] == "__SHELL_CMD__" {
		m.CmdInput.SetValue(fmt.Sprintf("ffmpeg %s", args[1]))
		return
	}

	fullArgs := FormatCommandLine(args)
	m.CmdInput.SetValue(fmt.Sprintf("ffmpeg %s", fullArgs))
}

// FilterSidebarForMedia updates available options based on media stream types.
func (m *Model) FilterSidebarForMedia() {
	if m.MediaInfo == nil {
		return
	}

	hasVideo := m.MediaInfo.Video != nil
	hasAudio := m.MediaInfo.Audio != nil

	var filtered []SidebarOption
	for _, opt := range allSidebarOptions {
		switch opt.Kind {
			case ModeCrop, ModeFrame, ModeSubtitles, ModeCompress, ModeSepAudio, ModeReplaceAudio:
				if hasVideo {
					filtered = append(filtered, opt)
				}
			case ModeAudioNorm:
				if hasAudio {
					filtered = append(filtered, opt)
				}
			case ModeSplit:
				title := "Split Video"
				if !hasVideo {
					title = "Split Audio"
				}
				filtered = append(filtered, SidebarOption{Title: title, Kind: ModeSplit})
			case ModeTrim, ModeConvert, ModeMetadata:
				filtered = append(filtered, opt)
		}
	}

	if len(filtered) > 0 {
		m.SidebarItems = filtered
		if m.SidebarIdx >= len(m.SidebarItems) {
			m.SidebarIdx = 0
		}
		m.SelectSidebarItem(m.SidebarIdx)
	}
}

// SelectSidebarItem initializes sub-options and edit state when a sidebar item is selected.
func (m *Model) SelectSidebarItem(idx int) {
	if idx < 0 || idx >= len(m.SidebarItems) {
		return
	}

	m.SubOptionsIdx = 0
	m.SubMenuFocusIdx = 0
	m.ParamInput1.SetValue("")
	m.ParamInput2.SetValue("")
	m.ParamInput3.SetValue("")
	m.ParamInput4.SetValue("")
	m.ParamInput5.SetValue("")

	item := m.SidebarItems[idx]
	switch item.Kind {
		case ModeCrop:
			m.EditOpts.ActiveTool = "crop"
			m.SubOptionsItems = []string{"9:16 (Shorts/Reels)", "1:1 (Square)", "16:9 (Widescreen)", "Auto Crop Black Bars"}
		case ModeTrim:
			m.EditOpts.ActiveTool = "trim"
			m.SubOptionsItems = []string{"Parameters Setup"}
			m.ParamInput1.Placeholder = "Start (e.g., 00:00:05 or 5)"
			m.ParamInput2.Placeholder = "End (e.g., 00:00:30 or 30)"
			if m.MediaInfo != nil {
				m.ParamInput1.SetValue("0")
				m.ParamInput2.SetValue(strconv.FormatFloat(m.MediaInfo.Duration.Seconds(), 'f', 2, 64))
			}
		case ModeSplit:
			m.EditOpts.ActiveTool = "split"
			m.SubOptionsItems = []string{"Parameters Setup"}
			m.ParamInput1.Placeholder = "Split point (e.g., 00:01:00 or 60)"
			m.ParamInput1.SetValue("0")
		case ModeAudioNorm:
			m.EditOpts.ActiveTool = "audio"
			m.EditOpts.NormalizeAudio = true
			m.SubOptionsItems = []string{"Loudnorm Equalizer Active"}
		case ModeFrame:
			m.EditOpts.ActiveTool = "frame"
			m.SubOptionsItems = []string{"Parameters Setup"}
			m.ParamInput1.Placeholder = "Timestamp (e.g., 00:00:05 or 5)"
			m.ParamInput1.SetValue("0")
		case ModeSubtitles:
			m.EditOpts.ActiveTool = "subtitles"
			m.SubOptionsItems = []string{"Hardburn Properties"}
			m.ParamInput1.Placeholder = "File Path (e.g., sub.srt)"
			m.ParamInput2.Placeholder = "Position (bottom/top/center)"
			m.ParamInput2.SetValue("bottom")
			m.ParamInput3.Placeholder = "Offset (e.g., 10)"
			m.ParamInput3.SetValue("10")
			m.ParamInput4.Placeholder = "Bg Color (black/none)"
			m.ParamInput4.SetValue("black")
			m.ParamInput5.Placeholder = "Text Color (white/yellow)"
			m.ParamInput5.SetValue("white")
		case ModeConvert:
			m.EditOpts.ActiveTool = "convert"
			if m.MediaInfo != nil && m.MediaInfo.Video == nil {
				m.SubOptionsItems = []string{"MP3", "WAV", "AAC", "FLAC", "OGG"}
			} else {
				m.SubOptionsItems = []string{"MP4", "MKV", "MOV", "AVI", "MP3 Pure Audio", "Animated GIF"}
			}
		case ModeCompress:
			m.EditOpts.ActiveTool = "compress"
			m.SubOptionsItems = []string{"CRF 23 (High Quality / Balanced)", "CRF 28 (High Compression / Social Sharing)"}
		case ModeSepAudio:
			m.EditOpts.ActiveTool = "sepaudio"
			m.SubOptionsItems = []string{"Extract Video (no audio) & Audio (.mp3 HQ)"}
		case ModeMetadata:
			m.EditOpts.ActiveTool = "metadata"
			m.SubOptionsItems = []string{"Remove EXIF/Stream Metadata Tags"}
		case ModeReplaceAudio:
			m.EditOpts.ActiveTool = "replaceaudio"
			m.SubOptionsItems = []string{"Audio Track Override"}
			m.ParamInput1.Placeholder = "New Audio File Path (.mp3, .wav, .m4a)"
	}
}
