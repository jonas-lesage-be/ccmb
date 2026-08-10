package video

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"ccmb/internal/media"
)

func (e *Extractor) shouldProcess(ctx context.Context, fullPath string) bool {
	isVideo, err := e.checkIsVideo(ctx, fullPath)
	if err != nil {
		return false
	}

	return isVideo
}

func (e *Extractor) checkIsVideo(ctx context.Context, fullPath string) (bool, error) {
	if e.VideoExtensions != nil {
		if ext := strings.ToLower(filepath.Ext(fullPath)); !e.VideoExtensions[ext] {
			return false, nil
		}
		return true, nil
	}

	ctx, cancel := context.WithTimeout(ctx, e.VideoExtractTimeout)
	defer cancel()

	frameCount, err := media.FrameCount(ctx, fullPath)
	if err != nil {
		return false, fmt.Errorf("failed to check if file is a video: %w", err)
	}

	return frameCount > 1, nil
}

func (e *Extractor) extractFrames(ctx context.Context, filePath string) error {
	baseName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	outputPattern := filepath.Join(e.TargetDir, baseName+e.FlatPathDelimiter+"frame_%d.jpg")

	log.Printf("Running FFmpeg wrapper onto: %s", filepath.Base(filePath))

	ctx, cancel := context.WithTimeout(ctx, e.VideoExtractTimeout)
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
		"-i", filePath,
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
