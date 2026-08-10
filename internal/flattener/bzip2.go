package flattener

import (
	"compress/bzip2"
	"context"
	"io"
)

func (f *Flattener) handleTarBZ2(ctx context.Context, path string) error {
	return f.handleTAR(ctx, path, func(r io.Reader) (io.Reader, io.Closer, error) {
		bzr := bzip2.NewReader(r)
		return bzr, nil, nil
	})
}
