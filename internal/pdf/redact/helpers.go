package redact

import (
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/v7/internal/pdf/pdfobj"
)

// bytesIndex is a helper to find a subsequence in a []byte
func bytesIndex(b, sub []byte) int {
	return strings.Index(string(b), string(sub))
}

// trailerHasEncrypt checks if the PDF declares document encryption: a real
// /Encrypt entry (indirect reference or inline dict) outside stream data.
// Plain "/Encrypt" text inside a content stream no longer counts.
func trailerHasEncrypt(data []byte) bool {
	return pdfobj.HasEncryptEntry(data)
}

// tryZlibDecompress attempts to decompress zlib data
func tryZlibDecompress(b []byte) ([]byte, error) {
	return pdfobj.InflateZlib(b)
}

// tryFlateDecompress attempts to decompress raw flate data
func tryFlateDecompress(b []byte) ([]byte, error) {
	return pdfobj.InflateFlate(b)
}

// findRootRef looks for /Root n m R in the PDF bytes.
func findRootRef(data []byte) (objNum int, genNum int, ok bool) {
	return pdfobj.FindRootRef(data)
}

func objGenNum(objGen map[int]int, objNum int) int {
	if objGen == nil {
		return 0
	}
	if g, ok := objGen[objNum]; ok {
		return g
	}
	return 0
}

func isPDFWhitespace(b byte) bool {
	return pdfobj.IsWhitespace(b)
}

// parseXRefStreams looks for XRef stream objects and uses them to augment objMap / objGen.
func parseXRefStreams(data []byte, objMap map[int][]byte, objGen map[int]int) {
	pdfobj.AugmentRawIntMap(data, objMap, objGen)
}
