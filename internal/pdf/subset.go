package pdf

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"sort"
)

// SubsetTTF creates a subset of the font containing only the specified glyphs
// This significantly reduces the embedded font size in the PDF
func SubsetTTF(font *TTFFont, usedGlyphs []uint16) ([]byte, map[uint16]uint16, error) {
	if len(usedGlyphs) == 0 {
		return nil, nil, errors.New("no glyphs to subset")
	}

	// Ensure glyphs are sorted and include .notdef (glyph 0)
	glyphSet := make(map[uint16]bool)
	glyphSet[0] = true // Always include .notdef
	for _, glyph := range usedGlyphs {
		if glyph < font.NumGlyphs {
			glyphSet[glyph] = true
		}
	}

	// Convert to sorted slice
	sortedGlyphs := make([]uint16, 0, len(glyphSet))
	for glyph := range glyphSet {
		sortedGlyphs = append(sortedGlyphs, glyph)
	}
	sort.Slice(sortedGlyphs, func(i, j int) bool {
		return sortedGlyphs[i] < sortedGlyphs[j]
	})

	// Create old-to-new glyph ID mapping
	oldToNew := make(map[uint16]uint16)
	for newID, oldID := range sortedGlyphs {
		oldToNew[oldID] = uint16(newID)
	}

	// Build the subset font
	return buildSubsetFont(font, sortedGlyphs, oldToNew)
}

// SubsetTTFForText creates a font subset containing only glyphs used in the given text
func SubsetTTFForText(font *TTFFont, text string) ([]byte, map[uint16]uint16, error) {
	usedGlyphs := font.GetUsedGlyphs(text)
	return SubsetTTF(font, usedGlyphs)
}

// buildSubsetFont creates a new TTF font file with only the specified glyphs
func buildSubsetFont(font *TTFFont, glyphs []uint16, oldToNew map[uint16]uint16) ([]byte, map[uint16]uint16, error) {
	var buf bytes.Buffer

	// Tables we need to include in the subset
	// Required tables: cmap, glyf, head, hhea, hmtx, loca, maxp, name, post
	// Optional but recommended: OS/2, cvt, fpgm, prep

	// Collect table data
	tables := make(map[string][]byte)

	// Generate required tables
	tables["head"] = subsetHead(font)
	tables["hhea"] = subsetHhea(font, uint16(len(glyphs)))
	tables["maxp"] = subsetMaxp(font, uint16(len(glyphs)))

	// Generate glyf and loca tables
	glyfData, locaData, isShortLoca := subsetGlyfAndLoca(font, glyphs)
	tables["glyf"] = glyfData
	tables["loca"] = locaData

	// Update head table with loca format
	if isShortLoca {
		tables["head"][50] = 0
		tables["head"][51] = 0
	} else {
		tables["head"][50] = 0
		tables["head"][51] = 1
	}

	// Generate hmtx table
	tables["hmtx"] = subsetHmtx(font, glyphs)

	// Generate cmap table with new glyph IDs
	tables["cmap"] = subsetCmap(font, oldToNew)

	// Generate post table (minimal version)
	tables["post"] = subsetPost(font)

	// Generate name table
	tables["name"] = subsetName(font)

	// Copy OS/2 table if present (with minor modifications)
	if os2Table, ok := font.Tables["OS/2"]; ok {
		if os2Table.Offset+os2Table.Length <= uint32(len(font.RawData)) {
			tables["OS/2"] = make([]byte, os2Table.Length)
			copy(tables["OS/2"], font.RawData[os2Table.Offset:os2Table.Offset+os2Table.Length])
		}
	}

	// Copy optional tables if they exist
	optionalTables := []string{"cvt ", "fpgm", "prep"}
	for _, tableName := range optionalTables {
		if entry, ok := font.Tables[tableName]; ok {
			if entry.Offset+entry.Length <= uint32(len(font.RawData)) {
				tables[tableName] = make([]byte, entry.Length)
				copy(tables[tableName], font.RawData[entry.Offset:entry.Offset+entry.Length])
			}
		}
	}

	// Calculate number of tables and offset table values
	numTables := uint16(len(tables))
	searchRange := uint16(1)
	entrySelector := uint16(0)
	for searchRange*2 <= numTables {
		searchRange *= 2
		entrySelector++
	}
	searchRange *= 16
	rangeShift := numTables*16 - searchRange

	// Write offset table
	if err := binary.Write(&buf, binary.BigEndian, uint32(0x00010000)); err != nil { // sfntVersion (TrueType)
		return nil, nil, err
	}
	if err := binary.Write(&buf, binary.BigEndian, numTables); err != nil {
		return nil, nil, err
	}
	if err := binary.Write(&buf, binary.BigEndian, searchRange); err != nil {
		return nil, nil, err
	}
	if err := binary.Write(&buf, binary.BigEndian, entrySelector); err != nil {
		return nil, nil, err
	}
	if err := binary.Write(&buf, binary.BigEndian, rangeShift); err != nil {
		return nil, nil, err
	}

	// Calculate table offsets
	tableOffset := uint32(12 + numTables*16) // After offset table and table directory

	// Sort table names for consistent output
	tableNames := make([]string, 0, len(tables))
	for name := range tables {
		tableNames = append(tableNames, name)
	}
	sort.Strings(tableNames)

	// Write table directory
	tableOffsets := make(map[string]uint32)
	for _, name := range tableNames {
		data := tables[name]

		// Pad table name to 4 bytes
		tag := []byte(name)
		for len(tag) < 4 {
			tag = append(tag, ' ')
		}

		checksum := calculateChecksum(data)
		length := uint32(len(data))

		buf.Write(tag[:4])
		if err := binary.Write(&buf, binary.BigEndian, checksum); err != nil {
			return nil, nil, err
		}
		if err := binary.Write(&buf, binary.BigEndian, tableOffset); err != nil {
			return nil, nil, err
		}
		if err := binary.Write(&buf, binary.BigEndian, length); err != nil {
			return nil, nil, err
		}

		tableOffsets[name] = tableOffset

		// Align to 4-byte boundary
		paddedLen := (length + 3) &^ 3
		tableOffset += paddedLen
	}

	// Write table data
	for _, name := range tableNames {
		data := tables[name]
		buf.Write(data)

		// Pad to 4-byte boundary
		padding := (4 - len(data)%4) % 4
		for i := 0; i < padding; i++ {
			buf.WriteByte(0)
		}
	}

	// Update head checksum adjustment
	result := buf.Bytes()
	headOffset := tableOffsets["head"]
	updateHeadChecksum(result, headOffset)

	return result, oldToNew, nil
}

