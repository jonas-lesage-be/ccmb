package filter

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"ccmb/internal/config"
)

// Filter is responsible for filtering out files with unsupported extensions.
type Filter struct {
	FilterExtensions map[string]bool
	TargetDir        string
}

// NewFilter creates a new instance of Filter using the application configuration.
func NewFilter(cfg *config.Config) *Filter {
	return &Filter{
		FilterExtensions: cfg.FilterExtensions,
		TargetDir:        cfg.TargetDir,
	}
}

// Execute looks up all files in TargetDir and deletes unsupported files.
func (f *Filter) Execute() error {
	log.Println("--- Removing files with an unsupported extension ---")

	files, err := os.ReadDir(f.TargetDir)
	if err != nil {
		return fmt.Errorf("failed to read target directory %s: %w", f.TargetDir, err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if ext := strings.ToLower(filepath.Ext(file.Name())); f.FilterExtensions[ext] {
			fullPath := filepath.Join(f.TargetDir, file.Name())
			if err := os.Remove(fullPath); err != nil {
				log.Printf("Failed to delete %s: %v", file.Name(), err)
			} else {
				log.Printf("Deleted unsupported file: %s", file.Name())
			}
		}
	}

	return nil
}
