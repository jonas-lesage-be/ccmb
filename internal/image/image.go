package image

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"ccmb/internal/config"
)

// Converter is responsible for converting image files.
type Converter struct {
	TargetDir string

	MaxWorkers int
	Timeout    time.Duration

	ImageExtensions          map[string]bool
	SupportedImageExtensions map[string]bool
}

type job struct {
	name string
	path string
}

// NewConverter creates a new instance of Converter using the application configuration.
func NewConverter(cfg *config.Config) *Converter {
	return &Converter{
		TargetDir: cfg.TargetDir,

		MaxWorkers: cfg.MaxImageWorkers,
		Timeout:    cfg.ImageConversionTimeout,

		ImageExtensions:          cfg.ImageExtensions,
		SupportedImageExtensions: cfg.SupportedImageExtensions,
	}
}

// Execute converts unsupported image files to supported formats.
func (c *Converter) Execute(ctx context.Context) error {
	slog.Info("Converting unsupported image files")
	files, err := os.ReadDir(c.TargetDir)
	if err != nil {
		return fmt.Errorf("failed to read target directory: %w", err)
	}

	jobs := c.jobsToProcess(ctx, files)
	if len(jobs) == 0 {
		return nil
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(c.MaxWorkers)

	for _, j := range jobs {
		g.Go(func() error {
			if err := c.processFile(ctx, j.name, j.path); err != nil {
				return fmt.Errorf("error converting %s: %w", j.name, err)
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("image conversion failed: %w", err)
	}

	return nil
}

func (c *Converter) jobsToProcess(ctx context.Context, files []os.DirEntry) []job {
	var jobs []job

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		path := filepath.Join(c.TargetDir, file.Name())
		ext := strings.ToLower(filepath.Ext(file.Name()))

		if c.SupportedImageExtensions[ext] {
			continue
		}

		if c.shouldProcess(ctx, path) {
			jobs = append(jobs, job{name: file.Name(), path: path})
		}
	}

	return jobs
}

func (c *Converter) shouldProcess(ctx context.Context, path string) bool {
	isImage, err := c.checkIsImage(ctx, path)
	if err != nil {
		return false
	}

	return isImage
}

func (c *Converter) processFile(ctx context.Context, fileName, path string) error {
	if err := c.convert(ctx, fileName, path); err != nil {
		return fmt.Errorf("failed to convert image: %w", err)
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete original image file: %w", err)
	}

	return nil
}
