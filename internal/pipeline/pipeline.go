package pipeline

import (
	"context"
	"log/slog"
	"time"

	"ccmb/internal/config"
	"ccmb/internal/filter"
	"ccmb/internal/flattener"
	"ccmb/internal/image"
	"ccmb/internal/textmerge"
	"ccmb/internal/video"
	"ccmb/internal/visualmerge"
)

type pipelineStep int

const (
	flattenerStep pipelineStep = iota
	filterStep
	imageConverterStep
	videoExtractorStep
	visualMergerStep
	textMergerStep
)

func (p pipelineStep) String() string {
	switch p {
	case flattenerStep:
		return "Flattener"
	case filterStep:
		return "Filter"
	case imageConverterStep:
		return "Image converter"
	case videoExtractorStep:
		return "Video extractor"
	case visualMergerStep:
		return "Visual merger"
	case textMergerStep:
		return "Text merger"
	default:
		return "Unknown step"
	}
}

// Execute runs the entire pipeline based on the provided configuration.
func Execute(cfg *config.Config) {
	ctx := context.Background()
	slog.Info("Pipeline initialized", "source", cfg.SourceDir, "target", cfg.TargetDir)

	// Execution step 1
	runIf(cfg.RunFlattener, flattenerStep, func() error {
		f := flattener.NewFlattener(cfg)
		return f.Execute(ctx)
	})

	// Execution step 2
	runIf(cfg.RunFilter, filterStep, func() error {
		f := filter.NewFilter(cfg)
		return f.Execute()
	})

	// Execution step 3
	runIf(cfg.RunImageConverter, imageConverterStep, func() error {
		c := image.NewConverter(cfg)
		return c.Execute(ctx)
	})

	// Execution step 4
	runIf(cfg.RunVideoExtractor, videoExtractorStep, func() error {
		e := video.NewExtractor(cfg)
		return e.Execute(ctx)
	})

	// Execution step 5
	runIf(cfg.RunVisualMerger, visualMergerStep, func() error {
		merger := visualmerge.NewMerger(cfg)
		return merger.Execute(ctx)
	})

	// Execution step 6
	runIf(cfg.RunTextMerger, textMergerStep, func() error {
		merger := textmerge.NewMerger(cfg)
		return merger.Execute(ctx)
	})

	slog.Info("Pipeline execution successfully completed")
}

func runIf(shouldRun bool, step pipelineStep, f func() error) {
	if !shouldRun {
		return
	}

	logger := slog.With("step", int(step+1), "name", step.String())
	logger.Info("Running step")

	startTime := time.Now()
	if err := f(); err != nil {
		logger.Error("Pipeline aborted", "err", err)
		return
	}
	duration := time.Since(startTime)

	logger.Info("Step completed successfully", "duration", duration)
}
