package pdf

import (
	"testing"
)

// TestPutRGBDataBufferDropsOversized asserts buffers past the pool cap are
// dropped instead of retained, keeping pooled memory bounded (C1).
func TestPutRGBDataBufferDropsOversized(t *testing.T) {
	big := make([]byte, maxPooledRGBDataCap+1024)
	putRGBDataBuffer(big)

	got := getRGBDataBuffer(16)
	if cap(got) > maxPooledRGBDataCap {
		t.Fatalf("pool retained oversized buffer: cap = %d, max = %d", cap(got), maxPooledRGBDataCap)
	}
	putRGBDataBuffer(got)
}

// TestPutRGBDataBufferBoundedReuse asserts normal buffers still recycle
// through the pool (sync.Pool reuse preserved under the cap guard).
func TestPutRGBDataBufferBoundedReuse(t *testing.T) {
	buf := getRGBDataBuffer(1024)
	if len(buf) != 1024 {
		t.Fatalf("len = %d, want 1024", len(buf))
	}
	putRGBDataBuffer(buf)

	got := getRGBDataBuffer(1024)
	if len(got) != 1024 {
		t.Fatalf("len = %d, want 1024", len(got))
	}
	if cap(got) > maxPooledRGBDataCap {
		t.Fatalf("recycled cap = %d exceeds max %d", cap(got), maxPooledRGBDataCap)
	}
	putRGBDataBuffer(got)

	putRGBDataBuffer(nil) // must not panic
}
