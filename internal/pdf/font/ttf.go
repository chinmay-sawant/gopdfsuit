package font

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
)

// TTFFont represents a parsed TrueType/OpenType font with all necessary data for PDF embedding
type TTFFont struct {
	// Font identification
	PostScriptName string // PostScript name for PDF
	FamilyName     string // Font family name
	FullName       string // Full font name
	Version        string // Font version

	// Font metrics (in font design units)
	UnitsPerEm   uint16
	Ascender     int16
	Descender    int16 // Negative value
	LineGap      int16
	CapHeight    int16
	XHeight      int16
	StemV        int16
	ItalicAngle  float64
	IsFixedPitch bool
	IsBold       bool
	IsItalic     bool

	// Bounding box
	BBox [4]int16 // xMin, yMin, xMax, yMax

	// Glyph data
	NumGlyphs   uint16
	GlyphWidths []uint16        // Width for each glyph ID
	CharToGlyph map[rune]uint16 // Unicode to glyph ID mapping (cmap)
	GlyphToChar map[uint16]rune // Reverse mapping for ToUnicode

	// Raw font data for embedding
	RawData []byte

	// Table offsets for subsetting
	Tables map[string]TableEntry
}

// TableEntry represents a font table's location in the file
type TableEntry struct {
	Tag      string
	Checksum uint32
	Offset   uint32
	Length   uint32
}

// LoadTTFFromFile loads and parses a TTF/OTF font from a file path
func LoadTTFFromFile(path string) (*TTFFont, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Join(errors.New("failed to read font file"), err)
	}
	return ParseTTF(data)
}

// LoadTTFFromData loads and parses a TTF/OTF font from raw bytes
func LoadTTFFromData(data []byte) (*TTFFont, error) {
	return ParseTTF(data)
}

// ParseTTF parses TrueType/OpenType font data
func ParseTTF(data []byte) (*TTFFont, error) {
	if len(data) < 12 {
		return nil, errors.New("font data too short")
	}

	off := 0
	sfntVersion, off, err := readU32BE(data, off)
	if err != nil {
		return nil, errors.Join(errors.New("failed to read sfntVersion"), err)
	}

	// Check for valid font format
	// 0x00010000 = TrueType, 0x4F54544F = 'OTTO' (OpenType/CFF)
	if sfntVersion != 0x00010000 && sfntVersion != 0x4F54544F {
		return nil, errors.New("unsupported font format: 0x" + strconv.FormatUint(uint64(sfntVersion), 16))
	}

	numTables, off, err := readU16BE(data, off)
	if err != nil {
		return nil, errors.Join(errors.New("failed to read numTables"), err)
	}
	// Skip searchRange, entrySelector, rangeShift
	off += 6
	if off > len(data) {
		return nil, errors.Join(errors.New("failed to seek"), io.ErrUnexpectedEOF)
	}

	font := &TTFFont{
		RawData:     data,
		Tables:      make(map[string]TableEntry, numTables), // PERF-192
		CharToGlyph: make(map[rune]uint16, 256),             // PERF-192: typical cmap size grows as needed
		GlyphToChar: make(map[uint16]rune, 256),             // PERF-192
	}

	// Read table directory
	for i := uint16(0); i < numTables; i++ {
		if off+16 > len(data) {
			return nil, errors.Join(errors.New("failed to read tag"), io.ErrUnexpectedEOF)
		}
		tag := string(data[off : off+4])
		off += 4
		var entry TableEntry
		entry.Tag = tag
		entry.Checksum, off, err = readU32BE(data, off)
		if err != nil {
			return nil, errors.Join(errors.New("failed to read checksum"), err)
		}
		entry.Offset, off, err = readU32BE(data, off)
		if err != nil {
			return nil, errors.Join(errors.New("failed to read offset"), err)
		}
		entry.Length, off, err = readU32BE(data, off)
		if err != nil {
			return nil, errors.Join(errors.New("failed to read length"), err)
		}
		font.Tables[entry.Tag] = entry
	}

	// Parse required tables
	if err := font.parseHead(data); err != nil {
		return nil, errors.Join(errors.New("failed to parse 'head' table"), err)
	}

	if err := font.parseHhea(data); err != nil {
		return nil, errors.Join(errors.New("failed to parse 'hhea' table"), err)
	}

	if err := font.parseMaxp(data); err != nil {
		return nil, errors.Join(errors.New("failed to parse 'maxp' table"), err)
	}

	if err := font.parseHmtx(data); err != nil {
		return nil, errors.Join(errors.New("failed to parse 'hmtx' table"), err)
	}

	if err := font.parseCmap(data); err != nil {
		return nil, errors.Join(errors.New("failed to parse 'cmap' table"), err)
	}

	if err := font.parseName(data); err != nil {
		// Name table is optional for basic functionality
		font.PostScriptName = "UnknownFont" //nolint:goconst
		font.FamilyName = "Unknown"
		font.FullName = "Unknown Font"
	}

	if err := font.parseOS2(data); err != nil {
		// OS/2 table is optional, set defaults
		font.CapHeight = int16(float64(font.UnitsPerEm) * 0.7)
		font.XHeight = int16(float64(font.UnitsPerEm) * 0.5)
		font.StemV = 80
	}

	if err := font.parsePost(data); err != nil {
		// post table is optional
		font.ItalicAngle = 0
		font.IsFixedPitch = false
	}

	return font, nil
}

