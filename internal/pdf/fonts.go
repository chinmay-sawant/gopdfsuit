package pdf

import (
	"fmt"
	"strings"
)

// PDF 2.0 compliant font definitions for the standard 14 fonts
// These include FirstChar, LastChar, Widths, and FontDescriptor as required by Arlington Model

// FontMetrics holds the complete metrics for a standard font
type FontMetrics struct {
	BaseFont       string
	FirstChar      int
	LastChar       int
	Widths         []int
	FontDescriptor FontDescriptor
}

// FontDescriptor holds font descriptor information
type FontDescriptor struct {
	FontName    string
	Flags       int
	FontBBox    [4]int
	ItalicAngle int
	Ascent      int
	Descent     int
	CapHeight   int
	StemV       int
	XHeight     int
}

// Standard Helvetica widths for WinAnsiEncoding (chars 32-255)
// These are the actual Adobe Helvetica metrics
// Note: helveticaWidths is already defined in xfdf.go, so we use a function to get it
var standardHelveticaWidths = []int{
	278, 278, 355, 556, 556, 889, 667, 191, 333, 333, 389, 584, 278, 333, 278, 278, // 32-47
	556, 556, 556, 556, 556, 556, 556, 556, 556, 556, 278, 278, 584, 584, 584, 556, // 48-63
	1015, 667, 667, 722, 722, 667, 611, 778, 722, 278, 500, 667, 556, 833, 722, 778, // 64-79
	667, 778, 722, 667, 611, 722, 667, 944, 667, 667, 611, 278, 278, 278, 469, 556, // 80-95
	333, 556, 556, 500, 556, 556, 278, 556, 556, 222, 222, 500, 222, 833, 556, 556, // 96-111
	556, 556, 333, 500, 278, 556, 500, 722, 500, 500, 500, 334, 260, 334, 584, 350, // 112-127
	556, 350, 222, 556, 333, 1000, 556, 556, 333, 1000, 667, 333, 1000, 350, 611, 350, // 128-143
	350, 222, 222, 333, 333, 350, 556, 1000, 333, 1000, 500, 333, 944, 350, 500, 667, // 144-159
	278, 333, 556, 556, 556, 556, 260, 556, 333, 737, 370, 556, 584, 333, 737, 333, // 160-175
	400, 584, 333, 333, 333, 556, 537, 278, 333, 333, 365, 556, 834, 834, 834, 611, // 176-191
	667, 667, 667, 667, 667, 667, 1000, 722, 667, 667, 667, 667, 278, 278, 278, 278, // 192-207
	722, 722, 778, 778, 778, 778, 778, 584, 778, 722, 722, 722, 722, 667, 667, 611, // 208-223
	556, 556, 556, 556, 556, 556, 889, 500, 556, 556, 556, 556, 278, 278, 278, 278, // 224-239
	556, 556, 556, 556, 556, 556, 556, 584, 611, 556, 556, 556, 556, 500, 556, 500, // 240-255
}

// Helvetica Bold widths
var helveticaBoldWidths = []int{
	278, 333, 474, 556, 556, 889, 722, 238, 333, 333, 389, 584, 278, 333, 278, 278, // 32-47
	556, 556, 556, 556, 556, 556, 556, 556, 556, 556, 333, 333, 584, 584, 584, 611, // 48-63
	975, 722, 722, 722, 722, 667, 611, 778, 722, 278, 556, 722, 611, 833, 722, 778, // 64-79
	667, 778, 722, 667, 611, 722, 667, 944, 667, 667, 611, 333, 278, 333, 584, 556, // 80-95
	333, 556, 611, 556, 611, 556, 333, 611, 611, 278, 278, 556, 278, 889, 611, 611, // 96-111
	611, 611, 389, 556, 333, 611, 556, 778, 556, 556, 500, 389, 280, 389, 584, 350, // 112-127
	556, 350, 278, 556, 500, 1000, 556, 556, 333, 1000, 667, 333, 1000, 350, 611, 350, // 128-143
	350, 278, 278, 500, 500, 350, 556, 1000, 333, 1000, 556, 333, 944, 350, 500, 667, // 144-159
	278, 333, 556, 556, 556, 556, 280, 556, 333, 737, 370, 556, 584, 333, 737, 333, // 160-175
	400, 584, 333, 333, 333, 611, 556, 278, 333, 333, 365, 556, 834, 834, 834, 611, // 176-191
	722, 722, 722, 722, 722, 722, 1000, 722, 667, 667, 667, 667, 278, 278, 278, 278, // 192-207
	722, 722, 778, 778, 778, 778, 778, 584, 778, 722, 722, 722, 722, 667, 667, 611, // 208-223
	556, 556, 556, 556, 556, 556, 889, 556, 556, 556, 556, 556, 278, 278, 278, 278, // 224-239
	611, 611, 611, 611, 611, 611, 611, 584, 611, 611, 611, 611, 611, 556, 611, 556, // 240-255
}