// subsetHead generates the head table for the subset font
func subsetHead(font *TTFFont) []byte {
	headTable := font.Tables["head"]
	result := make([]byte, headTable.Length)
	copy(result, font.RawData[headTable.Offset:headTable.Offset+headTable.Length])

	// Clear checksumAdjustment (will be recalculated)
	result[8] = 0
	result[9] = 0
	result[10] = 0
	result[11] = 0

	return result
}

// subsetHhea generates the hhea table with updated numberOfHMetrics
func subsetHhea(font *TTFFont, numGlyphs uint16) []byte {
	hheaTable := font.Tables["hhea"]
	result := make([]byte, hheaTable.Length)
	copy(result, font.RawData[hheaTable.Offset:hheaTable.Offset+hheaTable.Length])

	// Update numberOfHMetrics (last 2 bytes)
	binary.BigEndian.PutUint16(result[len(result)-2:], numGlyphs)

	return result
}

// subsetMaxp generates the maxp table with updated numGlyphs
func subsetMaxp(font *TTFFont, numGlyphs uint16) []byte {
	maxpTable := font.Tables["maxp"]
	result := make([]byte, maxpTable.Length)
	copy(result, font.RawData[maxpTable.Offset:maxpTable.Offset+maxpTable.Length])

	// Update numGlyphs (at offset 4)
	binary.BigEndian.PutUint16(result[4:], numGlyphs)

	return result
}

