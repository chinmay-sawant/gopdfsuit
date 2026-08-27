package font

import (
	"bytes"
	"encoding/binary"
	"sort"
)

// CompactUnusedGlyphs drops outline data for glyphs that are not in usedGlyphs,
// keeping original glyph IDs so existing PDF content streams stay valid.
// Returns the original bytes if the font is not a glyf-based TTF or nothing
// would be saved.
func CompactUnusedGlyphs(ttf []byte, usedGlyphs []uint16) ([]byte, error) {
	font, err := ParseTTF(ttf)
	if err != nil {
		return nil, err
	}
	if _, ok := font.Tables["glyf"]; !ok {
		return ttf, nil
	}

	used := make(map[uint16]bool, len(usedGlyphs)+1)
	used[0] = true
	for _, g := range usedGlyphs {
		if g < font.NumGlyphs {
			used[g] = true
		}
	}
	addCompositeComponents(font, used)
	if uint16(len(used)) >= font.NumGlyphs {
		return ttf, nil
	}

	glyf, loca, shortLoca := compactGlyfKeepIDs(font, used)
	tables := copyFontTables(font)
	tables["glyf"] = glyf
	tables["loca"] = loca
	if head, ok := tables["head"]; ok && len(head) > 51 {
		head = bytes.Clone(head)
		head[8], head[9], head[10], head[11] = 0, 0, 0, 0
		if shortLoca {
			head[50], head[51] = 0, 0
		} else {
			head[50], head[51] = 0, 1
		}
		tables["head"] = head
	}

	out := writeSFNT(tables)
	if len(out) == 0 || len(out) >= len(ttf) {
		return ttf, nil
	}
	return out, nil
}

func compactGlyfKeepIDs(font *TTFFont, used map[uint16]bool) (glyf []byte, loca []byte, shortLoca bool) {
	var glyfBuf bytes.Buffer
	offsets := make([]uint32, int(font.NumGlyphs)+1)
	for gid := uint16(0); gid < font.NumGlyphs; gid++ {
		offsets[gid] = uint32(glyfBuf.Len())
		if !used[gid] {
			continue
		}
		data := getGlyphData(font, gid)
		if len(data) == 0 {
			continue
		}
		glyfBuf.Write(data)
		if glyfBuf.Len()%2 != 0 {
			glyfBuf.WriteByte(0)
		}
	}
	offsets[font.NumGlyphs] = uint32(glyfBuf.Len())
	glyf = glyfBuf.Bytes()

	shortLoca = offsets[font.NumGlyphs] <= 0xFFFF*2
	var locaBuf bytes.Buffer
	if shortLoca {
		for _, off := range offsets {
			_ = binary.Write(&locaBuf, binary.BigEndian, uint16(off/2))
		}
	} else {
		for _, off := range offsets {
			_ = binary.Write(&locaBuf, binary.BigEndian, off)
		}
	}
	return glyf, locaBuf.Bytes(), shortLoca
}

func copyFontTables(font *TTFFont) map[string][]byte {
	tables := make(map[string][]byte, len(font.Tables))
	for name, entry := range font.Tables {
		end := entry.Offset + entry.Length
		if end > uint32(len(font.RawData)) {
			continue
		}
		tables[name] = font.RawData[entry.Offset:end]
	}
	return tables
}

func writeSFNT(tables map[string][]byte) []byte {
	if len(tables) == 0 {
		return nil
	}
	numTables := uint16(len(tables))
	searchRange := uint16(1)
	entrySelector := uint16(0)
	for searchRange*2 <= numTables {
		searchRange *= 2
		entrySelector++
	}
	searchRange *= 16
	rangeShift := numTables*16 - searchRange

	names := make([]string, 0, len(tables))
	for name := range tables {
		names = append(names, name)
	}
	sort.Strings(names)

	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.BigEndian, uint32(0x00010000))
	_ = binary.Write(&buf, binary.BigEndian, numTables)
	_ = binary.Write(&buf, binary.BigEndian, searchRange)
	_ = binary.Write(&buf, binary.BigEndian, entrySelector)
	_ = binary.Write(&buf, binary.BigEndian, rangeShift)

	tableOffset := uint32(12 + numTables*16)
	tableOffsets := make(map[string]uint32, len(names))
	for _, name := range names {
		data := tables[name]
		var tag [4]byte
		copy(tag[:], name)
		for i := len(name); i < 4; i++ {
			tag[i] = ' '
		}
		length := uint32(len(data))
		buf.Write(tag[:])
		_ = binary.Write(&buf, binary.BigEndian, calculateChecksum(data))
		_ = binary.Write(&buf, binary.BigEndian, tableOffset)
		_ = binary.Write(&buf, binary.BigEndian, length)
		tableOffsets[name] = tableOffset
		tableOffset += (length + 3) &^ 3
	}
	for _, name := range names {
		data := tables[name]
		buf.Write(data)
		pad := (4 - len(data)%4) % 4
		for range pad {
			buf.WriteByte(0)
		}
	}
	result := buf.Bytes()
	if off, ok := tableOffsets["head"]; ok {
		updateHeadChecksum(result, off)
	}
	return result
}
