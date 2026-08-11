package image

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"golang.org/x/sync/errgroup"

	"ccmb/internal/config"
	"ccmb/internal/media"
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
	slog.Debug("Converting unsupported image files")

	files, err := os.ReadDir(c.TargetDir)
	if err != nil {
		return fmt.Errorf("failed to read target directory: %w", err)
	}

	jobs := c.filterFiles(files)
	if len(jobs) == 0 {
		return nil
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(c.MaxWorkers)

	for _, j := range jobs {
		g.Go(func() error {
			return c.processJob(ctx, j)
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("image conversion failed: %w", err)
	}

	return nil
}

func (c *Converter) filterFiles(files []os.DirEntry) []job {
	var jobs []job

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		ext := strings.ToLower(filepath.Ext(name))

		if c.SupportedImageExtensions[ext] {
			continue
		}

		if c.ImageExtensions != nil && !c.ImageExtensions[ext] {
			continue
		}

		jobs = append(jobs, job{
			name: name,
			path: filepath.Join(c.TargetDir, name),
		})
	}

	return jobs
}

func (c *Converter) processJob(ctx context.Context, j job) error {
	isSingleFrameImage, err := c.isSingleFrameImage(ctx, j.path)
	if err != nil {
		return fmt.Errorf("error classifying %s: %w", j.name, err)
	}

	if !isSingleFrameImage {
		return nil
	}

	if err := c.processFile(ctx, j.name, j.path); err != nil {
		return fmt.Errorf("error converting %s: %w", j.name, err)
	}

	return nil
}

func (c *Converter) isSingleFrameImage(ctx context.Context, path string) (bool, error) {
	mtype, err := mimetype.DetectFile(path)
	if err != nil {
		return false, fmt.Errorf("failed to detect content type: %w", err)
	}

	if media.MainType(mtype.String()) != "image" {
		return false, nil
	}

	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	frameType, err := media.ProbeFrameType(ctx, path)
	if err != nil {
		return false, fmt.Errorf("failed to check if single-frame media file: %w", err)
	}

	return frameType == media.SingleFrameType, nil
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