// subsetGlyfAndLoca generates the glyf and loca tables for the subset
func subsetGlyfAndLoca(font *TTFFont, glyphs []uint16) ([]byte, []byte, bool) {
	glyfTable, hasGlyf := font.Tables["glyf"]
	locaTable, hasLoca := font.Tables["loca"]

	if !hasGlyf || !hasLoca {
		// Return empty tables if not present (e.g., CFF-based font)
		return []byte{}, []byte{0, 0}, true
	}

	// Determine loca format from head table
	headTable := font.Tables["head"]
	isShortLoca := font.RawData[headTable.Offset+50] == 0 && font.RawData[headTable.Offset+51] == 0

	// Read original loca table
	locaData := font.RawData[locaTable.Offset : locaTable.Offset+locaTable.Length]
	glyfData := font.RawData[glyfTable.Offset : glyfTable.Offset+glyfTable.Length]

	// Build new glyf data
	var newGlyf bytes.Buffer
	newOffsets := make([]uint32, len(glyphs)+1)

	for i, glyphID := range glyphs {
		newOffsets[i] = uint32(newGlyf.Len())

		// Get original glyph offset and length
		var offset, nextOffset uint32
		if isShortLoca {
			offset = uint32(binary.BigEndian.Uint16(locaData[int(glyphID)*2:])) * 2
			nextOffset = uint32(binary.BigEndian.Uint16(locaData[int(glyphID)*2+2:])) * 2
		} else {
			offset = binary.BigEndian.Uint32(locaData[int(glyphID)*4:])
			nextOffset = binary.BigEndian.Uint32(locaData[int(glyphID)*4+4:])
		}

		if nextOffset > offset && offset < uint32(len(glyfData)) {
			length := nextOffset - offset
			if offset+length > uint32(len(glyfData)) {
				length = uint32(len(glyfData)) - offset
			}
			newGlyf.Write(glyfData[offset : offset+length])

			// Pad to even boundary for short loca format
			if newGlyf.Len()%2 != 0 {
				newGlyf.WriteByte(0)
			}
		}
	}
	newOffsets[len(glyphs)] = uint32(newGlyf.Len())

	// Determine if we can use short loca format
	useShortLoca := newOffsets[len(glyphs)] <= 0xFFFF*2

	// Build new loca table
	var newLoca bytes.Buffer
	if useShortLoca {
		for _, offset := range newOffsets {
			if err := binary.Write(&newLoca, binary.BigEndian, uint16(offset/2)); err != nil {
				return nil, nil, false
			}
		}
	} else {
		for _, offset := range newOffsets {
			if err := binary.Write(&newLoca, binary.BigEndian, offset); err != nil {
				return nil, nil, false
			}
		}
	}

	return newGlyf.Bytes(), newLoca.Bytes(), useShortLoca
}

// subsetHmtx generates the hmtx table for the subset
func subsetHmtx(font *TTFFont, glyphs []uint16) []byte {
	var buf bytes.Buffer

	for _, glyphID := range glyphs {
		width := font.GetGlyphWidth(glyphID)
		if err := binary.Write(&buf, binary.BigEndian, width); err != nil {
			return nil
		}
		if err := binary.Write(&buf, binary.BigEndian, int16(0)); err != nil {
			return nil
		}
	}

	return buf.Bytes()
}

