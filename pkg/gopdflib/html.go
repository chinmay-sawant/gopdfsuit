// Package gopdflib provides HTML to PDF/Image conversion functionality.
package gopdflib

import (
	"github.com/chinmay-sawant/gopdfsuit/v4/internal/pdf"
)

// ConvertHTMLToPDF converts HTML content or a URL to a PDF document.
// This function requires Chrome/Chromium to be available on the system.
//
// Example - Convert HTML string:
//
//	req := gopdflib.HtmlToPDFRequest{
//	    HTML:        "<html><body><h1>Hello World</h1></body></html>",
//	    PageSize:    "A4",
//	    Orientation: "Portrait",
//	}
//	pdfBytes, err := gopdflib.ConvertHTMLToPDF(req)
//
// Example - Convert URL:
//
//	req := gopdflib.HtmlToPDFRequest{
//	    URL:      "https://example.com",
//	    PageSize: "Letter",
//	}
//	pdfBytes, err := gopdflib.ConvertHTMLToPDF(req)
func ConvertHTMLToPDF(req HtmlToPDFRequest) ([]byte, error) {
	return pdf.ConvertHTMLToPDF(req)
}

// ConvertHTMLToImage converts HTML content or a URL to an image.
// Supported formats: png, jpg/jpeg, svg (default: png).
// This function requires Chrome/Chromium to be available on the system.
//
// Example:
//
//	req := gopdflib.HtmlToImageRequest{
//	    HTML:   "<html><body><h1>Hello World</h1></body></html>",
//	    Format: "png",
//	    Width:  800,
//	    Height: 600,
//	}
//	imgBytes, err := gopdflib.ConvertHTMLToImage(req)
func ConvertHTMLToImage(req HtmlToImageRequest) ([]byte, error) {
	return pdf.ConvertHTMLToImage(req)
}
