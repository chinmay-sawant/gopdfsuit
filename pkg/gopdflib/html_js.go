//go:build js

package gopdflib

import (
	"fmt"

	"github.com/chinmay-sawant/gopdfsuit/v7/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/v7/internal/pdf"
)

// ConvertHTMLToPDF renders HTML in-browser via pure-Go gowkhtmltopdf. Pass
// inline markup as HTML, optionally with URL set as the document base URL
// so relative subresource references resolve against the page origin (this
// is the shape the WASM binding sends after pre-fetching page HTML via
// browser fetch). URL-only requests fail fast in the engine with guidance
// instead of dialing: raw sockets cannot work under js/wasm.
// Empty requests still report ErrInvalidInput, matching the server contract.
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

// ConvertHTMLToImage renders HTML in-browser via pure-Go gowkhtmltopdf
// (png/jpg; svg rejected as invalid input). Like ConvertHTMLToPDF it takes
// inline HTML plus an optional base URL; URL-only requests fail fast with
// guidance because the engine cannot dial under js/wasm.
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
