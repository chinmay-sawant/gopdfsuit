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
		return w
	},
}

// CompressBufPool recycles bytes.Buffer for compression output
var CompressBufPool = sync.Pool{
	New: func() any {
		buf := new(bytes.Buffer)
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

// PutCompressBuffer returns a compression buffer to the pool after resetting it.
func PutCompressBuffer(buf *bytes.Buffer) {
	if buf == nil {
		return
	}
	buf.Reset()
	CompressBufPool.Put(buf)
}

// WarmCompressionPools pre-fills zlib writer and buffer pools to reduce cold-start churn.
func WarmCompressionPools(n int) {
	if n < 1 {
		n = 1
	}
	for i := 0; i < n; i++ {
		buf := GetCompressBuffer()
		zw := GetZlibWriter(buf)
		_ = zw.Close()
		PutZlibWriter(zw)
		PutCompressBuffer(buf)
	}
}

const compressSampleBytes = 4096

// CompressContentStream zlib-compresses raw page bytes when smaller than raw.
// C5: if the first 4KB does not compress, skips the full pass (store-uncompressed).
func CompressContentStream(raw []byte) (compressed *bytes.Buffer, useFlate bool) {
	rawLen := len(raw)
	if rawLen == 0 {
		return nil, false
	}

	sampleLen := rawLen
	if sampleLen > compressSampleBytes {
		sampleLen = compressSampleBytes
	}
	sampleBuf := GetCompressBuffer()
	zw := GetZlibWriter(sampleBuf)
	if _, err := zw.Write(raw[:sampleLen]); err != nil {
		_ = zw.Close()
		PutZlibWriter(zw)
		PutCompressBuffer(sampleBuf)
		return nil, false
	}
	if err := zw.Close(); err != nil {
		PutZlibWriter(zw)
		PutCompressBuffer(sampleBuf)
		return nil, false
	}
	PutZlibWriter(zw)
	if sampleBuf.Len() >= sampleLen {
		PutCompressBuffer(sampleBuf)
		return nil, false
	}
	PutCompressBuffer(sampleBuf)

	compressedBuf := GetCompressBuffer()
	if grow := rawLen / 3; grow < 4096 {
		compressedBuf.Grow(4096)
	} else {
		compressedBuf.Grow(grow)
	}
	zw = GetZlibWriter(compressedBuf)
	if _, err := zw.Write(raw); err != nil {
		_ = zw.Close()
		PutZlibWriter(zw)
		PutCompressBuffer(compressedBuf)
		return nil, false
	}
	if err := zw.Close(); err != nil {
		PutZlibWriter(zw)
		PutCompressBuffer(compressedBuf)
		return nil, false
	}
	PutZlibWriter(zw)
	if compressedBuf.Len() >= rawLen {
		PutCompressBuffer(compressedBuf)
		return nil, false
	}
	return compressedBuf, true
}
