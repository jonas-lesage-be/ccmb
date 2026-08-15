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
		if err := c.convertFile(ctx, inputPath, name, ext); err != nil {
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

	nestedMediaFolder := filepath.Join(cleanMediaDir, "word", "media")
	dirFile, err := os.Open(filepath.Clean(nestedMediaFolder))
	if os.IsNotExist(err) {
		dirFile, err = os.Open(cleanMediaDir)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to open media directory handle: %w", err)
	}

	filenames, err := dirFile.Readdirnames(-1)
	if err != nil {
		tryCloseMediaDir(dirFile)
		return nil, fmt.Errorf("failed to read media filenames: %w", err)
	}

	for _, filename := range filenames {
		safeFilename := filepath.Base(filename)

		var pandocReference string
		var currentPath string

		if strings.Contains(dirFile.Name(), "word") {
			pandocReference = filepath.Join(
				filepath.Base(cleanMediaDir),
				"word",
				"media",
				safeFilename,
			)
			currentPath = filepath.Join(cleanMediaDir, "word", "media", safeFilename)
		} else {
			pandocReference = filepath.Join(filepath.Base(cleanMediaDir), safeFilename)
			currentPath = filepath.Join(cleanMediaDir, safeFilename)
		}

		flattenedImageName := fmt.Sprintf("%s%s%s", baseName, c.FlatPathDelimiter, safeFilename)
		newHomePath := filepath.Join(filepath.Clean(c.Dir), flattenedImageName)

		if err := os.Rename(filepath.Clean(currentPath), newHomePath); err != nil {
			tryCloseMediaDir(dirFile)
			return nil, fmt.Errorf("failed to move media asset: %w", err)
		}

		mappings[filepath.ToSlash(pandocReference)] = flattenedImageName
		mappings[filepath.FromSlash(pandocReference)] = flattenedImageName
	}

	tryCloseMediaDir(dirFile)

	if err := os.RemoveAll(cleanMediaDir); err != nil {
		return nil, fmt.Errorf("failed to remove media directory: %w", err)
	}

	return mappings, nil
}

func tryCloseMediaDir(f *os.File) {
	if err := f.Close(); err != nil {
		slog.Error("failed to close media directory handle", "err", err)
	}
}

func (c *Converter) fixMarkdownImageLinks(markdownPath string, mappings map[string]string) error {
	cleanPath := filepath.Clean(markdownPath)
	content, err := os.ReadFile(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to read markdown file %s: %w", cleanPath, err)
	}

	text := string(content)
	for oldPath, newFileName := range mappings {
		text = strings.ReplaceAll(text, oldPath, newFileName)
	}

	//nolint:gosec
	if err := os.WriteFile(cleanPath, []byte(text), c.TextFilePermissions); err != nil {
		return fmt.Errorf("failed to write markdown file %s: %w", cleanPath, err)
	}

	return nil
}
