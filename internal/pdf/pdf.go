package pdf

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/chinmay-sawant/gochromedp/pkg/gochromedp"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
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

// ConvertHTMLToPDF converts HTML content to PDF using gochromedp.
func ConvertHTMLToPDF(req models.HTMLToPDFRequest) ([]byte, error) {
	htmlDebugf("ConvertHTMLToPDF: Starting conversion. HTML length: %d, URL: %s", len(req.HTML), redactedURLForLog(req.URL))

	// Prepare options
	options := &gochromedp.ConvertOptions{
		PageSize:     req.PageSize,
		Orientation:  req.Orientation,
		MarginTop:    req.MarginTop,
		MarginRight:  req.MarginRight,
		MarginBottom: req.MarginBottom,
		MarginLeft:   req.MarginLeft,
	}

	// Handle PDF-specific options
	if req.Grayscale {
		options.Grayscale = true
	}
	// Note: LowQuality option not available in gochromedp ConvertOptions

	htmlDebugf("ConvertHTMLToPDF: Options prepared - PageSize: %s, Orientation: %s, Grayscale: %t",
		options.PageSize, options.Orientation, options.Grayscale)

	var pdfData []byte
	var err error

	switch {
	case req.HTML != "":
		htmlDebugf("ConvertHTMLToPDF: Converting HTML content")
		// Convert HTML content
		pdfData, err = gochromedp.ConvertHTMLToPDF(req.HTML, options)
	case req.URL != "":
		htmlDebugf("ConvertHTMLToPDF: Converting URL: %s", redactedURLForLog(req.URL))
		// Convert URL
		pdfData, err = gochromedp.ConvertURLToPDF(req.URL, options)
	default:
		return nil, fmt.Errorf("either HTML content or URL must be provided")
	}

	if err != nil {
		return nil, fmt.Errorf("PDF conversion failed: %w", err)
	}

	htmlDebugf("ConvertHTMLToPDF: Conversion successful. PDF size: %d bytes", len(pdfData))
	return pdfData, nil
}

// ConvertHTMLToImage converts HTML content to image using gochromedp.
func ConvertHTMLToImage(req models.HTMLToImageRequest) ([]byte, error) {
	htmlDebugf("ConvertHTMLToImage: Starting conversion. HTML length: %d, URL: %s, Format: %s", len(req.HTML), redactedURLForLog(req.URL), req.Format)

	// Prepare options
	options := &gochromedp.ConvertOptions{
		Format:  req.Format,
		Width:   req.Width,
		Height:  req.Height,
		Quality: req.Quality,
	}

	htmlDebugf("ConvertHTMLToImage: Options prepared - Format: %s, Width: %d, Height: %d, Quality: %d",
		options.Format, options.Width, options.Height, options.Quality)

	var imageData []byte
	var err error

	switch {
	case req.HTML != "":
		htmlDebugf("ConvertHTMLToImage: Converting HTML content")
		// Convert HTML content
		imageData, err = gochromedp.ConvertHTMLToImage(req.HTML, options)
	case req.URL != "":
		htmlDebugf("ConvertHTMLToImage: Converting URL: %s", redactedURLForLog(req.URL))
		// Convert URL
		imageData, err = gochromedp.ConvertURLToImage(req.URL, options)
	default:
		return nil, fmt.Errorf("either HTML content or URL must be provided")
	}

	if err != nil {
		return nil, fmt.Errorf("image conversion failed: %w", err)
	}

	htmlDebugf("ConvertHTMLToImage: Conversion successful. Image size: %d bytes", len(imageData))
	return imageData, nil
}
