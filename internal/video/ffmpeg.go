package video

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
)

func (e *Extractor) shouldProcessVideo(ctx context.Context, fullPath string) bool {
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

	//nolint:gosec
	cmd := exec.CommandContext(
		ctx,
		"ffprobe",
		"-v", "error",
		"-show_streams",
		"-select_streams", "v", // Only select video streams.
		"-show_entries", "stream=index", // Only output the stream index.
		"-of", "default=noprint_wrappers=1:nokey=1", // Strip the output.
		fullPath,
	)

	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("ffprobe command failed: %w", err)
	}

	cleanOutput := strings.TrimSpace(string(output))
	return cleanOutput != "", nil
}

func (e *Extractor) extractFrames(ctx context.Context, videoPath string) error {
	baseName := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))
	outputPattern := filepath.Join(e.TargetDir, baseName+"--frame_%d.jpg")

	log.Printf("Running FFmpeg wrapper onto: %s\n", filepath.Base(videoPath))

	ctx, cancel := context.WithTimeout(ctx, e.VideoExtractTimeout)
	defer cancel()

	//nolint:gosec
	cmd := exec.CommandContext(
		ctx,
		"ffmpeg",
		"-hwaccel", "auto",
		"-an",
		"-sn",
		"-i", videoPath,
		"-vf", "fps=1",
		"-fps_mode", "vfr",
		"-q:v", "2",
		"-threads", "0",
		"-start_number", "0",
		outputPattern,
		"-y",
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run ffmpeg command: %w", err)
	}

	return nil
}
