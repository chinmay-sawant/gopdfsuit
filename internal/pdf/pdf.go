package pdf

import (
	"fmt"
	"log"

	"github.com/chinmay-sawant/gochromedp/pkg/gochromedp"
	"github.com/chinmay-sawant/gopdfsuit/v4/internal/models"
)

// The original `pdf.go` was large. It has been split into smaller files by responsibility:
// - types.go        (page size and dimensions)
// - utils.go        (parsing helpers and string escaping)
// - pagemanager.go  (PageManager and page lifecycle)
// - draw.go         (drawing helpers: title, table, footer, watermark)
// - generator.go    (GenerateTemplatePDF and orchestration)
// - xfdf.go         (XFDF parsing and PDF form filling)

// This file intentionally left minimal to keep package build roots simple.

// ConvertHTMLToPDF converts HTML content to PDF using gochromedp.
func ConvertHTMLToPDF(req models.HTMLToPDFRequest) ([]byte, error) {
	log.Printf("ConvertHTMLToPDF: Starting conversion. HTML length: %d, URL: %s", len(req.HTML), req.URL)

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

	log.Printf("ConvertHTMLToPDF: Options prepared - PageSize: %s, Orientation: %s, Grayscale: %t",
		options.PageSize, options.Orientation, options.Grayscale)

	var pdfData []byte
	var err error

	if req.HTML != "" {
		log.Printf("ConvertHTMLToPDF: Converting HTML content")
		// Convert HTML content
		pdfData, err = gochromedp.ConvertHTMLToPDF(req.HTML, options)
	} else if req.URL != "" {
		log.Printf("ConvertHTMLToPDF: Converting URL: %s", req.URL)
		// Convert URL
		pdfData, err = gochromedp.ConvertURLToPDF(req.URL, options)
	} else {
		log.Printf("ConvertHTMLToPDF: Error - neither HTML nor URL provided")
		return nil, fmt.Errorf("either HTML content or URL must be provided")
	}

	if err != nil {
		log.Printf("ConvertHTMLToPDF: Conversion failed with error: %v", err)
		return nil, fmt.Errorf("PDF conversion failed: %v", err)
	}

	log.Printf("ConvertHTMLToPDF: Conversion successful. PDF size: %d bytes", len(pdfData))
	return pdfData, nil
}

// ConvertHTMLToImage converts HTML content to image using gochromedp.
func ConvertHTMLToImage(req models.HTMLToImageRequest) ([]byte, error) {
	log.Printf("ConvertHTMLToImage: Starting conversion. HTML length: %d, URL: %s, Format: %s", len(req.HTML), req.URL, req.Format)

	// Prepare options
	options := &gochromedp.ConvertOptions{
		Format:  req.Format,
		Width:   req.Width,
		Height:  req.Height,
		Quality: req.Quality,
	}

	log.Printf("ConvertHTMLToImage: Options prepared - Format: %s, Width: %d, Height: %d, Quality: %d",
		options.Format, options.Width, options.Height, options.Quality)

	var imageData []byte
	var err error

	if req.HTML != "" {
		log.Printf("ConvertHTMLToImage: Converting HTML content")
		// Convert HTML content
		imageData, err = gochromedp.ConvertHTMLToImage(req.HTML, options)
	} else if req.URL != "" {
		log.Printf("ConvertHTMLToImage: Converting URL: %s", req.URL)
		// Convert URL
		imageData, err = gochromedp.ConvertURLToImage(req.URL, options)
	} else {
		log.Printf("ConvertHTMLToImage: Error - neither HTML nor URL provided")
		return nil, fmt.Errorf("either HTML content or URL must be provided")
	}

	if err != nil {
		log.Printf("ConvertHTMLToImage: Conversion failed with error: %v", err)
		return nil, fmt.Errorf("image conversion failed: %v", err)
	}

	log.Printf("ConvertHTMLToImage: Conversion successful. Image size: %d bytes", len(imageData))
	return imageData, nil
}
