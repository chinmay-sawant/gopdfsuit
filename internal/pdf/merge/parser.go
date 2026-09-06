package merge

import (
	"github.com/chinmay-sawant/gopdfsuit/v7/internal/pdf/pdfobj"
)

// PDF parsing functions for the merge package.
//
// The implementations live in internal/pdf/pdfobj (the shared read seam);
// everything below is a thin alias so existing callers and tests keep
// compiling with identical behavior.

// DetectPDFVersion extracts the PDF version from the header (e.g., "1.4", "1.7", "2.0")
func DetectPDFVersion(data []byte) string {
	return pdfobj.DetectVersion(data)
}

// CompareVersions compares two PDF version strings
// Returns: 1 if v1 > v2, -1 if v1 < v2, 0 if equal
func CompareVersions(v1, v2 string) int {
	return pdfobj.CompareVersions(v1, v2)
}

// FindObjectBoundaries finds all PDF objects in the data
func FindObjectBoundaries(data []byte) []ObjectBoundary {
	bs := pdfobj.FindObjectBoundaries(data)
	out := make([]ObjectBoundary, len(bs))
	for i, b := range bs {
		out[i] = ObjectBoundary{
			ObjNum:    b.ObjNum,
			GenNum:    b.GenNum,
			Start:     b.Start,
			BodyStart: b.BodyStart,
			End:       b.End,
		}
	}
	return out
}

// FindEndObj finds the position right after "endobj" starting from pos
func FindEndObj(data []byte, pos int) int {
	return pdfobj.FindEndObj(data, pos)
}

// SkipStringLiteral skips a PDF string literal (...) handling escapes and nested parens
func SkipStringLiteral(data []byte, pos int) int {
	return pdfobj.SkipStringLiteral(data, pos)
}

// SkipHexString skips a PDF hex string <...>
func SkipHexString(data []byte, pos int) int {
	return pdfobj.SkipHexString(data, pos)
}

// SkipDictionary skips a PDF dictionary <<...>>
func SkipDictionary(data []byte, pos int) int {
	return pdfobj.SkipDictionary(data, pos)
}

// SkipArray skips a PDF array [...]
func SkipArray(data []byte, pos int) int {
	return pdfobj.SkipArray(data, pos)
}

// FindStreamStart finds the start of a stream in data, skipping strings/dicts
// Returns index of "stream" keyword, or -1
func FindStreamStart(data []byte) int {
	return pdfobj.FindStreamStart(data)
}

// ReplaceRefsOutsideStreams rewrites indirect references only outside stream blocks
func ReplaceRefsOutsideStreams(data []byte, offset int) []byte {
	return pdfobj.ReplaceRefsOutsideStreams(data, offset)
}

// HasSubstring checks if data contains substring
func HasSubstring(data, sub []byte) bool {
	return pdfobj.HasSubstring(data, sub)
}

// IsPageObject checks if the object body is a Page object
func IsPageObject(body []byte) bool {
	return pdfobj.IsPageObject(body)
}

// IsPagesTreeObject checks if the object is a Pages tree node
func IsPagesTreeObject(body []byte) bool {
	return pdfobj.IsPagesTreeObject(body)
}

// IsWidgetAnnotation checks if the object is a Widget annotation
func IsWidgetAnnotation(body []byte) bool {
	return pdfobj.IsWidgetAnnotation(body)
}

// IsFormField checks if the object has form field type
func IsFormField(body []byte) bool {
	return pdfobj.IsFormField(body)
}

// IsXObjectForm checks if object is a Form XObject (appearance stream)
func IsXObjectForm(body []byte) bool {
	return pdfobj.IsXObjectForm(body)
}

// IsObjectStream checks if the object is an Object Stream (ObjStm)
func IsObjectStream(body []byte) bool {
	return pdfobj.IsObjectStream(body)
}

// isWhitespace checks if byte is PDF whitespace
func isWhitespace(b byte) bool {
	return pdfobj.IsWhitespace(b)
}

// ParseObjectStream extracts objects from a compressed object stream
func ParseObjectStream(body []byte) map[int][]byte {
	return pdfobj.ParseObjectStream(body)
}
