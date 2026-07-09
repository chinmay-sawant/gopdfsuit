package font

import (
	"bytes"
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
	glyphSet := make(map[uint16]bool, len(usedGlyphs)+1) // PERF-192
	glyphSet[0] = true                                   // Always include .notdef
	for _, glyph := range usedGlyphs {
		if glyph < font.NumGlyphs {
			glyphSet[glyph] = true
		}
	}

	// Resolve composite glyph dependencies: composite glyphs reference
	// component glyphs by GID, so those components must also be in the subset.
	addCompositeComponents(font, glyphSet)

	// Convert to sorted slice
	sortedGlyphs := make([]uint16, 0, len(glyphSet))
	for glyph := range glyphSet {
		sortedGlyphs = append(sortedGlyphs, glyph)
	}
	sort.Slice(sortedGlyphs, func(i, j int) bool {
		return sortedGlyphs[i] < sortedGlyphs[j]
	})

	// Create old-to-new glyph ID mapping
	oldToNew := make(map[uint16]uint16, len(sortedGlyphs)) // PERF-192
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
	// Tables we need to include in the subset
	// Required tables: cmap, glyf, head, hhea, hmtx, loca, maxp, name, post
	// Optional but recommended: OS/2, cvt, fpgm, prep

	// Collect table data (~12 common TrueType tables)
	tables := make(map[string][]byte, 16) // PERF-192

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
				// Share underlying RawData (read-only for remaining subset assembly)
				tables[tableName] = font.RawData[entry.Offset : entry.Offset+entry.Length]
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

	// Write offset table (PERF-107: manual BE helpers, no binary.Write)
	// Pre-size: 12-byte header + 16 bytes per table directory entry + table data
	var totalTableBytes int
	for _, d := range tables {
		totalTableBytes += (len(d) + 3) &^ 3
	}
	buf := make([]byte, 0, 12+int(numTables)*16+totalTableBytes)

	putU32BE(&buf, 0x00010000) // sfntVersion (TrueType)
	putU16BE(&buf, numTables)
	putU16BE(&buf, searchRange)
	putU16BE(&buf, entrySelector)
	putU16BE(&buf, rangeShift)

	// Calculate table offsets
	tableOffset := uint32(12 + numTables*16) // After offset table and table directory

	// Sort table names for consistent output
	tableNames := make([]string, 0, len(tables))
	for name := range tables {
		tableNames = append(tableNames, name)
	}
	sort.Strings(tableNames)

	// Write table directory
	tableOffsets := make(map[string]uint32, len(tableNames)) // PERF-192
	for _, name := range tableNames {
		data := tables[name]

		// Pad table name to 4 bytes
		var tag [4]byte
		copy(tag[:], name)
		for i := len(name); i < 4; i++ {
			tag[i] = ' '
		}

		checksum := calculateChecksum(data)
		length := uint32(len(data))

		buf = append(buf, tag[:]...)
		putU32BE(&buf, checksum)
		putU32BE(&buf, tableOffset)
		putU32BE(&buf, length)

		tableOffsets[name] = tableOffset

		// Align to 4-byte boundary
		paddedLen := (length + 3) &^ 3
		tableOffset += paddedLen
	}

	// Write table data
	for _, name := range tableNames {
		data := tables[name]
		buf = append(buf, data...)

		// Pad to 4-byte boundary
		padding := (4 - len(data)%4) % 4
		for i := 0; i < padding; i++ {
			buf = append(buf, 0)
		}
	}

	// Update head checksum adjustment
	headOffset := tableOffsets["head"]
	updateHeadChecksum(buf, headOffset)

	return buf, oldToNew, nil
}

// subsetHead generates the head table for the subset font
func subsetHead(font *TTFFont) []byte {
	headTable := font.Tables["head"]
	result := bytes.Clone(font.RawData[headTable.Offset : headTable.Offset+headTable.Length])

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
	result := bytes.Clone(font.RawData[hheaTable.Offset : hheaTable.Offset+hheaTable.Length])

	// Update numberOfHMetrics (last 2 bytes)
	binary.BigEndian.PutUint16(result[len(result)-2:], numGlyphs)

	return result
}

