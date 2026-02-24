package font

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
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
		return nil, fmt.Errorf("failed to read font file: %w", err)
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

	font := &TTFFont{
		RawData:     data,
		Tables:      make(map[string]TableEntry),
		CharToGlyph: make(map[rune]uint16),
		GlyphToChar: make(map[uint16]rune),
	}

	r := bytes.NewReader(data)

	// Read offset table
	var sfntVersion uint32
	if err := binary.Read(r, binary.BigEndian, &sfntVersion); err != nil {
		return nil, fmt.Errorf("failed to read sfntVersion: %w", err)
	}

	// Check for valid font format
	// 0x00010000 = TrueType, 0x4F54544F = 'OTTO' (OpenType/CFF)
	if sfntVersion != 0x00010000 && sfntVersion != 0x4F54544F {
		return nil, fmt.Errorf("unsupported font format: 0x%08X", sfntVersion)
	}

	var numTables uint16
	if err := binary.Read(r, binary.BigEndian, &numTables); err != nil {
		return nil, fmt.Errorf("failed to read numTables: %w", err)
	}
	if _, err := r.Seek(6, io.SeekCurrent); err != nil { // Skip searchRange, entrySelector, rangeShift
		return nil, fmt.Errorf("failed to seek: %w", err)
	}

	// Read table directory
	for i := uint16(0); i < numTables; i++ {
		var tag [4]byte
		var entry TableEntry
		if _, err := r.Read(tag[:]); err != nil {
			return nil, fmt.Errorf("failed to read tag: %w", err)
		}
		if err := binary.Read(r, binary.BigEndian, &entry.Checksum); err != nil {
			return nil, fmt.Errorf("failed to read checksum: %w", err)
		}
		if err := binary.Read(r, binary.BigEndian, &entry.Offset); err != nil {
			return nil, fmt.Errorf("failed to read offset: %w", err)
		}
		if err := binary.Read(r, binary.BigEndian, &entry.Length); err != nil {
			return nil, fmt.Errorf("failed to read length: %w", err)
		}
		entry.Tag = string(tag[:])
		font.Tables[entry.Tag] = entry
	}

	// Parse required tables
	if err := font.parseHead(data); err != nil {
		return nil, fmt.Errorf("failed to parse 'head' table: %w", err)
	}

	if err := font.parseHhea(data); err != nil {
		return nil, fmt.Errorf("failed to parse 'hhea' table: %w", err)
	}

	if err := font.parseMaxp(data); err != nil {
		return nil, fmt.Errorf("failed to parse 'maxp' table: %w", err)
	}

	if err := font.parseHmtx(data); err != nil {
		return nil, fmt.Errorf("failed to parse 'hmtx' table: %w", err)
	}

	if err := font.parseCmap(data); err != nil {
		return nil, fmt.Errorf("failed to parse 'cmap' table: %w", err)
	}

	if err := font.parseName(data); err != nil {
		// Name table is optional for basic functionality
		font.PostScriptName = "UnknownFont"
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

	r := bytes.NewReader(data[table.Offset:])
	if _, err := r.Seek(18, io.SeekCurrent); err != nil { // Skip version, fontRevision, checksumAdjustment, magicNumber, flags
		return errors.New("failed to seek in head table")
	}

	if err := binary.Read(r, binary.BigEndian, &f.UnitsPerEm); err != nil {
		return fmt.Errorf("failed to read UnitsPerEm: %w", err)
	}

	if _, err := r.Seek(16, io.SeekCurrent); err != nil { // Skip created, modified dates
		return errors.New("failed to seek in head table")
	}

	if err := binary.Read(r, binary.BigEndian, &f.BBox[0]); err != nil {
		return fmt.Errorf("failed to read xMin: %w", err)
	}
	if err := binary.Read(r, binary.BigEndian, &f.BBox[1]); err != nil {
		return fmt.Errorf("failed to read yMin: %w", err)
	}
	if err := binary.Read(r, binary.BigEndian, &f.BBox[2]); err != nil {
		return fmt.Errorf("failed to read xMax: %w", err)
	}
	if err := binary.Read(r, binary.BigEndian, &f.BBox[3]); err != nil {
		return fmt.Errorf("failed to read yMax: %w", err)
	}

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

	r := bytes.NewReader(data[table.Offset:])
	if _, err := r.Seek(4, io.SeekCurrent); err != nil {
		return fmt.Errorf("failed to seek: %w", err)
	}

	if err := binary.Read(r, binary.BigEndian, &f.Ascender); err != nil {
		return fmt.Errorf("failed to read Ascender: %w", err)
	}
	if err := binary.Read(r, binary.BigEndian, &f.Descender); err != nil {
		return fmt.Errorf("failed to read Descender: %w", err)
	}
	if err := binary.Read(r, binary.BigEndian, &f.LineGap); err != nil {
		return fmt.Errorf("failed to read LineGap: %w", err)
	}

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

	r := bytes.NewReader(data[table.Offset:])
	if _, err := r.Seek(4, io.SeekCurrent); err != nil {
		return fmt.Errorf("failed to seek: %w", err)
	}

	if err := binary.Read(r, binary.BigEndian, &f.NumGlyphs); err != nil {
		return fmt.Errorf("failed to read NumGlyphs: %w", err)
	}

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

	var numberOfHMetrics uint16
	r := bytes.NewReader(data[hheaTable.Offset+34:])
	if err := binary.Read(r, binary.BigEndian, &numberOfHMetrics); err != nil {
		return fmt.Errorf("failed to read numberOfHMetrics: %w", err)
	}

	// Parse hmtx table
	f.GlyphWidths = make([]uint16, f.NumGlyphs)
	r = bytes.NewReader(data[table.Offset:])

	var lastWidth uint16
	for i := uint16(0); i < numberOfHMetrics; i++ {
		if err := binary.Read(r, binary.BigEndian, &f.GlyphWidths[i]); err != nil {
			return fmt.Errorf("failed to read GlyphWidths[%d]: %w", i, err)
		}
		if _, err := r.Seek(2, io.SeekCurrent); err != nil {
			return fmt.Errorf("failed to seek: %w", err)
		}
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

	r := bytes.NewReader(data[table.Offset:])
	if _, err := r.Seek(2, io.SeekCurrent); err != nil {
		return fmt.Errorf("failed to seek: %w", err)
	}

	var numTables uint16
	if err := binary.Read(r, binary.BigEndian, &numTables); err != nil {
		return fmt.Errorf("failed to read numTables: %w", err)
	}

	// Find best cmap subtable (prefer format 4 for BMP, format 12 for full Unicode)
	var bestOffset uint32
	var bestFormat uint16

	for i := uint16(0); i < numTables; i++ {
		var platformID, encodingID uint16
		var offset uint32
		if err := binary.Read(r, binary.BigEndian, &platformID); err != nil {
			return fmt.Errorf("failed to read platformID: %w", err)
		}
		if err := binary.Read(r, binary.BigEndian, &encodingID); err != nil {
			return fmt.Errorf("failed to read encodingID: %w", err)
		}
		if err := binary.Read(r, binary.BigEndian, &offset); err != nil {
			return fmt.Errorf("failed to read offset: %w", err)
		}

		// Windows Unicode BMP (format 4) or Full Unicode (format 12)
		if platformID == 3 && (encodingID == 1 || encodingID == 10) {
			// Check format
			formatReader := bytes.NewReader(data[table.Offset+offset:])
			var format uint16
			if err := binary.Read(formatReader, binary.BigEndian, &format); err != nil {
				return fmt.Errorf("failed to read format: %w", err)
			}

			// Prefer format 12 over format 4
			if format == 12 || (format == 4 && bestFormat != 12) {
				bestOffset = offset
				bestFormat = format
			}
		}

		// Unicode platform
		if platformID == 0 {
			formatReader := bytes.NewReader(data[table.Offset+offset:])
			var format uint16
			if err := binary.Read(formatReader, binary.BigEndian, &format); err != nil {
				return fmt.Errorf("failed to read format: %w", err)
			}

			if format == 12 || (format == 4 && bestFormat != 12) {
				bestOffset = offset
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
		return fmt.Errorf("unsupported cmap format: %d", bestFormat)
	}
}

// parseCmapFormat4 parses a format 4 cmap subtable (BMP characters)
func (f *TTFFont) parseCmapFormat4(data []byte, offset uint32) error {
	r := bytes.NewReader(data[offset:])
	if _, err := r.Seek(2, io.SeekCurrent); err != nil {
		return fmt.Errorf("failed to seek: %w", err)
	}

	var length uint16
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return fmt.Errorf("failed to read length: %w", err)
	}
	if _, err := r.Seek(2, io.SeekCurrent); err != nil {
		return fmt.Errorf("failed to seek: %w", err)
	}

	var segCountX2 uint16
	if err := binary.Read(r, binary.BigEndian, &segCountX2); err != nil {
		return fmt.Errorf("failed to read segCountX2: %w", err)
	}
	segCount := segCountX2 / 2

	if _, err := r.Seek(6, io.SeekCurrent); err != nil {
		return fmt.Errorf("failed to seek: %w", err)
	}

	// Read endCode array
	endCodes := make([]uint16, segCount)
	for i := uint16(0); i < segCount; i++ {
		if err := binary.Read(r, binary.BigEndian, &endCodes[i]); err != nil {
			return fmt.Errorf("failed to read endCodes: %w", err)
		}
	}

	if _, err := r.Seek(2, io.SeekCurrent); err != nil {
		return fmt.Errorf("failed to seek: %w", err)
	}

	// Read startCode array
	startCodes := make([]uint16, segCount)
	for i := uint16(0); i < segCount; i++ {
		if err := binary.Read(r, binary.BigEndian, &startCodes[i]); err != nil {
			return fmt.Errorf("failed to read startCodes: %w", err)
		}
	}

	// Read idDelta array
	idDeltas := make([]int16, segCount)
	for i := uint16(0); i < segCount; i++ {
		if err := binary.Read(r, binary.BigEndian, &idDeltas[i]); err != nil {
			return fmt.Errorf("failed to read idDeltas: %w", err)
		}
	}

	// Read idRangeOffset array
	idRangeOffsetPos, _ := r.Seek(0, io.SeekCurrent)
	idRangeOffsets := make([]uint16, segCount)
	for i := uint16(0); i < segCount; i++ {
		if err := binary.Read(r, binary.BigEndian, &idRangeOffsets[i]); err != nil {
			return fmt.Errorf("failed to read idRangeOffsets: %w", err)
		}
	}

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
				// Calculate offset into glyph ID array
				glyphIndexOffset := idRangeOffsetPos + int64(i)*2 + int64(idRangeOffsets[i]) + int64(c-startCodes[i])*2
				if glyphIndexOffset+2 <= int64(len(data[offset:])) {
					glyphReader := bytes.NewReader(data[offset+uint32(glyphIndexOffset):])
					if err := binary.Read(glyphReader, binary.BigEndian, &glyphID); err != nil {
						// Should probably handle error, but nested loop context, maybe break
						break
					}
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
	r := bytes.NewReader(data[offset:])
	if _, err := r.Seek(12, io.SeekCurrent); err != nil {
		return fmt.Errorf("failed to seek: %w", err)
	}

	var numGroups uint32
	if err := binary.Read(r, binary.BigEndian, &numGroups); err != nil {
		return fmt.Errorf("failed to read numGroups: %w", err)
	}

	for i := uint32(0); i < numGroups; i++ {
		var startCharCode, endCharCode, startGlyphID uint32
		if err := binary.Read(r, binary.BigEndian, &startCharCode); err != nil {
			return fmt.Errorf("failed to read startCharCode: %w", err)
		}
		if err := binary.Read(r, binary.BigEndian, &endCharCode); err != nil {
			return fmt.Errorf("failed to read endCharCode: %w", err)
		}
		if err := binary.Read(r, binary.BigEndian, &startGlyphID); err != nil {
			return fmt.Errorf("failed to read startGlyphID: %w", err)
		}

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
func (f *TTFFont) parseName(data []byte) error {
	table, ok := f.Tables["name"]
	if !ok {
		return errors.New("missing 'name' table")
	}

	r := bytes.NewReader(data[table.Offset:])
	if _, err := r.Seek(2, io.SeekCurrent); err != nil {
		return fmt.Errorf("failed to seek: %w", err)
	}

	var count, stringOffset uint16
	if err := binary.Read(r, binary.BigEndian, &count); err != nil {
		return fmt.Errorf("failed to read count: %w", err)
	}
	if err := binary.Read(r, binary.BigEndian, &stringOffset); err != nil {
		return fmt.Errorf("failed to read stringOffset: %w", err)
	}

	storageOffset := table.Offset + uint32(stringOffset)

	for i := uint16(0); i < count; i++ {
		var platformID, encodingID, languageID, nameID, length, offset uint16
		if err := binary.Read(r, binary.BigEndian, &platformID); err != nil {
			return fmt.Errorf("failed to read platformID: %w", err)
		}
		if err := binary.Read(r, binary.BigEndian, &encodingID); err != nil {
			return fmt.Errorf("failed to read encodingID: %w", err)
		}
		if err := binary.Read(r, binary.BigEndian, &languageID); err != nil {
			return fmt.Errorf("failed to read languageID: %w", err)
		}
		if err := binary.Read(r, binary.BigEndian, &nameID); err != nil {
			return fmt.Errorf("failed to read nameID: %w", err)
		}
		if err := binary.Read(r, binary.BigEndian, &length); err != nil {
			return fmt.Errorf("failed to read length: %w", err)
		}
		if err := binary.Read(r, binary.BigEndian, &offset); err != nil {
			return fmt.Errorf("failed to read offset: %w", err)
		}

		// Extract string (prefer platform 3 = Windows, encoding 1 = Unicode BMP)
		if platformID == 3 && encodingID == 1 {
			strStart := storageOffset + uint32(offset)
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
			strStart := storageOffset + uint32(offset)
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

	r := bytes.NewReader(data[table.Offset:])

	var version uint16
	if err := binary.Read(r, binary.BigEndian, &version); err != nil {
		return fmt.Errorf("failed to read version: %w", err)
	}

	if _, err := r.Seek(2, io.SeekCurrent); err != nil {
		return fmt.Errorf("failed to seek: %w", err)
	}

	var usWeightClass uint16
	if err := binary.Read(r, binary.BigEndian, &usWeightClass); err != nil {
		return fmt.Errorf("failed to read usWeightClass: %w", err)
	}
	f.IsBold = usWeightClass >= 700

	if _, err := r.Seek(60, io.SeekCurrent); err != nil { // Skip to fsSelection
		return fmt.Errorf("failed to seek: %w", err)
	}

	var fsSelection uint16
	if err := binary.Read(r, binary.BigEndian, &fsSelection); err != nil {
		return fmt.Errorf("failed to read fsSelection: %w", err)
	}
	f.IsItalic = (fsSelection & 0x0001) != 0

	if _, err := r.Seek(4, io.SeekCurrent); err != nil { // Skip sTypoAscender, sTypoDescender
		return fmt.Errorf("failed to seek: %w", err)
	}

	if version >= 2 && table.Length >= 96 {
		if _, err := r.Seek(16, io.SeekCurrent); err != nil { // Skip to sCapHeight (at offset 88)
			return fmt.Errorf("failed to seek: %w", err)
		}

		// sxHeight is at offset 86 in version 2+
		if _, err := r.Seek(-2, io.SeekCurrent); err != nil {
			return fmt.Errorf("failed to seek: %w", err)
		}
		if err := binary.Read(r, binary.BigEndian, &f.XHeight); err != nil {
			return fmt.Errorf("failed to read XHeight: %w", err)
		}
		if err := binary.Read(r, binary.BigEndian, &f.CapHeight); err != nil {
			return fmt.Errorf("failed to read CapHeight: %w", err)
		}
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

	r := bytes.NewReader(data[table.Offset:])
	if _, err := r.Seek(4, io.SeekCurrent); err != nil {
		return fmt.Errorf("failed to seek: %w", err)
	}

	// Read italic angle as fixed-point (16.16)
	var italicAngleFixed int32
	if err := binary.Read(r, binary.BigEndian, &italicAngleFixed); err != nil {
		return fmt.Errorf("failed to read italicAngleFixed: %w", err)
	}
	f.ItalicAngle = float64(italicAngleFixed) / 65536.0

	if _, err := r.Seek(4, io.SeekCurrent); err != nil {
		return fmt.Errorf("failed to seek: %w", err)
	}

	var isFixedPitch uint32
	if err := binary.Read(r, binary.BigEndian, &isFixedPitch); err != nil {
		return fmt.Errorf("failed to read isFixedPitch: %w", err)
	}
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
	glyphSet := make(map[uint16]bool)
	glyphSet[0] = true // Always include .notdef

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