// parseHead parses the 'head' table for basic font metrics
func (f *TTFFont) parseHead(data []byte) error {
	table, ok := f.Tables["head"]
	if !ok {
		return errors.New("missing 'head' table")
	}

	if table.Offset+54 > uint32(len(data)) {
		return errors.New("head table truncated")
	}

	base := int(table.Offset)
	// Skip version, fontRevision, checksumAdjustment, magicNumber, flags (18 bytes)
	off := base + 18

	units, off, err := readU16BE(data, off)
	if err != nil {
		return errors.Join(errors.New("failed to read UnitsPerEm"), err)
	}
	f.UnitsPerEm = units

	// Skip created, modified dates (16 bytes)
	off += 16

	xMin, off, err := readI16BE(data, off)
	if err != nil {
		return errors.Join(errors.New("failed to read xMin"), err)
	}
	yMin, off, err := readI16BE(data, off)
	if err != nil {
		return errors.Join(errors.New("failed to read yMin"), err)
	}
	xMax, off, err := readI16BE(data, off)
	if err != nil {
		return errors.Join(errors.New("failed to read xMax"), err)
	}
	yMax, _, err := readI16BE(data, off)
	if err != nil {
		return errors.Join(errors.New("failed to read yMax"), err)
	}
	f.BBox[0], f.BBox[1], f.BBox[2], f.BBox[3] = xMin, yMin, xMax, yMax

	return nil
}

// parseHhea parses the 'hhea' table for horizontal metrics
func (f *TTFFont) parseHhea(data []byte) error {
	table, ok := f.Tables["hhea"]
	if !ok {
		return errors.New("missing 'hhea' table")
	}

	if table.Offset+36 > uint32(len(data)) {
		return errors.New("hhea table truncated")
	}

	// Skip version (4 bytes)
	off := int(table.Offset) + 4

	asc, off, err := readI16BE(data, off)
	if err != nil {
		return errors.Join(errors.New("failed to read Ascender"), err)
	}
	desc, off, err := readI16BE(data, off)
	if err != nil {
		return errors.Join(errors.New("failed to read Descender"), err)
	}
	gap, _, err := readI16BE(data, off)
	if err != nil {
		return errors.Join(errors.New("failed to read LineGap"), err)
	}
	f.Ascender, f.Descender, f.LineGap = asc, desc, gap

	return nil
}

// parseMaxp parses the 'maxp' table for glyph count
func (f *TTFFont) parseMaxp(data []byte) error {
	table, ok := f.Tables["maxp"]
	if !ok {
		return errors.New("missing 'maxp' table")
	}

	if table.Offset+6 > uint32(len(data)) {
		return errors.New("maxp table truncated")
	}

	// Skip version (4 bytes)
	off := int(table.Offset) + 4
	num, _, err := readU16BE(data, off)
	if err != nil {
		return errors.Join(errors.New("failed to read NumGlyphs"), err)
	}
	f.NumGlyphs = num

	return nil
}

