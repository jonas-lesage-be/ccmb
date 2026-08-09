package video

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"ccmb/internal/config"
)

// Extractor is responsible for extracting frames from video files.
type Extractor struct {
	MaxVideoWorkerCount int
	FlatPathDelimiter   string
	VideoExtractTimeout time.Duration
	VideoExtensions     map[string]bool
	TargetDir           string
}

type job struct {
	fileName string
	fullPath string
}

// NewExtractor creates a new instance of Extractor using the application configuration.
func NewExtractor(cfg *config.Config) *Extractor {
	return &Extractor{
		MaxVideoWorkerCount: cfg.MaxVideoWorkerCount,
		FlatPathDelimiter:   cfg.FlatPathDelimiter,
		VideoExtractTimeout: cfg.VideoExtractTimeout,
		VideoExtensions:     cfg.VideoExtensions,
		TargetDir:           cfg.TargetDir,
	}
}

// Execute extracts 1 frame per second calls.
func (e *Extractor) Execute(ctx context.Context) error {
	log.Println("--- Extracting frames out of videos ---")
	files, err := os.ReadDir(e.TargetDir)
	if err != nil {
		return fmt.Errorf("failed to read target directory: %w", err)
	}

	jobs := e.jobsToProcess(ctx, files)
	if len(jobs) == 0 {
		return nil
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(e.MaxVideoWorkerCount)

	for _, j := range jobs {
		g.Go(func() error {
			if err := e.processFile(ctx, j.fileName, j.fullPath); err != nil {
				return fmt.Errorf("error processing %s: %w", j.fileName, err)
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
		var fullPath string

		if e.VideoExtensions != nil {
			ext := strings.ToLower(filepath.Ext(file.Name()))
			if shouldProcess = e.VideoExtensions[ext]; shouldProcess {
				fullPath = filepath.Join(e.TargetDir, file.Name())
			}
		} else {
			fullPath = filepath.Join(e.TargetDir, file.Name())
			shouldProcess = e.shouldProcess(ctx, fullPath)
		}

		if shouldProcess {
			jobs = append(jobs, job{fileName: file.Name(), fullPath: fullPath})
		}
	}

	return jobs
}

func (e *Extractor) processFile(ctx context.Context, fileName, fullPath string) error {
	if err := e.extractFrames(ctx, fullPath); err != nil {
		return fmt.Errorf("failed to extract frames: %w", err)
	}

	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete video file %s: %w", fileName, err)
	}

	return nil
}
