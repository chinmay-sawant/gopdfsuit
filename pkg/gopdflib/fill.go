// Package gopdflib provides PDF form filling functionality.
package gopdflib

import (
	"github.com/chinmay-sawant/gopdfsuit/v4/internal/pdf/form"
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
	return form.FillPDFWithXFDF(pdfBytes, xfdfBytes)
}
