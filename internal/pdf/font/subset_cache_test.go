package font

import (
	"testing"
)

func TestSubsetCacheReusesIdenticalGlyphSet(t *testing.T) {
	ClearSubsetCache()

	font := &TTFFont{PostScriptName: "TestFont"}
	usedGlyphs := []uint16{1, 5, 10}
	data := []byte("subset-bytes")
	oldToNew := map[uint16]uint16{1: 1, 5: 2, 10: 3}

	storeCachedSubset(font, usedGlyphs, data, oldToNew)

	cached, ok := lookupCachedSubset(font, usedGlyphs)
	if !ok {
		t.Fatal("expected cache hit")
	}
	if string(cached.data) != string(data) {
		t.Fatalf("cached data = %q, want %q", cached.data, data)
	}
	if len(cached.oldToNew) != len(oldToNew) {
		t.Fatalf("cached mapping length = %d, want %d", len(cached.oldToNew), len(oldToNew))
	}
}

func TestSubsetCacheBoundsEntries(t *testing.T) {
	ClearSubsetCache()

	base := &TTFFont{PostScriptName: "BoundFont"}
	data := []byte("x")
	mapping := map[uint16]uint16{1: 1}

	// Fill cache up to the limit with unique glyph sets.
	for i := uint16(0); i < maxSubsetCacheEntries+10; i++ {
		storeCachedSubset(base, []uint16{i + 100}, data, mapping)
	}

	// After an overflow clear, an early entry should no longer be present,
	// but a recently stored entry should still be cached.
	_, ok := lookupCachedSubset(base, []uint16{100})
	if ok {
		t.Fatal("expected early entry to be evicted after cache overflow")
	}

	_, ok = lookupCachedSubset(base, []uint16{maxSubsetCacheEntries + 100})
	if !ok {
		t.Fatal("expected recent entry to still be cached")
	}
}

func TestSubsetCacheClear(t *testing.T) {
	ClearSubsetCache()

	font := &TTFFont{PostScriptName: "ClearFont"}
	storeCachedSubset(font, []uint16{1}, []byte("d"), map[uint16]uint16{1: 1})

	if _, ok := lookupCachedSubset(font, []uint16{1}); !ok {
		t.Fatal("expected entry before clear")
	}

	ClearSubsetCache()

	if _, ok := lookupCachedSubset(font, []uint16{1}); ok {
		t.Fatal("expected cache miss after clear")
	}
}
