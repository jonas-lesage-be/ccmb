package filter

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"ccmb/internal/config"
	"ccmb/internal/flattener"
)

// Filter is responsible for filtering out files with unsupported extensions.
type Filter struct {
	TargetDir        string
	FilterExtensions map[string]bool
}

// NewFilter creates a new instance of Filter using the application configuration.
func NewFilter(cfg *config.Config) *Filter {
	return &Filter{
		TargetDir:        cfg.TargetDir,
		FilterExtensions: cfg.FilterExtensions,
	}
}

// Execute looks up all files in TargetDir and deletes unsupported files.
func (f *Filter) Execute() error {
	slog.Debug("Removing files with an unsupported extension")

	files, err := os.ReadDir(f.TargetDir)
	if err != nil {
		return fmt.Errorf("failed to read target directory %s: %w", f.TargetDir, err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if ext := flattener.Extension(file.Name()); f.FilterExtensions[ext] {
			path := filepath.Join(f.TargetDir, file.Name())
			if err := os.Remove(path); err != nil {
				slog.Error("Failed to delete file", "filename", file.Name(), "err", err)
			} else {
				slog.Debug("Deleted unsupported file", "filename", file.Name())
			}
		}
	}

	return nil
}
