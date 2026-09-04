package font

import (
	"bytes"
	"compress/zlib"
	"io"
	"testing"
)

func gunzip(t *testing.T, buf *bytes.Buffer) []byte {
	t.Helper()
	r, err := zlib.NewReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("zlib.NewReader: %v", err)
	}
	defer func() { _ = r.Close() }()
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read compressed: %v", err)
	}
	return out
}

// TestCompressCacheSamplingCollision keeps two pages distinct that would have
// collided under the old (len/first8/mid8/last8) fingerprint: same length,
// same first/middle/last 8 bytes, different body.
func TestCompressCacheSamplingCollision(t *testing.T) {
	ClearPageCompressCache()
	edge := []byte("ABCDEFGH")
	mk := func(fill byte) []byte {
		// [edge][fill*352][edge][fill*344][edge], len 720.
		// The old sampler read raw[:8], raw[360:368], raw[712:720]:
		// all three windows are edge in both fixtures.
		out := make([]byte, 0, 720)
		out = append(out, edge...)
		out = append(out, bytes.Repeat([]byte{fill}, 352)...)
		out = append(out, edge...)
		out = append(out, bytes.Repeat([]byte{fill}, 344)...)
		out = append(out, edge...)
		return out
	}
	rawA := mk('a')
	rawB := mk('b')
	if len(rawA) != len(rawB) {
		t.Fatal("fixture lengths must match")
	}
	if !bytes.Equal(rawA[:8], rawB[:8]) || !bytes.Equal(rawA[len(rawA)-8:], rawB[len(rawB)-8:]) {
		t.Fatal("fixture must share edge windows")
	}

	bufA, okA := CompressContentStreamCached(rawA)
	if !okA || bufA == nil {
		t.Fatal("expected compressed output for A")
	}
	decA := gunzip(t, bufA)
	PutCompressBuffer(bufA)

	bufB, okB := CompressContentStreamCached(rawB)
	if !okB || bufB == nil {
		t.Fatal("expected compressed output for B")
	}
	decB := gunzip(t, bufB)
	PutCompressBuffer(bufB)

	if !bytes.Equal(decA, rawA) {
		t.Fatal("page A stream corrupted by cache reuse")
	}
	if !bytes.Equal(decB, rawB) {
		t.Fatal("page B stream corrupted by cache reuse (sampling collision)")
	}
}

// TestPutCompressBufferDropsOversized asserts buffers with cap > 256 KiB are
// not retained by the pool.
func TestPutCompressBufferDropsOversized(t *testing.T) {
	for range 4 {
		CompressBufPool.Put(new(bytes.Buffer))
	}
	big := GetCompressBuffer()
	big.Grow(300 * 1024)
	if big.Cap() <= 256*1024 {
		t.Skip("could not grow an oversized buffer")
	}
	PutCompressBuffer(big)

	for range 64 {
		got := GetCompressBuffer()
		if got.Cap() > 256*1024 {
			t.Fatalf("oversized buffer retained in pool (cap %d)", got.Cap())
		}
		PutCompressBuffer(got)
	}
}
