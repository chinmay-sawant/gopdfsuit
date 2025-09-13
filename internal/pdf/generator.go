package pdf

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/internal/models"
	"github.com/gin-gonic/gin"
)

// GenerateTemplatePDF generates a PDF document with multi-page support based on a template
func GenerateTemplatePDF(c *gin.Context, template models.PDFTemplate) {
	var pdfBuffer bytes.Buffer
	xrefOffsets := make(map[int]int)

	// Get page dimensions from config
	pageConfig := template.Config
	pageDims := getPageDimensions(pageConfig.Page, pageConfig.PageAlignment)

	// Initialize page manager
	pageManager := NewPageManager(pageDims)

	// PDF Header
	pdfBuffer.WriteString("%PDF-1.7\n")
	pdfBuffer.WriteString("%âãÏÓ\n")

	// Object 1: Catalog
	xrefOffsets[1] = pdfBuffer.Len()
	pdfBuffer.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")

	// Generate all content first to know how many pages we need
	generateAllContent(template, pageManager)

	// Object 2: Pages (will be updated after we know total page count)
	xrefOffsets[2] = pdfBuffer.Len()
	pdfBuffer.WriteString("2 0 obj\n")
	pdfBuffer.WriteString(fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>\n",
		formatPageKids(pageManager.Pages), len(pageManager.Pages)))
	pdfBuffer.WriteString("endobj\n")

	// Calculate object IDs
	totalPages := len(pageManager.Pages)
	contentObjectStart := totalPages + 3               // Content objects start after pages
	fontObjectStart := contentObjectStart + totalPages // Fonts start after content

	// Generate page objects
	for i, pageID := range pageManager.Pages {
		xrefOffsets[pageID] = pdfBuffer.Len()
		pdfBuffer.WriteString(fmt.Sprintf("%d 0 obj\n", pageID))
		pdfBuffer.WriteString(fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %.2f %.2f] ",
			pageDims.Width, pageDims.Height))
		pdfBuffer.WriteString(fmt.Sprintf("/Contents %d 0 R ", contentObjectStart+i))
		pdfBuffer.WriteString(fmt.Sprintf("/Resources << /Font << /F1 %d 0 R /F2 %d 0 R /F3 %d 0 R /F4 %d 0 R >> >> >>\n",
			fontObjectStart, fontObjectStart+1, fontObjectStart+2, fontObjectStart+3))
		pdfBuffer.WriteString("endobj\n")
	}

	// Generate content stream objects
	for i, contentStream := range pageManager.ContentStreams {
		objectID := contentObjectStart + i
		xrefOffsets[objectID] = pdfBuffer.Len()
		pdfBuffer.WriteString(fmt.Sprintf("%d 0 obj\n", objectID))
		pdfBuffer.WriteString(fmt.Sprintf("<< /Length %d >>\n", contentStream.Len()))
		pdfBuffer.WriteString("stream\n")
		pdfBuffer.Write(contentStream.Bytes())
		pdfBuffer.WriteString("\nendstream\nendobj\n")
	}

	// Generate font objects
	fontNames := []string{"Helvetica", "Helvetica-Bold", "Helvetica-Oblique", "Helvetica-BoldOblique"}
	fontRefs := []string{"/F1", "/F2", "/F3", "/F4"}

	for i, fontName := range fontNames {
		objectID := fontObjectStart + i
		xrefOffsets[objectID] = pdfBuffer.Len()
		pdfBuffer.WriteString(fmt.Sprintf("%d 0 obj\n", objectID))
		pdfBuffer.WriteString(fmt.Sprintf("<< /Type /Font /Subtype /Type1 /Name %s /BaseFont /%s >>\n",
			fontRefs[i], fontName))
		pdfBuffer.WriteString("endobj\n")
	}

	// Cross-reference table
	totalObjects := fontObjectStart + 4
	xrefStart := pdfBuffer.Len()
	pdfBuffer.WriteString(fmt.Sprintf("xref\n0 %d\n0000000000 65535 f \n", totalObjects))
	for i := 1; i < totalObjects; i++ {
		pdfBuffer.WriteString(fmt.Sprintf("%010d 00000 n \n", xrefOffsets[i]))
	}

	// Trailer
	pdfBuffer.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R >>\n", totalObjects))
	pdfBuffer.WriteString("startxref\n")
	pdfBuffer.WriteString(strconv.Itoa(xrefStart) + "\n")
	pdfBuffer.WriteString("%%EOF\n")

	// HTTP Response
	filename := fmt.Sprintf("template-pdf-%d.pdf", time.Now().Unix())
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/pdf", pdfBuffer.Bytes())
}

// generateAllContent processes the template and generates content across multiple pages
func generateAllContent(template models.PDFTemplate, pageManager *PageManager) {
	// Initialize first page
	initializePage(pageManager.GetCurrentContentStream(), template.Config.PageBorder, template.Config.Watermark, pageManager.PageDimensions)

	// Title - Check if it fits on current page
	titleProps := parseProps(template.Title.Props)
	titleHeight := float64(titleProps.FontSize + 50) // Title + spacing

	if pageManager.CheckPageBreak(titleHeight) {
		pageManager.AddNewPage()
		initializePage(pageManager.GetCurrentContentStream(), template.Config.PageBorder, template.Config.Watermark, pageManager.PageDimensions)
	}

	drawTitle(pageManager.GetCurrentContentStream(), template.Title, titleProps, pageManager)

	// Tables - Process each table with automatic page breaks
	for _, table := range template.Table {
		drawTable(table, pageManager, template.Config.PageBorder, template.Config.Watermark)
	}

	// Draw footer and page numbers on every page (footer first to avoid overlap)
	totalPages := len(pageManager.Pages)
	for i := 0; i < totalPages; i++ {
		// Draw footer on this page if footer text provided
		if template.Footer.Text != "" {
			drawFooter(&pageManager.ContentStreams[i], template.Footer)
		}
		// Draw page number on this page
		drawPageNumber(&pageManager.ContentStreams[i], i+1, totalPages, pageManager.PageDimensions)
	}
}
