//go:build js

package pdf

import (
	"errors"
	"fmt"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
	gowkhtmltopdf "github.com/chinmay-sawant/gowkhtmltopdf"
)

// ErrUpstream marks URL/file HTML conversion as unavailable in the js/wasm
// build. Fetching remote pages needs net/http plus the SSRF guard in
// internal/handlers, so URL callers must use /api/v1/*. Inline HTML strings
// render fully in-browser below.
var ErrUpstream = errors.New("pdf: upstream failure: URL/file HTML conversion is server-side only, unsupported in WASM")

// ConvertHTMLToPDF renders an inline HTML string to PDF via pure-Go
// gowkhtmltopdf (no browser, embedded font fallback, no filesystem). URL or
// file sources return ErrUpstream. Document construction and the
// DPI/LowQuality/Options gap policy live in html_convert.go, shared with
// the server build.
func ConvertHTMLToPDF(req models.HTMLToPDFRequest) ([]byte, error) {
	content, err := inlineHTMLContent(req.HTML, req.URL)
	if err != nil {
		return nil, err
	}

	warnUnmappedHTMLOptions("ConvertHTMLToPDF", req.Options)
	return runPDFDocument(buildPDFDocument(req, content))
}

// ConvertHTMLToImage renders an inline HTML string to png/jpg via pure-Go
// gowkhtmltopdf. Format svg has no engine equivalent and fails fast; URL or
// file sources return ErrUpstream.
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

// inlineHTMLContent maps exactly-one-of HTML/URL onto inline content. URL
// and file sources need the server fetch path, so they fail here.
func inlineHTMLContent(html, rawURL string) (gowkhtmltopdf.Content, error) {
	switch {
	case html != "":
		return gowkhtmltopdf.HTML([]byte(html)), nil
	case rawURL != "":
		return gowkhtmltopdf.Content{}, ErrUpstream
	default:
		return gowkhtmltopdf.Content{}, fmt.Errorf("either HTML content or URL must be provided")
	}
}