// subsetMaxp generates the maxp table with updated numGlyphs
func subsetMaxp(font *TTFFont, numGlyphs uint16) []byte {
	maxpTable := font.Tables["maxp"]
	result := bytes.Clone(font.RawData[maxpTable.Offset : maxpTable.Offset+maxpTable.Length])

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
	// PERF-3: pre-size once to max glyph payload (whole glyf table upper bound)
	glyphScratch := make([]byte, len(glyfData))
	newOffsets := make([]uint32, len(glyphs)+1)

	// Build old-to-new GID mapping for this subset
	oldToNewGID := make(map[uint16]uint16, len(glyphs)) // PERF-192
	for newIdx, oldGID := range glyphs {
		oldToNewGID[oldGID] = uint16(newIdx)
	}

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
			gs := glyphScratch[:length]
			copy(gs, glyfData[offset:offset+length])

			// Remap component GID references in composite glyphs
			remapCompositeGIDs(gs, oldToNewGID)

			newGlyf.Write(gs)

			// Pad to even boundary for short loca format
			if newGlyf.Len()%2 != 0 {
				newGlyf.WriteByte(0)
			}
		}
	}
	newOffsets[len(glyphs)] = uint32(newGlyf.Len())

	// Determine if we can use short loca format
	useShortLoca := newOffsets[len(glyphs)] <= 0xFFFF*2

	// Build new loca table (PERF-107: putU16BE/putU32BE)
	var newLoca []byte
	if useShortLoca {
		newLoca = make([]byte, 0, len(newOffsets)*2)
		for _, offset := range newOffsets {
			putU16BE(&newLoca, uint16(offset/2))
		}
	} else {
		newLoca = make([]byte, 0, len(newOffsets)*4)
		for _, offset := range newOffsets {
			putU32BE(&newLoca, offset)
		}
	}

	return newGlyf.Bytes(), newLoca, useShortLoca
}

// subsetHmtx generates the hmtx table for the subset
func subsetHmtx(font *TTFFont, glyphs []uint16) []byte {
	buf := make([]byte, 0, len(glyphs)*4)
	gw := font.GlyphWidths

	for _, glyphID := range glyphs {
		var width uint16
		if int(glyphID) < len(gw) {
			width = gw[glyphID]
		}
		putU16BE(&buf, width)
		putI16BE(&buf, 0)
	}

	return buf
}

// subsetCmap generates a format 4 cmap table with remapped glyph IDs
//
//nolint:gocyclo
func subsetCmap(font *TTFFont, oldToNew map[uint16]uint16) []byte {
	// Build character to new glyph ID mapping
	charToNewGlyph := make(map[uint16]uint16, len(font.CharToGlyph)) // PERF-192
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

	// Build format 4 subtable (PERF-107)
	// Header 14 + 4 arrays of segCount u16s + reservedPad = 16 + 8*segCount
	format4 := make([]byte, 0, 16+int(segCount)*8)

	// Header
	putU16BE(&format4, 4)          // format
	putU16BE(&format4, 0)          // length (placeholder)
	putU16BE(&format4, 0)          // language
	putU16BE(&format4, segCount*2) // segCountX2
	putU16BE(&format4, searchRange)
	putU16BE(&format4, entrySelector)
	putU16BE(&format4, rangeShift)

	// endCode array
	for _, seg := range segments {
		putU16BE(&format4, seg.endCode)
	}

	// reservedPad
	putU16BE(&format4, 0)

	// startCode array
	for _, seg := range segments {
		putU16BE(&format4, seg.startCode)
	}

	// idDelta array
	for _, seg := range segments {
		putI16BE(&format4, seg.idDelta)
	}

	// idRangeOffset array (all zeros for our simple mapping)
	for range segments {
		putU16BE(&format4, 0)
	}

	// Update length
	binary.BigEndian.PutUint16(format4[2:], uint16(len(format4)))

	// Build cmap table
	// cmap header: version + numTables + encoding record (8) + format4
	buf := make([]byte, 0, 12+len(format4))
	putU16BE(&buf, 0) // version
	putU16BE(&buf, 1) // numTables

	// Encoding record
	putU16BE(&buf, 3)  // platformID (Windows)
	putU16BE(&buf, 1)  // encodingID (Unicode BMP)
	putU32BE(&buf, 12) // offset to subtable

	// Write format 4 subtable
	buf = append(buf, format4...)

	return buf
}

