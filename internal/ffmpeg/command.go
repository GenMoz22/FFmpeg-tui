package ffmpeg

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type EditOptions struct {
	InputFiles     []string
	ActiveTool     string
	CropPreset     string
	TrimStart      string
	TrimEnd        string
	SplitPoint     string
	ExtractFrame   string
	NormalizeAudio bool

	SubPath      string
	SubPos       string
	SubOffset    string
	SubBgColor   string
	SubTextColor string

	TargetFormat string

	GifFPS        string
	GifScale      string
	CRFValue      string
	AudioFilePath string
	OutputFile    string
}

// ParseDurationString converts duration strings (HH:MM:SS, MM:SS, or raw seconds) into total seconds.
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
			if err != nil {
				return 0, err
			}
			minutes, err = strconv.ParseFloat(parts[1], 64)
			if err != nil {
				return 0, err
			}
			seconds, err = strconv.ParseFloat(parts[2], 64)
			if err != nil {
				return 0, err
			}
			return (hours * 3600) + (minutes * 60) + seconds, nil
		} else if len(parts) == 2 {
			minutes, err = strconv.ParseFloat(parts[0], 64)
			if err != nil {
				return 0, err
			}
			seconds, err = strconv.ParseFloat(parts[1], 64)
			if err != nil {
				return 0, err
			}
			return (minutes * 60) + seconds, nil
		}
	}

	return strconv.ParseFloat(s, 64)
}

// ResolveFilePath implements smart path resolution for secondary files (subtitles, audio tracks).
// Priority 1: Direct or absolute path as specified by user.
// Priority 2: Relative to main media file directory (and its subdirectories).
// Priority 3: Fallback search from root system directory.
// Returns resolved absolute path if found, or empty string if not found.
func ResolveFilePath(targetPath, mainVideoPath string) (string, error) {
	targetPath = strings.TrimSpace(targetPath)
	if targetPath == "" {
		return "", fmt.Errorf("path provided is empty")
	}

	// Priority 1: Check exact specified relative/absolute path
	if absPath, err := filepath.Abs(targetPath); err == nil {
		if _, err := os.Stat(absPath); err == nil {
			return absPath, nil
		}
	}

	// Priority 2: Check relative to main media file directory
	if mainVideoPath != "" {
		videoAbs, err := filepath.Abs(mainVideoPath)
		if err == nil {
			videoDir := filepath.Dir(videoAbs)
			candidate := filepath.Join(videoDir, targetPath)
			if absCandidate, err := filepath.Abs(candidate); err == nil {
				if _, err := os.Stat(absCandidate); err == nil {
					return absCandidate, nil
				}
			}
		}
	}

	// Priority 3: Fallback from root directory
	rootDir := string(filepath.Separator)
	candidateRoot := filepath.Join(rootDir, targetPath)
	if absRootCandidate, err := filepath.Abs(candidateRoot); err == nil {
		if _, err := os.Stat(absRootCandidate); err == nil {
			return absRootCandidate, nil
		}
	}

	return "", fmt.Errorf("file '%s' not found in relative path, video directory, or root", targetPath)
}

// GetDerivedName constructs output filenames with custom suffixes or extensions.
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

// BuildCommand returns the parameters slice for execution.
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
				args = append(args, "-map", videoLabel, "-map", "0:a?", "-c:v", "libx264", "-c:a", "copy")
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
					outName1 := GetDerivedName(inFile, "_part1", "")
					outName2 := GetDerivedName(inFile, "_part2", "")

					splitPoint := opts.SplitPoint
					if splitPoint == "" {
						splitPoint = "0"
					}

					cmd1 := fmt.Sprintf("-ss 0 -i %s -to %s -c copy -avoid_negative_ts make_zero -y %s", inFile, splitPoint, outName1)
					cmd2 := fmt.Sprintf("-ss %s -i %s -c copy -avoid_negative_ts make_zero -y %s", splitPoint, inFile, outName2)

					fullCmd := fmt.Sprintf("%s && ffmpeg %s", cmd1, cmd2)
					return strings.Split(fullCmd, " ")

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

					resolvedSubPath, err := ResolveFilePath(opts.SubPath, inFile)
					if err != nil {
						resolvedSubPath = opts.SubPath
					}
					escapedSubPath := strings.ReplaceAll(resolvedSubPath, "\\", "/")
					escapedSubPath = strings.ReplaceAll(escapedSubPath, ":", "\\:")

					filterStr := fmt.Sprintf("subtitles='%s'", escapedSubPath)
					if len(styleExpr) > 0 {
						filterStr += fmt.Sprintf(":force_style='%s'", strings.Join(styleExpr, ","))
					}
					args = append(args, "-vf", filterStr)
					outName := GetDerivedName(inFile, "_sub", "")
					args = append(args, "-y", outName)

				case "convert":
					args = append(args, "-i", inFile)
					if opts.TargetFormat == "mp3" {
						args = append(args, "-vn", "-acodec", "libmp3lame", "-q:a", "0")
					} else {
						args = append(args, "-c:v", "copy", "-c:a", "copy")
					}
					outName := GetDerivedName(inFile, "", opts.TargetFormat)
					args = append(args, "-y", outName)

				case "gif":
					if opts.TrimStart != "" {
						args = append(args, "-ss", opts.TrimStart)
					}
					if opts.TrimEnd != "" {
						args = append(args, "-to", opts.TrimEnd)
					}
					args = append(args, "-i", inFile)

					scale := opts.GifScale
					if scale == "" {
						scale = "480"
					}
					fps := opts.GifFPS
					if fps == "" {
						fps = "15"
					}

					vf := fmt.Sprintf("fps=%s,scale=%s:-1:flags=lanczos,split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse", fps, scale)
					args = append(args, "-vf", vf)
					outName := GetDerivedName(inFile, "", "gif")
					args = append(args, "-y", outName)

				case "compress":
					args = append(args, "-i", inFile, "-vcodec", "libx264", "-crf", opts.CRFValue, "-preset", "medium")
					outName := GetDerivedName(inFile, "_crf"+opts.CRFValue, "")
					args = append(args, "-y", outName)

				case "sepaudio":
					outVideo := GetDerivedName(inFile, "_noaudio", "")
					outAudio := GetDerivedName(inFile, "", "mp3")
					args = append(args, "-i", inFile, "-an", "-vcodec", "copy", "-y", outVideo, "-i", inFile, "-vn", "-acodec", "libmp3lame", "-q:a", "0", "-y", outAudio)

				case "metadata":
					args = append(args, "-i", inFile, "-map_metadata", "-1", "-c", "copy")
					outName := GetDerivedName(inFile, "_clean", "")
					args = append(args, "-y", outName)

				case "replaceaudio":
					resolvedAudioPath, err := ResolveFilePath(opts.AudioFilePath, inFile)
					if err != nil {
						resolvedAudioPath = opts.AudioFilePath
					}
					args = append(args, "-i", inFile, "-i", resolvedAudioPath, "-c:v", "copy", "-c:a", "aac", "-map", "0:v:0", "-map", "1:a:0", "-shortest")
					outName := GetDerivedName(inFile, "_newaudio", "")
					args = append(args, "-y", outName)

				case "autocrop":
					args = append(args, "-i", inFile, "-vf", "cropdetect=24:2:0,crop=iw:ih", "-map", "0:v:0", "-map", "0:a?", "-c:a", "copy")
					outName := GetDerivedName(inFile, "_autocrop", "")
					args = append(args, "-y", outName)
	}

	return args
}
