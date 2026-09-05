package font

import (
	"testing"
)

func TestSubsetCacheReusesIdenticalGlyphSet(t *testing.T) {
	ClearSubsetCache()

	font := &TTFFont{PostScriptName: "TestFont", contentID: [32]byte{1}}
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

	base := &TTFFont{PostScriptName: "BoundFont", contentID: [32]byte{2}}
	data := []byte("x")
	mapping := map[uint16]uint16{1: 1}

	// Fill cache up to the limit with unique glyph sets.
	for i := range uint16(maxSubsetCacheEntries + 10) {
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

	font := &TTFFont{PostScriptName: "ClearFont", contentID: [32]byte{3}}
	storeCachedSubset(font, []uint16{1}, []byte("d"), map[uint16]uint16{1: 1})

	if _, ok := lookupCachedSubset(font, []uint16{1}); !ok {
		t.Fatal("expected entry before clear")
	}

	ClearSubsetCache()

	if _, ok := lookupCachedSubset(font, []uint16{1}); ok {
		t.Fatal("expected cache miss after clear")
	}
}

func TestSubsetCacheSeparatesFontContentsWithSameNameAndGlyphs(t *testing.T) {
	ClearSubsetCache()

	fontA := &TTFFont{
		PostScriptName: "SharedName",
		contentID:      [32]byte{1},
	}
	fontB := &TTFFont{
		PostScriptName: "SharedName",
		contentID:      [32]byte{2},
	}
	glyphs := []uint16{0, 7}

	storeCachedSubset(fontA, glyphs, []byte("font-a-subset"), map[uint16]uint16{0: 0, 7: 1})
	if _, ok := lookupCachedSubset(fontB, glyphs); ok {
		t.Fatal("font B reused font A subset despite different content identity")
	}

	storeCachedSubset(fontB, glyphs, []byte("font-b-subset"), map[uint16]uint16{0: 0, 7: 1})
	for name, test := range map[string]struct {
		font *TTFFont
		want string
	}{
		"font A": {font: fontA, want: "font-a-subset"},
		"font B": {font: fontB, want: "font-b-subset"},
	} {
		t.Run(name, func(t *testing.T) {
			cached, ok := lookupCachedSubset(test.font, glyphs)
			if !ok {
				t.Fatal("expected cache hit")
			}
			if string(cached.data) != test.want {
				t.Fatalf("cached subset = %q, want %q", cached.data, test.want)
			}
		})
	}
}
