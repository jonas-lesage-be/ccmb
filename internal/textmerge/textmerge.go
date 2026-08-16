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
	"ccmb/internal/pathsafe"
)

// Merger is responsible for merging text files into size-constrained text files.
type Merger struct {
	TargetDir           string
	TextFilePermissions os.FileMode

	VisualPartPrefix string
	TextPartPrefix   string
	MaxTextFileBytes int64

	FlatPathDelimiter string
	EscapedDelimiter  string
	DecodePlaceholder string
}

// ErrOutsideTargetDir is returned when a file is found outside the target directory.
var ErrOutsideTargetDir = errors.New("file is outside target directory")

// NewMerger creates a new instance of Merger using the application configuration.
func NewMerger(cfg *config.Config) *Merger {
	return &Merger{
		TargetDir:           cfg.TargetDir,
		TextFilePermissions: cfg.TextFilePermissions,

		VisualPartPrefix: cfg.VisualPartPrefix,
		TextPartPrefix:   cfg.TextPartPrefix,
		MaxTextFileBytes: cfg.MaxTextFileBytes,

		FlatPathDelimiter: cfg.FlatPathDelimiter,
		EscapedDelimiter:  cfg.EscapedDelimiter,
		DecodePlaceholder: cfg.DecodePlaceholder,
	}
}

// Execute merges all text files (.py, .js, .txt) into size-constrained text files.
func (m *Merger) Execute(ctx context.Context, tmpDir string) error {
	slog.Debug("Merging all text files")

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error before starting text merge: %w", err)
	}

	files, err := os.ReadDir(tmpDir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", tmpDir, err)
	}

	batch, errs := m.processFiles(ctx, files, tmpDir)
	if err := cleanUpFiles(batch); err != nil {
		errs = append(errs, fmt.Errorf("failed to clean up text files: %w", err))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (m *Merger) processFiles(
	ctx context.Context,
	files []os.DirEntry,
	baseDir string,
) ([]string, []error) {
	var batch []string
	var errs []error
	var builder strings.Builder

	partCounter := 1

	for _, file := range files {
		if err := ctx.Err(); err != nil {
			errs = append(errs, fmt.Errorf("context error during text merge: %w", err))
			break
		}

		if m.shouldSkip(file) {
			continue
		}

		name := file.Name()
		path := filepath.Join(baseDir, name)

		cleanPath := filepath.Clean(path)
		if !pathsafe.Contains(baseDir, cleanPath) {
			errs = append(
				errs,
				fmt.Errorf("failed to process %s: %w", name, ErrOutsideTargetDir),
			)
			continue
		}

		content, err := os.ReadFile(cleanPath)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to read file %s: %w", name, err))
			continue
		}

		originalPath := m.originalPathFromFlatName(name)
		builderLen := int64(builder.Len())
		expectedLen := builderLen + m.writeFileLen(originalPath, content)

		if expectedLen > m.MaxTextFileBytes && builderLen > 0 {
			m.flush(&builder, partCounter, &errs)
			partCounter++
		}

		batch = append(batch, path)

		if _, err := m.writeFile(&builder, originalPath, content); err != nil {
			errs = append(errs, fmt.Errorf("failed to write file %s: %w", name, err))
		}
	}

	if ctx.Err() == nil {
		m.flush(&builder, partCounter, &errs)
	}

	return batch, errs
}

func (m *Merger) shouldSkip(file os.DirEntry) bool {
	name := file.Name()
	return file.IsDir() ||
		strings.HasPrefix(name, m.VisualPartPrefix) ||
		strings.HasPrefix(name, m.TextPartPrefix)
}

func (m *Merger) originalPathFromFlatName(flatName string) string {
	return config.DecodeFlatName(
		flatName,
		m.EscapedDelimiter,
		m.DecodePlaceholder,
		m.FlatPathDelimiter,
	)
}

func (m *Merger) writeFileLen(path string, data []byte) int64 {
	headerLength := len("<file path=\"") + len(path) + len("\">\n")
	footerLength := len("\n</file>\n\n")
	return int64(headerLength + len(data) + footerLength)
}

func (m *Merger) writeFile(builder *strings.Builder, path string, data []byte) (int, error) {
	totalWritten := 0

	currentWritten, err := builder.WriteString("<file path=\"")
	if err != nil {
		return totalWritten, fmt.Errorf("failed to write file path to builder: %w", err)
	}
	totalWritten += currentWritten

	currentWritten, err = builder.WriteString(path)
	if err != nil {
		return totalWritten, fmt.Errorf("failed to write file path to builder: %w", err)
	}
	totalWritten += currentWritten

	currentWritten, err = builder.WriteString("\">\n")
	if err != nil {
		return totalWritten, fmt.Errorf("failed to write file header to builder: %w", err)
	}
	totalWritten += currentWritten

	currentWritten, err = builder.Write(data)
	if err != nil {
		return totalWritten, fmt.Errorf("failed to write file content to builder: %w", err)
	}
	totalWritten += currentWritten

	currentWritten, err = builder.WriteString("\n</file>\n\n")
	if err != nil {
		return totalWritten, fmt.Errorf("failed to write file footer to builder: %w", err)
	}
	totalWritten += currentWritten

	return totalWritten, nil
}

func (m *Merger) flush(b *strings.Builder, counter int, errs *[]error) {
	if b.Len() > 0 {
		if err := m.writeTextPart(b.String(), counter); err != nil {
			*errs = append(*errs, err)
		}
		b.Reset()
	}
}

func (m *Merger) writeTextPart(data string, counter int) error {
	outputName := filepath.Join(m.TargetDir, fmt.Sprintf("%s%d.txt", m.TextPartPrefix, counter))
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
