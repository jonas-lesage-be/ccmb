package video

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

// Extractor is responsible for extracting frames from video files.
type Extractor struct {
	TargetDir string

	MaxWorkers int
	Timeout    time.Duration

	FlatPathDelimiter string
	VideoExtensions   map[string]bool
}

type job struct {
	name string
	path string
}

// NewExtractor creates a new instance of Extractor using the application configuration.
func NewExtractor(cfg *config.Config) *Extractor {
	return &Extractor{
		TargetDir: cfg.TargetDir,

		MaxWorkers: cfg.MaxVideoWorkers,
		Timeout:    cfg.VideoExtractTimeout,

		FlatPathDelimiter: cfg.FlatPathDelimiter,
		VideoExtensions:   cfg.VideoExtensions,
	}
}

// Execute extracts 1 frame per second calls.
func (e *Extractor) Execute(ctx context.Context) error {
	slog.Debug("Extracting frames out of videos")

	files, err := os.ReadDir(e.TargetDir)
	if err != nil {
		return fmt.Errorf("failed to read target directory: %w", err)
	}

	jobs := e.jobsToProcess(ctx, files)
	if len(jobs) == 0 {
		return nil
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(e.MaxWorkers)

	for _, j := range jobs {
		g.Go(func() error {
			if err := e.processFile(ctx, j.name, j.path); err != nil {
				return fmt.Errorf("error processing %s: %w", j.name, err)
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("frame extraction failed: %w", err)
	}

	return nil
}

func (e *Extractor) jobsToProcess(ctx context.Context, files []os.DirEntry) []job {
	var jobs []job

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		var shouldProcess bool
		var path string

		if e.VideoExtensions != nil {
			ext := strings.ToLower(filepath.Ext(file.Name()))
			if shouldProcess = e.VideoExtensions[ext]; shouldProcess {
				path = filepath.Join(e.TargetDir, file.Name())
			}
		} else {
			path = filepath.Join(e.TargetDir, file.Name())
			shouldProcess = e.shouldProcess(ctx, path)
		}

		if shouldProcess {
			jobs = append(jobs, job{name: file.Name(), path: path})
		}
	}

	return jobs
}

func (e *Extractor) processFile(ctx context.Context, fileName, path string) error {
	if err := e.extractFrames(ctx, path); err != nil {
		return fmt.Errorf("failed to extract frames: %w", err)
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete video file %s: %w", fileName, err)
	}

	return nil
}
