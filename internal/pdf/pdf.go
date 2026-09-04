//go:build !js

package pdf

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
	gowkhtmltopdf "github.com/chinmay-sawant/gowkhtmltopdf"
)

// The original `pdf.go` was large. It has been split into smaller files by responsibility:
// - types.go        (page size and dimensions)
// - utils.go        (parsing helpers and string escaping)
// - pagemanager.go  (PageManager and page lifecycle)
// - draw.go         (drawing helpers: title, table, footer, watermark)
// - generator.go    (GenerateTemplatePDF and orchestration)
// - xfdf.go         (XFDF parsing and PDF form filling)

// This file intentionally left minimal to keep package build roots simple.

// htmlDebug gates ConvertHTML lifecycle logs. The per-step lines below are
// request noise in production; set GOPDFSUIT_DEBUG=1 to restore them when
// diagnosing conversion failures.
var htmlDebug = os.Getenv("GOPDFSUIT_DEBUG") == "1"

func htmlDebugf(format string, args ...any) {
	if htmlDebug {
		log.Printf(format, args...)
	}
}

// redactedURLForLog renders a fetch URL for logs as host + path only. Query,
// fragment, and userinfo are dropped because they may carry tokens or PII.
func redactedURLForLog(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return "(unparseable url)"
	}
	return u.Host + u.EscapedPath()
}

// ConvertHTMLToPDF converts HTML content to PDF using gowkhtmltopdf
// (pure-Go, no browser). PDFVersion/PDFProfile are intentionally left unset
// so the engine emits its default unclaimed PDF 1.4; pinning a3a-ua1/a4-ua2
// to match the template pipeline's PDF/A-4 and PDF/UA-2 claims is a Phase 4
// veraPDF re-baseline decision, not an adapter default.
func ConvertHTMLToPDF(req models.HTMLToPDFRequest) ([]byte, error) {
	htmlDebugf("ConvertHTMLToPDF: Starting conversion. HTML length: %d, URL: %s", len(req.HTML), redactedURLForLog(req.URL))

	content, err := htmlSourceContent(req.HTML, req.URL)
	if err != nil {
		return nil, err
	}

	// Defense in depth: handlers run validateFetchURL first; the restricted
	// policy additionally blocks private/link-local fetches and cross-host
	// redirects for the page plus its subresource CSS inside the engine.
	// DPI, LowQuality, and Options have no gowkhtmltopdf equivalent and are
	// accepted but ignored (see the mapping table in html_convert.go).
	warnUnmappedHTMLOptions("ConvertHTMLToPDF", req.Options)
	doc := buildPDFDocument(req, content)

	htmlDebugf("ConvertHTMLToPDF: Options prepared - PageSize: %s, Orientation: %s, Grayscale: %t",
		req.PageSize, req.Orientation, req.Grayscale)

	pdfData, err := runPDFDocument(doc)
	if err != nil {
		return nil, err
	}

	htmlDebugf("ConvertHTMLToPDF: Conversion successful. PDF size: %d bytes", len(pdfData))
	return pdfData, nil
}

// ConvertHTMLToImage converts HTML content to image using gowkhtmltopdf
// (pure-Go, no browser). Supported formats are png and jpg/jpeg; svg has no
// engine equivalent and fails fast as invalid input.
func ConvertHTMLToImage(req models.HTMLToImageRequest) ([]byte, error) {
	htmlDebugf("ConvertHTMLToImage: Starting conversion. HTML length: %d, URL: %s, Format: %s", len(req.HTML), redactedURLForLog(req.URL), req.Format)

	content, err := htmlSourceContent(req.HTML, req.URL)
	if err != nil {
		return nil, err
	}

	format, err := normalizeImageFormat(req.Format)
	if err != nil {
		return nil, err
	}

	htmlDebugf("ConvertHTMLToImage: Options prepared - Format: %s, Width: %d, Height: %d, Quality: %d",
		format, req.Width, req.Height, req.Quality)

	// Defense in depth: see ConvertHTMLToPDF on validateFetchURL plus the
	// restricted in-engine network policy. Options has no gowkhtmltopdf
	// equivalent and is accepted but ignored (see html_convert.go).
	warnUnmappedHTMLOptions("ConvertHTMLToImage", req.Options)
	imgDoc := buildImageDocument(req, content, format)

	imageData, err := runImageDocument(imgDoc)
	if err != nil {
		return nil, err
	}

	htmlDebugf("ConvertHTMLToImage: Conversion successful. Image size: %d bytes", len(imageData))
	return imageData, nil
}

// htmlSourceContent maps the request's exactly-one-of HTML/URL shape onto a
// gowkhtmltopdf Content value. Empty input fails fast before the engine runs.
func htmlSourceContent(html, rawURL string) (gowkhtmltopdf.Content, error) {
	switch {
	case html != "":
		return gowkhtmltopdf.HTML([]byte(html)), nil
	case rawURL != "":
		return gowkhtmltopdf.URL(rawURL), nil
	default:
		return gowkhtmltopdf.Content{}, fmt.Errorf("either HTML content or URL must be provided")
	}
}

// defaultMarginMM, parseMarginMM, and the Document/ImageDocument builders
// live in html_convert.go, shared with the WASM build.