// Helvetica Oblique widths (same as regular Helvetica)
var standardHelveticaObliqueWidths = standardHelveticaWidths

// Helvetica Bold Oblique widths (same as Helvetica Bold)
var standardHelveticaBoldObliqueWidths = helveticaBoldWidths

// GetFontMetrics returns the complete font metrics for a given font name
func GetFontMetrics(fontName string) FontMetrics {
	switch fontName {
	case "Helvetica":
		return FontMetrics{
			BaseFont:  "Helvetica",
			FirstChar: 32,
			LastChar:  255,
			Widths:    standardHelveticaWidths,
			FontDescriptor: FontDescriptor{
				FontName:    "Helvetica",
				Flags:       32, // Non-symbolic
				FontBBox:    [4]int{-166, -225, 1000, 931},
				ItalicAngle: 0,
				Ascent:      718,
				Descent:     -207,
				CapHeight:   718,
				StemV:       88,
				XHeight:     523,
			},
		}
	case "Helvetica-Bold":
		return FontMetrics{
			BaseFont:  "Helvetica-Bold",
			FirstChar: 32,
			LastChar:  255,
			Widths:    helveticaBoldWidths,
			FontDescriptor: FontDescriptor{
				FontName:    "Helvetica-Bold",
				Flags:       32 | 262144, // Non-symbolic + ForceBold
				FontBBox:    [4]int{-170, -228, 1003, 962},
				ItalicAngle: 0,
				Ascent:      718,
				Descent:     -207,
				CapHeight:   718,
				StemV:       140,
				XHeight:     532,
			},
		}
	case "Helvetica-Oblique":
		return FontMetrics{
			BaseFont:  "Helvetica-Oblique",
			FirstChar: 32,
			LastChar:  255,
			Widths:    standardHelveticaObliqueWidths,
			FontDescriptor: FontDescriptor{
				FontName:    "Helvetica-Oblique",
				Flags:       32 | 64, // Non-symbolic + Italic
				FontBBox:    [4]int{-170, -225, 1116, 931},
				ItalicAngle: -12,
				Ascent:      718,
				Descent:     -207,
				CapHeight:   718,
				StemV:       88,
				XHeight:     523,
			},
		}
	case "Helvetica-BoldOblique":
		return FontMetrics{
			BaseFont:  "Helvetica-BoldOblique",
			FirstChar: 32,
			LastChar:  255,
			Widths:    standardHelveticaBoldObliqueWidths,
			FontDescriptor: FontDescriptor{
				FontName:    "Helvetica-BoldOblique",
				Flags:       32 | 64 | 262144, // Non-symbolic + Italic + ForceBold
				FontBBox:    [4]int{-174, -228, 1114, 962},
				ItalicAngle: -12,
				Ascent:      718,
				Descent:     -207,
				CapHeight:   718,
				StemV:       140,
				XHeight:     532,
			},
		}
	default:
		// Default to Helvetica
		return GetFontMetrics("Helvetica")
	}
}

