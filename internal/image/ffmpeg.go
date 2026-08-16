package image

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"

	"ccmb/internal/media"
)

func (c *Converter) convert(ctx context.Context, name, path, mimeType string) error {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg is not installed or not found in PATH: %w", err)
	}

	ext := filepath.Ext(name)
	baseName := strings.TrimSuffix(name, ext)
	outputPath := filepath.Join(c.Dir, baseName+".png")

	slog.Debug("Converting unsupported image to PNG", "file", name)

	if strings.HasPrefix(mimeType, "image/svg") {
		return c.convertSVGToPNG(path, outputPath)
	}

	args := []string{
		// Use hardware acceleration if available.
		"-hwaccel", "auto",
		// Input file.
		"-i", path,
		// Use PNG codec.
		"-c:v", "png",
		// Use all available threads.
		"-threads", "0",
		// Overwrite output file if it exists.
		"-y",
		// Output file.
		outputPath,
	}

	if err := media.RunFFmpegCommand(ctx, c.Timeout, args...); err != nil {
		return fmt.Errorf("failed to convert image: %w", err)
	}

	return nil
}
