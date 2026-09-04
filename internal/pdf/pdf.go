//go:build !js

package pdf

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

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
	policy := gowkhtmltopdf.RestrictedNetworkPolicy()
	doc := &gowkhtmltopdf.Document{
		Pages:       []gowkhtmltopdf.Page{{Source: content}},
		PageSize:    req.PageSize,
		Orientation: req.Orientation,
		Margin: gowkhtmltopdf.Margin{
			Top:    parseMarginMM(req.MarginTop),
			Right:  parseMarginMM(req.MarginRight),
			Bottom: parseMarginMM(req.MarginBottom),
			Left:   parseMarginMM(req.MarginLeft),
		},
		Grayscale: req.Grayscale,
		Network:   &policy,
	}
	// Note: LowQuality, DPI, and free-form Options have no gowkhtmltopdf
	// equivalent and are accepted but ignored (see models.HTMLToPDFRequest).

	htmlDebugf("ConvertHTMLToPDF: Options prepared - PageSize: %s, Orientation: %s, Grayscale: %t",
		req.PageSize, req.Orientation, req.Grayscale)

	pdfData, err := doc.PDF(context.Background())
	if err != nil {
		return nil, fmt.Errorf("PDF conversion failed: %w", err)
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

	format := strings.ToLower(strings.TrimSpace(req.Format))
	if format == "" {
		format = "png"
	}
	if format == "svg" {
		return nil, fmt.Errorf("unsupported image format %q: gowkhtmltopdf supports png and jpg only", req.Format)
	}

	var crop *gowkhtmltopdf.Crop
	if req.CropWidth > 0 && req.CropHeight > 0 {
		crop = &gowkhtmltopdf.Crop{Left: req.CropX, Top: req.CropY, Width: req.CropWidth, Height: req.CropHeight}
	}

	htmlDebugf("ConvertHTMLToImage: Options prepared - Format: %s, Width: %d, Height: %d, Quality: %d",
		format, req.Width, req.Height, req.Quality)

	// Defense in depth: see ConvertHTMLToPDF on validateFetchURL plus the
	// restricted in-engine network policy.
	policy := gowkhtmltopdf.RestrictedNetworkPolicy()
	imgDoc := &gowkhtmltopdf.ImageDocument{
		Source:  content,
		Width:   req.Width,
		Height:  req.Height,
		Format:  format,
		Quality: req.Quality,
		Zoom:    req.Zoom,
		Crop:    crop,
		Network: &policy,
	}
	// Note: free-form Options has no gowkhtmltopdf equivalent and is
	// accepted but ignored (see models.HTMLToImageRequest).

	imageData, err := imgDoc.Image(context.Background())
	if err != nil {
		return nil, fmt.Errorf("image conversion failed: %w", err)
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

// defaultMarginMM is the fallback when a margin string is empty or
// unparseable. It matches the handler default of "10mm".
const defaultMarginMM = 10

// parseMarginMM parses "10mm"-style margin strings to millimetres. Bare
// numbers mean mm; cm, in, pt, and px (96dpi) are converted. Unparseable or
// negative values fall back to defaultMarginMM rather than failing the
// conversion.
func parseMarginMM(raw string) float64 {
	s := strings.TrimSpace(strings.ToLower(raw))
	if s == "" {
		return defaultMarginMM
	}
	mult := 1.0
	for _, suffix := range []struct {
		suffix string
		mult   float64
	}{
		{"mm", 1.0},
		{"cm", 10.0},
		{"in", 25.4},
		{"pt", 25.4 / 72.0},
		{"px", 25.4 / 96.0},
	} {
		if strings.HasSuffix(s, suffix.suffix) {
			mult = suffix.mult
			s = strings.TrimSpace(strings.TrimSuffix(s, suffix.suffix))
			break
		}
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil || v < 0 {
		return defaultMarginMM
	}
	return v * mult
}
