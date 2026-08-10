package textmerge

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"ccmb/internal/config"
	"ccmb/internal/conv"
	"ccmb/internal/pathsafe"
)

// Merger is responsible for merging text files into size-constrained text files.
type Merger struct {
	FlatPathDelimiter string
	EscapedDelimiter  string
	DecodePlaceholder string

	MaxTextFileBytes    int64
	TargetDir           string
	TextFilePermissions os.FileMode
}

// ErrOutsideTargetDir is returned when a file is found outside the target directory.
var ErrOutsideTargetDir = errors.New("file is outside target directory")

// NewMerger creates a new instance of Merger using the application configuration.
func NewMerger(cfg *config.Config) *Merger {
	return &Merger{
		FlatPathDelimiter: cfg.FlatPathDelimiter,
		EscapedDelimiter:  cfg.EscapedDelimiter,
		DecodePlaceholder: cfg.DecodePlaceholder,

		MaxTextFileBytes:    cfg.MaxTextFileBytes,
		TargetDir:           cfg.TargetDir,
		TextFilePermissions: cfg.TextFilePermissions,
	}
}

// Execute merges all text files (.py, .js, .txt) into size-constrained text files.
func (m *Merger) Execute(ctx context.Context) error {
	slog.Info("Merging all text files")

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error before starting text merge: %w", err)
	}

	files, err := os.ReadDir(m.TargetDir)
	if err != nil {
		return fmt.Errorf("failed to read target directory: %w", err)
	}

	textBatch, errs := m.processFiles(ctx, files)

	if err := cleanUpFiles(textBatch); err != nil {
		errs = append(errs, fmt.Errorf("failed to clean up text files: %w", err))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (m *Merger) processFiles(ctx context.Context, files []os.DirEntry) ([]string, []error) {
	var textBatch []string
	var errs []error
	var currentBuilder strings.Builder

	maxSizeBytes := m.MaxTextFileBytes * conv.MiB
	partCounter := 1

	for _, file := range files {
		if err := ctx.Err(); err != nil {
			errs = append(errs, fmt.Errorf("context error during text merge: %w", err))
			break
		}

		if m.shouldSkip(file) {
			continue
		}

		fullPath := filepath.Join(m.TargetDir, file.Name())
		textBatch = append(textBatch, fullPath)

		fileBlock, err := m.buildFileBlock(file.Name(), fullPath)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		if int64(currentBuilder.Len()+len(fileBlock)) > maxSizeBytes {
			m.flush(&currentBuilder, partCounter, &errs)
			partCounter++
			currentBuilder.Reset()
		}

		currentBuilder.WriteString(fileBlock)
	}

	if ctx.Err() == nil {
		m.flush(&currentBuilder, partCounter, &errs)
	}

	return textBatch, errs
}

func (m *Merger) shouldSkip(file os.DirEntry) bool {
	return file.IsDir() || strings.HasPrefix(file.Name(), "FinalResult_")
}

func (m *Merger) buildFileBlock(fileName, fullPath string) (string, error) {
	cleanPath := filepath.Clean(fullPath)

	if !pathsafe.Contains(m.TargetDir, cleanPath) {
		return "", fmt.Errorf("failed to process %s: %w", fileName, ErrOutsideTargetDir)
	}

	content, err := os.ReadFile(cleanPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", fileName, err)
	}

	originalPath := config.DecodeFlatName(
		fileName,
		m.EscapedDelimiter,
		m.DecodePlaceholder,
		m.FlatPathDelimiter,
	)

	return fmt.Sprintf(
		"<file path=\"%s\">\n%s\n</file>\n\n",
		originalPath,
		string(content),
	), nil
}

func (m *Merger) flush(b *strings.Builder, counter int, errs *[]error) {
	if b.Len() > 0 {
		if err := m.writeTextPart(b.String(), counter); err != nil {
			*errs = append(*errs, err)
		}
	}
}

func (m *Merger) writeTextPart(data string, counter int) error {
	outputName := filepath.Join(m.TargetDir, fmt.Sprintf("FinalResult_Text_Part_%d.txt", counter))
	slog.Info("Saving merged structural textual payload", "file", filepath.Base(outputName))

	if err := os.WriteFile(outputName, []byte(data), m.TextFilePermissions); err != nil {
		return fmt.Errorf("failed to write text part %d: %w", counter, err)
	}

	return nil
}

func cleanUpFiles(files []string) error {
	var errs []error

	for _, f := range files {
		if err := os.Remove(f); err != nil {
			errs = append(errs, fmt.Errorf("failed to delete %s: %w", f, err))
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
