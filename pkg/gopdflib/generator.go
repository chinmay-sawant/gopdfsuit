// Package gopdflib provides PDF generation from templates.
package gopdflib

import (
	"fmt"
	"sync"

	"github.com/bytedance/sonic"
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
// call multiple times; only the first call allocates. The guard is a
// sync.Once with no goroutines, so it is also safe single-threaded under
// GOOS=js GOARCH=wasm: the WASM entry point may call it once at startup or
// rely on the lazy ensureRuntimePools path inside each Generate call.
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

// MaxTemplateJSONBytes caps the template JSON document accepted by the
// JSON entry points below (8 MiB, shared with the HTTP handler body cap).
// The WASM shim enforces the same cap on the JSON string length before
// copying it into the module.
const MaxTemplateJSONBytes = 8 << 20

// DecodeTemplateJSON decodes a template JSON document into a PDFTemplate.
//
// This is the WASM template path: the JS caller passes the template as a
// JSON string (JSON.stringify(templateObject)), the shim copies the bytes,
// and this function decodes them with the same policy as the HTTP handler:
// models.PDFTemplate.PreallocForDecode sizes the slice backing arrays from
// the document length before decode, then the result is translated to the
// owned public type. Base64 imagedata/fontData stay plain JSON strings end
// to end; for very large assets prefer passing raw Uint8Array bytes plus an
// id map handled in JS instead of embedding data URIs in the JSON.
func DecodeTemplateJSON(data []byte) (PDFTemplate, error) {
	const op = "gopdflib: DecodeTemplateJSON"
	if len(data) == 0 {
		return PDFTemplate{}, invalidInputError(op, "needs a non-empty template JSON document")
	}
	if len(data) > MaxTemplateJSONBytes {
		return PDFTemplate{}, limitExceededError(op, "template JSON exceeds maximum size")
	}
	var in models.PDFTemplate
	in.PreallocForDecode(len(data), "")
	if err := sonic.Unmarshal(data, &in); err != nil {
		return PDFTemplate{}, fmt.Errorf("%w: %s: %w", ErrInvalidInput, op, err)
	}
	out, err := fromInternal[models.PDFTemplate, PDFTemplate](in)
	if err != nil {
		return PDFTemplate{}, fmt.Errorf("%w: %s: %w", ErrInvalidInput, op, err)
	}
	// The font hint is not part of the JSON shape, so it does not survive
	// translation above: carry it over explicitly.
	out.SetPrecomputedStandardFonts(in.PrecomputedStandardFonts()...)
	return out, nil
}

// GeneratePDFFromJSON decodes a template JSON document (see DecodeTemplateJSON)
// and generates the PDF in one call. It is the primary WASM generate binding:
// callable with a JS template object serialized to a JSON string.
func GeneratePDFFromJSON(data []byte) ([]byte, error) {
	template, err := DecodeTemplateJSON(data)
	if err != nil {
		return nil, err
	}
	return GeneratePDF(template)
}

// GeneratePDFBorrowedFromJSON is the pooled-buffer variant of
// GeneratePDFFromJSON. The caller owns the buffer until Release is called
// and MUST call Release exactly once, preferably via defer. See
// GeneratePDFBorrowed for the borrow rules.
func GeneratePDFBorrowedFromJSON(data []byte) (*BorrowedPDF, error) {
	template, err := DecodeTemplateJSON(data)
	if err != nil {
		return nil, err
	}
	return GeneratePDFBorrowed(template)
}
func GeneratePDF(template PDFTemplate) ([]byte, error) {
	ensureRuntimePools()
	in, err := toInternalTemplate(template)
	if err != nil {
		return nil, fmt.Errorf("%w: gopdflib: GeneratePDF template translation: %w", ErrInvalidInput, err)
	}
	out, err := pdf.GenerateTemplatePDF(in)
	if err != nil {
		return nil, wrapEngineError("gopdflib: GeneratePDF", err)
	}
	return out, nil
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
	in, err := toInternalTemplate(template)
	if err != nil {
		return nil, fmt.Errorf("%w: gopdflib: GeneratePDFBorrowed template translation: %w", ErrInvalidInput, err)
	}
	doc, err := pdf.GenerateTemplatePDFBorrowed(in)
	if err != nil {
		return nil, wrapEngineError("gopdflib: GeneratePDFBorrowed", err)
	}
	return doc, nil
}

// GetAvailableFonts returns a list of available fonts for PDF generation.
// This includes standard PDF fonts and any custom fonts that have been registered.
func GetAvailableFonts() []FontInfo {
	internal := pdf.GetAvailableFonts()
	out := make([]FontInfo, 0, len(internal))
	for _, f := range internal {
		pub, err := fromInternal[models.FontInfo, FontInfo](f)
		if err != nil {
			continue
		}
		out = append(out, pub)
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
