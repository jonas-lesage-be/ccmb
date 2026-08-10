package flattener

import (
	"context"
	"fmt"
	"io"

	"github.com/klauspost/compress/zstd"
)

func (f *Flattener) handleTarZST(ctx context.Context, path string) error {
	return f.handleTAR(ctx, path, func(r io.Reader) (io.Reader, io.Closer, error) {
		zsr, err := zstd.NewReader(r)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create zstd reader for %s: %w", path, err)
		}

		closerAdapter := io.Closer(funcCloser(func() error {
			zsr.Close()
			return nil
		}))
		return zsr, closerAdapter, nil
	})
}
