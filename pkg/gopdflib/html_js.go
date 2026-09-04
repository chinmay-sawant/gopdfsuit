//go:build js

package gopdflib

import (
	"fmt"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf"
)

// ConvertHTMLToPDF renders an inline HTML string in-browser via pure-Go
// gowkhtmltopdf. URL sources need the server fetch path (SSRF guard plus
// network) and return ErrUpstream; use POST /api/v1/htmltopdf for those.
// Empty requests still report ErrInvalidInput, matching the server contract.
func ConvertHTMLToPDF(req HTMLToPDFRequest) ([]byte, error) {
	const op = "gopdflib: ConvertHTMLToPDF"
	if req.HTML == "" && req.URL == "" {
		return nil, invalidInputError(op, "needs HTML content or a URL")
	}
	if req.URL != "" && req.HTML == "" {
		return nil, fmt.Errorf("%w: %s: URL conversion is server-side only", ErrUpstream, op)
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

// ConvertHTMLToImage renders an inline HTML string in-browser via pure-Go
// gowkhtmltopdf (png/jpg; svg rejected as invalid input). URL sources need
// the server fetch path and return ErrUpstream; use POST /api/v1/htmltoimage.
func ConvertHTMLToImage(req HTMLToImageRequest) ([]byte, error) {
	const op = "gopdflib: ConvertHTMLToImage"
	if req.HTML == "" && req.URL == "" {
		return nil, invalidInputError(op, "needs HTML content or a URL")
	}
	if req.Format == "svg" {
		return nil, invalidInputError(op, "format svg is not supported: use png or jpg")
	}
	if req.URL != "" && req.HTML == "" {
		return nil, fmt.Errorf("%w: %s: URL conversion is server-side only", ErrUpstream, op)
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
