package video

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"
)

func (e *Extractor) extractFrames(ctx context.Context, path string) error {
	baseName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	outputPattern := filepath.Join(e.TargetDir, baseName+e.FlatPathDelimiter+"frame_%d.jpg")

	slog.Debug("Running FFmpeg to extract frames", "file", filepath.Base(path))

	ctx, cancel := context.WithTimeout(ctx, e.Timeout)
	defer cancel()

	//nolint:gosec
	cmd := exec.CommandContext(
		ctx,
		"ffmpeg",
		// Use hardware acceleration if available.
		"-hwaccel", "auto",
		// No audio and subtitle streams.
		"-an",
		"-sn",
		// Input file.
		"-i", path,
		// Extract 1 frame per second using variable frame rate mode.
		"-vf", "fps=1",
		"-fps_mode", "vfr",
		// Use high quality.
		"-q:v", "2",
		// Use all available threads.
		"-threads", "0",
		// Overwrite output file if it exists.
		"-y",
		// Output pattern.
		outputPattern,
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run ffmpeg command: %w", err)
	}

	return nil
}
