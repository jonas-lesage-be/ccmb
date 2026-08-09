package media

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// FrameCount returns the number of frames in a file using ffprobe.
func FrameCount(ctx context.Context, fullPath string) (int, error) {
	//nolint:gosec
	cmd := exec.CommandContext(
		ctx,
		"ffprobe",
		// Verbose level: error.
		"-v", "error",
		// Show streams.
		"-show_streams",
		// Only select video streams.
		"-select_streams", "v",
		// Only show the number of read frames
		"-show_entries", "stream=nb_read_frames",
		// Count the number of frames.
		"-count_frames",
		// Strip the keys from the output.
		"-of", "default=noprint_wrappers=1:nokey=1",
		fullPath,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w", err)
	}

	cleanOutput := strings.TrimSpace(string(output))
	if cleanOutput == "" {
		return 0, nil
	}

	frameCount, err := strconv.Atoi(cleanOutput)
	if err != nil {
		return 0, fmt.Errorf("failed to convert frame count to integer: %w", err)
	}

	return frameCount, nil
}
