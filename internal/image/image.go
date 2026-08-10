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
	MaxImageWorkerCount      int
	ImageConversionTimeout   time.Duration
	ImageExtensions          map[string]bool
	SupportedImageExtensions map[string]bool
	TargetDir                string
}

type job struct {
	fileName string
	fullPath string
}

// NewConverter creates a new instance of Converter using the application configuration.
func NewConverter(cfg *config.Config) *Converter {
	return &Converter{
		MaxImageWorkerCount:      cfg.MaxImageWorkerCount,
		ImageConversionTimeout:   cfg.ImageConversionTimeout,
		ImageExtensions:          cfg.ImageExtensions,
		SupportedImageExtensions: cfg.SupportedImageExtensions,
		TargetDir:                cfg.TargetDir,
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
	g.SetLimit(c.MaxImageWorkerCount)

	for _, j := range jobs {
		g.Go(func() error {
			if err := c.processFile(ctx, j.fileName, j.fullPath); err != nil {
				return fmt.Errorf("error converting %s: %w", j.fileName, err)
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

		fullPath := filepath.Join(c.TargetDir, file.Name())
		ext := strings.ToLower(filepath.Ext(file.Name()))

		if c.SupportedImageExtensions[ext] {
			continue
		}

		if c.shouldProcess(ctx, fullPath) {
			jobs = append(jobs, job{fileName: file.Name(), fullPath: fullPath})
		}
	}

	return jobs
}

func (c *Converter) shouldProcess(ctx context.Context, fullPath string) bool {
	isImage, err := c.checkIsImage(ctx, fullPath)
	if err != nil {
		return false
	}

	return isImage
}

func (c *Converter) processFile(ctx context.Context, fileName, fullPath string) error {
	if err := c.convert(ctx, fileName, fullPath); err != nil {
		return fmt.Errorf("failed to convert image: %w", err)
	}

	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete original image file: %w", err)
	}

	return nil
}
