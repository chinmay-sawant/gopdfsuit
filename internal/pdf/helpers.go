package pdf

import (
	"regexp"
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/v7/internal/pdf/pdfobj"
)

var (
	encryptBytes = []byte("/Encrypt")
)

// bytesIndex is a helper to find a subsequence in a []byte
func bytesIndex(b, sub []byte) int {
	return strings.Index(string(b), string(sub))
}

// trailerHasEncrypt checks if trailer or any trailer 'Encrypt' appears
func trailerHasEncrypt(data []byte) bool {
	trRe := regexp.MustCompile(`trailer(?s).*?<<(.*?)>>`)
	for _, m := range trRe.FindAllSubmatch(data, -1) {
		if bytesIndex(m[1], encryptBytes) >= 0 {
			return true
		}
	}
	// also check for /Encrypt elsewhere
	return bytesIndex(data, encryptBytes) >= 0
}

// findRootRef looks for /Root n m R in the PDF bytes
func findRootRef(data []byte) (string, bool) {
	rootRe := regexp.MustCompile(`/Root\s+(\d+)\s+(\d+)\s+R`)
	if m := rootRe.FindSubmatch(data); m != nil {
		return string(m[1]) + " " + string(m[2]), true
	}
	return "", false
}

// parseXRefStreams looks for XRef stream objects and uses them to augment objMap.
// Owned by the pdfobj read seam; kept here as a thin adapter.
func parseXRefStreams(data []byte, objMap map[string][]byte) {
	pdfobj.AugmentStrMap(data, objMap)
}
