package video

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

// Extractor is responsible for extracting frames from video files.
type Extractor struct {
	Dir string

	MaxWorkers int
	Timeout    time.Duration
	Fps        float64

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
		Dir: cfg.TargetDir,

		MaxWorkers: cfg.MaxVideoWorkers,
		Timeout:    cfg.VideoExtractTimeout,
		Fps:        cfg.VideoFps,

		FlatPathDelimiter: cfg.FlatPathDelimiter,
		VideoExtensions:   cfg.VideoExtensions,
	}
}

// Execute extracts 1 frame per second calls.
func (e *Extractor) Execute(ctx context.Context) error {
	slog.Debug("Extracting frames out of videos")

	files, err := os.ReadDir(e.Dir)
	if err != nil {
		return fmt.Errorf("failed to read target directory: %w", err)
	}

	jobs := e.filterFiles(files)
	if len(jobs) == 0 {
		return nil
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(e.MaxWorkers)

	for _, j := range jobs {
		g.Go(func() error {
			return e.processJob(ctx, j)
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("frame extraction failed: %w", err)
	}

	return nil
}

func (e *Extractor) filterFiles(files []os.DirEntry) []job {
	var jobs []job

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		ext := strings.ToLower(filepath.Ext(name))

		if e.VideoExtensions != nil && !e.VideoExtensions[ext] {
			continue
		}

		jobs = append(jobs, job{
			name: name,
			path: filepath.Join(e.Dir, name),
		})
	}

	return jobs
}

func (e *Extractor) processJob(ctx context.Context, j job) error {
	shouldExtract, err := e.shouldExtract(ctx, j.path)
	if err != nil {
		return fmt.Errorf("error classifying %s: %w", j.name, err)
	}

	if !shouldExtract {
		slog.Debug("Skipping non-video, single-frame file", "file", j.name)
		return nil
	}

	if err := e.processFile(ctx, j.name, j.path); err != nil {
		return fmt.Errorf("error processing %s: %w", j.name, err)
	}

	return nil
}

func (e *Extractor) shouldExtract(ctx context.Context, path string) (bool, error) {
	mtype, err := mimetype.DetectFile(path)
	if err != nil {
		return false, fmt.Errorf("failed to detect content type: %w", err)
	}

	t := media.MainType(mtype.String())
	if t == "video" {
		return true, nil
	}
	if t != "image" {
		return false, nil
	}

	frameType, err := media.ProbeFrameType(ctx, path, e.Timeout)
	if err != nil {
		return false, fmt.Errorf("failed to probe frame type: %w", err)
	}

	return frameType == media.MultiFrameType, nil
}

func (e *Extractor) processFile(ctx context.Context, name, path string) error {
	if err := e.extractFrames(ctx, path); err != nil {
		return fmt.Errorf("failed to extract frames: %w", err)
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete video file %s: %w", name, err)
	}

	return nil
}
