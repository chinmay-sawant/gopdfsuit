// Package gopdflib provides PDF generation from templates.
package gopdflib

import (
	"sync"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf"
)

// warmPoolsOnce lazily initializes the runtime buffer/compression pools on
// the first GeneratePDF call instead of at import time, so importing this
// package does not pre-allocate ~2.75 MiB x GOMAXPROCS. WarmRuntimePools
// remains exported for callers (e.g. cmd/gopdfsuit/main.go) that want to
// warm pools explicitly at startup; it is idempotent.
var warmPoolsOnce sync.Once

// WarmRuntimePools pre-warms compression and buffer pools. It is safe to
// call multiple times; only the first call allocates.
func WarmRuntimePools() {
	warmPoolsOnce.Do(func() {
		pdf.WarmRuntimePools()
	})
}

func ensureRuntimePools() {
	WarmRuntimePools()
}

type BorrowedPDF = pdf.BorrowedPDF
type PDFCapacityHighWater = pdf.CapacityHighWater

// ResetPDFCapacityHighWater clears per-tier buffer high-water counters. Active when
// BENCH_DEBUG_CAPS=1 during generation.
func ResetPDFCapacityHighWater() {
	pdf.ResetPDFCapacityHighWater()
}

// SnapshotPDFCapacityHighWater returns the peak final len/cap per template tier.
func SnapshotPDFCapacityHighWater() PDFCapacityHighWater {
	return pdf.SnapshotPDFCapacityHighWater()
}

// GeneratePDF creates a PDF document from a template and returns the PDF bytes.
//
// Example usage:
//
//	template := gopdflib.PDFTemplate{
//	    Config: gopdflib.Config{
//	        Page:          "A4",
//	        PageAlignment: 1, // Portrait
//	    },
//	    Title: gopdflib.Title{
//	        Props: "Helvetica:18:100:center:0:0:0:0",
//	        Text:  "My Document",
//	    },
//	    Elements: []gopdflib.Element{
//	        {
//	            Type: "table",
//	            Table: &gopdflib.Table{
//	                MaxColumns:   2,
//	                ColumnWidths: []float64{1, 1},
//	                Rows: []gopdflib.Row{
//	                    {Row: []gopdflib.Cell{
//	                        {Props: "Helvetica:12:100:left:1:1:1:1", Text: "Column 1"},
//	                        {Props: "Helvetica:12:100:left:1:1:1:1", Text: "Column 2"},
//	                    }},
//	                },
//	            },
//	        },
//	    },
//	}
//
//	pdfBytes, err := gopdflib.GeneratePDF(template)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	os.WriteFile("output.pdf", pdfBytes, 0644)
func GeneratePDF(template PDFTemplate) ([]byte, error) {
	ensureRuntimePools()
	return pdf.GenerateTemplatePDF(toInternalTemplate(template))
}

// GeneratePDFBorrowed creates a PDF document without cloning the final pooled
// assembly buffer. The caller owns the buffer until Release is called and
// MUST call Release exactly once, preferably via defer:
//
//	doc, err := gopdflib.GeneratePDFBorrowed(template)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer doc.Release()
//	use(doc.Bytes())
//
// Bytes() borrows the pooled buffer: the slice is invalid after Release and
// must not be retained. Use CopyBytes when the bytes must outlive Release.
func GeneratePDFBorrowed(template PDFTemplate) (*BorrowedPDF, error) {
	ensureRuntimePools()
	return pdf.GenerateTemplatePDFBorrowed(toInternalTemplate(template))
}

// GetAvailableFonts returns a list of available fonts for PDF generation.
// This includes standard PDF fonts and any custom fonts that have been registered.
func GetAvailableFonts() []FontInfo {
	internal := pdf.GetAvailableFonts()
	out := make([]FontInfo, 0, len(internal))
	for _, f := range internal {
		out = append(out, mustFromInternal[models.FontInfo, FontInfo](f))
	}
	return out
}

// GetFontRegistry returns the global font registry for registering custom fonts.
// Use this to register custom TTF/OTF fonts before generating PDFs.
//
// Concurrency: the registry itself is mutex-guarded and each GeneratePDF call
// operates on an isolated per-generation clone, so concurrent GeneratePDF
// calls are safe. Registering (or clearing) fonts mutates the global and must
// happen-before concurrent generation, or be externally synchronized.
//
// Example:
//
//	registry := gopdflib.GetFontRegistry()
//	err := registry.RegisterFontFromFile("MyFont", "/path/to/font.ttf")
func GetFontRegistry() *pdf.CustomFontRegistry {
	return pdf.GetFontRegistry()
}
