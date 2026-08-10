package visualmerge

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"ccmb/internal/config"
)

// Merger is responsible for merging images and PDFs into size-constrained PDF files.
type Merger struct {
	TargetDir       string
	MaxPDFFileBytes int64

	FlatPathDelimiter string
	EscapedDelimiter  string
	DecodePlaceholder string

	VisualExtensions map[string]bool
}

// NewMerger creates a new instance of Merger using the application configuration.
func NewMerger(cfg *config.Config) *Merger {
	return &Merger{
		TargetDir:       cfg.TargetDir,
		MaxPDFFileBytes: cfg.MaxPDFFileBytes,

		FlatPathDelimiter: cfg.FlatPathDelimiter,
		EscapedDelimiter:  cfg.EscapedDelimiter,
		DecodePlaceholder: cfg.DecodePlaceholder,

		VisualExtensions: cfg.VisualExtensions,
	}
}

// Execute merges images and PDFs into size-constrained PDF files.
func (m *Merger) Execute(ctx context.Context) error {
	slog.Info("Converting and merging visual data")
	files, err := os.ReadDir(m.TargetDir)
	if err != nil {
		return fmt.Errorf("failed to read target directory %s: %w", m.TargetDir, err)
	}

	visualFiles := m.filterVisualFiles(files)
	if err := m.processVisualFiles(ctx, visualFiles); err != nil {
		return fmt.Errorf("failed to process visual files: %w", err)
	}
	if err := cleanUpFiles(visualFiles); err != nil {
		return fmt.Errorf("failed to clean up original visual files: %w", err)
	}

	return nil
}

func (m *Merger) filterVisualFiles(files []os.DirEntry) []string {
	var visualFiles []string

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if ext := strings.ToLower(filepath.Ext(file.Name())); m.VisualExtensions[ext] {
			visualFiles = append(visualFiles, filepath.Join(m.TargetDir, file.Name()))
		}
	}

	return visualFiles
}

func (m *Merger) processVisualFiles(ctx context.Context, files []string) error {
	var currentBatch []string
	var currentSizeBytes int64
	partCounter := 1

	for _, path := range files {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context error while processing file %s: %w", path, err)
		}

		convertedPDF, size, err := m.preparePDFComponent(ctx, path)
		if err != nil {
			slog.Warn("Skipping asset", "file", filepath.Base(path), "error", err)
			continue
		}

		if len(currentBatch) > 0 && currentSizeBytes+size > m.MaxPDFFileBytes {
			if err := mergeBatch(currentBatch, m.TargetDir, partCounter); err != nil {
				return fmt.Errorf("failed to merge intermediate batch %d: %w", partCounter, err)
			}
			partCounter++
			currentBatch = nil
			currentSizeBytes = 0
		}

		currentBatch = append(currentBatch, convertedPDF)
		currentSizeBytes += size
	}

	if len(currentBatch) > 0 {
		if err := mergeBatch(currentBatch, m.TargetDir, partCounter); err != nil {
			return fmt.Errorf("failed to merge final batch %d: %w", partCounter, err)
		}
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
