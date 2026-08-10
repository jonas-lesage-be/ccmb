package flattener

import (
	"context"
	"fmt"
	"io"

	"github.com/klauspost/compress/gzip"
)

func (f *Flattener) handleTarGZ(ctx context.Context, path string) error {
	return f.handleTAR(ctx, path, func(r io.Reader) (io.Reader, io.Closer, error) {
		gzr, err := gzip.NewReader(r)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create gzip reader for %s: %w", path, err)
		}
		return gzr, gzr, nil
	})
}
