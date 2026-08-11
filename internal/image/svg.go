package image

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers"
)

func (c *Converter) convertSVGToPNG(sourcePath, targetPath string) error {
	cleanPath := filepath.Clean(sourcePath)
	file, err := os.Open(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to open SVG file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			slog.Error("failed to close SVG file", "err", err)
		}
	}()

	cvs, err := canvas.ParseSVG(file)
	if err != nil {
		return fmt.Errorf("failed to parse SVG data: %w", err)
	}

	resolution := canvas.DPI(c.SvgCanvasResolution)
	if err := renderers.Write(targetPath, cvs, resolution); err != nil {
		return fmt.Errorf("failed to write PNG file: %w", err)
	}

	return nil
}
