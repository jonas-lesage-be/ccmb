package image

import (
	"fmt"
	"image"
	"image/gif"
	"image/png"
	"log/slog"
	"os"
	"path/filepath"
)

func (c *Converter) processGIF(srcPath, name string) error {
	in, err := os.Open(filepath.Clean(srcPath))
	if err != nil {
		return fmt.Errorf("failed to open source GIF: %w", err)
	}

	gifData, err := gif.DecodeAll(in)
	if err != nil {
		if closeErr := in.Close(); closeErr != nil {
			return fmt.Errorf(
				"failed to decode animated GIF streams: %w and failed to close file: %w",
				err,
				closeErr,
			)
		}
		return fmt.Errorf("failed to decode animated GIF streams: %w", err)
	}

	if err := in.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	frameCount := len(gifData.Image)
	slog.Debug("Processing GIF file", "file", name, "discovered_frames", frameCount)

	for i, frameImg := range gifData.Image {
		outputName := c.resolveFrameName(name, frameCount, i)
		outputPath := filepath.Join(c.Dir, outputName)

		if err := c.writeFrameToPNG(outputPath, frameImg, i); err != nil {
			return err
		}
	}

	if err := os.Remove(srcPath); err != nil {
		return fmt.Errorf("failed to delete original GIF file %s: %w", name, err)
	}

	return nil
}

func (c *Converter) resolveFrameName(name string, frameCount, index int) string {
	if frameCount == 1 {
		return name + c.FlatPathDelimiter + ".png"
	}
	return fmt.Sprintf("%s%sframe_%d.png", name, c.FlatPathDelimiter, index)
}

func (c *Converter) writeFrameToPNG(outputPath string, img image.Image, index int) error {
	out, err := os.OpenFile(
		filepath.Clean(outputPath),
		os.O_CREATE|os.O_EXCL|os.O_WRONLY,
		c.FilePermissions,
	)
	if err != nil {
		return fmt.Errorf("failed to create frame target PNG %d: %w", index, err)
	}

	encoder := png.Encoder{CompressionLevel: png.DefaultCompression}
	if err := encoder.Encode(out, img); err != nil {
		if closeErr := out.Close(); closeErr != nil {
			return fmt.Errorf(
				"failed to encode frame block %d to PNG: %w and failed to close file: %w",
				index,
				closeErr,
				err,
			)
		}
		return fmt.Errorf("failed to encode frame block %d to PNG: %w", index, err)
	}

	if err := out.Close(); err != nil {
		return fmt.Errorf("failed to close frame target PNG %d: %w", index, err)
	}

	return nil
}
