// Package gopdflib provides HTML to PDF/Image conversion functionality.
package gopdflib

import (
	"errors"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf"
)

// ConvertHTMLToPDF converts HTML content or a URL to a PDF document.
// This function requires Chrome/Chromium to be available on the system.
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
	if req.HTML == "" && req.URL == "" {
		return nil, errors.New("gopdflib: ConvertHTMLToPDF needs HTML content or a URL")
	}
	return pdf.ConvertHTMLToPDF(mustToInternal[HTMLToPDFRequest, models.HTMLToPDFRequest](req))
}

// ConvertHTMLToImage converts HTML content or a URL to an image.
// Supported formats: png, jpg/jpeg, svg (default: png).
// This function requires Chrome/Chromium to be available on the system.
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
	if req.HTML == "" && req.URL == "" {
		return nil, errors.New("gopdflib: ConvertHTMLToImage needs HTML content or a URL")
	}
	return pdf.ConvertHTMLToImage(mustToInternal[HTMLToImageRequest, models.HTMLToImageRequest](req))
}
