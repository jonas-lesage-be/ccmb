package flattener

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/klauspost/compress/gzip"
)

func (f *Flattener) handleTarGZ(ctx context.Context, tarGzPath string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error before processing tar.gz %s: %w", tarGzPath, err)
	}

	fi, err := os.Open(filepath.Clean(tarGzPath))
	if err != nil {
		return fmt.Errorf("failed to open tar.gz file %s: %w", tarGzPath, err)
	}
	defer func() {
		if errClose := fi.Close(); errClose != nil {
			log.Printf("failed to close tar.gz file stream: %v", errClose)
		}
	}()

	gzr, err := gzip.NewReader(fi)
	if err != nil {
		return fmt.Errorf("failed to initialize gzip reader for %s: %w", tarGzPath, err)
	}
	defer func() {
		if errClose := gzr.Close(); errClose != nil {
			log.Printf("failed to close gzip reader: %v", errClose)
		}
	}()

	return f.handleTAR(ctx, gzr, tarGzPath)
}
