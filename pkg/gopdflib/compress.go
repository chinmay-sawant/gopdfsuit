// Package gopdflib provides PDF compression functionality.
package gopdflib

import (
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/compress"
)

// CompressOptions controls how aggressively CompressPDF rewrites streams and images.
// Empty Level selects Medium (JPEG 75, max image edge 1275). JPEGQuality and
// MaxImageDim override the preset when greater than zero.
//
// Tiers match the Ghostscript product:
//
//	Light  — JPEG 92, max edge 1920
//	Medium — JPEG 75, max edge 1275
//	Heavy  — JPEG 50, max edge 612
type CompressOptions = compress.Options

// Compression tiers for CompressOptions.Level.
const (
	CompressLight  = compress.LevelLight
	CompressMedium = compress.LevelMedium
	CompressHeavy  = compress.LevelHeavy
)

// CompressPDF rewrites an existing PDF: bicubic image downsample and JPEG
// recompression at the chosen tier, unused TTF glyph outlines dropped,
// document metadata stripped, and streams Flate-compressed. Encrypted files
// are rejected.
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
	return compress.CompressPDF(data, opts)
}