// subsetPost generates a minimal post table
func subsetPost(font *TTFFont) []byte {
	buf := make([]byte, 0, 32)

	// Version 3.0 (no glyph names)
	putU32BE(&buf, 0x00030000)

	// italicAngle (16.16 fixed)
	italicAngleFixed := int32(font.ItalicAngle * 65536)
	putU32BE(&buf, uint32(italicAngleFixed))

	// underlinePosition, underlineThickness
	putI16BE(&buf, -100)
	putI16BE(&buf, 50)

	// isFixedPitch
	if font.IsFixedPitch {
		putU32BE(&buf, 1)
	} else {
		putU32BE(&buf, 0)
	}

	// minMemType42, maxMemType42, minMemType1, maxMemType1
	putU32BE(&buf, 0)
	putU32BE(&buf, 0)
	putU32BE(&buf, 0)
	putU32BE(&buf, 0)

	return buf
}

// subsetName generates a minimal name table
func subsetName(font *TTFFont) []byte {
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
	stringData.Grow(len(names) * 64)
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

	// Header (6) + records (12 each) + string data
	buf := make([]byte, 0, 6+len(records)*12+stringData.Len())

	// Write name table header
	putU16BE(&buf, 0)                         // format
	putU16BE(&buf, uint16(len(records)))      // count
	putU16BE(&buf, uint16(6+len(records)*12)) // stringOffset

	// Write name records
	for _, rec := range records {
		putU16BE(&buf, rec.platformID)
		putU16BE(&buf, rec.encodingID)
		putU16BE(&buf, rec.languageID)
		putU16BE(&buf, rec.nameID)
		putU16BE(&buf, rec.length)
		putU16BE(&buf, rec.offset)
	}

	// Write string data
	buf = append(buf, stringData.Bytes()...)

	return buf
}

