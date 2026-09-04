//go:build js

package pdf

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

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
// file sources return ErrUpstream.
func ConvertHTMLToPDF(req models.HTMLToPDFRequest) ([]byte, error) {
	content, err := inlineHTMLContent(req.HTML, req.URL)
	if err != nil {
		return nil, err
	}

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

	pdfData, err := doc.PDF(context.Background())
	if err != nil {
		return nil, fmt.Errorf("PDF conversion failed: %w", err)
	}
	return pdfData, nil
}

// ConvertHTMLToImage renders an inline HTML string to png/jpg via pure-Go
// gowkhtmltopdf. Format svg has no engine equivalent and fails fast; URL or
// file sources return ErrUpstream.
func ConvertHTMLToImage(req models.HTMLToImageRequest) ([]byte, error) {
	content, err := inlineHTMLContent(req.HTML, req.URL)
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

	imageData, err := imgDoc.Image(context.Background())
	if err != nil {
		return nil, fmt.Errorf("image conversion failed: %w", err)
	}
	return imageData, nil
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

// defaultMarginMM mirrors pdf.go: the fallback matches "10mm".
const defaultMarginMM = 10

// parseMarginMM mirrors pdf.go: "10mm"-style strings to millimetres.
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
