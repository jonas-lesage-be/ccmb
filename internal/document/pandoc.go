package document

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (c *Converter) convertFile(ctx context.Context, inputPath, filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	activeProcessingPath := inputPath
	activeFilename := filename
	originalFilename := filename

	if isLOConvertible(ext) {
		generatedPath, err := c.convertToIntermediate(ctx, inputPath, ext)
		if err != nil {
			return fmt.Errorf("failed to pre-convert file via LibreOffice: %w", err)
		}

		if err := os.Remove(inputPath); err != nil {
			return fmt.Errorf("failed to delete pre-converted raw source file: %w", err)
		}

		activeProcessingPath = generatedPath
		activeFilename = filepath.Base(generatedPath)
		ext = strings.ToLower(filepath.Ext(activeFilename))
	}

	if strings.HasSuffix(strings.ToLower(activeFilename), c.FlatPathDelimiter+".pdf") {
		return nil
	}

	if !isSupportedFileType(ext) {
		return nil
	}

	return c.runPandocConversion(ctx, activeProcessingPath, activeFilename, originalFilename, ext)
}

func (c *Converter) runPandocConversion(
	ctx context.Context,
	srcPath, filename, originalFilename, ext string,
) error {
	fromFormat := strings.TrimPrefix(ext, ".")
	cleanFlatName, outputName := c.derivePandocNames(filename, originalFilename)
	outputPath := filepath.Join(c.Dir, outputName)
	mediaDir := filepath.Join(c.Dir, "pandoc_tmp_"+strings.TrimSuffix(filename, ext))

	args := []string{
		srcPath,
		"-f", fromFormat,
		"-t", "markdown",
		"-o", outputPath,
	}

	isExtractable := isExtractableMediaFile(ext)
	if isExtractable {
		args = append(args, "--extract-media="+mediaDir)
	}

	//nolint:gosec
	cmd := exec.CommandContext(ctx, "pandoc", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pandoc failed to convert %s format to markdown: %w", fromFormat, err)
	}

	if isExtractable {
		if err := c.handleMediaExtraction(mediaDir, cleanFlatName, outputPath); err != nil {
			return fmt.Errorf("failed to handle media extraction: %w", err)
		}
	}

	if err := os.Remove(srcPath); err != nil {
		return fmt.Errorf("failed to remove temporary document after conversion: %w", err)
	}

	return nil
}

func (c *Converter) derivePandocNames(flatName, originalFilename string) (string, string) {
	cleanFlatName := flatName
	if lastIdx := strings.LastIndex(flatName, c.FlatPathDelimiter); lastIdx != -1 {
		lastSegment := flatName[lastIdx+len(c.FlatPathDelimiter):]
		if strings.HasPrefix(lastSegment, ".") {
			cleanFlatName = flatName[:lastIdx]
		}
	}

	return cleanFlatName, originalFilename + c.FlatPathDelimiter + ".md"
}

func (c *Converter) handleMediaExtraction(mediaDir, baseName, markdownPath string) error {
	mediaMappings, err := c.flattenPandocMedia(mediaDir, baseName)
	if err != nil {
		return fmt.Errorf("failed to flatten extracted document media: %w", err)
	}

	if len(mediaMappings) > 0 {
		if err := c.fixMarkdownImageLinks(markdownPath, mediaMappings); err != nil {
			return fmt.Errorf("failed to fix markdown image paths: %w", err)
		}
	}

	return nil
}

func isSupportedFileType(ext string) bool {
	switch ext {
	case ".docx", ".odt", ".rtf", ".epub", ".csv":
		return true
	default:
		return false
	}
}

func isExtractableMediaFile(ext string) bool {
	switch ext {
	case ".docx", ".odt":
		return true
	default:
		return false
	}
}
