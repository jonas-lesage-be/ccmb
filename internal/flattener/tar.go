package flattener

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"ccmb/internal/config"
)

type (
	wrapReaderFunc func(io.Reader) (io.Reader, io.Closer, error)
	funcCloser     func() error
)

func (f funcCloser) Close() error {
	return f()
}

func (f *Flattener) extension(path string) string {
	lowerPath := strings.ToLower(path)

	switch {
	case strings.HasSuffix(lowerPath, ".tar.gz"):
		return ".tar.gz"
	case strings.HasSuffix(lowerPath, ".tar.xz"):
		return ".tar.xz"
	case strings.HasSuffix(lowerPath, ".tar.zst"):
		return ".tar.zst"
	case strings.HasSuffix(lowerPath, ".tar.bz2"):
		return ".tar.bz2"
	}

	return filepath.Ext(lowerPath)
}

func (f *Flattener) handlePlainTAR(ctx context.Context, path string) error {
	return f.handleTAR(ctx, path, func(r io.Reader) (io.Reader, io.Closer, error) {
		return r, nil, nil
	})
}

func (f *Flattener) handleTAR(ctx context.Context, path string, wrapReader wrapReaderFunc) error {
	ext := f.extension(path)

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error before processing %s %s: %w", ext, path, err)
	}

	fi, err := os.Open(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("failed to open %s file %s: %w", ext, path, err)
	}
	defer func() {
		if errClose := fi.Close(); errClose != nil {
			log.Printf("failed to close %s file stream: %v", ext, errClose)
		}
	}()

	decompressor, closer, err := wrapReader(fi)
	if err != nil {
		return err
	}
	if closer != nil {
		defer func() {
			if errClose := closer.Close(); errClose != nil {
				log.Printf("failed to close %s reader: %v", ext, errClose)
			}
		}()
	}

	relPath, err := filepath.Rel(f.SourceDir, path)
	if err != nil {
		return fmt.Errorf("failed to get relative %s path: %w", ext, err)
	}

	prefix := config.EncodeFlatName(relPath, f.FlatPathDelimiter, f.EscapedDelimiter)
	prefix = strings.TrimSuffix(prefix, ext)

	return f.processTARStream(ctx, decompressor, path, prefix, ext)
}

func (f *Flattener) processTARStream(
	ctx context.Context,
	r io.Reader,
	path, prefix, ext string,
) error {
	tr := tar.NewReader(r)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading tar header from %s: %w", path, err)
		}

		if err := f.processTAREntry(ctx, tr, header, prefix, ext); err != nil {
			log.Printf("Failed to extract %s from %s: %v", header.Name, filepath.Base(path), err)
		}
	}

	return nil
}

func (f *Flattener) processTAREntry(
	ctx context.Context,
	tr *tar.Reader,
	header *tar.Header,
	prefix string,
	ext string,
) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error while processing %s %s: %w", ext, header.Name, err)
	}

	if header.Typeflag != tar.TypeReg || config.ShouldIgnore(header.Name) {
		return nil
	}

	memberFlatName := config.EncodeFlatName(
		header.Name,
		f.FlatPathDelimiter,
		f.EscapedDelimiter,
	)
	finalFlatName := fmt.Sprintf("%s%s%s", prefix, f.FlatPathDelimiter, memberFlatName)
	targetPath := filepath.Join(f.TargetDir, finalFlatName)

	return f.extractTARMember(ctx, tr, header, targetPath)
}

func (f *Flattener) extractTARMember(
	ctx context.Context,
	tr *tar.Reader,
	header *tar.Header,
	targetPath string,
) error {
	if err := f.extractFileStream(
		ctx,
		tr,
		targetPath,
		header.FileInfo().Mode(),
		header.Name,
	); err != nil {
		return fmt.Errorf("failed to extract tar member %s: %w", header.Name, err)
	}

	return nil
}
