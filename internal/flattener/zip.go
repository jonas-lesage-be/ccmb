package flattener

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	kzip "github.com/klauspost/compress/zip"

	"ccmb/internal/config"
)

func (f *Flattener) handleZIP(ctx context.Context, zipPath string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error before processing ZIP %s: %w", zipPath, err)
	}

	rc, err := kzip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip file %s: %w", zipPath, err)
	}
	defer func() {
		if err := rc.Close(); err != nil {
			slog.Error("failed to close zip reader", "err", err)
		}
	}()

	relZIP, err := filepath.Rel(f.SourceDir, zipPath)
	if err != nil {
		return fmt.Errorf("failed to get relative zip path: %w", err)
	}
	zipPrefix := config.EncodeFlatName(relZIP, f.FlatPathDelimiter, f.EscapedDelimiter)
	zipPrefix = strings.TrimSuffix(zipPrefix, filepath.Ext(zipPrefix))

	for _, file := range rc.File {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context error while processing ZIP %s: %w", zipPath, err)
		}

		if file.FileInfo().IsDir() || config.ShouldIgnore(file.Name) {
			continue
		}

		memberFlatName := config.EncodeFlatName(file.Name, f.FlatPathDelimiter, f.EscapedDelimiter)
		finalFlatName := fmt.Sprintf("%s%s%s", zipPrefix, f.FlatPathDelimiter, memberFlatName)
		targetPath := filepath.Join(f.TargetDir, finalFlatName)

		if err := f.extractZIPMember(ctx, file, targetPath); err != nil {
			slog.Error("failed to extract from zip", "member", file.Name, "err", err)
		}
	}

	return nil
}

func (f *Flattener) extractZIPMember(
	ctx context.Context,
	file *kzip.File,
	targetPath string,
) error {
	rc, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open zip member %s: %w", file.Name, err)
	}
	defer func() {
		if err := rc.Close(); err != nil {
			slog.Error("failed to close zip member stream", "err", err)
		}
	}()

	if err := f.extractFileStream(ctx, rc, targetPath, file.Mode(), file.Name); err != nil {
		if errors.Is(err, errFilteredExtension) {
			return nil
		}
		return fmt.Errorf("failed to extract zip member %s: %w", file.Name, err)
	}

	return nil
}
