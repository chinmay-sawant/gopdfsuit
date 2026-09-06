//go:build js

package pdf

import (
	"fmt"

	"github.com/chinmay-sawant/gopdfsuit/v7/internal/models"
	gowkhtmltopdf "github.com/chinmay-sawant/gowkhtmltopdf"
)

// ConvertHTMLToPDF renders pre-fetched HTML to PDF via pure-Go
// gowkhtmltopdf (no browser, embedded font fallback, no filesystem). The
// page URL, when present, is passed as the document base URL so relative
// subresource references resolve against it. URL-only requests fail fast
// here: the engine loader dials raw sockets for DNS/HTTP, which cannot work
// under js/wasm, so the binding layer (cmd/wasm) pre-fetches page HTML via
// browser fetch and hands over HTML plus URL together.
// Document construction and the DPI/LowQuality/Options gap policy live in
// html_convert.go, shared with the server build.
func ConvertHTMLToPDF(req models.HTMLToPDFRequest) ([]byte, error) {
	content, err := inlineHTMLContent(req.HTML, req.URL)
	if err != nil {
		return nil, err
	}

	warnUnmappedHTMLOptions("ConvertHTMLToPDF", req.Options)
	return runPDFDocument(buildPDFDocument(req, content))
}

// ConvertHTMLToImage renders pre-fetched HTML to png/jpg via pure-Go
// gowkhtmltopdf. Format svg has no engine equivalent and fails fast. The
// page URL, when present, is the document base URL for relative
// subresources. URL-only requests fail fast (see ConvertHTMLToPDF): the
// binding layer pre-fetches page HTML via browser fetch.
func ConvertHTMLToImage(req models.HTMLToImageRequest) ([]byte, error) {
	content, err := inlineHTMLContent(req.HTML, req.URL)
	if err != nil {
		return nil, err
	}

	format, err := normalizeImageFormat(req.Format)
	if err != nil {
		return nil, err
	}

	warnUnmappedHTMLOptions("ConvertHTMLToImage", req.Options)
	return runImageDocument(buildImageDocument(req, content, format))
}

// inlineHTMLContent maps the request onto engine content. HTML plus an
// optional page URL becomes an inline document with that URL as its base,
// so relative subresource references resolve against the page origin.
// URL-only input fails fast with guidance instead of reaching the engine
// loader: the loader dials raw sockets for DNS/HTTP, which cannot work
// under js/wasm (no sockets), so callers must pre-fetch the page HTML via
// browser fetch and pass HTML plus URL together.
func inlineHTMLContent(html, rawURL string) (gowkhtmltopdf.Content, error) {
	switch {
	case html != "":
		return gowkhtmltopdf.HTML([]byte(html), rawURL), nil
	case rawURL != "":
		return gowkhtmltopdf.Content{}, fmt.Errorf("URL-only conversion is unsupported in the browser build: fetch the page HTML with browser fetch and convert it as HTML with the page URL as base (sites without CORS headers cannot be fetched cross-origin)")
	default:
		return gowkhtmltopdf.Content{}, fmt.Errorf("either HTML content or URL must be provided")
	}
}

// htmlSourceContent keeps the shared context-aware conversion path on the
// same inline-only source policy as the public browser wrappers.
func htmlSourceContent(html, rawURL string) (gowkhtmltopdf.Content, error) {
	return inlineHTMLContent(html, rawURL)
}
