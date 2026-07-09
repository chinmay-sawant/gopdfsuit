//nolint:revive // package comment
package font

import (
	"bytes"
	"compress/zlib"
	"io"
	"sync"
)

// ZlibWriterPool recycles zlib writers to avoid allocation overhead
// Each zlib.NewWriter allocates ~256KB for compression tables
var ZlibWriterPool = sync.Pool{
	New: func() any {
		// Create writer that will be reset with actual buffer later
		w, _ := zlib.NewWriterLevel(io.Discard, zlib.BestSpeed)
		return w // *zlib.Writer (pointer) — Put does not box (PERF-110)
	},
}

// CompressBufPool recycles *bytes.Buffer for compression output.
// New always returns a pointer so Put never allocates a box (PERF-110).
var CompressBufPool = sync.Pool{
	New: func() any {
		buf := &bytes.Buffer{}
		buf.Grow(65536)
		return buf
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

// GetCompressBuffer returns a pooled compression buffer
func GetCompressBuffer() *bytes.Buffer {
	buf := CompressBufPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

// PutCompressBuffer returns a compression buffer to the pool.
// GetCompressBuffer resets on fetch, so we skip reset here.
func PutCompressBuffer(buf *bytes.Buffer) {
	if buf == nil {
		return
	}
	CompressBufPool.Put(buf)
}
