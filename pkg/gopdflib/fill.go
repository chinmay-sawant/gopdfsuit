// Package gopdflib provides PDF form filling functionality.
package gopdflib

import (
	"errors"

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
	if len(pdfBytes) == 0 {
		return nil, errors.New("gopdflib: FillPDFWithXFDF needs a non-empty PDF")
	}
	if len(xfdfBytes) == 0 {
		return nil, errors.New("gopdflib: FillPDFWithXFDF needs non-empty XFDF data")
	}
	return form.FillPDFWithXFDF(pdfBytes, xfdfBytes)
}
