package document

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"ccmb/internal/config"
)

// Converter is responsible for converting supported document formats into Markdown.
type Converter struct {
	Dir                 string
	TextFilePermissions os.FileMode

	FlatPathDelimiter  string
	DocumentExtensions map[string]bool
}

// NewConverter creates a new instance of Converter using the application configuration.
func NewConverter(cfg *config.Config) *Converter {
	return &Converter{
		Dir:                 cfg.TargetDir,
		TextFilePermissions: cfg.TextFilePermissions,

		FlatPathDelimiter:  cfg.FlatPathDelimiter,
		DocumentExtensions: cfg.DocumentExtensions,
	}
}

// Execute scans the target directory for supported document formats and converts them to Markdown.
func (c *Converter) Execute(ctx context.Context) error {
	if _, err := exec.LookPath("pandoc"); err != nil {
		return fmt.Errorf("pandoc is not installed or not found in PATH: %w", err)
	}

	slog.Debug("Converting supported documents to Markdown")

	files, err := os.ReadDir(c.Dir)
	if err != nil {
		return fmt.Errorf("failed to read input directory: %w", err)
	}

	for _, file := range files {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context error while processing files: %w", err)
		}

		if file.IsDir() {
			continue
		}

		name := file.Name()
		ext := filepath.Ext(name)
		if !c.DocumentExtensions[ext] {
			continue
		}

		inputPath := filepath.Join(c.Dir, name)
		if err := c.convertFile(ctx, inputPath, name); err != nil {
			return fmt.Errorf("failed to convert file %s: %w", name, err)
		}
	}

	return nil
}

func (c *Converter) prependHeader(filePath, docName string) error {
	cleanPath := filepath.Clean(filePath)
	content, err := os.ReadFile(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to read markdown file %s: %w", cleanPath, err)
	}

	header := fmt.Sprintf("# Document Source: %s\n\n", docName)
	newContent := append([]byte(header), content...)

	//nolint:gosec
	if err := os.WriteFile(cleanPath, newContent, c.TextFilePermissions); err != nil {
		return fmt.Errorf("failed to write markdown file %s: %w", cleanPath, err)
	}

	return nil
}

func (c *Converter) flattenPandocMedia(docMediaDir, baseName string) (map[string]string, error) {
	mappings := make(map[string]string)

	cleanMediaDir := filepath.Clean(docMediaDir)
	if _, err := os.Stat(cleanMediaDir); os.IsNotExist(err) {
		return mappings, nil
	}

	//nolint:gosec
	err := filepath.WalkDir(cleanMediaDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}

		safeFilename := filepath.Base(path)

		relPath, err := filepath.Rel(cleanMediaDir, path)
		if err != nil {
			return fmt.Errorf("failed to determine relative path for %s: %w", path, err)
		}
		pandocReference := filepath.Base(cleanMediaDir) + "/" + filepath.ToSlash(relPath)

		flattenedImageName := fmt.Sprintf("%s%s%s", baseName, c.FlatPathDelimiter, safeFilename)
		newHomePath := filepath.Join(c.Dir, flattenedImageName)

		if err := os.Rename(path, newHomePath); err != nil {
			return fmt.Errorf(
				"failed to move extracted media file %s to %s: %w",
				path,
				newHomePath,
				err,
			)
		}

		mappings[pandocReference] = flattenedImageName
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to process extracted media assets recursively: %w", err)
	}

	if err := os.RemoveAll(cleanMediaDir); err != nil {
		return nil, fmt.Errorf("failed to remove temporary media directory tree: %w", err)
	}

	return mappings, nil
}

func (c *Converter) fixMarkdownImageLinks(markdownPath string, mappings map[string]string) error {
	cleanPath := filepath.Clean(markdownPath)
	content, err := os.ReadFile(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to read markdown file %s: %w", cleanPath, err)
	}

	const pairCount = 2
	replacements := make([]string, 0, len(mappings)*pairCount)
	for oldPath, newFileName := range mappings {
		replacements = append(replacements, oldPath, newFileName)
	}

	replacer := strings.NewReplacer(replacements...)
	text := replacer.Replace(string(content))

	//nolint:gosec
	if err := os.WriteFile(cleanPath, []byte(text), c.TextFilePermissions); err != nil {
		return fmt.Errorf("failed to write markdown file %s: %w", cleanPath, err)
	}

	return nil
}
