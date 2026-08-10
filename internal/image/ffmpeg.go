package image

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"ccmb/internal/media"
)

func (c *Converter) checkIsImage(ctx context.Context, fullPath string) (bool, error) {
	if c.ImageExtensions != nil {
		if ext := strings.ToLower(filepath.Ext(fullPath)); !c.ImageExtensions[ext] {
			return false, nil
		}
		return true, nil
	}

	ctx, cancel := context.WithTimeout(ctx, c.ImageConversionTimeout)
	defer cancel()

	frameCount, err := media.FrameCount(ctx, fullPath)
	if err != nil {
		return false, fmt.Errorf("failed to check if file is an image: %w", err)
	}

	return frameCount == 1, nil
}

func (c *Converter) convert(ctx context.Context, fileName, fullPath string) error {
	baseName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	outputPath := filepath.Join(c.TargetDir, baseName+".png")

	log.Printf("Converting unsupported image to PNG: %s", fileName)

	ctx, cancel := context.WithTimeout(ctx, c.ImageConversionTimeout)
	defer cancel()

	//nolint:gosec
	cmd := exec.CommandContext(
		ctx,
		"ffmpeg",
		// Use hardware acceleration if available.
		"-hwaccel", "auto",
		// Input file.
		"-i", fullPath,
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
