package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"ffmpeg-tui/src/ffmpeg"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

// listenToProgress listens for execution progress updates from the background FFmpeg process
// and streams them back as Bubble Tea messages to trigger UI re-renders.
func listenToProgress(ch <-chan ffmpeg.ProgressMessage) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return MsgFFmpegProgress(msg)
	}
}

// delayedReset schedules a timer to reset the progress bar back to 0% after task completion.
func delayedReset() tea.Cmd {
	return tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
		return MsgResetProgressBar{}
	})
}

// syncInputsToEditOpts binds current active input text fields into the underlying EditOptions model.
func (m *Model) syncInputsToEditOpts() {
	switch m.EditOpts.ActiveTool {
		case "trim":
			m.EditOpts.TrimStart = strings.TrimSpace(m.ParamInput1.Value())
			m.EditOpts.TrimEnd = strings.TrimSpace(m.ParamInput2.Value())
		case "split":
			m.EditOpts.SplitPoint = strings.TrimSpace(m.ParamInput1.Value())
		case "frame":
			m.EditOpts.ExtractFrame = strings.TrimSpace(m.ParamInput1.Value())
		case "subtitles":
			m.EditOpts.SubPath = strings.TrimSpace(m.ParamInput1.Value())
			m.EditOpts.SubPos = strings.TrimSpace(m.ParamInput2.Value())
			m.EditOpts.SubOffset = strings.TrimSpace(m.ParamInput3.Value())
			m.EditOpts.SubBgColor = strings.TrimSpace(m.ParamInput4.Value())
			m.EditOpts.SubTextColor = strings.TrimSpace(m.ParamInput5.Value())
		case "gif":
			m.EditOpts.TrimStart = strings.TrimSpace(m.ParamInput1.Value())
			m.EditOpts.TrimEnd = strings.TrimSpace(m.ParamInput2.Value())
		case "replaceaudio":
			m.EditOpts.AudioFilePath = strings.TrimSpace(m.ParamInput1.Value())
	}
}

