package document

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

func (c *Converter) convertFile(ctx context.Context, inputPath, filename, ext string) error {
	baseName := strings.TrimSuffix(filename, ext)
	outputPath := filepath.Join(c.Dir, baseName+".md")
	docMediaDir := filepath.Join(c.Dir, baseName+"_media")

	args := []string{
		inputPath,
		"-t", "markdown",
		"--extract-media=" + docMediaDir,
		"-o", outputPath,
	}

	//nolint:gosec
	cmd := exec.CommandContext(ctx, "pandoc", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pandoc command failed: %w", err)
	}

	mediaMappings, err := c.flattenPandocMedia(docMediaDir, baseName)
	if err != nil {
		return fmt.Errorf("failed to flatten extracted document media: %w", err)
	}

	if len(mediaMappings) > 0 {
		if err := c.fixMarkdownImageLinks(outputPath, mediaMappings); err != nil {
			return fmt.Errorf("failed to fix markdown image paths: %w", err)
		}
	}

	if err := c.prependHeader(outputPath, baseName); err != nil {
		return fmt.Errorf("failed to prepend context header to markdown: %w", err)
	}

	return nil
}
