package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ffmpeg-tui/internal/ffmpeg"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

// listenToProgress listens for execution progress updates from the background FFmpeg process.
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

// clearErrorAfterTimeout retains error banners visible for 6 seconds to ensure usability.
func clearErrorAfterTimeout() tea.Cmd {
	return tea.Tick(time.Second*6, func(t time.Time) tea.Msg {
		return MsgClearValidationError{}
	})
}

// parseCmdArgs splits a command line string into a slice of arguments, respecting single and double quotes.
func parseCmdArgs(cmdStr string) []string {
	var args []string
	var current strings.Builder
	inDoubleQuotes := false
	inSingleQuotes := false

	for i := 0; i < len(cmdStr); i++ {
		r := cmdStr[i]
		switch r {
			case '"':
				if !inSingleQuotes {
					inDoubleQuotes = !inDoubleQuotes
				} else {
					current.WriteByte(r)
				}
			case '\'':
				if !inDoubleQuotes {
					inSingleQuotes = !inSingleQuotes
				} else {
					current.WriteByte(r)
				}
			case ' ', '\t':
				if inDoubleQuotes || inSingleQuotes {
					current.WriteByte(r)
				} else if current.Len() > 0 {
					args = append(args, current.String())
					current.Reset()
				}
			default:
				current.WriteByte(r)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}

// syncInputsToEditOpts binds current active input text fields into the underlying EditOptions model.
func (m *Model) syncInputsToEditOpts() {
	mainVid := ""
	if len(m.EditOpts.InputFiles) > 0 {
		mainVid = m.EditOpts.InputFiles[0]
	}

	switch m.EditOpts.ActiveTool {
		case "trim":
			m.EditOpts.TrimStart = strings.TrimSpace(m.ParamInput1.Value())
			m.EditOpts.TrimEnd = strings.TrimSpace(m.ParamInput2.Value())
		case "split":
			m.EditOpts.SplitPoint = strings.TrimSpace(m.ParamInput1.Value())
		case "frame":
			m.EditOpts.ExtractFrame = strings.TrimSpace(m.ParamInput1.Value())
		case "subtitles":
			rawPath := strings.TrimSpace(m.ParamInput1.Value())
			if resolved, err := ffmpeg.ResolveFilePath(rawPath, mainVid); err == nil {
				m.EditOpts.SubPath = resolved
			} else {
				m.EditOpts.SubPath = rawPath
			}
			m.EditOpts.SubPos = strings.TrimSpace(m.ParamInput2.Value())
			m.EditOpts.SubOffset = strings.TrimSpace(m.ParamInput3.Value())
			m.EditOpts.SubBgColor = strings.TrimSpace(m.ParamInput4.Value())
			m.EditOpts.SubTextColor = strings.TrimSpace(m.ParamInput5.Value())
		case "gif":
			m.EditOpts.TrimStart = strings.TrimSpace(m.ParamInput1.Value())
			m.EditOpts.TrimEnd = strings.TrimSpace(m.ParamInput2.Value())
		case "replaceaudio":
			rawPath := strings.TrimSpace(m.ParamInput1.Value())
			if resolved, err := ffmpeg.ResolveFilePath(rawPath, mainVid); err == nil {
				m.EditOpts.AudioFilePath = resolved
			} else {
				m.EditOpts.AudioFilePath = rawPath
			}
	}
}

// Update handles application state mutations, keyboard navigation, and asynchronous commands.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
		case MsgClearValidationError:
			m.ValidationError = ""
			return m, nil

		case tea.KeyMsg:
			switch msg.String() {
				case "ctrl+c", "esc":
					if m.CtxCancel != nil {
						m.CtxCancel()
					}
					return m, tea.Quit

				case "tab":
					m.ValidationError = ""
					if m.ActivePanel == PanelSidebar {
						m.ActivePanel = PanelSubOptions
						m.SubMenuFocusIdx = 0
						currentKind := m.SidebarItems[m.SidebarIdx].Kind
						if len(m.SubOptionsItems) > 0 && currentKind != ModeCrop && currentKind != ModeConvert && currentKind != ModeAudioNorm && currentKind != ModeCompress && currentKind != ModeSepAudio && currentKind != ModeMetadata {
							cmds = append(cmds, m.ParamInput1.Focus())
						}
					} else if m.ActivePanel == PanelSubOptions {
						m.ParamInput1.Blur()
						m.ParamInput2.Blur()
						m.ParamInput3.Blur()
						m.ParamInput4.Blur()
						m.ParamInput5.Blur()

						currentKind := m.SidebarItems[m.SidebarIdx].Kind
						if currentKind == ModeTrim {
							if m.SubMenuFocusIdx == 0 {
								m.SubMenuFocusIdx = 1
								cmds = append(cmds, m.ParamInput2.Focus())
							} else {
								m.ActivePanel = PanelConsole
								cmds = append(cmds, m.CmdInput.Focus())
							}
						} else if currentKind == ModeSubtitles {
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
						} else if currentKind == ModeConvert && m.EditOpts.ActiveTool == "gif" {
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
										m.ValidationError = ""
										if m.ActivePanel == PanelSidebar && m.SidebarIdx > 0 {
											m.SidebarIdx--
											m.SelectSidebarItem(m.SidebarIdx)
											m.syncInputsToEditOpts()
											m.UpdateLiveCommand()
										} else if m.ActivePanel == PanelSubOptions {
											currentKind := m.SidebarItems[m.SidebarIdx].Kind
											if (currentKind == ModeCrop || currentKind == ModeConvert || currentKind == ModeCompress) && m.SubOptionsIdx > 0 {
												m.SubOptionsIdx--
											}
										}

									case "down", "j":
										m.ValidationError = ""
										if m.ActivePanel == PanelSidebar && m.SidebarIdx < len(m.SidebarItems)-1 {
											m.SidebarIdx++
											m.SelectSidebarItem(m.SidebarIdx)
											m.syncInputsToEditOpts()
											m.UpdateLiveCommand()
										} else if m.ActivePanel == PanelSubOptions {
											currentKind := m.SidebarItems[m.SidebarIdx].Kind
											if (currentKind == ModeCrop || currentKind == ModeConvert || currentKind == ModeCompress) && m.SubOptionsIdx < len(m.SubOptionsItems)-1 {
												m.SubOptionsIdx++
											}
										}

									case "enter":
										m.ValidationError = ""
										if m.ActivePanel == PanelSidebar {
											m.SelectSidebarItem(m.SidebarIdx)

											currentKind := m.SidebarItems[m.SidebarIdx].Kind
											if currentKind != ModeAudioNorm && currentKind != ModeSepAudio && currentKind != ModeMetadata {
												m.ActivePanel = PanelSubOptions
											}

											if currentKind == ModeTrim || currentKind == ModeSplit || currentKind == ModeFrame || currentKind == ModeSubtitles || currentKind == ModeReplaceAudio {
												cmds = append(cmds, m.ParamInput1.Focus())
											}

											m.syncInputsToEditOpts()
											m.UpdateLiveCommand()

										} else if m.ActivePanel == PanelSubOptions {
											maxDuration := 0.0
											if m.MediaInfo != nil {
												maxDuration = m.MediaInfo.Duration.Seconds()
											}

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

											mainVid := ""
											if len(m.EditOpts.InputFiles) > 0 {
												mainVid = m.EditOpts.InputFiles[0]
											}

											currentKind := m.SidebarItems[m.SidebarIdx].Kind
											switch currentKind {
												case ModeCrop:
													if m.SubOptionsIdx == 3 {
														m.EditOpts.ActiveTool = "autocrop"
													} else {
														m.EditOpts.ActiveTool = "crop"
														presets := []string{"9:16", "1:1", "16:9"}
														m.EditOpts.CropPreset = presets[m.SubOptionsIdx]
													}
												case ModeTrim:
													t1, err1 := ffmpeg.ParseDurationString(m.ParamInput1.Value())
													t2, err2 := ffmpeg.ParseDurationString(m.ParamInput2.Value())
													if err1 != nil || err2 != nil {
														m.ValidationError = "INVALID TIMESTAMP: Use HH:MM:SS or seconds."
														cmds = append(cmds, clearErrorAfterTimeout())
														return m, tea.Batch(cmds...)
													}
													if maxDuration > 0 && (t1 > maxDuration || t2 > maxDuration || t1 >= t2) {
														m.ValidationError = fmt.Sprintf("OUT OF BOUNDS: Video length is %.2fs max.", maxDuration)
														cmds = append(cmds, clearErrorAfterTimeout())
														return m, tea.Batch(cmds...)
													}
													m.EditOpts.TrimStart = m.ParamInput1.Value()
													m.EditOpts.TrimEnd = m.ParamInput2.Value()
												case ModeSplit:
													sp, err := ffmpeg.ParseDurationString(m.ParamInput1.Value())
													if err != nil {
														m.ValidationError = "INVALID TIMESTAMP: Use HH:MM:SS or seconds."
														cmds = append(cmds, clearErrorAfterTimeout())
														return m, tea.Batch(cmds...)
													}
													if maxDuration > 0 && (sp >= maxDuration || sp <= 0) {
														m.ValidationError = fmt.Sprintf("OUT OF BOUNDS: Split point must be between 0s and %.2fs.", maxDuration)
														cmds = append(cmds, clearErrorAfterTimeout())
														return m, tea.Batch(cmds...)
													}
													m.EditOpts.SplitPoint = m.ParamInput1.Value()
												case ModeSubtitles:
													subPath := strings.TrimSpace(m.ParamInput1.Value())
													if subPath == "" {
														m.ValidationError = "ERROR: Missing subtitle file path."
														cmds = append(cmds, clearErrorAfterTimeout())
														return m, tea.Batch(cmds...)
													}
													resolved, err := ffmpeg.ResolveFilePath(subPath, mainVid)
													if err != nil {
														m.ValidationError = fmt.Sprintf("FILE NOT FOUND: Subtitle '%s' not found.", subPath)
														cmds = append(cmds, clearErrorAfterTimeout())
														return m, tea.Batch(cmds...)
													}
													m.ParamInput1.SetValue(resolved)
													m.EditOpts.SubPath = resolved
												case ModeConvert:
													if m.MediaInfo != nil && m.MediaInfo.Video == nil {
														m.EditOpts.ActiveTool = "convert"
														formats := []string{"mp3", "wav", "aac", "flac", "ogg"}
														m.EditOpts.TargetFormat = formats[m.SubOptionsIdx]
													} else {
														if m.SubOptionsIdx == 5 {
															m.EditOpts.ActiveTool = "gif"
															m.SubOptionsItems = []string{"Convert to Animated GIF"}
															m.ParamInput1.Placeholder = "Start Time (e.g., 00:00:00)"
															m.ParamInput2.Placeholder = "End Time (e.g., 00:00:05)"
															cmds = append(cmds, m.ParamInput1.Focus())
															m.syncInputsToEditOpts()
															m.UpdateLiveCommand()
															return m, tea.Batch(cmds...)
														} else {
															m.EditOpts.ActiveTool = "convert"
															formats := []string{"mp4", "mkv", "mov", "avi", "mp3"}
															m.EditOpts.TargetFormat = formats[m.SubOptionsIdx]
														}
													}
												case ModeCompress:
													if m.SubOptionsIdx == 0 {
														m.EditOpts.CRFValue = "23"
													} else {
														m.EditOpts.CRFValue = "28"
													}
												case ModeReplaceAudio:
													audioPath := strings.TrimSpace(m.ParamInput1.Value())
													if audioPath == "" {
														m.ValidationError = "ERROR: Missing target audio file path."
														cmds = append(cmds, clearErrorAfterTimeout())
														return m, tea.Batch(cmds...)
													}
													resolved, err := ffmpeg.ResolveFilePath(audioPath, mainVid)
													if err != nil {
														m.ValidationError = fmt.Sprintf("FILE NOT FOUND: Audio '%s' not found.", audioPath)
														cmds = append(cmds, clearErrorAfterTimeout())
														return m, tea.Batch(cmds...)
													}
													m.ParamInput1.SetValue(resolved)
													m.EditOpts.AudioFilePath = resolved
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

											rawCmd := strings.TrimPrefix(m.CmdInput.Value(), "ffmpeg ")
											customArgs := parseCmdArgs(rawCmd)
											m.ProgressChan = ffmpeg.ExecuteFFmpeg(ctx, customArgs, durationSec)
											cmds = append(cmds, listenToProgress(m.ProgressChan))
										}
			}

												case MsgMediaProbed:
													m.MediaInfo = msg.Info
													m.FilterSidebarForMedia()
													m.UpdateLiveCommand()

												case MsgError:
													m.ValidationError = fmt.Sprintf("Error probing file: %v", msg.Err)
													cmds = append(cmds, clearErrorAfterTimeout())

												case MsgFFmpegProgress:
													if msg.Err != nil {
														m.IsRunning = false
														m.ValidationError = fmt.Sprintf("FFmpeg execution failed: %v", msg.Err)
														m.History = append(m.History, HistoryItem{
															Action:  strings.ToUpper(m.EditOpts.ActiveTool),
																   Target:  msg.Err.Error(),
																   Success: false,
														})
														cmds = append(cmds, clearErrorAfterTimeout())
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
															inFile := m.EditOpts.InputFiles[0]
															p1Name := ffmpeg.GetDerivedName(inFile, "_part1", "")
															p2Name := ffmpeg.GetDerivedName(inFile, "_part2", "")
															outTarget = p1Name + " & " + p2Name
														}

														m.History = append(m.History, HistoryItem{
															Action:  strings.ToUpper(m.EditOpts.ActiveTool),
																   Target:  outTarget,
																   Success: true,
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

		m.syncInputsToEditOpts()
		m.UpdateLiveCommand()
	}

	return m, tea.Batch(cmds...)
}
