package flattener

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	"ccmb/internal/config"
)

// Flattener is responsible for flattening the directory structure of files.
type Flattener struct {
	saveMutex sync.Mutex

	MaxFlattenerWorkerCount int
	EstFileCount            int
	FlatPathDelimiter       string
	EscapedDelimiter        string

	MaxZIPFileBytes      int64
	SourceDir            string
	TargetDir            string
	TargetDirPermissions os.FileMode
}

// NewFlattener creates a new instance of Flattener using the application configuration.
func NewFlattener(cfg *config.Config) *Flattener {
	return &Flattener{
		MaxFlattenerWorkerCount: cfg.MaxFlattenerWorkerCount,
		EstFileCount:            cfg.EstFileCount,
		FlatPathDelimiter:       cfg.FlatPathDelimiter,
		EscapedDelimiter:        cfg.EscapedDelimiter,

		MaxZIPFileBytes:      cfg.MaxZIPFileBytes,
		SourceDir:            cfg.SourceDir,
		TargetDir:            cfg.TargetDir,
		TargetDirPermissions: cfg.TargetDirPermissions,
	}
}

// Execute scans the source directory and flattens files into the target directory.
func (f *Flattener) Execute(ctx context.Context) error {
	slog.Info("Starting directory flattening")

	if err := os.MkdirAll(f.TargetDir, f.TargetDirPermissions); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	absTarget, err := filepath.Abs(f.TargetDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute target path: %w", err)
	}

	filesToProcess, err := f.collectFiles(ctx, absTarget)
	if err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(f.MaxFlattenerWorkerCount)

	for _, path := range filesToProcess {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context error while processing file %s: %w", path, err)
		}

		g.Go(func() error {
			return f.processFile(ctx, path)
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("flattening failed: %w", err)
	}

	return nil
}

func (f *Flattener) collectFiles(ctx context.Context, absTarget string) ([]string, error) {
	filesToProcess := make([]string, 0, f.EstFileCount)

	err := filepath.WalkDir(f.SourceDir, func(srcPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %s: %w", srcPath, err)
		}

		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context error while walking source directory: %w", err)
		}

		if d.IsDir() || config.ShouldIgnore(srcPath) {
			return nil
		}

		absSrcPath, err := filepath.Abs(srcPath)
		if err != nil {
			return fmt.Errorf("failed to resolve absolute path for %s: %w", srcPath, err)
		}
		if strings.HasPrefix(absSrcPath, absTarget) {
			return nil
		}

		filesToProcess = append(filesToProcess, srcPath)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed walking source directory: %w", err)
	}

	return filesToProcess, nil
}

func (f *Flattener) processFile(ctx context.Context, path string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error while processing file %s: %w", path, err)
	}

	ext := f.extension(path)
	displayExt := strings.TrimPrefix(strings.ToUpper(ext), ".")
	handler := f.handlerForExtension(ext)

	if handler != nil {
		slog.Info("archive found, extracting", "type", displayExt, "file", filepath.Base(path))
		if err := handler(ctx, path); err != nil {
			return fmt.Errorf("error processing %s %s: %w", displayExt, path, err)
		}
		return nil
	}

	rel, err := filepath.Rel(f.SourceDir, path)
	if err != nil {
		return fmt.Errorf("failed to get relative path for %s: %w", path, err)
	}

	flatName := config.EncodeFlatName(rel, f.FlatPathDelimiter, f.EscapedDelimiter)
	targetPath := filepath.Join(f.TargetDir, flatName)

	if err := f.copyFileSecure(ctx, path, targetPath); err != nil {
		return fmt.Errorf("error copying %s: %w", path, err)
	}

	return nil
}

func (f *Flattener) copyFileSecure(ctx context.Context, src, dst string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error before copying file %s: %w", src, err)
	}

	f.saveMutex.Lock()
	dst = f.resolveCollision(dst)
	out, err := os.Create(filepath.Clean(dst))
	f.saveMutex.Unlock()

	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", dst, err)
	}

	shouldCleanup := true
	defer func() {
		if errClose := out.Close(); errClose != nil {
			slog.Error("failed to close written file", "error", errClose)
		}

		if shouldCleanup {
			if errRemove := os.Remove(dst); errRemove != nil {
				slog.Error("failed to remove incomplete file", "path", dst, "error", errRemove)
			}
		}
	}()

	in, err := os.Open(filepath.Clean(src))
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", src, err)
	}
	defer func() {
		if errClose := in.Close(); errClose != nil {
			slog.Error("failed to close source stream", "error", errClose)
		}
	}()

	if _, err = io.Copy(out, in); err != nil {
		return fmt.Errorf("failed copying content from %s to %s: %w", src, dst, err)
	}

	shouldCleanup = false
	return nil
}

func (f *Flattener) resolveCollision(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}

	ext := f.extension(path)
	base := strings.TrimSuffix(path, ext)

	for counter := 1; ; counter++ {
		newPath := fmt.Sprintf("%s_%d%s", base, counter, ext)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}
}