// Update handles application state mutations, keyboard navigation, and asynchronous commands.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	m.ValidationError = ""

	switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
				case "ctrl+c", "esc":
					if m.CtxCancel != nil {
						m.CtxCancel()
					}
					return m, tea.Quit

				case "tab":
					// Tab key handles panel focus rotation across Sidebar, SubOptions, and Console
					if m.ActivePanel == PanelSidebar {
						m.ActivePanel = PanelSubOptions
						m.SubMenuFocusIdx = 0
						if len(m.SubOptionsItems) > 0 && m.SidebarIdx != 0 && m.SidebarIdx != 6 && m.SidebarIdx != 3 && m.SidebarIdx != 7 && m.SidebarIdx != 8 && m.SidebarIdx != 9 {
							cmds = append(cmds, m.ParamInput1.Focus())
						}
					} else if m.ActivePanel == PanelSubOptions {
						m.ParamInput1.Blur()
						m.ParamInput2.Blur()
						m.ParamInput3.Blur()
						m.ParamInput4.Blur()
						m.ParamInput5.Blur()

						if m.SidebarIdx == 1 { // Trim tool multi-field cycling
							if m.SubMenuFocusIdx == 0 {
								m.SubMenuFocusIdx = 1
								cmds = append(cmds, m.ParamInput2.Focus())
							} else {
								m.ActivePanel = PanelConsole
								cmds = append(cmds, m.CmdInput.Focus())
							}
						} else if m.SidebarIdx == 5 { // Subtitles multi-field cycling
							if m.SubMenuFocusIdx < 4 {
								m.SubMenuFocusIdx++
								switch m.SubMenuFocusIdx {
									case 1:
										cmds = append(cmds, m.ParamInput2.Focus())
									case 2:
										cmds = append(cmds, m.ParamInput3.Focus())
									case 3:
										cmds = append(cmds, m.ParamInput4.Focus())
									case 4:
										cmds = append(cmds, m.ParamInput5.Focus())
								}
							} else {
								m.ActivePanel = PanelConsole
								cmds = append(cmds, m.CmdInput.Focus())
							}
						} else if m.SidebarIdx == 6 && m.SubOptionsIdx == 5 { // GIF conversion sub-option
							if m.SubMenuFocusIdx == 0 {
								m.SubMenuFocusIdx = 1
								cmds = append(cmds, m.ParamInput2.Focus())
							} else {
								m.ActivePanel = PanelConsole
								cmds = append(cmds, m.CmdInput.Focus())
							}
						} else {
							m.ActivePanel = PanelConsole
							cmds = append(cmds, m.CmdInput.Focus())
						}
					} else {
						m.CmdInput.Blur()
						m.ActivePanel = PanelSidebar
					}

									case "up", "k":
										if m.ActivePanel == PanelSidebar && m.SidebarIdx > 0 {
											m.SidebarIdx--
										} else if m.ActivePanel == PanelSubOptions && (m.SidebarIdx == 0 || m.SidebarIdx == 6 || m.SidebarIdx == 7) && m.SubOptionsIdx > 0 {
											m.SubOptionsIdx--
										}

									case "down", "j":
										if m.ActivePanel == PanelSidebar && m.SidebarIdx < len(m.SidebarItems)-1 {
											m.SidebarIdx++
										} else if m.ActivePanel == PanelSubOptions && (m.SidebarIdx == 0 || m.SidebarIdx == 6 || m.SidebarIdx == 7) && m.SubOptionsIdx < len(m.SubOptionsItems)-1 {
											m.SubOptionsIdx++
										}

									case "enter":
										if m.ActivePanel == PanelSidebar {
											// Reset inputs and focus state when switching sidebar options
											m.SubOptionsIdx = 0
											m.SubMenuFocusIdx = 0
											m.ParamInput1.SetValue("")
											m.ParamInput2.SetValue("")
											m.ParamInput3.SetValue("")
											m.ParamInput4.SetValue("")
											m.ParamInput5.SetValue("")

											switch m.SidebarIdx {
												case 0: // Crop Video
													m.EditOpts.ActiveTool = "crop"
													m.SubOptionsItems = []string{"9:16 (Shorts/Reels)", "1:1 (Square)", "16:9 (Widescreen)", "Auto Crop Black Bars"}
													m.ActivePanel = PanelSubOptions
												case 1: // Trim Segment
													m.EditOpts.ActiveTool = "trim"
													m.SubOptionsItems = []string{"Parameters Setup"}
													m.ParamInput1.Placeholder = "Start (e.g., 0)"
													m.ParamInput2.Placeholder = "End (e.g., 30)"
													if m.MediaInfo != nil {
														m.ParamInput1.SetValue("0")
														m.ParamInput2.SetValue(strconv.FormatFloat(m.MediaInfo.Duration.Seconds(), 'f', 2, 64))
													}
													m.ActivePanel = PanelSubOptions
													cmds = append(cmds, m.ParamInput1.Focus())
												case 2: // Split Video
													m.EditOpts.ActiveTool = "split"
													m.SubOptionsItems = []string{"Parameters Setup"}
													m.ParamInput1.Placeholder = "Split point in seconds (e.g., 60)"
													m.ParamInput1.SetValue("0")
													m.ActivePanel = PanelSubOptions
													cmds = append(cmds, m.ParamInput1.Focus())
												case 3: // Audio Normalization
													m.EditOpts.ActiveTool = "audio"
													m.EditOpts.NormalizeAudio = true
													m.SubOptionsItems = []string{"Loudnorm Equalizer Active"}
												case 4: // Frame Export
													m.EditOpts.ActiveTool = "frame"
													m.SubOptionsItems = []string{"Parameters Setup"}
													m.ParamInput1.Placeholder = "Timestamp (e.g., 5)"
													m.ParamInput1.SetValue("0")
													m.ActivePanel = PanelSubOptions
													cmds = append(cmds, m.ParamInput1.Focus())
												case 5: // Burn Subtitles
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
													m.ActivePanel = PanelSubOptions
													cmds = append(cmds, m.ParamInput1.Focus())
												case 6: // Convert Format
													m.EditOpts.ActiveTool = "convert"
													m.SubOptionsItems = []string{"MP4", "MKV", "MOV", "AVI", "MP3 Pure Audio", "Animated GIF"}
													m.ActivePanel = PanelSubOptions
												case 7: // CRF Compression
													m.EditOpts.ActiveTool = "compress"
													m.SubOptionsItems = []string{"CRF 23 (High Quality / Balanced)", "CRF 28 (High Compression / Social Sharing)"}
													m.ActivePanel = PanelSubOptions
												case 8: // Separate Video/Audio
													m.EditOpts.ActiveTool = "sepaudio"
													m.SubOptionsItems = []string{"Extract Video (no audio) & Audio (.mp3 HQ)"}
												case 9: // Strip Metadata
													m.EditOpts.ActiveTool = "metadata"
													m.SubOptionsItems = []string{"Remove EXIF/Stream Metadata Tags"}
												case 10: // Replace Audio Track
													m.EditOpts.ActiveTool = "replaceaudio"
													m.SubOptionsItems = []string{"Audio Track Override"}
													m.ParamInput1.Placeholder = "New Audio File Path (.mp3, .wav, .m4a)"
													m.ActivePanel = PanelSubOptions
													cmds = append(cmds, m.ParamInput1.Focus())
											}
											m.syncInputsToEditOpts()
											m.UpdateLiveCommand()

										} else if m.ActivePanel == PanelSubOptions {
											maxDuration := 0.0
											if m.MediaInfo != nil {
												maxDuration = m.MediaInfo.Duration.Seconds()
											}

											// Field transition logic within configuration steps
											if m.EditOpts.ActiveTool == "trim" && m.SubMenuFocusIdx == 0 {
												m.SubMenuFocusIdx = 1
												m.ParamInput1.Blur()
												cmds = append(cmds, m.ParamInput2.Focus())
												m.syncInputsToEditOpts()
												m.UpdateLiveCommand()
												return m, tea.Batch(cmds...)
											} else if m.EditOpts.ActiveTool == "gif" && m.SubMenuFocusIdx == 0 {
												m.SubMenuFocusIdx = 1
												m.ParamInput1.Blur()
												cmds = append(cmds, m.ParamInput2.Focus())
												m.syncInputsToEditOpts()
												m.UpdateLiveCommand()
												return m, tea.Batch(cmds...)
											} else if m.EditOpts.ActiveTool == "subtitles" && m.SubMenuFocusIdx < 4 {
												m.SubMenuFocusIdx++
												m.ParamInput1.Blur()
												m.ParamInput2.Blur()
												m.ParamInput3.Blur()
												m.ParamInput4.Blur()
												m.ParamInput5.Blur()
												switch m.SubMenuFocusIdx {
													case 1:
														cmds = append(cmds, m.ParamInput2.Focus())
													case 2:
														cmds = append(cmds, m.ParamInput3.Focus())
													case 3:
														cmds = append(cmds, m.ParamInput4.Focus())
													case 4:
														cmds = append(cmds, m.ParamInput5.Focus())
												}
												m.syncInputsToEditOpts()
												m.UpdateLiveCommand()
												return m, tea.Batch(cmds...)
											}

											// Validation logic and parameter execution setup
											switch m.SidebarIdx {
												case 0: // Crop Video
													if m.SubOptionsIdx == 3 {
														m.EditOpts.ActiveTool = "autocrop"
													} else {
														m.EditOpts.ActiveTool = "crop"
														presets := []string{"9:16", "1:1", "16:9"}
														m.EditOpts.CropPreset = presets[m.SubOptionsIdx]
													}
												case 1: // Trim Segment
													t1, _ := ffmpeg.ParseDurationString(m.ParamInput1.Value())
													t2, _ := ffmpeg.ParseDurationString(m.ParamInput2.Value())
													if maxDuration > 0 && (t1 > maxDuration || t2 > maxDuration || t1 >= t2) {
														m.ValidationError = fmt.Sprintf("OUT OF BOUNDS: Video length is %.2fs max.", maxDuration)
														return m, tea.Batch(cmds...)
													}
													m.EditOpts.TrimStart = m.ParamInput1.Value()
													m.EditOpts.TrimEnd = m.ParamInput2.Value()
												case 2: // Split Video
													sp, _ := ffmpeg.ParseDurationString(m.ParamInput1.Value())
													if maxDuration > 0 && (sp >= maxDuration || sp <= 0) {
														m.ValidationError = fmt.Sprintf("OUT OF BOUNDS: Split point must be between 0s and %.2fs.", maxDuration)
														return m, tea.Batch(cmds...)
													}
													m.EditOpts.SplitPoint = m.ParamInput1.Value()
												case 6: // Convert Format
													if m.SubOptionsIdx == 5 {
														m.EditOpts.ActiveTool = "gif"
														m.SubOptionsItems = []string{"Convert to Animated GIF"}
														m.ParamInput1.Placeholder = "Start Time (leave empty for full video)"
														m.ParamInput2.Placeholder = "End Time (leave empty for full video)"
														cmds = append(cmds, m.ParamInput1.Focus())
														m.syncInputsToEditOpts()
														m.UpdateLiveCommand()
														return m, tea.Batch(cmds...)
													} else {
														m.EditOpts.ActiveTool = "convert"
														formats := []string{"mp4", "mkv", "mov", "avi", "mp3"}
														m.EditOpts.TargetFormat = formats[m.SubOptionsIdx]
													}
												case 7: // Compress
													if m.SubOptionsIdx == 0 {
														m.EditOpts.CRFValue = "23"
													} else {
														m.EditOpts.CRFValue = "28"
													}
												case 10: // Replace Audio
													audioPath := strings.TrimSpace(m.ParamInput1.Value())
													if audioPath == "" {
														m.ValidationError = "ERROR: Missing target audio file path."
														return m, tea.Batch(cmds...)
													}
													m.EditOpts.AudioFilePath = audioPath
											}

											m.syncInputsToEditOpts()
											m.UpdateLiveCommand()
											m.ActivePanel = PanelConsole
											cmds = append(cmds, m.CmdInput.Focus())
											m.ParamInput1.Blur()
											m.ParamInput2.Blur()
											m.ParamInput3.Blur()
											m.ParamInput4.Blur()
											m.ParamInput5.Blur()

										} else if m.ActivePanel == PanelConsole {
											if m.IsRunning {
												return m, tea.Batch(cmds...)
											}

											m.IsRunning = true

											var ctx context.Context
											ctx, m.CtxCancel = context.WithCancel(context.Background())

											durationSec := 10.0
											if m.MediaInfo != nil {
												durationSec = m.MediaInfo.Duration.Seconds()
											}

											customArgs := strings.Split(strings.TrimPrefix(m.CmdInput.Value(), "ffmpeg "), " ")
											m.ProgressChan = ffmpeg.ExecuteFFmpeg(ctx, customArgs, durationSec)
											cmds = append(cmds, listenToProgress(m.ProgressChan))
										}
			}

												case MsgMediaProbed:
													m.MediaInfo = msg.Info
													m.UpdateLiveCommand()

												case MsgError:
													m.ValidationError = fmt.Sprintf("Error probing file: %v", msg.Err)

												case MsgFFmpegProgress:
													if msg.Err != nil {
														m.IsRunning = false
														m.ValidationError = fmt.Sprintf("FFmpeg execution failed: %v", msg.Err)
														return m, tea.Batch(cmds...)
													}

													progCmd := m.ProgressBar.SetPercent(msg.Percent)
													cmds = append(cmds, progCmd)

													if !msg.Done {
														cmds = append(cmds, listenToProgress(m.ProgressChan))
													} else {
														m.IsRunning = false

														outTarget := ffmpeg.GetDerivedName(m.EditOpts.InputFiles[0], "_"+m.EditOpts.ActiveTool, "")
														if m.EditOpts.ActiveTool == "frame" {
															outTarget = ffmpeg.GetDerivedName(m.EditOpts.InputFiles[0], "_frame"+m.EditOpts.ExtractFrame, ".png")
														} else if m.EditOpts.ActiveTool == "convert" {
															outTarget = ffmpeg.GetDerivedName(m.EditOpts.InputFiles[0], "", m.EditOpts.TargetFormat)
														} else if m.EditOpts.ActiveTool == "split" {
															outTarget = ffmpeg.GetDerivedName(m.EditOpts.InputFiles[0], "_part1", "") + " & " + ffmpeg.GetDerivedName(m.EditOpts.InputFiles[0], "_part2", "")
														}

														m.History = append(m.History, HistoryItem{
															Action: strings.ToUpper(m.EditOpts.ActiveTool),
																   Target: outTarget,
														})
														cmds = append(cmds, delayedReset())
													}

												case MsgResetProgressBar:
													m.ProgressBar.SetPercent(0)
													m.IsRunning = false

												case progress.FrameMsg:
													newProgressModel, cmd := m.ProgressBar.Update(msg)
													m.ProgressBar = newProgressModel.(progress.Model)
													cmds = append(cmds, cmd)
	}

	// Route text input keystrokes to active Bubble Tea text components
	if m.ActivePanel == PanelConsole {
		var cmd tea.Cmd
		m.CmdInput, cmd = m.CmdInput.Update(msg)
		cmds = append(cmds, cmd)
	} else if m.ActivePanel == PanelSubOptions {
		var c1, c2, c3, c4, c5 tea.Cmd
		m.ParamInput1, c1 = m.ParamInput1.Update(msg)
		m.ParamInput2, c2 = m.ParamInput2.Update(msg)
		m.ParamInput3, c3 = m.ParamInput3.Update(msg)
		m.ParamInput4, c4 = m.ParamInput4.Update(msg)
		m.ParamInput5, c5 = m.ParamInput5.Update(msg)
		cmds = append(cmds, c1, c2, c3, c4, c5)

		// Real-time synchronization: sync fields and recalculate live command string on keypress
		m.syncInputsToEditOpts()
		m.UpdateLiveCommand()
	}

	return m, tea.Batch(cmds...)
}