// parseHmtx parses the 'hmtx' table for glyph widths
func (f *TTFFont) parseHmtx(data []byte) error {
	table, ok := f.Tables["hmtx"]
	if !ok {
		return errors.New("missing 'hmtx' table")
	}

	// Need to get numberOfHMetrics from hhea table
	hheaTable := f.Tables["hhea"]
	if hheaTable.Offset+36 > uint32(len(data)) {
		return errors.New("hhea table truncated")
	}

	numberOfHMetrics, _, err := readU16BE(data, int(hheaTable.Offset)+34)
	if err != nil {
		return errors.Join(errors.New("failed to read numberOfHMetrics"), err)
	}

	// Parse hmtx table: each longHorMetric is advanceWidth (u16) + lsb (i16) = 4 bytes
	f.GlyphWidths = make([]uint16, f.NumGlyphs)
	off := int(table.Offset)

	var lastWidth uint16
	for i := uint16(0); i < numberOfHMetrics; i++ {
		if off+4 > len(data) {
			errPrefix := "failed to read GlyphWidths["
			errStr := make([]byte, 0, len(errPrefix)+5+1)
			errStr = append(errStr, errPrefix...)
			errStr = strconv.AppendInt(errStr, int64(i), 10) // PERF-15: AppendInt avoids strconv.Itoa alloc
			errStr = append(errStr, ']')
			return errors.Join(errors.New(string(errStr)), io.ErrUnexpectedEOF)
		}
		f.GlyphWidths[i] = binary.BigEndian.Uint16(data[off:]) // PERF-109: data[off:] avoids recomputing off+2 each loop
		// skip lsb (2 bytes)
		off += 4
		lastWidth = f.GlyphWidths[i]
	}

	// Remaining glyphs use the last advanceWidth
	for i := numberOfHMetrics; i < f.NumGlyphs; i++ {
		f.GlyphWidths[i] = lastWidth
	}

	return nil
}

// parseCmap parses the 'cmap' table for character to glyph mapping
func (f *TTFFont) parseCmap(data []byte) error {
	table, ok := f.Tables["cmap"]
	if !ok {
		return errors.New("missing 'cmap' table")
	}

	base := int(table.Offset)
	if base+4 > len(data) {
		return errors.Join(errors.New("failed to seek"), io.ErrUnexpectedEOF)
	}

	// Skip version (2 bytes)
	off := base + 2
	numTables, off, err := readU16BE(data, off)
	if err != nil {
		return errors.Join(errors.New("failed to read numTables"), err)
	}

	// Find best cmap subtable (prefer format 4 for BMP, format 12 for full Unicode)
	var bestOffset uint32
	var bestFormat uint16

	for i := uint16(0); i < numTables; i++ {
		if off+8 > len(data) {
			return errors.Join(errors.New("failed to read platformID"), io.ErrUnexpectedEOF)
		}
		platformID := binary.BigEndian.Uint16(data[off : off+2])
		encodingID := binary.BigEndian.Uint16(data[off+2 : off+4])
		subOffset := binary.BigEndian.Uint32(data[off+4 : off+8])
		off += 8

		// Windows Unicode BMP (format 4) or Full Unicode (format 12)
		if platformID == 3 && (encodingID == 1 || encodingID == 10) {
			formatOff := base + int(subOffset)
			format, _, ferr := readU16BE(data, formatOff)
			if ferr != nil {
				return errors.Join(errors.New("failed to read format"), ferr)
			}
			// Prefer format 12 over format 4
			if format == 12 || (format == 4 && bestFormat != 12) {
				bestOffset = subOffset
				bestFormat = format
			}
		}

		// Unicode platform
		if platformID == 0 {
			formatOff := base + int(subOffset)
			format, _, ferr := readU16BE(data, formatOff)
			if ferr != nil {
				return errors.Join(errors.New("failed to read format"), ferr)
			}
			if format == 12 || (format == 4 && bestFormat != 12) {
				bestOffset = subOffset
				bestFormat = format
			}
		}
	}

	if bestOffset == 0 {
		return errors.New("no suitable cmap subtable found")
	}

	// Parse the selected subtable
	switch bestFormat {
	case 4:
		return f.parseCmapFormat4(data, table.Offset+bestOffset)
	case 12:
		return f.parseCmapFormat12(data, table.Offset+bestOffset)
	default:
		return errors.New("unsupported cmap format: " + strconv.Itoa(int(bestFormat)))
	}
}

