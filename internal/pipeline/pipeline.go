package pipeline

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"ccmb/internal/config"
	"ccmb/internal/document"
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
	documentConverterStep
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
	case documentConverterStep:
		return "Document converter"
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

func (p pipelineStep) ShouldSkip(cfg *config.Config) bool {
	switch p {
	case flattenerStep:
		return cfg.SkipFlattener
	case filterStep:
		return cfg.SkipFilter
	case documentConverterStep:
		return cfg.SkipDocumentConverter
	case imageConverterStep:
		return cfg.SkipImageConverter
	case videoExtractorStep:
		return cfg.SkipVideoExtractor
	case visualMergerStep:
		return cfg.SkipVisualMerger
	case textMergerStep:
		return cfg.SkipTextMerger
	default:
		return true
	}
}

func (p pipelineStep) Func(cfg *config.Config) func(context.Context) error {
	switch p {
	case flattenerStep:
		return func(ctx context.Context) error {
			f := flattener.NewFlattener(cfg)
			return f.Execute(ctx)
		}
	case filterStep:
		return func(_ context.Context) error {
			f := filter.NewFilter(cfg)
			return f.Execute()
		}
	case documentConverterStep:
		return func(ctx context.Context) error {
			c := document.NewConverter(cfg)
			return c.Execute(ctx)
		}
	case imageConverterStep:
		return func(ctx context.Context) error {
			c := image.NewConverter(cfg)
			return c.Execute(ctx)
		}
	case videoExtractorStep:
		return func(ctx context.Context) error {
			e := video.NewExtractor(cfg)
			return e.Execute(ctx)
		}
	case visualMergerStep:
		return func(ctx context.Context) error {
			m := visualmerge.NewMerger(cfg)
			return m.Execute(ctx)
		}
	case textMergerStep:
		return func(ctx context.Context) error {
			m := textmerge.NewMerger(cfg)
			return m.Execute(ctx)
		}
	default:
		return nil
	}
}

// Pipeline represents the sequential processing pipeline.
type Pipeline struct {
	cfg   *config.Config
	steps []pipelineStep
}

// NewPipeline creates a new instance of Pipeline using the application configuration.
func NewPipeline(cfg *config.Config) *Pipeline {
	return &Pipeline{
		cfg: cfg,
		steps: []pipelineStep{
			flattenerStep,
			filterStep,
			documentConverterStep,
			imageConverterStep,
			videoExtractorStep,
			visualMergerStep,
			textMergerStep,
		},
	}
}

// Execute runs the entire pipeline based on the provided configuration.
func (p *Pipeline) Execute(ctx context.Context) error {
	slog.Info("Pipeline initialized", "source", p.cfg.SourceDir, "target", p.cfg.TargetDir)

	startTime := time.Now()
	for _, step := range p.steps {
		if err := p.runIf(ctx, step); err != nil {
			return fmt.Errorf("pipeline aborted at step %s: %w", step.String(), err)
		}
	}
	duration := time.Since(startTime)

	slog.Info("Pipeline execution successfully completed", "duration", duration)

	return nil
}

func (p *Pipeline) runIf(ctx context.Context, step pipelineStep) error {
	if step.ShouldSkip(p.cfg) {
		return nil
	}

	logger := slog.With("step", int(step+1), "name", step.String())
	logger.Info("Running step")

	startTime := time.Now()
	if err := step.Func(p.cfg)(ctx); err != nil {
		return fmt.Errorf("pipeline aborted: %w", err)
	}
	duration := time.Since(startTime)

	logger.Info("Step completed successfully", "duration", duration)

	return nil
}
