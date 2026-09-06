package ffmpeg

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"time"
)

type MediaInfo struct {
	Path     string        `json:"path"`
	Duration time.Duration `json:"duration"`
	Bitrate  int64         `json:"bitrate"`
	Size     int64         `json:"size"`
	Video    *VideoStream  `json:"video,omitempty"`
	Audio    *AudioStream  `json:"audio,omitempty"`
}

func (m *MediaInfo) FormatDuration() string {
	return fmt.Sprintf("%.2fs", m.Duration.Seconds())
}

type VideoStream struct {
	Codec       string `json:"codec"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	FPS         string `json:"fps"`
	AspectRatio string `json:"aspect_ratio"`
}

type AudioStream struct {
	Codec      string `json:"codec"`
	Channels   int    `json:"channels"`
	SampleRate string `json:"sample_rate"`
}

type ffprobeOutput struct {
	Format struct {
		Duration string `json:"duration"`
		BitRate  string `json:"bit_rate"`
		Size     string `json:"size"`
	} `json:"format"`
	Streams []struct {
		CodecType          string `json:"codec_type"`
		CodecName          string `json:"codec_name"`
		Width              int    `json:"width"`
		Height             int    `json:"height"`
		RFrameRate         string `json:"r_frame_rate"`
		DisplayAspectRatio string `json:"display_aspect_ratio"`
		Channels           int    `json:"channels"`
		SampleRate         string `json:"sample_rate"`
	} `json:"streams"`
}

func ProbeMedia(ctx context.Context, path string) (*MediaInfo, error) {
	args := []string{
		"-v", "error",
		"-show_entries", "format=duration,bit_rate,size:stream=codec_type,codec_name,width,height,r_frame_rate,display_aspect_ratio,channels,sample_rate",
		"-of", "json",
		path,
	}

	cmd := exec.CommandContext(ctx, "ffprobe", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var raw ffprobeOutput
	if err := json.Unmarshal(output, &raw); err != nil {
		return nil, err
	}

	info := &MediaInfo{Path: path}

	if d, err := strconv.ParseFloat(raw.Format.Duration, 64); err == nil {
		info.Duration = time.Duration(d * float64(time.Second))
	}
	if b, err := strconv.ParseInt(raw.Format.BitRate, 10, 64); err == nil {
		info.Bitrate = b
	}
	if s, err := strconv.ParseInt(raw.Format.Size, 10, 64); err == nil {
		info.Size = s
	}

	for _, stream := range raw.Streams {
		switch stream.CodecType {
			case "video":
				info.Video = &VideoStream{
					Codec:       stream.CodecName,
					Width:       stream.Width,
					Height:      stream.Height,
					FPS:         stream.RFrameRate,
					AspectRatio: stream.DisplayAspectRatio,
				}
			case "audio":
				info.Audio = &AudioStream{
					Codec:      stream.CodecName,
					Channels:   stream.Channels,
					SampleRate: stream.SampleRate,
				}
		}
	}

	return info, nil
}
