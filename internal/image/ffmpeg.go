package image

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"
)

func (c *Converter) convert(ctx context.Context, name, path, mimeType string) error {
	ext := filepath.Ext(name)
	baseName := strings.TrimSuffix(name, ext)
	outputPath := filepath.Join(c.TargetDir, baseName+".png")

	slog.Debug("Converting unsupported image to PNG", "file", name)

	if strings.HasPrefix(mimeType, "image/svg") {
		return c.convertSVGToPNG(path, outputPath)
	}

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
		// Use PNG codec.
		"-c:v", "png",
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