// calculateChecksum calculates the checksum for a font table
func calculateChecksum(data []byte) uint32 {
	if len(data)%4 != 0 {
		dst := make([]byte, len(data)+(4-len(data)%4))
		copy(dst, data)
		data = dst
	}

	var sum uint32
	for i := 0; i < len(data); i += 4 {
		sum += binary.BigEndian.Uint32(data[i:])
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
	buf := GetCompressBuffer()
	defer PutCompressBuffer(buf)
	w := GetZlibWriter(buf)
	if _, err := w.Write(data); err != nil {
		_ = w.Close()
		PutZlibWriter(w)
		return nil, err
	}
	if err := w.Close(); err != nil {
		PutZlibWriter(w)
		return nil, err
	}
	PutZlibWriter(w)
	return append([]byte(nil), buf.Bytes()...), nil
}

// TrueType composite glyph flags (from OpenType spec)
const (
	compositeArg1And2AreWords   = 0x0001
	compositeMoreComponents     = 0x0020
	compositeWeHaveAScale       = 0x0008
	compositeWeHaveAnXYScale    = 0x0040
	compositeWeHaveATwoByTwo    = 0x0080
	compositeWeHaveInstructions = 0x0100
)

// getGlyphData returns the raw glyf data for a given glyph ID using the loca table.
func getGlyphData(font *TTFFont, glyphID uint16) []byte {
	glyfTable, hasGlyf := font.Tables["glyf"]
	locaTable, hasLoca := font.Tables["loca"]
	if !hasGlyf || !hasLoca {
		return nil
	}

	headTable := font.Tables["head"]
	isShortLoca := font.RawData[headTable.Offset+50] == 0 && font.RawData[headTable.Offset+51] == 0

	locaData := font.RawData[locaTable.Offset : locaTable.Offset+locaTable.Length]
	glyfData := font.RawData[glyfTable.Offset : glyfTable.Offset+glyfTable.Length]

	var offset, nextOffset uint32
	if isShortLoca {
		if int(glyphID)*2+2 > len(locaData) {
			return nil
		}
		offset = uint32(binary.BigEndian.Uint16(locaData[int(glyphID)*2:])) * 2
		nextOffset = uint32(binary.BigEndian.Uint16(locaData[int(glyphID)*2+2:])) * 2
	} else {
		if int(glyphID)*4+4 > len(locaData) {
			return nil
		}
		offset = binary.BigEndian.Uint32(locaData[int(glyphID)*4:])
		nextOffset = binary.BigEndian.Uint32(locaData[int(glyphID)*4+4:])
	}

	if nextOffset <= offset || offset >= uint32(len(glyfData)) {
		return nil
	}
	length := nextOffset - offset
	if offset+length > uint32(len(glyfData)) {
		length = uint32(len(glyfData)) - offset
	}
	return glyfData[offset : offset+length]
}

// getCompositeComponentGIDs extracts the component glyph IDs referenced by a composite glyph.
// Returns nil if the glyph is not composite.
func getCompositeComponentGIDs(data []byte) []uint16 {
	if len(data) < 10 {
		return nil
	}
	// numberOfContours is the first int16; negative means composite
	numContours := int16(binary.BigEndian.Uint16(data[0:2]))
	if numContours >= 0 {
		return nil // simple glyph
	}

	var components []uint16
	// Component data starts at offset 10 (after header: numContours + xMin + yMin + xMax + yMax)
	pos := 10
	for pos+4 <= len(data) {
		flags := binary.BigEndian.Uint16(data[pos:])
		glyphIndex := binary.BigEndian.Uint16(data[pos+2:])
		components = append(components, glyphIndex)
		pos += 4

		// Skip arguments (offsets/points)
		if flags&compositeArg1And2AreWords != 0 {
			pos += 4 // two int16 args
		} else {
			pos += 2 // two int8 args
		}

		// Skip transform data
		switch {
		case flags&compositeWeHaveAScale != 0:
			pos += 2 // one F2Dot14
		case flags&compositeWeHaveAnXYScale != 0:
			pos += 4 // two F2Dot14
		case flags&compositeWeHaveATwoByTwo != 0:
			pos += 8 // four F2Dot14
		}

		if flags&compositeMoreComponents == 0 {
			break
		}
	}
	return components
}

// addCompositeComponents expands the glyph set to include all component glyphs
// referenced by composite glyphs, recursively.
func addCompositeComponents(font *TTFFont, glyphSet map[uint16]bool) {
	// Iterate until no new glyphs are added (handles nested composites)
	for {
		added := false
		for gid := range glyphSet {
			data := getGlyphData(font, gid)
			if data == nil {
				continue
			}
			components := getCompositeComponentGIDs(data)
			for _, compGID := range components {
				if compGID < font.NumGlyphs && !glyphSet[compGID] {
					glyphSet[compGID] = true
					added = true
				}
			}
		}
		if !added {
			break
		}
	}
}

// remapCompositeGIDs rewrites component glyph ID references in composite glyph data
// to use the new GIDs from the subset mapping.
func remapCompositeGIDs(data []byte, oldToNew map[uint16]uint16) {
	if len(data) < 10 {
		return
	}
	numContours := int16(binary.BigEndian.Uint16(data[0:2]))
	if numContours >= 0 {
		return // simple glyph, nothing to remap
	}

	pos := 10
	for pos+4 <= len(data) {
		flags := binary.BigEndian.Uint16(data[pos:])
		oldGID := binary.BigEndian.Uint16(data[pos+2:])

		if newGID, ok := oldToNew[oldGID]; ok {
			binary.BigEndian.PutUint16(data[pos+2:], newGID)
		}
		_ = oldGID
		pos += 4

		if flags&compositeArg1And2AreWords != 0 {
			pos += 4
		} else {
			pos += 2
		}

		switch {
		case flags&compositeWeHaveAScale != 0:
			pos += 2
		case flags&compositeWeHaveAnXYScale != 0:
			pos += 4
		case flags&compositeWeHaveATwoByTwo != 0:
			pos += 8
		}

		if flags&compositeMoreComponents == 0 {
			break
		}
	}
}
