//go:build !js

// Package gopdflib provides HTML to PDF/Image conversion functionality.
package gopdflib

import (
	"fmt"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf"
)

// ConvertHTMLToPDF converts HTML content or a URL to a PDF document.
// This function is pure-Go via gowkhtmltopdf; no browser required.
//
// Example - Convert HTML string:
//
//	req := gopdflib.HTMLToPDFRequest{
//	    HTML:        "<html><body><h1>Hello World</h1></body></html>",
//	    PageSize:    "A4",
//	    Orientation: "Portrait",
//	}
//	pdfBytes, err := gopdflib.ConvertHTMLToPDF(req)
//
// Example - Convert URL:
//
//	req := gopdflib.HTMLToPDFRequest{
//	    URL:      "https://example.com",
//	    PageSize: "Letter",
//	}
//	pdfBytes, err := gopdflib.ConvertHTMLToPDF(req)
func ConvertHTMLToPDF(req HTMLToPDFRequest) ([]byte, error) {
	const op = "gopdflib: ConvertHTMLToPDF"
	if req.HTML == "" && req.URL == "" {
		return nil, invalidInputError(op, "needs HTML content or a URL")
	}
	in, err := toInternal[HTMLToPDFRequest, models.HTMLToPDFRequest](req)
	if err != nil {
		return nil, fmt.Errorf("%w: %s request translation: %w", ErrInvalidInput, op, err)
	}
	out, err := pdf.ConvertHTMLToPDF(in)
	if err != nil {
		return nil, wrapEngineError(op, err)
	}
	return out, nil
}

// ConvertHTMLToImage converts HTML content or a URL to an image.
// Supported formats: png, jpg/jpeg (default: png). Format svg has no
// gowkhtmltopdf equivalent and is rejected as invalid input.
// This function is pure-Go via gowkhtmltopdf; no browser required.
//
// Example:
//
//	req := gopdflib.HTMLToImageRequest{
//	    HTML:   "<html><body><h1>Hello World</h1></body></html>",
//	    Format: "png",
//	    Width:  800,
//	    Height: 600,
//	}
//	imgBytes, err := gopdflib.ConvertHTMLToImage(req)
func ConvertHTMLToImage(req HTMLToImageRequest) ([]byte, error) {
	const op = "gopdflib: ConvertHTMLToImage"
	if req.HTML == "" && req.URL == "" {
		return nil, invalidInputError(op, "needs HTML content or a URL")
	}
	if req.Format == "svg" {
		return nil, invalidInputError(op, "format svg is not supported: use png or jpg")
	}
	in, err := toInternal[HTMLToImageRequest, models.HTMLToImageRequest](req)
	if err != nil {
		return nil, fmt.Errorf("%w: %s request translation: %w", ErrInvalidInput, op, err)
	}
	out, err := pdf.ConvertHTMLToImage(in)
	if err != nil {
		return nil, wrapEngineError(op, err)
	}
	return out, nil
}