// parseCmapFormat4 parses a format 4 cmap subtable (BMP characters)
func (f *TTFFont) parseCmapFormat4(data []byte, offset uint32) error {
	base := int(offset)
	if base+14 > len(data) {
		return errors.Join(errors.New("failed to seek"), io.ErrUnexpectedEOF)
	}

	// Skip format (2), read length (2), skip language (2), read segCountX2 (2)
	// skip searchRange, entrySelector, rangeShift (6) = total header 14 after format
	off := base + 2 // skip format
	length, off, err := readU16BE(data, off)
	if err != nil {
		return errors.Join(errors.New("failed to read length"), err)
	}
	_ = length
	off += 2 // skip language

	segCountX2, off, err := readU16BE(data, off)
	if err != nil {
		return errors.Join(errors.New("failed to read segCountX2"), err)
	}
	segCount := segCountX2 / 2
	off += 6 // skip searchRange, entrySelector, rangeShift

	// Read endCode array
	endCodes := make([]uint16, segCount)
	for i := uint16(0); i < segCount; i++ {
		endCodes[i], off, err = readU16BE(data, off)
		if err != nil {
			return errors.Join(errors.New("failed to read endCodes"), err)
		}
	}

	off += 2 // reservedPad
	if off > len(data) {
		return errors.Join(errors.New("failed to seek"), io.ErrUnexpectedEOF)
	}

	// Read startCode array
	startCodes := make([]uint16, segCount)
	for i := uint16(0); i < segCount; i++ {
		startCodes[i], off, err = readU16BE(data, off)
		if err != nil {
			return errors.Join(errors.New("failed to read startCodes"), err)
		}
	}

	// Read idDelta array
	idDeltas := make([]int16, segCount)
	for i := uint16(0); i < segCount; i++ {
		idDeltas[i], off, err = readI16BE(data, off)
		if err != nil {
			return errors.Join(errors.New("failed to read idDeltas"), err)
		}
	}

	// Read idRangeOffset array — position relative to subtable start
	idRangeOffsetPos := off - base
	idRangeOffsets := make([]uint16, segCount)
	for i := uint16(0); i < segCount; i++ {
		idRangeOffsets[i], off, err = readU16BE(data, off)
		if err != nil {
			return errors.Join(errors.New("failed to read idRangeOffsets"), err)
		}
	}

	subtableLen := len(data) - base
	// Build character to glyph mapping
	for i := uint16(0); i < segCount; i++ {
		if startCodes[i] == 0xFFFF {
			break
		}

		for c := startCodes[i]; c <= endCodes[i]; c++ {
			var glyphID uint16

			if idRangeOffsets[i] == 0 {
				glyphID = uint16(int32(c) + int32(idDeltas[i]))
			} else {
				// Calculate offset into glyph ID array (relative to subtable start)
				glyphIndexOffset := idRangeOffsetPos + int(i)*2 + int(idRangeOffsets[i]) + int(c-startCodes[i])*2
				if glyphIndexOffset+2 <= subtableLen {
					glyphID = binary.BigEndian.Uint16(data[base+glyphIndexOffset : base+glyphIndexOffset+2])
					if glyphID != 0 {
						glyphID = uint16(int32(glyphID) + int32(idDeltas[i]))
					}
				}
			}

			if glyphID != 0 && glyphID < f.NumGlyphs {
				f.CharToGlyph[rune(c)] = glyphID
				f.GlyphToChar[glyphID] = rune(c)
			}
		}
	}

	return nil
}

// parseCmapFormat12 parses a format 12 cmap subtable (full Unicode)
func (f *TTFFont) parseCmapFormat12(data []byte, offset uint32) error {
	// Skip format(2)+reserved(2)+length(4)+language(4) = 12 bytes
	off := int(offset) + 12
	numGroups, off, err := readU32BE(data, off)
	if err != nil {
		return errors.Join(errors.New("failed to read numGroups"), err)
	}

	for i := uint32(0); i < numGroups; i++ {
		if off+12 > len(data) {
			return errors.Join(errors.New("failed to read startCharCode"), io.ErrUnexpectedEOF)
		}
		startCharCode := binary.BigEndian.Uint32(data[off : off+4])
		endCharCode := binary.BigEndian.Uint32(data[off+4 : off+8])
		startGlyphID := binary.BigEndian.Uint32(data[off+8 : off+12])
		off += 12

		for c := startCharCode; c <= endCharCode; c++ {
			glyphID := uint16(startGlyphID + (c - startCharCode))
			if glyphID < f.NumGlyphs {
				f.CharToGlyph[rune(c)] = glyphID
				f.GlyphToChar[glyphID] = rune(c)
			}
		}
	}

	return nil
}

