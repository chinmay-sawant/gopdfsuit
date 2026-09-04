package pdf

import (
	"context"
	"fmt"
	"log"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
	gowkhtmltopdf "github.com/chinmay-sawant/gowkhtmltopdf"
)

// Shared HTML-convert core for the server (!js) and WASM (js) builds.
// pdf.go and pdf_js.go keep only their source-policy funcs
// (htmlSourceContent allows URL fetches; inlineHTMLContent returns
// ErrUpstream) plus thin Convert wrappers; everything below is identical
// on both targets.

// Request-field to gowkhtmltopdf-knob mapping (single table, both builds):
//
//	mapped:  PageSize, Orientation, margins, Grayscale -> Document fields
//	mapped:  Format, Width, Height, Quality, Zoom, Crop* -> ImageDocument fields
//	mapped:  HTML/URL -> Content (per-build source policy)
//	ignored: DPI, LowQuality -> Document has no such knob in gowkhtmltopdf
//	         v0.2.5 (verified: no DPI/LowQuality field); accepted, ignored
//	ignored: Options -> free-form wkhtmltopdf flags have no equivalent;
//	         accepted, ignored with a warnUnmappedHTMLOptions log
func warnUnmappedHTMLOptions(op string, opts map[string]string) {
	if len(opts) == 0 {
		return
	}
	log.Printf("%s: %d html option(s) have no gowkhtmltopdf equivalent and are ignored: %s",
		op, len(opts), strings.Join(slices.Sorted(maps.Keys(opts)), ","))
}

// buildPDFDocument maps a validated HTML-to-PDF request onto a
// gowkhtmltopdf Document with the restricted network policy.
func buildPDFDocument(req models.HTMLToPDFRequest, content gowkhtmltopdf.Content) *gowkhtmltopdf.Document {
	policy := gowkhtmltopdf.RestrictedNetworkPolicy()
	return &gowkhtmltopdf.Document{
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
}

// runPDFDocument renders doc and folds failures into the shared message.
func runPDFDocument(doc *gowkhtmltopdf.Document) ([]byte, error) {
	pdfData, err := doc.PDF(context.Background())
	if err != nil {
		return nil, fmt.Errorf("PDF conversion failed: %w", err)
	}
	return pdfData, nil
}

// normalizeImageFormat lowercases/trims the requested image format,
// defaulting empty to png. svg has no engine equivalent and fails fast.
func normalizeImageFormat(raw string) (string, error) {
	format := strings.ToLower(strings.TrimSpace(raw))
	if format == "" {
		format = "png"
	}
	if format == "svg" {
		return "", fmt.Errorf("unsupported image format %q: gowkhtmltopdf supports png and jpg only", raw)
	}
	return format, nil
}

// buildImageCrop maps a CropWidth x CropHeight request onto a crop rect;
// zero dimensions mean no crop.
func buildImageCrop(req models.HTMLToImageRequest) *gowkhtmltopdf.Crop {
	if req.CropWidth <= 0 || req.CropHeight <= 0 {
		return nil
	}
	return &gowkhtmltopdf.Crop{Left: req.CropX, Top: req.CropY, Width: req.CropWidth, Height: req.CropHeight}
}

// buildImageDocument maps a validated HTML-to-image request onto a
// gowkhtmltopdf ImageDocument with the restricted network policy.
func buildImageDocument(req models.HTMLToImageRequest, content gowkhtmltopdf.Content, format string) *gowkhtmltopdf.ImageDocument {
	policy := gowkhtmltopdf.RestrictedNetworkPolicy()
	return &gowkhtmltopdf.ImageDocument{
		Source:  content,
		Width:   req.Width,
		Height:  req.Height,
		Format:  format,
		Quality: req.Quality,
		Zoom:    req.Zoom,
		Crop:    buildImageCrop(req),
		Network: &policy,
	}
}

// runImageDocument renders imgDoc and folds failures into the shared message.
func runImageDocument(imgDoc *gowkhtmltopdf.ImageDocument) ([]byte, error) {
	imageData, err := imgDoc.Image(context.Background())
	if err != nil {
		return nil, fmt.Errorf("image conversion failed: %w", err)
	}
	return imageData, nil
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