// subsetCmap generates a format 4 cmap table with remapped glyph IDs
func subsetCmap(font *TTFFont, oldToNew map[uint16]uint16) []byte {
	var buf bytes.Buffer

	// Build character to new glyph ID mapping
	charToNewGlyph := make(map[uint16]uint16)
	for char, oldGlyph := range font.CharToGlyph {
		if char <= 0xFFFF {
			if newGlyph, ok := oldToNew[oldGlyph]; ok {
				charToNewGlyph[uint16(char)] = newGlyph
			}
		}
	}

	// Sort characters
	chars := make([]uint16, 0, len(charToNewGlyph))
	for char := range charToNewGlyph {
		chars = append(chars, char)
	}
	sort.Slice(chars, func(i, j int) bool {
		return chars[i] < chars[j]
	})

	// Build segments
	type segment struct {
		startCode uint16
		endCode   uint16
		idDelta   int16
	}

	var segments []segment
	if len(chars) > 0 {
		segStart := chars[0]
		prevChar := chars[0]
		prevGlyph := charToNewGlyph[chars[0]]

		for i := 1; i < len(chars); i++ {
			char := chars[i]
			glyph := charToNewGlyph[char]

			// Check if this continues the current segment
			if char == prevChar+1 && glyph == prevGlyph+1 {
				prevChar = char
				prevGlyph = glyph
			} else {
				// End current segment
				delta := int16(charToNewGlyph[segStart]) - int16(segStart)
				segments = append(segments, segment{segStart, prevChar, delta})

				// Start new segment
				segStart = char
				prevChar = char
				prevGlyph = glyph
			}
		}

		// Don't forget the last segment
		delta := int16(charToNewGlyph[segStart]) - int16(segStart)
		segments = append(segments, segment{segStart, prevChar, delta})
	}

	// Add terminating segment
	segments = append(segments, segment{0xFFFF, 0xFFFF, 1})

	segCount := uint16(len(segments))

	// Calculate searchRange, entrySelector, rangeShift
	searchRange := uint16(1)
	entrySelector := uint16(0)
	for searchRange*2 <= segCount {
		searchRange *= 2
		entrySelector++
	}
	searchRange *= 2
	rangeShift := segCount*2 - searchRange

	// Build format 4 subtable
	var format4 bytes.Buffer

	// Header
	if err := binary.Write(&format4, binary.BigEndian, uint16(4)); err != nil { // format
		return nil
	}
	if err := binary.Write(&format4, binary.BigEndian, uint16(0)); err != nil { // length (placeholder)
		return nil
	}
	if err := binary.Write(&format4, binary.BigEndian, uint16(0)); err != nil { // language
		return nil
	}
	if err := binary.Write(&format4, binary.BigEndian, segCount*2); err != nil { // segCountX2
		return nil
	}
	if err := binary.Write(&format4, binary.BigEndian, searchRange); err != nil {
		return nil
	}
	if err := binary.Write(&format4, binary.BigEndian, entrySelector); err != nil {
		return nil
	}
	if err := binary.Write(&format4, binary.BigEndian, rangeShift); err != nil {
		return nil
	}

	// endCode array
	for _, seg := range segments {
		if err := binary.Write(&format4, binary.BigEndian, seg.endCode); err != nil {
			return nil
		}
	}

	// reservedPad
	if err := binary.Write(&format4, binary.BigEndian, uint16(0)); err != nil {
		return nil
	}

	// startCode array
	for _, seg := range segments {
		if err := binary.Write(&format4, binary.BigEndian, seg.startCode); err != nil {
			return nil
		}
	}

	// idDelta array
	for _, seg := range segments {
		if err := binary.Write(&format4, binary.BigEndian, seg.idDelta); err != nil {
			return nil
		}
	}

	// idRangeOffset array (all zeros for our simple mapping)
	for range segments {
		if err := binary.Write(&format4, binary.BigEndian, uint16(0)); err != nil {
			return nil
		}
	}

	// Update length
	format4Data := format4.Bytes()
	binary.BigEndian.PutUint16(format4Data[2:], uint16(len(format4Data)))

	// Build cmap table
	// cmap header
	if err := binary.Write(&buf, binary.BigEndian, uint16(0)); err != nil { // version
		return nil
	}
	if err := binary.Write(&buf, binary.BigEndian, uint16(1)); err != nil { // numTables
		return nil
	}

	// Encoding record
	if err := binary.Write(&buf, binary.BigEndian, uint16(3)); err != nil { // platformID (Windows)
		return nil
	}
	if err := binary.Write(&buf, binary.BigEndian, uint16(1)); err != nil { // encodingID (Unicode BMP)
		return nil
	}
	if err := binary.Write(&buf, binary.BigEndian, uint32(12)); err != nil { // offset to subtable
		return nil
	}

	// Write format 4 subtable
	buf.Write(format4Data)

	return buf.Bytes()
}

// subsetPost generates a minimal post table
func subsetPost(font *TTFFont) []byte {
	var buf bytes.Buffer

	// Version 3.0 (no glyph names)
	if err := binary.Write(&buf, binary.BigEndian, uint32(0x00030000)); err != nil {
		return nil
	}

	// italicAngle (16.16 fixed)
	italicAngleFixed := int32(font.ItalicAngle * 65536)
	if err := binary.Write(&buf, binary.BigEndian, italicAngleFixed); err != nil {
		return nil
	}

	// underlinePosition, underlineThickness
	if err := binary.Write(&buf, binary.BigEndian, int16(-100)); err != nil {
		return nil
	}
	if err := binary.Write(&buf, binary.BigEndian, int16(50)); err != nil {
		return nil
	}

	// isFixedPitch
	if font.IsFixedPitch {
		if err := binary.Write(&buf, binary.BigEndian, uint32(1)); err != nil {
			return nil
		}
	} else {
		if err := binary.Write(&buf, binary.BigEndian, uint32(0)); err != nil {
			return nil
		}
	}

	// minMemType42, maxMemType42, minMemType1, maxMemType1
	if err := binary.Write(&buf, binary.BigEndian, uint32(0)); err != nil {
		return nil
	}
	if err := binary.Write(&buf, binary.BigEndian, uint32(0)); err != nil {
		return nil
	}
	if err := binary.Write(&buf, binary.BigEndian, uint32(0)); err != nil {
		return nil
	}
	if err := binary.Write(&buf, binary.BigEndian, uint32(0)); err != nil {
		return nil
	}

	return buf.Bytes()
}