// GenerateFontObject creates a complete PDF 2.0 compliant font object
// Returns the font object string and the FontDescriptor object ID used
func GenerateFontObject(fontName string, fontObjectID, fontDescriptorID, widthsArrayID int) string {
	metrics := GetFontMetrics(fontName)

	// Build compact font dictionary without deprecated /Name field for PDF 2.0
	return fmt.Sprintf("%d 0 obj\n<</Type/Font/Subtype/Type1/BaseFont/%s/Encoding/WinAnsiEncoding/FirstChar %d/LastChar %d/Widths %d 0 R/FontDescriptor %d 0 R>>\nendobj\n",
		fontObjectID, metrics.BaseFont, metrics.FirstChar, metrics.LastChar, widthsArrayID, fontDescriptorID)
}

// GenerateFontDescriptorObject creates a FontDescriptor object
func GenerateFontDescriptorObject(fontName string, objectID int) string {
	metrics := GetFontMetrics(fontName)
	fd := metrics.FontDescriptor

	// Compact format - single line
	return fmt.Sprintf("%d 0 obj\n<</Type/FontDescriptor/FontName/%s/Flags %d/FontBBox[%d %d %d %d]/ItalicAngle %d/Ascent %d/Descent %d/CapHeight %d/StemV %d/XHeight %d>>\nendobj\n",
		objectID, fd.FontName, fd.Flags, fd.FontBBox[0], fd.FontBBox[1], fd.FontBBox[2], fd.FontBBox[3],
		fd.ItalicAngle, fd.Ascent, fd.Descent, fd.CapHeight, fd.StemV, fd.XHeight)
}

// GenerateWidthsArrayObject creates a Widths array object
func GenerateWidthsArrayObject(fontName string, objectID int) string {
	metrics := GetFontMetrics(fontName)

	var widthsArray strings.Builder
	widthsArray.WriteString(fmt.Sprintf("%d 0 obj\n[", objectID))

	// Compact format - no newlines, minimal spacing
	for i, w := range metrics.Widths {
		if i > 0 {
			widthsArray.WriteString(" ")
		}
		widthsArray.WriteString(fmt.Sprintf("%d", w))
	}

	widthsArray.WriteString("]\nendobj\n")

	return widthsArray.String()
}

// GetHelveticaFontResourceString returns a complete inline font resource for XObjects
// This is used in form field appearance streams - optimized for minimal size
// arlingtonCompatible: if true, includes full font metrics for PDF 2.0 compliance
func GetHelveticaFontResourceString() string {
	metrics := GetFontMetrics("Helvetica")

	// Build compact widths array inline (no extra spaces)
	var widths strings.Builder
	widths.WriteString("[")
	for i, w := range metrics.Widths {
		if i > 0 {
			widths.WriteString(" ")
		}
		widths.WriteString(fmt.Sprintf("%d", w))
	}
	widths.WriteString("]")

	// Build compact font dictionary with inline FontDescriptor
	return fmt.Sprintf(`<</Type/Font/Subtype/Type1/BaseFont/Helvetica/Encoding/WinAnsiEncoding/FirstChar %d/LastChar %d/Widths %s/FontDescriptor<</Type/FontDescriptor/FontName/Helvetica/Flags 32/FontBBox[-166 -225 1000 931]/ItalicAngle 0/Ascent 718/Descent -207/CapHeight 718/StemV 88/XHeight 523>>>>`,
		metrics.FirstChar, metrics.LastChar, widths.String())
}

// GetSimpleHelveticaFontResourceString returns a simple inline font resource for XObjects
// This is used when Arlington compatibility is OFF - minimal font definition
func GetSimpleHelveticaFontResourceString() string {
	return `<</Type/Font/Subtype/Type1/BaseFont/Helvetica>>`
}

// GenerateSimpleFontObject creates a simple font object (non-Arlington mode)
// This is the legacy format without FirstChar, LastChar, Widths, and FontDescriptor
func GenerateSimpleFontObject(fontName string, fontRef string, fontObjectID int) string {
	return fmt.Sprintf("%d 0 obj\n<< /Type /Font /Subtype /Type1 /Name %s /BaseFont /%s >>\nendobj\n",
		fontObjectID, fontRef, fontName)
}
