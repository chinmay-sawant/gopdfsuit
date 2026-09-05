package handlers

import (
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
)

// This file owns HTML conversion request defaults and source checks.
// models.go is the type owner and stays untouched, so SetDefaults/Validate
// live here as constructor functions applied by the handlers before the
// service call. The field-to-knob mapping itself stays in
// internal/pdf/html_convert.go, the single mapping table.

// newHTMLToPDFRequest applies engine defaults to a decoded HTML-to-PDF
// request: A4 portrait, 10mm margins, 300 DPI.
func newHTMLToPDFRequest(req models.HTMLToPDFRequest) models.HTMLToPDFRequest {
	if req.PageSize == "" {
		req.PageSize = "A4"
	}
	if req.Orientation == "" {
		req.Orientation = "Portrait"
	}
	if req.MarginTop == "" {
		req.MarginTop = "10mm" //nolint:goconst
	}
	if req.MarginRight == "" {
		req.MarginRight = "10mm"
	}
	if req.MarginBottom == "" {
		req.MarginBottom = "10mm"
	}
	if req.MarginLeft == "" {
		req.MarginLeft = "10mm"
	}
	if req.DPI == 0 {
		req.DPI = 300
	}
	return req
}

// newHTMLToImageRequest applies engine defaults to a decoded HTML-to-image
// request: png format, quality 94, zoom 1.0.
func newHTMLToImageRequest(req models.HTMLToImageRequest) models.HTMLToImageRequest {
	req.Format = strings.ToLower(strings.TrimSpace(req.Format))
	switch req.Format {
	case "jpeg":
		req.Format = "jpg"
	case "":
		req.Format = "png"
	}
	if req.Quality == 0 {
		req.Quality = 94
	}
	if req.Zoom == 0 {
		req.Zoom = 1.0
	}
	return req
}
