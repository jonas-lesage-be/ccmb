package pathsafe

import (
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
)

// Contains checks if the targetPath is contained within the baseDir.
// It resolves both paths to their absolute forms and ensures that the targetPath
// starts with the baseDir followed by a path separator.
// This prevents directory traversal attacks and ensures that
// the targetPath is a subdirectory or file within the baseDir.
func Contains(baseDir, targetPath string) bool {
	if filepath.IsAbs(targetPath) {
		rel, err := filepath.Rel(baseDir, targetPath)
		if err != nil {
			return false
		}
		targetPath = rel
	}

	root, err := os.OpenRoot(baseDir)
	if err != nil {
		return false
	}
	defer func() {
		if cErr := root.Close(); cErr != nil {
			slog.Error("failed to close root directory", "error", cErr)
		}
	}()

	_, err = root.Stat(targetPath)
	if err != nil {
		return errors.Is(err, fs.ErrNotExist)
	}

	return true
}
