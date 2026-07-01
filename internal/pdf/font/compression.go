//nolint:revive // package comment
package font

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"sync"
)

// ZlibWriterPool recycles zlib writers to avoid allocation overhead
// Each zlib.NewWriter allocates ~256KB for compression tables
var ZlibWriterPool = sync.Pool{
	New: func() any {
		// Create writer that will be reset with actual buffer later.
		// NewWriterLevel can fail only for invalid levels; fall back to default.
		w, err := zlib.NewWriterLevel(io.Discard, zlib.BestSpeed)
		if err != nil {
			return zlib.NewWriter(io.Discard)
		}
		return w
	},
}

// CompressBufPool recycles bytes.Buffer for compression output
var CompressBufPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

// GetZlibWriter returns a pooled zlib writer reset to write to the given buffer
func GetZlibWriter(buf *bytes.Buffer) *zlib.Writer {
	w := ZlibWriterPool.Get().(*zlib.Writer)
	w.Reset(buf)
	return w
}

// PutZlibWriter returns a zlib writer to the pool
func PutZlibWriter(w *zlib.Writer) {
	ZlibWriterPool.Put(w)
}

// CloseZlibWriter closes the writer and returns it to the pool.
func CloseZlibWriter(w *zlib.Writer) error {
	if err := w.Close(); err != nil {
		PutZlibWriter(w)
		return fmt.Errorf("close zlib writer: %w", err)
	}
	PutZlibWriter(w)
	return nil
}

// GetCompressBuffer returns a pooled compression buffer
func GetCompressBuffer() *bytes.Buffer {
	buf := CompressBufPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}
