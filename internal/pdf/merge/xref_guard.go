package merge

import (
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/pdfobj"
)

// MaxXRefEntryWidth caps the total byte width of one cross-reference stream
// entry. Owned by pdfobj; kept here as an alias for existing callers.
const MaxXRefEntryWidth = pdfobj.MaxXRefEntryWidth

// MaxInflateBytes caps a single decompressed stream. Owned by pdfobj.
const MaxInflateBytes = pdfobj.MaxInflateBytes

// InflateCapped decompresses zlib-wrapped (rawFlate=false) or raw flate data,
// rejecting outputs over MaxInflateBytes. Owned by pdfobj.
func InflateCapped(b []byte, rawFlate bool) ([]byte, error) {
	return pdfobj.InflateCapped(b, rawFlate)
}

// ValidXRefWidths checks /W field widths from a cross-reference stream.
// Owned by pdfobj.
func ValidXRefWidths(w0, w1, w2 int) (total int, ok bool) {
	return pdfobj.ValidXRefWidths(w0, w1, w2)
}
