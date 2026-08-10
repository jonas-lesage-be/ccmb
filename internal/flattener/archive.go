package flattener

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

var errSizeExceedsLimit = errors.New("size exceeds maximum allowed limit")

func (f *Flattener) handlerForExtension(ext string) func(context.Context, string) error {
	switch ext {
	case ".zip":
		return f.handleZIP
	case ".tar.gz", ".tgz":
		return f.handleTarGZ
	// case ".tar.xz", ".txz":
	// 	return f.handleTarXZ
	case ".tar.zst", ".tzst":
		return f.handleTarZST
	case ".tar.bz2", ".tbz2":
		return f.handleTarBZ2
	case ".tar":
		return f.handlePlainTAR
	default:
		return nil
	}
}

func (f *Flattener) extractFileStream(
	ctx context.Context,
	src io.Reader,
	targetPath string,
	mode os.FileMode,
	memberName string,
) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error before extracting member %s: %w", memberName, err)
	}

	f.saveMutex.Lock()
	targetPath = f.resolveCollision(targetPath)
	cleanedPath := filepath.Clean(targetPath)

	out, err := os.OpenFile(
		cleanedPath,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		mode,
	)
	f.saveMutex.Unlock()

	if err != nil {
		return fmt.Errorf("failed to create target file %s: %w", targetPath, err)
	}

	shouldCleanup := true
	defer func() {
		if errClose := out.Close(); errClose != nil {
			slog.Error("failed to safely close output file", "error", errClose)
		}
		if shouldCleanup {
			if errRemove := os.Remove(cleanedPath); errRemove != nil {
				slog.Error(
					"failed to remove incomplete file",
					"path",
					cleanedPath,
					"error",
					errRemove,
				)
			}
		}
	}()

	limitedReader := io.LimitReader(src, f.MaxArchiveFileBytes+1)
	written, err := io.Copy(out, limitedReader)
	if err != nil {
		return fmt.Errorf("failed during copy: %w", err)
	}

	if written > f.MaxArchiveFileBytes {
		return fmt.Errorf("%w (%d bytes)", errSizeExceedsLimit, f.MaxArchiveFileBytes)
	}

	shouldCleanup = false
	return nil
}
