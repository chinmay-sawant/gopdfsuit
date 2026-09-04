package font

import (
	"encoding/binary"
	"testing"
)

// craftTestFont builds a minimal synthetic font with short loca format:
// 3 glyphs, head (54 bytes), hhea, maxp, glyf and a caller-sized loca.
func craftTestFont(locaEntries uint16) *TTFFont {
	raw := make([]byte, 512)
	put := func(off int, data []byte) {
		copy(raw[off:], data)
	}

	head := make([]byte, 54)
	put(0, head) // locaFormat stays 0 => short
	hhea := make([]byte, 36)
	put(64, hhea)
	maxp := make([]byte, 32)
	binary.BigEndian.PutUint16(maxp[4:], 3)
	put(128, maxp)

	// glyf: three 4-byte dummy glyphs
	glyf := []byte{0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0}
	put(192, glyf)

	// loca short entries: offsets in uint16 words (each glyph 4 bytes => 2 words)
	loca := make([]byte, int(locaEntries)*2)
	for i := range int(locaEntries) {
		binary.BigEndian.PutUint16(loca[i*2:], uint16(i*2))
	}
	put(256, loca)

	return &TTFFont{
		NumGlyphs:   3,
		CharToGlyph: map[rune]uint16{'A': 1, 'B': 2},
		GlyphWidths: []uint16{500, 500, 500},
		RawData:     raw,
		Tables: map[string]TableEntry{
			"head": {Tag: "head", Offset: 0, Length: 54},
			"hhea": {Tag: "hhea", Offset: 64, Length: 36},
			"maxp": {Tag: "maxp", Offset: 128, Length: 32},
			"glyf": {Tag: "glyf", Offset: 192, Length: uint32(len(glyf))},
			"loca": {Tag: "loca", Offset: 256, Length: uint32(len(loca))},
		},
	}
}

// TestSubsetGlyfTruncatedLoca proves a truncated loca table no longer panics:
// out-of-range glyphs are skipped instead of indexing past the slice.
func TestSubsetGlyfTruncatedLoca(t *testing.T) {
	font := craftTestFont(2) // only entries for glyphs 0..1; glyph 2 needs bytes [4:8)
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("subsetGlyfAndLoca panicked on truncated loca: %v", r)
			}
		}()
		glyf, loca, _ := subsetGlyfAndLoca(font, []uint16{0, 1, 2})
		if len(glyf) == 0 || len(loca) == 0 {
			t.Fatalf("expected non-empty subset output, got glyf=%d loca=%d", len(glyf), len(loca))
		}
	}()
}

// TestSubsetTablesMissingOrShort proves subsetHead/Hhea/Maxp return errors
// instead of indexing into missing or truncated tables.
func TestSubsetTablesMissingOrShort(t *testing.T) {
	font := craftTestFont(4)

	delete(font.Tables, "head")
	if _, err := subsetHead(font); err == nil {
		t.Fatalf("subsetHead with missing table: expected error")
	}
	font = craftTestFont(4)
	font.Tables["head"] = TableEntry{Tag: "head", Offset: 0, Length: 8}
	if _, err := subsetHead(font); err == nil {
		t.Fatalf("subsetHead with short table: expected error")
	}

	font = craftTestFont(4)
	delete(font.Tables, "hhea")
	if _, err := subsetHhea(font, 1); err == nil {
		t.Fatalf("subsetHhea with missing table: expected error")
	}
	font = craftTestFont(4)
	font.Tables["hhea"] = TableEntry{Tag: "hhea", Offset: 64, Length: 1}
	if _, err := subsetHhea(font, 1); err == nil {
		t.Fatalf("subsetHhea with short table: expected error")
	}

	font = craftTestFont(4)
	delete(font.Tables, "maxp")
	if _, err := subsetMaxp(font, 1); err == nil {
		t.Fatalf("subsetMaxp with missing table: expected error")
	}
	font = craftTestFont(4)
	font.Tables["maxp"] = TableEntry{Tag: "maxp", Offset: 128, Length: 4}
	if _, err := subsetMaxp(font, 1); err == nil {
		t.Fatalf("subsetMaxp with short table: expected error")
	}
}

// cmapGlyph decodes a format-4 cmap built by subsetCmap.
func cmapGlyph(t *testing.T, cmap []byte, char uint16) uint16 {
	t.Helper()
	if len(cmap) < 12 {
		t.Fatalf("cmap too short: %d", len(cmap))
	}
	sub := int(binary.BigEndian.Uint32(cmap[8:12]))
	if sub+14 > len(cmap) {
		t.Fatalf("cmap subtable offset out of range")
	}
	f4 := cmap[sub:]
	segCount := int(binary.BigEndian.Uint16(f4[6:8])) / 2
	p := 14
	endCode := make([]uint16, segCount)
	for i := range endCode {
		endCode[i] = binary.BigEndian.Uint16(f4[p+i*2:])
	}
	p += segCount*2 + 2
	startCode := make([]uint16, segCount)
	for i := range startCode {
		startCode[i] = binary.BigEndian.Uint16(f4[p+i*2:])
	}
	p += segCount * 2
	idDelta := make([]uint16, segCount)
	for i := range idDelta {
		idDelta[i] = binary.BigEndian.Uint16(f4[p+i*2:])
	}
	p += segCount * 2
	rangeBase := p
	for i := range endCode {
		if char < startCode[i] || char > endCode[i] {
			continue
		}
		ro := binary.BigEndian.Uint16(f4[p+i*2:])
		if ro == 0 {
			return char + idDelta[i]
		}
		off := rangeBase + i*2 + int(ro) + int(char-startCode[i])*2
		if off+2 > len(f4) {
			t.Fatalf("glyphIdArray offset out of range")
		}
		return binary.BigEndian.Uint16(f4[off:])
	}
	t.Fatalf("char U+%04X not covered by cmap", char)
	return 0
}

// TestSubsetCmapLargeGap proves a mapping whose idDelta overflows int16
// falls back to idRangeOffset segments and still round-trips exactly.
func TestSubsetCmapLargeGap(t *testing.T) {
	font := &TTFFont{
		CharToGlyph: map[rune]uint16{0x20: 10, 0xFFFE: 20},
	}
	// After subsetting, 0xFFFE maps to glyph 0: delta = -65534 overflows int16.
	oldToNew := map[uint16]uint16{10: 1, 20: 0}
	cmap := subsetCmap(font, oldToNew)
	if len(cmap) == 0 {
		t.Fatalf("subsetCmap returned empty table")
	}
	if got := cmapGlyph(t, cmap, 0x20); got != 1 {
		t.Fatalf("U+0020 mapped to %d, want 1", got)
	}
	if got := cmapGlyph(t, cmap, 0xFFFE); got != 0 {
		t.Fatalf("U+FFFE mapped to %d, want 0", got)
	}
}
