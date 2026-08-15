package flattener

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sync/errgroup"

	"ccmb/internal/config"
)

var errFilteredExtension = errors.New("file extension is blocked by filter configuration")

// Flattener is responsible for flattening the directory structure of files.
type Flattener struct {
	SourceDir            string
	TargetDir            string
	TargetDirPermissions os.FileMode

	MaxFlattenerWorkers int
	EstFileCount        int

	FlatPathDelimiter string
	EscapedDelimiter  string

	MaxArchiveFileBytes      int64
	FilterExtensions         map[string]bool
	SkipTARFlattener         bool
	SkipTARFlattenerExplicit bool
}

// NewFlattener creates a new instance of Flattener using the application configuration.
func NewFlattener(cfg *config.Config) *Flattener {
	return &Flattener{
		SourceDir:            cfg.SourceDir,
		TargetDir:            cfg.TargetDir,
		TargetDirPermissions: cfg.TargetDirPermissions,

		MaxFlattenerWorkers: cfg.MaxFlattenerWorkers,
		EstFileCount:        cfg.EstFileCount,

		FlatPathDelimiter: cfg.FlatPathDelimiter,
		EscapedDelimiter:  cfg.EscapedDelimiter,

		MaxArchiveFileBytes:      cfg.MaxArchiveFileBytes,
		FilterExtensions:         cfg.FilterExtensions,
		SkipTARFlattener:         cfg.SkipTARFlattener,
		SkipTARFlattenerExplicit: cfg.SkipTARFlattenerExplicit,
	}
}

// Execute scans the source directory and flattens files into the target directory.
func (f *Flattener) Execute(ctx context.Context) error {
	slog.Debug("Starting directory flattening")

	if f.SkipTARFlattener && !f.SkipTARFlattenerExplicit {
		slog.Info(
			"TAR archive flattening disabled",
			"reason",
			"unreliable on this platform, override with --skip-tar-flattener=false",
		)
	}

	if err := os.MkdirAll(f.TargetDir, f.TargetDirPermissions); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	absTarget, err := filepath.Abs(f.TargetDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute target path: %w", err)
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(f.MaxFlattenerWorkers)

	paths := make(chan string, f.MaxFlattenerWorkers)

	g.Go(func() error {
		defer close(paths)
		return f.walkFiles(ctx, absTarget, paths)
	})

	for range f.MaxFlattenerWorkers {
		g.Go(func() error {
			for path := range paths {
				if err := f.processFile(ctx, path); err != nil {
					return err
				}
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("flattening failed: %w", err)
	}

	return nil
}

func (f *Flattener) walkFiles(ctx context.Context, absTarget string, paths chan<- string) error {
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

		select {
		case paths <- srcPath:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
	if err != nil {
		return fmt.Errorf("failed walking source directory: %w", err)
	}

	return nil
}

func (f *Flattener) processFile(ctx context.Context, path string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error while processing file %s: %w", path, err)
	}

	ext := Extension(path)
	if f.FilterExtensions[ext] {
		slog.Debug("Skipping filtered file", "file", filepath.Base(path), "extension", ext)
		return nil
	}

	displayExt := strings.TrimPrefix(strings.ToUpper(ext), ".")
	handler := f.handlerForExtension(ext)

	if handler != nil {
		slog.Debug("Archive found, extracting", "type", displayExt, "file", filepath.Base(path))
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
		if errors.Is(err, errFilteredExtension) {
			return nil
		}
		return fmt.Errorf("error copying %s: %w", path, err)
	}

	return nil
}

func (f *Flattener) copyFileSecure(ctx context.Context, src, dst string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error before copying file %s: %w", src, err)
	}

	out, resolvedDst, err := f.createUnique(dst, f.TargetDirPermissions)
	if err != nil {
		if errors.Is(err, errFilteredExtension) {
			return errFilteredExtension
		}
		return fmt.Errorf("failed to create file %s: %w", dst, err)
	}

	shouldCleanup := true
	defer func() {
		if err := out.Close(); err != nil {
			slog.Error("failed to close written file", "err", err)
		}

		if shouldCleanup {
			if err := os.Remove(resolvedDst); err != nil {
				slog.Error("failed to remove incomplete file", "path", resolvedDst, "err", err)
			}
		}
	}()

	in, err := os.Open(filepath.Clean(src))
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", src, err)
	}
	defer func() {
		if err := in.Close(); err != nil {
			slog.Error("failed to close source stream", "err", err)
		}
	}()

	if _, err = io.Copy(out, in); err != nil {
		return fmt.Errorf("failed copying content from %s to %s: %w", src, resolvedDst, err)
	}

	shouldCleanup = false
	return nil
}

func (f *Flattener) createUnique(path string, mode os.FileMode) (*os.File, string, error) {
	ext := Extension(path)

	if f.FilterExtensions[ext] {
		slog.Debug(
			"Intercepting and dropping blocked extension entry at creation point",
			"extension",
			ext,
			"file",
			filepath.Base(path),
		)
		return nil, "", errFilteredExtension
	}

	base := strings.TrimSuffix(path, ext)
	candidate := path
	for counter := 1; ; counter++ {
		cleanCandidate := filepath.Clean(candidate)
		out, err := os.OpenFile(cleanCandidate, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
		if err == nil {
			return out, cleanCandidate, nil
		}

		if !os.IsExist(err) {
			return nil, "", fmt.Errorf("failed to create %s: %w", candidate, err)
		}

		candidate = fmt.Sprintf("%s_%d%s", base, counter, ext)
	}
}
