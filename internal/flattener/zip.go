package flattener

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	kzip "github.com/klauspost/compress/zip"

	"ccmb/internal/config"
)

type zipExtractionError struct {
	MemberName string
	Message    string
}

func (e *zipExtractionError) Error() string {
	return fmt.Sprintf("error extracting zip member %s: %s", e.MemberName, e.Message)
}

func (pf *Flattener) handleZIP(ctx context.Context, zipPath string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error before processing ZIP %s: %w", zipPath, err)
	}

	r, err := kzip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip file %s: %w", zipPath, err)
	}
	defer func() {
		if errClose := r.Close(); errClose != nil {
			log.Printf("failed to close zip reader: %v", errClose)
		}
	}()

	relZIP, err := filepath.Rel(pf.SourceDir, zipPath)
	if err != nil {
		return fmt.Errorf("failed to get relative zip path: %w", err)
	}
	zipPrefix := config.EncodeFlatName(relZIP, pf.FlatPathDelimiter, pf.EscapedDelimiter)
	zipPrefix = strings.TrimSuffix(zipPrefix, filepath.Ext(zipPrefix))

	for _, f := range r.File {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context error while processing ZIP %s: %w", zipPath, err)
		}

		if f.FileInfo().IsDir() || config.ShouldIgnore(f.Name) {
			continue
		}

		memberFlatName := config.EncodeFlatName(f.Name, pf.FlatPathDelimiter, pf.EscapedDelimiter)
		finalFlatName := fmt.Sprintf("%s--%s", zipPrefix, memberFlatName)
		targetPath := filepath.Join(pf.TargetDir, finalFlatName)

		if err := pf.extractZIPMember(ctx, f, targetPath); err != nil {
			log.Printf("Failed to extract %s from zip: %v", f.Name, err)
		}
	}

	return nil
}

func (pf *Flattener) extractZIPMember(ctx context.Context, f *kzip.File, targetPath string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error before extracting zip member %s: %w", f.Name, err)
	}

	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("failed to open zip member %s: %w", f.Name, err)
	}
	defer func() {
		if errClose := rc.Close(); errClose != nil {
			log.Printf("failed to close zip member stream: %v", errClose)
		}
	}()

	pf.saveMutex.Lock()
	targetPath = pf.resolveCollision(targetPath)
	cleanedPath := filepath.Clean(targetPath)
	out, err := os.OpenFile(
		cleanedPath,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		f.Mode(),
	)
	pf.saveMutex.Unlock()

	if err != nil {
		return fmt.Errorf("failed to create target file %s: %w", targetPath, err)
	}

	shouldCleanup := true
	defer func() {
		if errClose := out.Close(); errClose != nil {
			log.Printf("failed to safely close output file: %v", errClose)
		}
		if shouldCleanup {
			if errRemove := os.Remove(cleanedPath); errRemove != nil {
				log.Printf("failed to remove incomplete file %s: %v", cleanedPath, errRemove)
			}
		}
	}()

	limitedReader := io.LimitReader(rc, pf.MaxZIPFileBytes+1)
	written, err := io.Copy(out, limitedReader)
	if err != nil {
		return &zipExtractionError{
			MemberName: f.Name,
			Message:    err.Error(),
		}
	}

	if written > pf.MaxZIPFileBytes {
		return &zipExtractionError{
			MemberName: f.Name,
			Message:    "size exceeds maximum allowed limit",
		}
	}

	shouldCleanup = false
	return nil
}
