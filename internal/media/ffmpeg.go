package media

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

func runFFprobeCommand(ctx context.Context, path string, timeout time.Duration) ([]byte, error) {
	if _, err := exec.LookPath("ffprobe"); err != nil {
		return nil, fmt.Errorf("ffprobe is not installed or not found in PATH: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	args := []string{
		// Verbose level: error.
		"-v", "error",
		// Show streams.
		"-select_streams", "v",
		// Only decode the first 2 frames of the stream.
		"-read_intervals", "%+#2",
		// Count the number of frames.
		"-count_frames",
		// Only show the number of read frames
		"-show_entries", "stream=nb_read_frames",
		// Strip the keys from the output.
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	}

	//nolint:gosec
	cmd := exec.CommandContext(ctx, "ffprobe", args...)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w", err)
	}

	return output, nil
}

func RunFFmpegCommand(ctx context.Context, timeout time.Duration, args ...string) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	//nolint:gosec
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run ffmpeg command: %w", err)
	}

	return nil
}
