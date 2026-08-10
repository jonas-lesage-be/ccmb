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

func (c *Converter) checkIsImage(ctx context.Context, path string) (bool, error) {
	if c.ImageExtensions != nil {
		if ext := strings.ToLower(filepath.Ext(path)); !c.ImageExtensions[ext] {
			return false, nil
		}
		return true, nil
	}

	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	frameCount, err := media.FrameCount(ctx, path)
	if err != nil {
		return false, fmt.Errorf("failed to check if file is an image: %w", err)
	}

	return frameCount == 1, nil
}

func (c *Converter) convert(ctx context.Context, name, path string) error {
	baseName := strings.TrimSuffix(name, filepath.Ext(name))
	outputPath := filepath.Join(c.TargetDir, baseName+".png")

	slog.Debug("Converting unsupported image to PNG", "file", name)

	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	//nolint:gosec
	cmd := exec.CommandContext(
		ctx,
		"ffmpeg",
		// Use hardware acceleration if available.
		"-hwaccel", "auto",
		// Input file.
		"-i", path,
		// Use high quality.
		"-q:v", "2",
		// Use all available threads.
		"-threads", "0",
		// Overwrite output file if it exists.
		"-y",
		// Output file.
		outputPath,
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run ffmpeg command: %w", err)
	}

	return nil
}