// parseName parses the 'name' table for font names
//
//nolint:gocyclo
func (f *TTFFont) parseName(data []byte) error {
	table, ok := f.Tables["name"]
	if !ok {
		return errors.New("missing 'name' table")
	}

	base := int(table.Offset)
	if base+6 > len(data) {
		return errors.Join(errors.New("failed to seek"), io.ErrUnexpectedEOF)
	}

	// Skip format (2 bytes)
	off := base + 2
	count, off, err := readU16BE(data, off)
	if err != nil {
		return errors.Join(errors.New("failed to read count"), err)
	}
	stringOffset, off, err := readU16BE(data, off)
	if err != nil {
		return errors.Join(errors.New("failed to read stringOffset"), err)
	}

	storageOffset := table.Offset + uint32(stringOffset)

	for i := uint16(0); i < count; i++ {
		if off+12 > len(data) {
			return errors.Join(errors.New("failed to read platformID"), io.ErrUnexpectedEOF)
		}
		platformID := binary.BigEndian.Uint16(data[off : off+2])
		encodingID := binary.BigEndian.Uint16(data[off+2 : off+4])
		// languageID at off+4 — unused for extraction
		nameID := binary.BigEndian.Uint16(data[off+6 : off+8])
		length := binary.BigEndian.Uint16(data[off+8 : off+10])
		nameOffset := binary.BigEndian.Uint16(data[off+10 : off+12])
		off += 12

		// Extract string (prefer platform 3 = Windows, encoding 1 = Unicode BMP)
		if platformID == 3 && encodingID == 1 {
			strStart := storageOffset + uint32(nameOffset)
			strEnd := strStart + uint32(length)
			if strEnd <= uint32(len(data)) {
				// Convert UTF-16BE to string
				str := decodeUTF16BE(data[strStart:strEnd])
				switch nameID {
				case 1: // Font Family
					f.FamilyName = str
				case 4: // Full Name
					f.FullName = str
				case 6: // PostScript Name
					f.PostScriptName = str
				case 5: // Version
					f.Version = str
				}
			}
		}

		// Fallback to platform 1 (Macintosh) if needed
		if platformID == 1 && encodingID == 0 && f.PostScriptName == "" {
			strStart := storageOffset + uint32(nameOffset)
			strEnd := strStart + uint32(length)
			if strEnd <= uint32(len(data)) {
				str := string(data[strStart:strEnd])
				switch nameID {
				case 1:
					if f.FamilyName == "" {
						f.FamilyName = str
					}
				case 4:
					if f.FullName == "" {
						f.FullName = str
					}
				case 6:
					if f.PostScriptName == "" {
						f.PostScriptName = str
					}
				}
			}
		}
	}

	// Set defaults if not found
	if f.PostScriptName == "" {
		if f.FamilyName != "" {
			f.PostScriptName = sanitizePostScriptName(f.FamilyName)
		} else {
			f.PostScriptName = "UnknownFont"
		}
	}

	return nil
}

// parseOS2 parses the 'OS/2' table for additional metrics
func (f *TTFFont) parseOS2(data []byte) error {
	table, ok := f.Tables["OS/2"]
	if !ok {
		return errors.New("missing 'OS/2' table")
	}

	if table.Length < 78 {
		return errors.New("OS/2 table too short")
	}

	base := int(table.Offset)
	version, off, err := readU16BE(data, base)
	if err != nil {
		return errors.Join(errors.New("failed to read version"), err)
	}

	// Skip xAvgCharWidth (2)
	off += 2
	usWeightClass, off, err := readU16BE(data, off)
	if err != nil {
		return errors.Join(errors.New("failed to read usWeightClass"), err)
	}
	f.IsBold = usWeightClass >= 700

	// Match prior Seek(60) after usWeightClass, then read fsSelection field.
	off = base + 6 + 60
	fsSelection, off, err := readU16BE(data, off)
	if err != nil {
		return errors.Join(errors.New("failed to read fsSelection"), err)
	}
	f.IsItalic = (fsSelection & 0x0001) != 0

	// Skip 4 bytes; for v2+ land on sxHeight (offset 86) / sCapHeight (88).
	off += 4

	if version >= 2 && table.Length >= 96 {
		off += 14 // prior Seek(16) then Seek(-2)
		xh, off2, err := readI16BE(data, off)
		if err != nil {
			return errors.Join(errors.New("failed to read XHeight"), err)
		}
		ch, _, err := readI16BE(data, off2)
		if err != nil {
			return errors.Join(errors.New("failed to read CapHeight"), err)
		}
		f.XHeight = xh
		f.CapHeight = ch
	} else {
		// Estimate from ascender
		f.CapHeight = int16(float64(f.Ascender) * 0.7)
		f.XHeight = int16(float64(f.Ascender) * 0.5)
	}

	// Estimate StemV from weight class
	f.StemV = int16(50 + (usWeightClass-400)/10)
	if f.StemV < 50 {
		f.StemV = 50
	}
	if f.StemV > 200 {
		f.StemV = 200
	}

	return nil
}

