package pipeline

import (
	"context"
	"fmt"
	"log/slog"
	"os"
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

	if err := os.MkdirAll(p.cfg.TargetDir, p.cfg.TargetDirPermissions); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "ccmb_tmpdir_*")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer func() {
		if cleanErr := os.RemoveAll(tmpDir); cleanErr != nil {
			slog.Error("Failed to purge temporary directory", "err", cleanErr)
		}
	}()

	tmpCfg := *p.cfg
	tmpCfg.TargetDir = tmpDir

	for _, step := range p.steps {
		if step.ShouldSkip(p.cfg) {
			continue
		}

		if err := p.executeStep(ctx, step, &tmpCfg, tmpDir); err != nil {
			return err
		}
	}

	slog.Info("Pipeline execution successfully completed", "duration", time.Since(startTime))
	return nil
}

func (p *Pipeline) executeStep(
	ctx context.Context,
	step pipelineStep,
	tmpCfg *config.Config,
	tmpDir string,
) error {
	logger := slog.With("step", int(step+1), "name", step.String())
	logger.Info("Running step")
	startTime := time.Now()

	var err error
	switch step {
	case flattenerStep:
		f := flattener.NewFlattener(tmpCfg)
		err = f.Execute(ctx)
	case filterStep:
		f := filter.NewFilter(tmpCfg)
		err = f.Execute()
	case documentConverterStep:
		c := document.NewConverter(tmpCfg)
		err = c.Execute(ctx)
	case imageConverterStep:
		c := image.NewConverter(tmpCfg)
		err = c.Execute(ctx)
	case videoExtractorStep:
		e := video.NewExtractor(tmpCfg)
		err = e.Execute(ctx)
	case visualMergerStep:
		m := visualmerge.NewMerger(p.cfg)
		err = m.Execute(ctx, tmpDir)
	case textMergerStep:
		m := textmerge.NewMerger(p.cfg)
		err = m.Execute(ctx, tmpDir)
	}

	if err != nil {
		return fmt.Errorf("pipeline aborted at step %s: %w", step.String(), err)
	}

	logger.Info("Step completed successfully", "duration", time.Since(startTime))

	return nil
}
