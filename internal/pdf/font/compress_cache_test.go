package font

import (
	"bytes"
	"testing"
)

func TestCompressContentStreamCachedReusesResult(t *testing.T) {
	ClearPageCompressCache()
	raw := bytes.Repeat([]byte("BT /F1 12 Tf (hello) Tj ET\n"), 128)

	buf1, ok1 := CompressContentStreamCached(raw)
	if !ok1 || buf1 == nil {
		t.Fatal("expected compressed output")
	}
	len1 := buf1.Len()
	PutCompressBuffer(buf1)

	buf2, ok2 := CompressContentStreamCached(raw)
	if !ok2 || buf2 == nil {
		t.Fatal("expected cache hit")
	}
	if buf2.Len() != len1 {
		t.Fatalf("cache hit length = %d, want %d", buf2.Len(), len1)
	}
	PutCompressBuffer(buf2)

	var total int64
	for i := range compressShards {
		total += compressShards[i].count.Load()
	}
	if total != 1 {
		t.Fatalf("cache entry count = %d, want 1", total)
	}
}

func TestCompressContentStreamCachedStoresUncompressed(t *testing.T) {
	ClearPageCompressCache()
	raw := make([]byte, 64)
	for i := range raw {
		raw[i] = byte(i)
	}

	buf, ok := CompressContentStreamCached(raw)
	if ok || buf != nil {
		t.Fatal("expected store-uncompressed cache entry")
	}

	buf2, ok2 := CompressContentStreamCached(raw)
	if ok2 || buf2 != nil {
		t.Fatal("expected cached store-uncompressed decision")
	}

	var total int64
	for i := range compressShards {
		total += compressShards[i].count.Load()
	}
	if total != 1 {
		t.Fatalf("cache entry count = %d, want 1", total)
	}
}
