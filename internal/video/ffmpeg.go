package video

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"

	"ccmb/internal/media"
)

func (e *Extractor) extractFrames(ctx context.Context, path string) error {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg is not installed or not found in PATH: %w", err)
	}

	filename := filepath.Base(path)
	outputPattern := filepath.Join(e.Dir, filename+e.FlatPathDelimiter+"frame_%d.jpg")

	slog.Debug("Running FFmpeg to extract frames", "file", filename)

	fpsArg := fmt.Sprintf("fps=%g", e.Fps)
	args := []string{
		// Use hardware acceleration if available.
		"-hwaccel", "auto",
		// No audio and subtitle streams.
		"-an",
		"-sn",
		// Input file.
		"-i", path,
		// Extract frames using variable frame rate mode.
		"-vf", fpsArg,
		"-fps_mode", "vfr",
		// Use high quality.
		"-q:v", "2",
		// Use all available threads.
		"-threads", "0",
		// Overwrite output file if it exists.
		"-y",
		// Output pattern.
		outputPattern,
	}

	if err := media.RunFFmpegCommand(ctx, e.Timeout, args...); err != nil {
		return fmt.Errorf("failed to extract frames: %w", err)
	}

	return nil
}
