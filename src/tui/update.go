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

func listenToProgress(ch <-chan ffmpeg.ProgressMessage) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return MsgFFmpegProgress(msg)
	}
}

func delayedReset() tea.Cmd {
	return tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
		return MsgResetProgressBar{}
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	m.ValidationError = ""

	switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
				case "ctrl+c", "q":
					if m.CtxCancel != nil {
						m.CtxCancel()
					}
					return m, tea.Quit

				case "tab":
					if m.ActivePanel == PanelSidebar {
						m.ActivePanel = PanelSubOptions
						m.SubMenuFocusIdx = 0
						// Lists (Crop/Convert) don't need ParamInput keyboard focusing
						if len(m.SubOptionsItems) > 0 && m.SidebarIdx != 0 && m.SidebarIdx != 6 && m.SidebarIdx != 3 {
							cmds = append(cmds, m.ParamInput1.Focus())
						}
					} else if m.ActivePanel == PanelSubOptions {
						m.ParamInput1.Blur()
						m.ParamInput2.Blur()
						m.ParamInput3.Blur()
						m.ParamInput4.Blur()
						m.ParamInput5.Blur()

						if m.SidebarIdx == 1 || m.SidebarIdx == 2 {
							if m.SubMenuFocusIdx == 0 {
								m.SubMenuFocusIdx = 1
								cmds = append(cmds, m.ParamInput2.Focus())
							} else {
								m.ActivePanel = PanelConsole
								cmds = append(cmds, m.CmdInput.Focus())
							}
						} else if m.SidebarIdx == 5 {
							if m.SubMenuFocusIdx < 4 {
								m.SubMenuFocusIdx++
								switch m.SubMenuFocusIdx {
									case 1: cmds = append(cmds, m.ParamInput2.Focus())
									case 2: cmds = append(cmds, m.ParamInput3.Focus())
									case 3: cmds = append(cmds, m.ParamInput4.Focus())
									case 4: cmds = append(cmds, m.ParamInput5.Focus())
								}
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
										} else if m.ActivePanel == PanelSubOptions && (m.SidebarIdx == 0 || m.SidebarIdx == 6) && m.SubOptionsIdx > 0 {
											m.SubOptionsIdx--
										}

									case "down", "j":
										if m.ActivePanel == PanelSidebar && m.SidebarIdx < len(m.SidebarItems)-1 {
											m.SidebarIdx++
										} else if m.ActivePanel == PanelSubOptions && (m.SidebarIdx == 0 || m.SidebarIdx == 6) && m.SubOptionsIdx < len(m.SubOptionsItems)-1 {
											m.SubOptionsIdx++
										}

									case "enter":
										if m.ActivePanel == PanelSidebar {
											m.SubOptionsIdx = 0
											m.SubMenuFocusIdx = 0
											m.ParamInput1.SetValue("")
											m.ParamInput2.SetValue("")
											m.ParamInput3.SetValue("")
											m.ParamInput4.SetValue("")
											m.ParamInput5.SetValue("")

											switch m.SidebarIdx {
												case 0: // Crop
													m.EditOpts.ActiveTool = "crop"
													m.SubOptionsItems = []string{"9:16 (Shorts/Reels)", "1:1 (Square)", "16:9 (Widescreen)"}
													m.ActivePanel = PanelSubOptions
												case 1: // Trim
													m.EditOpts.ActiveTool = "trim"
													m.SubOptionsItems = []string{"Parameters Setup"}
													m.ParamInput1.Placeholder = "Start (e.g., 0 or 00:00:00)"
													m.ParamInput2.Placeholder = "End (e.g., 30 or 00:00:30)"
													if m.MediaInfo != nil {
														m.ParamInput1.SetValue("0")
														m.ParamInput2.SetValue(strconv.FormatFloat(m.MediaInfo.Duration.Seconds(), 'f', 2, 64))
													}
													m.ActivePanel = PanelSubOptions
													cmds = append(cmds, m.ParamInput1.Focus())
												case 2: // Split
													m.EditOpts.ActiveTool = "split"
													m.SubOptionsItems = []string{"Parameters Setup"}
													m.ParamInput1.Placeholder = "Split point (e.g., 10 or 00:00:10)"
													m.ActivePanel = PanelSubOptions
													cmds = append(cmds, m.ParamInput1.Focus())
												case 3: // Audio
													m.EditOpts.ActiveTool = "audio"
													m.EditOpts.NormalizeAudio = true
													m.SubOptionsItems = []string{"Loudnorm Equalizer Active"}
												case 4: // Frame
													m.EditOpts.ActiveTool = "frame"
													m.SubOptionsItems = []string{"Parameters Setup"}
													m.ParamInput1.Placeholder = "Timestamp (e.g., 5 or 00:00:05)"
													m.ParamInput1.SetValue("0")
													m.ActivePanel = PanelSubOptions
													cmds = append(cmds, m.ParamInput1.Focus())
												case 5: // Subtitles
													m.EditOpts.ActiveTool = "subtitles"
													m.SubOptionsItems = []string{"Hardburn Style Properties Layout"}
													m.ParamInput1.Placeholder = "File Path (e.g., sub.srt)"
													m.ParamInput2.Placeholder = "Position (bottom/top/center)"
													m.ParamInput2.SetValue("bottom")
													m.ParamInput3.Placeholder = "Vertical Offset (e.g., 20)"
													m.ParamInput3.SetValue("10")
													m.ParamInput4.Placeholder = "Background Color (black/none)"
													m.ParamInput4.SetValue("black")
													m.ParamInput5.Placeholder = "Text Color (white/yellow)"
													m.ParamInput5.SetValue("white")
													m.ActivePanel = PanelSubOptions
													cmds = append(cmds, m.ParamInput1.Focus())
												case 6: // Convert Formats
													m.EditOpts.ActiveTool = "convert"
													m.SubOptionsItems = []string{"Container Format: MP4", "Container Format: MKV", "Container Format: MOV", "Container Format: AVI", "Pure Audio Extraction: MP3"}
													m.ActivePanel = PanelSubOptions
											}
											m.UpdateLiveCommand()
										} else if m.ActivePanel == PanelSubOptions {
											maxDuration := 0.0
											if m.MediaInfo != nil {
												maxDuration = m.MediaInfo.Duration.Seconds()
											}

											switch m.EditOpts.ActiveTool {
												case "crop":
													presets := []string{"9:16", "1:1", "16:9"}
													m.EditOpts.CropPreset = presets[m.SubOptionsIdx]
												case "trim":
													t1, _ := ffmpeg.ParseDurationString(m.ParamInput1.Value())
													t2, _ := ffmpeg.ParseDurationString(m.ParamInput2.Value())
													if t1 > maxDuration || t2 > maxDuration || t1 >= t2 {
														m.ValidationError = fmt.Sprintf("OUT OF BOUNDS: Video length is %.2fs max.", maxDuration)
														return m, tea.Batch(cmds...)
													}
													m.EditOpts.TrimStart = m.ParamInput1.Value()
													m.EditOpts.TrimEnd = m.ParamInput2.Value()
												case "split":
													sp, _ := ffmpeg.ParseDurationString(m.ParamInput1.Value())
													if sp <= 0 || sp >= maxDuration {
														m.ValidationError = fmt.Sprintf("INVALID TIMESTAMP: Split point outside range (Max %.2fs).", maxDuration)
														return m, tea.Batch(cmds...)
													}
													m.EditOpts.SplitPoint = m.ParamInput1.Value()
												case "frame":
													fp, _ := ffmpeg.ParseDurationString(m.ParamInput1.Value())
													if fp > maxDuration || fp < 0 {
														m.ValidationError = fmt.Sprintf("OUT OF BOUNDS: Target must be between 0 and %.2fs.", maxDuration)
														return m, tea.Batch(cmds...)
													}
													m.EditOpts.ExtractFrame = m.ParamInput1.Value()
												case "subtitles":
													m.EditOpts.SubPath = m.ParamInput1.Value()
													m.EditOpts.SubPos = m.ParamInput2.Value()
													m.EditOpts.SubOffset = m.ParamInput3.Value()
													m.EditOpts.SubBgColor = m.ParamInput4.Value()
													m.EditOpts.SubTextColor = m.ParamInput5.Value()
												case "convert":
													formats := []string{"mp4", "mkv", "mov", "avi", "mp3"}
													m.EditOpts.TargetFormat = formats[m.SubOptionsIdx]
											}
											m.UpdateLiveCommand()
											m.ActivePanel = PanelConsole
											cmds = append(cmds, m.CmdInput.Focus())
											m.ParamInput1.Blur()
											m.ParamInput2.Blur()
											m.ParamInput3.Blur()
											m.ParamInput4.Blur()
											m.ParamInput5.Blur()
										} else if m.ActivePanel == PanelConsole {
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

												case MsgFFmpegProgress:
													progCmd := m.ProgressBar.SetPercent(msg.Percent)
													cmds = append(cmds, progCmd)
													if !msg.Done && msg.Err == nil {
														cmds = append(cmds, listenToProgress(m.ProgressChan))
													} else if msg.Done {
														outTarget := ffmpeg.GetDerivedName(m.EditOpts.InputFiles[0], "_"+m.EditOpts.ActiveTool, "")
														if m.EditOpts.ActiveTool == "frame" {
															outTarget = ffmpeg.GetDerivedName(m.EditOpts.InputFiles[0], "_frame"+m.EditOpts.ExtractFrame, ".png")
														} else if m.EditOpts.ActiveTool == "convert" {
															outTarget = ffmpeg.GetDerivedName(m.EditOpts.InputFiles[0], "", m.EditOpts.TargetFormat)
														}
														m.History = append(m.History, HistoryItem{
															Action: strings.ToUpper(m.EditOpts.ActiveTool),
																   Target: outTarget,
														})
														cmds = append(cmds, delayedReset())
													}

												case MsgResetProgressBar:
													m.ProgressBar.SetPercent(0)

												case progress.FrameMsg:
													newProgressModel, cmd := m.ProgressBar.Update(msg)
													m.ProgressBar = newProgressModel.(progress.Model)
													cmds = append(cmds, cmd)
	}

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
	}

	return m, tea.Batch(cmds...)
}
