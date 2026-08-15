package visualmerge

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sync/errgroup"

	"ccmb/internal/config"
)

// Merger is responsible for merging images and PDFs into size-constrained PDF files.
type Merger struct {
	TargetDir       string
	MaxWorkers      int
	MaxPDFFileBytes int64

	FlatPathDelimiter string
	EscapedDelimiter  string
	DecodePlaceholder string

	VisualExtensions map[string]bool
}

type job struct {
	index int
	name  string
	path  string
}

type jobResult struct {
	path string
	size int64
}

// NewMerger creates a new instance of Merger using the application configuration.
func NewMerger(cfg *config.Config) *Merger {
	return &Merger{
		TargetDir:       cfg.TargetDir,
		MaxWorkers:      cfg.MaxVisualMergeWorkers,
		MaxPDFFileBytes: cfg.MaxPDFFileBytes,

		FlatPathDelimiter: cfg.FlatPathDelimiter,
		EscapedDelimiter:  cfg.EscapedDelimiter,
		DecodePlaceholder: cfg.DecodePlaceholder,

		VisualExtensions: cfg.VisualExtensions,
	}
}

// Execute merges images and PDFs into size-constrained PDF files.
func (m *Merger) Execute(ctx context.Context, tmpDir string) error {
	slog.Debug("Converting and merging visual data")

	files, err := os.ReadDir(tmpDir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", tmpDir, err)
	}

	jobs := m.filterFiles(files, tmpDir)
	if len(jobs) == 0 {
		return nil
	}

	if err := m.processFiles(ctx, jobs); err != nil {
		return fmt.Errorf("failed to process visual files: %w", err)
	}

	originalPaths := make([]string, len(jobs))
	for i, j := range jobs {
		originalPaths[i] = j.path
	}
	if err := cleanUpFiles(originalPaths); err != nil {
		return fmt.Errorf("failed to clean up temporary visual files: %w", err)
	}

	return nil
}

func (m *Merger) filterFiles(files []os.DirEntry, baseDir string) []job {
	var jobs []job
	indexCounter := 0

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if m.VisualExtensions != nil && !m.VisualExtensions[ext] {
			continue
		}

		jobs = append(jobs, job{
			index: indexCounter,
			name:  name,
			path:  filepath.Join(baseDir, name),
		})
		indexCounter++
	}

	return jobs
}

func (m *Merger) processFiles(ctx context.Context, jobs []job) error {
	results := make([]jobResult, len(jobs))
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(m.MaxWorkers)

	for _, j := range jobs {
		g.Go(func() error {
			return m.processJob(ctx, results, j)
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to process visual files concurrently: %w", err)
	}

	if err := m.mergeBatches(results); err != nil {
		return fmt.Errorf("failed to merge batches: %w", err)
	}

	return nil
}

func (m *Merger) processJob(ctx context.Context, results []jobResult, j job) error {
	convertedPDF, size, err := m.preparePDFComponent(ctx, j)
	if err != nil {
		slog.Warn("Skipping asset due to generation error", "file", j.name, "err", err)
		return nil
	}
	results[j.index] = jobResult{path: convertedPDF, size: size}
	return nil
}

func (m *Merger) mergeBatches(results []jobResult) error {
	var batch []string
	var currentSizeBytes int64
	partCounter := 1

	for _, res := range results {
		if res.path == "" {
			continue
		}

		if len(batch) > 0 && currentSizeBytes+res.size > m.MaxPDFFileBytes {
			if err := m.mergeBatch(batch, partCounter); err != nil {
				return fmt.Errorf("failed to merge intermediate batch %d: %w", partCounter, err)
			}
			partCounter++
			batch = nil
			currentSizeBytes = 0
		}

		batch = append(batch, res.path)
		currentSizeBytes += res.size
	}

	if len(batch) > 0 {
		if err := m.mergeBatch(batch, partCounter); err != nil {
			return fmt.Errorf("failed to merge final batch %d: %w", partCounter, err)
		}
	}

	return nil
}

func cleanUpFiles(files []string) error {
	var errs []error

	for _, f := range files {
		if err := os.Remove(f); err != nil && !errors.Is(err, os.ErrNotExist) {
			errs = append(errs, fmt.Errorf("failed to delete %s: %w", f, err))
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