// subsetName generates a minimal name table
func subsetName(font *TTFFont) []byte {
	var buf bytes.Buffer

	// We'll include: Copyright, Family, Subfamily, UniqueID, FullName, PostScriptName
	names := []struct {
		nameID uint16
		value  string
	}{
		{0, "Subset font"},       // Copyright
		{1, font.FamilyName},     // Family
		{2, "Regular"},           // Subfamily
		{4, font.FullName},       // Full name
		{5, font.Version},        // Version
		{6, font.PostScriptName}, // PostScript name
	}

	// Calculate string storage
	var stringData bytes.Buffer
	type nameRecord struct {
		platformID uint16
		encodingID uint16
		languageID uint16
		nameID     uint16
		length     uint16
		offset     uint16
	}

	var records []nameRecord
	for _, name := range names {
		// Windows Unicode BMP
		offset := uint16(stringData.Len())
		encoded := encodeUTF16BE(name.value)
		stringData.Write(encoded)

		records = append(records, nameRecord{
			platformID: 3,      // Windows
			encodingID: 1,      // Unicode BMP
			languageID: 0x0409, // English US
			nameID:     name.nameID,
			length:     uint16(len(encoded)),
			offset:     offset,
		})
	}

	// Write name table header
	if err := binary.Write(&buf, binary.BigEndian, uint16(0)); err != nil { // format
		return nil
	}
	if err := binary.Write(&buf, binary.BigEndian, uint16(len(records))); err != nil { // count
		return nil
	}
	if err := binary.Write(&buf, binary.BigEndian, uint16(6+len(records)*12)); err != nil { // stringOffset
		return nil
	}

	// Write name records
	for _, rec := range records {
		if err := binary.Write(&buf, binary.BigEndian, rec.platformID); err != nil {
			return nil
		}
		if err := binary.Write(&buf, binary.BigEndian, rec.encodingID); err != nil {
			return nil
		}
		if err := binary.Write(&buf, binary.BigEndian, rec.languageID); err != nil {
			return nil
		}
		if err := binary.Write(&buf, binary.BigEndian, rec.nameID); err != nil {
			return nil
		}
		if err := binary.Write(&buf, binary.BigEndian, rec.length); err != nil {
			return nil
		}
		if err := binary.Write(&buf, binary.BigEndian, rec.offset); err != nil {
			return nil
		}
	}

	// Write string data
	buf.Write(stringData.Bytes())

	return buf.Bytes()
}

// calculateChecksum calculates the checksum for a font table
func calculateChecksum(data []byte) uint32 {
	// Pad to 4-byte boundary
	padded := data
	if len(data)%4 != 0 {
		padded = make([]byte, len(data)+(4-len(data)%4))
		copy(padded, data)
	}

	var sum uint32
	for i := 0; i < len(padded); i += 4 {
		sum += binary.BigEndian.Uint32(padded[i:])
	}

	return sum
}

// updateHeadChecksum updates the checksumAdjustment field in the head table
func updateHeadChecksum(fontData []byte, headOffset uint32) {
	// Calculate checksum of entire font
	fontChecksum := calculateChecksum(fontData)

	// checksumAdjustment = 0xB1B0AFBA - fontChecksum
	adjustment := uint32(0xB1B0AFBA) - fontChecksum

	// Write to head table (offset 8)
	binary.BigEndian.PutUint32(fontData[headOffset+8:], adjustment)
}

// encodeUTF16BE encodes a string as UTF-16BE
func encodeUTF16BE(s string) []byte {
	var buf bytes.Buffer
	for _, r := range s {
		if r <= 0xFFFF {
			buf.WriteByte(byte(r >> 8))
			buf.WriteByte(byte(r))
		} else {
			// Surrogate pair
			r -= 0x10000
			high := uint16(0xD800 + (r >> 10))
			low := uint16(0xDC00 + (r & 0x3FF))
			buf.WriteByte(byte(high >> 8))
			buf.WriteByte(byte(high))
			buf.WriteByte(byte(low >> 8))
			buf.WriteByte(byte(low))
		}
	}
	return buf.Bytes()
}

// CompressFontData compresses font data using zlib (FlateDecode)
func CompressFontData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	_, err := w.Write(data)
	if err != nil {
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