// parsePost parses the 'post' table for PostScript data
func (f *TTFFont) parsePost(data []byte) error {
	table, ok := f.Tables["post"]
	if !ok {
		return errors.New("missing 'post' table")
	}

	if table.Length < 32 {
		return errors.New("post table too short")
	}

	base := int(table.Offset)
	// Skip version (4 bytes)
	off := base + 4

	// Read italic angle as fixed-point (16.16)
	if off+4 > len(data) {
		return errors.Join(errors.New("failed to read italicAngleFixed"), io.ErrUnexpectedEOF)
	}
	italicAngleFixed := int32(binary.BigEndian.Uint32(data[off : off+4]))
	off += 4
	f.ItalicAngle = float64(italicAngleFixed) / 65536.0

	// Skip underlinePosition, underlineThickness (4 bytes)
	off += 4
	if off+4 > len(data) {
		return errors.Join(errors.New("failed to read isFixedPitch"), io.ErrUnexpectedEOF)
	}
	isFixedPitch := binary.BigEndian.Uint32(data[off : off+4])
	f.IsFixedPitch = isFixedPitch != 0

	return nil
}

// GetGlyphWidth returns the width of a glyph in font design units
func (f *TTFFont) GetGlyphWidth(glyphID uint16) uint16 {
	if int(glyphID) < len(f.GlyphWidths) {
		return f.GlyphWidths[glyphID]
	}
	return 0
}

// GetCharWidth returns the width of a character in font design units
func (f *TTFFont) GetCharWidth(char rune) uint16 {
	if glyphID, ok := f.CharToGlyph[char]; ok {
		return f.GetGlyphWidth(glyphID)
	}
	// Return .notdef glyph width (glyph 0)
	return f.GetGlyphWidth(0)
}

// GetCharWidthScaled returns the width of a character scaled to PDF units (1/1000 em)
func (f *TTFFont) GetCharWidthScaled(char rune) int {
	width := f.GetCharWidth(char)
	return int(math.Round(float64(width) * 1000.0 / float64(f.UnitsPerEm)))
}

// GetUsedGlyphs returns a sorted list of glyph IDs used by the given text
func (f *TTFFont) GetUsedGlyphs(text string) []uint16 {
	glyphSet := make(map[uint16]bool, len(text)+1) // PERF-192
	glyphSet[0] = true                             // Always include .notdef

	for _, char := range text {
		if glyphID, ok := f.CharToGlyph[char]; ok {
			glyphSet[glyphID] = true
		}
	}

	glyphs := make([]uint16, 0, len(glyphSet))
	for glyph := range glyphSet {
		glyphs = append(glyphs, glyph)
	}

	sort.Slice(glyphs, func(i, j int) bool {
		return glyphs[i] < glyphs[j]
	})

	return glyphs
}

// GetPDFFlags returns the PDF font descriptor Flags value
func (f *TTFFont) GetPDFFlags() int {
	flags := 0

	if f.IsFixedPitch {
		flags |= 1 // FixedPitch
	}
	// flags |= 2 // Serif (would need to detect)
	// flags |= 4 // Symbolic (for symbol fonts)
	// flags |= 8 // Script (for script fonts)
	flags |= 32 // Nonsymbolic (standard Latin characters)
	if f.IsItalic {
		flags |= 64 // Italic
	}
	// flags |= 65536 // AllCap
	// flags |= 131072 // SmallCap
	if f.IsBold {
		flags |= 262144 // ForceBold
	}

	return flags
}

// Helper functions

func decodeUTF16BE(data []byte) string {
	if len(data)%2 != 0 {
		return ""
	}

	runes := make([]rune, 0, len(data)/2)
	for i := 0; i < len(data); i += 2 {
		r := rune(data[i])<<8 | rune(data[i+1])
		// Handle surrogate pairs
		if r >= 0xD800 && r <= 0xDBFF && i+2 < len(data) {
			low := rune(data[i+2])<<8 | rune(data[i+3])
			if low >= 0xDC00 && low <= 0xDFFF {
				r = 0x10000 + (r-0xD800)<<10 + (low - 0xDC00)
				i += 2
			}
		}
		runes = append(runes, r)
	}

	return string(runes)
}

func sanitizePostScriptName(name string) string {
	result := make([]byte, 0, len(name))
	for _, c := range name {
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			result = append(result, byte(c))
		}
	}
	if len(result) == 0 {
		return "UnknownFont"
	}
	return string(result)
}
