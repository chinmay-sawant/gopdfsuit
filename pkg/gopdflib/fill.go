// Package gopdflib provides PDF form filling functionality.
package gopdflib

import (
	"github.com/chinmay-sawant/gopdfsuit/v7/internal/pdf/form"
)

// FillPDFWithXFDF fills a PDF form with data from an XFDF file.
// XFDF (XML Forms Data Format) is an XML-based format for representing
// form data and annotations in PDF documents.
//
// WASM: the browser binding goFillPDF passes pdfBytes and xfdfBytes as two
// Uint8Array values (via js.CopyBytesToGo); this signature already takes
// plain byte slices so no adapter is needed on the Go side.
//
// Limit: the /NeedAppearances true flag is applied as a byte-level patch to
// the AcroForm dictionary. When the AcroForm dictionary lives inside a
// compressed object stream (/ObjStm) the byte patch cannot reach it, so
// viewers may not regenerate field appearances on open for such files.
// Field values written into object streams are still updated by the
// object-stream-aware fill path; only the appearance flag is affected.
//
// Example:
//
//	pdfBytes, _ := os.ReadFile("form.pdf")
//	xfdfBytes, _ := os.ReadFile("data.xfdf")
//	filled, err := gopdflib.FillPDFWithXFDF(pdfBytes, xfdfBytes)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	os.WriteFile("filled.pdf", filled, 0644)
func FillPDFWithXFDF(pdfBytes, xfdfBytes []byte) ([]byte, error) {
	const op = "gopdflib: FillPDFWithXFDF"
	if len(pdfBytes) == 0 {
		return nil, invalidInputError(op, "needs a non-empty PDF")
	}
	if len(xfdfBytes) == 0 {
		return nil, invalidInputError(op, "needs non-empty XFDF data")
	}
	out, err := form.FillPDFWithXFDF(pdfBytes, xfdfBytes)
	if err != nil {
		return nil, wrapEngineError(op, err)
	}
	return out, nil
}
