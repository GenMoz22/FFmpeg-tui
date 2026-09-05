package ffmpeg

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type ProgressMessage struct {
	Percent float64
	Done    bool
	Err     error
}

// ExecuteFFmpeg executes ffmpeg and streams progress percentage over a channel.
func ExecuteFFmpeg(ctx context.Context, args []string, totalDurationSec float64) <-chan ProgressMessage {
	outChan := make(chan ProgressMessage, 100)

	go func() {
		defer close(outChan)

		var cmd *exec.Cmd
		fullArgsStr := strings.Join(args, " ")

		if strings.Contains(fullArgsStr, "&&") {
			if runtime.GOOS == "windows" {
				cmd = exec.CommandContext(ctx, "cmd", "/C", "ffmpeg "+fullArgsStr)
			} else {
				cmd = exec.CommandContext(ctx, "sh", "-c", "ffmpeg "+fullArgsStr)
			}
		} else {
			cmd = exec.CommandContext(ctx, "ffmpeg", args...)
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			outChan <- ProgressMessage{Err: fmt.Errorf("failed to get stderr pipe: %w", err)}
			return
		}

		if err := cmd.Start(); err != nil {
			outChan <- ProgressMessage{Err: fmt.Errorf("failed to start ffmpeg: %w", err)}
			return
		}

		scanner := bufio.NewScanner(stderr)
		scanner.Split(bufio.ScanLines)

		for scanner.Scan() {
			line := scanner.Text()

			// Parse progress timestamp e.g. "time=00:01:23.45"
			if strings.Contains(line, "time=") {
				timeIdx := strings.Index(line, "time=")
				if timeIdx != -1 {
					timeStr := line[timeIdx+5:]
					spaceIdx := strings.Index(timeStr, " ")
					if spaceIdx != -1 {
						timeStr = timeStr[:spaceIdx]
					}

					parsedSec, err := ParseDurationString(timeStr)
					if err == nil && totalDurationSec > 0 {
						pct := parsedSec / totalDurationSec
						if pct > 1.0 {
							pct = 1.0
						}
						outChan <- ProgressMessage{Percent: pct}
					}
				}
			}
		}

		if err := cmd.Wait(); err != nil {
			outChan <- ProgressMessage{Err: fmt.Errorf("ffmpeg process exited with error: %w", err)}
			return
		}

		outChan <- ProgressMessage{Percent: 1.0, Done: true}
	}()

	return outChan
}
