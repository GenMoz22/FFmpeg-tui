package tui

import "ffmpeg-tui/src/ffmpeg"

// MsgFFmpegProgress carries progress percentage and execution status from FFmpeg.
type MsgFFmpegProgress ffmpeg.ProgressMessage

// MsgResetProgressBar signals the UI to reset the progress bar indicator to 0%.
type MsgResetProgressBar struct{}

// MsgClearValidationError signals the UI to safely clear validation error banners after display timeout.
type MsgClearValidationError struct{}

// MsgMediaProbed carries the result of probing video/audio media metadata.
type MsgMediaProbed struct {
	Info *ffmpeg.MediaInfo
}

// MsgError represents an internal error message passed to the UI update loop.
type MsgError struct {
	Err error
}
