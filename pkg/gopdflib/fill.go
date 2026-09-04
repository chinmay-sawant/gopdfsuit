// Package gopdflib provides PDF form filling functionality.
package gopdflib

import (
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/form"
)

// FillPDFWithXFDF fills a PDF form with data from an XFDF file.
// XFDF (XML Forms Data Format) is an XML-based format for representing
// form data and annotations in PDF documents.
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
