// Package gopdflib provides HTML to PDF/Image conversion functionality.
package gopdflib

import (
	"github.com/chinmay-sawant/gopdfsuit/v5/internal/pdf"
)

// ConvertHTMLToPDF converts HTML content or a URL to a PDF document.
// Requires Chrome/Chromium to be available on the system.
func ConvertHTMLToPDF(req HTMLToPDFRequest) ([]byte, error) {
	return pdf.ConvertHTMLToPDF(req)
}

// ConvertHTMLToImage converts HTML content or a URL to an image (png, jpg, svg).
// Requires Chrome/Chromium to be available on the system.
func ConvertHTMLToImage(req HTMLToImageRequest) ([]byte, error) {
	return pdf.ConvertHTMLToImage(req)
}
