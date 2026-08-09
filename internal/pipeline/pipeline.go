package pipeline

import (
	"context"
	"log"

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
	log.Printf("Pipeline initialized. Source: %s | Target: %s\n", cfg.SourceDir, cfg.TargetDir)

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

	log.Println("Pipeline execution successfully completed!")
}

func runIf(shouldRun bool, step pipelineStep, f func() error) {
	if !shouldRun {
		return
	}

	log.Printf("=== Running STEP %d: %s ===", step+1, step)

	if err := f(); err != nil {
		log.Fatalf("Pipeline aborted at step %d (%s): %v", step+1, step, err)
		return
	}

	log.Printf("=== STEP %d: %s completed successfully ===", step+1, step)
}
