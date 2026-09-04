// Package gopdflib provides PDF compression functionality.
package gopdflib

import (
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/compress"
)

// MaxCompressInputBytes is the largest PDF CompressPDF accepts (32 MiB,
// shared with the HTTP and WASM entry points).
const MaxCompressInputBytes = compress.MaxInputBytes

// CompressPDF rewrites an existing PDF: bicubic image downsample and JPEG
// recompression at the chosen tier, unused TTF glyph outlines dropped,
// document metadata stripped, and streams Flate-compressed. Encrypted files
// are rejected. Input larger than MaxCompressInputBytes is rejected.
//
// Example:
//
//	pdfBytes, _ := os.ReadFile("document.pdf")
//	out, err := gopdflib.CompressPDF(pdfBytes, gopdflib.CompressOptions{Level: gopdflib.CompressHeavy})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	os.WriteFile("document-compressed.pdf", out, 0644)
func CompressPDF(data []byte, opts CompressOptions) ([]byte, error) {
	const op = "gopdflib: CompressPDF"
	if len(data) == 0 {
		return nil, invalidInputError(op, "needs a non-empty PDF")
	}
	if len(data) > MaxCompressInputBytes {
		return nil, limitExceededError(op, "PDF exceeds maximum size")
	}
	out, err := compress.CompressPDF(data, toInternalCompressOptions(opts))
	if err != nil {
		return nil, wrapEngineError(op, err)
	}
	return out, nil
}
