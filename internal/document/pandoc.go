package document

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func (c *Converter) convertFile(ctx context.Context, inputPath, filename string) error {
	outputPath := filepath.Join(c.Dir, filename+".md")
	docMediaDir := filepath.Join(c.Dir, "pandoc_tmp_"+filename)

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

	mediaMappings, err := c.flattenPandocMedia(docMediaDir, filename)
	if err != nil {
		return fmt.Errorf("failed to flatten extracted document media: %w", err)
	}

	if len(mediaMappings) > 0 {
		if err := c.fixMarkdownImageLinks(outputPath, mediaMappings); err != nil {
			return fmt.Errorf("failed to fix markdown image paths: %w", err)
		}
	}

	if err := c.prependHeader(outputPath, filename); err != nil {
		return fmt.Errorf("failed to prepend context header to markdown: %w", err)
	}

	if err := os.Remove(filepath.Clean(inputPath)); err != nil {
		return fmt.Errorf("failed to remove original document after conversion: %w", err)
	}

	return nil
}
