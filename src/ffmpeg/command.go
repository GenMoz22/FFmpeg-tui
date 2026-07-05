package ffmpeg

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

type EditOptions struct {
	InputFiles     []string
	ActiveTool     string // "crop", "trim", "split", "audio", "frame", "subtitles", "convert"
	CropPreset     string
	TrimStart      string
	TrimEnd        string
	SplitPoint     string
	ExtractFrame   string
	NormalizeAudio bool

	// Subtitles configuration parameters
	SubPath        string
	SubPos         string
	SubOffset      string
	SubBgColor     string
	SubTextColor  string

	// Target format conversion override extension
	TargetFormat   string // "mp4", "mkv", "mov", "avi", "mp3"

	OutputFile     string
}

func ParseDurationString(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	if strings.Contains(s, ":") {
		parts := strings.Split(s, ":")
		var hours, minutes, seconds float64
		var err error

		if len(parts) == 3 {
			hours, err = strconv.ParseFloat(parts[0], 64)
			if err != nil { return 0, err }
			minutes, err = strconv.ParseFloat(parts[1], 64)
			if err != nil { return 0, err }
			seconds, err = strconv.ParseFloat(parts[2], 64)
			if err != nil { return 0, err }
			return (hours * 3600) + (minutes * 60) + seconds, nil
		} else if len(parts) == 2 {
			minutes, err = strconv.ParseFloat(parts[0], 64)
			if err != nil { return 0, err }
			seconds, err = strconv.ParseFloat(parts[1], 64)
			if err != nil { return 0, err }
			return (minutes * 60) + seconds, nil
		}
	}
	return strconv.ParseFloat(s, 64)
}

func GetDerivedName(inputPath, suffix, extOverride string) string {
	base := filepath.Base(inputPath)
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)

	targetExt := ext
	if extOverride != "" {
		if !strings.HasPrefix(extOverride, ".") {
			targetExt = "." + extOverride
		} else {
			targetExt = extOverride
		}
	}

	return fmt.Sprintf("%s%s%s", nameWithoutExt, suffix, targetExt)
}

func BuildCommand(opts EditOptions) []string {
	var args []string
	if len(opts.InputFiles) == 0 {
		return []string{"-i", "placeholder.mp4", "output.mp4"}
	}

	inFile := opts.InputFiles[0]

	switch opts.ActiveTool {
		case "crop":
			args = append(args, "-i", inFile)
			var filterComplex []string
			videoLabel := "[0:v]"

			switch opts.CropPreset {
				case "9:16":
					filterComplex = append(filterComplex, fmt.Sprintf("%scrop=ih*(9/16):ih[cropped]", videoLabel))
					videoLabel = "[cropped]"
				case "1:1":
					filterComplex = append(filterComplex, fmt.Sprintf("%scrop=ih:ih[cropped]", videoLabel))
					videoLabel = "[cropped]"
				case "16:9":
					filterComplex = append(filterComplex, fmt.Sprintf("%scrop=iw:iw*(9/16)[cropped]", videoLabel))
					videoLabel = "[cropped]"
			}

			if len(filterComplex) > 0 {
				args = append(args, "-filter_complex", strings.Join(filterComplex, ";"))
				args = append(args, "-map", videoLabel)
			}
			outName := GetDerivedName(inFile, "_crop", "")
			args = append(args, "-y", outName)

				case "trim":
					if opts.TrimStart != "" {
						args = append(args, "-ss", opts.TrimStart)
					}
					if opts.TrimEnd != "" {
						args = append(args, "-to", opts.TrimEnd)
					}
					args = append(args, "-i", inFile)
					outName := GetDerivedName(inFile, "_trim", "")
					args = append(args, "-y", outName)

				case "split":
					args = append(args, "-to", opts.SplitPoint, "-i", inFile)
					args = append(args, "-ss", opts.SplitPoint, "-i", inFile)

					outName1 := GetDerivedName(inFile, "_split1", "")
					outName2 := GetDerivedName(inFile, "_split2", "")
					args = append(args, "-map", "0", "-y", outName1, "-map", "1", "-y", outName2)

				case "audio":
					args = append(args, "-i", inFile)
					if opts.NormalizeAudio {
						args = append(args, "-af", "loudnorm=I=-16:TP=-1.5:LRA=11")
					}
					outName := GetDerivedName(inFile, "_normalized", "")
					args = append(args, "-y", outName)

				case "frame":
					if opts.ExtractFrame != "" {
						args = append(args, "-ss", opts.ExtractFrame)
					}
					args = append(args, "-i", inFile, "-vframes", "1")

					sanitizedSec := strings.ReplaceAll(opts.ExtractFrame, ":", "-")
					if sanitizedSec == "" {
						sanitizedSec = "0"
					}
					outName := GetDerivedName(inFile, "_frame"+sanitizedSec, ".png")
					args = append(args, "-y", outName)

				case "subtitles":
					args = append(args, "-i", inFile)

					var styleExpr []string
					if opts.SubBgColor != "" && opts.SubBgColor != "none" {
						styleExpr = append(styleExpr, fmt.Sprintf("OutlineColour=&H80%s", opts.SubBgColor))
					}

					alignment := "2"
					if opts.SubPos == "top" {
						alignment = "6"
					} else if opts.SubPos == "center" {
						alignment = "10"
					}
					styleExpr = append(styleExpr, fmt.Sprintf("Alignment=%s", alignment))

					filterStr := fmt.Sprintf("subtitles='%s'", opts.SubPath)
					if len(styleExpr) > 0 {
						filterStr += fmt.Sprintf(":force_style='%s'", strings.Join(styleExpr, ","))
					}

					args = append(args, "-vf", filterStr)
					outName := GetDerivedName(inFile, "_sub", "")
					args = append(args, "-y", outName)

				case "convert":
					// Direct streaming copy layout or native multi-format encapsulation processing
					args = append(args, "-i", inFile)
					if opts.TargetFormat == "mp3" {
						// Extract audio streams exclusively when transcoding down to pure MP3 format
						args = append(args, "-vn", "-acodec", "libmp3lame")
					} else {
						args = append(args, "-c:v", "copy", "-c:a", "copy")
					}
					outName := GetDerivedName(inFile, "", opts.TargetFormat)
					args = append(args, "-y", outName)
	}

	return args
}
