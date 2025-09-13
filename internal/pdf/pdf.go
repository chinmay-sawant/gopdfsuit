package pdf

import (
	"fmt"

	"github.com/chinmay-sawant/gochromedp/pkg/gochromedp"
	"github.com/chinmay-sawant/gopdfsuit/internal/models"
)

// The original `pdf.go` was large. It has been split into smaller files by responsibility:
// - types.go        (page size and dimensions)
// - utils.go        (parsing helpers and string escaping)
// - pagemanager.go  (PageManager and page lifecycle)
// - draw.go         (drawing helpers: title, table, footer, watermark)
// - generator.go    (GenerateTemplatePDF and orchestration)
// - xfdf.go         (XFDF parsing and PDF form filling)

// This file intentionally left minimal to keep package build roots simple.

// ConvertHTMLToPDF converts HTML content to PDF using gochromedp
func ConvertHTMLToPDF(req models.HtmlToPDFRequest) ([]byte, error) {
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

	var pdfData []byte
	var err error

	if req.HTML != "" {
		// Convert HTML content
		pdfData, err = gochromedp.ConvertHTMLToPDF(req.HTML, options)
	} else if req.URL != "" {
		// Convert URL
		pdfData, err = gochromedp.ConvertURLToPDF(req.URL, options)
	} else {
		return nil, fmt.Errorf("either HTML content or URL must be provided")
	}

	if err != nil {
		return nil, fmt.Errorf("PDF conversion failed: %v", err)
	}

	return pdfData, nil
}

// ConvertHTMLToImage converts HTML content to image using gochromedp
func ConvertHTMLToImage(req models.HtmlToImageRequest) ([]byte, error) {
	// Prepare options
	options := &gochromedp.ConvertOptions{
		Format:  req.Format,
		Width:   req.Width,
		Height:  req.Height,
		Quality: req.Quality,
	}

	var imageData []byte
	var err error

	if req.HTML != "" {
		// Convert HTML content
		imageData, err = gochromedp.ConvertHTMLToImage(req.HTML, options)
	} else if req.URL != "" {
		// Convert URL
		imageData, err = gochromedp.ConvertURLToImage(req.URL, options)
	} else {
		return nil, fmt.Errorf("either HTML content or URL must be provided")
	}

	if err != nil {
		return nil, fmt.Errorf("image conversion failed: %v", err)
	}

	return imageData, nil
}
