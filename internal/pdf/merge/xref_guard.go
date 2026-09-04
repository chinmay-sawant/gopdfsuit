package merge

import (
	"bytes"
	"compress/flate"
	"compress/zlib"
	"fmt"
	"io"
)

// MaxXRefEntryWidth caps the total byte width of one cross-reference stream
// entry (/W array sum). Field widths come from the file, so they must be
// validated before slicing: W=[0 0 0] would make the entry loop never
// advance (hang) and negative widths panic on slicing.
const MaxXRefEntryWidth = 32

// MaxInflateBytes caps a single decompressed stream (48 MiB, matching the
// merge parser and compress tiers) so a crafted Flate stream cannot act as
// a zip bomb.
const MaxInflateBytes = 48 << 20

// InflateCapped decompresses zlib-wrapped (rawZlib=false) or raw flate data,
// rejecting outputs over MaxInflateBytes.
func InflateCapped(b []byte, rawFlate bool) ([]byte, error) {
	var r io.ReadCloser
	if rawFlate {
		r = flate.NewReader(bytes.NewReader(b))
	} else {
		var err error
		r, err = zlib.NewReader(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
	}
	defer func() {
		_ = r.Close()
	}()
	var out bytes.Buffer
	n, err := io.Copy(&out, io.LimitReader(r, int64(MaxInflateBytes)+1))
	if err != nil {
		return nil, err
	}
	if n > int64(MaxInflateBytes) {
		return nil, fmt.Errorf("decompressed stream exceeds %d bytes", MaxInflateBytes)
	}
	return out.Bytes(), nil
}

// ValidXRefWidths checks /W field widths from a cross-reference stream.
// It returns the entry width and true when the widths are safe to slice
// with: all non-negative, total in (0, MaxXRefEntryWidth].
func ValidXRefWidths(w0, w1, w2 int) (total int, ok bool) {
	if w0 < 0 || w1 < 0 || w2 < 0 {
		return 0, false
	}
	total = w0 + w1 + w2
	if total <= 0 || total > MaxXRefEntryWidth {
		return 0, false
	}
	return total, true
}
