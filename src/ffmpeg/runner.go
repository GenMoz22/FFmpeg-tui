package ffmpeg

import (
	"bufio"
	"context"
	"os/exec"
	"strconv"
	"strings"
)

type ProgressMessage struct {
	Percent float64
	Done    bool
	Err     error
}

func ExecuteFFmpeg(ctx context.Context, args []string, totalDurationSec float64) <-chan ProgressMessage {
	outChan := make(chan ProgressMessage)

	go func() {
		defer close(outChan)

		fullArgs := append([]string{"-progress", "pipe:1"}, args...)
		cmd := exec.CommandContext(ctx, "ffmpeg", fullArgs...)

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			outChan <- ProgressMessage{Err: err}
			return
		}

		if err := cmd.Start(); err != nil {
			outChan <- ProgressMessage{Err: err}
			return
		}

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "out_time_us=") {
				timeUsStr := strings.TrimPrefix(line, "out_time_us=")
				if timeUs, err := strconv.ParseFloat(timeUsStr, 64); err == nil && totalDurationSec > 0 {
					currentSec := timeUs / 1000000.0
					percent := currentSec / totalDurationSec
					if percent > 1.0 {
						percent = 1.0
					}
					outChan <- ProgressMessage{Percent: percent}
				}
			}
		}

		if err := cmd.Wait(); err != nil {
			outChan <- ProgressMessage{Err: err}
			return
		}

		outChan <- ProgressMessage{Percent: 1.0, Done: true}
	}()

	return outChan
}
